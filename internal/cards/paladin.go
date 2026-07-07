package cards

// paladinCards are the Paladin class cards (sixth playable class). Target scope is
// the genre-staple Basic + Classic Paladin set; mechanics are 1:1 with well-worn
// staples while names + art + rules text are wholly original (see HANDOFF "Legal
// rules"). The class theme is Holy Light — blessings, protection (Aegis), weapons,
// and justice/retribution Secrets.
//
// Basic cards carry NO rarity (empty Rarity = no gem); Classic cards do. HS
// "Divine Shield" is our keyword Aegis; HS "Draenei" is our tribe Riftborn.
var paladinCards = []Card{
	// --- Spells ---

	{ID: "valor_blessing", Name: "Blessing of Valor", Type: TypeSpell, Class: ClassPaladin, Cost: 1,
		Text:   "Give a minion +3 Attack.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 3, Target: TargetMinion}},

	{ID: "warding_hand", Name: "Warding Hand", Type: TypeSpell, Class: ClassPaladin, Cost: 1,
		Text:   "Give a minion Aegis.",
		Effect: &Effect{Kind: EffectBuff, Target: TargetMinion, Grant: []Keyword{KeywordAegis}}},

	{ID: "meekness", Name: "Meekness", Type: TypeSpell, Class: ClassPaladin, Cost: 1,
		Text:   "Change a minion's Attack to 1.",
		Effect: &Effect{Kind: EffectSetAttack, Amount: 1, Target: TargetMinion}},

	{ID: "insight_blessing", Name: "Blessing of Insight", Type: TypeSpell, Class: ClassPaladin, Rarity: RarityCommon, Cost: 1,
		Text:   "Choose a minion. Whenever it attacks, draw a card.",
		Effect: &Effect{Kind: EffectGrantDrawOnAttack, Target: TargetMinion}},

	{ID: "radiant_mending", Name: "Radiant Mending", Type: TypeSpell, Class: ClassPaladin, Cost: 2,
		Text:   "Restore 8 Health to your hero.",
		Effect: &Effect{Kind: EffectHeal, Amount: 8, Target: TargetNone, Area: AreaFriendlyHero}},

	{ID: "great_leveling", Name: "Great Leveling", Type: TypeSpell, Class: ClassPaladin, Rarity: RarityRare, Cost: 2,
		Text:   "Change the Health of ALL minions to 1.",
		Effect: &Effect{Kind: EffectSetHealth, Amount: 1, Target: TargetNone, Area: AreaAllMinions}},

	{ID: "wrathhammer", Name: "Wrathful Hammer", Type: TypeSpell, Class: ClassPaladin, Cost: 3,
		Text:   "Deal 3 damage. Draw a card.",
		Effect: &Effect{Kind: EffectDamage, Amount: 3, Target: TargetAny, ThenDraw: 1}},

	{ID: "zealots_verdict", Name: "Zealot's Verdict", Type: TypeSpell, Class: ClassPaladin, Cost: 5,
		Text:   "Draw a card and deal damage equal to its Cost to a minion.",
		Effect: &Effect{Kind: EffectDrawAndBolt, Target: TargetMinion}},

	{ID: "providence", Name: "Providence", Type: TypeSpell, Class: ClassPaladin, Rarity: RarityRare, Cost: 3,
		Text:   "Draw cards until you have as many in hand as your opponent.",
		Effect: &Effect{Kind: EffectDrawToOpponent, Target: TargetNone}},

	{ID: "hallowed_ground", Name: "Hallowed Ground", Type: TypeSpell, Class: ClassPaladin, Cost: 4,
		Text:   "Deal 2 damage to all enemies.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaEnemyChars}},

	{ID: "royal_blessing", Name: "Blessing of Royalty", Type: TypeSpell, Class: ClassPaladin, Cost: 4,
		Text:   "Give a minion +4/+4.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 4, BuffHP: 4, Target: TargetMinion}},

	{ID: "exalted_might", Name: "Exalted Might", Type: TypeSpell, Class: ClassPaladin, Rarity: RarityRare, Cost: 5,
		Text:   "Double a minion's Attack.",
		Effect: &Effect{Kind: EffectDoubleAttack, Target: TargetMinion}},

	{ID: "aegis_hymn", Name: "Hymn of Aegis", Type: TypeSpell, Class: ClassPaladin, Rarity: RarityRare, Cost: 5,
		Text:   "Give your minions Aegis.",
		Effect: &Effect{Kind: EffectBuff, Target: TargetNone, Area: AreaFriendlyMinions, Grant: []Keyword{KeywordAegis}}},

	{ID: "avenging_light", Name: "Avenging Light", Type: TypeSpell, Class: ClassPaladin, Rarity: RarityEpic, Cost: 6,
		Text:   "Deal 8 damage randomly split among all enemies.",
		Effect: &Effect{Kind: EffectMissiles, Amount: 1, Count: 8, Target: TargetNone, Area: AreaEnemyChars}},

	{ID: "laying_of_hands", Name: "Laying of Hands", Type: TypeSpell, Class: ClassPaladin, Rarity: RarityEpic, Cost: 8,
		Text:   "Restore 8 Health to your hero. Draw 3 cards.",
		Effect: &Effect{Kind: EffectHeal, Amount: 8, Target: TargetNone, Area: AreaFriendlyHero, Then: &Effect{Kind: EffectDraw, Amount: 3, Target: TargetNone}}},

	// --- Secrets ---

	{ID: "retribution_vow", Name: "Vow of Retribution", Type: TypeSecret, Class: ClassPaladin, Rarity: RarityCommon, Cost: 1,
		Text:   "Secret: When your hero takes damage, deal that much to the enemy hero.",
		Secret: &SecretDef{Trigger: OnHeroDamaged, Kind: SecretReflectHeroDamage}},

	{ID: "valiant_ward", Name: "Valiant Ward", Type: TypeSecret, Class: ClassPaladin, Rarity: RarityCommon, Cost: 1,
		Text:   "Secret: When an enemy attacks, summon a 2/1 Defender as the new target.",
		Secret: &SecretDef{Trigger: OnEnemyAttack, Kind: SecretSummonDefender, Summon: "oath_defender"}},

	{ID: "second_dawn", Name: "Second Dawn", Type: TypeSecret, Class: ClassPaladin, Rarity: RarityCommon, Cost: 1,
		Text:   "Secret: When a friendly minion dies, return it to life with 1 Health.",
		Secret: &SecretDef{Trigger: OnFriendlyDeath, Kind: SecretResummonFriendly}},

	{ID: "penance_seal", Name: "Seal of Penance", Type: TypeSecret, Class: ClassPaladin, Rarity: RarityCommon, Cost: 1,
		Text:   "Secret: After your opponent plays a minion, reduce its Health to 1.",
		Secret: &SecretDef{Trigger: OnEnemyPlayMinion, Kind: SecretReduceHealth, Amount: 1}},

	// --- Minions ---

	{ID: "dawnguard_templar", Name: "Dawnguard Templar", Type: TypeMinion, Class: ClassPaladin, Rarity: RarityRare, Cost: 2, Attack: 3, Health: 2,
		Text:     "Onset: Give a friendly minion Aegis.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, Target: TargetFriendlyMinion, Grant: []Keyword{KeywordAegis}}}}},

	{ID: "riftwarden_pacifier", Name: "Riftwarden Pacifier", Type: TypeMinion, Class: ClassPaladin, Rarity: RarityRare, Cost: 3, Attack: 3, Health: 3, Tribe: TribeRiftborn,
		Text:     "Onset: Change an enemy minion's Attack to 1.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSetAttack, Amount: 1, Target: TargetEnemyMinion}}}},

	{ID: "crown_guardian", Name: "Crown Guardian", Type: TypeMinion, Class: ClassPaladin, Cost: 7, Attack: 5, Health: 7,
		Text:     "Taunt. Onset: Restore 6 Health to your hero.",
		Keywords: []Keyword{KeywordTaunt},
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectHeal, Amount: 6, Target: TargetNone, Area: AreaFriendlyHero}}}},

	{ID: "highlord_valdric", Name: "Highlord Valdric", Type: TypeMinion, Class: ClassPaladin, Rarity: RarityLegendary, Cost: 8, Attack: 6, Health: 6,
		Text:     "Aegis, Taunt. Final Gasp: Equip a 5/3 weapon.",
		Keywords: []Keyword{KeywordAegis, KeywordTaunt},
		Triggers: []Trigger{{When: OnDeath, Effect: Effect{Kind: EffectEquip, Target: TargetNone, EquipWeapon: "dawnbringer"}}}},

	// --- Weapons ---

	{ID: "dawnmace", Name: "Dawnmace", Type: TypeWeapon, Class: ClassPaladin, Cost: 1, Attack: 1, Durability: 4},

	{ID: "verdict_edge", Name: "Edge of Verdict", Type: TypeWeapon, Class: ClassPaladin, Rarity: RarityEpic, Cost: 3, Attack: 1, Durability: 5,
		Text:          "After you summon a minion, give it +1/+1 and this loses 1 Durability.",
		SummonBuffAtk: 1, SummonBuffHP: 1},

	{ID: "pureheart_blade", Name: "Pureheart Blade", Type: TypeWeapon, Class: ClassPaladin, Cost: 4, Attack: 4, Durability: 2,
		Text:           "Whenever your hero attacks, restore 3 Health to it.",
		WeaponHealHero: 3},

	// --- Hero power ---

	{ID: "muster", Name: "Muster", Type: TypeHeroPower, Class: ClassPaladin, Cost: 2,
		Text:   "Hero Power: Summon a 1/1 Recruit.",
		Effect: &Effect{Kind: EffectSummon, Target: TargetNone, Summon: "lightsworn_recruit"}},

	// --- Tokens (summon-only / equipped; excluded from decks) ---

	{ID: "lightsworn_recruit", Name: "Lightsworn Recruit", Type: TypeMinion, Class: ClassPaladin, Cost: 1, Attack: 1, Health: 1, Token: true},

	{ID: "oath_defender", Name: "Oath Defender", Type: TypeMinion, Class: ClassPaladin, Cost: 1, Attack: 2, Health: 1, Token: true},

	{ID: "dawnbringer", Name: "Dawnbringer", Type: TypeWeapon, Class: ClassPaladin, Cost: 0, Attack: 5, Durability: 3, Token: true},
}
