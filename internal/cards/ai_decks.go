package cards

// Prebuilt AI opponent decks. One is picked at random per vs-AI match. Each class
// offers exactly TWO archetypes — a fast FACE deck and a slower MIDRANGE/tempo
// deck — so the bot plays to a coherent game plan instead of a random pile.
//
// Cards are chosen to be "bot-friendly": the AI is a greedy single-turn planner,
// so every card must be strong PLAYED ON CURVE with no combo, sequencing, or
// holding. Excluded across the board: spell-cast synergy payoffs, cost-reduction /
// "next thing free" enablers, self-bounce / copy / swap-with-hand, discard, and
// cards that need a set-up board state. Legendaries are included freely where they
// are simple standalone bombs (big bodies, recurring random damage, automatic
// deathrattle/end-of-turn value) — the deck rule caps each legendary id at 1 copy
// but allows many DISTINCT legendaries.
//
// Each must be a legal 30-card deck for its class (≤2 of any id, ≤1 of any
// legendary) — enforced by TestAIDecksAreLegal. When a new class becomes playable,
// add its two decks here and a case to AIDecks.

// --- Mage ---

var aiMageFace = []string{
	"shadow_prowler", "shadow_prowler",
	"tusker_runt", "tusker_runt",
	"frostlance", "frostlance",
	"arcane_barrage", "arcane_barrage",
	"mirefang_raptor", "mirefang_raptor",
	"rimebolt", "rimebolt",
	"magma_brute", "magma_brute",
	"dire_rider", "dire_rider",
	"forge_rifleman", "forge_rifleman",
	"pyrebolt", "pyrebolt",
	"ironpike_commando", "ironpike_commando",
	"duskblade", "duskblade",
	"reckless_skyrider", "reckless_skyrider",
	"snarlmaw",
	"nightmare_lord",
	"emberlord_vrakgar",
	"pyrecataclysm",
}

var aiMageMidrange = []string{
	"sylvan_archer", "sylvan_archer",
	"frostlance", "frostlance",
	"glacial_splinter", "glacial_splinter",
	"rimebolt", "rimebolt",
	"frostfont_elemental", "frostfont_elemental",
	"ironfur_bear", "ironfur_bear",
	"frostshear", "frostshear",
	"tinker_inventor", "tinker_inventor",
	"frostwind_brute", "frostwind_brute",
	"bulwark_shieldmaster", "bulwark_shieldmaster",
	"frost_tempest", "frost_tempest",
	"harbor_bodyguard", "harbor_bodyguard",
	"pyrebolt", "pyrebolt",
	"hornelder_chief",
	"cragmaw",
	"emberwing_matron",
	"spelltide_wyrm",
}

var aiMageDecks = [][]string{aiMageFace, aiMageMidrange}

// --- Hunter ---

var aiHunterFace = []string{
	"keen_arrow", "keen_arrow", "packleader_wolf", "packleader_wolf",
	"tusker_runt", "tusker_runt", "fledgling_hawk", "fledgling_hawk",
	"mirefang_raptor", "mirefang_raptor", "carrion_hyena", "carrion_hyena",
	"famished_vulture", "famished_vulture",
	"thornvale_panther", "thornvale_panther", "feral_command", "feral_command",
	"dire_rider", "dire_rider", "call_the_pack", "call_the_pack",
	"kennel_master", "kennel_master", "trampling_brute", "trampling_brute",
	"the_gorehound", "cinder_baron", "molten_hound", "apex_saurian",
}

var aiHunterMidrange = []string{
	"packleader_wolf", "packleader_wolf", "keen_arrow", "keen_arrow",
	"famished_vulture", "famished_vulture", "carrion_hyena", "carrion_hyena",
	"river_snapper", "river_snapper", "fang_alpha", "fang_alpha",
	"ironfur_bear", "ironfur_bear", "silverback_elder", "silverback_elder",
	"call_the_pack", "call_the_pack", "culling_shot", "culling_shot",
	"kennel_master", "kennel_master", "marsh_snapjaw",
	"blasting_shot", "blasting_shot", "trampling_brute",
	"snarlmaw", "hornelder_chief", "the_gorehound", "apex_saurian",
}

var aiHunterDecks = [][]string{aiHunterFace, aiHunterMidrange}

// --- Warrior ---

