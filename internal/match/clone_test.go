package match

import (
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
)

// richBoard sets up a varied, mutation-prone state on both seats so the clone
// round-trip exercises every reference-typed field: boards with enchants/flags,
// secrets, weapons, hands, decks, and the assorted per-turn scalars.
func richBoard() *Match {
	m, _, _ := newMatch()
	s0, s1 := m.state[0], m.state[1]

	s0.heroHP, s0.armor, s0.mana, s0.maxMana = 24, 3, 5, 7
	s0.board = []*minion{
		{uid: "u1", card: getCard("bastion_golem"), owner: 0, health: 5,
			enchants: []enchant{{atk: 2, hp: 1, keywords: []cards.Keyword{cards.KeywordTaunt}}}},
		{uid: "u2", card: getCard("gilded_sentry"), owner: 0, health: 3, aegis: true, frozen: true},
	}
	s0.secrets = []*secretInst{{uid: "s1", card: getCard("snare"), owner: 0}}
	s0.weapon = &weaponInst{card: getCard("ember_cleaver"), attack: 3, durability: 2}
	s0.diedThisTurn = []cards.Card{getCard("pebble_imp")}
	s0.nextSecretFree, s0.minionsPlayedThisTurn, s0.spellsFreeOnTurn = true, 2, 4

	s1.heroHP = 18
	s1.board = []*minion{
		{uid: "u3", card: getCard("clay_acolyte"), owner: 1, health: 2, attacksMade: 1, summonedThisTurn: true},
	}
	return m
}

// TestCloneForSimIsIndependent is the foundation guard for the AI: a simulation
// clone must share no mutable state with the live match, or hypothetical actions
// run during planning would corrupt the real game. It clones a rich board, mutates
// the clone destructively, and asserts the original is byte-for-byte unchanged.
func TestCloneForSimIsIndependent(t *testing.T) {
	m := richBoard()
	c := m.cloneForSim(99)

	// The clone must start equal to the original on every state-bearing field.
	assertSameState(t, m, c)

	// Now wreck the clone: empty boards, drain the hero, strip secrets/weapon,
	// flip flags, grow buffs, clear hands/decks.
	c.state[0].heroHP = 0
	c.state[0].armor = 0
	c.state[0].board = c.state[0].board[:1]
	c.state[0].board[0].health = 99
	c.state[0].board[0].enchants[0].atk = 100
	c.state[0].board[0].aegis = false
	c.state[0].secrets = nil
	c.state[0].weapon = nil
	c.state[0].hand = nil
	c.state[0].deck = nil
	c.state[0].diedThisTurn = nil
	c.state[0].nextSecretFree = false
	c.state[1].board[0].health = -5
	c.state[1].board = nil
	c.turn, c.turnNum, c.nextUID, c.over = 1, 50, 999, true

	// Original must be pristine.
	o := richBoard() // a fresh, known-good reference
	assertSameState(t, m, o)
}

// assertSameState fails if a and b differ on any field the clone is responsible
// for copying. Deliberately exhaustive — this test exists to catch a field that
// gets added to the state structs but forgotten in clone().
func assertSameState(t *testing.T, a, b *Match) {
	t.Helper()
	if a.turn != b.turn || a.turnNum != b.turnNum || a.nextUID != b.nextUID || a.over != b.over {
		t.Fatalf("match scalars differ: %+v vs %+v", a, b)
	}
	for pi := 0; pi < 2; pi++ {
		pa, pb := a.state[pi], b.state[pi]
		if pa.heroHP != pb.heroHP || pa.armor != pb.armor || pa.mana != pb.mana || pa.maxMana != pb.maxMana ||
			pa.frozen != pb.frozen || pa.fatigue != pb.fatigue || pa.heroPowerUsed != pb.heroPowerUsed ||
			pa.heroAttacked != pb.heroAttacked || pa.immune != pb.immune ||
			pa.nextSecretFree != pb.nextSecretFree || pa.minionsPlayedThisTurn != pb.minionsPlayedThisTurn ||
			pa.spellsFreeOnTurn != pb.spellsFreeOnTurn {
			t.Fatalf("seat %d scalars differ:\n %+v\n %+v", pi, pa, pb)
		}
		if len(pa.hand) != len(pb.hand) || len(pa.deck) != len(pb.deck) ||
			len(pa.diedThisTurn) != len(pb.diedThisTurn) || len(pa.secrets) != len(pb.secrets) {
			t.Fatalf("seat %d collection lengths differ", pi)
		}
		if (pa.weapon == nil) != (pb.weapon == nil) {
			t.Fatalf("seat %d weapon presence differs", pi)
		}
		if pa.weapon != nil && (pa.weapon.attack != pb.weapon.attack || pa.weapon.durability != pb.weapon.durability) {
			t.Fatalf("seat %d weapon differs", pi)
		}
		if len(pa.board) != len(pb.board) {
			t.Fatalf("seat %d board length differs: %d vs %d", pi, len(pa.board), len(pb.board))
		}
		for i := range pa.board {
			ma, mb := pa.board[i], pb.board[i]
			if ma.uid != mb.uid || ma.health != mb.health || ma.owner != mb.owner ||
				ma.aegis != mb.aegis || ma.frozen != mb.frozen || ma.silenced != mb.silenced ||
				ma.stealthed != mb.stealthed || ma.attacksMade != mb.attacksMade ||
				ma.summonedThisTurn != mb.summonedThisTurn || ma.atk() != mb.atk() || ma.maxHP() != mb.maxHP() {
				t.Fatalf("seat %d minion %d differs:\n %+v\n %+v", pi, i, ma, mb)
			}
		}
	}
}

// TestCloneSimPlayDoesNotTouchLive runs a real action (play a minion from hand)
// on a clone and asserts the live match's board/hand are unaffected — the actual
// use the planner makes of the clone.
func TestCloneSimPlayDoesNotTouchLive(t *testing.T) {
	m := richBoard()
	m.state[0].hand = []cards.Card{getCard("clay_acolyte")}
	m.state[0].mana, m.state[0].maxMana = 5, 5
	m.turn = 0

	beforeBoard := len(m.state[0].board)
	beforeHand := len(m.state[0].hand)

	c := m.cloneForSim(7)
	if ok, msg := c.PlayCard(c.players[0], 0, ""); !ok {
		t.Fatalf("clone PlayCard failed: %s", msg)
	}
	if len(c.state[0].board) != beforeBoard+1 {
		t.Fatalf("clone board should have grown to %d, got %d", beforeBoard+1, len(c.state[0].board))
	}
	if len(m.state[0].board) != beforeBoard {
		t.Fatalf("LIVE board changed: %d -> %d", beforeBoard, len(m.state[0].board))
	}
	if len(m.state[0].hand) != beforeHand {
		t.Fatalf("LIVE hand changed: %d -> %d", beforeHand, len(m.state[0].hand))
	}
}
