package cards

// mageCards are the Mage class cards. Target scope is the HS Classic set's 15
// Mage cards (mechanics 1:1; names + art wholly original — see HANDOFF "Legal
// rules"). Currently 11 of the 15 are in (e.g. `codex_of_insight`,
// `frostshear`); the remaining 4 are mechanic-gated and land as
// their mechanics are built (see TASKS.md):
//   - `arcane_adept`, `spellwarden_magus` → cost reduction
//   - `glacial_splinter`  → conditional (if the target is Frozen)
//   - `decoy_ward`        → retarget-spell secret
//
// Secrets keep Type TypeSecret (the engine drives the hidden secret zone off it);
// the Mage class color makes them blue like any other Mage spell.
var mageCards = []Card{
	// --- Secrets (Go-handled triggers; fire on the opponent's action) ---

	{ID: "glacial_ward", Name: "Glacial Ward", Type: TypeSecret, Class: ClassMage, Rarity: RarityCommon, Cost: 3,
		Text:   "Secret: When your hero is attacked, gain 8 Armor.",
		Secret: &SecretDef{Trigger: OnHeroAttacked, Kind: SecretGainArmor, Amount: 8}},

	{ID: "echo_glass", Name: "Echo Glass", Type: TypeSecret, Class: ClassMage, Rarity: RarityCommon, Cost: 3,
		Text:   "Secret: When your opponent plays a minion, summon a copy of it.",
		Secret: &SecretDef{Trigger: OnEnemyPlayMinion, Kind: SecretCopyMinion}},

	{ID: "nullrune", Name: "Nullrune", Type: TypeSecret, Class: ClassMage, Rarity: RarityRare, Cost: 3,
		Text:   "Secret: When your opponent casts a spell, counter it.",
		Secret: &SecretDef{Trigger: OnEnemyCastSpell, Kind: SecretCounterSpell}},

	{ID: "cinder_trap", Name: "Cinder Trap", Type: TypeSecret, Class: ClassMage, Rarity: RarityRare, Cost: 3,
		Text:   "Secret: When a minion attacks your hero, destroy it.",
		Secret: &SecretDef{Trigger: OnEnemyAttackHero, Kind: SecretDestroyAttacker}},

	{ID: "decoy_ward", Name: "Decoy Ward", Type: TypeSecret, Class: ClassMage, Rarity: RarityEpic, Cost: 3,
		Text:   "Secret: When an enemy casts a spell on one of your minions, summon a 1/3 and make it the new target.",
		Secret: &SecretDef{Trigger: OnEnemyCastSpell, Kind: SecretRetargetSpell, Summon: "conjured_decoy"}},

	// --- Minions ---

	{ID: "arcane_wyrmling", Name: "Arcane Wyrmling", Type: TypeMinion, Class: ClassMage, Rarity: RarityCommon, Cost: 1, Attack: 1, Health: 3,
		Text:     "Whenever you cast a spell, gain +1 Attack.",
		Triggers: []Trigger{{When: OnSpellCast, Effect: Effect{Kind: EffectBuff, BuffAtk: 1, Target: TargetSelf}}}},

	{ID: "warded_scholar", Name: "Warded Scholar", Type: TypeMinion, Class: ClassMage, Rarity: RarityRare, Cost: 4, Attack: 3, Health: 3,
		Text: "At the end of your turn, if you control a Secret, gain +2/+2.",
		Triggers: []Trigger{{When: OnTurnEnd, Condition: CondControlSecret,
			Effect: Effect{Kind: EffectBuff, BuffAtk: 2, BuffHP: 2, Target: TargetSelf}}}},

	{ID: "emberforge_magus", Name: "Emberforge Magus", Type: TypeMinion, Class: ClassMage, Rarity: RarityLegendary, Cost: 7, Attack: 5, Health: 7,
		Text:     "Whenever you cast a spell, add a Pyrebolt to your hand.",
		Triggers: []Trigger{{When: OnSpellCast, Effect: Effect{Kind: EffectGenerate, Generate: "pyrebolt", Target: TargetNone}}}},

	{ID: "arcane_adept", Name: "Arcane Adept", Type: TypeMinion, Class: ClassMage, Rarity: RarityCommon, Cost: 2, Attack: 3, Health: 2,
		Text:     "Your spells cost (1) less.",
		CostAura: &CostAura{Delta: -1, Scope: CostScopeFriendly, Type: TypeSpell}},

	{ID: "spellwarden_magus", Name: "Spellwarden Magus", Type: TypeMinion, Class: ClassMage, Rarity: RarityRare, Cost: 3, Attack: 4, Health: 3,
		Text:     "Onset: The next Secret you play this turn costs (0).",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectFreeNextSecret, Target: TargetNone}}}},

	// --- Spells ---

	{ID: "frost_tempest", Name: "Frost Tempest", Type: TypeSpell, Class: ClassMage, Rarity: RarityRare, Cost: 6,
		Text:   "Deal 2 damage to all enemy minions and Freeze them.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaEnemyMinions, Freeze: true}},

	{ID: "frostshear", Name: "Frostshear", Type: TypeSpell, Class: ClassMage, Rarity: RarityCommon, Cost: 3,
		Text:   "Freeze a minion and the minions next to it, and deal 1 damage to them.",
		Effect: &Effect{Kind: EffectDamage, Amount: 1, Target: TargetMinion, Area: AreaSplash, Freeze: true}},

	{ID: "pyrecataclysm", Name: "Pyrecataclysm", Type: TypeSpell, Class: ClassMage, Rarity: RarityEpic, Cost: 10,
		Text:   "Deal 10 damage.",
		Effect: &Effect{Kind: EffectDamage, Amount: 10, Target: TargetAny}},

	{ID: "codex_of_insight", Name: "Codex of Insight", Type: TypeSpell, Class: ClassMage, Rarity: RarityRare, Cost: 1,
		Text:   "Add a random Mage spell to your hand.",
		Effect: &Effect{Kind: EffectGenerateRandom, Target: TargetNone, GenClass: ClassMage, GenType: TypeSpell}},

	{ID: "glacial_splinter", Name: "Glacial Splinter", Type: TypeSpell, Class: ClassMage, Rarity: RarityCommon, Cost: 2,
		Text:   "Deal 2 damage to a minion. If it was Frozen, draw a card.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetMinion, DrawIfFrozen: 1}},

	{ID: "frostlance", Name: "Frostlance", Type: TypeSpell, Class: ClassMage, Rarity: RarityCommon, Cost: 1,
		Text:   "Freeze a character. If it was already Frozen, deal 4 damage instead.",
		Effect: &Effect{Kind: EffectDamage, Target: TargetAny, Freeze: true, FrozenDamage: 4}},

	{ID: "frostward_aegis", Name: "Frostward Aegis", Type: TypeSecret, Class: ClassMage, Rarity: RarityEpic, Cost: 3,
		Text:   "Secret: When your hero takes fatal damage, prevent it and become Immune this turn.",
		Secret: &SecretDef{Trigger: OnFatalDamage, Kind: SecretIceBlock}},

	// --- Tokens (summon/generate-only; excluded from decks and Seek) ---

	// Pyrebolt is the burn spell `emberforge_magus` adds to hand. It is
	// NOT a Classic collectible, so it is a token (never built into a deck).
	{ID: "pyrebolt", Name: "Pyrebolt", Type: TypeSpell, Class: ClassMage, Cost: 4, Token: true,
		Text:   "Deal 6 damage.",
		Effect: &Effect{Kind: EffectDamage, Amount: 6, Target: TargetAny}},

	// Conjured Decoy is the 1/3 Decoy Ward summons to soak a redirected spell.
	{ID: "conjured_decoy", Name: "Conjured Decoy", Type: TypeMinion, Class: ClassMage, Cost: 1, Attack: 1, Health: 3, Token: true},

	// --- Hero power (fixed; not collectible, never in a hand or deck) ---

	{ID: "fire_dart", Name: "Fire Dart", Type: TypeHeroPower, Class: ClassMage, Cost: 2,
		Text:   "Hero Power: Deal 1 damage.",
		Effect: &Effect{Kind: EffectDamage, Amount: 1, Target: TargetAny}},
}
