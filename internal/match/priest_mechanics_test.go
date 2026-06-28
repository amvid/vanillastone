package match

import (
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
)

// prodCard pulls a card straight from the production registry, bypassing the
// test fixtures (some ids — `mend`, `hush` — exist as both a fixture and a real
// Priest card; these tests want the real one).
func prodCard(id string) cards.Card {
	c, _ := cards.Get(id)
	return c
}

// TestDawnwardSigilBuffsAndDraws: the buff (+2 Health) and the chained draw (Then)
// both resolve — proving the generic effect-chaining path works, which several
// Priest cards rely on.
func TestDawnwardSigilBuffsAndDraws(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "f", "granite_watcher", 2, 3, true)
	m.state[0].deck = []cards.Card{getCard("granite_watcher")} // ensure the draw finds a card
	castFrom(t, m, a, 0, "dawnward_sigil", "f")
	f := findMinion(m.state[0].board, "f")
	if f == nil || f.health != 5 || f.maxHP() != 5 {
		t.Fatalf("Dawnward Sigil should give +2 Health (3->5), got %v", f)
	}
	if len(m.state[0].hand) != 1 {
		t.Fatalf("Dawnward Sigil should also draw a card, hand=%d", len(m.state[0].hand))
	}
}

// TestGloomWordAcheMaxAttack: destroys a minion with <=3 Attack, but a bigger
// minion is an illegal target — the max-attack target gate must reject, not silently
// no-op (it's the only thing stopping a 0-mana kill on a fat threat).
func TestGloomWordAcheMaxAttack(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "small", "granite_watcher", 3, 3, true)
	castFrom(t, m, a, 0, "gloom_word_ache", "small")
	if findMinion(m.state[1].board, "small") != nil {
		t.Fatal("Gloom Word: Ache should destroy a 3-Attack minion")
	}

	m2, a2, _ := newMatch()
	place(m2, 1, "big", "crag_ogre", 6, 7, true)
	m2.state[0].mana, m2.state[0].maxMana = 10, 10
	m2.state[0].hand = []cards.Card{prodCard("gloom_word_ache")}
	m2.sendStateAll()
	if ok, _ := m2.PlayCard(a2, 0, "big"); ok {
		t.Fatal("Gloom Word: Ache must reject a 6-Attack minion")
	}
}

// TestGloomWordDemiseMinAttack: destroys a minion with >=5 Attack, rejects a small one.
func TestGloomWordDemiseMinAttack(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "big", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "gloom_word_demise", "big")
	if findMinion(m.state[1].board, "big") != nil {
		t.Fatal("Gloom Word: Demise should destroy a 6-Attack minion")
	}

	m2, a2, _ := newMatch()
	place(m2, 1, "small", "granite_watcher", 2, 3, true)
	m2.state[0].mana, m2.state[0].maxMana = 10, 10
	m2.state[0].hand = []cards.Card{prodCard("gloom_word_demise")}
	m2.sendStateAll()
	if ok, _ := m2.PlayCard(a2, 0, "small"); ok {
		t.Fatal("Gloom Word: Demise must reject a 2-Attack minion")
	}
}

// TestGloomWordUndoingArea: destroys only the 5+ Attack minions across the board,
// sparing the smaller ones — the attack filter must apply per-target on an AoE.
func TestGloomWordUndoingArea(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "big", "crag_ogre", 6, 7, true)
	place(m, 1, "small", "granite_watcher", 2, 3, true)
	place(m, 0, "mybig", "war_colossus", 7, 7, true)
	castFrom(t, m, a, 0, "gloom_word_undoing", "")
	if findMinion(m.state[1].board, "big") != nil {
		t.Fatal("Undoing should destroy the enemy 6-Attack minion")
	}
	if findMinion(m.state[0].board, "mybig") != nil {
		t.Fatal("Undoing should also destroy my own 7-Attack minion (it hits ALL minions)")
	}
	if findMinion(m.state[1].board, "small") == nil {
		t.Fatal("Undoing should spare the 2-Attack minion")
	}
}

