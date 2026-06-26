package match

import (
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
)

// TestSoulTitheDrawsAndSelfDamages: the Warlock hero power (Life Tap) draws a card
// and deals 2 damage to its own hero.
func TestSoulTitheDrawsAndSelfDamages(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroPower = getCard("soul_tithe")
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].heroHP = 30
	m.state[0].hand = nil
	m.state[0].deck = []cards.Card{getCard("shadow_lance")}
	if ok, msg := m.HeroPower(a, ""); !ok {
		t.Fatalf("Soul Tithe should resolve: %s", msg)
	}
	if m.state[0].heroHP != 28 {
		t.Fatalf("Soul Tithe should deal 2 to own hero (28), got %d", m.state[0].heroHP)
	}
	if len(m.state[0].hand) != 1 {
		t.Fatalf("Soul Tithe should draw 1, hand=%d", len(m.state[0].hand))
	}
}

// TestEmberImpSelfDamage: Ember Imp's onset deals 3 to its own hero.
func TestEmberImpSelfDamage(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 30
	castFrom(t, m, a, 0, "ember_imp", "")
	if m.state[0].heroHP != 27 {
		t.Fatalf("Ember Imp should deal 3 to own hero (27), got %d", m.state[0].heroHP)
	}
	if findMinion(m.state[0].board, m.state[0].board[len(m.state[0].board)-1].uid) == nil {
		t.Fatal("Ember Imp should still be on the board")
	}
}

// TestSoulEmberDamageAndDiscard: Soul Ember deals 4 and discards one random card
// from the caster's hand.
func TestSoulEmberDamageAndDiscard(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "x", "marsh_snapjaw", 2, 7, true)
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].hand = []cards.Card{getCard("soul_ember"), getCard("shadow_lance"), getCard("hexfire")}
	m.sendStateAll()
	if ok, msg := m.PlayCard(a, 0, "x"); !ok { // play Soul Ember at index 0
		t.Fatalf("Soul Ember should resolve: %s", msg)
	}
	if x := findMinion(m.state[1].board, "x"); x == nil || x.health != 3 {
		t.Fatalf("Soul Ember should deal 4 (7->3), got %v", x)
	}
	// 3 in hand - 1 played - 1 discarded = 1 left.
	if len(m.state[0].hand) != 1 {
		t.Fatalf("Soul Ember should discard one card (hand=1), got %d", len(m.state[0].hand))
	}
}

// TestGnawingFiendDiscards: Gnawing Fiend's onset discards a random card from hand.
func TestGnawingFiendDiscards(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].hand = []cards.Card{getCard("gnawing_fiend"), getCard("shadow_lance"), getCard("hexfire")}
	m.sendStateAll()
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("Gnawing Fiend should resolve: %s", msg)
	}
	// 3 in hand - 1 played - 1 discarded = 1 left.
	if len(m.state[0].hand) != 1 {
		t.Fatalf("Gnawing Fiend should discard one card (hand=1), got %d", len(m.state[0].hand))
	}
}

// TestMortalWhisperDrawsOnKill: Mortal Whisper draws only when its 1 damage kills.
func TestMortalWhisperDrawsOnKill(t *testing.T) {
	// Kills -> draw.
	m, a, _ := newMatch()
	place(m, 1, "weak", "granite_watcher", 2, 1, true)
	castFrom(t, m, a, 0, "mortal_whisper", "weak")
	if len(m.state[0].hand) != 1 {
		t.Fatalf("Mortal Whisper kill should draw 1, hand=%d", len(m.state[0].hand))
	}
	// Survives -> no draw.
	m2, a2, _ := newMatch()
	place(m2, 1, "tough", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m2, a2, 0, "mortal_whisper", "tough")
	if len(m2.state[0].hand) != 0 {
		t.Fatalf("Mortal Whisper survivor should not draw, hand=%d", len(m2.state[0].hand))
	}
}

