package match

import (
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
)

// TestMeeknessSetsAttackToOne: Meekness sets a minion's Attack to 1 via an
// enchantment (Silence would restore it — that path is shared with EffectSetAtkToHealth).
func TestMeeknessSetsAttackToOne(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "ogre", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "meekness", "ogre")
	if o := findMinion(m.state[1].board, "ogre"); o == nil || o.atk() != 1 {
		t.Fatalf("Meekness should set Attack to 1, got %v", o)
	}
}

// TestRiftwardenPacifierOnset: the onset changes an enemy minion's Attack to 1.
func TestRiftwardenPacifierOnset(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "ogre", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "riftwarden_pacifier", "ogre")
	if o := findMinion(m.state[1].board, "ogre"); o == nil || o.atk() != 1 {
		t.Fatalf("Riftwarden Pacifier should set the enemy's Attack to 1, got %v", o)
	}
}

// TestGreatLevelingSetsAllHealthToOne: Equality-style — every minion on both
// boards ends at 1 max/current Health.
func TestGreatLevelingSetsAllHealthToOne(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "mine", "war_colossus", 7, 7, true)
	place(m, 1, "ogre", "crag_ogre", 6, 7, true)
	castFrom(t, m, a, 0, "great_leveling", "")
	for _, id := range []struct {
		side int
		uid  string
	}{{0, "mine"}, {1, "ogre"}} {
		mn := findMinion(m.state[id.side].board, id.uid)
		if mn == nil || mn.health != 1 || mn.maxHP() != 1 {
			t.Fatalf("Great Leveling should set %s to 1/1 Health, got %v", id.uid, mn)
		}
	}
}

// TestExaltedMightDoublesAttack: Blessed-Champion-style — Attack doubles.
func TestExaltedMightDoublesAttack(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "f", "granite_watcher", 3, 3, true)
	castFrom(t, m, a, 0, "exalted_might", "f")
	if f := findMinion(m.state[0].board, "f"); f == nil || f.atk() != 6 {
		t.Fatalf("Exalted Might should double Attack (3->6), got %v", f)
	}
}

// TestWardingHandGrantsAegisShield: granting Aegis must raise the LIVE pop-shield
// (not just the keyword) so the next hit is negated — the whole point of the card.
func TestWardingHandGrantsAegisShield(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "f", "granite_watcher", 2, 3, true)
	castFrom(t, m, a, 0, "warding_hand", "f")
	f := findMinion(m.state[0].board, "f")
	if f == nil || !f.aegis {
		t.Fatalf("Warding Hand should give the live Aegis shield, got %v", f)
	}
	m.damageMinion(f, 5, "x")
	if f.aegis || f.health != 3 {
		t.Fatalf("Aegis should absorb the hit (shield pops, no health lost), got %v", f)
	}
}

// TestAegisHymnGrantsAllFriendlyMinions: gives every friendly minion the Aegis
// shield; enemy minions are untouched.
func TestAegisHymnGrantsAllFriendlyMinions(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "f1", "granite_watcher", 2, 3, true)
	place(m, 0, "f2", "clay_acolyte", 3, 2, true)
	place(m, 1, "e", "granite_watcher", 2, 3, true)
	castFrom(t, m, a, 0, "aegis_hymn", "")
	if f := findMinion(m.state[0].board, "f1"); f == nil || !f.aegis {
		t.Fatalf("Hymn of Aegis should shield f1, got %v", f)
	}
	if f := findMinion(m.state[0].board, "f2"); f == nil || !f.aegis {
		t.Fatalf("Hymn of Aegis should shield f2, got %v", f)
	}
	if e := findMinion(m.state[1].board, "e"); e == nil || e.aegis {
		t.Fatal("Hymn of Aegis must not shield enemy minions")
	}
}

