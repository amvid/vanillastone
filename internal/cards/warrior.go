package cards

// warriorCards are the Warrior class cards (third playable class). Target scope is
// the genre-staple Basic + Classic Warrior set; mechanics are 1:1 with well-worn
// staples while names + art + rules text are wholly original (see HANDOFF "Legal
// rules"). The class theme is Armor, weapons, and damage/Enrage payoffs.
//
// Basic cards carry NO rarity (empty Rarity = no gem); Classic cards do.
var warriorCards = []Card{
	// --- Spells ---

	{ID: "goading_strike", Name: "Goading Strike", Type: TypeSpell, Class: ClassWarrior, Rarity: RarityCommon, Cost: 0,
		Text:   "Deal 1 damage to a minion and give it +2 Attack.",
		Effect: &Effect{Kind: EffectDamage, Amount: 1, BuffAtk: 2, Target: TargetMinion}},

	{ID: "hammer_blow", Name: "Hammer Blow", Type: TypeSpell, Class: ClassWarrior, Rarity: RarityCommon, Cost: 1,
		Text:   "Deal 2 damage to a minion. If it survives, draw a card.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, DrawIfSurvives: true, Target: TargetMinion}},

	{ID: "hone_edge", Name: "Hone Edge", Type: TypeSpell, Class: ClassWarrior, Rarity: RarityCommon, Cost: 1,
		Text:   "If you have a weapon, give it +1/+1. Otherwise equip a 1/3 weapon.",
		Effect: &Effect{Kind: EffectEquip, Target: TargetNone, EquipWeapon: "whetstone_blade", UpgradeIfWeapon: true, BuffAtk: 1, BuffHP: 1}},

	{ID: "bulwark_bash", Name: "Bulwark Bash", Type: TypeSpell, Class: ClassWarrior, Rarity: RarityEpic, Cost: 1,
		Text:   "Deal damage to a minion equal to your Armor.",
		Effect: &Effect{Kind: EffectDamage, ScaleByArmor: true, Target: TargetMinion}},

	{ID: "finishing_cut", Name: "Finishing Cut", Type: TypeSpell, Class: ClassWarrior, Cost: 1,
		Text:   "Destroy a damaged enemy minion.",
		Effect: &Effect{Kind: EffectDestroy, Target: TargetEnemyMinion, ReqDamaged: true}},

	{ID: "steel_cyclone", Name: "Steel Cyclone", Type: TypeSpell, Class: ClassWarrior, Cost: 1,
		Text:   "Deal 1 damage to ALL minions.",
		Effect: &Effect{Kind: EffectDamage, Amount: 1, Target: TargetNone, Area: AreaAllMinions}},

	{ID: "war_frenzy", Name: "War Frenzy", Type: TypeSpell, Class: ClassWarrior, Rarity: RarityCommon, Cost: 2,
		Text:   "Draw a card for each damaged friendly character.",
		Effect: &Effect{Kind: EffectDrawPerDamaged, Target: TargetNone}},

	{ID: "berserk_surge", Name: "Berserk Surge", Type: TypeSpell, Class: ClassWarrior, Cost: 2,
		Text:   "Give a damaged minion +3/+3.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 3, BuffHP: 3, Target: TargetMinion, ReqDamaged: true}},

	{ID: "rallying_roar", Name: "Rallying Roar", Type: TypeSpell, Class: ClassWarrior, Rarity: RarityRare, Cost: 2,
		Text:   "Your minions can't be reduced below 1 Health this turn. Draw a card.",
		Effect: &Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone, GuardMinions: true}},

	{ID: "wide_swing", Name: "Wide Swing", Type: TypeSpell, Class: ClassWarrior, Cost: 2,
		Text:   "Deal 2 damage to two random enemy minions.",
		Effect: &Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaRandomEnemyMinion, Count: 2}},

	{ID: "valiant_strike", Name: "Valiant Strike", Type: TypeSpell, Class: ClassWarrior, Cost: 2,
		Text:   "Give your hero +4 Attack this turn.",
		Effect: &Effect{Kind: EffectHeroAttack, Amount: 4, Target: TargetNone}},

	{ID: "bracing_guard", Name: "Bracing Guard", Type: TypeSpell, Class: ClassWarrior, Rarity: RarityCommon, Cost: 2,
		Text:   "Gain 5 Armor. Draw a card.",
		Effect: &Effect{Kind: EffectArmor, Amount: 5, ThenDraw: 1, Target: TargetNone}},

	{ID: "headlong_rush", Name: "Headlong Rush", Type: TypeSpell, Class: ClassWarrior, Cost: 3,
		Text:   "Give a friendly minion +2 Attack and Charge.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 2, Target: TargetFriendlyMinion, Grant: []Keyword{KeywordCharge}}},

	{ID: "deathblow_swing", Name: "Deathblow Swing", Type: TypeSpell, Class: ClassWarrior, Rarity: RarityRare, Cost: 4,
		Text:   "Deal 4 damage. If you have 12 or less Health, deal 6 instead.",
		Effect: &Effect{Kind: EffectDamage, Amount: 4, ReqOwnHealthAtMost: 12, AmountIfReq: 6, Target: TargetAny}},

	{ID: "pit_brawl", Name: "Pit Brawl", Type: TypeSpell, Class: ClassWarrior, Rarity: RarityEpic, Cost: 5,
		Text:   "Destroy all minions except one (chosen at random).",
		Effect: &Effect{Kind: EffectBrawl, Target: TargetNone}},

	// --- Minions ---

	{ID: "whipcrack_overseer", Name: "Whipcrack Overseer", Type: TypeMinion, Class: ClassWarrior, Rarity: RarityCommon, Cost: 2, Attack: 2, Health: 3,
		Text:     "Onset: Deal 1 damage to a minion and give it +2 Attack.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 1, BuffAtk: 2, Target: TargetMinion}}}},

	{ID: "platewright", Name: "Platewright", Type: TypeMinion, Class: ClassWarrior, Rarity: RarityRare, Cost: 2, Attack: 1, Health: 4,
		Text:     "Whenever a friendly minion takes damage, gain 1 Armor.",
		Triggers: []Trigger{{When: OnFriendlyMinionDamage, Effect: Effect{Kind: EffectArmor, Amount: 1, Target: TargetNone}}}},

	{ID: "battle_marshal", Name: "Battle Marshal", Type: TypeMinion, Class: ClassWarrior, Cost: 3, Attack: 2, Health: 3,
		Text:     "Whenever you summon a minion with 3 or less Attack, give it Charge.",
		Triggers: []Trigger{{When: OnFriendlySummon, SubjectMaxAttack: 3, Effect: Effect{Kind: EffectBuff, Target: TargetSubject, Grant: []Keyword{KeywordCharge}}}}},

	{ID: "ragebound_brute", Name: "Ragebound Brute", Type: TypeMinion, Class: ClassWarrior, Rarity: RarityRare, Cost: 3, Attack: 2, Health: 4,
		Text:     "Whenever a minion takes damage, gain +1 Attack.",
		Triggers: []Trigger{{When: OnAnyMinionDamage, Effect: Effect{Kind: EffectBuff, Target: TargetSelf, BuffAtk: 1}}}},

	{ID: "forgehold_smith", Name: "Forgehold Smith", Type: TypeMinion, Class: ClassWarrior, Rarity: RarityCommon, Cost: 4, Attack: 3, Health: 3,
		Text:     "Onset: Equip a 2/2 weapon.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectEquip, Target: TargetNone, EquipWeapon: "keenedge_blade"}}}},

	{ID: "ironguard_elite", Name: "Ironguard Elite", Type: TypeMinion, Class: ClassWarrior, Cost: 4, Attack: 4, Health: 3,
		Text:     "Charge.",
		Keywords: []Keyword{KeywordCharge}},

	{ID: "warchief_gorthak", Name: "Warchief Gorthak", Type: TypeMinion, Class: ClassWarrior, Rarity: RarityLegendary, Cost: 8, Attack: 4, Health: 9,
		Text:     "Charge. Has +6 Attack while damaged.",
		Keywords: []Keyword{KeywordCharge},
		Enrage:   &Aura{Atk: 6}},

	// --- Weapons ---

	{ID: "cindersplit_axe", Name: "Cindersplit Axe", Type: TypeWeapon, Class: ClassWarrior, Cost: 2, Attack: 3, Durability: 2},

	{ID: "runesteel_reaper", Name: "Runesteel Reaper", Type: TypeWeapon, Class: ClassWarrior, Cost: 5, Attack: 5, Durability: 2},

	{ID: "bloodwail", Name: "Bloodwail", Type: TypeWeapon, Class: ClassWarrior, Rarity: RarityEpic, Cost: 7, Attack: 7, Durability: 1,
		Text:         "Attacking a minion costs 1 Attack instead of 1 Durability.",
		WearByAttack: true},

	// --- Hero power ---

	{ID: "shore_up", Name: "Shore Up", Type: TypeHeroPower, Class: ClassWarrior, Cost: 2,
		Text:   "Hero Power: Gain 2 Armor.",
		Effect: &Effect{Kind: EffectArmor, Amount: 2, Target: TargetNone}},

	// --- Tokens (weapons equipped by cards above; not collectible) ---

	{ID: "keenedge_blade", Name: "Keenedge Blade", Type: TypeWeapon, Class: ClassWarrior, Cost: 2, Attack: 2, Durability: 2, Token: true},

	{ID: "whetstone_blade", Name: "Whetstone Blade", Type: TypeWeapon, Class: ClassWarrior, Cost: 1, Attack: 1, Durability: 3, Token: true},
}
