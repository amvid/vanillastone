package match

import (
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
)

// castFrom sets the caster's mana and hand to the single card, refreshes the
// snapshot, and plays it at targetID ("" for untargeted). Fails the test if the
// play is rejected.
func castFrom(t *testing.T, m *Match, f *fakeSender, pi int, cardID, targetID string) {
	t.Helper()
	m.state[pi].mana, m.state[pi].maxMana = 10, 10
	m.state[pi].hand = []cards.Card{getCard(cardID)}
	m.sendStateAll()
	if ok, msg := m.PlayCard(f, 0, targetID); !ok {
		t.Fatalf("%s should resolve: %s", cardID, msg)
	}
}

func countCard(board []*minion, cardID string) int {
	n := 0
	for _, mn := range board {
		if mn.card.ID == cardID {
			n++
		}
	}
	return n
}

// --- Wave B: small extensions ---

// TestQuarryBrandSetsHealthToOne: Quarry Brand changes a minion's Health (max +
// current) to 1.
func TestQuarryBrandSetsHealthToOne(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "ogre", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "quarry_brand", "ogre")
	ogre := findMinion(m.state[1].board, "ogre")
	if ogre == nil {
		t.Fatal("target should survive (health 1, not dead)")
	}
	if ogre.health != 1 || ogre.maxHP() != 1 {
		t.Fatalf("Quarry Brand should set Health to 1, got %d/%d", ogre.atk(), ogre.maxHP())
	}
}

// TestFeralCommandConditional: Feral Command deals 3 normally, 5 when the caster
// controls a Beast.
func TestFeralCommandConditional(t *testing.T) {
	// No Beast controlled -> 3 damage.
	m, a, _ := newMatch()
	castFrom(t, m, a, 0, "feral_command", oppHeroTarget)
	if hp := m.state[1].heroHP; hp != 27 {
		t.Fatalf("Feral Command without a Beast should deal 3 (30->27), got %d", hp)
	}
	// Beast controlled -> 5 damage.
	m2, a2, _ := newMatch()
	place(m2, 0, "wolf", "packleader_wolf", 1, 1, true) // a Beast
	castFrom(t, m2, a2, 0, "feral_command", oppHeroTarget)
	if hp := m2.state[1].heroHP; hp != 25 {
		t.Fatalf("Feral Command with a Beast should deal 5 (30->25), got %d", hp)
	}
}

// TestVolleyShotHitsTwoRandomEnemies: Volley Shot deals 3 to exactly two enemy
// minions (no more, no less) when three are available.
func TestVolleyShotHitsTwoRandomEnemies(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "m1", "marsh_snapjaw", 2, 7, true)
	place(m, 1, "m2", "marsh_snapjaw", 2, 7, true)
	place(m, 1, "m3", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "volley_shot", "")
	hit := 0
	for _, mn := range m.state[1].board {
		if mn.health == 4 { // 7 - 3
			hit++
		} else if mn.health != 7 {
			t.Fatalf("a minion took the wrong damage, health=%d", mn.health)
		}
	}
	if hit != 2 {
		t.Fatalf("Volley Shot should hit exactly two enemy minions, hit %d", hit)
	}
}

// TestCullingShotDestroysAnEnemyMinion: Culling Shot destroys the (only) enemy
// minion.
func TestCullingShotDestroysAnEnemyMinion(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "big", "war_colossus", 7, 7, true)
	castFrom(t, m, a, 0, "culling_shot", "")
	if len(m.state[1].board) != 0 {
		t.Fatalf("Culling Shot should destroy the enemy minion, board=%d", len(m.state[1].board))
	}
}

// --- Wave D: auras / keywords / multi-effect / seek ---