var aiWarriorFace = []string{
	"hone_edge", "hone_edge",
	"hammer_blow", "hammer_blow",
	"goading_strike", "goading_strike",
	"cindersplit_axe", "cindersplit_axe",
	"tideblade_raider", "tideblade_raider",
	"finblade_warrior", "finblade_warrior",
	"whipcrack_overseer", "whipcrack_overseer",
	"valiant_strike", "valiant_strike",
	"wide_swing", "wide_swing",
	"dire_rider", "dire_rider",
	"ironguard_elite", "ironguard_elite",
	"deathblow_swing", "deathblow_swing",
	"duskblade", "duskblade",
	"runesteel_reaper",
	"reckless_skyrider",
	"the_gorehound",
	"warchief_gorthak",
}

var aiWarriorMidrange = []string{
	"hammer_blow", "hammer_blow",
	"steel_cyclone", "steel_cyclone",
	"cindersplit_axe", "cindersplit_axe",
	"whipcrack_overseer", "whipcrack_overseer",
	"bracing_guard", "bracing_guard",
	"frostpaw_grunt", "frostpaw_grunt",
	"ragebound_brute", "ragebound_brute",
	"ironfur_bear", "ironfur_bear",
	"forgehold_smith", "forgehold_smith",
	"ironguard_elite", "ironguard_elite",
	"bulwark_shieldmaster", "bulwark_shieldmaster",
	"runesteel_reaper", "runesteel_reaper",
	"pit_brawl",
	"hornelder_chief",
	"snarlmaw",
	"cinder_baron",
	"cragmaw",
	"spelltide_wyrm",
}

var aiWarriorDecks = [][]string{aiWarriorFace, aiWarriorMidrange}

// --- Warlock ---

var aiWarlockFace = []string{
	"ember_imp", "ember_imp",
	"hollow_guardian", "hollow_guardian",
	"shadow_prowler", "shadow_prowler",
	"plague_gremlin", "plague_gremlin",
	"mortal_whisper", "mortal_whisper",
	"hexfire", "hexfire",
	"thornvale_panther", "thornvale_panther",
	"dire_rider", "dire_rider",
	"shadow_lance", "shadow_lance",
	"duskblade", "duskblade",
	"doom_kiss",
	"reckless_skyrider", "reckless_skyrider",
	"dread_colossus", "dread_colossus",
	"reckless_vanguard",
	"the_gorehound",
	"molten_hound",
	"cragmaw",
	"dread_warden",
}

var aiWarlockMidrange = []string{
	"mortal_whisper", "mortal_whisper",
	"hexfire", "hexfire",
	"siphon_vitae", "siphon_vitae",
	"hollow_guardian", "hollow_guardian",
	"frostpaw_grunt", "frostpaw_grunt",
	"shadow_lance", "shadow_lance",
	"ironfur_bear", "ironfur_bear",
	"infernal_blaze", "infernal_blaze",
	"bulwark_shieldmaster",
	"doom_kiss", "doom_kiss",
	"harbor_bodyguard", "harbor_bodyguard",
	"soul_harvest",
	"grave_knight",
	"snarlmaw",
	"hornelder_chief",
	"dread_colossus",
	"emberlord_vrakgar",
	"dread_warden", "dread_warden",
	"overlord_xathul",
}

var aiWarlockDecks = [][]string{aiWarlockFace, aiWarlockMidrange}

// aiPriestFace is a fast Holy/Shadow tempo deck: cheap bodies + buffs + efficient
// removal (Searing Light, Gloom Word, Psychic Lance reach), bot-friendly (every card
// is strong played straight on curve, no combos/sequencing).
var aiPriestFace = []string{
	"crimson_subduer", "crimson_subduer",
	"searing_light", "searing_light",
	"dawnvale_acolyte", "dawnvale_acolyte",
	"psychic_lance", "psychic_lance",
	"gloom_word_ache", "gloom_word_ache",
	"harborlight_chaplain", "harborlight_chaplain",
	"silverback_elder", "silverback_elder",
	"radiant_burst", "radiant_burst",
	"zealots_blessing", "zealots_blessing",
	"moonsilver_guardian", "moonsilver_guardian",
	"sanctum_warden", "sanctum_warden",
	"gloom_word_demise", "gloom_word_demise",
	"harbor_bodyguard", "harbor_bodyguard",
	"crag_ogre",
	"war_colossus",
	"dawnguard_protector",
	"dominate_will",
}