// TestRadiantBurstDamageThenHeal: deals 2 to enemy minions and heals friendly
// characters — the damage and the chained heal must both land.
func TestRadiantBurstDamageThenHeal(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 20
	place(m, 1, "e", "crag_ogre", 6, 7, true)
	place(m, 0, "f", "crag_ogre", 6, 7, true)
	m.damageMinion(findMinion(m.state[0].board, "f"), 2, "x") // 7 -> 5, now actually damaged
	castFrom(t, m, a, 0, "radiant_burst", "")
	if e := findMinion(m.state[1].board, "e"); e == nil || e.health != 5 {
		t.Fatalf("Radiant Burst should deal 2 to the enemy minion (7->5), got %v", e)
	}
	if m.state[0].heroHP != 22 {
		t.Fatalf("Radiant Burst should heal my hero 2 (22), got %d", m.state[0].heroHP)
	}
	if f := findMinion(m.state[0].board, "f"); f == nil || f.health != 7 {
		t.Fatalf("Radiant Burst should heal my minion 2 (5->7), got %v", f)
	}
}

// TestSoulKindleSetsAttackToHealth: Inner-Fire-style — Attack becomes current Health.
func TestSoulKindleSetsAttackToHealth(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "f", "granite_watcher", 2, 3, true)
	castFrom(t, m, a, 0, "soul_kindle", "f")
	if f := findMinion(m.state[0].board, "f"); f == nil || f.atk() != 3 {
		t.Fatalf("Soul Kindle should set Attack to Health (3), got %v", f)
	}
}

// TestLumenWispAttackTracksHealth: Attack always equals current Health, follows
// damage, and Silence cancels it (back to the base 0 Attack).
func TestLumenWispAttackTracksHealth(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "w", "lumen_wisp", 0, 4, true)
	w := findMinion(m.state[0].board, "w")
	if w.atk() != 4 {
		t.Fatalf("Lumen Wisp Attack should equal Health (4), got %d", w.atk())
	}
	m.damageMinion(w, 2, "x")
	if w.atk() != 2 {
		t.Fatalf("Lumen Wisp Attack should track Health after damage (2), got %d", w.atk())
	}
	m.silence(w)
	if w.atk() != 0 {
		t.Fatalf("Silence should cancel the always-equal rule (base 0), got %d", w.atk())
	}
}

// TestSoulMirrorDoublesHealth: doubles the target's current Health.
func TestSoulMirrorDoublesHealth(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "f", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "soul_mirror", "f")
	if f := findMinion(m.state[0].board, "f"); f == nil || f.health != 14 || f.maxHP() != 14 {
		t.Fatalf("Soul Mirror should double Health (7->14), got %v", f)
	}
}

// TestCrimsonSubduerDebuffPersistsUntilNextTurn: the -2 Attack lasts through the
// opponent's turn and expires at the caster's NEXT turn start — not at the caster's
// own end of turn (which is what makes the debuff actually defuse an attacker).
func TestCrimsonSubduerDebuffPersistsUntilNextTurn(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "e", "crag_ogre", 4, 7, true)
	castFrom(t, m, a, 0, "crimson_subduer", "e")
	if e := findMinion(m.state[1].board, "e"); e == nil || e.atk() != 2 {
		t.Fatalf("Crimson Subduer should give -2 Attack (4->2), got %v", e)
	}
	m.turnDuration = 0
	m.endTurnLocked() // turn 0 -> 1 (opponent's turn): debuff still active
	if e := findMinion(m.state[1].board, "e"); e == nil || e.atk() != 2 {
		t.Fatalf("debuff must persist through the opponent's turn (2), got %v", e)
	}
	m.endTurnLocked() // turn 1 -> 0 (my next turn start): debuff expires
	if e := findMinion(m.state[1].board, "e"); e == nil || e.atk() != 4 {
		t.Fatalf("debuff should expire at my next turn start (4), got %v", e)
	}
}

