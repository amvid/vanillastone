package match

import (
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
)

// castCombo plays cardID with Chain (Combo) armed: it seeds cardsPlayedThisTurn so
// the engine treats a prior card as already played this turn. Mirrors castFrom
// otherwise. Chain matters here because the WHOLE point of these cards is the
// combo upside — a test that couldn't tell chain from no-chain wouldn't guard it.
func castCombo(t *testing.T, m *Match, f *fakeSender, pi int, cardID, targetID string, priorCards int) {
	t.Helper()
	m.state[pi].mana, m.state[pi].maxMana = 10, 10
	m.state[pi].cardsPlayedThisTurn = priorCards
	m.state[pi].hand = []cards.Card{getCard(cardID)}
	m.sendStateAll()
	if ok, msg := m.PlayCard(f, 0, targetID); !ok {
		t.Fatalf("%s (chain) should resolve: %s", cardID, msg)
	}
}

// TestBlindsideNeedsUndamagedTarget: Backstab hits only a full-Health minion, and
// is rejected outright when the only minion is already damaged (the whole card is
// the tempo of a "free" hit on something untouched).
func TestBlindsideNeedsUndamagedTarget(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "ogre", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "blindside", "ogre")
	if o := findMinion(m.state[1].board, "ogre"); o == nil || o.health != 5 {
		t.Fatalf("Blindside should deal 2 to an undamaged minion (7->5), got %v", o)
	}
	// Now the ogre is damaged: Blindside can no longer target it, so with no other
	// target the spell must be rejected (a spell can't cast at nothing).
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].hand = []cards.Card{getCard("blindside")}
	m.sendStateAll()
	if ok, _ := m.PlayCard(a, 0, "ogre"); ok {
		t.Fatal("Blindside must be rejected against a damaged minion")
	}
}

// TestColdVenomChain: +2 Attack normally, +4 under Chain — the combo doubling is
// the reason to hold a card back and play it second.
func TestColdVenomChain(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "m1", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "cold_venom", "m1")
	if o := findMinion(m.state[0].board, "m1"); o.atk() != 8 {
		t.Fatalf("Cold Venom with no Chain should give +2 Attack (6->8), got %d", o.atk())
	}
	m2 := &minion{uid: "m2", card: getCard("crag_ogre"), owner: 0, health: 7}
	m.state[0].board = append(m.state[0].board, m2)
	castCombo(t, m, a, 0, "cold_venom", "m2", 1)
	if m2.atk() != 10 {
		t.Fatalf("Cold Venom under Chain should give +4 Attack (6->10), got %d", m2.atk())
	}
}

// TestLacerateChain: 2 damage normally, 4 under Chain.
func TestLacerateChain(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "t1", "crag_ogre", 6, 9, true)
	castFrom(t, m, a, 0, "lacerate", "t1")
	if o := findMinion(m.state[1].board, "t1"); o.health != 7 {
		t.Fatalf("Lacerate no-Chain should deal 2 (9->7), got %d", o.health)
	}
	place(m, 1, "t2", "crag_ogre", 6, 9, true)
	castCombo(t, m, a, 0, "lacerate", "t2", 1)
	if o := findMinion(m.state[1].board, "t2"); o.health != 5 {
		t.Fatalf("Lacerate under Chain should deal 4 (9->5), got %d", o.health)
	}
}

// TestGuildAgentChainDamage: no battlecry without Chain (a vanilla 3/3), deals 3
// with Chain. Encodes that the combo body is conditional, not free.
func TestGuildAgentChainDamage(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "face", "crag_ogre", 6, 9, true)
	castFrom(t, m, a, 0, "guild_agent", "") // no Chain: plays as a plain body, no target needed
	if o := findMinion(m.state[1].board, "face"); o.health != 9 {
		t.Fatalf("Guild Agent without Chain must deal no damage, got %d", o.health)
	}
	castCombo(t, m, a, 0, "guild_agent", "face", 1)
	if o := findMinion(m.state[1].board, "face"); o.health != 6 {
		t.Fatalf("Guild Agent under Chain should deal 3 (9->6), got %d", o.health)
	}
}

// TestGuildRingleaderChainSummon: summons a 2/1 Guild Thug only under Chain.
func TestGuildRingleaderChainSummon(t *testing.T) {
	m, a, _ := newMatch()
	castFrom(t, m, a, 0, "guild_ringleader", "")
	if n := countCard(m.state[0].board, "guild_thug"); n != 0 {
		t.Fatalf("Guild Ringleader without Chain must not summon, got %d thugs", n)
	}
	castCombo(t, m, a, 0, "guild_ringleader", "", 1)
	if n := countCard(m.state[0].board, "guild_thug"); n != 1 {
		t.Fatalf("Guild Ringleader under Chain should summon a Guild Thug, got %d", n)
	}
}

