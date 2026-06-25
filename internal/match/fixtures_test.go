package match

import "github.com/amvid/vanillastone/internal/cards"

// testCards are white-box fixtures for the match engine tests. They were the
// original demo cards before the card pool was rebuilt; the tests exercise engine
// mechanics (onset, finalGasp, secrets, the keyword set, weapons, …) against
// these stable fixtures rather than the shipping pool, so the pool can change
// freely without rewriting engine coverage.
//
// Two fixtures reference REAL production tokens (broken_golem) because the summon
// path resolves tokens through the production registry, and the Seek fixture
// uses the minion pool (the shipping spell pool is too small to offer three).
var testCards = map[string]cards.Card{
	"pebble_imp":      {ID: "pebble_imp", Name: "Pebble Imp", Type: cards.TypeMinion, Cost: 1, Attack: 1, Health: 1},
	"clay_acolyte":    {ID: "clay_acolyte", Name: "Clay Acolyte", Type: cards.TypeMinion, Cost: 2, Attack: 3, Health: 2},
	"granite_watcher": {ID: "granite_watcher", Name: "Granite Watcher", Type: cards.TypeMinion, Cost: 2, Attack: 2, Health: 3},
	"thicket_stalker": {ID: "thicket_stalker", Name: "Thicket Stalker", Type: cards.TypeMinion, Cost: 3, Attack: 3, Health: 3},

	"spark_adept": {ID: "spark_adept", Name: "Spark Adept", Type: cards.TypeMinion, Cost: 2, Attack: 2, Health: 2,
		Text:     "Onset: Deal 2 damage to any character.",
		Triggers: []cards.Trigger{{When: cards.OnPlay, Effect: cards.Effect{Kind: cards.EffectDamage, Amount: 2, Target: cards.TargetAny}}}},
	"ember_striker": {ID: "ember_striker", Name: "Ember Striker", Type: cards.TypeMinion, Cost: 2, Attack: 2, Health: 1,
		Text:     "Onset: Deal 1 damage to an enemy minion.",
		Triggers: []cards.Trigger{{When: cards.OnPlay, Effect: cards.Effect{Kind: cards.EffectDamage, Amount: 1, Target: cards.TargetEnemyMinion}}}},
	"bog_warden": {ID: "bog_warden", Name: "Bog Warden", Type: cards.TypeMinion, Cost: 3, Attack: 2, Health: 2,
		Text:     "Onset: Summon a 2/1 Broken Golem.",
		Triggers: []cards.Trigger{{When: cards.OnPlay, Effect: cards.Effect{Kind: cards.EffectSummon, Summon: "broken_golem", Count: 1, Target: cards.TargetNone}}}},
	"brood_mother": {ID: "brood_mother", Name: "Brood Mother", Type: cards.TypeMinion, Cost: 2, Attack: 2, Health: 2,
		Text:     "Final Gasp: Summon a 2/1 Broken Golem.",
		Triggers: []cards.Trigger{{When: cards.OnDeath, Effect: cards.Effect{Kind: cards.EffectSummon, Summon: "broken_golem", Count: 1, Target: cards.TargetNone}}}},
	"cinder_husk": {ID: "cinder_husk", Name: "Cinder Husk", Type: cards.TypeMinion, Cost: 3, Attack: 3, Health: 2,
		Text:     "Final Gasp: Deal 2 damage to the enemy hero.",
		Triggers: []cards.Trigger{{When: cards.OnDeath, Effect: cards.Effect{Kind: cards.EffectDamage, Amount: 2, Target: cards.TargetNone, Area: cards.AreaEnemyHero}}}},
	"volatile_wisp": {ID: "volatile_wisp", Name: "Volatile Wisp", Type: cards.TypeMinion, Cost: 1, Attack: 1, Health: 1,
		Text:     "Final Gasp: Deal 1 damage to a random enemy.",
		Triggers: []cards.Trigger{{When: cards.OnDeath, Effect: cards.Effect{Kind: cards.EffectDamage, Amount: 1, Target: cards.TargetRandomEnemy}}}},

	"bastion_golem": {ID: "bastion_golem", Name: "Bastion Golem", Type: cards.TypeMinion, Cost: 4, Attack: 3, Health: 5,
		Text: "Taunt.", Keywords: []cards.Keyword{cards.KeywordTaunt}},
	"swift_raptor": {ID: "swift_raptor", Name: "Swift Raptor", Type: cards.TypeMinion, Cost: 3, Attack: 3, Health: 2,
		Text: "Charge.", Keywords: []cards.Keyword{cards.KeywordCharge}},
	"lurking_stalker": {ID: "lurking_stalker", Name: "Lurking Stalker", Type: cards.TypeMinion, Cost: 3, Attack: 3, Health: 3,
		Text: "Rush.", Keywords: []cards.Keyword{cards.KeywordRush}},
	"gilded_sentry": {ID: "gilded_sentry", Name: "Gilded Sentry", Type: cards.TypeMinion, Cost: 3, Attack: 2, Health: 3,
		Text: "Aegis.", Keywords: []cards.Keyword{cards.KeywordAegis}},

	"gale_harrier": {ID: "gale_harrier", Name: "Gale Harrier", Type: cards.TypeMinion, Cost: 3, Attack: 2, Health: 3,
		Text: "Twinstrike.", Keywords: []cards.Keyword{cards.KeywordTwinstrike}},
	"veil_stalker": {ID: "veil_stalker", Name: "Veil Stalker", Type: cards.TypeMinion, Cost: 1, Attack: 2, Health: 1,
		Text: "Stealth.", Keywords: []cards.Keyword{cards.KeywordStealth}},
	"toxic_fang": {ID: "toxic_fang", Name: "Toxic Fang", Type: cards.TypeMinion, Cost: 2, Attack: 1, Health: 3,
		Text: "Poisonous.", Keywords: []cards.Keyword{cards.KeywordPoisonous}},
	"bloodthorn_knight": {ID: "bloodthorn_knight", Name: "Bloodthorn Knight", Type: cards.TypeMinion, Cost: 4, Attack: 3, Health: 4,
		Text: "Lifesteal.", Keywords: []cards.Keyword{cards.KeywordLifesteal}},
	"ember_scribe": {ID: "ember_scribe", Name: "Ember Scribe", Type: cards.TypeMinion, Cost: 2, Attack: 1, Health: 3,
		Text: "Spell Damage +1.", SpellDamage: 1},
	"pack_leader": {ID: "pack_leader", Name: "Pack Leader", Type: cards.TypeMinion, Cost: 3, Attack: 2, Health: 3,
		Text: "Your other minions have +1 Attack.", Aura: &cards.Aura{Atk: 1}},

	"cinder_bolt": {ID: "cinder_bolt", Name: "Cinder Bolt", Type: cards.TypeSpell, Cost: 2,
		Text:   "Deal 3 damage to any character.",
		Effect: &cards.Effect{Kind: cards.EffectDamage, Amount: 3, Target: cards.TargetAny}},
	"mend": {ID: "mend", Name: "Mend", Type: cards.TypeSpell, Cost: 1,
		Text:   "Restore 4 Health to any character.",
		Effect: &cards.Effect{Kind: cards.EffectHeal, Amount: 4, Target: cards.TargetAny}},
	"whetstone": {ID: "whetstone", Name: "Whetstone", Type: cards.TypeSpell, Cost: 1,
		Text:   "Give a friendly minion +2/+1.",
		Effect: &cards.Effect{Kind: cards.EffectBuff, BuffAtk: 2, BuffHP: 1, Target: cards.TargetFriendlyMinion}},
	"quake": {ID: "quake", Name: "Quake", Type: cards.TypeSpell, Cost: 3,
		Text:   "Deal 1 damage to all enemy minions.",
		Effect: &cards.Effect{Kind: cards.EffectDamage, Amount: 1, Target: cards.TargetNone, Area: cards.AreaEnemyMinions}},
	"frost_snap": {ID: "frost_snap", Name: "Frost Snap", Type: cards.TypeSpell, Cost: 2,
		Text:   "Deal 1 damage to a character and Freeze it.",
		Effect: &cards.Effect{Kind: cards.EffectDamage, Amount: 1, Target: cards.TargetAny, Freeze: true}},
	"permafrost": {ID: "permafrost", Name: "Permafrost", Type: cards.TypeSpell, Cost: 3,
		Text:   "Freeze all enemy minions.",
		Effect: &cards.Effect{Kind: cards.EffectDamage, Amount: 0, Target: cards.TargetNone, Area: cards.AreaEnemyMinions, Freeze: true}},
	"hush": {ID: "hush", Name: "Hush", Type: cards.TypeSpell, Cost: 1,
		Text:   "Silence a minion.",
		Effect: &cards.Effect{Kind: cards.EffectSilence, Target: cards.TargetMinion}},
	"drain_touch": {ID: "drain_touch", Name: "Drain Touch", Type: cards.TypeSpell, Cost: 2,
		Text:   "Deal 2 damage to a character. Lifesteal.",
		Effect: &cards.Effect{Kind: cards.EffectDamage, Amount: 2, Target: cards.TargetAny, Lifesteal: true}},

	"snare": {ID: "snare", Name: "Snare", Type: cards.TypeSecret, Cost: 2,
		Text:   "Secret: When an enemy minion attacks your hero, destroy it.",
		Secret: &cards.SecretDef{Trigger: cards.OnEnemyAttackHero, Kind: cards.SecretDestroyAttacker}},
	"mimic": {ID: "mimic", Name: "Mimic", Type: cards.TypeSecret, Cost: 3,
		Text:   "Secret: When an enemy plays a minion, summon a copy.",
		Secret: &cards.SecretDef{Trigger: cards.OnEnemyPlayMinion, Kind: cards.SecretCopyMinion}},
	"nullify": {ID: "nullify", Name: "Nullify", Type: cards.TypeSecret, Cost: 3,
		Text:   "Secret: When an enemy casts a spell, counter it.",
		Secret: &cards.SecretDef{Trigger: cards.OnEnemyCastSpell, Kind: cards.SecretCounterSpell}},
	"frost_ward": {ID: "frost_ward", Name: "Frost Ward", Type: cards.TypeSecret, Cost: 3,
		Text:   "Secret: When your hero is attacked, gain 8 Armor.",
		Secret: &cards.SecretDef{Trigger: cards.OnHeroAttacked, Kind: cards.SecretGainArmor, Amount: 8}},

	// Seek fixtures use the MINION pool (the production spell pool is too small
	// to offer three distinct options).
	"arcane_insight": {ID: "arcane_insight", Name: "Arcane Insight", Type: cards.TypeMinion, Cost: 2, Attack: 2, Health: 2,
		Text:     "Onset: Seek a minion.",
		Triggers: []cards.Trigger{{When: cards.OnPlay, Effect: cards.Effect{Kind: cards.EffectSeek, Pool: cards.SeekMinion, Target: cards.TargetNone}}}},
	"wild_summons": {ID: "wild_summons", Name: "Wild Summons", Type: cards.TypeMinion, Cost: 3, Attack: 3, Health: 2,
		Text:     "Onset: Seek a minion.",
		Triggers: []cards.Trigger{{When: cards.OnPlay, Effect: cards.Effect{Kind: cards.EffectSeek, Pool: cards.SeekMinion, Target: cards.TargetNone}}}},

	"ember_cleaver": {ID: "ember_cleaver", Name: "Ember Cleaver", Type: cards.TypeWeapon, Cost: 2, Attack: 3, Durability: 2,
		Text: "A 3/2 weapon."},
	"quartz_spike": {ID: "quartz_spike", Name: "Quartz Spike", Type: cards.TypeWeapon, Cost: 3, Attack: 2, Durability: 3,
		Text: "A 2/3 weapon."},

	// Transform / summon-spell fixtures: the engine supports these effects, but no
	// Classic shipping card uses them, so the tests exercise fixtures.
	"hex_bolt": {ID: "hex_bolt", Name: "Hex Bolt", Type: cards.TypeSpell, Cost: 4,
		Text:   "Transform a minion into a 2/1 Broken Golem.",
		Effect: &cards.Effect{Kind: cards.EffectTransform, Transform: "broken_golem", Target: cards.TargetMinion}},
	"twin_summons": {ID: "twin_summons", Name: "Twin Summons", Type: cards.TypeSpell, Cost: 3,
		Text:   "Summon two 2/1 Broken Golems.",
		Effect: &cards.Effect{Kind: cards.EffectSummon, Summon: "broken_golem", Count: 2, Target: cards.TargetNone}},

	// Destroy / enemy-target / friendly-hero-heal fixtures: the engine supports
	// these effects/rules, but no Classic shipping card uses them yet, so the tests
	// exercise fixtures.
	"banish_rite": {ID: "banish_rite", Name: "Banish Rite", Type: cards.TypeSpell, Cost: 4,
		Text:   "Destroy a minion.",
		Effect: &cards.Effect{Kind: cards.EffectDestroy, Target: cards.TargetMinion}},
	"headsman": {ID: "headsman", Name: "Headsman", Type: cards.TypeMinion, Cost: 5, Attack: 4, Health: 3,
		Text:     "Onset: Destroy an enemy minion.",
		Triggers: []cards.Trigger{{When: cards.OnPlay, Effect: cards.Effect{Kind: cards.EffectDestroy, Target: cards.TargetEnemyMinion}}}},
	"bombard_captain": {ID: "bombard_captain", Name: "Bombard Captain", Type: cards.TypeMinion, Cost: 5, Attack: 4, Health: 2,
		Text:     "Onset: Deal 2 damage to an enemy character.",
		Triggers: []cards.Trigger{{When: cards.OnPlay, Effect: cards.Effect{Kind: cards.EffectDamage, Amount: 2, Target: cards.TargetEnemy}}}},
	"tavern_medic_fx": {ID: "tavern_medic_fx", Name: "Tavern Medic", Type: cards.TypeMinion, Cost: 2, Attack: 2, Health: 2,
		Text:     "Onset: Restore 4 Health to your hero.",
		Triggers: []cards.Trigger{{When: cards.OnPlay, Effect: cards.Effect{Kind: cards.EffectHeal, Amount: 4, Target: cards.TargetFriendlyHero}}}},
}

// getCard resolves a card id for the white-box tests: a test fixture if one
// exists, otherwise the production registry (tokens, Mana Surge, the hero power).
func getCard(id string) cards.Card {
	if c, ok := testCards[id]; ok {
		return c
	}
	c, _ := cards.Get(id)
	return c
}

// testDeck materializes a list of ids into Card values via getCard.
func testDeck(ids []string) []cards.Card {
	out := make([]cards.Card, 0, len(ids))
	for _, id := range ids {
		out = append(out, getCard(id))
	}
	return out
}