// TestScoutAheadSeeksFromDeck: Scout Ahead offers the top 3 of the caster's own
// deck (removing them); the pick goes to hand, the rest are discarded.
func TestScoutAheadSeeksFromDeck(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].deck = testDeck([]string{"river_snapper", "ironfur_bear", "mirefang_raptor", "fang_alpha"})
	castFrom(t, m, a, 0, "scout_ahead", "")
	if m.pending == nil || len(m.pending.options) != 3 {
		t.Fatalf("Scout Ahead should present 3 options from the deck, got %v", m.pending)
	}
	if len(m.state[0].deck) != 1 {
		t.Fatalf("the 3 offered cards should leave the deck (4->1), got %d", len(m.state[0].deck))
	}
	want := m.pending.options[1].ID
	if ok, msg := m.Choose(a, 1); !ok {
		t.Fatalf("Choose should resolve: %s", msg)
	}
	if n := countCardHand(m.state[0].hand, want); n != 1 {
		t.Fatalf("the chosen card should be in hand once, got %d", n)
	}
}

func countCardHand(hand []cards.Card, id string) int {
	n := 0
	for _, c := range hand {
		if c.ID == id {
			n++
		}
	}
	return n
}

// TestSignalFlareClearsStealthSecretsDraws: Signal Flare strips Stealth from all
// minions, destroys every enemy Secret, and draws a card.
func TestSignalFlareClearsStealthSecretsDraws(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "sneak", "jungle_stalker", 5, 5, true) // a Stealth minion
	if !findMinion(m.state[1].board, "sneak").stealthed {
		t.Fatal("setup: minion should start Stealthed")
	}
	placeSecret(m, 1, "blasting_snare")
	placeSecret(m, 1, "marksman_trap")
	m.state[0].deck = testDeck([]string{"river_snapper", "fang_alpha"})
	castFrom(t, m, a, 0, "signal_flare", "")
	if findMinion(m.state[1].board, "sneak").stealthed {
		t.Fatal("Signal Flare should strip Stealth")
	}
	if len(m.state[1].secrets) != 0 {
		t.Fatalf("Signal Flare should destroy all enemy Secrets, %d left", len(m.state[1].secrets))
	}
	if len(m.state[0].hand) != 1 { // spell left hand (0), then drew 1
		t.Fatalf("Signal Flare should draw a card, hand=%d", len(m.state[0].hand))
	}
}

// TestUnleashSummonsPerEnemyMinion: Unleash the Pack summons one 1/1 Charge Hound
// per enemy minion.
func TestUnleashSummonsPerEnemyMinion(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "e1", "river_snapper", 2, 3, true)
	place(m, 1, "e2", "river_snapper", 2, 3, true)
	castFrom(t, m, a, 0, "unleash_the_pack", "")
	if n := countCard(m.state[0].board, "snarling_hound"); n != 2 {
		t.Fatalf("Unleash should summon one Hound per enemy minion (2), got %d", n)
	}
	hound := m.state[0].board[0]
	if !hound.has(cards.KeywordCharge) {
		t.Fatal("the summoned Hounds should have Charge")
	}
}

// TestBestialFuryBuffAndImmune: Bestial Fury gives a friendly Beast +2 Attack and
// Immune this turn; both expire at end of turn.
func TestBestialFuryBuffAndImmune(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "beast", "guardian_bear", 4, 4, true)
	castFrom(t, m, a, 0, "bestial_fury", "beast")
	beast := findMinion(m.state[0].board, "beast")
	if beast.atk() != 6 {
		t.Fatalf("Bestial Fury should give +2 Attack (4->6), got %d", beast.atk())
	}
	if !beast.has(cards.KeywordImmune) {
		t.Fatal("Bestial Fury should grant Immune")
	}
	if dealt := m.damageMinion(beast, 5, "x"); dealt != 0 || beast.health != 4 {
		t.Fatalf("an Immune minion should ignore damage, dealt=%d health=%d", dealt, beast.health)
	}
	m.EndTurn(a) // temp buff + Immune expire
	beast = findMinion(m.state[0].board, "beast")
	if beast.atk() != 4 || beast.has(cards.KeywordImmune) {
		t.Fatalf("Bestial Fury should expire at end of turn, atk=%d immune=%v", beast.atk(), beast.has(cards.KeywordImmune))
	}
}

