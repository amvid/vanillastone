// Package cards is the server's source of truth for card definitions. Phase 2
// added vanilla minions; Phase 3 adds spells with data-driven effects
// (damage/heal/buff) — the common effects HANDOFF says are data, not Go
// handlers. JSON-backed defs and real decks arrive in later phases. Custom IP
// only — see HANDOFF "Legal rules".
package cards

import (
	"fmt"
	"slices"
	"sort"
)

// Class is the card's class. Color in the client is driven by class, NOT by
// card type: neutral cards render in the default (parchment) frame, Mage cards
// in the class color (blue). Secrets are Mage spells, so they pick up the Mage
// color via this field — there is no per-type coloring anymore.
type Class string

const (
	ClassNeutral Class = "neutral"
	ClassMage    Class = "mage"
	ClassHunter  Class = "hunter" // reserved; no cards yet (deck class only, "coming soon")
)

// PlayableClasses lists the hero classes a deck may be built for. A deck binds
// to exactly one of these; its cards must be that class or neutral. Mage only
// for now (the game is Mage-only); Hunter is reserved but not yet playable.
func PlayableClasses() []Class { return []Class{ClassMage} }

// classPlayable reports whether decks may be built for this class.
func classPlayable(c Class) bool {
	for _, p := range PlayableClasses() {
		if p == c {
			return true
		}
	}
	return false
}

// Rarity is the card's collectible rarity (drives the rarity gem on the card
// face). Tokens and the hero power leave it empty.
type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityRare      Rarity = "rare"
	RarityEpic      Rarity = "epic"
	RarityLegendary Rarity = "legendary"
)

// Type distinguishes card kinds.
type Type string

const (
	TypeMinion    Type = "minion"
	TypeSpell     Type = "spell"
	TypeSecret    Type = "secret"    // played from hand into a hidden secret zone (Phase 7)
	TypeWeapon    Type = "weapon"    // equipped to the hero; attack + durability (Phase 8)
	TypeHeroPower Type = "heroPower" // the hero's reusable ability; not in hand (Phase 8)
)

// EffectKind is what an effect does. Spells and triggers (onset/finalGasp)
// share this vocabulary.
type EffectKind string

const (
	EffectDamage          EffectKind = "damage"
	EffectHeal            EffectKind = "heal"
	EffectBuff            EffectKind = "buff"
	EffectSummon          EffectKind = "summon"          // put token minions onto the caster's board
	EffectSilence         EffectKind = "silence"         // strip a minion's enchantments, keywords, triggers
	EffectSeek            EffectKind = "seek"            // present 3 cards from a pool; the player picks one (Phase 7)
	EffectMana            EffectKind = "mana"            // give the caster temporary mana this turn (Mana Surge, Phase 9)
	EffectDraw            EffectKind = "draw"            // the caster draws Amount cards (default 1) from their deck
	EffectGenerate        EffectKind = "generate"        // add a specific card (Effect.Generate id) to the caster's hand
	EffectDestroy         EffectKind = "destroy"         // destroy the target minion outright (ignores Aegis)
	EffectTransform       EffectKind = "transform"       // replace the target minion with a token (Effect.Transform id)
	EffectBounce          EffectKind = "bounce"          // return the target minion to its owner's hand as its base card
	EffectMissiles        EffectKind = "missiles"        // fire Count missiles of Amount damage, each at a random OTHER character
	EffectKillSecret      EffectKind = "killSecret"      // destroy a random enemy Secret (no reveal)
	EffectCopy            EffectKind = "copy"            // the played minion becomes a copy of the targeted minion (`visage_thief`)
	EffectSwapStats       EffectKind = "swapStats"       // swap the target minion's Attack and Health (`addled_brewer`)
	EffectConsumeShields  EffectKind = "consumeShields"  // strip every Aegis in play, buff self per shield removed (`crimson_reaver`)
	EffectResummonDead    EffectKind = "resummonDead"    // summon all of the caster's minions that died this turn (`revenant_priestess`)
	EffectGenerateRandom  EffectKind = "generateRandom"  // add a random card from a filtered pool to the caster's hand (`codex_of_insight`, `gleamwing`)
	EffectSummonRandom    EffectKind = "summonRandom"    // summon a random minion from a filtered pool (or explicit GenIDs) onto the caster's board (`wilds_beastcaller` / `gearmaster_cog`)
	EffectMindControl     EffectKind = "mindControl"     // take control of a random enemy minion (`wraithqueen_selvara` / `mesmer_adept`)
	EffectTutorTribe      EffectKind = "tutorTribe"      // draw a random card of Effect.Tribe from the caster's deck (`corsair_macaw`)
	EffectTransformRandom EffectKind = "transformRandom" // transform a random OTHER minion into a random token from GenIDs (`sprocket_tinkerer`)
	EffectSetHealth       EffectKind = "setHealth"       // set the target hero's Health to Amount (`emberqueen_valtha`)
	EffectFreeNextSecret  EffectKind = "freeNextSecret"  // the caster's next Secret this turn costs 0 (`spellwarden_magus`)
	EffectEnemySpellsFree EffectKind = "enemySpellsFree" // the opponent's spells cost 0 on their next turn (`fizzle_sparkmuddle`)
	EffectGiveOppMana     EffectKind = "giveOppMana"     // give the opponent an empty Mana Crystal (maxMana +1, capped) (`runed_golem`)
	EffectSwapWithHand    EffectKind = "swapWithHand"    // swap the source minion with a random minion in the caster's hand (`clockwork_swapbot`)

	// Weapon-manipulation battlecries (untargeted; operate on the hero weapons).
	EffectGainWeaponAttack EffectKind = "gainWeaponAttack" // buff self by the caster's weapon Attack (`tidereaver`)
	EffectChipWeapon       EffectKind = "chipWeapon"       // remove Amount Durability from the opponent's weapon (`brine_cutter`)
	EffectBuffWeapon       EffectKind = "buffWeapon"       // give the caster's weapon +BuffAtk Attack / +BuffHP Durability (`captain_brackwater`)
	EffectDestroyWeapon    EffectKind = "destroyWeapon"    // destroy the opponent's weapon; the caster draws cards = its Durability (`relic_breaker`)
)

