package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"path/filepath"

	"github.com/amvid/vanillastone/internal/auth"
	"github.com/amvid/vanillastone/internal/protocol"
	"github.com/amvid/vanillastone/internal/store"
	"github.com/coder/websocket"
)

// dial opens a ws client against the test server's /ws.
func dial(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	c, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { c.Close(websocket.StatusNormalClosure, "") })
	return c
}

func send(t *testing.T, c *websocket.Conn, v any) {
	t.Helper()
	if err := c.Write(context.Background(), websocket.MessageText, protocol.Marshal(v)); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// recv reads one message and returns its decoded type + raw bytes.
func recv(t *testing.T, c *websocket.Conn) (string, []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var env protocol.Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return env.Type, data
}

// expect reads until it sees a message of the wanted type, returning its bytes.
// Skips other types (e.g. "lobby" presence updates or a "waiting" before
// "match_start").
func expect(t *testing.T, c *websocket.Conn, want string) []byte {
	t.Helper()
	for i := 0; i < 20; i++ {
		typ, data := recv(t, c)
		if typ == want {
			return data
		}
	}
	t.Fatalf("did not receive %q", want)
	return nil
}

// waitPlay reads state messages until one signals play has begun (Mulligan
// false), so the test does not race the mulligan -> play transition.
func waitPlay(t *testing.T, c *websocket.Conn) {
	t.Helper()
	for i := 0; i < 20; i++ {
		typ, data := recv(t, c)
		if typ != protocol.TypeState {
			continue
		}
		var st protocol.State
		json.Unmarshal(data, &st)
		if !st.Mulligan {
			return
		}
	}
	t.Fatalf("play did not begin after mulligan")
}

func newServer(t *testing.T) (*httptest.Server, *auth.Auth, *Server) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	au := auth.New(st)
	ts := NewServer(au, st)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ts.HandleWS)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, au, ts
}

// account registers a user and logs in, returning a session token.
func account(t *testing.T, au *auth.Auth, name string) string {
	t.Helper()
	if err := au.Register(name, "password123"); err != nil {
		t.Fatalf("register %s: %v", name, err)
	}
	token, err := au.Login(name, "password123")
	if err != nil {
		t.Fatalf("login %s: %v", name, err)
	}
	return token
}

// TestMatchTurnPingPong verifies the core Phase 1 intent: the server is the sole
// authority on turn order. Two players are matched; the turn alternates only on
// the current player's end_turn, and an off-turn end_turn is rejected without
// changing whose turn it is.
func TestMatchTurnPingPong(t *testing.T) {
	srv, au, _ := newServer(t)
	aliceTok := account(t, au, "alice")
	bobTok := account(t, au, "bob")

	a := dial(t, srv)
	send(t, a, protocol.Auth{Type: protocol.TypeAuth, Token: aliceTok})
	var aJoined protocol.Joined
	json.Unmarshal(expect(t, a, protocol.TypeJoined), &aJoined)
	// Auth lands in the lobby; queue explicitly. a is first -> acts first.
	send(t, a, protocol.FindMatch{Type: protocol.TypeFindMatch})

	b := dial(t, srv)
	send(t, b, protocol.Auth{Type: protocol.TypeAuth, Token: bobTok})
	var bJoined protocol.Joined
	json.Unmarshal(expect(t, b, protocol.TypeJoined), &bJoined)
	send(t, b, protocol.FindMatch{Type: protocol.TypeFindMatch})

	// Both should now see match_start. The first to queue (alice) acts first.
	var msA protocol.MatchStart
	json.Unmarshal(expect(t, a, protocol.TypeMatchStart), &msA)
	var msB protocol.MatchStart
	json.Unmarshal(expect(t, b, protocol.TypeMatchStart), &msB)

	if msA.Turn != aJoined.You {
		t.Fatalf("first queued player should act first: turn=%s alice=%s", msA.Turn, aJoined.You)
	}
	if msA.Turn != msB.Turn {
		t.Fatalf("both clients must agree on whose turn: a=%s b=%s", msA.Turn, msB.Turn)
	}
	if !msA.Mulligan {
		t.Fatalf("match should open in the mulligan phase")
	}

	// Complete the opening mulligan (keep all cards) for both players, then wait
	// until each sees play begin (a state with Mulligan=false).
	send(t, a, protocol.Mulligan{Type: protocol.TypeMulligan})
	send(t, b, protocol.Mulligan{Type: protocol.TypeMulligan})
	waitPlay(t, a)
	waitPlay(t, b)

	// Off-turn player (bob) tries to end turn -> rejected, no state change.
	send(t, b, protocol.EndTurn{Type: protocol.TypeEndTurn})
	var errMsg protocol.Error
	json.Unmarshal(expect(t, b, protocol.TypeError), &errMsg)
	if errMsg.Msg != "not your turn" {
		t.Fatalf("expected 'not your turn', got %q", errMsg.Msg)
	}

	// Current player (alice) ends turn -> both get state, turn flips to bob.
	send(t, a, protocol.EndTurn{Type: protocol.TypeEndTurn})
	var stA protocol.State
	json.Unmarshal(expect(t, a, protocol.TypeState), &stA)
	var stB protocol.State
	json.Unmarshal(expect(t, b, protocol.TypeState), &stB)
	if stA.Turn != bJoined.You || stB.Turn != bJoined.You {
		t.Fatalf("turn should flip to bob: a=%s b=%s bob=%s", stA.Turn, stB.Turn, bJoined.You)
	}
	if stA.TurnNum != 1 {
		t.Fatalf("turnNum should be 1 after first end_turn, got %d", stA.TurnNum)
	}

	// Now bob ends turn -> flips back to alice, turnNum 2.
	send(t, b, protocol.EndTurn{Type: protocol.TypeEndTurn})
	var st2 protocol.State
	json.Unmarshal(expect(t, b, protocol.TypeState), &st2)
	if st2.Turn != aJoined.You || st2.TurnNum != 2 {
		t.Fatalf("turn should return to alice at turnNum 2: turn=%s num=%d", st2.Turn, st2.TurnNum)
	}
}

