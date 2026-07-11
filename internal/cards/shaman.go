package cards

// shamanCards are the Shaman class cards (ninth / final playable class). Scope is
// the genre-staple Basic + Classic Shaman set using the OLD-ERA numbers
// transcribed from the user's screenshots. Mechanics are 1:1 with well-worn
// staples while names + art + rules text are wholly original (see HANDOFF "Legal
// rules"). The class theme is elemental shamanism — lightning/storms, fire,
// frost, living stone, carved totems, wolf & ancestral spirits, and Overload.
//
// Basic cards carry NO rarity (empty Rarity = no gem); Classic cards do. NEW
// mechanic: Overload (X) — playing the card locks X of your Mana Crystals at the
// START of your next turn. NEW tribe: Totem. Kept keyword renames already in the
// engine: Twinstrike (= Windfury), Aegis (= Divine Shield), Riftborn tribe
// (= Draenei). The hero power `call_totem` summons a random Totem token.
var shamanCards = []Card{
	// --- Spells (Basic) ---

	{ID: "spirit_mend", Name: "Spirit Mend", Type: TypeSpell, Class: ClassShaman, Cost: 0,
		Text: "Restore a minion to full Health and give it Taunt.",
		Effect: &Effect{Kind: EffectHeal, Amount: 999, Target: TargetMinion,
			Then: &Effect{Kind: EffectBuff, Target: TargetMinion, Grant: []Keyword{KeywordTaunt}}}},

	{ID: "totem_bulwark", Name: "Totem Bulwark", Type: TypeSpell, Class: ClassShaman, Cost: 0,
		Text:   "Give your Totems +2 Health.",
		Effect: &Effect{Kind: EffectBuff, BuffHP: 2, Target: TargetNone, Area: AreaFriendlyTribe, Tribe: TribeTotem}},

	{ID: "frost_jolt", Name: "Frost Jolt", Type: TypeSpell, Class: ClassShaman, Cost: 1,
		Text:   "Deal 1 damage to an enemy character and Freeze it.",
		Effect: &Effect{Kind: EffectDamage, Amount: 1, Target: TargetEnemy, Freeze: true}},

	{ID: "stonefury", Name: "Stonefury", Type: TypeSpell, Class: ClassShaman, Cost: 2,
		Text:   "Give a friendly character +3 Attack this turn.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 3, Target: TargetFriendlyChar, Temporary: true}},

	{ID: "galeforce", Name: "Galeforce", Type: TypeSpell, Class: ClassShaman, Cost: 2,
		Text:   "Give a minion Twinstrike.",
		Effect: &Effect{Kind: EffectBuff, Target: TargetMinion, Grant: []Keyword{KeywordTwinstrike}}},

	{ID: "toadcurse", Name: "Toadcurse", Type: TypeSpell, Class: ClassShaman, Cost: 3,
		Text:   "Transform a minion into a 0/1 Frog with Taunt.",
		Effect: &Effect{Kind: EffectTransform, Transform: "spirit_frog", Target: TargetMinion}},

	{ID: "bloodsurge", Name: "Bloodsurge", Type: TypeSpell, Class: ClassShaman, Cost: 5,
		Text:   "Give your minions +3 Attack this turn.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 3, Target: TargetNone, Area: AreaFriendlyMinions, Temporary: true}},

	// --- Minions (Basic) ---

	{ID: "embertongue_totem", Name: "Embertongue Totem", Type: TypeMinion, Class: ClassShaman, Cost: 2, Attack: 0, Health: 3, Tribe: TribeTotem,
		Text: "Adjacent minions have +2 Attack.",
		Aura: &Aura{Atk: 2, Adjacent: true}},

	{ID: "galecaller", Name: "Galecaller", Type: TypeMinion, Class: ClassShaman, Cost: 4, Attack: 3, Health: 3, Tribe: TribeRiftborn,
		Text:     "Onset: Give a friendly minion Twinstrike.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, Target: TargetFriendlyMinion, Grant: []Keyword{KeywordTwinstrike}}}}},

	{ID: "cinder_elemental", Name: "Cinder Elemental", Type: TypeMinion, Class: ClassShaman, Cost: 6, Attack: 6, Health: 5, Tribe: TribeElemental,
		Text:     "Onset: Deal 4 damage.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 4, Target: TargetAny}}}},

	// --- Spells (Classic) ---

	{ID: "stonejolt", Name: "Stonejolt", Type: TypeSpell, Class: ClassShaman, Rarity: RarityCommon, Cost: 1,
		Text: "Silence a minion, then deal 1 damage to it.",
		Effect: &Effect{Kind: EffectSilence, Target: TargetMinion,
			Then: &Effect{Kind: EffectDamage, Amount: 1, Target: TargetMinion}}},

	{ID: "split_bolt", Name: "Split Bolt", Type: TypeSpell, Class: ClassShaman, Rarity: RarityCommon, Cost: 1, Overload: 2,
		Text:   "Deal 2 damage to 2 random enemy minions. Overload (2).",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaRandomEnemyMinion, Count: 2}},

	{ID: "voltstrike", Name: "Voltstrike", Type: TypeSpell, Class: ClassShaman, Rarity: RarityCommon, Cost: 1, Overload: 1,
		Text:   "Deal 3 damage. Overload (1).",
		Effect: &Effect{Kind: EffectDamage, Amount: 3, Target: TargetAny}},

	{ID: "spirit_bond", Name: "Spirit Bond", Type: TypeSpell, Class: ClassShaman, Rarity: RarityRare, Cost: 2,
		Text: "Give a minion \"Final Gasp: Resummon this minion.\"",
		Effect: &Effect{Kind: EffectGrantFinalGasp, Target: TargetMinion,
			FinalGasp: &Effect{Kind: EffectSummon, SummonSelf: true, Target: TargetNone}}},

	{ID: "distant_sight", Name: "Distant Sight", Type: TypeSpell, Class: ClassShaman, Rarity: RarityEpic, Cost: 3,
		Text:   "Draw a card. That card costs (3) less.",
		Effect: &Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone, DrawCostDelta: -3}},

	{ID: "wolfspirit_call", Name: "Wolfspirit Call", Type: TypeSpell, Class: ClassShaman, Rarity: RarityRare, Cost: 3, Overload: 1,
		Text:   "Summon two 2/3 Spirit Wolves with Taunt. Overload (1).",
		Effect: &Effect{Kind: EffectSummon, Summon: "spirit_wolf", Count: 2, Target: TargetNone}},

	{ID: "magma_burst", Name: "Magma Burst", Type: TypeSpell, Class: ClassShaman, Rarity: RarityRare, Cost: 3, Overload: 2,
		Text:   "Deal 5 damage. Overload (2).",
		Effect: &Effect{Kind: EffectDamage, Amount: 5, Target: TargetAny}},

	{ID: "tempest_surge", Name: "Tempest Surge", Type: TypeSpell, Class: ClassShaman, Rarity: RarityRare, Cost: 3, Overload: 1,
		Text:   "Deal 3 damage to all enemy minions. Overload (1).",
		Effect: &Effect{Kind: EffectDamage, Amount: 3, Target: TargetNone, Area: AreaEnemyMinions}},

	// --- Weapons (Classic) ---

	{ID: "tempest_axe", Name: "Tempest Axe", Type: TypeWeapon, Class: ClassShaman, Rarity: RarityCommon, Cost: 2, Attack: 2, Durability: 3, Overload: 1,
		Text: "Overload (1)."},

	{ID: "ruinhammer", Name: "Ruinhammer", Type: TypeWeapon, Class: ClassShaman, Rarity: RarityEpic, Cost: 5, Attack: 2, Durability: 8, Overload: 2,
		Text: "Twinstrike, Overload (2).", Keywords: []Keyword{KeywordTwinstrike}},

	// --- Minions (Classic) ---

	{ID: "gale_wisp", Name: "Gale Wisp", Type: TypeMinion, Class: ClassShaman, Rarity: RarityCommon, Cost: 1, Attack: 3, Health: 1, Tribe: TribeElemental, Overload: 2,
		Text: "Twinstrike. Overload (2).", Keywords: []Keyword{KeywordTwinstrike}},

	{ID: "tidewater_totem", Name: "Tidewater Totem", Type: TypeMinion, Class: ClassShaman, Rarity: RarityRare, Cost: 3, Attack: 0, Health: 3, Tribe: TribeTotem,
		Text:     "At the end of your turn, draw a card.",
		Triggers: []Trigger{{When: OnTurnEnd, Effect: Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone}}}},

	{ID: "riftbound_elemental", Name: "Riftbound Elemental", Type: TypeMinion, Class: ClassShaman, Rarity: RarityCommon, Cost: 3, Attack: 3, Health: 4, Tribe: TribeElemental,
		Text:     "After you play a card with Overload, gain +1/+1.",
		Triggers: []Trigger{{When: OnPlayOverload, Effect: Effect{Kind: EffectBuff, Target: TargetSelf, BuffAtk: 1, BuffHP: 1}}}},

	{ID: "bedrock_elemental", Name: "Bedrock Elemental", Type: TypeMinion, Class: ClassShaman, Rarity: RarityEpic, Cost: 5, Attack: 7, Health: 8, Tribe: TribeElemental, Overload: 2,
		Text: "Taunt. Overload (2).", Keywords: []Keyword{KeywordTaunt}},

	{ID: "zephiron_stormlord", Name: "Zephiron the Stormlord", Type: TypeMinion, Class: ClassShaman, Rarity: RarityLegendary, Cost: 8, Attack: 3, Health: 6, Tribe: TribeElemental,
		Text:     "Charge, Aegis, Taunt, Twinstrike.",
		Keywords: []Keyword{KeywordCharge, KeywordAegis, KeywordTaunt, KeywordTwinstrike}},

	// --- Hero power ---

	{ID: "call_totem", Name: "Call Totem", Type: TypeHeroPower, Class: ClassShaman, Cost: 2,
		Text: "Hero Power: Summon a random Totem.",
		Effect: &Effect{Kind: EffectSummonTotem, Target: TargetNone,
			GenIDs: []string{"mending_totem", "ember_totem", "stoneshell_totem", "stormcrest_totem"}}},

	// --- Tokens (summon-only / generated; excluded from decks) ---

	{ID: "mending_totem", Name: "Mending Totem", Type: TypeMinion, Class: ClassShaman, Cost: 1, Attack: 0, Health: 2, Tribe: TribeTotem, Token: true,
		Text:     "At the end of your turn, restore 1 Health to all friendly minions.",
		Triggers: []Trigger{{When: OnTurnEnd, Effect: Effect{Kind: EffectHeal, Amount: 1, Target: TargetNone, Area: AreaFriendlyMinions}}}},

	{ID: "ember_totem", Name: "Ember Totem", Type: TypeMinion, Class: ClassShaman, Cost: 1, Attack: 1, Health: 1, Tribe: TribeTotem, Token: true},

	{ID: "stoneshell_totem", Name: "Stoneshell Totem", Type: TypeMinion, Class: ClassShaman, Cost: 1, Attack: 0, Health: 2, Tribe: TribeTotem, Token: true,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},

	{ID: "stormcrest_totem", Name: "Stormcrest Totem", Type: TypeMinion, Class: ClassShaman, Cost: 1, Attack: 0, Health: 2, Tribe: TribeTotem, Token: true, SpellDamage: 1,
		Text: "Spell Damage +1."},

	{ID: "spirit_wolf", Name: "Spirit Wolf", Type: TypeMinion, Class: ClassShaman, Cost: 2, Attack: 2, Health: 3, Token: true,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},

	{ID: "spirit_frog", Name: "Frog", Type: TypeMinion, Class: ClassShaman, Cost: 0, Attack: 0, Health: 1, Tribe: TribeBeast, Token: true,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},
}
