package match

import (
	"math/rand"

	"github.com/amvid/vanillastone/internal/cards"
)

// botSender is a Sender with no socket: it carries an id/name (so action
// validation that checks players[turn].ID() passes) but discards every message.
// Used both as the live AI player's seat and as the stub seats on a simulation
// clone (where broadcasts must go nowhere).
type botSender struct {
	id   string
	name string
}

func (b botSender) ID() string   { return b.id }
func (b botSender) Name() string { return b.name }
func (b botSender) Send([]byte)  {}

// cloneForSim returns a deep, independent copy of the match for hypothetical
// action simulation. Applying actions to the copy never touches the live match.
// The copy gets:
//   - deep-copied per-seat state (board minions, secrets, weapon, hand, deck),
//   - no-op (botSender) seats so finish()/broadcast() send nowhere,
//   - a fresh RNG seeded from the caller (NOT read off m.rng, so cloning is pure),
//   - timers disabled (turnDuration 0 → activeTurnDuration arms no time.AfterFunc),
//   - an empty event log / history / observers.
//
// cards.Card values (in hand/deck, on minions, secrets, weapon, hero power) are
// copied shallowly: their definition slices/pointers (Triggers, Effect, Aura, …)
// are immutable and never mutated in play, and simulation only mutates
// health/enchants/flags — so sharing those backings across the clone is safe.
func (m *Match) cloneForSim(seed int64) *Match {
	c := &Match{
		ID:      m.ID + "#sim",
		players: [2]Sender{botSender{m.players[0].ID(), m.players[0].Name()}, botSender{m.players[1].ID(), m.players[1].Name()}},
		turn:    m.turn,
		turnNum: m.turnNum,
		nextUID: m.nextUID,
		over:    m.over,
		rng:     rand.New(rand.NewSource(seed)),

		observers:    make(map[Sender]int),
		turnDuration: 0,                              // disable the auto-end timer on the throwaway
		aiSeat:       -1,                             // a sim must never spawn the async bot driver when its turn is ended (lookahead)
		aiRng:        rand.New(rand.NewSource(seed)), // lets nested planning (opponent-reply sim) seed its own sub-clones
	}
	c.state[0] = m.state[0].clone()
	c.state[1] = m.state[1].clone()
	if m.pending != nil {
		c.pending = &pendingChoice{
			player:  m.pending.player,
			options: append([]cards.Card(nil), m.pending.options...),
		}
	}
	if m.mulligan != nil {
		c.mulligan = &mulliganState{done: m.mulligan.done}
	}
	return c
}

// clone deep-copies a player's state: every slice gets a fresh backing array and
// every owned pointer (minion, secret, weapon) a fresh instance, so the copy
// shares no mutable state with the original. cards.Card values inside are copied
// by value (see cloneForSim for why that's safe).
func (ps *playerState) clone() *playerState {
	cp := *ps // scalars + heroPower (value) + slice/pointer headers; reference types fixed below
	cp.hand = append([]cards.Card(nil), ps.hand...)
	cp.deck = append([]cards.Card(nil), ps.deck...)
	cp.diedThisTurn = append([]cards.Card(nil), ps.diedThisTurn...)
	cp.board = make([]*minion, len(ps.board))
	for i, mn := range ps.board {
		cp.board[i] = mn.clone()
	}
	cp.secrets = make([]*secretInst, len(ps.secrets))
	for i, s := range ps.secrets {
		s2 := *s
		cp.secrets[i] = &s2
	}
	if ps.weapon != nil {
		w := *ps.weapon
		cp.weapon = &w
	}
	return &cp
}

// clone deep-copies a minion instance, giving it a fresh enchants slice so buffs
// added/stripped on the copy don't leak to the original. enchant.keywords is
// shared shallowly (read-only after creation, like card definitions).
func (mn *minion) clone() *minion {
	cp := *mn
	cp.enchants = append([]enchant(nil), mn.enchants...)
	return &cp
}