// SeekPool selects the card pool an EffectSeek offers.
type SeekPool string

const (
	SeekSpell  SeekPool = "spell"
	SeekMinion SeekPool = "minion"
)

// TargetRule constrains what an effect may target. The player-chosen rules
// (any/friendlyMinion/enemyMinion) need a target picked at play time;
// none/randomEnemy resolve server-side with no player choice.
type TargetRule string

const (
	TargetNone           TargetRule = "none"   // untargeted (Area resolves it)
	TargetAny            TargetRule = "any"    // any character: any minion or either hero
	TargetMinion         TargetRule = "minion" // any minion, either side (not heroes)
	TargetFriendlyMinion TargetRule = "friendlyMinion"
	TargetEnemyMinion    TargetRule = "enemyMinion"
	TargetEnemy          TargetRule = "enemy"          // any enemy character: an enemy minion or the enemy hero
	TargetFriendlyHero   TargetRule = "friendlyHero"   // the caster's own hero
	TargetHero           TargetRule = "hero"           // either hero (friendly or enemy), not minions (`emberqueen_valtha`)
	TargetRandomEnemy    TargetRule = "randomEnemy"    // server picks a random enemy character (RNG)
	TargetSelf           TargetRule = "self"           // the effect's own source minion (edge triggers, e.g. self-buff)
	TargetRandomFriendly TargetRule = "randomFriendly" // a random OTHER friendly minion (trigger effects)
)

// AreaRule selects an untargeted group for mass effects.
type AreaRule string

const (
	AreaNone              AreaRule = ""
	AreaEnemyMinions      AreaRule = "enemyMinions"
	AreaEnemyHero         AreaRule = "enemyHero"
	AreaAllMinions        AreaRule = "allMinions"        // every minion on both boards
	AreaAdjacent          AreaRule = "adjacent"          // the minions either side of the anchor (excludes it)
	AreaSplash            AreaRule = "splash"            // the anchor minion AND the minions either side of it
	AreaRandomEnemyMinion AreaRule = "randomEnemyMinion" // one random enemy minion (optionally stat-filtered by MaxAttack)
	AreaFriendlyTribe     AreaRule = "friendlyTribe"     // the caster's OTHER friendly minions of Effect.Tribe
	AreaAllCharacters     AreaRule = "allCharacters"     // both heroes and every minion on both boards
	AreaOtherCharacters   AreaRule = "otherCharacters"   // every character except the anchor minion (self-anchored)
	AreaOtherMinions      AreaRule = "otherMinions"      // every minion on both boards except the anchor minion (self-anchored)
	AreaFriendlyChars     AreaRule = "friendlyChars"     // the caster's hero AND every friendly minion (`darkscale_mender`)
	AreaEnemyChars        AreaRule = "enemyChars"        // the enemy hero AND every enemy minion (`arcane_barrage` missiles)
)

