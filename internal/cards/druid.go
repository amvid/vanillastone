package cards

// druidCards are the Druid class cards (seventh playable class). Target scope is
// the genre-staple Basic + Classic Druid set, using the ORIGINAL Classic-era
// numbers (ramp Innervate, Charge/doomed Force-of-Nature treants, the 4/4
// self-buff Druid of the Claw, draw-2 Ancient of Lore) rather than later
// reworks. Mechanics are 1:1 with well-worn staples while names + art + rules
// text are wholly original (see HANDOFF "Legal rules"). The class theme is
// Nature — mana ramp, beasts, treants, claws/fangs, moonlight.
//
// Basic cards carry NO rarity (empty Rarity = no gem); Classic cards do. The
// "Choose One" mechanic is our keyword Duality (Card.Choices — two options, one
// picked at play time); the treant token is our Thornling.
var druidCards = []Card{
	// --- Spells (Basic) ---

	{ID: "mana_bloom", Name: "Mana Bloom", Type: TypeSpell, Class: ClassDruid, Cost: 0,
		Text:   "Gain 2 Mana Crystals this turn only.",
		Effect: &Effect{Kind: EffectMana, Amount: 2, Target: TargetNone}},

	{ID: "moonbeam", Name: "Moonbeam", Type: TypeSpell, Class: ClassDruid, Cost: 0,
		Text:   "Deal 1 damage.",
		Effect: &Effect{Kind: EffectDamage, Amount: 1, Target: TargetAny}},

	{ID: "feral_claws", Name: "Feral Claws", Type: TypeSpell, Class: ClassDruid, Cost: 1,
		Text:   "Give your hero +2 Attack this turn. Gain 2 Armor.",
		Effect: &Effect{Kind: EffectHeroAttack, Amount: 2, Target: TargetNone, Then: &Effect{Kind: EffectArmor, Amount: 2, Target: TargetNone}}},

	{ID: "wild_mark", Name: "Wild Mark", Type: TypeSpell, Class: ClassDruid, Cost: 2,
		Text:   "Give a minion +2/+2 and Taunt.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 2, BuffHP: 2, Target: TargetMinion, Grant: []Keyword{KeywordTaunt}}},

	{ID: "verdant_growth", Name: "Verdant Growth", Type: TypeSpell, Class: ClassDruid, Cost: 2,
		Text:   "Gain an empty Mana Crystal.",
		Effect: &Effect{Kind: EffectRampMana, Amount: 1, Empty: true, OverflowGenerate: "overflow_mana", Target: TargetNone}},

	{ID: "mending_touch", Name: "Mending Touch", Type: TypeSpell, Class: ClassDruid, Cost: 3,
		Text:   "Restore 8 Health.",
		Effect: &Effect{Kind: EffectHeal, Amount: 8, Target: TargetAny}},

	{ID: "feral_howl", Name: "Feral Howl", Type: TypeSpell, Class: ClassDruid, Cost: 3,
		Text:   "Give your characters +2 Attack this turn.",
		Effect: &Effect{Kind: EffectHeroAttack, Amount: 2, Target: TargetNone, Then: &Effect{Kind: EffectBuff, BuffAtk: 2, Target: TargetNone, Area: AreaFriendlyMinions, Temporary: true}}},

	{ID: "claw_sweep", Name: "Claw Sweep", Type: TypeSpell, Class: ClassDruid, Cost: 4,
		Text:   "Deal 4 damage to an enemy and 1 damage to all other enemies.",
		Effect: &Effect{Kind: EffectDamage, Amount: 4, OtherEnemyAmount: 1, Target: TargetEnemy}},

	{ID: "starbolt", Name: "Starbolt", Type: TypeSpell, Class: ClassDruid, Cost: 6,
		Text:   "Deal 5 damage. Draw a card.",
		Effect: &Effect{Kind: EffectDamage, Amount: 5, Target: TargetAny, ThenDraw: 1}},

	// --- Spells (Classic) ---

	{ID: "wild_reclaim", Name: "Wild Reclaim", Type: TypeSpell, Class: ClassDruid, Rarity: RarityCommon, Cost: 1,
		Text:   "Destroy a minion. Your opponent draws 2 cards.",
		Effect: &Effect{Kind: EffectDestroy, Target: TargetMinion, Then: &Effect{Kind: EffectDraw, Amount: 2, ToOpponent: true, Target: TargetNone}}},

	{ID: "savage_slash", Name: "Savage Slash", Type: TypeSpell, Class: ClassDruid, Rarity: RarityRare, Cost: 1,
		Text:   "Deal damage equal to your hero's Attack to a minion.",
		Effect: &Effect{Kind: EffectDamage, ScaleByHeroAttack: true, Target: TargetMinion}},

	{ID: "might_of_the_grove", Name: "Might of the Grove", Type: TypeSpell, Class: ClassDruid, Rarity: RarityCommon, Cost: 2,
		Text: "Duality - Give your minions +1/+1; or Summon a 3/2 Panther.",
		Choices: []Choice{
			{Text: "Give your minions +1/+1.", Effect: Effect{Kind: EffectBuff, BuffAtk: 1, BuffHP: 1, Target: TargetNone, Area: AreaFriendlyMinions}},
			{Text: "Summon a 3/2 Panther.", Effect: Effect{Kind: EffectSummon, Summon: "grove_panther", Target: TargetNone}},
		}},

	{ID: "thornlash", Name: "Thornlash", Type: TypeSpell, Class: ClassDruid, Rarity: RarityCommon, Cost: 2,
		Text: "Duality - Deal 3 damage to a minion; or 1 damage and draw a card.",
		Choices: []Choice{
			{Text: "Deal 3 damage to a minion.", Effect: Effect{Kind: EffectDamage, Amount: 3, Target: TargetMinion}},
			{Text: "Deal 1 damage to a minion and draw a card.", Effect: Effect{Kind: EffectDamage, Amount: 1, Target: TargetMinion, ThenDraw: 1}},
		}},

	{ID: "wild_endowment", Name: "Wild Endowment", Type: TypeSpell, Class: ClassDruid, Rarity: RarityCommon, Cost: 3,
		Text: "Duality - Give a minion +4 Attack; or +4 Health and Taunt.",
		Choices: []Choice{
			{Text: "Give a minion +4 Attack.", Effect: Effect{Kind: EffectBuff, BuffAtk: 4, Target: TargetMinion}},
			{Text: "Give a minion +4 Health and Taunt.", Effect: Effect{Kind: EffectBuff, BuffHP: 4, Target: TargetMinion, Grant: []Keyword{KeywordTaunt}}},
		}},

	{ID: "forest_soul", Name: "Forest Soul", Type: TypeSpell, Class: ClassDruid, Rarity: RarityCommon, Cost: 3,
		Text:   "Give your minions \"Final Gasp: Summon a 2/2 Thornling.\"",
		Effect: &Effect{Kind: EffectGrantFinalGasp, Target: TargetNone, Area: AreaFriendlyMinions, FinalGasp: &Effect{Kind: EffectSummon, Summon: "thornling", Target: TargetNone}}},

	{ID: "savage_bite", Name: "Savage Bite", Type: TypeSpell, Class: ClassDruid, Rarity: RarityRare, Cost: 4,
		Text:   "Give your hero +4 Attack this turn. Gain 4 Armor.",
		Effect: &Effect{Kind: EffectHeroAttack, Amount: 4, Target: TargetNone, Then: &Effect{Kind: EffectArmor, Amount: 4, Target: TargetNone}}},

	{ID: "verdant_bounty", Name: "Verdant Bounty", Type: TypeSpell, Class: ClassDruid, Rarity: RarityRare, Cost: 5,
		Text: "Duality - Gain 2 Mana Crystals; or Draw 3 cards.",
		Choices: []Choice{
			{Text: "Gain 2 Mana Crystals.", Effect: Effect{Kind: EffectRampMana, Amount: 2, Target: TargetNone}},
			{Text: "Draw 3 cards.", Effect: Effect{Kind: EffectDraw, Amount: 3, Target: TargetNone}},
		}},

	{ID: "star_rain", Name: "Star Rain", Type: TypeSpell, Class: ClassDruid, Rarity: RarityRare, Cost: 5,
		Text: "Duality - Deal 5 damage to a minion; or 2 damage to all enemy minions.",
		Choices: []Choice{
			{Text: "Deal 5 damage to a minion.", Effect: Effect{Kind: EffectDamage, Amount: 5, Target: TargetMinion}},
			{Text: "Deal 2 damage to all enemy minions.", Effect: Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaEnemyMinions}},
		}},

	{ID: "call_the_grove", Name: "Call the Grove", Type: TypeSpell, Class: ClassDruid, Rarity: RarityEpic, Cost: 6,
		Text:   "Summon three 2/2 Thornlings with Charge. At the end of the turn, destroy them.",
		Effect: &Effect{Kind: EffectSummon, Summon: "thornling", Count: 3, Target: TargetNone, SummonGrant: []Keyword{KeywordCharge}, SummonDestroyEndOfTurn: true}},

	{ID: "wilds_gift", Name: "Wild's Gift", Type: TypeSpell, Class: ClassDruid, Rarity: RarityCommon, Cost: 8,
		Text:   "Give your minions +2/+2 and Taunt.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 2, BuffHP: 2, Target: TargetNone, Area: AreaFriendlyMinions, Grant: []Keyword{KeywordTaunt}}},

	// --- Minions ---

	{ID: "grove_warden", Name: "Grove Warden", Type: TypeMinion, Class: ClassDruid, Rarity: RarityRare, Cost: 4, Attack: 2, Health: 4,
		Text: "Duality - Deal 2 damage; or Silence a minion.",
		Choices: []Choice{
			{Text: "Deal 2 damage.", Effect: Effect{Kind: EffectDamage, Amount: 2, Target: TargetAny}},
			{Text: "Silence a minion.", Effect: Effect{Kind: EffectSilence, Target: TargetMinion}},
		}},

	{ID: "clawform_druid", Name: "Clawform Druid", Type: TypeMinion, Class: ClassDruid, Rarity: RarityCommon, Cost: 5, Attack: 4, Health: 4,
		Text: "Duality - +2 Attack and Charge; or +2 Health and Taunt.",
		Choices: []Choice{
			{Text: "+2 Attack and Charge.", Effect: Effect{Kind: EffectBuff, BuffAtk: 2, Target: TargetSelf, Grant: []Keyword{KeywordCharge}}},
			{Text: "+2 Health and Taunt.", Effect: Effect{Kind: EffectBuff, BuffHP: 2, Target: TargetSelf, Grant: []Keyword{KeywordTaunt}}},
		}},

	{ID: "elder_of_wisdom", Name: "Elder of Wisdom", Type: TypeMinion, Class: ClassDruid, Rarity: RarityEpic, Cost: 7, Attack: 5, Health: 5,
		Text: "Duality - Draw 2 cards; or Restore 5 Health.",
		Choices: []Choice{
			{Text: "Draw 2 cards.", Effect: Effect{Kind: EffectDraw, Amount: 2, Target: TargetNone}},
			{Text: "Restore 5 Health to your hero.", Effect: Effect{Kind: EffectHeal, Amount: 5, Target: TargetNone, Area: AreaFriendlyHero}},
		}},

	{ID: "elder_of_battle", Name: "Elder of Battle", Type: TypeMinion, Class: ClassDruid, Rarity: RarityEpic, Cost: 7, Attack: 5, Health: 5,
		Text: "Duality - +5 Attack; or +5 Health and Taunt.",
		Choices: []Choice{
			{Text: "+5 Attack.", Effect: Effect{Kind: EffectBuff, BuffAtk: 5, Target: TargetSelf}},
			{Text: "+5 Health and Taunt.", Effect: Effect{Kind: EffectBuff, BuffHP: 5, Target: TargetSelf, Grant: []Keyword{KeywordTaunt}}},
		}},

	{ID: "sylvaros", Name: "Sylvaros", Type: TypeMinion, Class: ClassDruid, Rarity: RarityLegendary, Cost: 8, Attack: 5, Health: 8,
		Text: "Duality - Give your other minions +2/+2; or Summon two 2/2 Thornlings with Taunt.",
		Choices: []Choice{
			{Text: "Give your other minions +2/+2.", Effect: Effect{Kind: EffectBuff, BuffAtk: 2, BuffHP: 2, Target: TargetNone, Area: AreaOtherFriendlyMinions}},
			{Text: "Summon two 2/2 Thornlings with Taunt.", Effect: Effect{Kind: EffectSummon, Summon: "thornling", Count: 2, Target: TargetNone, SummonGrant: []Keyword{KeywordTaunt}}},
		}},

	{ID: "barkhide_colossus", Name: "Barkhide Colossus", Type: TypeMinion, Class: ClassDruid, Cost: 8, Attack: 8, Health: 8,
		Text:     "Taunt.",
		Keywords: []Keyword{KeywordTaunt}},

	// --- Hero power ---

	{ID: "wild_shape", Name: "Wild Shape", Type: TypeHeroPower, Class: ClassDruid, Cost: 2,
		Text:   "Hero Power: +1 Attack this turn. +1 Armor.",
		Effect: &Effect{Kind: EffectHeroAttack, Amount: 1, Target: TargetNone, Then: &Effect{Kind: EffectArmor, Amount: 1, Target: TargetNone}}},

	// --- Tokens (summon-only / generated; excluded from decks) ---

	{ID: "overflow_mana", Name: "Overflowing Mana", Type: TypeSpell, Class: ClassDruid, Cost: 0, Token: true,
		Text:   "Draw a card.",
		Effect: &Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone}},

	{ID: "grove_panther", Name: "Grove Panther", Type: TypeMinion, Class: ClassDruid, Cost: 3, Attack: 3, Health: 2, Tribe: TribeBeast, Token: true},

	{ID: "thornling", Name: "Thornling", Type: TypeMinion, Class: ClassDruid, Cost: 2, Attack: 2, Health: 2, Token: true},
}