// aiPriestMidrange is a slower control deck: heal/draw engine (Dawnward Sigil into
// Dawnvale Acolyte), heavy removal (both Gloom Words + Undoing AoE), sticky bodies,
// Pyre of Faith reach, topped by the doubling legendary. All cards stand alone.
var aiPriestMidrange = []string{
	"dawnward_sigil", "dawnward_sigil",
	"searing_light", "searing_light",
	"dawnvale_acolyte", "dawnvale_acolyte",
	"gloom_word_ache", "gloom_word_ache",
	"harborlight_chaplain", "harborlight_chaplain",
	"radiant_burst", "radiant_burst",
	"gloom_word_demise", "gloom_word_demise",
	"ironforge_brute", "granite_warden",
	"sanctum_warden", "sanctum_warden",
	"pyre_of_faith", "pyre_of_faith",
	"gloom_word_undoing",
	"moonsilver_guardian", "moonsilver_guardian",
	"harbor_bodyguard", "harbor_bodyguard",
	"dawnguard_protector",
	"crag_ogre",
	"war_colossus",
	"dominate_will",
	"oracle_velneth",
}

var aiPriestDecks = [][]string{aiPriestFace, aiPriestMidrange}

// --- Paladin ---

// aiPaladinFace is a weapon/tempo aggro deck: cheap weapons + weapon-synergy
// neutral bodies, Wrathful Hammer reach, Hallowed Ground clear, all strong played
// straight on curve (no secrets/combos — bot-friendly).
var aiPaladinFace = []string{
	"dawnmace", "dawnmace",
	"tideblade_raider", "tideblade_raider",
	"finblade_warrior", "finblade_warrior",
	"dawnguard_templar", "dawnguard_templar",
	"verdict_edge", "verdict_edge",
	"wrathhammer", "wrathhammer",
	"tusker_runt", "tusker_runt",
	"dire_rider", "dire_rider",
	"pureheart_blade", "pureheart_blade",
	"hallowed_ground", "hallowed_ground",
	"duskblade", "duskblade",
	"reckless_skyrider", "reckless_skyrider",
	"mirefang_raptor", "mirefang_raptor",
	"molten_hound",
	"the_gorehound",
	"war_colossus",
	"highlord_valdric",
}

// aiPaladinMidrange is a slower blessing/heal control deck: Aegis bodies, Wrathful
// Hammer + Hallowed Ground removal, Pureheart Blade + Radiant Mending sustain,
// sticky neutral walls, topped by the class legendary. Every card stands alone.
var aiPaladinMidrange = []string{
	"dawnmace", "dawnmace",
	"dawnguard_templar", "dawnguard_templar",
	"riftwarden_pacifier", "riftwarden_pacifier",
	"wrathhammer", "wrathhammer",
	"verdict_edge", "verdict_edge",
	"radiant_mending", "radiant_mending",
	"silverback_elder", "silverback_elder",
	"hallowed_ground", "hallowed_ground",
	"pureheart_blade", "pureheart_blade",
	"moonsilver_guardian", "moonsilver_guardian",
	"harbor_bodyguard", "harbor_bodyguard",
	"ironforge_brute",
	"granite_warden",
	"dawnguard_protector",
	"crown_guardian",
	"crag_ogre",
	"war_colossus",
	"laying_of_hands",
	"highlord_valdric",
}

var aiPaladinDecks = [][]string{aiPaladinFace, aiPaladinMidrange}

// aiDruidFace is a nature tempo/beast deck: cheap ramp + reach (Mana Bloom,
// Moonbeam, Feral Claws), Might of the Grove board buffs, Clawform Druid's Charge
// mode, Claw Sweep clear, Treant swarm (Call the Grove), sticky neutral bodies,
// topped by the class legendary. Leans on options the greedy bot plays straight
// (untargeted / self buffs / plain-targeted removal).
var aiDruidFace = []string{
	"mana_bloom", "mana_bloom",
	"moonbeam", "moonbeam",
	"feral_claws", "feral_claws",
	"might_of_the_grove", "might_of_the_grove",
	"claw_sweep", "claw_sweep",
	"silverback_elder", "silverback_elder",
	"grove_warden", "grove_warden",
	"clawform_druid", "clawform_druid",
	"harbor_bodyguard", "harbor_bodyguard",
	"dire_rider", "dire_rider",
	"tusker_runt", "tusker_runt",
	"mirefang_raptor", "mirefang_raptor",
	"starbolt",
	"call_the_grove",
	"elder_of_battle",
	"molten_hound",
	"war_colossus",
	"sylvaros",
}