// Effect is an effect's data-driven behavior. Amount is damage/heal magnitude;
// BuffAtk/BuffHP are buff deltas; Summon/Count drive token summons. Target/Area
// decide who it hits.
type Effect struct {
	Kind                   EffectKind `json:"kind"`
	Amount                 int        `json:"amount,omitempty"`
	BuffAtk                int        `json:"buffAtk,omitempty"`
	BuffHP                 int        `json:"buffHP,omitempty"`
	Target                 TargetRule `json:"target"`
	Area                   AreaRule   `json:"area,omitempty"`
	Summon                 string     `json:"summon,omitempty"`                 // token card ID (EffectSummon)
	Count                  int        `json:"count,omitempty"`                  // number of tokens (default 1)
	Freeze                 bool       `json:"freeze,omitempty"`                 // also Freeze each character the effect hits
	Lifesteal              bool       `json:"lifesteal,omitempty"`              // damage dealt heals the caster's hero
	Pool                   SeekPool   `json:"pool,omitempty"`                   // EffectSeek: which card pool to offer
	Generate               string     `json:"generate,omitempty"`               // EffectGenerate: card ID added to the caster's hand
	Transform              string     `json:"transform,omitempty"`              // EffectTransform: token card ID the target becomes
	Grant                  []Keyword  `json:"grant,omitempty"`                  // EffectBuff: keywords granted to each target (e.g. Taunt)
	Temporary              bool       `json:"temporary,omitempty"`              // EffectBuff: the buff lasts only until the end of the controller's turn
	MaxAttack              int        `json:"maxAttack,omitempty"`              // AreaRandomEnemyMinion: only consider minions with Attack <= this (0 = no filter)
	PerCardInHand          bool       `json:"perCardInHand,omitempty"`          // EffectBuff: scale the buff by the caster's current hand size
	PerOtherFriendlyMinion bool       `json:"perOtherFriendlyMinion,omitempty"` // EffectBuff: scale the buff by the caster's OTHER friendly minions (`frostpaw_warlord`)
	ReqAttack              int        `json:"reqAttack,omitempty"`              // targeted effect: the target minion must have Attack >= this (0 = no requirement)
	ReqTaunt               bool       `json:"reqTaunt,omitempty"`               // targeted effect: the target minion must have Taunt
	ReqTribe               Tribe      `json:"reqTribe,omitempty"`               // targeted effect: the target minion must be of this tribe (`shellback_crab`)
	DrawIfFrozen           int        `json:"drawIfFrozen,omitempty"`           // EffectDamage: draw this many cards if the target minion was Frozen (`glacial_splinter`)
	SelfBuffAtk            int        `json:"selfBuffAtk,omitempty"`            // onset rider: buff the played minion by this Attack after the effect resolves (`shellback_crab`)
	SelfBuffHP             int        `json:"selfBuffHP,omitempty"`             // onset rider: buff the played minion by this Health after the effect resolves (`shellback_crab`)
	GrantSpellDamage       int        `json:"grantSpellDamage,omitempty"`       // EffectBuff: grant the target minion +N Spell Damage (`runeward_sage`)
	ToOpponent             bool       `json:"toOpponent,omitempty"`             // EffectGenerate: add the card(s) to the OPPONENT's hand instead of the caster's (`grovelord_brakka`)
	DestroyNextTurn        bool       `json:"destroyNextTurn,omitempty"`        // EffectBuff: destroy the buffed minion at the start of its owner's next turn (`dream_waking_nightmare`)
	ExceptCardID           string     `json:"exceptCardID,omitempty"`           // AreaAllCharacters/AllMinions: skip minions with this card ID (`dream_emerald_reckoning` spares Dreamwardens)
	Tribe                  Tribe      `json:"tribe,omitempty"`                  // AreaFriendlyTribe: restrict the buff to other friendly minions of this tribe
	SummonForOpponent      bool       `json:"summonForOpponent,omitempty"`      // EffectSummon: summon onto the OPPONENT's board instead of the caster's
	DiscardHand            bool       `json:"discardHand,omitempty"`            // EffectDestroy: also discard the caster's remaining hand (`voidwyrm_tyrant`)
	FrozenDamage           int        `json:"frozenDamage,omitempty"`           // EffectDamage: if the (single) target is already Frozen, deal this instead of freezing (`frostlance`)
	ThenDraw               int        `json:"thenDraw,omitempty"`               // EffectDamage: the caster draws this many cards after the damage resolves
	CountMax               int        `json:"countMax,omitempty"`               // EffectSummon: when > Count, summon a random Count..CountMax tokens
	ReqOppMinions          int        `json:"reqOppMinions,omitempty"`          // EffectMindControl: only fire if the opponent controls at least this many minions (`mesmer_adept`)
	ReqDeckAllOdd          bool       `json:"reqDeckAllOdd,omitempty"`          // EffectDraw: only fire if every card left in the caster's deck is odd-cost (`shadowtail_familiar`)
	DrawWeaponDurability   bool       `json:"drawWeaponDurability,omitempty"`   // EffectDestroyWeapon: also draw cards = the broken weapon's durability (`relic_breaker`); plain destroy leaves it false (`corroding_ooze`)

	// Random-pool generation (EffectGenerateRandom / EffectSummonRandom): pick one
	// card at random from the collectible cards matching every set filter below.
	// EffectTransformRandom ignores these and uses GenIDs (its two token outcomes).
	GenClass  Class    `json:"genClass,omitempty"`
	GenType   Type     `json:"genType,omitempty"`
	GenRarity Rarity   `json:"genRarity,omitempty"`
	GenTribe  Tribe    `json:"genTribe,omitempty"`
	GenIDs    []string `json:"genIDs,omitempty"` // explicit outcome pool (EffectTransformRandom token ids)
}

