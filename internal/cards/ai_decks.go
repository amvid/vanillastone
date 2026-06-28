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
	default:
		return nil
	}
	out := make([][]string, len(src))
	for i, d := range src {
		out[i] = append([]string(nil), d...)
	}
	return out
}
