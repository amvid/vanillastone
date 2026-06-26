// Package transport handles WebSocket connections: it upgrades requests, runs
// per-connection read/write loops, and routes client messages into the lobby
// and match. The client is a dumb renderer; all decisions happen here/below.
package transport

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/amvid/vanillastone/internal/auth"
	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/lobby"
	"github.com/amvid/vanillastone/internal/match"
	"github.com/amvid/vanillastone/internal/protocol"
	"github.com/amvid/vanillastone/internal/store"
	"github.com/coder/websocket"
)

// defaultGraceWindow is how long a dropped player's seat is held open for a
// reconnect before the match is forfeited on their behalf. Stored per-Server
// (see Server.graceWindow) so tests can shorten it without a shared global.
const defaultGraceWindow = 60 * time.Second

// activeSeat records the match a username is seated in and the player-slot id
// they hold, so a reconnecting login can be swapped back into the same seat
// (Match.Reattach) instead of being sent to the lobby.
type activeSeat struct {
	m    *match.Match
	slot string // player-slot id (turn identity) to re-adopt on reconnect
}

// inviteRec is one player's outstanding direct invite: who they challenged and
// the deck they locked in at send time.
type inviteRec struct {
	target string
	deckID int64
}

// Server wires the WebSocket endpoint to the lobby. It also tracks the set of
// authenticated connections so it can report live presence counts to the lobby.
type Server struct {
	auth   *auth.Auth
	store  *store.Store
	lobby  *lobby.Lobby
	nextID atomic.Int64

	graceWindow time.Duration // how long a dropped seat is held for reconnect

	mu      sync.Mutex
	clients map[*Client]struct{}   // authenticated, live connections
	active  map[string]activeSeat  // username -> the live match + slot they occupy
	grace   map[string]*time.Timer // username -> pending forfeit timer during a disconnect grace window
	queued  map[string]struct{}    // usernames currently waiting in the matchmaking queue
	invites map[string]inviteRec   // inviter username -> their outstanding direct invite
}

// NewServer returns a transport server with a fresh lobby. The store is used to
// load a player's chosen deck when they queue.
func NewServer(a *auth.Auth, st *store.Store) *Server {
	return &Server{
		auth:        a,
		store:       st,
		lobby:       lobby.New(),
		graceWindow: defaultGraceWindow,
		clients:     make(map[*Client]struct{}),
		active:      make(map[string]activeSeat),
		grace:       make(map[string]*time.Timer),
		queued:      make(map[string]struct{}),
		invites:     make(map[string]inviteRec),
	}
}

// Client is one connected player. It implements match.Sender. Outbound writes
// go through send so the single writer goroutine owns the socket. match is
// atomic because presence broadcasts read it from other goroutines.
type Client struct {
	id    string
	name  string
	send  chan []byte
	match atomic.Pointer[match.Match]
	// spectating is the match this client is currently watching as a spectator
	// (nil when not spectating), so leaving/disconnecting can deregister it as an
	// observer. Distinct from match (a seated player), which a spectator never has.
	spectating atomic.Pointer[match.Match]
}

// ID implements match.Sender.
func (c *Client) ID() string { return c.id }

// Name implements match.Sender: the authenticated display username.
func (c *Client) Name() string { return c.name }

// Send implements match.Sender: queues a message for the writer goroutine.
// Non-blocking drop if the buffer is full keeps the match from stalling.
func (c *Client) Send(b []byte) {
	select {
	case c.send <- b:
	default:
		log.Printf("client %s send buffer full, dropping message", c.id)
	}
}

// Heartbeat tunables: ping every pingInterval, and consider the connection dead
// if a pong doesn't arrive within pingTimeout.
const (
	pingInterval = 15 * time.Second
	pingTimeout  = 10 * time.Second
)