// EventType identifies a game event a trigger reacts to. Phase 4 wires on_play
// (onset) and on_death (finalGasp); the rest of the taxonomy is data on
// top later.
type EventType string

const (
	OnPlay  EventType = "on_play"  // onset: fires when the minion is summoned from hand
	OnDeath EventType = "on_death" // finalGasp: fires when the minion dies

	// Edge triggers (Phase 4 extended): fire on ongoing game events for minions
	// already in play. All controller-scoped except OnAnyMinionDeath.
	OnTurnStart      EventType = "on_turn_start"       // the controller's turn begins
	OnTurnEnd        EventType = "on_turn_end"         // the controller's turn ends
	OnAnyTurnEnd     EventType = "on_any_turn_end"     // ANY turn ends (either player's) — global (`cragmaw`)
	OnSpellCast      EventType = "on_spell_cast"       // the controller casts a spell (incl. secrets)
	OnFriendlySummon EventType = "on_friendly_summon"  // another friendly minion is summoned
	OnFriendlyDeath  EventType = "on_friendly_death"   // another friendly minion dies
	OnAnyMinionDeath EventType = "on_any_minion_death" // any minion (either side) dies
	OnHeal           EventType = "on_heal"             // any character is healed (global)
	OnSecretPlayed   EventType = "on_secret_played"    // any Secret is played (global)
	OnPlayCard       EventType = "on_play_card"        // the controller plays any card (after it resolves)
	OnDamage         EventType = "on_damage"           // this minion takes damage (fires on the damaged minion only — draw-on-damage minion)

	// Secret triggers (Phase 7+): fire on the ENEMY's action, from the secret
	// owner's perspective.
	OnEnemyAttackHero EventType = "on_enemy_attack_hero" // an enemy MINION attacks the owner's hero
	OnEnemyPlayMinion EventType = "on_enemy_play_minion" // an enemy plays a minion
	OnEnemyCastSpell  EventType = "on_enemy_cast_spell"  // an enemy casts a spell
	OnHeroAttacked    EventType = "on_hero_attacked"     // the owner's hero is attacked (minion OR weapon)
	OnFatalDamage     EventType = "on_fatal_damage"      // the owner's hero would take fatal damage (`frostward_aegis`; fired only from the damageHero hook, never via triggerSecrets)
)

// SecretKind is a secret's Go-handled behavior (HANDOFF: weird effects are
// handlers by ID, not data). Each fires when its SecretDef.Trigger occurs.
type SecretKind string

