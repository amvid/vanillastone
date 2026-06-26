package match

import (
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
)

// hurt lowers a placed minion's current Health below its max so it reads as
// "damaged" (Enrage / ReqDamaged / on-damage triggers don't fire on the set-up
// itself — only a real damage instance does).
func hurt(m *Match, owner int, uid string, to int) {
	if mn := findMinion(m.state[owner].board, uid); mn != nil {
		mn.health = to
	}
}

// TestShoreUpGainsArmor: the Warrior hero power gives the hero 2 Armor.
func TestShoreUpGainsArmor(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroPower = getCard("shore_up")
	m.state[0].mana, m.state[0].maxMana = 10, 10
	if ok, msg := m.HeroPower(a, ""); !ok {
		t.Fatalf("Shore Up should resolve: %s", msg)
	}
	if m.state[0].armor != 2 {
		t.Fatalf("Shore Up should give 2 Armor, got %d", m.state[0].armor)
	}
}

// TestGoadingStrikeDamageAndBuff: Goading Strike deals 1 damage to a minion AND
// gives it +2 Attack.
func TestGoadingStrikeDamageAndBuff(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "x", "granite_watcher", 2, 3, true)
	castFrom(t, m, a, 0, "goading_strike", "x")
	x := findMinion(m.state[1].board, "x")
	if x == nil || x.health != 2 || x.atk() != 4 {
		t.Fatalf("Goading Strike should leave a 4/2 (was 2/3, -1 hp +2 atk), got %d/%d", x.atk(), x.health)
	}
}

// TestHammerBlowDrawsIfSurvives: Hammer Blow deals 2 and draws a card only when
// the target survives.
func TestHammerBlowDrawsIfSurvives(t *testing.T) {
	// Survives -> draw.
	m, a, _ := newMatch()
	place(m, 1, "big", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "hammer_blow", "big")
	if h := len(m.state[0].hand); h != 1 {
		t.Fatalf("Hammer Blow on a survivor should draw 1 (hand=1), got %d", h)
	}
	// Dies -> no draw.
	m2, a2, _ := newMatch()
	place(m2, 1, "small", "granite_watcher", 2, 2, true)
	castFrom(t, m2, a2, 0, "hammer_blow", "small")
	if findMinion(m2.state[1].board, "small") != nil {
		t.Fatal("Hammer Blow should kill a 2-Health minion")
	}
	if h := len(m2.state[0].hand); h != 0 {
		t.Fatalf("Hammer Blow on a kill should not draw (hand=0), got %d", h)
	}
}

// TestBerserkSurgeRequiresDamaged: Berserk Surge can target only a damaged
// minion, and buffs it +3/+3.
func TestBerserkSurgeRequiresDamaged(t *testing.T) {
	// Undamaged -> rejected.
	m, a, _ := newMatch()
	place(m, 0, "fresh", "granite_watcher", 2, 3, true)
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].hand = []cards.Card{getCard("berserk_surge")}
	if ok, _ := m.PlayCard(a, 0, "fresh"); ok {
		t.Fatal("Berserk Surge should reject an undamaged target")
	}
	// Damaged -> +3/+3.
	m2, a2, _ := newMatch()
	place(m2, 0, "hurt", "granite_watcher", 2, 3, true)
	hurt(m2, 0, "hurt", 1)
	castFrom(t, m2, a2, 0, "berserk_surge", "hurt")
	x := findMinion(m2.state[0].board, "hurt")
	if x == nil || x.atk() != 5 || x.maxHP() != 6 {
		t.Fatalf("Berserk Surge should make the damaged 2/3 a 5/6, got %d/%d", x.atk(), x.maxHP())
	}
}