// TestDominateWillTakesChosenMinion: targeted mind control moves the chosen enemy
// minion to my board.
func TestDominateWillTakesChosenMinion(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "e", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "dominate_will", "e")
	if findMinion(m.state[1].board, "e") != nil {
		t.Fatal("Dominate Will should remove the minion from the enemy board")
	}
	if findMinion(m.state[0].board, "e") == nil {
		t.Fatal("Dominate Will should put the minion on my board")
	}
}

// TestCabalMindbinderMaxAttack: the onset steals a <=2 Attack enemy minion; a
// bigger one leaves it on the enemy board (the onset fizzles, the body still plays).
func TestCabalMindbinderMaxAttack(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "weak", "granite_watcher", 2, 3, true)
	castFrom(t, m, a, 0, "cabal_mindbinder", "weak")
	if findMinion(m.state[0].board, "weak") == nil {
		t.Fatal("Cabal Mindbinder should steal a 2-Attack minion")
	}

	m2, a2, _ := newMatch()
	place(m2, 1, "strong", "crag_ogre", 6, 7, true)
	m2.state[0].mana, m2.state[0].maxMana = 10, 10
	m2.state[0].hand = []cards.Card{prodCard("cabal_mindbinder")}
	m2.sendStateAll()
	m2.PlayCard(a2, 0, "strong") // onset fizzles (no legal target), body still plays
	if findMinion(m2.state[1].board, "strong") == nil {
		t.Fatal("Cabal Mindbinder must NOT steal a 6-Attack minion")
	}
}

// TestGloomThrallReturnsAtTurnEnd: temporary mind control — the stolen minion fights
// for me this turn, then returns to its owner at end of turn.
func TestGloomThrallReturnsAtTurnEnd(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "e", "granite_watcher", 3, 3, true)
	castFrom(t, m, a, 0, "gloom_thrall", "e")
	if findMinion(m.state[0].board, "e") == nil {
		t.Fatal("Gloom Thrall should give me control this turn")
	}
	m.turnDuration = 0
	m.endTurnLocked()
	if findMinion(m.state[1].board, "e") == nil {
		t.Fatal("Gloom Thrall minion should return to its owner at end of turn")
	}
	if findMinion(m.state[0].board, "e") != nil {
		t.Fatal("Gloom Thrall minion should no longer be on my board")
	}
}

// TestGreatHushSilencesAllAndDraws: silences every enemy minion and cantrips.
func TestGreatHushSilencesAllAndDraws(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "t", "bastion_golem", 3, 5, true) // a Taunt minion
	m.state[0].deck = []cards.Card{getCard("granite_watcher")}
	castFrom(t, m, a, 0, "great_hush", "")
	if tn := findMinion(m.state[1].board, "t"); tn == nil || !tn.silenced || tn.has(cards.KeywordTaunt) {
		t.Fatalf("Great Hush should silence the enemy minion, got %v", tn)
	}
	if len(m.state[0].hand) != 1 {
		t.Fatalf("Great Hush should draw a card, hand=%d", len(m.state[0].hand))
	}
}

// TestSoulreaverDevour: destroys the target minion and the played body gains its Health.
func TestSoulreaverDevour(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "prey", "crag_ogre", 6, 5, true) // 5 Health to absorb
	castFrom(t, m, a, 0, "soulreaver_nyssa", "prey")
	if findMinion(m.state[1].board, "prey") != nil {
		t.Fatal("Soulreaver should destroy the target")
	}
	// Soulreaver Nyssa is a 7/1; +5 Health -> 6.
	var nyssa *minion
	for _, mn := range m.state[0].board {
		if mn.card.ID == "soulreaver_nyssa" {
			nyssa = mn
		}
	}
	if nyssa == nil || nyssa.health != 6 || nyssa.maxHP() != 6 {
		t.Fatalf("Soulreaver should gain the prey's 5 Health (1->6), got %v", nyssa)
	}
}

