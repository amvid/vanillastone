package match

import (
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
)

// castChoice plays a Duality (Choose One) card at targetID with option index
// choice. Mirrors castFrom but threads the chosen option through PlayCardAt.
func castChoice(t *testing.T, m *Match, f *fakeSender, pi int, cardID, targetID string, choice int) {
	t.Helper()
	m.state[pi].mana, m.state[pi].maxMana = 10, 10
	m.state[pi].hand = []cards.Card{getCard(cardID)}
	m.sendStateAll()
	if ok, msg := m.PlayCardAt(f, 0, targetID, -1, choice); !ok {
		t.Fatalf("%s (choice %d) should resolve: %s", cardID, choice, msg)
	}
}

// TestManaBloomGivesTempMana: Innervate-style ramp — +2 available mana this turn
// (current mana only, not the crystal cap).
func TestManaBloomGivesTempMana(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = []cards.Card{getCard("mana_bloom")}
	m.state[0].mana, m.state[0].maxMana = 1, 1
	m.sendStateAll()
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("Mana Bloom should resolve: %s", msg)
	}
	if m.state[0].mana != 3 {
		t.Fatalf("Mana Bloom should give +2 mana this turn (1->3), got %d", m.state[0].mana)
	}
	if m.state[0].maxMana != 1 {
		t.Fatalf("Mana Bloom must not raise the crystal cap, got %d", m.state[0].maxMana)
	}
}

// TestVerdantGrowthEmptyCrystal: Wild Growth — +1 crystal cap, but empty (current
// mana unchanged).
func TestVerdantGrowthEmptyCrystal(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = []cards.Card{getCard("verdant_growth")}
	m.state[0].mana, m.state[0].maxMana = 2, 3 // pay the 2 cost, leaving 0
	m.sendStateAll()
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("Verdant Growth should resolve: %s", msg)
	}
	if m.state[0].maxMana != 4 {
		t.Fatalf("Verdant Growth should add an empty crystal (cap 3->4), got %d", m.state[0].maxMana)
	}
	if m.state[0].mana != 0 {
		t.Fatalf("Verdant Growth's crystal is EMPTY — current mana must not change, got %d", m.state[0].mana)
	}
}

// TestVerdantGrowthOverflow: at the 10-crystal cap it instead adds the Excess
// Mana token (`overflow_mana`) to hand.
func TestVerdantGrowthOverflow(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = []cards.Card{getCard("verdant_growth")}
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.sendStateAll()
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("Verdant Growth should resolve: %s", msg)
	}
	if m.state[0].maxMana != 10 {
		t.Fatalf("cap already at 10 must stay 10, got %d", m.state[0].maxMana)
	}
	if len(m.state[0].hand) != 1 || m.state[0].hand[0].ID != "overflow_mana" {
		t.Fatalf("overflow should add overflow_mana to hand, got %v", m.state[0].hand)
	}
}

// TestSavageSlashScalesByHeroAttack: Savagery deals damage equal to the hero's
// current Attack.
func TestSavageSlashScalesByHeroAttack(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "ogre", "crag_ogre", 6, 7, true)
	m.state[0].heroAtkThisTurn = 3 // hero Attack 3 (e.g. after Feral Claws)
	castFrom(t, m, a, 0, "savage_slash", "ogre")
	if o := findMinion(m.state[1].board, "ogre"); o == nil || o.health != 4 {
		t.Fatalf("Savage Slash should deal hero Attack (3) to the minion (7->4), got %v", o)
	}
}