// aiDruidMidrange is a slower nature ramp/control deck: mana acceleration (Mana
// Bloom / Verdant Growth / Verdant Bounty) into beefy bodies, Claw Sweep clear,
// Ancients, sturdy neutral walls, topped by the class legendary. Every card
// stands alone for the bot.
var aiDruidMidrange = []string{
	"mana_bloom", "mana_bloom",
	"verdant_growth", "verdant_growth",
	"feral_claws", "feral_claws",
	"silverback_elder", "silverback_elder",
	"grove_warden", "grove_warden",
	"verdant_bounty", "verdant_bounty",
	"clawform_druid", "clawform_druid",
	"harbor_bodyguard", "harbor_bodyguard",
	"moonsilver_guardian", "moonsilver_guardian",
	"claw_sweep", "claw_sweep",
	"might_of_the_grove", "might_of_the_grove",
	"barkhide_colossus",
	"elder_of_wisdom",
	"elder_of_battle",
	"wilds_gift",
	"dawnguard_protector",
	"crag_ogre",
	"war_colossus",
	"sylvaros",
}

var aiDruidDecks = [][]string{aiDruidFace, aiDruidMidrange}

// aiRogueTempo is a dagger/poison tempo deck: cheap removal (Blindside / Lacerate),
// reach (Sly Jab), Chain (Combo) bodies (Guild Ringleader / Guild Agent), a weapon,
// and Onset stealth/poison bodies over sticky neutral bodies, topped by the class
// legendary. Every card stands alone if the bot plays it without a Chain trigger.
var aiRogueTempo = []string{
	"blindside", "blindside",
	"sly_jab", "sly_jab",
	"quickstab", "quickstab",
	"lacerate", "lacerate",
	"blade_fan", "blade_fan",
	"guild_ringleader", "guild_ringleader",
	"guild_agent", "guild_agent",
	"silverback_elder", "silverback_elder",
	"assassins_edge",
	"final_cut", "final_cut",
	"masked_infiltrator", "masked_infiltrator",
	"plague_carrier", "plague_carrier",
	"harbor_bodyguard", "harbor_bodyguard",
	"dire_rider",
	"snatcher_brute",
	"molten_hound",
	"war_colossus",
	"shadowlord_vex",
}

// aiRogueMidrange is a slower dagger/control deck: bounce tempo (Waylay), removal
// (Blindside / Lacerate / Final Cut / Blade Fan), a weapon, sturdy neutral walls,
// and draw (Hasty Dash), topped by the class legendary. Stands alone for the bot.
var aiRogueMidrange = []string{
	"blindside", "blindside",
	"waylay", "waylay",
	"lacerate", "lacerate",
	"blade_fan", "blade_fan",
	"guild_agent", "guild_agent",
	"silverback_elder", "silverback_elder",
	"final_cut", "final_cut",
	"assassins_edge",
	"masked_infiltrator", "masked_infiltrator",
	"plague_carrier", "plague_carrier",
	"moonsilver_guardian", "moonsilver_guardian",
	"harbor_bodyguard", "harbor_bodyguard",
	"hasty_dash",
	"snatcher_brute",
	"crag_ogre",
	"dawnguard_protector",
	"war_colossus",
	"molten_hound",
	"shadowlord_vex",
}

var aiRogueDecks = [][]string{aiRogueTempo, aiRogueMidrange}

