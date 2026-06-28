package cards

// priestCards are the Priest class cards (fifth playable class). Target scope is
// the genre-staple Basic + Classic Priest set; mechanics are 1:1 with well-worn
// staples while names + art + rules text are wholly original (see HANDOFF "Legal
// rules"). The class theme is Holy healing + Shadow mind-control / card-copy.
//
// Basic cards carry NO rarity (empty Rarity = no gem); Classic cards do.
var priestCards = []Card{
	// --- Spells ---

	{ID: "ring_of_renewal", Name: "Ring of Renewal", Type: TypeSpell, Class: ClassPriest, Rarity: RarityRare, Cost: 0,
		Text:   "Restore 4 Health to ALL minions.",
		Effect: &Effect{Kind: EffectHeal, Amount: 4, Target: TargetNone, Area: AreaAllMinions}},

	{ID: "hush", Name: "Hush", Type: TypeSpell, Class: ClassPriest, Rarity: RarityCommon, Cost: 0,
		Text:   "Silence a minion.",
		Effect: &Effect{Kind: EffectSilence, Target: TargetMinion}},

	{ID: "dawnward_sigil", Name: "Dawnward Sigil", Type: TypeSpell, Class: ClassPriest, Cost: 1,
		Text:   "Give a minion +2 Health. Draw a card.",
		Effect: &Effect{Kind: EffectBuff, BuffHP: 2, Target: TargetMinion, Then: &Effect{Kind: EffectDraw, Target: TargetNone}}},

	{ID: "searing_light", Name: "Searing Light", Type: TypeSpell, Class: ClassPriest, Cost: 1,
		Text:   "Deal 3 damage to a minion.",
		Effect: &Effect{Kind: EffectDamage, Amount: 3, Target: TargetMinion}},

	{ID: "pried_thought", Name: "Pried Thought", Type: TypeSpell, Class: ClassPriest, Cost: 1,
		Text:   "Put a copy of a random card in your opponent's hand into your hand.",
		Effect: &Effect{Kind: EffectCopyOppHand, Count: 1, Target: TargetNone}},

	{ID: "mending_light", Name: "Mending Light", Type: TypeSpell, Class: ClassPriest, Cost: 1,
		Text:   "Restore 5 Health to your hero.",
		Effect: &Effect{Kind: EffectHeal, Amount: 5, Target: TargetNone, Area: AreaFriendlyHero}},

	{ID: "soul_kindle", Name: "Soul Kindle", Type: TypeSpell, Class: ClassPriest, Cost: 1,
		Text:   "Change a minion's Attack to be equal to its Health.",
		Effect: &Effect{Kind: EffectSetAtkToHealth, Target: TargetMinion}},

	{ID: "gloom_word_demise", Name: "Gloom Word: Demise", Type: TypeSpell, Class: ClassPriest, Cost: 3,
		Text:   "Destroy a minion with 5 or more Attack.",
		Effect: &Effect{Kind: EffectDestroy, Target: TargetMinion, ReqAttack: 5}},

	{ID: "gloom_word_ache", Name: "Gloom Word: Ache", Type: TypeSpell, Class: ClassPriest, Cost: 2,
		Text:   "Destroy a minion with 3 or less Attack.",
		Effect: &Effect{Kind: EffectDestroy, Target: TargetMinion, ReqMaxAttack: 3}},

	{ID: "mind_larceny", Name: "Mind Larceny", Type: TypeSpell, Class: ClassPriest, Rarity: RarityCommon, Cost: 3,
		Text:   "Copy 2 cards in your opponent's deck and add them to your hand.",
		Effect: &Effect{Kind: EffectCopyOppDeck, Count: 2, Target: TargetNone}},

	{ID: "psychic_lance", Name: "Psychic Lance", Type: TypeSpell, Class: ClassPriest, Cost: 2,
		Text:   "Deal 5 damage to the enemy hero.",
		Effect: &Effect{Kind: EffectDamage, Amount: 5, Target: TargetNone, Area: AreaEnemyHero}},

	{ID: "soul_mirror", Name: "Soul Mirror", Type: TypeSpell, Class: ClassPriest, Rarity: RarityCommon, Cost: 2,
		Text:   "Double a minion's Health.",
		Effect: &Effect{Kind: EffectDoubleHealth, Target: TargetMinion}},

	{ID: "umbral_shift", Name: "Umbral Shift", Type: TypeSpell, Class: ClassPriest, Rarity: RarityEpic, Cost: 3,
		Text:   "Your Hero Power becomes 'Deal 2 damage.'",
		Effect: &Effect{Kind: EffectSetHeroPower, Target: TargetNone, HeroPowerID: "gloom_spike"}},

	{ID: "radiant_burst", Name: "Radiant Burst", Type: TypeSpell, Class: ClassPriest, Cost: 5,
		Text:   "Deal 2 damage to all enemy minions. Restore 2 Health to all friendly characters.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaEnemyMinions, Then: &Effect{Kind: EffectHeal, Amount: 2, Target: TargetNone, Area: AreaFriendlyChars}}},

	{ID: "gloom_thrall", Name: "Gloom Thrall", Type: TypeSpell, Class: ClassPriest, Rarity: RarityRare, Cost: 4,
		Text:   "Gain control of an enemy minion with 3 or less Attack until end of turn.",
		Effect: &Effect{Kind: EffectMindControl, Target: TargetEnemyMinion, ReqMaxAttack: 3, TempControl: true}},

	{ID: "zealots_blessing", Name: "Zealot's Blessing", Type: TypeSpell, Class: ClassPriest, Cost: 4,
		Text:   "Give a minion +2/+6.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 2, BuffHP: 6, Target: TargetMinion}},

	{ID: "great_hush", Name: "Great Hush", Type: TypeSpell, Class: ClassPriest, Rarity: RarityRare, Cost: 4,
		Text:   "Silence all enemy minions. Draw a card.",
		Effect: &Effect{Kind: EffectSilence, Target: TargetNone, Area: AreaEnemyMinions, ThenDraw: 1}},

	{ID: "phantom_summons", Name: "Phantom Summons", Type: TypeSpell, Class: ClassPriest, Rarity: RarityEpic, Cost: 4,
		Text:   "Put a copy of a random minion from your opponent's deck into the battlefield.",
		Effect: &Effect{Kind: EffectSummonFromOppDeck, Target: TargetNone}},

	{ID: "gloom_word_undoing", Name: "Gloom Word: Undoing", Type: TypeSpell, Class: ClassPriest, Rarity: RarityEpic, Cost: 4,
		Text:   "Destroy all minions with 5 or more Attack.",
		Effect: &Effect{Kind: EffectDestroy, Target: TargetNone, Area: AreaAllMinions, ReqAttack: 5}},

	{ID: "pyre_of_faith", Name: "Pyre of Faith", Type: TypeSpell, Class: ClassPriest, Rarity: RarityRare, Cost: 6,
		Text:   "Deal 5 damage. Restore 5 Health to your hero.",
		Effect: &Effect{Kind: EffectDamage, Amount: 5, Target: TargetAny, Then: &Effect{Kind: EffectHeal, Amount: 5, Target: TargetNone, Area: AreaFriendlyHero}}},

	{ID: "dominate_will", Name: "Dominate Will", Type: TypeSpell, Class: ClassPriest, Cost: 10,
		Text:   "Take control of an enemy minion.",
		Effect: &Effect{Kind: EffectMindControl, Target: TargetEnemyMinion}},

	// --- Minions ---

	{ID: "mindspun_wraith", Name: "Mindspun Wraith", Type: TypeMinion, Class: ClassPriest, Cost: 1, Attack: 1, Health: 2, Tribe: TribeUndead,
		Text:     "Onset: Copy a card in your opponent's deck and add it to your hand.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectCopyOppDeck, Count: 1, Target: TargetNone}}}},

	{ID: "dawnvale_acolyte", Name: "Dawnvale Acolyte", Type: TypeMinion, Class: ClassPriest, Cost: 1, Attack: 1, Health: 3,
		Text:     "Whenever a minion is healed, draw a card.",
		Triggers: []Trigger{{When: OnMinionHealed, Effect: Effect{Kind: EffectDraw, Target: TargetNone}}}},

	{ID: "crimson_subduer", Name: "Crimson Subduer", Type: TypeMinion, Class: ClassPriest, Rarity: RarityCommon, Cost: 1, Attack: 2, Health: 1,
		Text:     "Onset: Give an enemy minion -2 Attack until your next turn.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, BuffAtk: -2, Target: TargetEnemyMinion, TempUntilNextTurn: true}}}},

	{ID: "harborlight_chaplain", Name: "Harborlight Chaplain", Type: TypeMinion, Class: ClassPriest, Rarity: RarityCommon, Cost: 2, Attack: 2, Health: 3,
		Text:     "Onset: Give a friendly minion +2 Health.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, BuffHP: 2, Target: TargetFriendlyMinion}}}},

	{ID: "radiant_font", Name: "Radiant Font", Type: TypeMinion, Class: ClassPriest, Rarity: RarityRare, Cost: 2, Attack: 0, Health: 5,
		Text:     "At the start of your turn, restore 3 Health to a damaged friendly character.",
		Triggers: []Trigger{{When: OnTurnStart, Effect: Effect{Kind: EffectHeal, Amount: 3, Target: TargetRandomDamagedFriendly}}}},

	{ID: "lumen_wisp", Name: "Lumen Wisp", Type: TypeMinion, Class: ClassPriest, Rarity: RarityCommon, Cost: 3, Attack: 0, Health: 4, Tribe: TribeElemental,
		Text:            "This minion's Attack is always equal to its Health.",
		AtkEqualsHealth: true},

	{ID: "auralast_zealot", Name: "Auralast Zealot", Type: TypeMinion, Class: ClassPriest, Rarity: RarityRare, Cost: 4, Attack: 3, Health: 5, Tribe: TribeRiftborn,
		Text:            "Your cards and powers that restore Health now deal damage instead.",
		HealsDealDamage: true},

	{ID: "prism_moth", Name: "Prism Moth", Type: TypeMinion, Class: ClassPriest, Rarity: RarityEpic, Cost: 5, Attack: 4, Health: 4, Tribe: TribeBeast,
		Text:     "Onset: If your deck has only odd-Cost cards, double the Health of your other minions.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDoubleHealth, Target: TargetNone, Area: AreaOtherFriendlyMinions, ReqDeckAllOdd: true}}}},

	{ID: "sanctum_warden", Name: "Sanctum Warden", Type: TypeMinion, Class: ClassPriest, Rarity: RarityCommon, Cost: 5, Attack: 5, Health: 6,
		Text:     "Onset: Give a friendly minion +3 Health.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, BuffHP: 3, Target: TargetFriendlyMinion}}}},

	{ID: "cabal_mindbinder", Name: "Cabal Mindbinder", Type: TypeMinion, Class: ClassPriest, Rarity: RarityEpic, Cost: 6, Attack: 4, Health: 5,
		Text:     "Onset: Take control of an enemy minion that has 2 or less Attack.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectMindControl, Target: TargetEnemyMinion, ReqMaxAttack: 2}}}},

	{ID: "soulreaver_nyssa", Name: "Soulreaver Nyssa", Type: TypeMinion, Class: ClassPriest, Rarity: RarityLegendary, Cost: 7, Attack: 7, Health: 1,
		Text:     "Onset: Destroy a minion and gain its Health.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDevour, Target: TargetMinion}}}},

	{ID: "oracle_velneth", Name: "Oracle Velneth", Type: TypeMinion, Class: ClassPriest, Rarity: RarityLegendary, Cost: 7, Attack: 7, Health: 7, Tribe: TribeRiftborn,
		Text:              "Double the damage and healing of your spells and Hero Power.",
		DoublesCastOutput: true},

	// --- Hero power ---

	{ID: "mend", Name: "Mend", Type: TypeHeroPower, Class: ClassPriest, Cost: 2,
		Text:   "Hero Power: Restore 2 Health.",
		Effect: &Effect{Kind: EffectHeal, Amount: 2, Target: TargetAny}},

	// --- Tokens (replacement hero power from `umbral_shift`; not collectible) ---

	{ID: "gloom_spike", Name: "Gloom Spike", Type: TypeHeroPower, Class: ClassPriest, Cost: 2, Token: true,
		Text:   "Hero Power: Deal 2 damage.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetAny}},
}
