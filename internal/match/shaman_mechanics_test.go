package match

import (
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
)

// findByCard returns the first minion on board whose base card id matches — used
// for summoned/resummoned minions, which get fresh numeric uids (findMinion keys
// off uid).
func findByCard(board []*minion, cardID string) *minion {
	for _, mn := range board {
		if mn.card.ID == cardID {
			return mn
		}
	}
	return nil
}

// TestOverloadLocksNextTurn: playing an Overload card queues locked crystals that
// bite at the START of the caster's next turn (the whole risk of the mechanic —
// cheap now, poorer next turn), then free up the turn after.
func TestOverloadLocksNextTurn(t *testing.T) {
	m, a, _ := newMatch()
	castFrom(t, m, a, 0, "voltstrike", oppHeroTarget) // Overload (1)
	if got := m.state[0].overloadNext; got != 1 {
		t.Fatalf("Voltstrike should queue Overload 1, got %d", got)
	}
	// The caster's next turn begins: the crystal is locked (10 max, 9 usable).
	m.startTurn(0)
	if m.state[0].overloadLocked != 1 || m.state[0].mana != 9 {
		t.Fatalf("Overload should lock 1 crystal next turn (mana 9/locked 1), got mana=%d locked=%d",
			m.state[0].mana, m.state[0].overloadLocked)
	}
	if m.state[0].overloadNext != 0 {
		t.Fatalf("Overload queue should clear after locking, got %d", m.state[0].overloadNext)
	}
	// The turn after: no lingering lock.
	m.startTurn(0)
	if m.state[0].overloadLocked != 0 || m.state[0].mana != 10 {
		t.Fatalf("Overload should free up the following turn (mana 10/locked 0), got mana=%d locked=%d",
			m.state[0].mana, m.state[0].overloadLocked)
	}
}

// TestUnboundElementalGrowsOnOverload: Riftbound Elemental gains +1/+1 whenever
// its controller plays an Overload card — the payoff that makes Overload a plan,
// not just a tax. A non-Overload card must NOT grow it.
func TestUnboundElementalGrowsOnOverload(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "ub", "riftbound_elemental", 3, 4, true)
	castFrom(t, m, a, 0, "voltstrike", oppHeroTarget) // Overload (1)
	if ub := findMinion(m.state[0].board, "ub"); ub.atk() != 4 || ub.maxHP() != 5 {
		t.Fatalf("Riftbound should grow to 4/5 after an Overload card, got %d/%d", ub.atk(), ub.maxHP())
	}
	// A no-Overload spell leaves it unchanged.
	castFrom(t, m, a, 0, "frost_jolt", oppHeroTarget)
	if ub := findMinion(m.state[0].board, "ub"); ub.atk() != 4 || ub.maxHP() != 5 {
		t.Fatalf("Riftbound must not grow on a non-Overload card, got %d/%d", ub.atk(), ub.maxHP())
	}
}

// TestDoomhammerAttacksTwice: a Twinstrike (Windfury) weapon lets the hero attack
// twice in a turn — the reason to run the weapon over a vanilla one.
func TestDoomhammerAttacksTwice(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].weapon = &weaponInst{card: getCard("ruinhammer"), attack: 2, durability: 8}
	m.sendStateAll()
	if ok, msg := m.Attack(a, selfHeroTarget, oppHeroTarget); !ok {
		t.Fatalf("first hero attack should resolve: %s", msg)
	}
	if ok, msg := m.Attack(a, selfHeroTarget, oppHeroTarget); !ok {
		t.Fatalf("Twinstrike weapon should allow a SECOND hero attack: %s", msg)
	}
	if ok, _ := m.Attack(a, selfHeroTarget, oppHeroTarget); ok {
		t.Fatal("hero must not attack a THIRD time with a Twinstrike weapon")
	}
	if hp := m.state[1].heroHP; hp != 26 { // 30 - 2 - 2
		t.Fatalf("two 2-damage hero attacks should leave 26, got %d", hp)
	}
}