// TestSiphonVitaeLifesteal: Siphon Vitae deals 2 and heals the caster's hero by 2.
func TestSiphonVitaeLifesteal(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 20
	place(m, 1, "x", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "siphon_vitae", "x")
	if x := findMinion(m.state[1].board, "x"); x == nil || x.health != 5 {
		t.Fatalf("Siphon Vitae should deal 2 (7->5), got %v", x)
	}
	if m.state[0].heroHP != 22 {
		t.Fatalf("Siphon Vitae should heal hero 2 (22), got %d", m.state[0].heroHP)
	}
}

// TestHexfireDamageOrBuff: Hexfire damages a minion, but buffs a friendly Demon
// +2/+2 instead.
func TestHexfireDamageOrBuff(t *testing.T) {
	// Friendly Demon -> +2/+2.
	m, a, _ := newMatch()
	place(m, 0, "imp", "runt_imp", 1, 1, true)
	castFrom(t, m, a, 0, "hexfire", "imp")
	if imp := findMinion(m.state[0].board, "imp"); imp == nil || imp.atk() != 3 || imp.maxHP() != 3 {
		t.Fatalf("Hexfire on a friendly Demon should make it 3/3, got %v", imp)
	}
	// Enemy minion -> 2 damage.
	m2, a2, _ := newMatch()
	place(m2, 1, "x", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m2, a2, 0, "hexfire", "x")
	if x := findMinion(m2.state[1].board, "x"); x == nil || x.health != 5 {
		t.Fatalf("Hexfire on an enemy should deal 2 (7->5), got %v", x)
	}
}

// TestDarkBargainDestroysDemonAndHeals: Dark Bargain destroys a friendly Demon and
// restores 5 Health; it can't target a non-Demon.
func TestDarkBargainDestroysDemonAndHeals(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 20
	place(m, 0, "imp", "runt_imp", 1, 1, true)
	castFrom(t, m, a, 0, "dark_bargain", "imp")
	if findMinion(m.state[0].board, "imp") != nil {
		t.Fatal("Dark Bargain should destroy the friendly Demon")
	}
	if m.state[0].heroHP != 25 {
		t.Fatalf("Dark Bargain should heal 5 (25), got %d", m.state[0].heroHP)
	}
	// A non-Demon friendly minion is not a legal target.
	m2, a2, _ := newMatch()
	place(m2, 0, "g", "granite_watcher", 2, 3, true)
	m2.state[0].mana, m2.state[0].maxMana = 10, 10
	m2.state[0].hand = []cards.Card{getCard("dark_bargain")}
	if ok, _ := m2.PlayCard(a2, 0, "g"); ok {
		t.Fatal("Dark Bargain should reject a non-Demon target")
	}
}

// TestSoulHarvestDestroysAndHeals: Soul Harvest destroys any minion and restores 3.
func TestSoulHarvestDestroysAndHeals(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 20
	place(m, 1, "big", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "soul_harvest", "big")
	if findMinion(m.state[1].board, "big") != nil {
		t.Fatal("Soul Harvest should destroy the minion")
	}
	if m.state[0].heroHP != 23 {
		t.Fatalf("Soul Harvest should heal 3 (23), got %d", m.state[0].heroHP)
	}
}

// TestDoomKissSummonsOnKill: Doom Kiss deals 2 and, when it kills a minion, summons
// a random Demon for the caster.
func TestDoomKissSummonsOnKill(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "x", "granite_watcher", 2, 2, true)
	before := len(m.state[0].board)
	castFrom(t, m, a, 0, "doom_kiss", "x")
	if findMinion(m.state[1].board, "x") != nil {
		t.Fatal("Doom Kiss should kill a 2-Health minion")
	}
	if len(m.state[0].board) != before+1 {
		t.Fatalf("Doom Kiss kill should summon a Demon for the caster, board %d->%d", before, len(m.state[0].board))
	}
	if mn := m.state[0].board[len(m.state[0].board)-1]; mn.card.Tribe != cards.TribeDemon {
		t.Fatalf("Doom Kiss should summon a Demon, got tribe %q", mn.card.Tribe)
	}
}