// TestShadowlordVexScalesByCardsPlayed: Edwin gains +2/+2 for each OTHER card
// played this turn — it must NOT count itself. With two prior cards it is a
// 2/2 + (2 * +2/+2) = 6/6.
func TestShadowlordVexScalesByCardsPlayed(t *testing.T) {
	m, a, _ := newMatch()
	castCombo(t, m, a, 0, "shadowlord_vex", "", 2)
	if len(m.state[0].board) != 1 {
		t.Fatalf("Vex should be on board, got %d minions", len(m.state[0].board))
	}
	v := m.state[0].board[0]
	if v.atk() != 6 || v.maxHP() != 6 {
		t.Fatalf("Vex after 2 prior cards should be 6/6, got %v", v)
	}
}

// TestSkullcrackChainReturns: 2 to the enemy face always; under Chain the spell
// comes back to hand at the caster's next turn start (the combo grants recursion).
func TestSkullcrackChainReturns(t *testing.T) {
	m, a, _ := newMatch()
	m.state[1].heroHP = 30
	castCombo(t, m, a, 0, "skullcrack", "", 1)
	if m.state[1].heroHP != 28 {
		t.Fatalf("Skullcrack should deal 2 to the enemy hero (30->28), got %d", m.state[1].heroHP)
	}
	if len(m.state[0].returnToHandNextTurn) != 1 || m.state[0].returnToHandNextTurn[0].ID != "skullcrack" {
		t.Fatalf("Skullcrack under Chain should queue a return, got %v", m.state[0].returnToHandNextTurn)
	}
	m.state[0].hand = nil
	m.startTurn(0)
	if countHand(m.state[0].hand, "skullcrack") != 1 {
		t.Fatal("Skullcrack should be back in hand at the caster's next turn start")
	}
}

// TestSkullcrackNoChainNoReturn: without Chain it is a one-shot.
func TestSkullcrackNoChainNoReturn(t *testing.T) {
	m, a, _ := newMatch()
	castFrom(t, m, a, 0, "skullcrack", "")
	if len(m.state[0].returnToHandNextTurn) != 0 {
		t.Fatalf("Skullcrack without Chain must not return, got %v", m.state[0].returnToHandNextTurn)
	}
}

// TestTurncoatHitsNeighbors: Betrayal — the forced minion deals its Attack to the
// two minions beside it, and is itself unharmed.
func TestTurncoatHitsNeighbors(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "left", "crag_ogre", 6, 9, true)
	place(m, 1, "mid", "crag_ogre", 6, 9, true)
	place(m, 1, "right", "crag_ogre", 6, 9, true)
	castFrom(t, m, a, 0, "turncoat", "mid")
	l, mid, r := findMinion(m.state[1].board, "left"), findMinion(m.state[1].board, "mid"), findMinion(m.state[1].board, "right")
	if l.health != 3 || r.health != 3 {
		t.Fatalf("Turncoat should deal the mid minion's 6 Attack to both neighbours (9->3), got L=%d R=%d", l.health, r.health)
	}
	if mid.health != 9 {
		t.Fatalf("Turncoat must not harm the forced minion, got %d", mid.health)
	}
}

// TestBladeWhirlSweepsWithWeapon: Blade Flurry destroys the weapon and deals its
// Attack to every enemy minion; fizzles with no weapon.
func TestBladeWhirlSweepsWithWeapon(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].weapon = &weaponInst{card: getCard("assassins_edge"), attack: 2, durability: 5}
	place(m, 1, "e1", "crag_ogre", 6, 9, true)
	place(m, 1, "e2", "crag_ogre", 6, 9, true)
	castFrom(t, m, a, 0, "blade_whirl", "")
	if m.state[0].weapon != nil {
		t.Fatal("Blade Whirl should destroy the caster's weapon")
	}
	for _, id := range []string{"e1", "e2"} {
		if o := findMinion(m.state[1].board, id); o == nil || o.health != 7 {
			t.Fatalf("Blade Whirl should deal the weapon's 2 to %s (9->7), got %v", id, o)
		}
	}
}

// TestGroundworkDiscountsNextSpell: Preparation drops the next spell's cost by 2
// and is consumed by that spell (not permanent).
func TestGroundworkDiscountsNextSpell(t *testing.T) {
	m, a, _ := newMatch()
	castFrom(t, m, a, 0, "groundwork", "")
	if m.state[0].nextSpellDiscount != 2 {
		t.Fatalf("Groundwork should arm a 2-mana discount, got %d", m.state[0].nextSpellDiscount)
	}
	if c := m.effectiveCost(0, getCard("skullcrack")); c != 1 {
		t.Fatalf("Groundwork should make a 3-cost spell cost 1, got %d", c)
	}
	castFrom(t, m, a, 0, "skullcrack", "") // casting a spell consumes the discount
	if m.state[0].nextSpellDiscount != 0 {
		t.Fatalf("the discount should be consumed by the next spell, got %d", m.state[0].nextSpellDiscount)
	}
}