// TestFinishingCutRequiresDamaged: Finishing Cut destroys a damaged enemy minion,
// but cannot target a healthy one.
func TestFinishingCutRequiresDamaged(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "fresh", "marsh_snapjaw", 2, 7, true)
	m.state[0].mana, m.state[0].maxMana = 10, 10
	m.state[0].hand = []cards.Card{getCard("finishing_cut")}
	if ok, _ := m.PlayCard(a, 0, "fresh"); ok {
		t.Fatal("Finishing Cut should reject an undamaged target")
	}
	m2, a2, _ := newMatch()
	place(m2, 1, "hurt", "marsh_snapjaw", 2, 7, true)
	hurt(m2, 1, "hurt", 3)
	castFrom(t, m2, a2, 0, "finishing_cut", "hurt")
	if findMinion(m2.state[1].board, "hurt") != nil {
		t.Fatal("Finishing Cut should destroy a damaged enemy minion")
	}
}

// TestBulwarkBashScalesByArmor: Bulwark Bash deals damage equal to the caster's
// Armor.
func TestBulwarkBashScalesByArmor(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].armor = 3
	place(m, 1, "x", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "bulwark_bash", "x")
	if x := findMinion(m.state[1].board, "x"); x == nil || x.health != 4 {
		t.Fatalf("Bulwark Bash with 3 Armor should deal 3 (7->4), got %v", x)
	}
}

// TestDeathblowSwingScalesWithOwnHealth: Deathblow Swing deals 4 normally, 6 when
// the caster's hero is at or below 12 Health.
func TestDeathblowSwingScalesWithOwnHealth(t *testing.T) {
	// Above the threshold -> 4.
	m, a, _ := newMatch()
	m.state[0].heroHP = 13
	castFrom(t, m, a, 0, "deathblow_swing", oppHeroTarget)
	if hp := m.state[1].heroHP; hp != 26 {
		t.Fatalf("Deathblow Swing at 13 Health should deal 4 (30->26), got %d", hp)
	}
	// At the threshold -> 6.
	m2, a2, _ := newMatch()
	m2.state[0].heroHP = 12
	castFrom(t, m2, a2, 0, "deathblow_swing", oppHeroTarget)
	if hp := m2.state[1].heroHP; hp != 24 {
		t.Fatalf("Deathblow Swing at 12 Health should deal 6 (30->24), got %d", hp)
	}
}

// TestValiantStrikeGivesHeroAttack: Valiant Strike gives the hero +4 Attack this
// turn with no weapon, letting it attack a minion.
func TestValiantStrikeGivesHeroAttack(t *testing.T) {
	m, a, _ := newMatch()
	castFrom(t, m, a, 0, "valiant_strike", "")
	if v := heroAttackValue(m.state[0]); v != 4 {
		t.Fatalf("Valiant Strike should give the hero 4 Attack, got %d", v)
	}
	place(m, 1, "x", "marsh_snapjaw", 0, 7, true) // 0-attack so no retaliation noise
	if ok, msg := m.Attack(a, selfHeroTarget, "x"); !ok {
		t.Fatalf("hero should be able to attack with Valiant Strike: %s", msg)
	}
	if x := findMinion(m.state[1].board, "x"); x == nil || x.health != 3 {
		t.Fatalf("hero attack should deal 4 (7->3), got %v", x)
	}
}

// TestBracingGuardArmorAndDraw: Bracing Guard gives 5 Armor and draws a card.
func TestBracingGuardArmorAndDraw(t *testing.T) {
	m, a, _ := newMatch()
	castFrom(t, m, a, 0, "bracing_guard", "")
	if m.state[0].armor != 5 {
		t.Fatalf("Bracing Guard should give 5 Armor, got %d", m.state[0].armor)
	}
	if h := len(m.state[0].hand); h != 1 {
		t.Fatalf("Bracing Guard should draw 1 (hand=1), got %d", h)
	}
}