const (
	SecretDestroyAttacker SecretKind = "destroyAttacker" // destroy the attacking minion, negate the attack
	SecretCopyMinion      SecretKind = "copyMinion"      // summon a copy of the played minion for the owner
	SecretCounterSpell    SecretKind = "counterSpell"    // counter the enemy spell (it fizzles)
	SecretGainArmor       SecretKind = "gainArmor"       // the owner's hero gains Amount armor
	SecretRetargetSpell   SecretKind = "retargetSpell"   // enemy casts a spell on the owner's minion → summon Summon and make it the new target (`decoy_ward`)
	SecretIceBlock        SecretKind = "iceBlock"        // fatal damage to the owner's hero → prevent it and make the hero Immune this turn (`frostward_aegis`)
)

// SecretDef binds a secret's trigger event to its behavior. Amount carries a
// magnitude for kinds that need one (e.g. gainArmor).
type SecretDef struct {
	Trigger EventType  `json:"trigger"`
	Kind    SecretKind `json:"kind"`
	Amount  int        `json:"amount,omitempty"`
	Summon  string     `json:"summon,omitempty"` // token id summoned by the secret (e.g. `decoy_ward`'s decoy)
}

// TriggerCondition gates an edge trigger: it fires only when the condition holds
// for the reacting minion's controller. Empty = always fires.
type TriggerCondition string

const (
	CondNone          TriggerCondition = ""
	CondControlSecret TriggerCondition = "controlSecret" // the controller has at least one active Secret
)

// Trigger binds an event to an effect. A minion's onset is an on_play
// trigger; its finalGasp is an on_death trigger. Condition (edge triggers
// only) further gates when it fires.
type Trigger struct {
	When         EventType        `json:"when"`
	Effect       Effect           `json:"effect"`
	Condition    TriggerCondition `json:"condition,omitempty"`
	SubjectTribe Tribe            `json:"subjectTribe,omitempty"` // summon/death triggers: fire only when the subject minion is of this tribe
	Chance       int              `json:"chance,omitempty"`       // percent chance the trigger fires (0 = always); `lucky_angler`'s 50% draw
}

// Keyword is a static minion ability baked into the card (Phase 5 wave 1).
// Aegis is the card's starting state; the live shield is tracked per
// instance in the match (it pops). Frozen is NOT a keyword — it's a runtime
// status applied by effects.
type Keyword string

const (
	KeywordTaunt       Keyword = "taunt"       // enemies must attack this first
	KeywordCharge      Keyword = "charge"      // may attack anything the turn it is played
	KeywordRush        Keyword = "rush"        // may attack minions (not heroes) the turn it is played
	KeywordAegis       Keyword = "aegis"       // first damage instance is negated
	KeywordTwinstrike  Keyword = "twinstrike"  // may attack twice per turn
	KeywordStealth     Keyword = "stealth"     // untargetable by the enemy until it attacks
	KeywordPoisonous   Keyword = "poisonous"   // any damage it deals to a minion destroys it
	KeywordLifesteal   Keyword = "lifesteal"   // damage it deals heals its controller's hero
	KeywordElusive     Keyword = "elusive"     // cannot be targeted by spells or hero powers (either side)
	KeywordCantAttack  Keyword = "cantAttack"  // can never attack
	KeywordFreezeOnHit Keyword = "freezeOnHit" // Freezes any character it deals combat damage to (`frostfont_elemental`)
)

// Tribe is a minion's creature type (drives tribal synergies). Most minions have
// none. Purely informational until a synergy card references it.
type Tribe string

const (
	TribeNone      Tribe = ""
	TribeBeast     Tribe = "beast"
	TribeGilkin    Tribe = "gilkin"
	TribePirate    Tribe = "pirate"
	TribeDragon    Tribe = "dragon"
	TribeMech      Tribe = "mech"
	TribeDemon     Tribe = "demon"
	TribeElemental Tribe = "elemental"
	TribeUndead    Tribe = "undead"
	TribeRiftborn  Tribe = "riftborn"
)

// Aura is a continuous stat buff a minion grants to the controller's other
// minions while it is in play and not silenced. Recomputed by the match each time
// the board changes. Tribe (optional) restricts it to other minions of that
// tribe; Adjacent restricts it to the immediate neighbours.
type Aura struct {
	Atk      int   `json:"atk,omitempty"`
	HP       int   `json:"hp,omitempty"`
	Tribe    Tribe `json:"tribe,omitempty"`    // only buff other minions of this tribe (empty = all)
	Adjacent bool  `json:"adjacent,omitempty"` // only buff the immediate neighbours
}

