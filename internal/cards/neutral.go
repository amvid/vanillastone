package cards

// neutralCards are the Neutral cards, scoped to the HS Classic set (target: 105;
// see TASKS.md). Each is a 1:1 mechanic clone of a Classic neutral with a wholly
// original name + art (mechanics aren't copyrightable; names/art are — see HANDOFF
// "Legal rules"). Cards needing mechanics we have not built yet (enrage, adjacency,
// cost modification, tribes, bounce, weapon-manipulation, temp buffs, special
// legendaries, …) are omitted until those mechanics land — see TASKS.md backlog.
//
// NOTE: some engine features currently have NO collectible card because no Classic
// neutral uses them at our build stage (weapons, aura, destroy, transform, the
// enemy/friendly-hero target rules). The engine keeps them; cards arrive with the
// relevant wave.
var neutralCards = []Card{
	// --- Vanilla & single-keyword minions ---

	{ID: "mote", Name: "Mote", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 0, Attack: 1, Health: 1},

	{ID: "silver_page", Name: "Silver Page", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 1, Attack: 1, Health: 1,
		Text: "Aegis.", Keywords: []Keyword{KeywordAegis}},

	{ID: "shadow_prowler", Name: "Shadow Prowler", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 1, Attack: 2, Health: 1,
		Text: "Stealth.", Keywords: []Keyword{KeywordStealth}},

	{ID: "fledgling_hawk", Name: "Fledgling Hawk", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 1, Attack: 1, Health: 1,
		Text: "Twinstrike.", Tribe: TribeBeast, Keywords: []Keyword{KeywordTwinstrike}},

	{ID: "shield_lackey", Name: "Shield Lackey", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 1, Attack: 0, Health: 4,
		Text: "Taunt.", Tribe: TribeRiftborn, Keywords: []Keyword{KeywordTaunt}},

	{ID: "gale_seer", Name: "Gale Seer", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 3, Attack: 2, Health: 3,
		Text: "Twinstrike.", Keywords: []Keyword{KeywordTwinstrike}},

	{ID: "thornvale_panther", Name: "Thornvale Panther", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 3, Attack: 4, Health: 2,
		Text: "Stealth.", Tribe: TribeBeast, Keywords: []Keyword{KeywordStealth}},

	{ID: "dawn_crusader", Name: "Dawn Crusader", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 3, Attack: 3, Health: 1,
		Text: "Aegis.", Keywords: []Keyword{KeywordAegis}},

	{ID: "venom_serpent", Name: "Venom Serpent", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 3, Attack: 2, Health: 3,
		Text: "Poisonous.", Tribe: TribeBeast, Keywords: []Keyword{KeywordPoisonous}},

	{ID: "glimmerwing_drake", Name: "Glimmerwing Drake", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 2, Attack: 3, Health: 2,
		Text: "Can't be targeted by spells or Hero Powers.", Tribe: TribeDragon, Keywords: []Keyword{KeywordElusive}},

	{ID: "stoneveil_watcher", Name: "Stoneveil Watcher", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 2, Attack: 4, Health: 5,
		Text: "Can't attack.", Keywords: []Keyword{KeywordCantAttack}},

	{ID: "granite_warden", Name: "Granite Warden", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 4, Attack: 1, Health: 7,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},

	{ID: "moonsilver_guardian", Name: "Moonsilver Guardian", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 4, Attack: 3, Health: 3,
		Text: "Aegis.", Keywords: []Keyword{KeywordAegis}},

	{ID: "marsh_lurker", Name: "Marsh Lurker", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 5, Attack: 3, Health: 6,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},

	{ID: "jungle_stalker", Name: "Jungle Stalker", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 5, Attack: 5, Health: 5,
		Text: "Stealth.", Tribe: TribeBeast, Keywords: []Keyword{KeywordStealth}},

	{ID: "vanguard_champion", Name: "Vanguard Champion", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 6, Attack: 4, Health: 2,
		Text: "Charge. Aegis.", Keywords: []Keyword{KeywordCharge, KeywordAegis}},

	{ID: "dawnguard_protector", Name: "Dawnguard Protector", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 6, Attack: 4, Health: 5,
		Text: "Taunt. Aegis.", Keywords: []Keyword{KeywordTaunt, KeywordAegis}},

	{ID: "galewing_harpy", Name: "Galewing Harpy", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 6, Attack: 4, Health: 5,
		Text: "Twinstrike.", Keywords: []Keyword{KeywordTwinstrike}},

	{ID: "veiled_assassin", Name: "Veiled Assassin", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 7, Attack: 7, Health: 5,
		Text: "Stealth.", Keywords: []Keyword{KeywordStealth}},

	{ID: "spelltide_wyrm", Name: "Spelltide Wyrm", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 9, Attack: 4, Health: 12,
		Text: "Spell Damage +5.", Tribe: TribeDragon, SpellDamage: 5},

	// --- Enrage minions (Atk bonus active only while damaged) ---

	{ID: "riled_rooster", Name: "Riled Rooster", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 1, Attack: 1, Health: 1,
		Text: "Enrage: +5 Attack.", Enrage: &Aura{Atk: 5}},

	{ID: "frenzied_brave", Name: "Frenzied Brave", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 2, Attack: 2, Health: 3,
		Text: "Enrage: +3 Attack.", Enrage: &Aura{Atk: 3}},

	{ID: "highland_guardian", Name: "Highland Guardian", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 3, Attack: 2, Health: 3,
		Text: "Taunt. Enrage: +3 Attack.", Keywords: []Keyword{KeywordTaunt}, Enrage: &Aura{Atk: 3}},

	// --- Onset / FinalGasp minions ---

	{ID: "plague_gremlin", Name: "Plague Gremlin", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 1, Attack: 2, Health: 1,
		Text:     "Final Gasp: Deal 2 damage to the enemy hero.",
		Triggers: []Trigger{{When: OnDeath, Effect: Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaEnemyHero}}}},

	{ID: "pilfer_imp", Name: "Pilfer Imp", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 2, Attack: 2, Health: 1,
		Text:     "Final Gasp: Draw a card.",
		Triggers: []Trigger{{When: OnDeath, Effect: Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone}}}},

	{ID: "vael_emberscribe", Name: "Vael Emberscribe", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 2, Attack: 1, Health: 1, Tribe: TribeUndead,
		Text:        "Spell Damage +1. Final Gasp: Draw a card.",
		SpellDamage: 1,
		Triggers:    []Trigger{{When: OnDeath, Effect: Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone}}}},

	{ID: "hushwing_owl", Name: "Hushwing Owl", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 3, Attack: 2, Health: 1,
		Text:     "Onset: Silence a minion.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSilence, Target: TargetMinion}}}},

	{ID: "earthroot_healer", Name: "Earthroot Healer", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 3, Attack: 3, Health: 3,
		Text:     "Onset: Restore 3 Health.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectHeal, Amount: 3, Target: TargetAny}}}},

	{ID: "goad_imp", Name: "Goad Imp", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 1, Attack: 1, Health: 1,
		Text:     "Onset: Give a minion +2 Attack this turn.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, BuffAtk: 2, Target: TargetMinion, Temporary: true}}}},

	{ID: "ironforge_brute", Name: "Ironforge Brute", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 4, Attack: 4, Health: 4,
		Text:     "Onset: Give a minion +2 Attack this turn.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, BuffAtk: 2, Target: TargetMinion, Temporary: true}}}},

	{ID: "tavern_apprentice", Name: "Tavern Apprentice", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 2, Attack: 3, Health: 2,
		Text:     "Onset: Return a friendly minion to your hand.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBounce, Target: TargetFriendlyMinion}}}},

	{ID: "elder_brewkeeper", Name: "Elder Brewkeeper", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 4, Attack: 5, Health: 4,
		Text:     "Onset: Return a friendly minion to your hand.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBounce, Target: TargetFriendlyMinion}}}},

	{ID: "wardstone_sentinel", Name: "Wardstone Sentinel", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 2, Attack: 2, Health: 3,
		Text:     "Onset: Give adjacent minions Taunt.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, Target: TargetNone, Area: AreaAdjacent, Grant: []Keyword{KeywordTaunt}}}}},

	{ID: "bannerguard", Name: "Bannerguard", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 4, Attack: 2, Health: 3,
		Text:     "Onset: Give adjacent minions +1/+1 and Taunt.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, BuffAtk: 1, BuffHP: 1, Target: TargetNone, Area: AreaAdjacent, Grant: []Keyword{KeywordTaunt}}}}},

	{ID: "reaper_golem", Name: "Reaper Golem", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 3, Attack: 2, Health: 3, Tribe: TribeMech,
		Text:     "Final Gasp: Summon a 2/1 Broken Golem.",
		Triggers: []Trigger{{When: OnDeath, Effect: Effect{Kind: EffectSummon, Summon: "broken_golem", Count: 1, Target: TargetNone}}}},

	{ID: "errant_knight", Name: "Errant Knight", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 5, Attack: 4, Health: 4,
		Text:     "Onset: Summon a 2/2 Errant Squire.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSummon, Summon: "errant_squire", Count: 1, Target: TargetNone}}}},

	{ID: "emberwing_matron", Name: "Emberwing Matron", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 9, Attack: 8, Health: 8, Tribe: TribeDragon,
		Text: "Onset: Summon 1/1 Whelps until your board is full.",
		// Count 6 = the most empty slots possible after Matron occupies one; the
		// summon path discards once the board is full, so this fills exactly.
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSummon, Summon: "emberwing_whelp", Count: 6, Target: TargetNone}}}},

	{ID: "rime_elemental", Name: "Rime Elemental", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 6, Attack: 5, Health: 5, Tribe: TribeElemental,
		Text:     "Onset: Freeze a character.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 0, Target: TargetAny, Freeze: true}}}},

	{ID: "tavern_medic", Name: "Tavern Medic", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 6, Attack: 5, Health: 4,
		Text:     "Onset: Restore 4 Health to your hero.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectHeal, Amount: 4, Target: TargetFriendlyHero}}}},

	{ID: "hornelder_chief", Name: "Hornelder Chief", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 6, Attack: 5, Health: 5,
		Text:     "Final Gasp: Summon a 5/5 Hornelder Heir.",
		Triggers: []Trigger{{When: OnDeath, Effect: Effect{Kind: EffectSummon, Summon: "hornelder_heir", Count: 1, Target: TargetNone}}}},

	// --- Edge-trigger minions (react to ongoing game events while in play) ---

	{ID: "dagger_tosser", Name: "Dagger Tosser", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 2, Attack: 3, Health: 2,
		Text:     "After you summon a minion, deal 1 damage to a random enemy.",
		Triggers: []Trigger{{When: OnFriendlySummon, Effect: Effect{Kind: EffectDamage, Amount: 1, Target: TargetRandomEnemy}}}},

	{ID: "carrion_fiend", Name: "Carrion Fiend", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 3, Attack: 3, Health: 3, Tribe: TribeUndead,
		Text:     "Whenever a minion dies, gain +1 Attack.",
		Triggers: []Trigger{{When: OnAnyMinionDeath, Effect: Effect{Kind: EffectBuff, BuffAtk: 1, Target: TargetSelf}}}},

	{ID: "siege_engine", Name: "Siege Engine", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 3, Attack: 1, Health: 4, Tribe: TribeMech,
		Text:     "At the start of your turn, deal 2 damage to a random enemy.",
		Triggers: []Trigger{{When: OnTurnStart, Effect: Effect{Kind: EffectDamage, Amount: 2, Target: TargetRandomEnemy}}}},

	{ID: "cabal_overseer", Name: "Cabal Overseer", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 4, Attack: 4, Health: 2, Tribe: TribeUndead,
		Text:     "Whenever one of your other minions dies, draw a card.",
		Triggers: []Trigger{{When: OnFriendlyDeath, Effect: Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone}}}},

	{ID: "adept_tutor", Name: "Adept Tutor", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 4, Attack: 3, Health: 5,
		Text:     "Whenever you cast a spell, summon a 1/1 Pupil.",
		Triggers: []Trigger{{When: OnSpellCast, Effect: Effect{Kind: EffectSummon, Summon: "tutors_pupil", Count: 1, Target: TargetNone}}}},

	{ID: "ashflame_zealot", Name: "Ashflame Zealot", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 2, Attack: 3, Health: 2,
		Text:     "After you cast a spell, deal 1 damage to all minions.",
		Triggers: []Trigger{{When: OnSpellCast, Effect: Effect{Kind: EffectDamage, Amount: 1, Target: TargetNone, Area: AreaAllMinions}}}},

	{ID: "bazaar_crier", Name: "Bazaar Crier", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 6, Attack: 4, Health: 4,
		Text:     "Whenever you cast a spell, draw a card.",
		Triggers: []Trigger{{When: OnSpellCast, Effect: Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone}}}},

	{ID: "dawn_tender", Name: "Dawn Tender", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 1, Attack: 1, Health: 2,
		Text:     "Whenever a character is healed, gain +2 Attack.",
		Triggers: []Trigger{{When: OnHeal, Effect: Effect{Kind: EffectBuff, BuffAtk: 2, Target: TargetSelf}}}},

	{ID: "rune_warden", Name: "Rune Warden", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 1, Attack: 1, Health: 2,
		Text:     "Whenever a Secret is played, gain +1/+1.",
		Triggers: []Trigger{{When: OnSecretPlayed, Effect: Effect{Kind: EffectBuff, BuffAtk: 1, BuffHP: 1, Target: TargetSelf}}}},

	{ID: "relic_seeker", Name: "Relic Seeker", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 3, Attack: 2, Health: 2,
		Text:     "Whenever you play a card, gain +1/+1.",
		Triggers: []Trigger{{When: OnPlayCard, Effect: Effect{Kind: EffectBuff, BuffAtk: 1, BuffHP: 1, Target: TargetSelf}}}},

	{ID: "acolyte_novice", Name: "Acolyte Novice", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 1, Attack: 2, Health: 1,
		Text:     "At the end of your turn, give another random friendly minion +1 Health.",
		Triggers: []Trigger{{When: OnTurnEnd, Effect: Effect{Kind: EffectBuff, BuffHP: 1, Target: TargetRandomFriendly}}}},

	{ID: "forge_hand", Name: "Forge Hand", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 2, Attack: 1, Health: 3,
		Text:     "At the end of your turn, give another random friendly minion +1 Attack.",
		Triggers: []Trigger{{When: OnTurnEnd, Effect: Effect{Kind: EffectBuff, BuffAtk: 1, Target: TargetRandomFriendly}}}},

	{ID: "spellrage_acolyte", Name: "Spellrage Acolyte", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 2, Attack: 1, Health: 3,
		Text:     "Whenever you cast a spell, gain +2 Attack this turn.",
		Triggers: []Trigger{{When: OnSpellCast, Effect: Effect{Kind: EffectBuff, BuffAtk: 2, Target: TargetSelf, Temporary: true}}}},

	{ID: "managlutton", Name: "Managlutton", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 8, Attack: 4, Health: 8, Tribe: TribeElemental,
		Text:     "Whenever you cast a spell, gain +2/+2.",
		Triggers: []Trigger{{When: OnSpellCast, Effect: Effect{Kind: EffectBuff, BuffAtk: 2, BuffHP: 2, Target: TargetSelf}}}},

	{ID: "ruin_oracle", Name: "Ruin Oracle", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 2, Attack: 0, Health: 7,
		Text:     "At the start of your turn, destroy all minions.",
		Triggers: []Trigger{{When: OnTurnStart, Effect: Effect{Kind: EffectDestroy, Target: TargetNone, Area: AreaAllMinions}}}},

	{ID: "snarlmaw", Name: "Snarlmaw", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 6, Attack: 4, Health: 4,
		Text:     "At the end of your turn, summon a 2/2 Snarl Pup with Taunt.",
		Triggers: []Trigger{{When: OnTurnEnd, Effect: Effect{Kind: EffectSummon, Summon: "snarl_pup", Count: 1, Target: TargetNone}}}},

	// --- Special / conditional battlecries (sub-wave 1: untargeted/random) ---

	{ID: "wounded_duelist", Name: "Wounded Duelist", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 3, Attack: 4, Health: 7,
		Text:     "Onset: Deal 4 damage to himself.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 4, Target: TargetSelf}}}},

	{ID: "powder_tosser", Name: "Powder Tosser", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 2, Attack: 3, Health: 2,
		Text:     "Onset: Deal 3 damage randomly split among all other characters.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectMissiles, Amount: 1, Count: 3, Target: TargetNone}}}},

	{ID: "trampling_brute", Name: "Trampling Brute", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 5, Attack: 3, Health: 5, Tribe: TribeBeast,
		Text:     "Onset: Destroy a random enemy minion with 2 or less Attack.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDestroy, Target: TargetNone, Area: AreaRandomEnemyMinion, MaxAttack: 2}}}},

	{ID: "duskscale_drake", Name: "Duskscale Drake", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 4, Attack: 4, Health: 1, Tribe: TribeDragon,
		Text:     "Onset: Gain +1 Health for each card in your hand.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, BuffHP: 1, Target: TargetSelf, PerCardInHand: true}}}},

	{ID: "covert_saboteur", Name: "Covert Saboteur", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 4, Attack: 5, Health: 4,
		Text:     "Onset: Destroy a random enemy Secret.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectKillSecret, Target: TargetNone}}}},

	// --- Special / conditional battlecries (sub-wave 2: target-condition + copy) ---

	{ID: "trophy_hunter", Name: "Trophy Hunter", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 5, Attack: 4, Health: 2,
		Text:     "Onset: Destroy a minion with 7 or more Attack.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDestroy, Target: TargetMinion, ReqAttack: 7}}}},

	{ID: "grave_knight", Name: "Grave Knight", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 6, Attack: 4, Health: 5,
		Text:     "Onset: Destroy an enemy minion with Taunt.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDestroy, Target: TargetEnemyMinion, ReqTaunt: true}}}},

	{ID: "visage_thief", Name: "Visage Thief", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 5, Attack: 3, Health: 3,
		Text:     "Onset: Become a copy of a chosen minion.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectCopy, Target: TargetMinion}}}},

	// --- Special / conditional battlecries (sub-wave 3: swap-stats + consume-shields) ---

	{ID: "addled_brewer", Name: "Addled Brewer", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 2, Attack: 2, Health: 2,
		Text:     "Onset: Swap a minion's Attack and Health.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSwapStats, Target: TargetMinion}}}},

	{ID: "crimson_reaver", Name: "Crimson Reaver", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 3, Attack: 3, Health: 3,
		Text:     "Onset: All minions lose Aegis. Gain +3/+3 for each Shield lost.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectConsumeShields, BuffAtk: 3, BuffHP: 3, Target: TargetNone}}}},

	// --- Weapon-manipulation battlecries ---

	{ID: "brine_cutter", Name: "Brine Cutter", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 1, Attack: 1, Health: 2, Tribe: TribePirate,
		Text:     "Onset: Remove 1 Durability from your opponent's weapon.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectChipWeapon, Amount: 1, Target: TargetNone}}}},

	{ID: "tidereaver", Name: "Tidereaver", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 2, Attack: 2, Health: 3, Tribe: TribePirate,
		Text:     "Onset: Gain Attack equal to your weapon's Attack.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectGainWeaponAttack, Target: TargetNone}}}},

	{ID: "captain_brackwater", Name: "Captain Brackwater", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 5, Attack: 5, Health: 4, Tribe: TribePirate,
		Text:     "Onset: Give your weapon +1/+1.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuffWeapon, BuffAtk: 1, BuffHP: 1, Target: TargetNone}}}},

	{ID: "relic_breaker", Name: "Relic Breaker", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 5, Attack: 5, Health: 4,
		Text:     "Onset: Destroy your opponent's weapon and draw cards equal to its Durability.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDestroyWeapon, Target: TargetNone, DrawWeaponDurability: true}}}},

	// --- Tribe auras + tribe-synergy minions ---

	{ID: "fang_alpha", Name: "Fang Alpha", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 2, Attack: 2, Health: 2, Tribe: TribeBeast,
		Text: "Adjacent minions have +1 Attack.",
		Aura: &Aura{Atk: 1, Adjacent: true}},

	{ID: "reef_warchief", Name: "Reef Warchief", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 3, Attack: 3, Health: 3, Tribe: TribeGilkin,
		Text: "Your other Gilkins have +2 Attack.",
		Aura: &Aura{Atk: 2, Tribe: TribeGilkin}},

	{ID: "tidehook_captain", Name: "Tidehook Captain", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 3, Attack: 3, Health: 3, Tribe: TribePirate,
		Text: "Your other Pirates have +1/+1.",
		Aura: &Aura{Atk: 1, HP: 1, Tribe: TribePirate}},

	{ID: "tidescry_oracle", Name: "Tidescry Oracle", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 3, Attack: 2, Health: 3, Tribe: TribeGilkin,
		Text:     "Onset: Give your other Gilkins +2 Health.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, BuffHP: 2, Target: TargetNone, Area: AreaFriendlyTribe, Tribe: TribeGilkin}}}},

	{ID: "brackish_caller", Name: "Brackish Caller", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 1, Attack: 1, Health: 2, Tribe: TribeGilkin,
		Text:     "Whenever you summon a Gilkin, gain +1 Attack.",
		Triggers: []Trigger{{When: OnFriendlySummon, SubjectTribe: TribeGilkin, Effect: Effect{Kind: EffectBuff, BuffAtk: 1, Target: TargetSelf}}}},

	// --- Special legendaries (all-character AoE + summon-for-opponent) ---

	{ID: "cinder_baron", Name: "Cinder Baron", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 7, Attack: 7, Health: 5, Tribe: TribeElemental,
		Text:     "At the end of your turn, deal 2 damage to all other characters.",
		Triggers: []Trigger{{When: OnTurnEnd, Effect: Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaOtherCharacters}}}},

	{ID: "rotgut_horror", Name: "Rotgut Horror", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 5, Attack: 4, Health: 4, Tribe: TribeUndead,
		Text:     "Taunt. Final Gasp: Deal 2 damage to all characters.",
		Keywords: []Keyword{KeywordTaunt},
		Triggers: []Trigger{{When: OnDeath, Effect: Effect{Kind: EffectDamage, Amount: 2, Target: TargetNone, Area: AreaAllCharacters}}}},

	{ID: "the_gorehound", Name: "The Gorehound", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 6, Attack: 9, Health: 7, Tribe: TribeBeast,
		Text:     "Final Gasp: Summon a 3/3 Gorehound Whelp for your opponent.",
		Triggers: []Trigger{{When: OnDeath, Effect: Effect{Kind: EffectSummon, Summon: "gorehound_whelp", Count: 1, Target: TargetNone, SummonForOpponent: true}}}},

	{ID: "voidwyrm_tyrant", Name: "Voidwyrm Tyrant", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 10, Attack: 12, Health: 12, Tribe: TribeDragon,
		Text:     "Onset: Destroy all other minions and discard your hand.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDestroy, Target: TargetNone, Area: AreaOtherMinions, DiscardHand: true}}}},

	{ID: "cragmaw", Name: "Cragmaw", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 8, Attack: 7, Health: 7,
		Text:     "At the end of each turn, gain +1/+1.",
		Triggers: []Trigger{{When: OnAnyTurnEnd, Effect: Effect{Kind: EffectBuff, BuffAtk: 1, BuffHP: 1, Target: TargetSelf}}}},

	{ID: "revenant_priestess", Name: "Revenant Priestess", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 6, Attack: 5, Health: 7,
		Text:     "Onset: Summon all friendly minions that died this turn.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectResummonDead, Target: TargetNone}}}},

	// --- Random-generation legendaries / minions ---

	{ID: "gleamwing", Name: "Gleamwing", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 2, Attack: 3, Health: 2, Tribe: TribeDragon,
		Text:     "Onset: Add a random Legendary minion to your hand.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectGenerateRandom, Target: TargetNone, GenType: TypeMinion, GenRarity: RarityLegendary}}}},

	{ID: "sprocket_tinkerer", Name: "Sprocket Tinkerer", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 3, Attack: 3, Health: 3,
		Text:     "Onset: Transform another random minion into a 5/5 or a 1/1.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectTransformRandom, Target: TargetNone, GenIDs: []string{"thornback_saurian", "bramble_squirrel"}}}}},

	{ID: "wilds_beastcaller", Name: "Wilds Beastcaller", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 7, Attack: 5, Health: 5,
		Text:     "Onset: Summon a random Beast.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSummonRandom, Target: TargetNone, GenType: TypeMinion, GenTribe: TribeBeast}}}},

	{ID: "emberqueen_valtha", Name: "Emberqueen Valtha", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 9, Attack: 8, Health: 8, Tribe: TribeDragon,
		Text:     "Onset: Set a hero's Health to 15.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSetHealth, Amount: 15, Target: TargetHero}}}},

	// --- Cost-modification minions ---

	{ID: "mana_leech", Name: "Mana Leech", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 2, Attack: 2, Health: 2,
		Text:     "All minions cost (1) more.",
		CostAura: &CostAura{Delta: 1, Scope: CostScopeAll, Type: TypeMinion}},

	{ID: "pocket_conjurer", Name: "Pocket Conjurer", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 2, Attack: 2, Health: 2,
		Text:     "Your first minion each turn costs (1) less.",
		CostAura: &CostAura{Delta: -1, Scope: CostScopeFriendly, Type: TypeMinion, FirstMinionEachTurn: true}},

	{ID: "tidecolossus", Name: "Tidecolossus", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 10, Attack: 8, Health: 8,
		Text:     "Costs (1) less for each other minion on the battlefield.",
		CostRule: &CostRule{PerBoardMinion: -1}},

	{ID: "fizzle_sparkmuddle", Name: "Fizzle Sparkmuddle", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 2, Attack: 4, Health: 4,
		Text:     "Onset: Your opponent's spells cost (0) next turn.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectEnemySpellsFree, Target: TargetNone}}}},

	// --- Tribe-conditional onset ---

	{ID: "shellback_crab", Name: "Shellback Crab", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 1, Attack: 1, Health: 2, Tribe: TribeBeast,
		Text:     "Onset: Destroy a Gilkin and gain +2/+2.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDestroy, Target: TargetMinion, ReqTribe: TribeGilkin, SelfBuffAtk: 2, SelfBuffHP: 2}}}},

	// --- Misc legendaries / one-offs ---

	{ID: "nightmare_lord", Name: "Nightmare Lord", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 6, Attack: 7, Health: 5, Tribe: TribeDemon,
		Text:     "After you play a card, summon a 2/1 Satyr.",
		Triggers: []Trigger{{When: OnPlayCard, Effect: Effect{Kind: EffectSummon, Summon: "thornwood_satyr", Count: 1, Target: TargetNone}}}},

	{ID: "imp_warden", Name: "Imp Warden", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 3, Attack: 1, Health: 5, Tribe: TribeDemon,
		Text: "At the end of your turn, deal 1 damage to this minion and summon a 1/1 Imp.",
		Triggers: []Trigger{
			{When: OnTurnEnd, Effect: Effect{Kind: EffectDamage, Amount: 1, Target: TargetSelf}},
			{When: OnTurnEnd, Effect: Effect{Kind: EffectSummon, Summon: "imp_whelp", Count: 1, Target: TargetNone}}}},

	{ID: "runed_golem", Name: "Runed Golem", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 3, Attack: 4, Health: 2,
		Text:     "Charge. Onset: Give your opponent a Mana Crystal.",
		Keywords: []Keyword{KeywordCharge},
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectGiveOppMana, Target: TargetNone}}}},

	{ID: "runeward_sage", Name: "Runeward Sage", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 4, Attack: 2, Health: 5,
		Text:     "Onset: Give adjacent minions Spell Damage +1.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, Target: TargetNone, Area: AreaAdjacent, GrantSpellDamage: 1}}}},

	{ID: "grovelord_brakka", Name: "Grovelord Brakka", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 3, Attack: 5, Health: 5, Tribe: TribeBeast,
		Text:     "Onset: Give your opponent 2 Jungle Gifts.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectGenerate, Generate: "jungle_gift", Count: 2, ToOpponent: true, Target: TargetNone}}}},

	{ID: "lucky_angler", Name: "Lucky Angler", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 2, Attack: 0, Health: 4,
		Text:     "At the start of your turn, you have a 50% chance to draw an extra card.",
		Triggers: []Trigger{{When: OnTurnStart, Chance: 50, Effect: Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone}}}},

	{ID: "archivist_solenne", Name: "Archivist Solenne", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 2, Attack: 0, Health: 4,
		Text:         "Whenever a player casts a spell, put a copy into the other player's hand.",
		CopiesSpells: true},

	{ID: "tollkeeper_brute", Name: "Tollkeeper Brute", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 5, Attack: 7, Health: 6,
		Text:     "Your minions cost (3) more.",
		CostAura: &CostAura{Delta: 3, Scope: CostScopeFriendly, Type: TypeMinion}},

	{ID: "dread_buccaneer", Name: "Dread Buccaneer", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 4, Attack: 3, Health: 3, Tribe: TribePirate,
		Text:     "Taunt. Costs (1) less for each Attack of your weapon.",
		Keywords: []Keyword{KeywordTaunt},
		CostRule: &CostRule{PerOwnWeaponAttack: -1}},

	{ID: "grudge_smith", Name: "Grudge Smith", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 5, Attack: 4, Health: 6, Tribe: TribeUndead,
		Text:            "Enrage: Your weapon has +2 Attack.",
		EnrageWeaponAtk: 2},

	{ID: "chronlord_zhal", Name: "Chronlord Zhal", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 9, Attack: 8, Health: 8, Tribe: TribeDragon,
		Text:        "While in play, players have 15 seconds to take their turns.",
		TurnSeconds: 15},

	{ID: "tideblade_raider", Name: "Tideblade Raider", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 1, Attack: 2, Health: 1, Tribe: TribePirate,
		Text:             "Has Charge while you have a weapon.",
		ChargeWithWeapon: true},

	{ID: "moonfury_stalker", Name: "Moonfury Stalker", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 3, Attack: 3, Health: 3,
		Text:        "Enrage: +1 Attack and Twinstrike.",
		Enrage:      &Aura{Atk: 1},
		EnrageGrant: []Keyword{KeywordTwinstrike}},

	{ID: "clockwork_swapbot", Name: "Clockwork Swapbot", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 3, Attack: 0, Health: 3, Tribe: TribeMech,
		Text:     "At the start of your turn, swap this minion with a random one in your hand.",
		Triggers: []Trigger{{When: OnTurnStart, Effect: Effect{Kind: EffectSwapWithHand, Target: TargetSelf}}}},

	{ID: "dreamwarden_ylena", Name: "Dreamwarden Ylena", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 9, Attack: 4, Health: 12, Tribe: TribeDragon,
		Text: "At the end of your turn, add a random Dream card to your hand.",
		Triggers: []Trigger{{When: OnTurnEnd, Effect: Effect{Kind: EffectGenerateRandom, Target: TargetNone,
			GenIDs: []string{"dream_daydream", "dream_verdant_drake", "dream_gleeful_sister", "dream_waking_nightmare", "dream_emerald_reckoning"}}}}},

	// --- Hall of Fame clones (Classic cards later rotated to HoF; see
	//     .notes/classic-mapping.md "HALL OF FAME"). Wave HoF-1: pure data. ---

	{ID: "hexbreaker_warden", Name: "Hexbreaker Warden", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 4, Attack: 4, Health: 3,
		Text:     "Onset: Silence a minion.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSilence, Target: TargetMinion}}}},

	{ID: "cobalt_loreling", Name: "Cobalt Loreling", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 5, Attack: 4, Health: 4, Tribe: TribeDragon,
		Text:        "Spell Damage +1. Onset: Draw a card.",
		SpellDamage: 1,
		Triggers:    []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone}}}},

	{ID: "reckless_vanguard", Name: "Reckless Vanguard", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 5, Attack: 6, Health: 2,
		Text:     "Charge. Onset: Summon two 1/1 Whelps for your opponent.",
		Keywords: []Keyword{KeywordCharge},
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSummon, Summon: "emberwing_whelp", Count: 2, Target: TargetNone, SummonForOpponent: true}}}},

	{ID: "emberlord_vrakgar", Name: "Emberlord Vrakgar", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 8, Attack: 8, Health: 8, Tribe: TribeElemental,
		Text:     "Can't attack. At the end of your turn, deal 8 damage to a random enemy.",
		Keywords: []Keyword{KeywordCantAttack},
		Triggers: []Trigger{{When: OnTurnEnd, Effect: Effect{Kind: EffectDamage, Amount: 8, Target: TargetRandomEnemy}}}},

	// --- Hall of Fame Wave HoF-2/HoF-3 (small engine + new mechanics). ---

	{ID: "brineseer_diviner", Name: "Brineseer Diviner", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 3, Attack: 2, Health: 2, Tribe: TribeGilkin,
		Text: "Onset: Each player draws 2 cards.",
		Triggers: []Trigger{
			{When: OnPlay, Effect: Effect{Kind: EffectDraw, Amount: 2, Target: TargetNone}},
			{When: OnPlay, Effect: Effect{Kind: EffectDraw, Amount: 2, Target: TargetNone, ToOpponent: true}},
		}},

	{ID: "anguished_scribe", Name: "Anguished Scribe", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityCommon, Cost: 3, Attack: 1, Health: 4,
		Text:     "Whenever this minion takes damage, draw a card.",
		Triggers: []Trigger{{When: OnDamage, Effect: Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone}}}},

	{ID: "mesmer_adept", Name: "Mesmer Adept", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 5, Attack: 3, Health: 3,
		Text:     "Onset: If your opponent has 4 or more minions, take control of a random one.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectMindControl, Target: TargetNone, ReqOppMinions: 4}}}},

	{ID: "corsair_macaw", Name: "Corsair Macaw", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 2, Attack: 1, Health: 1, Tribe: TribeBeast,
		Text:     "Onset: Draw a Pirate from your deck.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectTutorTribe, Target: TargetNone, Tribe: TribePirate}}}},

	{ID: "crag_colossus", Name: "Crag Colossus", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 12, Attack: 8, Health: 8, Tribe: TribeElemental,
		Text:     "Costs (1) less for each other card in your hand.",
		CostRule: &CostRule{PerCardInHand: -1}},

	{ID: "magma_behemoth", Name: "Magma Behemoth", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityEpic, Cost: 20, Attack: 8, Health: 8, Tribe: TribeElemental,
		Text:     "Costs (1) less for each Health your hero is missing.",
		CostRule: &CostRule{PerMissingHealth: -1}},

	{ID: "brinelord_gorrak", Name: "Brinelord Gorrak", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 4, Attack: 2, Health: 4, Tribe: TribeGilkin,
		Text:         "Charge. Has +1 Attack for each other Gilkin in play.",
		Keywords:     []Keyword{KeywordCharge},
		SelfCountAtk: &SelfCountAtk{Tribe: TribeGilkin, Atk: 1}},

	{ID: "warhorn_chieftain", Name: "Warhorn Chieftain", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 5, Attack: 5, Health: 5,
		Text: "Onset: Give both players a random Warhorn Anthem.",
		Triggers: []Trigger{
			{When: OnPlay, Effect: Effect{Kind: EffectGenerateRandom, Target: TargetNone, GenIDs: warhornAnthems}},
			{When: OnPlay, Effect: Effect{Kind: EffectGenerateRandom, Target: TargetNone, ToOpponent: true, GenIDs: warhornAnthems}},
		}},

	{ID: "gearmaster_cog", Name: "Gearmaster Cog", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 6, Attack: 6, Health: 6, Tribe: TribeMech,
		Text:     "Onset: Summon a random Contraption.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSummonRandom, Target: TargetNone, GenIDs: cogContraptions}}}},

	{ID: "duskwarden_genmar", Name: "Duskwarden Genmar", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 6, Attack: 6, Health: 5,
		Text: "Start of Game: If your deck has only even-Cost cards, your starting Hero Power costs (1)."},

	{ID: "wraithqueen_selvara", Name: "Wraithqueen Selvara", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 6, Attack: 5, Health: 5, Tribe: TribeUndead,
		Text:     "Final Gasp: Take control of a random enemy minion.",
		Triggers: []Trigger{{When: OnDeath, Effect: Effect{Kind: EffectMindControl, Target: TargetNone}}}},

	{ID: "lunar_devourer", Name: "Lunar Devourer", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityLegendary, Cost: 9, Attack: 7, Health: 8, Tribe: TribeBeast,
		Text: "Start of Game: If your deck has only odd-Cost cards, upgrade your Hero Power."},

	{ID: "shadowtail_familiar", Name: "Shadowtail Familiar", Type: TypeMinion, Class: ClassNeutral, Rarity: RarityRare, Cost: 3, Attack: 3, Health: 3, Tribe: TribeBeast,
		Text:        "Spell Damage +1. Onset: If your deck has only odd-Cost cards, draw a card.",
		SpellDamage: 1,
		Triggers:    []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone, ReqDeckAllOdd: true}}}},

	// --- Basic set (the free starter cards). No rarity gem (Basic cards have
	// none). Mechanics 1:1 with the genre's staple starter set; names + art
	// wholly original (custom IP). Three Basic neutrals need engine work and land
	// in a later batch: a heal-all-friendly-characters minion, a
	// +1/+1-per-other-friendly-minion warlord, and a destroy-enemy-weapon minion
	// (the latter gated on weapons, Phase 8).

	// 1-cost
	{ID: "sylvan_archer", Name: "Sylvan Archer", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 1,
		Text:     "Onset: Deal 1 damage.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 1, Target: TargetAny}}}},

	{ID: "hearthguard_footman", Name: "Hearthguard Footman", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 2,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},

	{ID: "brineherald", Name: "Brineherald", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 1, Tribe: TribeGilkin,
		Text: "Your other Gilkin have +1 Attack.",
		Aura: &Aura{Atk: 1, Tribe: TribeGilkin}},

	{ID: "reef_raider", Name: "Reef Raider", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 2, Health: 1, Tribe: TribeGilkin},

	{ID: "tusker_runt", Name: "Tusker Runt", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 1, Tribe: TribeBeast,
		Text: "Charge.", Keywords: []Keyword{KeywordCharge}},

	{ID: "hexbone_healer", Name: "Hexbone Healer", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 2, Health: 1,
		Text:     "Onset: Restore 2 Health.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectHeal, Amount: 2, Target: TargetAny}}}},

	// 2-cost
	{ID: "corroding_ooze", Name: "Corroding Ooze", Type: TypeMinion, Class: ClassNeutral, Cost: 2, Attack: 3, Health: 2,
		Text:     "Onset: Destroy your opponent's weapon.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDestroyWeapon, Target: TargetNone}}}},

	{ID: "mirefang_raptor", Name: "Mirefang Raptor", Type: TypeMinion, Class: ClassNeutral, Cost: 2, Attack: 3, Health: 2, Tribe: TribeBeast},

	{ID: "finblade_warrior", Name: "Finblade Warrior", Type: TypeMinion, Class: ClassNeutral, Cost: 2, Attack: 2, Health: 1, Tribe: TribeGilkin,
		Text: "Charge.", Keywords: []Keyword{KeywordCharge}},

	{ID: "frostpaw_grunt", Name: "Frostpaw Grunt", Type: TypeMinion, Class: ClassNeutral, Cost: 2, Attack: 2, Health: 2,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},

	{ID: "runescale_kobold", Name: "Runescale Kobold", Type: TypeMinion, Class: ClassNeutral, Cost: 2, Attack: 2, Health: 2,
		Text: "Spell Damage +1.", SpellDamage: 1},

	{ID: "tideling_hunter", Name: "Tideling Hunter", Type: TypeMinion, Class: ClassNeutral, Cost: 2, Attack: 2, Health: 1, Tribe: TribeGilkin,
		Text:     "Onset: Summon a 1/1 Tideling Scout.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSummon, Summon: "tideling_scout", Count: 1, Target: TargetNone}}}},

	{ID: "tinker_novice", Name: "Tinker Novice", Type: TypeMinion, Class: ClassNeutral, Cost: 2, Attack: 1, Health: 1,
		Text:     "Onset: Draw a card.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone}}}},

	{ID: "river_snapper", Name: "River Snapper", Type: TypeMinion, Class: ClassNeutral, Cost: 2, Attack: 2, Health: 3, Tribe: TribeBeast},

	// 3-cost
	{ID: "spire_mage", Name: "Spire Mage", Type: TypeMinion, Class: ClassNeutral, Cost: 3, Attack: 1, Health: 4,
		Text: "Spell Damage +1.", SpellDamage: 1},

	{ID: "forge_rifleman", Name: "Forge Rifleman", Type: TypeMinion, Class: ClassNeutral, Cost: 3, Attack: 2, Health: 2,
		Text:     "Onset: Deal 1 damage.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 1, Target: TargetAny}}}},

	{ID: "ironfur_bear", Name: "Ironfur Bear", Type: TypeMinion, Class: ClassNeutral, Cost: 3, Attack: 3, Health: 3, Tribe: TribeBeast,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},

	{ID: "magma_brute", Name: "Magma Brute", Type: TypeMinion, Class: ClassNeutral, Cost: 3, Attack: 5, Health: 1, Tribe: TribeElemental},

	{ID: "warband_leader", Name: "Warband Leader", Type: TypeMinion, Class: ClassNeutral, Cost: 3, Attack: 2, Health: 2,
		Text: "Your other minions have +1 Attack.",
		Aura: &Aura{Atk: 1}},

	{ID: "razorthorn_hunter", Name: "Razorthorn Hunter", Type: TypeMinion, Class: ClassNeutral, Cost: 3, Attack: 2, Health: 3,
		Text:     "Onset: Summon a 1/1 Thornback Boar.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSummon, Summon: "thornback_boar", Count: 1, Target: TargetNone}}}},

	{ID: "sunderlight_cleric", Name: "Sunderlight Cleric", Type: TypeMinion, Class: ClassNeutral, Cost: 3, Attack: 3, Health: 2,
		Text:     "Onset: Give a friendly minion +1/+1.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, BuffAtk: 1, BuffHP: 1, Target: TargetFriendlyMinion}}}},

	{ID: "silverback_elder", Name: "Silverback Elder", Type: TypeMinion, Class: ClassNeutral, Cost: 3, Attack: 1, Health: 4, Tribe: TribeBeast,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},

	{ID: "dire_rider", Name: "Dire Rider", Type: TypeMinion, Class: ClassNeutral, Cost: 3, Attack: 3, Health: 1,
		Text: "Charge.", Keywords: []Keyword{KeywordCharge}},

	// 4-cost
	{ID: "frostwind_brute", Name: "Frostwind Brute", Type: TypeMinion, Class: ClassNeutral, Cost: 4, Attack: 4, Health: 5},

	{ID: "whelpforge_mechanic", Name: "Whelpforge Mechanic", Type: TypeMinion, Class: ClassNeutral, Cost: 4, Attack: 2, Health: 4,
		Text:     "Onset: Summon a 2/1 Clockwork Whelp.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectSummon, Summon: "clockwork_whelp", Count: 1, Target: TargetNone}}}},

	{ID: "tinker_inventor", Name: "Tinker Inventor", Type: TypeMinion, Class: ClassNeutral, Cost: 4, Attack: 2, Health: 4,
		Text:     "Onset: Draw a card.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDraw, Amount: 1, Target: TargetNone}}}},

	{ID: "marsh_snapjaw", Name: "Marsh Snapjaw", Type: TypeMinion, Class: ClassNeutral, Cost: 4, Attack: 2, Health: 7, Tribe: TribeBeast},

	{ID: "runefist_ogre", Name: "Runefist Ogre", Type: TypeMinion, Class: ClassNeutral, Cost: 4, Attack: 4, Health: 4,
		Text: "Spell Damage +1.", SpellDamage: 1},

	{ID: "bulwark_shieldmaster", Name: "Bulwark Shieldmaster", Type: TypeMinion, Class: ClassNeutral, Cost: 4, Attack: 3, Health: 5,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},

	{ID: "bastion_knight", Name: "Bastion Knight", Type: TypeMinion, Class: ClassNeutral, Cost: 4, Attack: 2, Health: 5,
		Text: "Charge.", Keywords: []Keyword{KeywordCharge}},

	// 5-cost
	{ID: "harbor_bodyguard", Name: "Harbor Bodyguard", Type: TypeMinion, Class: ClassNeutral, Cost: 5, Attack: 5, Health: 4,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},

	{ID: "darkscale_mender", Name: "Darkscale Mender", Type: TypeMinion, Class: ClassNeutral, Cost: 5, Attack: 4, Health: 5,
		Text:     "Onset: Restore 2 Health to all friendly characters.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectHeal, Amount: 2, Target: TargetNone, Area: AreaFriendlyChars}}}},

	{ID: "frostpaw_warlord", Name: "Frostpaw Warlord", Type: TypeMinion, Class: ClassNeutral, Cost: 5, Attack: 4, Health: 4,
		Text:     "Onset: Gain +1/+1 for each other friendly minion.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectBuff, BuffAtk: 1, BuffHP: 1, Target: TargetSelf, PerOtherFriendlyMinion: true}}}},

	{ID: "gorebound_berserker", Name: "Gorebound Berserker", Type: TypeMinion, Class: ClassNeutral, Cost: 5, Attack: 2, Health: 7,
		Text: "Enrage: +3 Attack.", Enrage: &Aura{Atk: 3}},

	{ID: "duskblade", Name: "Duskblade", Type: TypeMinion, Class: ClassNeutral, Cost: 5, Attack: 4, Health: 4,
		Text:     "Onset: Deal 3 damage to the enemy hero.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 3, Target: TargetNone, Area: AreaEnemyHero}}}},

	{ID: "ironpike_commando", Name: "Ironpike Commando", Type: TypeMinion, Class: ClassNeutral, Cost: 5, Attack: 4, Health: 2,
		Text:     "Onset: Deal 2 damage.",
		Triggers: []Trigger{{When: OnPlay, Effect: Effect{Kind: EffectDamage, Amount: 2, Target: TargetAny}}}},

	// 6-cost
	{ID: "elder_spellweaver", Name: "Elder Spellweaver", Type: TypeMinion, Class: ClassNeutral, Cost: 6, Attack: 4, Health: 7,
		Text: "Spell Damage +1.", SpellDamage: 1},

	{ID: "crag_ogre", Name: "Crag Ogre", Type: TypeMinion, Class: ClassNeutral, Cost: 6, Attack: 6, Health: 7},

	{ID: "arena_champion", Name: "Arena Champion", Type: TypeMinion, Class: ClassNeutral, Cost: 6, Attack: 6, Health: 5,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},

	{ID: "reckless_skyrider", Name: "Reckless Skyrider", Type: TypeMinion, Class: ClassNeutral, Cost: 6, Attack: 5, Health: 2,
		Text: "Charge.", Keywords: []Keyword{KeywordCharge}},

	// 7-cost
	{ID: "molten_hound", Name: "Molten Hound", Type: TypeMinion, Class: ClassNeutral, Cost: 7, Attack: 9, Health: 5, Tribe: TribeBeast},

	{ID: "battlehorn_champion", Name: "Battlehorn Champion", Type: TypeMinion, Class: ClassNeutral, Cost: 7, Attack: 6, Health: 6,
		Text: "Your other minions have +1/+1.",
		Aura: &Aura{Atk: 1, HP: 1}},

	{ID: "war_colossus", Name: "War Colossus", Type: TypeMinion, Class: ClassNeutral, Cost: 7, Attack: 7, Health: 7},

	// --- Tokens (summon-only; excluded from decks and Seek) ---

	// Basic-set tokens (Tideling Hunter, Razorthorn Hunter, Whelpforge Mechanic).
	{ID: "tideling_scout", Name: "Tideling Scout", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 1, Tribe: TribeGilkin, Token: true},
	{ID: "thornback_boar", Name: "Thornback Boar", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 1, Tribe: TribeBeast, Token: true},
	{ID: "clockwork_whelp", Name: "Clockwork Whelp", Type: TypeMinion, Class: ClassNeutral, Cost: 2, Attack: 2, Health: 1, Tribe: TribeMech, Token: true},

	{ID: "broken_golem", Name: "Broken Golem", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 2, Health: 1, Token: true},
	{ID: "errant_squire", Name: "Errant Squire", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 2, Health: 2, Token: true},
	{ID: "hornelder_heir", Name: "Hornelder Heir", Type: TypeMinion, Class: ClassNeutral, Cost: 4, Attack: 5, Health: 5, Token: true},
	{ID: "tutors_pupil", Name: "Tutor's Pupil", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 1, Token: true},
	{ID: "snarl_pup", Name: "Snarl Pup", Type: TypeMinion, Class: ClassNeutral, Cost: 2, Attack: 2, Health: 2, Token: true,
		Text: "Taunt.", Keywords: []Keyword{KeywordTaunt}},
	{ID: "emberwing_whelp", Name: "Emberwing Whelp", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 1, Token: true},
	{ID: "gorehound_whelp", Name: "Gorehound Whelp", Type: TypeMinion, Class: ClassNeutral, Cost: 3, Attack: 3, Health: 3, Token: true},
	{ID: "thornback_saurian", Name: "Thornback Saurian", Type: TypeMinion, Class: ClassNeutral, Cost: 5, Attack: 5, Health: 5, Tribe: TribeBeast, Token: true},
	{ID: "bramble_squirrel", Name: "Bramble Squirrel", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 1, Tribe: TribeBeast, Token: true},
	{ID: "thornwood_satyr", Name: "Thornwood Satyr", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 2, Health: 1, Tribe: TribeDemon, Token: true},
	{ID: "imp_whelp", Name: "Imp Whelp", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 1, Tribe: TribeDemon, Token: true},
	{ID: "jungle_gift", Name: "Jungle Gift", Type: TypeSpell, Class: ClassNeutral, Cost: 1, Token: true,
		Text:   "Give a minion +1/+1.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 1, BuffHP: 1, Target: TargetMinion}},

	// Dream cards — the random pool Dreamwarden Ylena adds to hand (summon/generate
	// only; never built into a deck).
	{ID: "dream_daydream", Name: "Daydream", Type: TypeSpell, Class: ClassNeutral, Cost: 0, Token: true,
		Text:   "Return a minion to its owner's hand.",
		Effect: &Effect{Kind: EffectBounce, Target: TargetMinion}},
	{ID: "dream_verdant_drake", Name: "Verdant Drake", Type: TypeMinion, Class: ClassNeutral, Cost: 4, Attack: 7, Health: 6, Tribe: TribeDragon, Token: true},
	{ID: "dream_gleeful_sister", Name: "Gleeful Sister", Type: TypeMinion, Class: ClassNeutral, Cost: 3, Attack: 3, Health: 5, Token: true,
		Text: "Can't be targeted by spells or Hero Powers.", Keywords: []Keyword{KeywordElusive}},
	{ID: "dream_waking_nightmare", Name: "Waking Nightmare", Type: TypeSpell, Class: ClassNeutral, Cost: 0, Token: true,
		Text:   "Give a minion +5/+5. At the start of your next turn, destroy it.",
		Effect: &Effect{Kind: EffectBuff, BuffAtk: 5, BuffHP: 5, Target: TargetMinion, DestroyNextTurn: true}},
	{ID: "dream_emerald_reckoning", Name: "Emerald Reckoning", Type: TypeSpell, Class: ClassNeutral, Cost: 2, Token: true,
		Text:   "Deal 5 damage to all characters except Dreamwardens.",
		Effect: &Effect{Kind: EffectDamage, Amount: 5, Target: TargetNone, Area: AreaAllCharacters, ExceptCardID: "dreamwarden_ylena"}},

	// Warhorn Anthems — the random spell tokens Warhorn Chieftain gives both
	// players. Generate-only; never built into a deck.
	{ID: "anthem_muster", Name: "Anthem of the Muster", Type: TypeSpell, Class: ClassNeutral, Cost: 0, Token: true,
		Text:   "Summon three, four, or five 1/1 Recruits.",
		Effect: &Effect{Kind: EffectSummon, Summon: "tide_recruit", Count: 3, CountMax: 5, Target: TargetNone}},
	{ID: "anthem_warsong", Name: "Anthem of War", Type: TypeSpell, Class: ClassNeutral, Cost: 0, Token: true,
		Text:   "Summon a random Warsong fighter.",
		Effect: &Effect{Kind: EffectSummonRandom, Target: TargetNone, GenIDs: []string{"warsong_grunt", "warsong_reaver"}}},
	{ID: "anthem_ambush", Name: "Anthem of Ambush", Type: TypeSpell, Class: ClassNeutral, Cost: 0, Token: true,
		Text:   "Deal 4 damage. Draw a card.",
		Effect: &Effect{Kind: EffectDamage, Amount: 4, Target: TargetAny, ThenDraw: 1}},
	{ID: "tide_recruit", Name: "Recruit", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 1, Token: true},
	{ID: "warsong_grunt", Name: "Warsong Grunt", Type: TypeMinion, Class: ClassNeutral, Cost: 4, Attack: 4, Health: 2, Token: true},
	{ID: "warsong_reaver", Name: "Warsong Reaver", Type: TypeMinion, Class: ClassNeutral, Cost: 4, Attack: 2, Health: 4, Token: true},

	// Contraptions — the random Mech tokens Gearmaster Cog summons. Summon-only.
	{ID: "cog_emboldener", Name: "Emboldener Cog", Type: TypeMinion, Class: ClassNeutral, Cost: 0, Attack: 0, Health: 4, Tribe: TribeMech, Token: true,
		Text:     "At the end of your turn, give a random friendly minion +1/+1.",
		Triggers: []Trigger{{When: OnTurnEnd, Effect: Effect{Kind: EffectBuff, BuffAtk: 1, BuffHP: 1, Target: TargetRandomFriendly}}}},
	{ID: "cog_beacon", Name: "Beacon Cog", Type: TypeMinion, Class: ClassNeutral, Cost: 0, Attack: 0, Health: 1, Tribe: TribeMech, Token: true,
		Text: "At the start of your turn, destroy this minion and draw 3 cards.",
		Triggers: []Trigger{
			{When: OnTurnStart, Effect: Effect{Kind: EffectDraw, Amount: 3, Target: TargetNone}},
			{When: OnTurnStart, Effect: Effect{Kind: EffectDestroy, Target: TargetSelf}},
		}},
	{ID: "cog_polymorpher", Name: "Polymorpher Cog", Type: TypeMinion, Class: ClassNeutral, Cost: 0, Attack: 0, Health: 3, Tribe: TribeMech, Token: true,
		Text:     "At the start of your turn, transform a random minion into a 1/1 Chick.",
		Triggers: []Trigger{{When: OnTurnStart, Effect: Effect{Kind: EffectTransformRandom, Target: TargetSelf, GenIDs: []string{"cog_chick"}}}}},
	{ID: "cog_mender", Name: "Mender Cog", Type: TypeMinion, Class: ClassNeutral, Cost: 0, Attack: 0, Health: 3, Tribe: TribeMech, Token: true,
		Text:     "At the end of your turn, restore 6 Health to your hero.",
		Triggers: []Trigger{{When: OnTurnEnd, Effect: Effect{Kind: EffectHeal, Amount: 6, Target: TargetFriendlyHero}}}},
	{ID: "cog_chick", Name: "Chick", Type: TypeMinion, Class: ClassNeutral, Cost: 1, Attack: 1, Health: 1, Tribe: TribeBeast, Token: true},

	// Mana Surge — granted to the second player (Phase 9). Token spell: not in any
	// deck or Seek pool.
	{ID: "mana_surge", Name: "Mana Surge", Type: TypeSpell, Class: ClassNeutral, Cost: 0, Token: true,
		Text:   "Gain 1 Mana Crystal this turn only.",
		Effect: &Effect{Kind: EffectMana, Amount: 1, Target: TargetNone}},
}

// warhornAnthems / cogContraptions are the random token pools Warhorn Chieftain
// and Gearmaster Cog draw from (defined here, next to the tokens they reference).
var (
	warhornAnthems  = []string{"anthem_muster", "anthem_warsong", "anthem_ambush"}
	cogContraptions = []string{"cog_emboldener", "cog_beacon", "cog_polymorpher", "cog_mender"}
)