// TestZealotsVerdictDrawsAndBolts: draws the top card and deals its Cost to the
// target minion (Holy-Wrath-style).
func TestZealotsVerdictDrawsAndBolts(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "ogre", "crag_ogre", 6, 7, true)
	m.state[0].deck = []cards.Card{getCard("crag_ogre")} // top card cost 6
	castFrom(t, m, a, 0, "zealots_verdict", "ogre")
	if o := findMinion(m.state[1].board, "ogre"); o == nil || o.health != 1 {
		t.Fatalf("Zealot's Verdict should deal 6 (the drawn card's Cost) to the ogre (7->1), got %v", o)
	}
	if len(m.state[0].hand) != 1 {
		t.Fatalf("Zealot's Verdict should also draw the card, hand=%d", len(m.state[0].hand))
	}
}

// TestProvidenceDrawsToOpponentHand: Divine-Favor-style — draw until the caster's
// hand matches the opponent's size.
func TestProvidenceDrawsToOpponentHand(t *testing.T) {
	m, a, _ := newMatch()
	m.state[1].hand = testDeck([]string{"pebble_imp", "pebble_imp", "pebble_imp", "pebble_imp", "pebble_imp"})
	m.state[0].deck = testDeck([]string{"pebble_imp", "pebble_imp", "pebble_imp", "pebble_imp", "pebble_imp", "pebble_imp"})
	castFrom(t, m, a, 0, "providence", "") // sets hand to [providence] then plays it (hand -> 0)
	if len(m.state[0].hand) != 5 {
		t.Fatalf("Providence should draw up to the opponent's 5 cards, hand=%d", len(m.state[0].hand))
	}
}

// TestInsightBlessingDrawsOnAttack: a minion given Blessing of Insight draws its
// controller a card each time it attacks.
func TestInsightBlessingDrawsOnAttack(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "clay_acolyte", 3, 2, true)
	m.state[0].deck = testDeck([]string{"pebble_imp"})
	castFrom(t, m, a, 0, "insight_blessing", "atk") // hand -> 0 after the cast
	if ok, msg := m.Attack(a, "atk", oppHeroTarget); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	if len(m.state[0].hand) != 1 {
		t.Fatalf("Blessing of Insight should draw a card on attack, hand=%d", len(m.state[0].hand))
	}
}

// TestPureheartBladeHealsOnHeroAttack: Truesilver-style — the hero heals 3 whenever
// it attacks with the weapon.
func TestPureheartBladeHealsOnHeroAttack(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 20
	m.state[0].weapon = &weaponInst{card: prodCard("pureheart_blade"), attack: 4, durability: 2}
	if ok, msg := m.Attack(a, selfHeroTarget, oppHeroTarget); !ok {
		t.Fatalf("hero attack should resolve: %s", msg)
	}
	if m.state[0].heroHP != 23 {
		t.Fatalf("Pureheart Blade should heal the hero 3 (20->23), got %d", m.state[0].heroHP)
	}
	if m.state[1].heroHP != heroMaxHP-4 {
		t.Fatalf("the weapon should still deal 4 to the enemy hero, got %d", m.state[1].heroHP)
	}
}

// TestVerdictEdgeBuffsSummonedMinion: Sword-of-Justice-style — each summoned minion
// gets +1/+1 and the weapon loses 1 Durability.
func TestVerdictEdgeBuffsSummonedMinion(t *testing.T) {
	m, _, _ := newMatch()
	m.state[0].weapon = &weaponInst{card: prodCard("verdict_edge"), attack: 1, durability: 5}
	mn := m.summonMinion(0, getCard("pebble_imp")) // base 1/1
	if mn == nil || mn.atk() != 2 || mn.maxHP() != 2 {
		t.Fatalf("Edge of Verdict should buff the summon to 2/2, got %v", mn)
	}
	if m.state[0].weapon == nil || m.state[0].weapon.durability != 4 {
		t.Fatalf("Edge of Verdict should lose 1 Durability (5->4), got %v", m.state[0].weapon)
	}
}

// TestMusterSummonsRecruit: the Paladin hero power summons a 1/1 Recruit token.
func TestMusterSummonsRecruit(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroPower = prodCard("muster")
	m.state[0].mana, m.state[0].maxMana = 10, 10
	if ok, msg := m.HeroPower(a, ""); !ok {
		t.Fatalf("Muster should resolve: %s", msg)
	}
	if n := countCard(m.state[0].board, "lightsworn_recruit"); n != 1 {
		t.Fatalf("Muster should summon one Recruit, got %d", n)
	}
}