// TestClawSweepHitsAllEnemies: Swipe deals 4 to the chosen enemy and 1 to every
// other enemy character (hero + minions).
func TestClawSweepHitsAllEnemies(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "target", "crag_ogre", 6, 7, true)
	place(m, 1, "other", "war_colossus", 7, 7, true)
	castFrom(t, m, a, 0, "claw_sweep", "target")
	if tgt := findMinion(m.state[1].board, "target"); tgt == nil || tgt.health != 3 {
		t.Fatalf("Claw Sweep should deal 4 to the target (7->3), got %v", tgt)
	}
	if oth := findMinion(m.state[1].board, "other"); oth == nil || oth.health != 6 {
		t.Fatalf("Claw Sweep should deal 1 to other enemy minions (7->6), got %v", oth)
	}
	if m.state[1].heroHP != 29 {
		t.Fatalf("Claw Sweep should deal 1 to the enemy hero (30->29), got %d", m.state[1].heroHP)
	}
}

// TestFeralHowlBuffsAllFriendlyCharacters: Savage Roar gives the hero +2 Attack
// this turn and every friendly minion +2 Attack (temporary).
func TestFeralHowlBuffsAllFriendlyCharacters(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "wolf", "packleader_wolf", 2, 3, true)
	castFrom(t, m, a, 0, "feral_howl", "")
	if m.state[0].heroAtkThisTurn != 2 {
		t.Fatalf("Feral Howl should give the hero +2 Attack, got %d", m.state[0].heroAtkThisTurn)
	}
	if w := findMinion(m.state[0].board, "wolf"); w == nil || w.atk() != 4 {
		t.Fatalf("Feral Howl should give friendly minions +2 Attack (2->4), got %v", w)
	}
}

// TestForestSoulGrantsFinalGasp: Soul of the Forest gives friendly minions a
// deathrattle that summons a 2/2 Thornling.
func TestForestSoulGrantsFinalGasp(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "wolf", "packleader_wolf", 2, 1, true)
	castFrom(t, m, a, 0, "forest_soul", "")
	// Kill the buffed minion; its granted finalGasp should summon a Thornling.
	wolf := findMinion(m.state[0].board, "wolf")
	m.damageMinion(wolf, 5, "x")
	m.resolveDeaths()
	if countCard(m.state[0].board, "thornling") != 1 {
		t.Fatalf("Forest Soul's granted Final Gasp should summon a Thornling, board=%v", m.state[0].board)
	}
}

// TestCallTheGroveChargeAndDoomed: Force of Nature summons three 2/2 Thornlings
// with Charge, each marked to die at the end of the turn.
func TestCallTheGroveChargeAndDoomed(t *testing.T) {
	m, a, _ := newMatch()
	castFrom(t, m, a, 0, "call_the_grove", "")
	if n := countCard(m.state[0].board, "thornling"); n != 3 {
		t.Fatalf("Call the Grove should summon three Thornlings, got %d", n)
	}
	for _, mn := range m.state[0].board {
		if mn.card.ID != "thornling" {
			continue
		}
		if !mn.has(cards.KeywordCharge) {
			t.Fatalf("Call the Grove's Thornlings should have Charge, got %v", mn)
		}
		if !mn.destroyAtTurnEnd {
			t.Fatalf("Call the Grove's Thornlings should die at end of turn, got %v", mn)
		}
	}
}

// TestDualitySpellBranches: Might of the Grove — option 0 buffs your minions
// +1/+1; option 1 summons a 3/2 Panther.
func TestDualitySpellBranches(t *testing.T) {
	// Option 0: buff.
	m, a, _ := newMatch()
	place(m, 0, "wolf", "packleader_wolf", 2, 3, true)
	castChoice(t, m, a, 0, "might_of_the_grove", "", 0)
	if w := findMinion(m.state[0].board, "wolf"); w == nil || w.atk() != 3 || w.maxHP() != 4 {
		t.Fatalf("Might of the Grove option 0 should give +1/+1 (2/3->3/4), got %v", w)
	}
	// Option 1: summon a Panther.
	m2, a2, _ := newMatch()
	castChoice(t, m2, a2, 0, "might_of_the_grove", "", 1)
	if countCard(m2.state[0].board, "grove_panther") != 1 {
		t.Fatalf("Might of the Grove option 1 should summon a Grove Panther, board=%v", m2.state[0].board)
	}
}

