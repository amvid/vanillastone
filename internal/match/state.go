package match

import (
	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/protocol"
	"sort"
)

// Players returns both participants (read-only; order is fixed at creation).
func (m *Match) Players() [2]Sender {
	return m.players
}

// SetRanked marks the match competitive: its result is persisted when it ends.
// Only matchmaking-queue games are ranked (not invites or vs-AI).
func (m *Match) SetRanked(v bool) {
	m.mu.Lock()
	m.ranked = v
	m.mu.Unlock()
}

// SetRanks records each seat's ladder position (0 = unranked / AI) for the
// in-game nameplate. Call before play begins; the value is a match-start
// snapshot and does not update mid-game.
func (m *Match) SetRanks(r0, r1 int) {
	m.mu.Lock()
	m.rank = [2]int{r0, r1}
	m.mu.Unlock()
}

// OnEnd registers a one-shot callback invoked with the winning seat index when a
// hero dies. The transport uses it to persist a ranked result; the callback must
// not block (it runs while the match lock is held).
func (m *Match) OnEnd(fn func(winnerSeat int)) {
	m.mu.Lock()
	m.onEnd = fn
	m.mu.Unlock()
}

// Ranked reports whether this match's result should be persisted.
func (m *Match) Ranked() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ranked
}

// SeatClass returns the deck class played at seat i (immutable after creation).
func (m *Match) SeatClass(i int) cards.Class { return m.class[i] }

// Over reports whether the match has finished (a hero died). Used by the
// transport layer to count players still in active games.
func (m *Match) Over() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.over
}

// Seats reports whether c currently occupies one of the two player slots, by
// pointer identity. The transport uses it on a disconnect to tell whether the
// dropping connection is still the live occupant of its seat: a client that has
// already been swapped out by a reconnecting takeover (see Reattach) no longer
// seats, so its drop must not start a forfeit timer.
func (m *Match) Seats(c Sender) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.players[0] == c || m.players[1] == c
}

// Reattach swaps a reconnecting player back into their existing seat. c must
// carry the slot's player id (the transport re-adopts it before calling), so
// indexOf resolves the right slot; the dead Sender is replaced by c and a fresh
// full snapshot is pushed so the returning client re-syncs. A pending Seek
// for that player has its prompt re-sent (the one-shot prompt was lost with the
// old socket). Returns false if the match already ended (nothing to rejoin).
func (m *Match) Reattach(c Sender) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.over {
		return false
	}
	i := m.indexOf(c)
	m.players[i] = c
	m.resetLog()
	m.sendResyncTo(i)
	if m.pending != nil && m.pending.player == i {
		m.sendSeekTo(i)
	} else if m.pending != nil && m.pending.player == 1-i {
		// Reconnected while the opponent is mid-Seek: re-show the indicator.
		c.Send(protocol.Marshal(protocol.OppSeek{Type: protocol.TypeOppSeek, Count: len(m.pending.options)}))
	}
	return true
}

// --- internal helpers (caller holds m.mu) ---

// startTurn refills the player's mana (ramping by one to a cap) and wakes their
// minions: summon sickness clears and the per-turn attack is reset. Freeze is
// NOT cleared here — it thaws at end of turn (see thawAfterTurn).
func (m *Match) startTurn(pi int) {
	// A new turn begins: reset the "died this turn" window for both players (deaths
	// in the just-ended turn no longer count for `revenant_priestess`-style effects).
	m.state[0].diedThisTurn = nil
	m.state[1].diedThisTurn = nil
	ps := m.state[pi]
	if ps.maxMana < maxMana {
		ps.maxMana++
	}
	ps.mana = ps.maxMana
	ps.heroPowerUsed = false
	ps.heroAttacked = false
	ps.nextSecretFree = false    // `spellwarden_magus`'s "this turn" free secret expires
	ps.minionsPlayedThisTurn = 0 // reset `pocket_conjurer`'s first-minion counter
	for _, mn := range ps.board {
		mn.summonedThisTurn = false
		mn.attacksMade = 0
		if mn.destroyAtTurnStart { // Nightmare: the buffed minion dies now (finish() resolves it)
			mn.health = 0
		}
	}
	m.drawCard(pi) // draw for the turn (fatigue if the deck is empty)
	m.fireTriggers(pi, cards.OnTurnStart, nil)
	m.scheduleTurnTimer()
}

// shuffleDeck shuffles player pi's draw pile with the match RNG.
func (m *Match) shuffleDeck(pi int) {
	d := m.state[pi].deck
	m.rng.Shuffle(len(d), func(i, j int) { d[i], d[j] = d[j], d[i] })
}

