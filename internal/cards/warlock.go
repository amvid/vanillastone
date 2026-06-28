package cards

// warlockCards are the Warlock class cards (fourth playable class). Target scope is
// the genre-staple Basic + Classic Warlock set; mechanics are 1:1 with well-worn
// staples while names + art + rules text are wholly original (see HANDOFF "Legal
// rules"). The class theme is Demons, shadow, discard, and paying your own life
// force (self-damage / soul-tap) for power.
//
// Basic cards carry NO rarity (empty Rarity = no gem); Classic cards do. All
// Demon minions use TribeDemon.
var warlockCards = []Card{
	// --- Spells ---

	{ID: "dark_bargain", Name: "Dark Bargain", Type: TypeSpell, Class: ClassWarlock, Cost: 0,
		Text:   "Destroy a friendly Demon. Restore 5 Health to your hero.",
		Effect: &Effect{Kind: EffectDestroy, Target: TargetFriendlyMinion, ReqTribe: TribeDemon, HealHero: 5}},

	{ID: "creeping_rot", Name: "Creeping Rot", Type: TypeSpell, Class: ClassWarlock, Cost: 1,
		Text:   "Choose an enemy minion. At the start of your turn, destroy it.",
		Effect: &Effect{Kind: EffectCorrupt, Target: TargetEnemyMinion}},

	{ID: "mortal_whisper", Name: "Mortal Whisper", Type: TypeSpell, Class: ClassWarlock, Cost: 1,
		Text:   "Deal 1 damage to a minion. If that kills it, draw a card.",
		Effect: &Effect{Kind: EffectDamage, Amount: 1, DrawIfKills: true, Target: TargetMinion}},

	{ID: "soul_ember", Name: "Soul Ember", Type: TypeSpell, Class: ClassWarlock, Cost: 1,
		Text:   "Deal 4 damage. Discard a random card.",
		Effect: &Effect{Kind: EffectDamage, Amount: 4, DiscardRandom: 1, Target: TargetAny}},

	{ID: "forbidden_might", Name: "Forbidden Might", Type: TypeSpell, Class: ClassWarlock, Rarity: RarityCommon, Cost: 1,
		Text:   "Give a friendly minion +4/+4 until end of turn. Then, it dies.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 4, BuffHP: 4, Temporary: true, DestroyEndOfTurn: true, Target: TargetFriendlyMinion}},

	{ID: "dark_summons", Name: "Dark Summons", Type: TypeSpell, Class: ClassWarlock, Rarity: RarityCommon, Cost: 1,
		Text:   "Add a random Demon to your hand.",
		Effect: &Effect{Kind: EffectGenerateRandom, Target: TargetNone, GenType: TypeMinion, GenTribe: TribeDemon}},

	{ID: "hexfire", Name: "Hexfire", Type: TypeSpell, Class: ClassWarlock, Rarity: RarityCommon, Cost: 2,
		Text:   "Deal 2 damage to a minion. If it's a friendly Demon, give it +2/+2 instead.",
		Effect: &Effect{Kind: EffectDemonfire, Amount: 2, BuffAtk: 2, BuffHP: 2, Tribe: TribeDemon, Target: TargetMinion}},

	{ID: "siphon_vitae", Name: "Siphon Vitae", Type: TypeSpell, Class: ClassWarlock, Cost: 3,
		Text:   "Deal 2 damage. Restore 2 Health to your hero.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Lifesteal: true, Target: TargetAny}},

	{ID: "shadow_lance", Name: "Shadow Lance", Type: TypeSpell, Class: ClassWarlock, Cost: 3,
		Text:   "Deal 4 damage to a minion.",
		Effect: &Effect{Kind: EffectDamage, Amount: 4, Target: TargetMinion}},

	{ID: "call_the_brood", Name: "Call the Brood", Type: TypeSpell, Class: ClassWarlock, Rarity: RarityCommon, Cost: 3,
		Text:   "Draw 2 Demons from your deck.",
		Effect: &Effect{Kind: EffectTutorTribe, Tribe: TribeDemon, Count: 2, TutorFallback: "runt_imp", Target: TargetNone}},

	{ID: "infernal_blaze", Name: "Infernal Blaze", Type: TypeSpell, Class: ClassWarlock, Cost: 4,
		Text:   "Deal 3 damage to all characters.",
		Effect: &Effect{Kind: EffectDamage, Amount: 3, Target: TargetNone, Area: AreaAllCharacters}},

	{ID: "gloomflare", Name: "Gloomflare", Type: TypeSpell, Class: ClassWarlock, Rarity: RarityRare, Cost: 4,
		Text:   "Destroy a friendly minion and deal its Attack damage to all enemy minions.",
		Effect: &Effect{Kind: EffectShadowflame, Target: TargetFriendlyMinion}},

	{ID: "doom_kiss", Name: "Doom Kiss", Type: TypeSpell, Class: ClassWarlock, Rarity: RarityEpic, Cost: 5,
		Text:   "Deal 2 damage to a character. If that kills it, summon a random Demon.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetAny, SummonRandomIfKills: true, GenType: TypeMinion, GenTribe: TribeDemon}},

	{ID: "soul_harvest", Name: "Soul Harvest", Type: TypeSpell, Class: ClassWarlock, Rarity: RarityRare, Cost: 6,
		Text:   "Destroy a minion. Restore 3 Health to your hero.",
		Effect: &Effect{Kind: EffectDestroy, Target: TargetMinion, HealHero: 3}},

	{ID: "the_unmaking", Name: "The Unmaking", Type: TypeSpell, Class: ClassWarlock, Rarity: RarityEpic, Cost: 8,
		Text:   "Destroy all minions.",
		Effect: &Effect{Kind: EffectDestroy, Target: TargetNone, Area: AreaAllMinions}},

	// --- Minions ---

	{ID: "gloom_imp", Name: "Gloom Imp", Type: TypeMinion, Class: ClassWarlock, Rarity: RarityCommon, Cost: 1, Attack: 0, Health: 1, Tribe: TribeDemon,
		Text:     "Stealth. At the end of your turn, give another random friendly minion +1 Health.",
		Keywords: []Keyword{KeywordStealth},
		Triggers: []Trigger{{When: OnTurnEnd, Effect: Effect{Kind: EffectBuff, Target: TargetRandomFriendly, BuffHP: 1}}}},

	{ID: "ember_imp", Name: "Ember Imp", Type: TypeMinion, Class: ClassWarlock, Rarity: RarityCommon, Cost: 1, Attack: 3, Health: 2, Tribe: TribeDemon,
		Text:     "Onset: Deal 3 damage to your hero.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 3, Target: TargetNone, Area: AreaFriendlyHero}}}},

	{ID: "hollow_guardian", Name: "Hollow Guardian", Type: TypeMinion, Class: ClassWarlock, Cost: 1, Attack: 1, Health: 3, Tribe: TribeDemon,
		Text:     "Taunt.",
		Keywords: []Keyword{KeywordTaunt}},

	{ID: "gnawing_fiend", Name: "Gnawing Fiend", Type: TypeMinion, Class: ClassWarlock, Cost: 2, Attack: 4, Health: 3, Tribe: TribeDemon,
		Text:     "Onset: Discard a random card.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDiscard, Count: 1, Target: TargetNone}}}},

	{ID: "dark_gateway", Name: "Dark Gateway", Type: TypeMinion, Class: ClassWarlock, Rarity: RarityCommon, Cost: 4, Attack: 0, Health: 4,
		Text:     "Your minions cost (2) less, but not less than (1).",
		CostAura: &CostAura{Delta: -2, Scope: CostScopeFriendly, Type: TypeMinion, MinResult: 1}},

	{ID: "chained_brute", Name: "Chained Brute", Type: TypeMinion, Class: ClassWarlock, Rarity: RarityRare, Cost: 3, Attack: 3, Health: 5, Tribe: TribeDemon,
		Text:     "Taunt. Onset: Destroy one of your Mana Crystals.",
		Keywords: []Keyword{KeywordTaunt},
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectLoseMana, Amount: 1, Target: TargetNone}}}},

	{ID: "ravening_horror", Name: "Ravening Horror", Type: TypeMinion, Class: ClassWarlock, Rarity: RarityRare, Cost: 3, Attack: 3, Health: 3, Tribe: TribeDemon,
		Text:     "Onset: Destroy both adjacent minions and gain their Attack and Health.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectConsumeAdjacent, Target: TargetNone, Area: AreaAdjacent}}}},

	{ID: "abyssal_brute", Name: "Abyssal Brute", Type: TypeMinion, Class: ClassWarlock, Rarity: RarityEpic, Cost: 4, Attack: 5, Health: 6, Tribe: TribeDemon,
		Text:     "Onset: Deal 5 damage to your hero.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 5, Target: TargetNone, Area: AreaFriendlyHero}}}},

	{ID: "terror_fiend", Name: "Terror Fiend", Type: TypeMinion, Class: ClassWarlock, Cost: 5, Attack: 5, Health: 7, Tribe: TribeDemon,
		Text:     "Charge. Onset: Discard two random cards.",
		Keywords: []Keyword{KeywordCharge},
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDiscard, Count: 2, Target: TargetNone}}}},

	{ID: "dread_colossus", Name: "Dread Colossus", Type: TypeMinion, Class: ClassWarlock, Cost: 6, Attack: 6, Health: 6, Tribe: TribeDemon,
		Text:     "Onset: Deal 1 damage to all other characters.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 1, Target: TargetNone, Area: AreaOtherCharacters}}}},

	{ID: "dread_warden", Name: "Dread Warden", Type: TypeMinion, Class: ClassWarlock, Rarity: RarityRare, Cost: 7, Attack: 5, Health: 8, Tribe: TribeDemon,
		Text:     "Taunt. Your other Demons have +1 Attack.",
		Keywords: []Keyword{KeywordTaunt},
		Aura:     &Aura{Atk: 1, Tribe: TribeDemon}},

	{ID: "overlord_xathul", Name: "Overlord Xathul", Type: TypeMinion, Class: ClassWarlock, Rarity: RarityLegendary, Cost: 9, Attack: 3, Health: 15, Tribe: TribeDemon,
		Text: "Onset: Destroy your hero and replace it with Overlord Xathul.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectReplaceHero, Target: TargetNone, Amount: 15,
			HeroArt: "overlord_xathul_hero", HeroPowerID: "infernal_eruption", EquipWeapon: "gore_scythe"}}}},

	// --- Hero power ---

	{ID: "soul_tithe", Name: "Soul Tithe", Type: TypeHeroPower, Class: ClassWarlock, Cost: 2,
		Text:   "Hero Power: Draw a card and take 2 damage.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaFriendlyHero, ThenDraw: 1}},

	// --- Tokens (summoned/generated/equipped by cards above; not collectible) ---

	{ID: "runt_imp", Name: "Runt Imp", Type: TypeMinion, Class: ClassWarlock, Cost: 1, Attack: 1, Health: 1, Tribe: TribeDemon, Token: true},

	{ID: "abyss_horror", Name: "Abyss Horror", Type: TypeMinion, Class: ClassWarlock, Cost: 6, Attack: 6, Health: 6, Tribe: TribeDemon, Token: true},

	{ID: "gore_scythe", Name: "Gore Scythe", Type: TypeWeapon, Class: ClassWarlock, Cost: 3, Attack: 3, Durability: 8, Token: true},

	{ID: "infernal_eruption", Name: "Infernal Eruption", Type: TypeHeroPower, Class: ClassWarlock, Cost: 2, Token: true,
		Text:   "Hero Power: Summon a 6/6 Demon.",
		Effect: &Effect{Kind: EffectSummon, Summon: "abyss_horror", Target: TargetNone}},
}