// TestForbiddenMightBuffsThenDies: Forbidden Might gives +4/+4 then the minion dies
// at end of turn.
func TestForbiddenMightBuffsThenDies(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "g", "granite_watcher", 2, 3, true)
	castFrom(t, m, a, 0, "forbidden_might", "g")
	if g := findMinion(m.state[0].board, "g"); g == nil || g.atk() != 6 || g.maxHP() != 7 {
		t.Fatalf("Forbidden Might should make a 6/7, got %v", g)
	}
	m.EndTurn(a)
	if findMinion(m.state[0].board, "g") != nil {
		t.Fatal("Forbidden Might minion should die at end of turn")
	}
}

// TestChainedBruteLosesManaCrystal: Chained Brute's onset destroys one Mana Crystal.
func TestChainedBruteLosesManaCrystal(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].maxMana, m.state[0].mana = 10, 10
	m.state[0].hand = []cards.Card{getCard("chained_brute")}
	m.sendStateAll()
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("Chained Brute should resolve: %s", msg)
	}
	if m.state[0].maxMana != 9 {
		t.Fatalf("Chained Brute should destroy a Mana Crystal (maxMana 9), got %d", m.state[0].maxMana)
	}
}

// TestRaveningHorrorConsumesAdjacent: Ravening Horror destroys both neighbours and
// gains their combined Attack/Health.
func TestRaveningHorrorConsumesAdjacent(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "L", "granite_watcher", 2, 3, true)
	place(m, 0, "R", "marsh_snapjaw", 4, 5, true)
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].hand = []cards.Card{getCard("ravening_horror")}
	m.sendStateAll()
	if ok, msg := m.PlayCardAt(a, 0, "", 1); !ok { // insert between L and R
		t.Fatalf("Ravening Horror should resolve: %s", msg)
	}
	if findMinion(m.state[0].board, "L") != nil || findMinion(m.state[0].board, "R") != nil {
		t.Fatal("Ravening Horror should destroy both adjacent minions")
	}
	rh := m.state[0].board[len(m.state[0].board)-1]
	// 3/3 base + (2+4 atk, 3+5 hp) = 9/11.
	if rh.atk() != 9 || rh.maxHP() != 11 {
		t.Fatalf("Ravening Horror should gain neighbours' stats (9/11), got %d/%d", rh.atk(), rh.maxHP())
	}
}

// TestGloomflareSacrificeAoE: Gloomflare destroys a friendly minion and deals its
// Attack to every enemy minion.
func TestGloomflareSacrificeAoE(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "fuel", "marsh_snapjaw", 4, 5, true) // 4 Attack
	place(m, 1, "e1", "marsh_snapjaw", 2, 7, true)
	place(m, 1, "e2", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "gloomflare", "fuel")
	if findMinion(m.state[0].board, "fuel") != nil {
		t.Fatal("Gloomflare should destroy the friendly minion")
	}
	for _, id := range []string{"e1", "e2"} {
		if e := findMinion(m.state[1].board, id); e == nil || e.health != 3 {
			t.Fatalf("Gloomflare should deal 4 to each enemy (7->3), %s=%v", id, e)
		}
	}
}

// TestTheUnmakingDestroysAllMinions: The Unmaking wipes every minion on both boards.
func TestTheUnmakingDestroysAllMinions(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "f", "granite_watcher", 2, 3, true)
	place(m, 1, "e", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "the_unmaking", "")
	if len(m.state[0].board) != 0 || len(m.state[1].board) != 0 {
		t.Fatalf("The Unmaking should clear both boards, got %d/%d", len(m.state[0].board), len(m.state[1].board))
	}
}

// TestCreepingRotDelayedDestroy: Creeping Rot destroys the enemy minion at the start
// of the caster's NEXT turn (it survives the opponent's turn first).
func TestCreepingRotDelayedDestroy(t *testing.T) {
	m, a, b := newMatch()
	place(m, 1, "x", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "creeping_rot", "x")
	if findMinion(m.state[1].board, "x") == nil {
		t.Fatal("Creeping Rot should not destroy immediately")
	}
	m.EndTurn(a) // -> opponent's turn; the minion still lives
	if findMinion(m.state[1].board, "x") == nil {
		t.Fatal("Creeping Rot target should survive the opponent's turn")
	}
	m.EndTurn(b) // -> back to caster; it dies at the caster's turn start
	if findMinion(m.state[1].board, "x") != nil {
		t.Fatal("Creeping Rot should destroy the minion at the caster's next turn start")
	}
}