// TestSingleSessionKicksPrevious verifies that logging in as an account that is
// already connected reaps the earlier connection: the first client is told it
// was logged in elsewhere. This is the intent — one live session per account.
func TestSingleSessionKicksPrevious(t *testing.T) {
	srv, au, _ := newServer(t)
	tok := account(t, au, "carol")

	a := dial(t, srv)
	send(t, a, protocol.Auth{Type: protocol.TypeAuth, Token: tok})
	expect(t, a, protocol.TypeJoined)

	// Same account, second connection — the first should be kicked.
	b := dial(t, srv)
	send(t, b, protocol.Auth{Type: protocol.TypeAuth, Token: tok})
	expect(t, b, protocol.TypeJoined)

	var e protocol.Error
	json.Unmarshal(expect(t, a, protocol.TypeError), &e)
	if e.Msg != "logged in elsewhere" {
		t.Fatalf("first session should be kicked with 'logged in elsewhere', got %q", e.Msg)
	}
}

// TestReconnectResumesMatch verifies the Phase 10 intent: a player who drops
// mid-match rejoins the same match (re-adopting their seat) instead of
// forfeiting, the opponent is told the player left and returned, and the match
// continues authoritatively from where it stood.
func TestReconnectResumesMatch(t *testing.T) {
	srv, au, _ := newServer(t)
	aliceTok := account(t, au, "alice")
	bobTok := account(t, au, "bob")

	a := dial(t, srv)
	send(t, a, protocol.Auth{Type: protocol.TypeAuth, Token: aliceTok})
	var aJoined protocol.Joined
	json.Unmarshal(expect(t, a, protocol.TypeJoined), &aJoined)
	send(t, a, protocol.FindMatch{Type: protocol.TypeFindMatch})

	b := dial(t, srv)
	send(t, b, protocol.Auth{Type: protocol.TypeAuth, Token: bobTok})
	var bJoined protocol.Joined
	json.Unmarshal(expect(t, b, protocol.TypeJoined), &bJoined)
	send(t, b, protocol.FindMatch{Type: protocol.TypeFindMatch})

	expect(t, a, protocol.TypeMatchStart)
	expect(t, b, protocol.TypeMatchStart)

	// Both keep their opening hands; alice (first) is to act once play begins.
	send(t, a, protocol.Mulligan{Type: protocol.TypeMulligan})
	send(t, b, protocol.Mulligan{Type: protocol.TypeMulligan})
	waitPlay(t, a)
	waitPlay(t, b)

	// Alice drops. The opponent is told her connection went away (grace window
	// begins; no forfeit yet).
	a.Close(websocket.StatusGoingAway, "")
	var oc protocol.OppConn
	json.Unmarshal(expect(t, b, protocol.TypeOppConn), &oc)
	if oc.Connected {
		t.Fatalf("opponent should be told alice disconnected (connected=false)")
	}

	// Alice reconnects with the same account: she swaps back into her seat with
	// the SAME player id (turn identity preserved), and gets a resync snapshot.
	a2 := dial(t, srv)
	send(t, a2, protocol.Auth{Type: protocol.TypeAuth, Token: aliceTok})
	var a2Joined protocol.Joined
	json.Unmarshal(expect(t, a2, protocol.TypeJoined), &a2Joined)
	if a2Joined.You != aJoined.You {
		t.Fatalf("reconnect must re-adopt the original seat id: was %s, got %s", aJoined.You, a2Joined.You)
	}
	var resync protocol.State
	json.Unmarshal(expect(t, a2, protocol.TypeState), &resync)
	if !resync.Resync {
		t.Fatalf("reconnect snapshot should be flagged resync")
	}
	if resync.Turn != aJoined.You {
		t.Fatalf("resync should show it is still alice's turn: turn=%s alice=%s", resync.Turn, aJoined.You)
	}

	// Bob is told alice is back.
	json.Unmarshal(expect(t, b, protocol.TypeOppConn), &oc)
	if !oc.Connected {
		t.Fatalf("opponent should be told alice reconnected (connected=true)")
	}

	// The match continues authoritatively: alice (reconnected) ends her turn and
	// the turn flips to bob — proving no forfeit happened and her seat is live.
	send(t, a2, protocol.EndTurn{Type: protocol.TypeEndTurn})
	var stB protocol.State
	json.Unmarshal(expect(t, b, protocol.TypeState), &stB)
	if stB.Turn != bJoined.You {
		t.Fatalf("turn should flip to bob after alice's post-reconnect end_turn: turn=%s bob=%s", stB.Turn, bJoined.You)
	}
}

