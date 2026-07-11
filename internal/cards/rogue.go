package cards

// rogueCards are the Rogue class cards (eighth playable class). Scope is the
// genre-staple Basic + Classic Rogue set using the ORIGINAL / beta-era numbers
// (cheap Fan of Knives, 4-mana Assassinate, 5-mana Sprint, a 4-mana 2/5
// weapon, 2-mana Blade Flurry) transcribed from the user's screenshots rather
// than later reworks. Mechanics are 1:1 with well-worn staples while names + art
// + rules text are wholly original (see HANDOFF "Legal rules"). The class theme
// is shadows, daggers, poison, and cheap tempo.
//
// Basic cards carry NO rarity (empty Rarity = no gem); Classic cards do. HS
// "Combo" is our keyword Chain: a bonus that triggers only if the caster already
// played another card this turn. A spell's ChainEffect replaces its base Effect
// when Chain is active; a minion/weapon's ChainOnset is a battlecry that only
// fires under Chain. Kept keywords: Stealth, Poisonous. The thief token is the
// Guild Thug; the hero-power dagger is the Night Shiv.
var rogueCards = []Card{
	// --- Spells (Basic) ---

	{ID: "blindside", Name: "Blindside", Type: TypeSpell, Class: ClassRogue, Cost: 0,
		Text:   "Deal 2 damage to an undamaged minion.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetMinion, ReqUndamaged: true}},

	{ID: "venom_coat", Name: "Venom Coat", Type: TypeSpell, Class: ClassRogue, Cost: 1,
		Text:   "Give your weapon +2 Attack.",
		Effect: &Effect{Kind: EffectBuffWeapon, BuffAtk: 2, Target: TargetNone}},

	{ID: "sly_jab", Name: "Sly Jab", Type: TypeSpell, Class: ClassRogue, Cost: 1,
		Text:   "Deal 3 damage to the enemy hero.",
		Effect: &Effect{Kind: EffectDamage, Amount: 3, Target: TargetNone, Area: AreaEnemyHero}},

	{ID: "waylay", Name: "Waylay", Type: TypeSpell, Class: ClassRogue, Cost: 2,
		Text:   "Return an enemy minion to your opponent's hand.",
		Effect: &Effect{Kind: EffectBounce, Target: TargetEnemyMinion}},

	{ID: "quickstab", Name: "Quickstab", Type: TypeSpell, Class: ClassRogue, Cost: 2,
		Text:   "Deal 1 damage. Draw a card.",
		Effect: &Effect{Kind: EffectDamage, Amount: 1, Target: TargetAny, ThenDraw: 1}},

	{ID: "blade_fan", Name: "Blade Fan", Type: TypeSpell, Class: ClassRogue, Cost: 3,
		Text:   "Deal 1 damage to all enemy minions. Draw a card.",
		Effect: &Effect{Kind: EffectDamage, Amount: 1, Target: TargetNone, Area: AreaEnemyMinions, ThenDraw: 1}},

	{ID: "final_cut", Name: "Final Cut", Type: TypeSpell, Class: ClassRogue, Cost: 4,
		Text:   "Destroy an enemy minion.",
		Effect: &Effect{Kind: EffectDestroy, Target: TargetEnemyMinion}},

	{ID: "hasty_dash", Name: "Hasty Dash", Type: TypeSpell, Class: ClassRogue, Cost: 5,
		Text:   "Draw 4 cards.",
		Effect: &Effect{Kind: EffectDraw, Amount: 4, Target: TargetNone}},

	{ID: "vanishing_act", Name: "Vanishing Act", Type: TypeSpell, Class: ClassRogue, Cost: 6,
		Text:   "Return all minions to their owner's hand.",
		Effect: &Effect{Kind: EffectBounceAll, Target: TargetNone}},

	// Assassin's Edge — a vanilla weapon (beta 2/5).
	{ID: "assassins_edge", Name: "Assassin's Edge", Type: TypeWeapon, Class: ClassRogue, Cost: 4, Attack: 2, Durability: 5},

	// --- Spells (Classic) ---

	{ID: "shadow_veil", Name: "Shadow Veil", Type: TypeSpell, Class: ClassRogue, Rarity: RarityCommon, Cost: 1,
		Text:   "Give your minions Stealth until your next turn.",
		Effect: &Effect{Kind: EffectBuff, Target: TargetNone, Area: AreaFriendlyMinions, Grant: []Keyword{KeywordStealth}, TempUntilNextTurn: true}},

	{ID: "groundwork", Name: "Groundwork", Type: TypeSpell, Class: ClassRogue, Rarity: RarityEpic, Cost: 0,
		Text:   "The next spell you cast this turn costs (2) less.",
		Effect: &Effect{Kind: EffectDiscountNextSpell, Amount: 2, Target: TargetNone}},

	{ID: "slip_away", Name: "Slip Away", Type: TypeSpell, Class: ClassRogue, Rarity: RarityCommon, Cost: 0,
		Text:   "Return a friendly minion to your hand. It costs (2) less.",
		Effect: &Effect{Kind: EffectBounce, Target: TargetFriendlyMinion, BounceCostDelta: -2}},

	{ID: "cold_venom", Name: "Cold Venom", Type: TypeSpell, Class: ClassRogue, Rarity: RarityCommon, Cost: 1,
		Text:        "Give a minion +2 Attack. Chain: +4 Attack instead.",
		Effect:      &Effect{Kind: EffectBuff, BuffAtk: 2, Target: TargetMinion},
		ChainEffect: &Effect{Kind: EffectBuff, BuffAtk: 4, Target: TargetMinion}},

	{ID: "pickpocket", Name: "Pickpocket", Type: TypeSpell, Class: ClassRogue, Rarity: RarityCommon, Cost: 1,
		Text:   "Add a random card from another class to your hand.",
		Effect: &Effect{Kind: EffectPickpocket, Target: TargetNone}},

	{ID: "turncoat", Name: "Turncoat", Type: TypeSpell, Class: ClassRogue, Rarity: RarityCommon, Cost: 2,
		Text:   "Force an enemy minion to deal its damage to the minions next to it.",
		Effect: &Effect{Kind: EffectForceAttackNeighbors, Target: TargetEnemyMinion}},

	{ID: "blade_whirl", Name: "Blade Whirl", Type: TypeSpell, Class: ClassRogue, Rarity: RarityCommon, Cost: 2,
		Text:   "Destroy your weapon and deal its damage to all enemy minions.",
		Effect: &Effect{Kind: EffectWeaponSweep, Target: TargetNone}},

	{ID: "lacerate", Name: "Lacerate", Type: TypeSpell, Class: ClassRogue, Rarity: RarityCommon, Cost: 2,
		Text:        "Deal 2 damage. Chain: Deal 4 damage instead.",
		Effect:      &Effect{Kind: EffectDamage, Amount: 2, Target: TargetAny},
		ChainEffect: &Effect{Kind: EffectDamage, Amount: 4, Target: TargetAny}},

	{ID: "skullcrack", Name: "Skullcrack", Type: TypeSpell, Class: ClassRogue, Rarity: RarityRare, Cost: 3,
		Text:               "Deal 2 damage to the enemy hero. Chain: Return this to your hand next turn.",
		Effect:             &Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaEnemyHero},
		ChainReturnsToHand: true},

	// --- Minions ---

	{ID: "guild_ringleader", Name: "Guild Ringleader", Type: TypeMinion, Class: ClassRogue, Rarity: RarityCommon, Cost: 2, Attack: 2, Health: 2,
		Text:       "Chain: Summon a 2/1 Guild Thug.",
		ChainOnset: &Effect{Kind: EffectSummon, Summon: "guild_thug", Target: TargetNone}},

	{ID: "guild_agent", Name: "Guild Agent", Type: TypeMinion, Class: ClassRogue, Rarity: RarityRare, Cost: 3, Attack: 3, Health: 3,
		Text:       "Chain: Deal 3 damage.",
		ChainOnset: &Effect{Kind: EffectDamage, Amount: 3, Target: TargetAny}},

	{ID: "shadowlord_vex", Name: "Vex the Shadowlord", Type: TypeMinion, Class: ClassRogue, Rarity: RarityLegendary, Cost: 3, Attack: 2, Health: 2,
		Text:       "Chain: Gain +2/+2 for each other card you've played this turn.",
		ChainOnset: &Effect{Kind: EffectBuff, BuffAtk: 2, BuffHP: 2, Target: TargetSelf, PerCardPlayedThisTurn: true}},

	{ID: "masked_infiltrator", Name: "Masked Infiltrator", Type: TypeMinion, Class: ClassRogue, Rarity: RarityRare, Cost: 4, Attack: 4, Health: 4,
		Text:     "Onset: Give a friendly minion Stealth.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, Target: TargetFriendlyMinion, Grant: []Keyword{KeywordStealth}}}}},

	{ID: "plague_carrier", Name: "Plague Carrier", Type: TypeMinion, Class: ClassRogue, Rarity: RarityRare, Cost: 4, Attack: 3, Health: 3, Tribe: TribeUndead,
		Text:     "Onset: Give a friendly minion Poisonous.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, Target: TargetFriendlyMinion, Grant: []Keyword{KeywordPoisonous}}}}},

	{ID: "snatcher_brute", Name: "Snatcher Brute", Type: TypeMinion, Class: ClassRogue, Rarity: RarityEpic, Cost: 6, Attack: 5, Health: 3, Tribe: TribeUndead,
		Text:       "Chain: Return a minion to its owner's hand.",
		ChainOnset: &Effect{Kind: EffectBounce, Target: TargetMinion}},

	// --- Weapons (Classic) ---

	{ID: "ruin_dagger", Name: "Ruin Dagger", Type: TypeWeapon, Class: ClassRogue, Rarity: RarityEpic, Cost: 3, Attack: 2, Durability: 2,
		Text:       "Onset: Deal 1 damage. Chain: Deal 2 instead.",
		Triggers:   []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 1, Target: TargetAny}}},
		ChainOnset: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetAny}},

	// --- Hero power ---

	{ID: "hone_blade", Name: "Hone Blade", Type: TypeHeroPower, Class: ClassRogue, Cost: 2,
		Text:   "Hero Power: Equip a 1/2 Dagger.",
		Effect: &Effect{Kind: EffectEquip, EquipWeapon: "night_shiv", Target: TargetNone}},

	// --- Tokens (summon-only / generated; excluded from decks) ---

	{ID: "night_shiv", Name: "Night Shiv", Type: TypeWeapon, Class: ClassRogue, Cost: 1, Attack: 1, Durability: 2, Token: true},

	{ID: "guild_thug", Name: "Guild Thug", Type: TypeMinion, Class: ClassRogue, Cost: 2, Attack: 2, Health: 1, Token: true},
}
