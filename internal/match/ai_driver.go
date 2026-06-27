package match

import (
	"math/rand"
	"time"

	"github.com/amvid/vanillastone/internal/cards"
)

// botActionDelay paces the bot's actions so a human can follow what it does
// (each play/attack/hero-power/choice is spaced out, like watching an opponent).
// The client now buffers actions and plays them one at a time at its own animation
// pace (App.tsx action queue), so this no longer has to exceed the client timeline
// to avoid clobbering — it just sets a natural send cadence. Kept near a typical
// action's animation length so the client queue stays shallow. A var (not const)
// so tests can zero it for fast, deterministic runs.
var botActionDelay = 1200 * time.Millisecond

// mulliganKeepCost is the highest mana cost the bot keeps in its opening hand;
// anything dearer is tossed to dig for an early curve. 3 keeps the 1–3 drops that
// matter most in the first turns.
const mulliganKeepCost = 3

// NewVsAI creates a match against an AI opponent. Who goes first is decided by a
// coin flip: the human takes seat 0 (first) or seat 1 (second, and so gets Mana
// Surge), with the bot on the other seat. botName is the opponent's display name.
// The caller still drives the lifecycle: Start() then DriveBotMulligan().
func NewVsAI(id string, human Sender, seed int64, humanDeck, botDeck []cards.Card, botName string) *Match {
	bot := botSender{id: "ai:" + id, name: botName}
	// A throwaway stream off the seed (not the game RNG) picks the seating, so it
	// varies per match without disturbing the deterministic game/AI randomness.
	humanFirst := rand.New(rand.NewSource(seed)).Intn(2) == 0
	var m *Match
	if humanFirst {
		m = New(id, human, bot, seed, humanDeck, botDeck)
		m.EnableAI(1, seed)
	} else {
		m = New(id, bot, human, seed, botDeck, humanDeck)
		m.EnableAI(0, seed)
	}
	return m
}

// EnableAI turns the match's given seat into an AI opponent: a planner-only RNG
// stream (independent of the game RNG so simulation never consumes game
// randomness) and the seat marker the turn loop checks. Call before Start.
func (m *Match) EnableAI(seat int, seed int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.aiSeat = seat
	m.aiRng = rand.New(rand.NewSource(seed))
}

// DriveBotMulligan auto-submits the bot's opening mulligan so play can begin.
// Exported wrapper for the lobby to call after Start() on a vs-AI match.
func (m *Match) DriveBotMulligan() { m.driveBotMulligan() }

// isAITurn reports whether the AI controls the seat whose turn it currently is.
// Caller holds m.mu.
func (m *Match) isAITurn() bool {
	return m.aiSeat >= 0 && m.turn == m.aiSeat && !m.over && m.mulligan == nil
}

// driveBotMulligan auto-submits the bot's opening mulligan, tossing expensive
// cards to dig for an early curve (botMulliganTosses). Safe to call after Start on
// a vs-AI match; a no-op if there's no AI or the mulligan already resolved.
func (m *Match) driveBotMulligan() {
	m.mu.Lock()
	ai := m.aiSeat
	inMulligan := m.mulligan != nil && ai >= 0
	var toss []int
	if inMulligan {
		toss = m.botMulliganTosses(ai)
	}
	m.mu.Unlock()
	if !inMulligan {
		return
	}
	go func() {
		time.Sleep(botActionDelay)
		m.Mulligan(m.players[ai], toss)
	}()
}

// botMulliganTosses picks the opening-hand cards the bot replaces: anything that
// costs more than mulliganKeepCost, so it digs for a playable early curve instead
// of keeping a clump of late-game cards. Caller holds m.mu.
func (m *Match) botMulliganTosses(seat int) []int {
	var toss []int
	for i, c := range m.state[seat].hand {
		if c.Cost > mulliganKeepCost {
			toss = append(toss, i)
		}
	}
	return toss
}

// runBotTurn drives one full AI turn in its own goroutine: greedily apply the
// best-scoring action until nothing improves the position (or a safety bound),
// resolving any Seek the bot's own battlecries open, then end the turn. Each
// action is paced by botActionDelay for readability. All mutations go through the
// same authoritative handlers a human uses.
func (m *Match) runBotTurn(seat int) {
	for i := 0; i < maxBotMoves; i++ {
		m.mu.Lock()
		if m.over || m.turn != seat || m.mulligan != nil {
			m.mu.Unlock()
			return
		}
		pendingForMe := m.pending != nil && m.pending.player == seat
		m.mu.Unlock()

		if pendingForMe {
			m.botChoose(seat)
			continue
		}

		m.mu.Lock()
		mv, ok := m.planBest(seat)
		m.mu.Unlock()
		if !ok {
			break
		}
		time.Sleep(botActionDelay)
		mv.applyTo(m, seat)
	}
	time.Sleep(botActionDelay)
	m.EndTurn(m.players[seat])
}

// botChoose resolves a pending Seek for the bot, picking the highest-value
// option (a simple body/cost heuristic — good enough for the dumb opponent).
func (m *Match) botChoose(seat int) {
	m.mu.Lock()
	idx := 0
	if m.pending != nil && m.pending.player == seat {
		idx = bestSeekIndex(m.pending.options)
	}
	m.mu.Unlock()
	time.Sleep(botActionDelay)
	m.Choose(m.players[seat], idx)
}

// bestSeekIndex picks a Seek option: the highest atk+health minion, else
// the highest-cost card. Deliberately crude — Seek is rare on the dumb bot.
func bestSeekIndex(opts []cards.Card) int {
	best, bestVal := 0, -1
	for i, c := range opts {
		v := c.Cost
		if c.Type == cards.TypeMinion {
			v = c.Attack + c.Health + 100 // prefer a body over a spell/weapon
		}
		if v > bestVal {
			best, bestVal = i, v
		}
	}
	return best
}