// CostScope decides whose hand cards a CostAura affects.
type CostScope string

const (
	CostScopeFriendly CostScope = "friendly" // only the aura controller's cards
	CostScopeAll      CostScope = "all"      // both players' cards
)

// CostAura is a continuous hand-cost modifier a minion grants while in play and
// not silenced. Delta>0 raises cost, Delta<0 lowers it (the final cost is floored
// at 0). Type (optional) restricts it to a card type; FirstMinionEachTurn
// (`pocket_conjurer`) only discounts the controller's first minion of the turn.
type CostAura struct {
	Delta               int       `json:"delta"`
	Scope               CostScope `json:"scope,omitempty"`
	Type                Type      `json:"type,omitempty"`
	FirstMinionEachTurn bool      `json:"firstMinionEachTurn,omitempty"`
}

// CostRule is an intrinsic per-card cost modifier: the card's own printed cost
// depends on board state. `tidecolossus`: -1 per minion on the battlefield;
// `dread_buccaneer`: -1 per point of the caster's weapon Attack.
type CostRule struct {
	PerBoardMinion     int `json:"perBoardMinion,omitempty"`     // delta per minion on either board (`tidecolossus`: -1)
	PerOwnWeaponAttack int `json:"perOwnWeaponAttack,omitempty"` // delta per point of the caster's weapon Attack (`dread_buccaneer`: -1)
	PerCardInHand      int `json:"perCardInHand,omitempty"`      // delta per OTHER card in the caster's hand (`crag_colossus`: -1)
	PerMissingHealth   int `json:"perMissingHealth,omitempty"`   // delta per point of Health the caster's hero is missing (`magma_behemoth`: -1)
}

// SelfCountAtk grants a minion +Atk Attack for each OTHER minion of Tribe on the
// battlefield (both boards). Recomputed with the auras (so Silence cancels it).
// `brinelord_gorrak`: +1 Attack per other Gilkin.
type SelfCountAtk struct {
	Tribe Tribe `json:"tribe"`
	Atk   int   `json:"atk"`
}

// Card is a card definition. Minions use Attack/Health; spells use Effect.
// Text is human-readable rules text shown in the client's hover box; vanilla
// minions with no special behavior leave it empty.
type Card struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Type             Type          `json:"type"`
	Class            Class         `json:"class,omitempty"`  // neutral (default) or mage — drives client color
	Rarity           Rarity        `json:"rarity,omitempty"` // common/rare/epic/legendary; empty for tokens & hero power
	Cost             int           `json:"cost"`
	Attack           int           `json:"attack,omitempty"`     // minions and weapons
	Health           int           `json:"health,omitempty"`     // minions only
	Durability       int           `json:"durability,omitempty"` // weapons only
	Text             string        `json:"text,omitempty"`
	Effect           *Effect       `json:"effect,omitempty"`           // spells only
	Triggers         []Trigger     `json:"triggers,omitempty"`         // minions: onset / finalGasp
	Keywords         []Keyword     `json:"keywords,omitempty"`         // minions: static keywords
	Tribe            Tribe         `json:"tribe,omitempty"`            // minions: creature type (tribal synergies)
	SpellDamage      int           `json:"spellDamage,omitempty"`      // minions: bonus added to the controller's spell damage
	Aura             *Aura         `json:"aura,omitempty"`             // minions: continuous buff to other friendly minions
	CostAura         *CostAura     `json:"costAura,omitempty"`         // minions: continuous hand-cost modifier while in play
	CostRule         *CostRule     `json:"costRule,omitempty"`         // intrinsic per-card cost modifier (cost depends on board state)
	SelfCountAtk     *SelfCountAtk `json:"selfCountAtk,omitempty"`     // minions: +Atk per other minion of a tribe in play (`brinelord_gorrak`)
	Enrage           *Aura         `json:"enrage,omitempty"`           // minions: stat bonus active only while damaged (Atk only; HP unsupported)
	EnrageWeaponAtk  int           `json:"enrageWeaponAtk,omitempty"`  // minions: while this is damaged, the controller's weapon gets +N Attack (`grudge_smith`)
	TurnSeconds      int           `json:"turnSeconds,omitempty"`      // minions: while in play, caps every turn to N seconds (`chronlord_zhal`)
	Secret           *SecretDef    `json:"secret,omitempty"`           // secrets only: trigger + behavior
	CopiesSpells     bool          `json:"copiesSpells,omitempty"`     // while in play, every spell cast adds a copy to the non-caster's hand (`archivist_solenne`)
	ChargeWithWeapon bool          `json:"chargeWithWeapon,omitempty"` // has Charge only while its controller has a weapon equipped (`tideblade_raider`)
	EnrageGrant      []Keyword     `json:"enrageGrant,omitempty"`      // keywords granted while this minion is damaged (`moonfury_stalker`: Twinstrike)
	Token            bool          `json:"token,omitempty"`            // summon-only; excluded from Seek/decks
}