// aiShamanOverload is an elemental burn/tempo deck: cheap Overload burn
// (Voltstrike / Magma Burst), lightning bodies (Gale Wisp), AoE (Split Bolt /
// Tempest Surge), a weapon, an Overload payoff body, and sturdy Elemental/neutral
// walls topped by the class legendary. Every card stands alone on curve.
var aiShamanOverload = []string{
	"voltstrike", "voltstrike",
	"frost_jolt", "frost_jolt",
	"gale_wisp", "gale_wisp",
	"split_bolt", "split_bolt",
	"tempest_axe", "tempest_axe",
	"magma_burst", "magma_burst",
	"tempest_surge", "tempest_surge",
	"riftbound_elemental", "riftbound_elemental",
	"silverback_elder", "silverback_elder",
	"harbor_bodyguard", "harbor_bodyguard",
	"bedrock_elemental", "bedrock_elemental",
	"cinder_elemental", "cinder_elemental",
	"molten_hound",
	"war_colossus",
	"crag_ogre",
	"dire_rider",
	"arena_champion",
	"zephiron_stormlord",
}

// aiShamanMidrange is a slower elemental control deck: removal + AoE (Tempest
// Surge / Magma Burst), draw (Distant Sight / Tidewater Totem), Overload payoff,
// big Elemental bodies, and sturdy neutral walls, topped by the class legendary.
var aiShamanMidrange = []string{
	"frost_jolt", "frost_jolt",
	"voltstrike", "voltstrike",
	"tempest_surge", "tempest_surge",
	"magma_burst", "magma_burst",
	"distant_sight", "distant_sight",
	"tidewater_totem", "tidewater_totem",
	"riftbound_elemental", "riftbound_elemental",
	"galecaller", "galecaller",
	"silverback_elder", "silverback_elder",
	"moonsilver_guardian", "moonsilver_guardian",
	"bedrock_elemental", "bedrock_elemental",
	"cinder_elemental", "cinder_elemental",
	"crag_ogre",
	"dawnguard_protector",
	"war_colossus",
	"molten_hound",
	"harbor_bodyguard",
	"zephiron_stormlord",
}

var aiShamanDecks = [][]string{aiShamanOverload, aiShamanMidrange}

// StarterDeck is a named prebuilt deck offered to players in the deckbuilder as a
// one-click prefill. Cards mirror the vs-AI decks (AIDecks); the names are flavor
// only and belong to our custom set (no external IP).
type StarterDeck struct {
	Name  string
	Cards []string
}

// starterNames labels each class's two AIDecks archetypes, in the SAME order
// AIDecks returns them: [aggro/face, midrange/control]. Names reference our own
// custom cards/keywords only.
var starterNames = map[Class][2]string{
	ClassMage:    {"Emberlord Tempo", "Rimebound Control"},
	ClassHunter:  {"Pack Frenzy", "Apex Wild"},
	ClassWarrior: {"Reckless Onslaught", "Runesteel Wall"},
	ClassWarlock: {"Hexfire Rush", "Overlord's Pact"},
	ClassPriest:  {"Radiant Zeal", "Gloomward Control"},
	ClassPaladin: {"Dawnmace Tempo", "Hallowed Bulwark"},
	ClassDruid:   {"Grove Frenzy", "Verdant Ramp"},
	ClassRogue:   {"Shadow Tempo", "Venom Control"},
	ClassShaman:  {"Stormcaller Overload", "Elemental Control"},
}

// StarterDecks returns the named prefill decks for a class (fresh copies, safe for
// the caller to mutate), or nil if the class has none. Order matches AIDecks.
func StarterDecks(class Class) []StarterDeck {
	decks := AIDecks(class)
	if decks == nil {
		return nil
	}
	names := starterNames[class]
	out := make([]StarterDeck, len(decks))
	for i, d := range decks {
		name := ""
		if i < len(names) {
			name = names[i]
		}
		out[i] = StarterDeck{Name: name, Cards: d}
	}
	return out
}

// AIDecks returns copies of the prebuilt AI decks for a class, or nil if the
// class has none. The caller picks one at random.
func AIDecks(class Class) [][]string {
	var src [][]string
	switch class {
	case ClassMage:
		src = aiMageDecks
	case ClassHunter:
		src = aiHunterDecks
	case ClassWarrior:
		src = aiWarriorDecks
	case ClassWarlock:
		src = aiWarlockDecks
	case ClassPriest:
		src = aiPriestDecks
	case ClassPaladin:
		src = aiPaladinDecks
	case ClassDruid:
		src = aiDruidDecks
	case ClassRogue:
		src = aiRogueDecks
	case ClassShaman:
		src = aiShamanDecks
	default:
		return nil
	}
	out := make([][]string, len(src))
	for i, d := range src {
		out[i] = append([]string(nil), d...)
	}
	return out
}