// TestAncestralSpiritResummons: Ancestral Spirit grants a minion a deathrattle
// that brings back a FRESH copy — the point is recovering the body after a trade.
func TestAncestralSpiritResummons(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "ele", "cinder_elemental", 6, 5, true) // 6/5 body
	castFrom(t, m, a, 0, "spirit_bond", "ele")
	// Kill it: the granted deathrattle resummons a fresh Cinder Elemental.
	ele := findMinion(m.state[0].board, "ele")
	ele.health = 0
	m.finish()
	reborn := findByCard(m.state[0].board, "cinder_elemental")
	if reborn == nil {
		t.Fatal("Ancestral Spirit should resummon the minion on death")
	}
	if reborn.maxHP() != 5 || reborn.atk() != 6 {
		t.Fatalf("resummoned copy should be a fresh 6/5, got %d/%d", reborn.atk(), reborn.maxHP())
	}
}

// TestRockbiterBuffsHeroThisTurn: Rockbiter can pump the HERO's Attack with no
// weapon, and only for this turn — enabling a face swing off an empty board.
func TestRockbiterBuffsHeroThisTurn(t *testing.T) {
	m, a, _ := newMatch()
	m.turn = 0
	castFrom(t, m, a, 0, "stonefury", selfHeroTarget)
	if m.state[0].heroAtkThisTurn != 3 {
		t.Fatalf("Rockbiter should give the hero +3 Attack, got %d", m.state[0].heroAtkThisTurn)
	}
	if !heroCanAttack(m.state[0]) {
		t.Fatal("hero with +3 Attack (no weapon) should be able to attack")
	}
	// End of turn clears it (it was "this turn" only).
	m.endTurnLocked()
	if m.state[0].heroAtkThisTurn != 0 {
		t.Fatalf("Rockbiter's hero Attack must clear at end of turn, got %d", m.state[0].heroAtkThisTurn)
	}
}

// TestRockbiterBuffsMinionThisTurn: the same spell can instead pump a friendly
// minion, and only for this turn.
func TestRockbiterBuffsMinionThisTurn(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "w", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "stonefury", "w")
	if w := findMinion(m.state[0].board, "w"); w.atk() != 9 {
		t.Fatalf("Rockbiter should give the minion +3 Attack (6->9), got %d", w.atk())
	}
	m.endTurnLocked()
	if w := findMinion(m.state[0].board, "w"); w.atk() != 6 {
		t.Fatalf("the +3 should expire at turn end (back to 6), got %d", w.atk())
	}
}

// TestFarSightDiscountsDrawnCard: Far Sight draws a card AND makes that specific
// card 3 cheaper in hand — the reason to run it over a plain cantrip.
func TestFarSightDiscountsDrawnCard(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].deck = []cards.Card{getCard("cinder_elemental")} // printed cost 6
	m.sendStateAll()
	castFrom(t, m, a, 0, "distant_sight", "")
	drawn := m.state[0].hand[len(m.state[0].hand)-1]
	if drawn.ID != "cinder_elemental" || drawn.Cost != 3 {
		t.Fatalf("Far Sight should draw the card at Cost 3 (6-3), got %s cost %d", drawn.ID, drawn.Cost)
	}
}

// TestHexTransformsToFrog: Hex turns any minion into a 0/1 Frog with Taunt,
// wiping its stats and text — hard removal of a big threat.
func TestHexTransformsToFrog(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "big", "war_colossus", 7, 7, true)
	castFrom(t, m, a, 0, "toadcurse", "big")
	frog := findMinion(m.state[1].board, "big") // same uid, new card
	if frog == nil || frog.card.ID != "spirit_frog" {
		t.Fatalf("Hex should transform the minion into a Frog, got %v", frog)
	}
	if frog.atk() != 0 || frog.maxHP() != 1 || !frog.has(cards.KeywordTaunt) {
		t.Fatalf("the Frog should be a 0/1 with Taunt, got %d/%d taunt=%v", frog.atk(), frog.maxHP(), frog.has(cards.KeywordTaunt))
	}
}