// TestGraceForfeitsIfNoReconnect verifies the other side of the grace window: a
// player who drops and does NOT come back within the window forfeits, so the
// opponent wins. Without reconnect this is the old immediate-forfeit behavior,
// now merely delayed.
func TestGraceForfeitsIfNoReconnect(t *testing.T) {
	srv, au, ts := newServer(t)
	ts.graceWindow = 100 * time.Millisecond
	aliceTok := account(t, au, "alice")
	bobTok := account(t, au, "bob")

	a := dial(t, srv)
	send(t, a, protocol.Auth{Type: protocol.TypeAuth, Token: aliceTok})
	expect(t, a, protocol.TypeJoined)
	send(t, a, protocol.FindMatch{Type: protocol.TypeFindMatch})

	b := dial(t, srv)
	send(t, b, protocol.Auth{Type: protocol.TypeAuth, Token: bobTok})
	var bJoined protocol.Joined
	json.Unmarshal(expect(t, b, protocol.TypeJoined), &bJoined)
	send(t, b, protocol.FindMatch{Type: protocol.TypeFindMatch})

	expect(t, a, protocol.TypeMatchStart)
	expect(t, b, protocol.TypeMatchStart)
	send(t, a, protocol.Mulligan{Type: protocol.TypeMulligan})
	send(t, b, protocol.Mulligan{Type: protocol.TypeMulligan})
	waitPlay(t, a)
	waitPlay(t, b)

	// Alice drops and never returns: after the (shortened) grace window, the
	// match forfeits and bob is declared the winner.
	a.Close(websocket.StatusGoingAway, "")
	var over protocol.GameOver
	json.Unmarshal(expect(t, b, protocol.TypeGameOver), &over)
	if over.Winner != bJoined.You {
		t.Fatalf("bob should win when alice fails to reconnect: winner=%s bob=%s", over.Winner, bJoined.You)
	}
}