// TestDawnvaleAcolyteDrawsOnMinionHealOnly: a minion heal draws a card; a hero heal
// does not (the "minion is healed" wording is the whole point of the card).
func TestDawnvaleAcolyteDrawsOnMinionHealOnly(t *testing.T) {
	// Minion heal -> draw.
	m, a, _ := newMatch()
	place(m, 0, "acolyte", "dawnvale_acolyte", 1, 3, true)
	place(m, 0, "hurt", "crag_ogre", 6, 7, true)
	m.damageMinion(findMinion(m.state[0].board, "hurt"), 4, "x") // 7 -> 3, now damaged
	m.state[0].deck = []cards.Card{getCard("granite_watcher")}
	m.state[0].hand = nil
	castFrom(t, m, a, 0, "ring_of_renewal", "") // heals all minions
	if len(m.state[0].hand) == 0 {
		t.Fatal("Dawnvale Acolyte should draw when a minion is healed")
	}

	// Hero heal only -> no draw.
	m2, a2, _ := newMatch()
	place(m2, 0, "acolyte", "dawnvale_acolyte", 1, 3, true)
	m2.state[0].heroHP = 20
	m2.state[0].hand = nil
	castFrom(t, m2, a2, 0, "mending_light", "") // heals the hero only
	if len(m2.state[0].hand) != 0 {
		t.Fatalf("Dawnvale Acolyte must NOT draw on a hero-only heal, hand=%d", len(m2.state[0].hand))
	}
}

// TestAuralastZealotFlipsHealToDamage: with the zealot in play, a "heal" deals
// damage instead — the Auchenai conversion applies to the hero power heal.
func TestAuralastZealotFlipsHealToDamage(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "z", "auralast_zealot", 3, 5, true)
	m.state[0].heroPower = prodCard("mend")
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].heroHP = 20
	if ok, msg := m.HeroPower(a, selfHeroTarget); !ok {
		t.Fatalf("Mend should resolve: %s", msg)
	}
	if m.state[0].heroHP != 18 {
		t.Fatalf("Auralast Zealot should turn the +2 heal into 2 damage (20->18), got %d", m.state[0].heroHP)
	}
}

// TestOracleVelnethDoublesSpellAndPower: doubles spell damage and hero-power healing
// while in play.
func TestOracleVelnethDoublesSpellAndPower(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "velen", "oracle_velneth", 7, 7, true)
	place(m, 1, "e", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "searing_light", "e")
	if e := findMinion(m.state[1].board, "e"); e == nil || e.health != 1 {
		t.Fatalf("Velen should double Searing Light to 6 (7->1), got %v", e)
	}
	// Hero-power heal doubled: Mend +2 -> +4.
	m.state[0].heroPower = prodCard("mend")
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].heroPowerUsed = false
	m.state[0].heroHP = 20
	if ok, msg := m.HeroPower(a, selfHeroTarget); !ok {
		t.Fatalf("Mend should resolve: %s", msg)
	}
	if m.state[0].heroHP != 24 {
		t.Fatalf("Velen should double Mend to +4 (20->24), got %d", m.state[0].heroHP)
	}
}

// TestUmbralShiftSwapsHeroPower: Shadowform-style — the hero power becomes the
// 2-damage Gloom Spike and can fire immediately.
func TestUmbralShiftSwapsHeroPower(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroPower = prodCard("mend")
	castFrom(t, m, a, 0, "umbral_shift", "")
	if m.state[0].heroPower.ID != "gloom_spike" {
		t.Fatalf("Umbral Shift should swap the hero power to gloom_spike, got %q", m.state[0].heroPower.ID)
	}
	m.state[0].mana, m.state[0].maxMana = 10, 10
	place(m, 1, "e", "crag_ogre", 6, 7, true)
	if ok, msg := m.HeroPower(a, "e"); !ok {
		t.Fatalf("Gloom Spike should resolve: %s", msg)
	}
	if e := findMinion(m.state[1].board, "e"); e == nil || e.health != 5 {
		t.Fatalf("Gloom Spike should deal 2 (7->5), got %v", e)
	}
}