// TestWarFrenzyDrawsPerDamaged: War Frenzy draws one card per damaged friendly
// character (hero + minions).
func TestWarFrenzyDrawsPerDamaged(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 20 // damaged hero
	place(m, 0, "m1", "marsh_snapjaw", 2, 7, true)
	place(m, 0, "m2", "marsh_snapjaw", 2, 7, true)
	place(m, 0, "m3", "marsh_snapjaw", 2, 7, true)
	hurt(m, 0, "m1", 3)
	hurt(m, 0, "m2", 3) // m3 stays full
	castFrom(t, m, a, 0, "war_frenzy", "")
	if h := len(m.state[0].hand); h != 3 {
		t.Fatalf("War Frenzy with hero + 2 damaged minions should draw 3, got %d", h)
	}
}

// TestPitBrawlLeavesOne: Pit Brawl destroys every minion but one.
func TestPitBrawlLeavesOne(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "a1", "marsh_snapjaw", 2, 7, true)
	place(m, 0, "a2", "marsh_snapjaw", 2, 7, true)
	place(m, 1, "b1", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "pit_brawl", "")
	if n := len(m.state[0].board) + len(m.state[1].board); n != 1 {
		t.Fatalf("Pit Brawl should leave exactly one minion, got %d", n)
	}
}

// TestSteelCycloneHitsAllMinions: Steel Cyclone deals 1 to every minion on both
// boards.
func TestSteelCycloneHitsAllMinions(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "f", "marsh_snapjaw", 2, 7, true)
	place(m, 1, "e", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "steel_cyclone", "")
	f, e := findMinion(m.state[0].board, "f"), findMinion(m.state[1].board, "e")
	if f == nil || e == nil || f.health != 6 || e.health != 6 {
		t.Fatalf("Steel Cyclone should deal 1 to all minions, got %v %v", f, e)
	}
}

// TestRallyingRoarFloorsHealth: after Rallying Roar, the caster's minions can't be
// reduced below 1 Health this turn (and it draws a card).
func TestRallyingRoarFloorsHealth(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "x", "marsh_snapjaw", 2, 7, true)
	castFrom(t, m, a, 0, "rallying_roar", "")
	if h := len(m.state[0].hand); h != 1 {
		t.Fatalf("Rallying Roar should draw 1, got %d", h)
	}
	m.damageMinion(findMinion(m.state[0].board, "x"), 99, "test")
	if x := findMinion(m.state[0].board, "x"); x == nil || x.health != 1 {
		t.Fatalf("Rallying Roar should floor the minion at 1 Health, got %v", x)
	}
}

// TestWhipcrackOverseerBattlecry: Whipcrack Overseer's Onset deals 1 to a minion
// and gives it +2 Attack.
func TestWhipcrackOverseerBattlecry(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "x", "granite_watcher", 2, 3, true)
	castFrom(t, m, a, 0, "whipcrack_overseer", "x")
	x := findMinion(m.state[1].board, "x")
	if x == nil || x.health != 2 || x.atk() != 4 {
		t.Fatalf("Whipcrack Overseer should leave the target a 4/2, got %d/%d", x.atk(), x.health)
	}
}

// TestPlatewrightGainsArmorOnFriendlyDamage: Platewright gives the hero 1 Armor
// whenever a friendly minion takes damage.
func TestPlatewrightGainsArmorOnFriendlyDamage(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "smith", "platewright", 1, 4, true)
	place(m, 0, "ally", "marsh_snapjaw", 2, 7, true)
	m.damageMinion(findMinion(m.state[0].board, "ally"), 2, "test")
	if m.state[0].armor != 1 {
		t.Fatalf("Platewright should give 1 Armor on a friendly minion's damage, got %d", m.state[0].armor)
	}
}

// TestRageboundBruteGrowsOnAnyDamage: Ragebound Brute gains +1 Attack whenever any
// minion (either side) takes damage.
func TestRageboundBruteGrowsOnAnyDamage(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "rb", "ragebound_brute", 2, 4, true)
	place(m, 1, "foe", "marsh_snapjaw", 2, 7, true)
	m.damageMinion(findMinion(m.state[1].board, "foe"), 2, "test")
	if rb := findMinion(m.state[0].board, "rb"); rb == nil || rb.atk() != 3 {
		t.Fatalf("Ragebound Brute should be 3 Attack after one damage instance, got %v", rb)
	}
}