// heartbeat pings the connection on a ticker and cancels ctx (reaping the
// connection) if a ping fails or times out, so dropped clients are removed from
// presence promptly instead of lingering until TCP times out.
func (s *Server) heartbeat(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, pingCancel := context.WithTimeout(ctx, pingTimeout)
			err := conn.Ping(pingCtx)
			pingCancel()
			if err != nil {
				cancel()
				return
			}
		}
	}
}

// HandleWS upgrades the connection and serves one client until disconnect.
func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("ws accept: %v", err)
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	c := &Client{
		id:   "p" + strconv.FormatInt(s.nextID.Add(1), 10),
		send: make(chan []byte, 16),
	}

	// Writer goroutine owns the socket for writes.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-c.send:
				if err := conn.Write(ctx, websocket.MessageText, b); err != nil {
					cancel()
					return
				}
			}
		}
	}()

	// Heartbeat goroutine: ping periodically and reap the connection if a pong
	// doesn't come back in time. Without this a dropped client (e.g. a page
	// reload whose old socket isn't cleanly closed) lingers until TCP notices.
	go s.heartbeat(ctx, cancel, conn)

	s.readLoop(ctx, conn, c)

	// Cleanup: leave the queue if still waiting. If dropping mid-match while still
	// the live occupant of the seat, hold it open for a grace window (the player
	// may reconnect) and tell the opponent; the timer forfeits if no one returns.
	// If the seat was already taken over by a reconnect, just deregister.
	s.lobby.Remove(c)
	s.mu.Lock()
	delete(s.queued, c.name)
	s.mu.Unlock()
	s.clearInvites(c.name)
	s.stopSpectating(c)
	if m := c.match.Load(); m != nil && !m.Over() && m.Seats(c) {
		s.startGrace(c, m)
		s.notifyOpp(m, c, false)
	} else {
		c.match.Store(nil)
	}
	s.deregister(c)
}

// readLoop reads and dispatches client messages until error/close.
func (s *Server) readLoop(ctx context.Context, conn *websocket.Conn, c *Client) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return // closed or context cancelled
		}
		var env protocol.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "bad json"}))
			continue
		}
		switch env.Type {
		case protocol.TypeAuth:
			s.handleAuth(data, c)
		case protocol.TypeFindMatch:
			s.handleFindMatch(data, c)
		case protocol.TypeEnterLobby:
			s.handleEnterLobby(c)
		case protocol.TypeEndTurn:
			s.handleEndTurn(c)
		case protocol.TypePlayCard:
			s.handlePlayCard(data, c)
		case protocol.TypeAttack:
			s.handleAttack(data, c)
		case protocol.TypeConcede:
			s.handleConcede(c)
		case protocol.TypeChoose:
			s.handleChoose(data, c)
		case protocol.TypeHeroPower:
			s.handleHeroPower(data, c)
		case protocol.TypeMulligan:
			s.handleMulligan(data, c)
		case protocol.TypeInvite:
			s.handleInvite(data, c)
		case protocol.TypeInviteCancel:
			s.handleInviteCancel(c)
		case protocol.TypeInviteRespond:
			s.handleInviteRespond(data, c)
		case protocol.TypeSpectate:
			s.handleSpectate(data, c)
		case protocol.TypeIntent:
			s.handleIntent(data, c)
		default:
			c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "unknown type: " + env.Type}))
		}
	}
}

func (s *Server) handleAuth(data []byte, c *Client) {
	if c.name != "" {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "already authenticated"}))
		return
	}
	var a protocol.Auth
	if err := json.Unmarshal(data, &a); err != nil || a.Token == "" {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "token required"}))
		return
	}
	name, ok := s.auth.Username(a.Token)
	if !ok {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "invalid or expired token"}))
		return
	}
	c.name = name
	// Reconnect: if this user is still seated in a live match, swap them back into
	// it (re-adopting their slot id) instead of returning them to the lobby.
	if s.tryReconnect(c) {
		return
	}
	// Single session per account: kick any earlier connection logged in as this
	// user before this one takes over.
	s.kickExisting(name, c)
	c.Send(protocol.Marshal(protocol.Joined{Type: protocol.TypeJoined, You: c.id, Name: name}))

	// Authentication lands the player in the lobby; matchmaking starts only when
	// the client sends FindMatch (the "Play" button).
	s.register(c)
}