// TestAuthRejectsBadToken verifies an unauthenticated connection cannot join:
// a bogus token gets an error, not a Joined.
func TestAuthRejectsBadToken(t *testing.T) {
	srv, _, _ := newServer(t)
	c := dial(t, srv)
	send(t, c, protocol.Auth{Type: protocol.TypeAuth, Token: "not-a-real-token"})
	var e protocol.Error
	json.Unmarshal(expect(t, c, protocol.TypeError), &e)
	if e.Msg != "invalid or expired token" {
		t.Fatalf("expected invalid token error, got %q", e.Msg)
	}
}

// authJoin dials, authenticates, and returns the connection plus the player's
// Joined id. Drains the initial lobby presence message is left to the caller via
// expect (it skips non-matching types).
func authJoin(t *testing.T, srv *httptest.Server, tok string) (*websocket.Conn, string) {
	t.Helper()
	c := dial(t, srv)
	send(t, c, protocol.Auth{Type: protocol.TypeAuth, Token: tok})
	var j protocol.Joined
	json.Unmarshal(expect(t, c, protocol.TypeJoined), &j)
	return c, j.You
}

// TestInviteAcceptStartsMatch verifies the intent of direct invites: a lobby
// player challenges another, the target is prompted, and accepting starts a
// match between exactly those two (the inviter acting first) — no global queue.
func TestInviteAcceptStartsMatch(t *testing.T) {
	srv, au, _ := newServer(t)
	aTok := account(t, au, "alice")
	bTok := account(t, au, "bob")

	a, aID := authJoin(t, srv, aTok)
	b, bID := authJoin(t, srv, bTok)

	// alice invites bob; bob is prompted with alice as the challenger.
	send(t, a, protocol.Invite{Type: protocol.TypeInvite, Target: "bob"})
	var got protocol.InviteReceived
	json.Unmarshal(expect(t, b, protocol.TypeInviteReceived), &got)
	if got.From != "alice" {
		t.Fatalf("invite should come from alice, got %q", got.From)
	}

	// bob accepts -> both enter the mulligan; alice (the inviter) acts first.
	send(t, b, protocol.InviteRespond{Type: protocol.TypeInviteRespond, From: "alice", Accept: true})
	var msA, msB protocol.MatchStart
	json.Unmarshal(expect(t, a, protocol.TypeMatchStart), &msA)
	json.Unmarshal(expect(t, b, protocol.TypeMatchStart), &msB)
	if msA.Turn != aID {
		t.Fatalf("inviter (alice) should act first: turn=%s alice=%s", msA.Turn, aID)
	}
	if msA.Turn != msB.Turn {
		t.Fatalf("both clients must agree on whose turn: a=%s b=%s", msA.Turn, msB.Turn)
	}
	_ = bID
}

// TestInviteDeclined verifies a refused invite notifies the inviter and starts
// no match.
func TestInviteDeclined(t *testing.T) {
	srv, au, _ := newServer(t)
	aTok := account(t, au, "alice")
	bTok := account(t, au, "bob")

	a, _ := authJoin(t, srv, aTok)
	b, _ := authJoin(t, srv, bTok)

	send(t, a, protocol.Invite{Type: protocol.TypeInvite, Target: "bob"})
	expect(t, b, protocol.TypeInviteReceived)

	send(t, b, protocol.InviteRespond{Type: protocol.TypeInviteRespond, From: "alice", Accept: false})
	var dec protocol.InviteDeclined
	json.Unmarshal(expect(t, a, protocol.TypeInviteDeclined), &dec)
	if dec.By != "bob" {
		t.Fatalf("decline should be by bob, got %q", dec.By)
	}
}

// TestInviteOneOutstanding verifies the rule that a player may hold only one
// outstanding invite: a second invite is rejected until the first is cancelled.
func TestInviteOneOutstanding(t *testing.T) {
	srv, au, _ := newServer(t)
	aTok := account(t, au, "alice")
	account(t, au, "bob")
	cTok := account(t, au, "carol")

	a, _ := authJoin(t, srv, aTok)
	c, _ := authJoin(t, srv, cTok)
	_ = c

	send(t, a, protocol.Invite{Type: protocol.TypeInvite, Target: "carol"})
	expect(t, c, protocol.TypeInviteReceived)

	// A second invite while one is outstanding is rejected.
	send(t, a, protocol.Invite{Type: protocol.TypeInvite, Target: "bob"})
	var e protocol.Error
	json.Unmarshal(expect(t, a, protocol.TypeError), &e)
	if e.Msg != "cancel your current invite first" {
		t.Fatalf("second invite should be rejected, got %q", e.Msg)
	}

	// Cancelling notifies the invitee and frees a new invite.
	send(t, a, protocol.InviteCancel{Type: protocol.TypeInviteCancel})
	var cancelled protocol.InviteCancelled
	json.Unmarshal(expect(t, c, protocol.TypeInviteCancelled), &cancelled)
	if cancelled.From != "alice" {
		t.Fatalf("cancel should be from alice, got %q", cancelled.From)
	}
}