// TestDreadColossusHitsAllOthers: Dread Colossus's onset deals 1 to every other
// character, never itself.
func TestDreadColossusHitsAllOthers(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP, m.state[1].heroHP = 30, 30
	place(m, 1, "e", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "dread_colossus", "")
	if m.state[0].heroHP != 29 || m.state[1].heroHP != 29 {
		t.Fatalf("Dread Colossus should hit both heroes for 1, got %d/%d", m.state[0].heroHP, m.state[1].heroHP)
	}
	if e := findMinion(m.state[1].board, "e"); e == nil || e.health != 6 {
		t.Fatalf("Dread Colossus should deal 1 to the enemy minion (7->6), got %v", e)
	}
	col := m.state[0].board[len(m.state[0].board)-1]
	if col.health != col.maxHP() {
		t.Fatalf("Dread Colossus should not damage itself, got %d/%d", col.health, col.maxHP())
	}
}

// TestDreadWardenDemonAura: Dread Warden gives the caster's OTHER Demons +1 Attack,
// not non-Demons and not itself.
func TestDreadWardenDemonAura(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "warden", "dread_warden", 5, 8, true)
	place(m, 0, "imp", "runt_imp", 1, 1, true)         // Demon -> +1
	place(m, 0, "rock", "granite_watcher", 2, 3, true) // non-Demon -> unaffected
	m.refreshAuras()
	if imp := findMinion(m.state[0].board, "imp"); imp.atk() != 2 {
		t.Fatalf("Dread Warden should give a friendly Demon +1 Attack (2), got %d", imp.atk())
	}
	if rock := findMinion(m.state[0].board, "rock"); rock.atk() != 2 {
		t.Fatalf("Dread Warden should not buff a non-Demon, got %d", rock.atk())
	}
	if warden := findMinion(m.state[0].board, "warden"); warden.atk() != 5 {
		t.Fatalf("Dread Warden should not buff itself, got %d", warden.atk())
	}
}

// TestGloomImpEndOfTurnBuff: Gloom Imp gives another random friendly minion +1
// Health at the end of the owner's turn.
func TestGloomImpEndOfTurnBuff(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "imp", "gloom_imp", 0, 1, true)
	place(m, 0, "ally", "granite_watcher", 2, 3, true)
	m.EndTurn(a)
	if ally := findMinion(m.state[0].board, "ally"); ally == nil || ally.maxHP() != 4 {
		t.Fatalf("Gloom Imp should give the ally +1 Health (max 4), got %v", ally)
	}
}

// TestDarkSummonsAddsDemon: Dark Summons adds a random Demon minion to hand.
func TestDarkSummonsAddsDemon(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = nil
	castFrom(t, m, a, 0, "dark_summons", "")
	// castFrom set the hand to just the spell; after playing it, one Demon is added.
	if len(m.state[0].hand) != 1 {
		t.Fatalf("Dark Summons should add one card, hand=%d", len(m.state[0].hand))
	}
	if c := m.state[0].hand[0]; c.Type != cards.TypeMinion || c.Tribe != cards.TribeDemon {
		t.Fatalf("Dark Summons should add a Demon minion, got %s/%s", c.Type, c.Tribe)
	}
}