// Has reports whether the card has the given keyword.
func (c Card) Has(k Keyword) bool {
	return slices.Contains(c.Keywords, k)
}

// Onset returns the minion's on_play effect, or nil if it has none.
func (c Card) Onset() *Effect {
	for i := range c.Triggers {
		if c.Triggers[i].When == OnPlay {
			return &c.Triggers[i].Effect
		}
	}
	return nil
}

// FinalGasps returns the minion's on_death effects in declared order.
func (c Card) FinalGasps() []Effect {
	return c.TriggersFor(OnDeath)
}

// TriggersFor returns the effects of every trigger reacting to event `when`, in
// declared order. Used by the match's edge-trigger dispatch.
func (c Card) TriggersFor(when EventType) []Effect {
	var out []Effect
	for _, t := range c.Triggers {
		if t.When == when {
			out = append(out, t.Effect)
		}
	}
	return out
}

// set is the full card registry, keyed by ID. It is assembled at init from the
// per-class card files: neutral.go (neutralCards, incl. tokens + Mana Surge) and
// mage.go (mageCards, incl. the Mage hero power). Mechanics follow well-worn
// genre staples; names and art are wholly original (custom IP). See HANDOFF
// "Legal rules".
var set = map[string]Card{}

func init() {
	for _, list := range [][]Card{neutralCards, mageCards} {
		for _, c := range list {
			if _, dup := set[c.ID]; dup {
				panic("duplicate card id: " + c.ID)
			}
			set[c.ID] = c
		}
	}
}

// Get returns the card with the given ID.
func Get(id string) (Card, bool) {
	c, ok := set[id]
	return c, ok
}

// MageHeroPower returns the Mage hero power (Fire Dart). The only hero for now.
func MageHeroPower() Card {
	return set["fire_dart"]
}

// UpgradedMageHeroPower returns the upgraded Mage hero power (`lunar_devourer`'s
// Start of Game): Deal 2 damage instead of 1. Returns a fresh copy with its own Effect so
// the shared set entry is never mutated.
func UpgradedMageHeroPower() Card {
	hp := set["fire_dart"]
	e := *hp.Effect
	e.Amount = 2
	hp.Effect = &e
	hp.Text = "Hero Power: Deal 2 damage."
	return hp
}

// Deck rules (Phase 9). A legal deck is exactly DeckSize cards, at most
// MaxCopies of any single card, every id drawn from the buildable pool.
const (
	DeckSize  = 30
	MaxCopies = 2
)

// ManaSurge returns Mana Surge, granted to the player going second.
func ManaSurge() Card { return set["mana_surge"] }

// inPool reports whether a card may be built into a deck: a real (non-token)
// card of a playable type. Hero powers and tokens are excluded.
func inPool(c Card) bool {
	if c.Token || c.Type == TypeHeroPower {
		return false
	}
	switch c.Type {
	case TypeMinion, TypeSpell, TypeSecret, TypeWeapon:
		return true
	}
	return false
}