// TestBestialFuryRequiresBeast: Bestial Fury cannot target a non-Beast.
func TestBestialFuryRequiresBeast(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "guy", "ironforge_brute", 3, 3, true) // not a Beast
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].hand = []cards.Card{getCard("bestial_fury")}
	m.sendStateAll()
	if ok, _ := m.PlayCard(a, 0, "guy"); ok {
		t.Fatal("Bestial Fury must reject a non-Beast target")
	}
}

// TestBlastingShotSplash: Blasting Shot deals 5 to the target minion and 2 to its
// neighbours.
func TestBlastingShotSplash(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "l", "war_colossus", 7, 7, true)
	place(m, 1, "mid", "war_colossus", 7, 7, true)
	place(m, 1, "r", "war_colossus", 7, 7, true)
	castFrom(t, m, a, 0, "blasting_shot", "mid")
	if mid := findMinion(m.state[1].board, "mid"); mid == nil || mid.health != 2 {
		t.Fatalf("Blasting Shot should deal 5 to the target (7->2), got %v", mid)
	}
	for _, uid := range []string{"l", "r"} {
		if n := findMinion(m.state[1].board, uid); n == nil || n.health != 5 {
			t.Fatalf("Blasting Shot should deal 2 to neighbour %s (7->5), got %v", uid, n)
		}
	}
}

// TestTundraChargerGrantsCharge: Tundra Charger gives friendly Beasts Charge via
// its aura, so a Beast summoned this turn can attack immediately.
func TestTundraChargerGrantsCharge(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "rhino", "tundra_charger", 2, 5, true)
	beast := m.summonMinion(0, getCard("mirefang_raptor")) // summon-sick Beast
	m.refreshAuras()
	if !beast.has(cards.KeywordCharge) {
		t.Fatal("Tundra Charger should grant Charge to friendly Beasts")
	}
	if !m.canAttack(beast) {
		t.Fatal("a Beast with the granted Charge should be able to attack the turn it is summoned")
	}
}

// --- Wave C: secrets / traps ---

// TestBlastingSnareDamagesAllEnemies: when the owner's hero is attacked, Blasting
// Snare deals 2 to every enemy character; the attack still lands.
func TestBlastingSnareDamagesAllEnemies(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "war_colossus", 7, 7, true) // attacker (survives the 2 AoE)
	place(m, 0, "ally", "war_colossus", 3, 5, true)
	placeSecret(m, 1, "blasting_snare")
	if ok, msg := m.Attack(a, "atk", oppHeroTarget); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	if hp := m.state[0].heroHP; hp != 28 {
		t.Fatalf("Blasting Snare should deal 2 to the attacking hero (30->28), got %d", hp)
	}
	if ally := findMinion(m.state[0].board, "ally"); ally == nil || ally.health != 3 {
		t.Fatalf("Blasting Snare should deal 2 to enemy minions (5->3), got %v", ally)
	}
	if atk := findMinion(m.state[0].board, "atk"); atk == nil || atk.health != 5 {
		t.Fatalf("the attacker should take the 2 AoE (7->5), got %v", atk)
	}
	if hp := m.state[1].heroHP; hp != 23 { // 30 - 7 (attack still lands, non-cancel)
		t.Fatalf("Blasting Snare does not cancel; defender hero should take the 7, got %d", hp)
	}
}

// TestSnaringTrapBouncesAttacker: Snaring Trap returns the attacker to hand at +2
// cost and cancels the attack.
func TestSnaringTrapBouncesAttacker(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = nil                             // room for the bounced card (the fixture hand is over the cap)
	place(m, 0, "atk", "mirefang_raptor", 3, 2, true) // cost 2 Beast
	placeSecret(m, 1, "snaring_trap")
	if ok, msg := m.Attack(a, "atk", oppHeroTarget); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	if findMinion(m.state[0].board, "atk") != nil {
		t.Fatal("Snaring Trap should remove the attacker from the board")
	}
	if m.state[1].heroHP != 30 {
		t.Fatalf("Snaring Trap cancels the attack; hero should be unharmed, got %d", m.state[1].heroHP)
	}
	if countCardHand(m.state[0].hand, "mirefang_raptor") != 1 {
		t.Fatal("the bounced minion should return to hand")
	}
	for _, c := range m.state[0].hand {
		if c.ID == "mirefang_raptor" && c.Cost != getCard("mirefang_raptor").Cost+2 {
			t.Fatalf("the bounced card should cost (2) more, got %d", c.Cost)
		}
	}
}

