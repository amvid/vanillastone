package cards

// Prebuilt AI opponent decks. One is picked at random per vs-AI match. Each must
// be a legal 30-card deck for its class (≤2 of any id, ≤1 legendary) — enforced
// by TestAIDecksAreLegal. When a new class becomes playable, add its decks here
// and a case to AIDecks.

// aiMageAggro is a cheap, aggressive curve: lots of early bodies plus burn to
// close, leaning on the bot's "race when ahead" lethal lens.
var aiMageAggro = []string{
	"acolyte_novice", "acolyte_novice",
	"shadow_prowler", "shadow_prowler",
	"plague_gremlin", "plague_gremlin",
	"arcane_wyrmling", "arcane_wyrmling",
	"ashflame_zealot", "ashflame_zealot",
	"tavern_apprentice", "tavern_apprentice",
	"arcane_adept", "arcane_adept",
	"spellwarden_magus", "spellwarden_magus",
	"crimson_reaver", "crimson_reaver",
	"frostlance", "frostlance",
	"glacial_splinter", "glacial_splinter",
	"frostshear", "frostshear",
	"powder_tosser", "powder_tosser",
	"warded_scholar", "warded_scholar",
	"frost_tempest", "frost_tempest",
}

// aiMageMidrange is a stickier, value-oriented list: solid bodies up the curve,
// board clears, and the class legendary on top.
var aiMageMidrange = []string{
	"arcane_wyrmling", "arcane_wyrmling",
	"codex_of_insight", "codex_of_insight",
	"arcane_adept", "arcane_adept",
	"stoneveil_watcher", "stoneveil_watcher",
	"glimmerwing_drake", "glimmerwing_drake",
	"spellwarden_magus", "spellwarden_magus",
	"carrion_fiend", "carrion_fiend",
	"earthroot_healer", "earthroot_healer",
	"frostshear", "frostshear",
	"warded_scholar", "warded_scholar",
	"ironforge_brute", "ironforge_brute",
	"errant_knight", "errant_knight",
	"frost_tempest", "frost_tempest",
	"rime_elemental", "rime_elemental",
	"dawnguard_protector",
	"emberforge_magus",
}

// aiMageDecks is the pool the AI draws from for a Mage opponent (the curated
// default plus two flavors).
var aiMageDecks = [][]string{
	defaultMageDeck,
	aiMageAggro,
	aiMageMidrange,
}

// aiHunterAggro is a fast Beast-tempo curve: cheap bodies plus the class's
// tribal payoffs, leaning on the bot's "race when ahead" lethal lens.
var aiHunterAggro = []string{
	"keen_arrow", "keen_arrow",
	"packleader_wolf", "packleader_wolf",
	"tusker_runt", "tusker_runt",
	"famished_vulture", "famished_vulture",
	"carrion_hyena", "carrion_hyena",
	"fang_alpha", "fang_alpha",
	"mirefang_raptor", "mirefang_raptor",
	"river_snapper", "river_snapper",
	"call_the_pack", "call_the_pack",
	"ironfur_bear", "ironfur_bear",
	"kennel_master", "kennel_master",
	"razorthorn_hunter", "razorthorn_hunter",
	"trampling_brute", "trampling_brute",
	"mane_lioness", "mane_lioness",
	"molten_hound",
	"apex_saurian",
}

// aiHunterDecks is the pool the AI draws from for a Hunter opponent (the curated
// default plus an aggro flavor).
var aiHunterDecks = [][]string{
	defaultHunterDeck,
	aiHunterAggro,
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
	default:
		return nil
	}
	out := make([][]string, len(src))
	for i, d := range src {
		out[i] = append([]string(nil), d...)
	}
	return out
}