// tryReconnect swaps a just-authenticated client back into a live match it still
// holds a seat in, cancelling any pending forfeit timer. It re-adopts the seat's
// player id (so turn identity is unchanged for both sides), pushes a fresh
// snapshot via Match.Reattach, and tells the opponent the player is back. Joined
// is sent BEFORE Reattach so the client records its id before the resync state
// arrives. Returns false (caller continues with the normal lobby flow) when the
// user has no live seat.
func (s *Server) tryReconnect(c *Client) bool {
	s.mu.Lock()
	seat, ok := s.active[c.name]
	if ok && (seat.m == nil || seat.m.Over()) {
		delete(s.active, c.name) // stale entry for a finished match
		ok = false
	}
	if ok {
		if t := s.grace[c.name]; t != nil { // returning in time: cancel the forfeit
			t.Stop()
			delete(s.grace, c.name)
		}
	}
	s.mu.Unlock()
	if !ok {
		return false
	}

	// A takeover login is possible (another live tab for this account); kick it
	// before seating this connection.
	s.kickExisting(c.name, c)
	c.id = seat.slot
	c.match.Store(seat.m)
	c.Send(protocol.Marshal(protocol.Joined{Type: protocol.TypeJoined, You: c.id, Name: c.name}))
	s.register(c)
	seat.m.Reattach(c) // pushes the resync snapshot to the returning client
	s.notifyOpp(seat.m, c, true)
	return true
}

// kickExisting tells every other live connection authenticated as name that it
// was logged in elsewhere. The client closes itself on this notice (clearing its
// token); its disconnect then forfeits any active match in HandleWS cleanup. We
// notify rather than force-close so the message reliably reaches the client
// before the socket drops.
func (s *Server) kickExisting(name string, except *Client) {
	s.mu.Lock()
	var victims []*Client
	for cl := range s.clients {
		if cl != except && cl.name == name {
			victims = append(victims, cl)
		}
	}
	s.mu.Unlock()
	for _, v := range victims {
		v.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "logged in elsewhere"}))
	}
}

// handleFindMatch enters the authenticated client into the matchmaking queue
// with the deck named by the FindMatch message (falling back to a default deck
// when none is chosen, the deck is missing, or it fails validation).
func (s *Server) handleFindMatch(data []byte, c *Client) {
	if c.name == "" {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "not authenticated"}))
		return
	}
	if cm := c.match.Load(); cm != nil && !cm.Over() {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "already in a match"}))
		return
	}
	var fm protocol.FindMatch
	json.Unmarshal(data, &fm) // best-effort; zero DeckID => default deck
	s.clearInvites(c.name)    // entering the queue drops any pending invites
	s.stopSpectating(c)       // queuing leaves any match being spectated
	deck := s.deckFor(c.name, fm.DeckID)
	c.match.Store(nil)
	s.mu.Lock()
	delete(s.active, c.name) // starting fresh: drop any prior (finished) seat
	s.mu.Unlock()
	if fm.VsAI {
		s.startVsAI(c, deck, fm.AIClass, fm.AIDeckID)
		s.broadcastPresence()
		return
	}
	if m := s.lobby.Join(c, deck); m != nil {
		// Both clients learn the match via MatchStart broadcast in m.Start();
		// record the ref on both so end_turn/play_card can find it.
		s.mu.Lock()
		delete(s.queued, c.name) // matched immediately: no longer waiting
		s.mu.Unlock()
		m.SetRanked(true) // matchmaking-queue games count toward the ladder
		s.attachMatch(m)
	} else {
		s.mu.Lock()
		s.queued[c.name] = struct{}{} // waiting in the queue
		s.mu.Unlock()
		c.Send(protocol.Marshal(protocol.Waiting{Type: protocol.TypeWaiting}))
	}
	s.broadcastPresence()
}