// TestMarksmanTrapDamagesPlayedMinion: Marksman's Trap deals 4 to a minion the
// opponent plays.
func TestMarksmanTrapDamagesPlayedMinion(t *testing.T) {
	m, a, _ := newMatch()
	placeSecret(m, 1, "marksman_trap")
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].hand = []cards.Card{getCard("marsh_snapjaw")} // 2/7
	m.sendStateAll()
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("playing a minion should resolve: %s", msg)
	}
	played := m.state[0].board[0]
	if played.health != 3 { // 7 - 4
		t.Fatalf("Marksman's Trap should deal 4 to the played minion (7->3), got %d", played.health)
	}
}

// TestFeintTrapRedirectsAttack: Feint Trap redirects an attack on the owner's
// hero to another random character. With no other minions, the only other target
// is the attacker's own hero.
func TestFeintTrapRedirectsAttack(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "war_colossus", 7, 7, true)
	placeSecret(m, 1, "feint_trap")
	if ok, msg := m.Attack(a, "atk", oppHeroTarget); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	if m.state[1].heroHP != 30 {
		t.Fatalf("Feint Trap should redirect away from the defender hero, got %d", m.state[1].heroHP)
	}
	if m.state[0].heroHP != 23 { // redirected onto the attacker's own hero (30 - 7)
		t.Fatalf("the attack should hit the only other character (attacker hero), got %d", m.state[0].heroHP)
	}
}

// TestSerpentTrapSummonsOnMinionAttacked: Serpent Trap summons three 1/1 Serpents
// when one of the owner's minions is attacked.
func TestSerpentTrapSummonsOnMinionAttacked(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "war_colossus", 7, 7, true)
	place(m, 1, "wall", "war_colossus", 1, 9, true) // the defender's attacked minion
	placeSecret(m, 1, "serpent_trap")
	if ok, msg := m.Attack(a, "atk", "wall"); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	if n := countCard(m.state[1].board, "coil_serpent"); n != 3 {
		t.Fatalf("Serpent Trap should summon three Serpents, got %d", n)
	}
}

// --- Wave E: weapons ---

// TestHawkeyeBowGainsDurabilityOnSecretReveal: Hawkeye Bow gains +1 Durability
// whenever one of the wielder's Secrets is revealed.
func TestHawkeyeBowGainsDurabilityOnSecretReveal(t *testing.T) {
	m, _, _ := newMatch()
	m.state[0].weapon = &weaponInst{card: getCard("hawkeye_bow"), attack: 3, durability: 2}
	placeSecret(m, 0, "blasting_snare") // a non-cancelling secret owned by player 0
	m.triggerSecrets(0, cards.OnHeroAttacked, secretCtx{})
	if m.state[0].weapon == nil || m.state[0].weapon.durability != 3 {
		t.Fatalf("Hawkeye Bow should gain +1 Durability on a Secret reveal (2->3), got %v", m.state[0].weapon)
	}
}

// TestDuelistsLongbowImmuneWhileAttacking: the hero takes no retaliation when
// attacking with Duelist's Longbow.
func TestDuelistsLongbowImmuneWhileAttacking(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].weapon = &weaponInst{card: getCard("duelists_longbow"), attack: 5, durability: 2}
	place(m, 1, "wall", "war_colossus", 4, 9, true)
	if ok, msg := m.Attack(a, selfHeroTarget, "wall"); !ok {
		t.Fatalf("hero attack should resolve: %s", msg)
	}
	if m.state[0].heroHP != 30 {
		t.Fatalf("Duelist's Longbow should make the hero Immune while attacking (no retaliation), got %d", m.state[0].heroHP)
	}
	if wall := findMinion(m.state[1].board, "wall"); wall == nil || wall.health != 4 { // 9 - 5
		t.Fatalf("the weapon should still deal its damage (9->4), got %v", wall)
	}
}