// TestSpectateMirrorsPlayerPOV verifies the intent of spectator mode: a watcher
// receives the watched player's exact point of view — that player's hand revealed,
// the opponent's hidden — follows the live game, and is told when it ends. The
// watched player's hidden information must NOT leak from the opponent's side.
func TestSpectateMirrorsPlayerPOV(t *testing.T) {
	srv, au, _ := newServer(t)
	aTok := account(t, au, "alice")
	bTok := account(t, au, "bob")
	cTok := account(t, au, "carol")

	a, aID := authJoin(t, srv, aTok)
	b, bID := authJoin(t, srv, bTok)

	// alice invites bob, bob accepts -> alice acts first. Both mulligan to begin play.
	send(t, a, protocol.Invite{Type: protocol.TypeInvite, Target: "bob"})
	expect(t, b, protocol.TypeInviteReceived)
	send(t, b, protocol.InviteRespond{Type: protocol.TypeInviteRespond, From: "alice", Accept: true})
	expect(t, a, protocol.TypeMatchStart)
	expect(t, b, protocol.TypeMatchStart)
	send(t, a, protocol.Mulligan{Type: protocol.TypeMulligan})
	send(t, b, protocol.Mulligan{Type: protocol.TypeMulligan})
	waitPlay(t, a)
	waitPlay(t, b)

	// carol logs in and spectates alice. She first gets spectate_start, then a
	// snapshot from alice's POV.
	c, _ := authJoin(t, srv, cTok)
	send(t, c, protocol.Spectate{Type: protocol.TypeSpectate, Target: "alice"})
	var ss protocol.SpectateStart
	json.Unmarshal(expect(t, c, protocol.TypeSpectateStart), &ss)
	if ss.Target != "alice" {
		t.Fatalf("spectate_start target = %q, want alice", ss.Target)
	}
	var st protocol.State
	json.Unmarshal(expect(t, c, protocol.TypeState), &st)
	if st.You != aID {
		t.Fatalf("spectator POV You = %q, want alice's id %q", st.You, aID)
	}
	// Alice's hand is revealed (Self), bob's is hidden (Opp: count only, no cards).
	if len(st.Self.Hand) == 0 {
		t.Fatalf("watched player's hand should be revealed to the spectator")
	}
	if len(st.Opp.Hand) != 0 {
		t.Fatalf("opponent's hand must stay hidden from the spectator, got %d cards", len(st.Opp.Hand))
	}
	if st.Opp.HandCount == 0 {
		t.Fatalf("opponent's hand count should still be visible")
	}

	// The watched player is told who is watching.
	var watchers protocol.Spectators
	json.Unmarshal(expect(t, a, protocol.TypeSpectators), &watchers)
	if len(watchers.Names) != 1 || watchers.Names[0] != "carol" {
		t.Fatalf("alice should see carol watching, got %v", watchers.Names)
	}

	// The spectator follows the live game: alice ends her turn, the snapshot the
	// spectator receives shows the turn passing to bob.
	send(t, a, protocol.EndTurn{Type: protocol.TypeEndTurn})
	for i := 0; i < 20; i++ {
		_, data := recv(t, c)
		var s protocol.State
		if json.Unmarshal(data, &s); s.Type == protocol.TypeState && s.Turn == bID {
			break
		}
		if i == 19 {
			t.Fatalf("spectator did not observe the turn pass to bob")
		}
	}

	// When the match ends, the spectator is notified too.
	send(t, a, protocol.Concede{Type: protocol.TypeConcede})
	var over protocol.GameOver
	json.Unmarshal(expect(t, c, protocol.TypeGameOver), &over)
	if over.Winner != bID {
		t.Fatalf("game_over winner = %q, want bob %q", over.Winner, bID)
	}
}