// TestTotemicCallSummonsUniqueTotem: the hero power summons a Totem the player
// does NOT already control — with three of four out, it must produce the fourth,
// never a duplicate.
func TestTotemicCallSummonsUniqueTotem(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroPower = getCard("call_totem")
	m.state[0].mana, m.state[0].maxMana = 10, 10
	place(m, 0, "t1", "mending_totem", 0, 2, true)
	place(m, 0, "t2", "ember_totem", 1, 1, true)
	place(m, 0, "t3", "stoneshell_totem", 0, 2, true)
	m.sendStateAll()
	if ok, msg := m.HeroPower(a, ""); !ok {
		t.Fatalf("Totemic Call should resolve: %s", msg)
	}
	if findByCard(m.state[0].board, "stormcrest_totem") == nil {
		t.Fatal("with three totems out, Totemic Call must summon the missing fourth (Stormcrest)")
	}
	if len(m.state[0].board) != 4 {
		t.Fatalf("expected 4 totems on board, got %d", len(m.state[0].board))
	}
}

// TestTotemicMightBuffsOnlyTotems: Totemic Might pumps your Totems' Health and
// leaves non-Totem minions untouched — a totem-tribe payoff.
func TestTotemicMightBuffsOnlyTotems(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "tot", "ember_totem", 1, 1, true)
	place(m, 0, "guy", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "totem_bulwark", "")
	if tot := findMinion(m.state[0].board, "tot"); tot.maxHP() != 3 {
		t.Fatalf("Totemic Might should give the Totem +2 Health (1->3), got %d", tot.maxHP())
	}
	if guy := findMinion(m.state[0].board, "guy"); guy.maxHP() != 7 {
		t.Fatalf("Totemic Might must not buff a non-Totem, got %d", guy.maxHP())
	}
}

// TestFrostShockDamagesAndFreezes: Frost Shock both pings AND freezes the target,
// buying a turn against an attacker.
func TestFrostShockDamagesAndFreezes(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "atk", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "frost_jolt", "atk")
	atk := findMinion(m.state[1].board, "atk")
	if atk.health != 6 {
		t.Fatalf("Frost Shock should deal 1 (7->6), got %d", atk.health)
	}
	if !atk.frozen {
		t.Fatal("Frost Shock should Freeze its target")
	}
}

// TestEarthShockSilencesThenDamages: Earth Shock strips a minion (removing a
// Taunt/deathrattle/aura) THEN pings it — the ordering lets it kill a 1-Health
// body whose Health came from an enchantment it just silenced away.
func TestEarthShockSilencesThenDamages(t *testing.T) {
	m, a, _ := newMatch()
	// A 0/1 totem buffed to 0/3 by an enchant: silence drops it back to 0/1, then
	// the 1 damage kills it. Guards the silence-BEFORE-damage ordering.
	mn := &minion{uid: "t", card: getCard("ember_totem"), owner: 1,
		enchants: []enchant{{hp: 2}}, health: 3}
	m.state[1].board = append(m.state[1].board, mn)
	castFrom(t, m, a, 0, "stonejolt", "t")
	m.finish()
	if findMinion(m.state[1].board, "t") != nil {
		t.Fatal("Earth Shock should silence (0/3 -> 0/1) then deal 1, killing the totem")
	}
}

// TestSplitBoltHitsTwoRandomEnemyMinions: Forked Lightning spreads its damage
// across two distinct enemy minions, not one — board control, not single-target.
func TestSplitBoltHitsTwoRandomEnemyMinions(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "e1", "crag_ogre", 6, 7, true)
	place(m, 1, "e2", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "split_bolt", "")
	e1, e2 := findMinion(m.state[1].board, "e1"), findMinion(m.state[1].board, "e2")
	if e1.health != 5 || e2.health != 5 {
		t.Fatalf("Split Bolt should deal 2 to each of two enemy minions (7->5), got %d and %d", e1.health, e2.health)
	}
}