// dealOpening moves the top n cards of player pi's deck into their hand. Used
// only for the opening hand, which never fatigues or overdraws.
func (m *Match) dealOpening(pi, n int) {
	ps := m.state[pi]
	if n > len(ps.deck) {
		n = len(ps.deck)
	}
	ps.hand = append(ps.hand, ps.deck[:n]...)
	ps.deck = ps.deck[n:]
}

// drawCard draws the top card of player pi's deck into their hand. An empty deck
// inflicts escalating fatigue damage (armor absorbs it) instead. A full hand
// burns the drawn card (overdraw). Emits the appropriate event.
func (m *Match) drawCard(pi int) {
	ps := m.state[pi]
	if len(ps.deck) == 0 {
		ps.fatigue++
		n := ps.fatigue
		absorbed := min(ps.armor, n)
		ps.armor -= absorbed
		ps.heroHP -= n - absorbed
		m.emit(protocol.Event{Kind: "fatigue", Target: m.pid(pi), Amount: n})
		return
	}
	c := ps.deck[0]
	ps.deck = ps.deck[1:]
	if len(ps.hand) >= maxHand {
		// Overdraw: the card is destroyed. Its face is revealed to both players so
		// the burn animation can show what was lost.
		m.emitBurn(pi, c)
		return
	}
	ps.hand = append(ps.hand, c)
}

// emitBurn logs a destroyed card (overdraw or full-hand discard). The card face
// is carried so the client's burn animation can reveal it to both players.
func (m *Match) emitBurn(pi int, c cards.Card) {
	cv := cardView(c)
	m.emit(protocol.Event{Kind: "burn", Target: m.pid(pi), Card: &cv})
}

// thawAfterTurn unfreezes player pi's characters that did not attack this turn,
// so a character frozen during the opponent's turn stays frozen for exactly one
// of pi's turns. Called for the player whose turn is ending. A hero never
// attacks (no weapons yet), so a frozen hero always thaws after one turn.
func (m *Match) thawAfterTurn(pi int) {
	ps := m.state[pi]
	ps.frozen = false
	for _, mn := range ps.board {
		if mn.frozen && !mn.hasAttacked() {
			mn.frozen = false
		}
	}
}

// clearTempBuffs removes "this turn" buffs from every minion (both boards) at end
// of turn, clamping current health to the reduced max. Temp buffs only exist
// during the turn they were cast, so clearing both sides is safe and correct.
func (m *Match) clearTempBuffs() {
	for pi := 0; pi < 2; pi++ {
		for _, mn := range m.state[pi].board {
			kept := mn.enchants[:0]
			had := false
			for _, e := range mn.enchants {
				if e.temp {
					had = true
					continue
				}
				kept = append(kept, e)
			}
			if had {
				mn.enchants = kept
				if mn.health > mn.maxHP() {
					mn.health = mn.maxHP()
				}
			}
		}
	}
}

// bounceMinion removes a minion from its owner's board and returns it to that
// owner's hand as its base card (all enchantments/damage reset). A full hand
// burns it. Emits a "bounce" event. Caller resolves auras/deaths in finish().
func (m *Match) bounceMinion(mn *minion, owner int) {
	m.bounceMinionCost(mn, owner, 0)
}