// TestDualityMinionSelfBuff: Druid of the Claw — option 0 = +2 Attack & Charge
// (a 6/4 that can attack at once); option 1 = +2 Health & Taunt (a 4/6 Taunt).
func TestDualityMinionSelfBuff(t *testing.T) {
	// Option 0: +2 Attack and Charge.
	m, a, _ := newMatch()
	castChoice(t, m, a, 0, "clawform_druid", "", 0)
	c := m.state[0].board[len(m.state[0].board)-1]
	if c.atk() != 6 || c.maxHP() != 4 {
		t.Fatalf("Clawform Druid option 0 should be a 6/4, got %d/%d", c.atk(), c.maxHP())
	}
	if !c.has(cards.KeywordCharge) {
		t.Fatalf("Clawform Druid option 0 should have Charge, got %v", c)
	}
	// Option 1: +2 Health and Taunt.
	m2, a2, _ := newMatch()
	castChoice(t, m2, a2, 0, "clawform_druid", "", 1)
	c2 := m2.state[0].board[len(m2.state[0].board)-1]
	if c2.atk() != 4 || c2.maxHP() != 6 {
		t.Fatalf("Clawform Druid option 1 should be a 4/6, got %d/%d", c2.atk(), c2.maxHP())
	}
	if !c2.has(cards.KeywordTaunt) {
		t.Fatalf("Clawform Druid option 1 should have Taunt, got %v", c2)
	}
}

// TestGroveWardenDualityTargeted: Keeper-of-the-Grove-style minion — option 0
// deals 2 damage to a target; the minion still enters play.
func TestGroveWardenDualityTargeted(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "ogre", "crag_ogre", 6, 7, true)
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].hand = []cards.Card{getCard("grove_warden")}
	m.sendStateAll()
	if ok, msg := m.PlayCardAt(a, 0, "ogre", -1, 0); !ok {
		t.Fatalf("Grove Warden (deal 2) should resolve: %s", msg)
	}
	if o := findMinion(m.state[1].board, "ogre"); o == nil || o.health != 5 {
		t.Fatalf("Grove Warden option 0 should deal 2 to the target (7->5), got %v", o)
	}
	if countCard(m.state[0].board, "grove_warden") != 1 {
		t.Fatalf("Grove Warden should also enter play, board=%v", m.state[0].board)
	}
}

// TestWildShapeHeroPower: Shapeshift — +1 hero Attack this turn and +1 Armor.
func TestWildShapeHeroPower(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroPower = getCard("wild_shape")
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.sendStateAll()
	if ok, msg := m.HeroPower(a, ""); !ok {
		t.Fatalf("Wild Shape should resolve: %s", msg)
	}
	if m.state[0].heroAtkThisTurn != 1 {
		t.Fatalf("Wild Shape should give +1 hero Attack, got %d", m.state[0].heroAtkThisTurn)
	}
	if m.state[0].armor != 1 {
		t.Fatalf("Wild Shape should give +1 Armor, got %d", m.state[0].armor)
	}
}

// TestWildReclaimDestroysAndOpponentDraws: Naturalize destroys a minion and the
// OPPONENT draws 2 cards.
func TestWildReclaimDestroysAndOpponentDraws(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "ogre", "crag_ogre", 6, 7, true)
	m.state[1].hand = nil // the fixture opening hand is over the cap; start empty so draws land
	m.state[1].deck = testDeck([]string{"crag_ogre", "war_colossus", "molten_hound"})
	before := len(m.state[1].hand)
	castFrom(t, m, a, 0, "wild_reclaim", "ogre")
	if findMinion(m.state[1].board, "ogre") != nil {
		t.Fatal("Wild Reclaim should destroy the target minion")
	}
	if got := len(m.state[1].hand) - before; got != 2 {
		t.Fatalf("Wild Reclaim should make the opponent draw 2, drew %d", got)
	}
}