// TestHighlordValdricEquipsWeaponOnDeath: the Final Gasp equips a 5/3 weapon.
func TestHighlordValdricEquipsWeaponOnDeath(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "val", "highlord_valdric", 6, 6, true)
	v := findMinion(m.state[0].board, "val")
	v.aegis = false
	v.health = 0
	m.finish()
	w := m.state[0].weapon
	if w == nil || w.attack != 5 || w.durability != 3 {
		t.Fatalf("Highlord Valdric's Final Gasp should equip a 5/3 weapon, got %v", w)
	}
}

// TestRetributionVowReflectsHeroDamage: Eye-for-an-Eye — damage to the owner's hero
// is reflected to the enemy hero.
func TestRetributionVowReflectsHeroDamage(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "clay_acolyte", 3, 2, true) // 3 Attack
	placeSecret(m, 1, "retribution_vow")
	beforeMine := m.state[0].heroHP
	if ok, msg := m.Attack(a, "atk", oppHeroTarget); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	if m.state[1].heroHP != heroMaxHP-3 {
		t.Fatalf("enemy hero should take the 3 attack, got %d", m.state[1].heroHP)
	}
	if m.state[0].heroHP != beforeMine-3 {
		t.Fatalf("Vow of Retribution should reflect 3 to my hero, got %d (was %d)", m.state[0].heroHP, beforeMine)
	}
	if len(m.state[1].secrets) != 0 {
		t.Fatal("the secret should be consumed")
	}
}

// TestValiantWardSummonsDefenderAndRedirects: Noble-Sacrifice — an enemy attack is
// redirected onto a freshly-summoned 2/1 Defender, sparing the hero.
func TestValiantWardSummonsDefenderAndRedirects(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "thicket_stalker", 3, 3, true) // 3/3 attacker
	placeSecret(m, 1, "valiant_ward")
	if ok, msg := m.Attack(a, "atk", oppHeroTarget); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	if m.state[1].heroHP != heroMaxHP {
		t.Fatalf("Valiant Ward should spare the hero, got %d", m.state[1].heroHP)
	}
	if atk := findMinion(m.state[0].board, "atk"); atk == nil || atk.health != 1 {
		t.Fatalf("the attacker should take the 2/1 Defender's retaliation (3->1), got %v", atk)
	}
}

// TestSecondDawnResummonsAtOneHealth: Redemption — a dead friendly minion returns
// to life at 1 Health (as its base card, no buffs).
func TestSecondDawnResummonsAtOneHealth(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "victim", "granite_watcher", 5, 3, true) // buffed on the board
	placeSecret(m, 0, "second_dawn")
	v := findMinion(m.state[0].board, "victim")
	v.health = 0
	m.finish()
	if len(m.state[0].board) != 1 {
		t.Fatalf("Second Dawn should resummon the minion, board=%d", len(m.state[0].board))
	}
	r := m.state[0].board[0]
	if r.health != 1 || r.maxHP() != 3 || r.atk() != 2 {
		t.Fatalf("Second Dawn should return a base 2/3 at 1 Health, got %v", r)
	}
	if len(m.state[0].secrets) != 0 {
		t.Fatal("the secret should be consumed")
	}
}

// TestPenanceSealReducesPlayedMinionHealth: Repentance — after the opponent plays a
// minion, its Health is set to 1.
func TestPenanceSealReducesPlayedMinionHealth(t *testing.T) {
	m, a, _ := newMatch()
	placeSecret(m, 1, "penance_seal")
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].hand = []cards.Card{getCard("crag_ogre")} // 6/7
	m.sendStateAll()
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("playing the minion should resolve: %s", msg)
	}
	ogre := m.state[0].board[len(m.state[0].board)-1]
	if ogre.health != 1 || ogre.maxHP() != 1 {
		t.Fatalf("Seal of Penance should reduce the played minion's Health to 1, got %v", ogre)
	}
	if len(m.state[1].secrets) != 0 {
		t.Fatal("the secret should be consumed")
	}
}