// TestCallTheBroodTutorsDemons: Call the Brood draws two Demons from the deck, and
// adds Runt Imp tokens when the deck has no Demons left.
func TestCallTheBroodTutorsDemons(t *testing.T) {
	// Deck has 2 Demons -> both drawn.
	m, a, _ := newMatch()
	m.state[0].hand = nil
	m.state[0].deck = []cards.Card{getCard("ember_imp"), getCard("hollow_guardian"), getCard("shadow_lance")}
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].hand = []cards.Card{getCard("call_the_brood")}
	m.sendStateAll()
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("Call the Brood should resolve: %s", msg)
	}
	demons := 0
	for _, c := range m.state[0].hand {
		if c.Tribe == cards.TribeDemon {
			demons++
		}
	}
	if demons != 2 {
		t.Fatalf("Call the Brood should draw 2 Demons, got %d", demons)
	}
	if len(m.state[0].deck) != 1 || m.state[0].deck[0].ID != "shadow_lance" {
		t.Fatalf("Call the Brood should leave only the non-Demon in deck, got %v", m.state[0].deck)
	}
	// No Demons in deck -> two Runt Imp fallbacks.
	m2, a2, _ := newMatch()
	m2.state[0].deck = nil
	m2.state[0].mana, m2.state[0].maxMana = 10, 10
	m2.state[0].hand = []cards.Card{getCard("call_the_brood")}
	m2.sendStateAll()
	if ok, msg := m2.PlayCard(a2, 0, ""); !ok {
		t.Fatalf("Call the Brood should resolve: %s", msg)
	}
	if countCardInHand(m2.state[0].hand, "runt_imp") != 2 {
		t.Fatalf("Call the Brood with no Demons should add 2 Runt Imps, got %d", countCardInHand(m2.state[0].hand, "runt_imp"))
	}
}

// TestDarkGatewayCostFloor: Dark Gateway makes the caster's minions cost 2 less,
// floored at 1 (never 0).
func TestDarkGatewayCostFloor(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "portal", "dark_gateway", 0, 4, true)
	// A 6-cost minion -> 4; a 2-cost minion -> 1 (floor), not 0.
	big := getCard("marsh_snapjaw") // 6-cost neutral minion
	if got := m.effectiveCost(0, big); got != big.Cost-2 {
		t.Fatalf("Dark Gateway should reduce a 6-drop by 2 (%d), got %d", big.Cost-2, got)
	}
	twoDrop := getCard("gnawing_fiend") // 2-cost minion -> floored at 1, not 0
	if got := m.effectiveCost(0, twoDrop); got != 1 {
		t.Fatalf("Dark Gateway should floor a 2-drop at 1, got %d", got)
	}
}

// TestOverlordXathulReplacesHero: Overlord Xathul replaces the hero (15 Health,
// new hero power, equipped weapon) and the played minion does not stay on board.
// The new hero power then summons a 6/6 Demon.
func TestOverlordXathulReplacesHero(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 22
	castFrom(t, m, a, 0, "overlord_xathul", "")
	ps := m.state[0]
	if ps.heroHP != 15 {
		t.Fatalf("Overlord Xathul should set Health to 15, got %d", ps.heroHP)
	}
	if ps.heroPower.ID != "infernal_eruption" {
		t.Fatalf("Overlord Xathul should swap the hero power, got %q", ps.heroPower.ID)
	}
	if ps.heroArt != "overlord_xathul_hero" {
		t.Fatalf("Overlord Xathul should set the hero portrait art, got %q", ps.heroArt)
	}
	if ps.weapon == nil || ps.weapon.attack != 3 || ps.weapon.durability != 8 {
		t.Fatalf("Overlord Xathul should equip a 3/8 weapon, got %v", ps.weapon)
	}
	if countCard(ps.board, "overlord_xathul") != 0 {
		t.Fatal("Overlord Xathul should not remain on the board")
	}
	// New hero power: summon a 6/6 Demon.
	ps.mana, ps.maxMana = 10, 10
	ps.heroPowerUsed = false
	if ok, msg := m.HeroPower(a, ""); !ok {
		t.Fatalf("Infernal Eruption should resolve: %s", msg)
	}
	last := ps.board[len(ps.board)-1]
	if last.card.ID != "abyss_horror" || last.atk() != 6 || last.maxHP() != 6 {
		t.Fatalf("Infernal Eruption should summon a 6/6 Abyss Horror, got %v", last)
	}
}

// countCardInHand counts hand cards with the given id.
func countCardInHand(hand []cards.Card, id string) int {
	n := 0
	for _, c := range hand {
		if c.ID == id {
			n++
		}
	}
	return n
}