// handleEnterLobby returns the client to the lobby: drops it from the queue,
// abandons any finished match, and refreshes presence counts for everyone.
func (s *Server) handleEnterLobby(c *Client) {
	s.lobby.Remove(c)
	s.stopSpectating(c)
	c.match.Store(nil)
	s.mu.Lock()
	delete(s.active, c.name) // left the (finished) match; not reconnectable
	delete(s.queued, c.name)
	s.mu.Unlock()
	s.clearInvites(c.name)
	s.broadcastPresence()
}

// handleSpectate starts c watching another player's live match from that
// player's point of view. Rejected if c is in its own match, the target is
// offline, or the target is not currently in a match. SpectateStart is sent
// before AddObserver's snapshot so the client enters its read-only view first.
func (s *Server) handleSpectate(data []byte, c *Client) {
	if c.name == "" {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "not authenticated"}))
		return
	}
	var sp protocol.Spectate
	if err := json.Unmarshal(data, &sp); err != nil || sp.Target == "" {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "bad spectate"}))
		return
	}
	if sp.Target == c.name {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "cannot spectate yourself"}))
		return
	}
	if cm := c.match.Load(); cm != nil && !cm.Over() {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "leave your match first"}))
		return
	}
	tc := s.byName(sp.Target)
	if tc == nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "player is offline"}))
		return
	}
	m := tc.match.Load()
	if m == nil || m.Over() {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "player is not in a match"}))
		return
	}
	seat := m.SeatOf(tc)
	if seat < 0 {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "player is not in a match"}))
		return
	}
	s.stopSpectating(c) // drop any match already being watched
	// SpectateStart first so the client switches to its spectator view before the
	// AddObserver snapshot arrives (both go through the ordered send channel). The
	// client only renders once the snapshot lands, so a failed AddObserver leaves
	// it in the lobby.
	c.Send(protocol.Marshal(protocol.SpectateStart{Type: protocol.TypeSpectateStart, Target: sp.Target}))
	if !m.AddObserver(c, seat) {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "player is not in a match"}))
		return
	}
	c.spectating.Store(m)
}

// handleIntent relays an acting player's ephemeral aiming hint (hovered hand card /
// inspected minion / in-progress aim) to their opponent and that opponent's spectators.
// Non-authoritative and best-effort: dropped silently if the sender isn't in a live
// match or the payload is malformed. Spectators (no seat) are no-op'd inside RelayIntent.
func (s *Server) handleIntent(data []byte, c *Client) {
	m := c.match.Load()
	if m == nil || m.Over() {
		return
	}
	var in protocol.Intent
	if err := json.Unmarshal(data, &in); err != nil {
		return
	}
	m.RelayIntent(c, protocol.OppIntent{
		Type:      protocol.TypeOppIntent,
		HoverHand: in.HoverHand,
		Hover:     in.Hover,
		AimFrom:   in.AimFrom,
		AimTo:     in.AimTo,
	})
}

// stopSpectating deregisters c as an observer of any match it is watching.
func (s *Server) stopSpectating(c *Client) {
	if m := c.spectating.Load(); m != nil {
		m.RemoveObserver(c)
		c.spectating.Store(nil)
	}
}

// byNameLocked returns one authenticated client with the given username (the
// single-session kick keeps it to one). Caller holds s.mu.
func (s *Server) byNameLocked(name string) *Client {
	for cl := range s.clients {
		if cl.name == name {
			return cl
		}
	}
	return nil
}

// availableLocked reports whether cl is a lobby player free to be matched:
// online, not in a live match, not queued, no pending invite of their own.
// Caller holds s.mu.
func (s *Server) availableLocked(cl *Client) bool {
	if cl == nil || cl.name == "" {
		return false
	}
	if _, q := s.queued[cl.name]; q {
		return false
	}
	if m := cl.match.Load(); m != nil && !m.Over() {
		return false
	}
	return true
}