// bounceMinionCost is bounceMinion with a permanent cost increase on the returned
// card (`snaring_trap`: +2). The bump rides on the hand card's printed cost, so
// later cost auras/rules still apply on top of it.
func (m *Match) bounceMinionCost(mn *minion, owner, costBump int) {
	board := m.state[owner].board
	idx := -1
	for i, x := range board {
		if x == mn {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	m.state[owner].board = append(board[:idx], board[idx+1:]...)
	base, ok := cards.Get(mn.card.ID)
	if !ok {
		base = mn.card
	}
	base.Cost += costBump
	if len(m.state[owner].hand) >= maxHand {
		m.emitBurn(owner, base)
		return
	}
	m.state[owner].hand = append(m.state[owner].hand, base)
	m.emit(protocol.Event{Kind: "bounce", Target: mn.uid, Name: mn.card.Name})
}

func (m *Match) indexOf(c Sender) int {
	if m.players[0].ID() == c.ID() {
		return 0
	}
	return 1
}

// sendStateAll pushes a per-player snapshot — plus the current action's ordered
// event log — to both players.
func (m *Match) sendStateAll() {
	for i := range m.players {
		m.sendStateTo(i)
	}
}

// stateFor builds player i's snapshot (their hand revealed, the opponent's
// hidden) carrying the given event log. resync marks a reconnect/initial snapshot
// whose Events is the full recent history to REPLACE the client log. Caller holds
// m.mu.
func (m *Match) stateFor(i int, events []protocol.Event, resync bool) protocol.State {
	if events == nil {
		events = []protocol.Event{}
	}
	return protocol.State{
		Type:     protocol.TypeState,
		Turn:     m.players[m.turn].ID(),
		TurnNum:  m.turnNum,
		You:      m.players[i].ID(),
		Self:     m.selfView(i, m.players[i].Name()),
		Opp:      m.oppView(1-i, m.players[1-i].Name()),
		Events:   events,
		Mulligan: m.mulligan != nil,
		TurnSecs: m.turnSecondsLeft(),
		Resync:   resync,
	}
}

// sendStateTo pushes player i's snapshot (with the current event log) to player i
// and to every spectator bound to seat i. Used mid-mulligan to confirm one
// player's submission without leaking it to the opponent.
func (m *Match) sendStateTo(i int) {
	b := protocol.Marshal(m.stateFor(i, m.log, false))
	m.players[i].Send(b)
	m.fanout(i, b)
}

// sendResyncTo pushes player i a reconnect snapshot: the current state plus the
// rolling event history (Resync=true tells the client to REPLACE its log rather
// than append), so a returning player recovers the log — including events that
// resolved while they were disconnected. Caller holds m.mu.
func (m *Match) sendResyncTo(i int) {
	b := protocol.Marshal(m.stateFor(i, m.history, true))
	m.players[i].Send(b)
	m.fanout(i, b)
}

// RelayIntent forwards an acting player's ephemeral aiming hint to the viewers who
// see that player as the opponent: the OTHER seat's player + the spectators bound to
// that other seat (those whose "opp" is the acting player). It is never stored, never
// logged, and never touches game state — purely a presentation cue. No-op if c is not
// a seated player or the match is over. Long id strings are dropped (cosmetic field,
// no need to relay anything bigger than a real id).
func (m *Match) RelayIntent(c Sender, oi protocol.OppIntent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.over {
		return
	}
	i := -1
	for s, p := range m.players {
		if p.ID() == c.ID() {
			i = s
			break
		}
	}
	if i < 0 {
		return // not a seated player (e.g. a spectator) — never relay
	}
	if len(oi.Hover) > 64 || len(oi.AimFrom) > 64 || len(oi.AimTo) > 64 {
		return
	}
	b := protocol.Marshal(oi)
	m.players[1-i].Send(b)
	m.fanout(1-i, b)
}

// fanout sends b to every spectator bound to seat. Caller holds m.mu.
func (m *Match) fanout(seat int, b []byte) {
	for obs, s := range m.observers {
		if s == seat {
			obs.Send(b)
		}
	}
}

// SeatOf returns the player-slot index (0 or 1) that c occupies, or -1 if c is
// not a player in this match. Locks internally.
func (m *Match) SeatOf(c Sender) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, p := range m.players {
		if p.ID() == c.ID() {
			return i
		}
	}
	return -1
}

// AddObserver registers c as a spectator of seat (0/1): it begins receiving the
// per-seat snapshots that player gets (their hand revealed, the opponent's
// hidden) but cannot act. A fresh resync snapshot — recent history included — is
// pushed immediately so the spectator's screen matches the watched player's. If
// that player is mid-Seek, an opp_seek indicator is sent too (the option
// faces stay hidden from spectators, like the waiting opponent sees). Returns
// false if the match has ended (nothing to watch) or seat is out of range.
func (m *Match) AddObserver(c Sender, seat int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.over || seat < 0 || seat > 1 {
		return false
	}
	m.observers[c] = seat
	c.Send(protocol.Marshal(m.stateFor(seat, m.history, true)))
	if m.pending != nil {
		c.Send(protocol.Marshal(protocol.OppSeek{Type: protocol.TypeOppSeek, Count: len(m.pending.options)}))
	}
	m.notifySpectators()
	return true
}

// RemoveObserver drops c from the spectator set (on leave/disconnect). Safe to
// call for a client that is not observing. Locks internally.
func (m *Match) RemoveObserver(c Sender) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.observers[c]; !ok {
		return
	}
	delete(m.observers, c)
	m.notifySpectators()
}

// notifySpectators tells both players who is currently watching (sorted names),
// so each can show a "being watched" badge. Sent to the players only, not the
// spectators. Caller holds m.mu.
func (m *Match) notifySpectators() {
	names := make([]string, 0, len(m.observers))
	for obs := range m.observers {
		names = append(names, obs.Name())
	}
	sort.Strings(names)
	b := protocol.Marshal(protocol.Spectators{Type: protocol.TypeSpectators, Names: names})
	for _, p := range m.players {
		p.Send(b)
	}
}