// TestForgeholdSmithEquips: Forgehold Smith's Onset equips a 2/2 weapon.
func TestForgeholdSmithEquips(t *testing.T) {
	m, a, _ := newMatch()
	castFrom(t, m, a, 0, "forgehold_smith", "")
	w := m.state[0].weapon
	if w == nil || w.attack != 2 || w.durability != 2 {
		t.Fatalf("Forgehold Smith should equip a 2/2 weapon, got %v", w)
	}
}

// TestHoneEdgeEquipsOrUpgrades: Hone Edge equips a 1/3 weapon with no weapon, but
// buffs an existing one +1/+1.
func TestHoneEdgeEquipsOrUpgrades(t *testing.T) {
	// No weapon -> equip 1/3.
	m, a, _ := newMatch()
	castFrom(t, m, a, 0, "hone_edge", "")
	if w := m.state[0].weapon; w == nil || w.attack != 1 || w.durability != 3 {
		t.Fatalf("Hone Edge with no weapon should equip a 1/3, got %v", w)
	}
	// Has a weapon -> +1/+1.
	m2, a2, _ := newMatch()
	m2.state[0].weapon = &weaponInst{card: getCard("keenedge_blade"), attack: 2, durability: 2}
	castFrom(t, m2, a2, 0, "hone_edge", "")
	if w := m2.state[0].weapon; w == nil || w.attack != 3 || w.durability != 3 {
		t.Fatalf("Hone Edge with a weapon should make it 3/3, got %v", w)
	}
}

// TestBattleMarshalGrantsCharge: Battle Marshal gives Charge to a summoned minion
// with 3 or less Attack, but not to a bigger one.
func TestBattleMarshalGrantsCharge(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "marshal", "battle_marshal", 2, 3, true)
	small := m.summonMinion(0, getCard("pebble_imp")) // 1 Attack
	if small == nil || !small.has(cards.KeywordCharge) {
		t.Fatal("Battle Marshal should give Charge to a 1-Attack summon")
	}
	big := m.summonMinion(0, getCard("bombard_captain")) // 4 Attack
	if big == nil || big.has(cards.KeywordCharge) {
		t.Fatal("Battle Marshal should NOT give Charge to a 4-Attack summon")
	}
}

// TestBloodwailWearsByAttack: attacking a minion with Bloodwail spends 1 Attack,
// not Durability; attacking the hero spends Durability as usual.
func TestBloodwailWearsByAttack(t *testing.T) {
	// Attack a minion -> Attack drops, Durability holds.
	m, a, _ := newMatch()
	m.state[0].weapon = &weaponInst{card: getCard("bloodwail"), attack: 7, durability: 1}
	place(m, 1, "x", "marsh_snapjaw", 0, 7, true) // 0-attack: no retaliation
	if ok, msg := m.Attack(a, selfHeroTarget, "x"); !ok {
		t.Fatalf("hero should attack the minion: %s", msg)
	}
	if w := m.state[0].weapon; w == nil || w.attack != 6 || w.durability != 1 {
		t.Fatalf("Bloodwail attacking a minion should be 6/1, got %v", w)
	}
	// Attack the hero -> Durability spent (weapon breaks at 0).
	m2, a2, _ := newMatch()
	m2.state[0].weapon = &weaponInst{card: getCard("bloodwail"), attack: 7, durability: 1}
	if ok, msg := m2.Attack(a2, selfHeroTarget, oppHeroTarget); !ok {
		t.Fatalf("hero should attack the enemy hero: %s", msg)
	}
	if m2.state[0].weapon != nil {
		t.Fatal("Bloodwail attacking the hero should spend its last Durability and break")
	}
}