// handleInvite records a direct invite from c to another lobby player and
// prompts that player. Only one outstanding invite per inviter (a second is
// rejected — the client must cancel first).
func (s *Server) handleInvite(data []byte, c *Client) {
	if c.name == "" {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "not authenticated"}))
		return
	}
	var in protocol.Invite
	if err := json.Unmarshal(data, &in); err != nil || in.Target == "" {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "bad invite"}))
		return
	}
	if in.Target == c.name {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "cannot invite yourself"}))
		return
	}
	if cm := c.match.Load(); cm != nil && !cm.Over() {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "already in a match"}))
		return
	}
	s.mu.Lock()
	if _, ok := s.invites[c.name]; ok {
		s.mu.Unlock()
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "cancel your current invite first"}))
		return
	}
	if _, q := s.queued[c.name]; q {
		s.mu.Unlock()
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "leave the queue first"}))
		return
	}
	tc := s.byNameLocked(in.Target)
	if !s.availableLocked(tc) {
		s.mu.Unlock()
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "player is not available"}))
		return
	}
	s.invites[c.name] = inviteRec{target: in.Target, deckID: in.DeckID}
	s.mu.Unlock()
	tc.Send(protocol.Marshal(protocol.InviteReceived{Type: protocol.TypeInviteReceived, From: c.name}))
}

// handleInviteCancel withdraws c's outstanding invite and tells the invitee.
func (s *Server) handleInviteCancel(c *Client) {
	s.mu.Lock()
	rec, ok := s.invites[c.name]
	if ok {
		delete(s.invites, c.name)
	}
	s.mu.Unlock()
	if ok {
		if tc := s.byName(rec.target); tc != nil {
			tc.Send(protocol.Marshal(protocol.InviteCancelled{Type: protocol.TypeInviteCancelled, From: c.name}))
		}
	}
}

// handleInviteRespond answers an invite from r.From. Decline notifies the
// inviter; accept starts the match directly (inviter acts first), each side
// playing the deck they chose.
func (s *Server) handleInviteRespond(data []byte, c *Client) {
	if c.name == "" {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "not authenticated"}))
		return
	}
	var r protocol.InviteRespond
	if err := json.Unmarshal(data, &r); err != nil || r.From == "" {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "bad invite_respond"}))
		return
	}
	s.mu.Lock()
	rec, ok := s.invites[r.From]
	if !ok || rec.target != c.name {
		s.mu.Unlock()
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "invite no longer valid"}))
		return
	}
	delete(s.invites, r.From)
	ic := s.byNameLocked(r.From)
	s.mu.Unlock()

	if !r.Accept {
		if ic != nil {
			ic.Send(protocol.Marshal(protocol.InviteDeclined{Type: protocol.TypeInviteDeclined, By: c.name}))
		}
		return
	}

	// Accept: both sides must still be free.
	s.mu.Lock()
	ok = s.availableLocked(ic) && s.availableLocked(c)
	s.mu.Unlock()
	if !ok {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "player is no longer available"}))
		if ic != nil {
			ic.Send(protocol.Marshal(protocol.InviteDeclined{Type: protocol.TypeInviteDeclined, By: c.name}))
		}
		return
	}

	deckA := s.deckFor(ic.name, rec.deckID)
	deckB := s.deckFor(c.name, r.DeckID)
	// Drop any other pending invites for either player, plus stale seats/queue.
	s.clearInvites(ic.name)
	s.clearInvites(c.name)
	s.mu.Lock()
	delete(s.active, ic.name)
	delete(s.active, c.name)
	delete(s.queued, ic.name)
	delete(s.queued, c.name)
	s.mu.Unlock()

	m := s.lobby.StartMatch(ic, deckA, c, deckB)
	s.attachMatch(m)
	s.broadcastPresence()
}