// DeckPoolIDs returns the sorted ids of every card a deck may include.
func DeckPoolIDs() []string {
	var ids []string
	for id, c := range set {
		if inPool(c) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

// ValidateDeck checks a deck (a list of card ids) against the deck rules: exactly
// DeckSize cards, at most MaxCopies of any id, every id in the buildable pool,
// the deck bound to a playable class, and every card either that class or
// neutral. Returns nil if legal, otherwise an error describing the first
// violation.
func ValidateDeck(ids []string, class Class) error {
	if !classPlayable(class) {
		return fmt.Errorf("invalid deck class %q", class)
	}
	if len(ids) != DeckSize {
		return fmt.Errorf("deck must have exactly %d cards, got %d", DeckSize, len(ids))
	}
	counts := make(map[string]int, len(ids))
	for _, id := range ids {
		c, ok := set[id]
		if !ok {
			return fmt.Errorf("unknown card %q", id)
		}
		if !inPool(c) {
			return fmt.Errorf("card %q cannot be in a deck", id)
		}
		if c.Class != ClassNeutral && c.Class != class {
			return fmt.Errorf("card %q is not a %s or neutral card", id, class)
		}
		counts[id]++
		// Legendaries are limited to a single copy per deck.
		limit := MaxCopies
		if c.Rarity == RarityLegendary {
			limit = 1
		}
		if counts[id] > limit {
			return fmt.Errorf("too many copies of %q (max %d)", id, limit)
		}
	}
	return nil
}

// defaultMageDeck is a hand-curated, playable 30-card Mage deck used when a
// player queues without one of their own. It is a smooth-curve freeze/tempo
// list: cheap spell-synergy bodies + cheap removal early, sticky four-drops
// (Aegis / Taunt), and Frost Tempest as a board clear up top. Showcases
// the Mage class (freeze spells, spell-cost discount, `emberforge_magus`-style payoff)
// alongside solid neutral stats. One legendary (cap-legal). Keep it 30 cards,
// ≤2 of any id, ≤1 legendary — TestDefaultDeckIsLegal enforces it.
//
// When more classes become playable, give each its own curated default here and
// pick by class in DefaultDeck.
var defaultMageDeck = []string{
	// 1-drops: spell payoff + draw.
	"arcane_wyrmling", "arcane_wyrmling",
	"codex_of_insight", "codex_of_insight",
	// 2-drops: cheap removal, spell discount, elusive body.
	"arcane_adept", "arcane_adept",
	"glacial_splinter", "glacial_splinter",
	"glimmerwing_drake", "glimmerwing_drake",
	// 3-drops: freeze + tempo.
	"frostshear", "frostshear",
	"spellwarden_magus", "spellwarden_magus",
	// 4-drops: sticky bodies (grows / Aegis / vanilla / Taunt).
	"warded_scholar", "warded_scholar",
	"moonsilver_guardian", "moonsilver_guardian",
	"ironforge_brute",
	"granite_warden",
	// 5-drops: extra bodies + stealth threat.
	"errant_knight", "errant_knight",
	"jungle_stalker", "jungle_stalker",
	// 6-drops: board clear + freeze + Taunt/Aegis.
	"frost_tempest", "frost_tempest",
	"rime_elemental", "rime_elemental",
	"dawnguard_protector",
	// Top end: legendary spell payoff.
	"emberforge_magus",
}

// DefaultDeck returns a legal, curated 30-card Mage deck used when a player
// queues without having built one. The slice is copied so callers can't mutate
// the shared list.
func DefaultDeck() []string {
	return append([]string(nil), defaultMageDeck...)
}

// Deck materializes a list of card ids into Card values.
func Deck(ids []string) []Card {
	out := make([]Card, 0, len(ids))
	for _, id := range ids {
		if c, ok := set[id]; ok {
			out = append(out, c)
		}
	}
	return out
}

// RandomGenPoolIDs returns the sorted ids of every collectible (non-token, non
// hero-power) card matching all of the given filters. An empty filter value means
// "no constraint on that field". Used by random-generate / random-summon effects;
// the caller's RNG alone picks which id, so the sorted order keeps tests
// deterministic.
func RandomGenPoolIDs(class Class, typ Type, rarity Rarity, tribe Tribe) []string {
	var ids []string
	for id, c := range set {
		if c.Token || c.Type == TypeHeroPower {
			continue
		}
		if class != "" && c.Class != class {
			continue
		}
		if typ != "" && c.Type != typ {
			continue
		}
		if rarity != "" && c.Rarity != rarity {
			continue
		}
		if tribe != TribeNone && c.Tribe != tribe {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// SeekPoolIDs returns the card ids a Seek of the given pool may offer:
// non-token cards of the matching type. The result is sorted so the caller's RNG
// alone determines which are picked (deterministic for tests).
func SeekPoolIDs(pool SeekPool) []string {
	var want Type
	switch pool {
	case SeekSpell:
		want = TypeSpell
	case SeekMinion:
		want = TypeMinion
	default:
		return nil
	}
	var ids []string
	for id, c := range set {
		if c.Token || c.Type != want {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
