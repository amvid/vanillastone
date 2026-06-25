package cards

// hunterCards are the Hunter class cards (second playable class). Target scope is
// the HS Basic + Classic Hunter set; mechanics are 1:1 with genre staples while
// names + art are wholly original (see HANDOFF "Legal rules"). This first wave
// covers only the cards that need NO new engine support — Beast-tribe synergies,
// an onset buff, a final gasp, a random-summon spell, and the hero power.
// The mechanic-gated Hunter cards (traps/secrets, conditional damage, set-health,
// keyword auras, weapon triggers, multi-effect spells) land as their engine
// support is built — see HANDOFF "Open / next" Hunter phases.
//
// Basic cards carry NO rarity (empty Rarity = no gem); Classic cards do.
var hunterCards = []Card{
	// --- Spells ---

	{ID: "keen_arrow", Name: "Keen Arrow", Type: TypeSpell, Class: ClassHunter, Cost: 1,
		Text:   "Deal 2 damage.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetAny}},

	{ID: "call_the_pack", Name: "Call the Pack", Type: TypeSpell, Class: ClassHunter, Cost: 3,
		Text:   "Summon a random Beast companion.",
		Effect: &Effect{Kind: EffectSummonRandom, Target: TargetNone, GenIDs: []string{"frenzy_boar", "pride_lion", "guardian_bear"}}},

	{ID: "quarry_brand", Name: "Quarry Brand", Type: TypeSpell, Class: ClassHunter, Cost: 1,
		Text:   "Change a minion's Health to 1.",
		Effect: &Effect{Kind: EffectSetHealth, Amount: 1, Target: TargetMinion}},

	{ID: "feral_command", Name: "Feral Command", Type: TypeSpell, Class: ClassHunter, Cost: 3,
		Text:   "Deal 3 damage. If you control a Beast, deal 5 instead.",
		Effect: &Effect{Kind: EffectDamage, Amount: 3, ReqControlTribe: TribeBeast, AmountIfReq: 5, Target: TargetAny}},

	{ID: "volley_shot", Name: "Volley Shot", Type: TypeSpell, Class: ClassHunter, Cost: 4,
		Text:   "Deal 3 damage to two random enemy minions.",
		Effect: &Effect{Kind: EffectDamage, Amount: 3, Target: TargetNone, Area: AreaRandomEnemyMinion, Count: 2}},

	{ID: "culling_shot", Name: "Culling Shot", Type: TypeSpell, Class: ClassHunter, Rarity: RarityCommon, Cost: 3,
		Text:   "Destroy a random enemy minion.",
		Effect: &Effect{Kind: EffectDestroy, Target: TargetNone, Area: AreaRandomEnemyMinion}},

	{ID: "scout_ahead", Name: "Scout Ahead", Type: TypeSpell, Class: ClassHunter, Cost: 1,
		Text:   "Seek a card from your deck.",
		Effect: &Effect{Kind: EffectSeek, Target: TargetNone, FromDeck: true}},

	{ID: "signal_flare", Name: "Signal Flare", Type: TypeSpell, Class: ClassHunter, Rarity: RarityRare, Cost: 1,
		Text:   "All minions lose Stealth. Destroy all enemy Secrets. Draw a card.",
		Effect: &Effect{Kind: EffectFlare, Target: TargetNone}},

	{ID: "unleash_the_pack", Name: "Unleash the Pack", Type: TypeSpell, Class: ClassHunter, Rarity: RarityRare, Cost: 3,
		Text:   "For each enemy minion, summon a 1/1 Hound with Charge.",
		Effect: &Effect{Kind: EffectSummon, Target: TargetNone, Summon: "snarling_hound", CountPerEnemyMinion: true}},

	{ID: "bestial_fury", Name: "Bestial Fury", Type: TypeSpell, Class: ClassHunter, Rarity: RarityEpic, Cost: 1,
		Text:   "Give a friendly Beast +2 Attack and Immune this turn.",
		Effect: &Effect{Kind: EffectBuff, Target: TargetFriendlyMinion, ReqTribe: TribeBeast, BuffAtk: 2, Grant: []Keyword{KeywordImmune}, Temporary: true}},

	{ID: "blasting_shot", Name: "Blasting Shot", Type: TypeSpell, Class: ClassHunter, Rarity: RarityRare, Cost: 5,
		Text:   "Deal 5 damage to a minion and 2 damage to adjacent ones.",
		Effect: &Effect{Kind: EffectDamage, Amount: 5, SplashAmount: 2, Target: TargetMinion}},

	// --- Minions ---

	{ID: "packleader_wolf", Name: "Packleader Wolf", Type: TypeMinion, Class: ClassHunter, Cost: 1, Attack: 1, Health: 1,
		Tribe: TribeBeast,
		Text:  "Your other Beasts have +1 Attack.",
		Aura:  &Aura{Atk: 1, Tribe: TribeBeast}},

	{ID: "famished_vulture", Name: "Famished Vulture", Type: TypeMinion, Class: ClassHunter, Cost: 2, Attack: 2, Health: 1,
		Tribe:    TribeBeast,
		Text:     "Whenever you summon a Beast, draw a card.",
		Triggers: []Trigger{{When: OnFriendlySummon, SubjectTribe: TribeBeast, Effect: Effect{Kind: EffectDraw, Amount: 1}}}},

	{ID: "carrion_hyena", Name: "Carrion Hyena", Type: TypeMinion, Class: ClassHunter, Rarity: RarityCommon, Cost: 2, Attack: 2, Health: 2,
		Tribe:    TribeBeast,
		Text:     "Whenever a friendly Beast dies, gain +2/+1.",
		Triggers: []Trigger{{When: OnFriendlyDeath, SubjectTribe: TribeBeast, Effect: Effect{Kind: EffectBuff, Target: TargetSelf, BuffAtk: 2, BuffHP: 1}}}},

	{ID: "kennel_master", Name: "Kennel Master", Type: TypeMinion, Class: ClassHunter, Cost: 4, Attack: 4, Health: 3,
		Text:     "Onset: Give a friendly Beast +2/+2 and Taunt.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, Target: TargetFriendlyMinion, ReqTribe: TribeBeast, BuffAtk: 2, BuffHP: 2, Grant: []Keyword{KeywordTaunt}}}}},

	{ID: "mane_lioness", Name: "Mane Lioness", Type: TypeMinion, Class: ClassHunter, Rarity: RarityRare, Cost: 6, Attack: 6, Health: 5,
		Tribe:    TribeBeast,
		Text:     "Final Gasp: Summon two 2/2 Cubs.",
		Triggers: []Trigger{{When: OnDeath, Effect: Effect{Kind: EffectSummon, Summon: "savanna_cub", Count: 2}}}},

	{ID: "apex_saurian", Name: "Apex Saurian", Type: TypeMinion, Class: ClassHunter, Rarity: RarityLegendary, Cost: 9, Attack: 8, Health: 8,
		Tribe:    TribeBeast,
		Text:     "Charge",
		Keywords: []Keyword{KeywordCharge}},

	{ID: "tundra_charger", Name: "Tundra Charger", Type: TypeMinion, Class: ClassHunter, Cost: 5, Attack: 2, Health: 5,
		Tribe: TribeBeast,
		Text:  "Your Beasts have Charge.",
		Aura:  &Aura{Tribe: TribeBeast, Grant: []Keyword{KeywordCharge}}},

	// --- Secrets (traps; fire on the opponent's action) ---

	{ID: "blasting_snare", Name: "Blasting Snare", Type: TypeSecret, Class: ClassHunter, Rarity: RarityCommon, Cost: 2,
		Text:   "Secret: When your hero is attacked, deal 2 damage to all enemies.",
		Secret: &SecretDef{Trigger: OnHeroAttacked, Kind: SecretDamageAll, Amount: 2}},

	{ID: "snaring_trap", Name: "Snaring Trap", Type: TypeSecret, Class: ClassHunter, Rarity: RarityCommon, Cost: 2,
		Text:   "Secret: When an enemy minion attacks, return it to its owner's hand. It costs (2) more.",
		Secret: &SecretDef{Trigger: OnEnemyAttack, Kind: SecretBounceAttacker, Amount: 2}},

	{ID: "marksman_trap", Name: "Marksman's Trap", Type: TypeSecret, Class: ClassHunter, Rarity: RarityCommon, Cost: 2,
		Text:   "Secret: After your opponent plays a minion, deal 4 damage to it.",
		Secret: &SecretDef{Trigger: OnEnemyPlayMinion, Kind: SecretDamageMinion, Amount: 4}},

	{ID: "feint_trap", Name: "Feint Trap", Type: TypeSecret, Class: ClassHunter, Rarity: RarityRare, Cost: 2,
		Text:   "Secret: When an enemy attacks your hero, instead it attacks another random character.",
		Secret: &SecretDef{Trigger: OnEnemyAttackHero, Kind: SecretRetargetAttack}},

	{ID: "serpent_trap", Name: "Serpent Trap", Type: TypeSecret, Class: ClassHunter, Rarity: RarityEpic, Cost: 2,
		Text:   "Secret: When one of your minions is attacked, summon three 1/1 Serpents.",
		Secret: &SecretDef{Trigger: OnMinionAttacked, Kind: SecretSummon, Amount: 3, Summon: "coil_serpent"}},

	// --- Weapons ---

	{ID: "hawkeye_bow", Name: "Hawkeye Bow", Type: TypeWeapon, Class: ClassHunter, Rarity: RarityCommon, Cost: 3, Attack: 3, Durability: 2,
		Text:             "Whenever a friendly Secret is revealed, gain +1 Durability.",
		WeaponSecretGain: true},

	{ID: "duelists_longbow", Name: "Duelist's Longbow", Type: TypeWeapon, Class: ClassHunter, Rarity: RarityEpic, Cost: 7, Attack: 5, Durability: 2,
		Text:            "Your hero is Immune while attacking.",
		ImmuneAttacking: true},

	// --- Hero power ---

	{ID: "quick_shot", Name: "Quick Shot", Type: TypeHeroPower, Class: ClassHunter, Cost: 2,
		Text:   "Hero Power: Deal 2 damage to the enemy hero.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaEnemyHero}},

	// --- Tokens (summon-only; excluded from decks) ---

	{ID: "frenzy_boar", Name: "Frenzy Boar", Type: TypeMinion, Class: ClassHunter, Cost: 3, Attack: 4, Health: 2,
		Tribe: TribeBeast, Token: true, Text: "Charge", Keywords: []Keyword{KeywordCharge}},

	{ID: "pride_lion", Name: "Pride Lion", Type: TypeMinion, Class: ClassHunter, Cost: 3, Attack: 2, Health: 4,
		Tribe: TribeBeast, Token: true, Text: "Your other minions have +1 Attack.", Aura: &Aura{Atk: 1}},

	{ID: "guardian_bear", Name: "Guardian Bear", Type: TypeMinion, Class: ClassHunter, Cost: 3, Attack: 4, Health: 4,
		Tribe: TribeBeast, Token: true, Text: "Taunt", Keywords: []Keyword{KeywordTaunt}},

	{ID: "savanna_cub", Name: "Savanna Cub", Type: TypeMinion, Class: ClassHunter, Cost: 2, Attack: 2, Health: 2,
		Tribe: TribeBeast, Token: true},

	{ID: "snarling_hound", Name: "Snarling Hound", Type: TypeMinion, Class: ClassHunter, Cost: 1, Attack: 1, Health: 1,
		Tribe: TribeBeast, Token: true, Text: "Charge", Keywords: []Keyword{KeywordCharge}},

	{ID: "coil_serpent", Name: "Coil Serpent", Type: TypeMinion, Class: ClassHunter, Cost: 1, Attack: 1, Health: 1,
		Tribe: TribeBeast, Token: true},
}