// clearInvites drops every invite that name is part of and notifies the other
// side: an invite name sent is cancelled to its target; invites aimed at name
// are declined to their inviters (name is no longer available).
func (s *Server) clearInvites(name string) {
	s.mu.Lock()
	var cancelTo string    // target to tell the invite from name is gone
	var declineTo []string // inviters to tell name declined / left
	if rec, ok := s.invites[name]; ok {
		delete(s.invites, name)
		cancelTo = rec.target
	}
	for inv, rec := range s.invites {
		if rec.target == name {
			delete(s.invites, inv)
			declineTo = append(declineTo, inv)
		}
	}
	s.mu.Unlock()
	if cancelTo != "" {
		if tc := s.byName(cancelTo); tc != nil {
			tc.Send(protocol.Marshal(protocol.InviteCancelled{Type: protocol.TypeInviteCancelled, From: name}))
		}
	}
	for _, inv := range declineTo {
		if ic := s.byName(inv); ic != nil {
			ic.Send(protocol.Marshal(protocol.InviteDeclined{Type: protocol.TypeInviteDeclined, By: name}))
		}
	}
}

// byName returns one authenticated client with the given username, locking
// internally (use when s.mu is not already held).
func (s *Server) byName(name string) *Client {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.byNameLocked(name)
}

// attachMatch records the match ref on both players so end_turn can find it,
// records each player's seat (match + slot id) so a dropped player can reconnect,
// stamps each seat's current ladder rank for the in-game nameplate, and wires the
// end-of-game hook that persists a ranked result.
func (s *Server) attachMatch(m *match.Match) {
	s.mu.Lock()
	for _, p := range m.Players() {
		if cl, ok := p.(*Client); ok {
			cl.match.Store(m)
			s.active[cl.name] = activeSeat{m: m, slot: cl.id}
		}
	}
	s.mu.Unlock()

	// Ladder rank per seat (0 for the AI / a player with no ranked games yet).
	players := m.Players()
	rankOf := func(p match.Sender) int {
		if _, ok := p.(*Client); ok {
			return s.store.Rank(p.Name())
		}
		return 0
	}
	oldRanks := [2]int{rankOf(players[0]), rankOf(players[1])}
	m.SetRanks(oldRanks[0], oldRanks[1])

	// Persist the result when a hero dies. ranked/players/oldRanks are captured here
	// (all known at match start) so the hook touches no match-locked state and can
	// run synchronously inside the match — see Match.OnEnd. Unranked games (vs-AI,
	// invites) are a no-op.
	ranked := m.Ranked()
	m.OnEnd(func(winnerSeat int) {
		if !ranked {
			return
		}
		ws, ls := winnerSeat, 1-winnerSeat
		if err := s.store.RecordResult(
			players[ws].Name(), players[ls].Name(),
			string(m.SeatClass(ws)), string(m.SeatClass(ls)),
		); err != nil {
			log.Printf("record result %s beat %s: %v", players[ws].Name(), players[ls].Name(), err)
			return
		}
		// Tell each player their new ladder position vs. where they started, for the
		// win/loss screen's rank-change indicator. AI sends are no-ops.
		for seat := 0; seat < 2; seat++ {
			players[seat].Send(protocol.Marshal(protocol.RankUpdate{
				Type:    protocol.TypeRankUpdate,
				OldRank: oldRanks[seat],
				NewRank: s.store.Rank(players[seat].Name()),
			}))
		}
	})
}

// clearActive drops every seat entry pointing at m (both players), called when
// the match ends so finished matches are not reconnectable.
func (s *Server) clearActive(m *match.Match) {
	s.mu.Lock()
	for name, seat := range s.active {
		if seat.m == m {
			delete(s.active, name)
		}
	}
	s.mu.Unlock()
}