// TestSlipAwayBouncesCheaper: Shadowstep returns a friendly minion to hand at 2
// less than its printed cost (the tempo/re-battlecry enabler).
func TestSlipAwayBouncesCheaper(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "sb", "silverback_elder", 3, 5, true) // 3-cost minion
	castFrom(t, m, a, 0, "slip_away", "sb")
	if len(m.state[0].board) != 0 {
		t.Fatal("Slip Away should return the minion to hand (board empty)")
	}
	got := -1
	for _, c := range m.state[0].hand {
		if c.ID == "silverback_elder" {
			got = c.Cost
		}
	}
	if got != 1 {
		t.Fatalf("Slip Away should return silverback_elder at cost 1 (3-2), got %d", got)
	}
}

// TestPickpocketAddsForeignCard: Pilfer adds a card that is neither the caster's
// class nor neutral — that "off-class" reach is the card's identity.
func TestPickpocketAddsForeignCard(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroPower = getCard("hone_blade") // caster's class = rogue
	m.state[0].hand = nil
	castFrom(t, m, a, 0, "pickpocket", "")
	if len(m.state[0].hand) != 1 {
		t.Fatalf("Pickpocket should add exactly one card, got %d", len(m.state[0].hand))
	}
	got := m.state[0].hand[0]
	if got.Class == cards.ClassRogue || got.Class == cards.ClassNeutral || got.Class == "" {
		t.Fatalf("Pickpocket must add a card from another class, got %s (%s)", got.ID, got.Class)
	}
}

// TestShadowVeilStealthsUntilNextTurn: Conceal hides your minions, and the Stealth
// wears off at your next turn start (it is a one-turn dodge, not permanent).
func TestShadowVeilStealthsUntilNextTurn(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "b1", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "shadow_veil", "")
	b1 := findMinion(m.state[0].board, "b1")
	if !b1.stealthed {
		t.Fatal("Shadow Veil should Stealth your minions")
	}
	m.startTurn(0) // the caster's next turn begins
	if b1.stealthed {
		t.Fatal("Shadow Veil's Stealth should expire at the caster's next turn start")
	}
}

// TestPlagueCarrierGrantsPoisonous: the Onset gives a chosen friendly minion
// Poisonous (kept even after the carrier leaves, via a granted keyword).
func TestPlagueCarrierGrantsPoisonous(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "ally", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "plague_carrier", "ally")
	if o := findMinion(m.state[0].board, "ally"); o == nil || !o.has(cards.KeywordPoisonous) {
		t.Fatalf("Plague Carrier should give the ally Poisonous, got %v", o)
	}
}

// TestRuinDaggerChainBattlecry: the weapon equips and its battlecry deals 1
// normally, 2 under Chain.
func TestRuinDaggerChainBattlecry(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "v1", "crag_ogre", 6, 9, true)
	castFrom(t, m, a, 0, "ruin_dagger", "v1")
	if w := m.state[0].weapon; w == nil || w.attack != 2 || w.durability != 2 {
		t.Fatalf("Ruin Dagger should equip a 2/2 weapon, got %v", w)
	}
	if o := findMinion(m.state[1].board, "v1"); o.health != 8 {
		t.Fatalf("Ruin Dagger no-Chain battlecry should deal 1 (9->8), got %d", o.health)
	}
	place(m, 1, "v2", "crag_ogre", 6, 9, true)
	castCombo(t, m, a, 0, "ruin_dagger", "v2", 1)
	if o := findMinion(m.state[1].board, "v2"); o.health != 7 {
		t.Fatalf("Ruin Dagger under Chain should deal 2 (9->7), got %d", o.health)
	}
}

// TestVanishingActBouncesAll: Vanish returns every minion on both boards to hand.
func TestVanishingActBouncesAll(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "f1", "crag_ogre", 6, 7, true)
	place(m, 1, "e1", "crag_ogre", 6, 7, true)
	m.state[0].hand, m.state[1].hand = nil, nil
	castFrom(t, m, a, 0, "vanishing_act", "")
	if len(m.state[0].board) != 0 || len(m.state[1].board) != 0 {
		t.Fatalf("Vanishing Act should clear both boards, got %d / %d", len(m.state[0].board), len(m.state[1].board))
	}
	if countHand(m.state[0].hand, "crag_ogre") != 1 || countHand(m.state[1].hand, "crag_ogre") != 1 {
		t.Fatal("Vanishing Act should return each minion to its owner's hand")
	}
}

// TestHoneBladeEquipsDagger: the Rogue hero power equips a 1/2 Night Shiv.
func TestHoneBladeEquipsDagger(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroPower = getCard("hone_blade")
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.sendStateAll()
	if ok, msg := m.HeroPower(a, ""); !ok {
		t.Fatalf("Hone Blade should resolve: %s", msg)
	}
	w := m.state[0].weapon
	if w == nil || w.card.ID != "night_shiv" || w.attack != 1 || w.durability != 2 {
		t.Fatalf("Hone Blade should equip a 1/2 Night Shiv, got %v", w)
	}
}

// countHand counts cards of cardID in a hand slice.
func countHand(hand []cards.Card, cardID string) int {
	n := 0
	for _, c := range hand {
		if c.ID == cardID {
			n++
		}
	}
	return n
}