// TestCopyEffectsPullFromOpponent: hand-copy and deck-copy add the opponent's cards
// to my hand; the originals stay with the opponent.
func TestCopyEffectsPullFromOpponent(t *testing.T) {
	// Pried Thought: copy from the opponent's HAND.
	m, a, _ := newMatch()
	m.state[1].hand = []cards.Card{getCard("crag_ogre")}
	castFrom(t, m, a, 0, "pried_thought", "")
	if len(m.state[0].hand) != 1 {
		t.Fatalf("Pried Thought should copy 1 card into my hand, got %d", len(m.state[0].hand))
	}
	if len(m.state[1].hand) != 1 {
		t.Fatal("Pried Thought must not remove the card from the opponent's hand")
	}

	// Mind Larceny: copy 2 from the opponent's DECK.
	m2, a2, _ := newMatch()
	m2.state[1].deck = []cards.Card{getCard("crag_ogre"), getCard("granite_watcher")}
	m2.state[0].hand = nil
	castFrom(t, m2, a2, 0, "mind_larceny", "")
	if len(m2.state[0].hand) != 2 {
		t.Fatalf("Mind Larceny should copy 2 cards, got %d", len(m2.state[0].hand))
	}
	if len(m2.state[1].deck) != 2 {
		t.Fatal("Mind Larceny must not remove the cards from the opponent's deck")
	}
}

// TestPhantomSummonsFromOppDeck: summons a copy of a random minion from the
// opponent's deck onto my board.
func TestPhantomSummonsFromOppDeck(t *testing.T) {
	m, a, _ := newMatch()
	m.state[1].deck = []cards.Card{getCard("crag_ogre")}
	m.state[0].board = nil
	castFrom(t, m, a, 0, "phantom_summons", "")
	if len(m.state[0].board) != 1 || m.state[0].board[0].card.ID != "crag_ogre" {
		t.Fatalf("Phantom Summons should summon a copy of the enemy deck minion, got %v", m.state[0].board)
	}
	if len(m.state[1].deck) != 1 {
		t.Fatal("Phantom Summons must not remove the minion from the opponent's deck")
	}
}

// TestPrismMothDoubleHealthOnlyWhenAllOdd: doubles other friendly minions' Health
// only when the deck is all odd-cost.
func TestPrismMothDoubleHealthOnlyWhenAllOdd(t *testing.T) {
	// All-odd deck -> doubles.
	m, a, _ := newMatch()
	place(m, 0, "ally", "crag_ogre", 6, 7, true)
	m.state[0].deck = []cards.Card{prodCard("searing_light")} // cost 1 (odd)
	castFrom(t, m, a, 0, "prism_moth", "")
	if ally := findMinion(m.state[0].board, "ally"); ally == nil || ally.health != 14 {
		t.Fatalf("Prism Moth should double the ally's Health (7->14) with an all-odd deck, got %v", ally)
	}

	// Even card present -> no doubling.
	m2, a2, _ := newMatch()
	place(m2, 0, "ally", "crag_ogre", 6, 7, true)
	m2.state[0].deck = []cards.Card{prodCard("crag_ogre")} // cost 6 (even)
	castFrom(t, m2, a2, 0, "prism_moth", "")
	if ally := findMinion(m2.state[0].board, "ally"); ally == nil || ally.health != 7 {
		t.Fatalf("Prism Moth should NOT double with an even-cost card in deck, got %v", ally)
	}
}

// TestRadiantFontHealsAtTurnStart: the start-of-turn trigger restores Health to a
// damaged friendly character.
func TestRadiantFontHealsAtTurnStart(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "font", "radiant_font", 0, 5, true)
	place(m, 0, "hurt", "crag_ogre", 6, 7, true)
	m.damageMinion(findMinion(m.state[0].board, "hurt"), 5, "x") // 7 -> 2, now damaged
	m.turn = 1
	m.turnDuration = 0
	m.endTurnLocked() // -> turn 0 start: the font fires
	if h := findMinion(m.state[0].board, "hurt"); h == nil || h.health != 5 {
		t.Fatalf("Radiant Font should restore 3 to the damaged minion (2->5), got %v", h)
	}
}