// startGrace holds c's seat open for graceWindow, then forfeits the match on
// their behalf if they have not reconnected (the seat still belongs to c). The
// timer is keyed by username so a reconnect can cancel it.
func (s *Server) startGrace(c *Client, m *match.Match) {
	name := c.name
	s.mu.Lock()
	if old := s.grace[name]; old != nil {
		old.Stop()
	}
	s.grace[name] = time.AfterFunc(s.graceWindow, func() {
		s.mu.Lock()
		delete(s.grace, name)
		seat, ok := s.active[name]
		s.mu.Unlock()
		// Skip if the player reconnected (seat swapped to a new client),
		// re-queued, or the match already ended some other way.
		if !ok || seat.m != m || m.Over() || !m.Seats(c) {
			return
		}
		m.Concede(c) // opponent wins; game_over is broadcast by the match
		s.clearActive(m)
		s.broadcastPresence()
	})
	s.mu.Unlock()
}

// notifyOpp tells c's opponent in match m whether c's connection is now live.
func (s *Server) notifyOpp(m *match.Match, c *Client, connected bool) {
	msg := protocol.Marshal(protocol.OppConn{Type: protocol.TypeOppConn, Connected: connected})
	for _, p := range m.Players() {
		if cl, ok := p.(*Client); ok && cl != c {
			cl.Send(msg)
		}
	}
}

// register adds an authenticated client to the presence set and broadcasts the
// updated counts.
func (s *Server) register(c *Client) {
	s.mu.Lock()
	s.clients[c] = struct{}{}
	s.mu.Unlock()
	s.broadcastPresence()
}

// deregister removes a client (on disconnect) and broadcasts updated counts.
func (s *Server) deregister(c *Client) {
	s.mu.Lock()
	_, was := s.clients[c]
	delete(s.clients, c)
	s.mu.Unlock()
	if was {
		s.broadcastPresence()
	}
}

// broadcastPresence sends current online / in-game counts to every connected
// client. Counts are by distinct username, not by connection, so a reload (which
// can briefly leave a stale connection until its close is detected) or a second
// tab for the same account does not inflate the numbers.
func (s *Server) broadcastPresence() {
	s.mu.Lock()
	// One representative client per distinct username (single-session kick keeps
	// it to one live connection per account anyway).
	byName := make(map[string]*Client, len(s.clients))
	recipients := make([]*Client, 0, len(s.clients))
	for cl := range s.clients {
		recipients = append(recipients, cl)
		if _, ok := byName[cl.name]; !ok {
			byName[cl.name] = cl
		}
	}
	queued := make(map[string]struct{}, len(s.queued))
	for n := range s.queued {
		queued[n] = struct{}{}
	}
	s.mu.Unlock()

	players := make([]protocol.PlayerInfo, 0, len(byName))
	inGame := 0
	for name, cl := range byName {
		info := protocol.PlayerInfo{Name: name, Status: "lobby"}
		if m := cl.match.Load(); m != nil && !m.Over() {
			info.Status = "in_game"
			info.MatchID = m.ID
			for _, p := range m.Players() {
				if p != nil && p.Name() != name {
					info.Vs = p.Name()
				}
			}
			inGame++
		} else if _, ok := queued[name]; ok {
			info.Status = "waiting"
		}
		players = append(players, info)
	}
	sort.Slice(players, func(i, j int) bool { return players[i].Name < players[j].Name })

	msg := protocol.Marshal(protocol.Lobby{
		Type: protocol.TypeLobby, Online: len(byName), InGame: inGame, Players: players,
	})
	for _, cl := range recipients {
		cl.Send(msg)
	}
}

func (s *Server) handleEndTurn(c *Client) {
	m := c.match.Load()
	if m == nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "not in a match"}))
		return
	}
	if ok, msg := m.EndTurn(c); !ok {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: msg}))
	}
}

func (s *Server) handlePlayCard(data []byte, c *Client) {
	m := c.match.Load()
	if m == nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "not in a match"}))
		return
	}
	var p protocol.PlayCard
	if err := json.Unmarshal(data, &p); err != nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "bad play_card"}))
		return
	}
	pos := -1
	if p.Pos != nil {
		pos = *p.Pos
	}
	if ok, msg := m.PlayCardAt(c, p.HandIndex, p.TargetID, pos); !ok {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: msg}))
	}
}

func (s *Server) handleAttack(data []byte, c *Client) {
	m := c.match.Load()
	if m == nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "not in a match"}))
		return
	}
	var a protocol.Attack
	if err := json.Unmarshal(data, &a); err != nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "bad attack"}))
		return
	}
	if ok, msg := m.Attack(c, a.AttackerID, a.TargetID); !ok {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: msg}))
	}
}

func (s *Server) handleConcede(c *Client) {
	m := c.match.Load()
	if m == nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "not in a match"}))
		return
	}
	if ok, msg := m.Concede(c); !ok {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: msg}))
	}
}

func (s *Server) handleChoose(data []byte, c *Client) {
	m := c.match.Load()
	if m == nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "not in a match"}))
		return
	}
	var ch protocol.Choose
	if err := json.Unmarshal(data, &ch); err != nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "bad choose"}))
		return
	}
	if ok, msg := m.Choose(c, ch.Index); !ok {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: msg}))
	}
}

func (s *Server) handleHeroPower(data []byte, c *Client) {
	m := c.match.Load()
	if m == nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "not in a match"}))
		return
	}
	var hp protocol.HeroPower
	if err := json.Unmarshal(data, &hp); err != nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "bad hero_power"}))
		return
	}
	if ok, msg := m.HeroPower(c, hp.TargetID); !ok {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: msg}))
	}
}

func (s *Server) handleMulligan(data []byte, c *Client) {
	m := c.match.Load()
	if m == nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "not in a match"}))
		return
	}
	var mu protocol.Mulligan
	if err := json.Unmarshal(data, &mu); err != nil {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: "bad mulligan"}))
		return
	}
	if ok, msg := m.Mulligan(c, mu.Indices); !ok {
		c.Send(protocol.Marshal(protocol.Error{Type: protocol.TypeError, Msg: msg}))
	}
}

// startVsAI matches c immediately against an AI opponent. When aiDeckID names one
// of the player's own saved decks the bot plays that; otherwise it plays a random
// prebuilt deck of the requested class (falling back to a playable class when the
// class is unknown / has no decks). The human plays the deck they queued with.
func (s *Server) startVsAI(c *Client, humanDeck []cards.Card, aiClass string, aiDeckID int64) {
	class := cards.Class(aiClass)
	decks := cards.AIDecks(class)
	if len(decks) == 0 {
		class = cards.PlayableClasses()[0]
		decks = cards.AIDecks(class)
	}
	botDeck := cards.Deck(decks[rand.Intn(len(decks))])
	if aiDeckID != 0 {
		botDeck = s.deckFor(c.name, aiDeckID) // bot plays one of the player's saved decks
	}
	m := s.lobby.StartVsAI(c, humanDeck, botDeck, aiName(class))
	s.attachMatch(m)
}

// aiName is the AI opponent's display name for a class (e.g. "AI Mage").
func aiName(class cards.Class) string {
	s := string(class)
	if s == "" {
		return "AI Opponent"
	}
	return "AI " + strings.ToUpper(s[:1]) + s[1:]
}

// deckFor returns the deck cards a player will play: their saved deck deckID if
// it exists and is legal. When deckID is 0, missing, or fails validation it
// falls back to the player's first saved deck (every account is seeded with a
// starter deck at registration). A curated default is the last-resort safety net
// for any legacy account with no decks, so queuing never fails.
func (s *Server) deckFor(username string, deckID int64) []cards.Card {
	if deckID != 0 {
		if d, err := s.store.GetDeck(username, deckID); err == nil && cards.ValidateDeck(d.Cards, cards.Class(d.Class)) == nil {
			return cards.Deck(d.Cards)
		}
	}
	if decks, err := s.store.ListDecks(username); err == nil {
		for _, d := range decks {
			if cards.ValidateDeck(d.Cards, cards.Class(d.Class)) == nil {
				return cards.Deck(d.Cards)
			}
		}
	}
	return cards.Deck(cards.DefaultDeck())
}
