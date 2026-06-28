// Package match holds a single in-memory 1v1 match and owns all game truth.
// Phase 2 models minions, mana, combat, and win/lose; Phase 3 spells; Phase 4
// adds triggers (onset/finalGasp) and an ordered event log. The server is
// authoritative: every mutation goes through the match mutex and the resulting
// full snapshot — plus the action's ordered event log — is pushed to both
// players (each sees only its own hand).
package match

import (
	"math/rand"
	"sync"
	"time"

	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/protocol"
)

const (
	maxMana    = 10
	maxBoard   = 7
	maxHand    = 10
	maxSecrets = 5
	startHP    = 30
	heroMaxHP  = 30

	openingFirst  = 3 // opening-hand size for the player going first
	openingSecond = 4 // opening-hand size for the player going second (also gets Mana Surge)

	maxHistory = 60 // rolling event log kept for reconnect replay (client shows the last 40)

	selfHeroTarget = "selfHero"
	oppHeroTarget  = "oppHero"

	turnLimit = 75 * time.Second // a turn auto-ends after this (HS-style)

	// mulliganLimit is the backstop for the opening mulligan: a player who hasn't
	// submitted by then is force-kept so a disconnect can't hang the match. The
	// client drives the visible ~20s countdown and auto-submits first; this sits
	// above that (plus the match-found splash) so it only bites a dead client.
	mulliganLimit = 30 * time.Second
)

// Sender is the minimal view of a connected player the match needs: an id and
// a way to push a server message. Implemented by transport.Client.
type Sender interface {
	ID() string
	Name() string // display username (for the opponent's nameplate)
	Send([]byte)
}

// enchant is a persistent stat modifier applied to a minion (e.g. Whetstone
// +2/+1). Enchantments stack and are stripped together by Silence (Phase 6),
// which is why buffs live here rather than being folded into base stats.
type enchant struct {
	atk         int
	hp          int
	spellDamage int             // Spell Damage granted by this buff (`runeward_sage`)
	keywords    []cards.Keyword // keywords granted by this buff (e.g. `bannerguard`'s Taunt)
	temp        bool            // expires at the end of the controller's turn ("+X this turn")

	// tempNextTurn buffs/debuffs expire at the START of tempOwner's next turn
	// ("until your next turn" — `crimson_subduer`'s -2 Attack on an enemy minion).
	tempNextTurn bool
	tempOwner    int
}

// minion is a minion instance in play. uid is unique within the match. Attack
// and max health are derived (base card stats + enchantments + aura) via
// atk()/maxHP(); only current health is stored. Attack eligibility (Phase 5/6):
// summonedThisTurn = summon sickness, attacksMade = attacks used this turn
// (eligibility = attacksMade < attacksPerTurn, which is 2 for Twinstrike — read
// live, so Twinstrike gained mid-turn still grants the extra attack). Charge/Rush
// bypass sickness; frozen blocks attacking;
// aegis negates the next damage instance; stealthed hides it from enemy
// targeting until it attacks; silenced strips its enchantments/keywords/triggers.
type minion struct {
	uid                string
	card               cards.Card
	owner              int             // player index that controls it (drives finalGasp sides/summons)
	enchants           []enchant       // persistent buffs (stripped by Silence)
	auraAtk            int             // attack granted by active auras, recomputed each board change
	auraHP             int             // max-health granted by active auras, recomputed each board change
	auraKeywords       []cards.Keyword // keywords granted by active auras (`tundra_charger`: Charge), recomputed each board change
	health             int             // current health; <= 0 means dead
	summonedThisTurn   bool
	attacksMade        int
	frozen             bool
	aegis              bool
	stealthed          bool
	silenced           bool
	destroyAtTurnStart bool // scheduled to die at the start of its owner's next turn (Nightmare)
	destroyAtTurnEnd   bool // scheduled to die at the end of THIS turn (`forbidden_might`)
	corrupted          bool // `creeping_rot`: scheduled to be destroyed at the start of corruptedBy's next turn
	corruptedBy        int  // player index whose turn-start destroys this minion (valid only when corrupted)
	returnAtTurnEnd    bool // `gloom_thrall`: a temporarily mind-controlled minion returns to returnTo at this turn's end
	returnTo           int  // owner index to return control to (valid only when returnAtTurnEnd)
}

// has reports whether the minion currently has keyword k: its card grants it and
// it has not been silenced.
func (mn *minion) has(k cards.Keyword) bool {
	// Silence permanently strips the card's OWN keywords (and enrage-granted ones),
	// but keywords granted by buffs/auras applied AFTER the silence still count —
	// silence is a one-time strip, not a permanent suppressor.
	if !mn.silenced && mn.card.Has(k) {
		return true
	}
	for _, e := range mn.enchants { // keywords granted by a buff (e.g. `bannerguard`'s Taunt)
		for _, gk := range e.keywords {
			if gk == k {
				return true
			}
		}
	}
	for _, gk := range mn.auraKeywords { // keywords granted by an in-play aura (`tundra_charger`: Charge)
		if gk == k {
			return true
		}
	}
	if mn.enraged() { // keywords granted only while damaged (`moonfury_stalker`: Twinstrike)
		for _, gk := range mn.card.EnrageGrant {
			if gk == k {
				return true
			}
		}
	}
	return false
}

// atk is the minion's current Attack: base card attack plus enchantments and
// aura, floored at 0.
func (mn *minion) atk() int {
	// `lumen_wisp`: Attack is always equal to current Health (silence cancels it).
	if !mn.silenced && mn.card.AtkEqualsHealth {
		if mn.health < 0 {
			return 0
		}
		return mn.health
	}
	a := mn.card.Attack + mn.auraAtk
	for _, e := range mn.enchants {
		a += e.atk
	}
	if mn.enraged() {
		a += mn.card.Enrage.Atk
	}
	if a < 0 {
		return 0
	}
	return a
}

// enraged reports whether the minion's Enrage bonus is currently active: it has
// an Enrage, is not silenced, and is damaged (current health below its max). The
// bonus is recomputed on every read, so it appears/vanishes as the minion takes
// damage or is healed — no stored state, no recompute hook.
func (mn *minion) enraged() bool {
	return !mn.silenced && mn.card.Enrage != nil && mn.health < mn.maxHP()
}

// maxHP is the minion's current maximum Health: base card health plus all
// enchantments.
func (mn *minion) maxHP() int {
	h := mn.card.Health + mn.auraHP
	for _, e := range mn.enchants {
		h += e.hp
	}
	return h
}

// attacksPerTurn is how many times the minion may attack each turn: 2 with
// Twinstrike, otherwise 1.
func (mn *minion) attacksPerTurn() int {
	if mn.has(cards.KeywordTwinstrike) {
		return 2
	}
	return 1
}

// hasAttacked reports whether the minion has used any of its attacks this turn.
func (mn *minion) hasAttacked() bool { return mn.attacksMade > 0 }

// spellDamageOf is the minion's live Spell Damage contribution (0 if silenced):
// its printed Spell Damage plus any granted by enchantments (`runeward_sage`).
func spellDamageOf(mn *minion) int {
	sp := 0
	if !mn.silenced { // silence strips the printed Spell Damage, not post-silence enchant grants
		sp = mn.card.SpellDamage
	}
	for _, e := range mn.enchants {
		sp += e.spellDamage
	}
	return sp
}

// secretInst is an active secret in a player's hidden secret zone (Phase 7). The
// opponent sees only that a secret exists (a count), never which one, until it
// triggers and is revealed.
type secretInst struct {
	uid   string
	card  cards.Card
	owner int
}

// weaponInst is the hero's equipped weapon: its card plus the current attack and
// remaining durability (both mutable so weapon-buff battlecries can raise them).
// It is destroyed when durability reaches 0 or it is replaced.
type weaponInst struct {
	card       cards.Card
	attack     int
	durability int
}

// playerState is one side's board, hand, hero, mana, secret zone, hero power, and
// weapon. frozen marks the hero as frozen (now meaningful: a frozen hero cannot
// attack with its weapon). heroPowerUsed / heroAttacked are per-turn flags.
type playerState struct {
	heroHP        int
	armor         int
	frozen        bool
	mana          int
	maxMana       int
	hand          []cards.Card
	deck          []cards.Card // draw pile (index 0 = top); shuffled at match start
	fatigue       int          // escalating self-damage taken each draw from an empty deck
	board         []*minion
	diedThisTurn  []cards.Card // base cards of this player's minions that died since the current turn began (`revenant_priestess`); cleared each turn start
	secrets       []*secretInst
	heroPower     cards.Card
	heroPowerUsed bool
	heroArt       string // overrides the class-derived hero portrait art id (`overlord_xathul` hero replacement); "" = class default
	weapon        *weaponInst
	heroAttacked  bool
	immune        bool // hero ignores all damage this turn (e.g. `frostward_aegis`)

	heroAtkThisTurn       int  // `valiant_strike`: bonus hero Attack for this turn only (no weapon needed); cleared at turn end
	minMinHealth1ThisTurn bool // `rallying_roar`: this player's minions can't drop below 1 Health this turn; cleared at turn end

	nextSecretFree        bool // `spellwarden_magus`: the next Secret played this turn costs 0 (consumed on play; cleared each turn start)
	minionsPlayedThisTurn int  // `pocket_conjurer`: drives the "first minion each turn" cost discount (reset each turn start)
	spellsFreeOnTurn      int  // `fizzle_sparkmuddle`: the turnNum during which this player's spells cost 0 (0 = never; a specific future turn)
}

// pendingChoice is a Seek awaiting the player's pick. While set, the match
// rejects that player's other actions until a Choose resolves it.
type pendingChoice struct {
	player  int
	options []cards.Card
}

// mulliganState is the opening mulligan phase (Phase 9). Both players replace
// chosen opening-hand cards simultaneously; play begins once both submit. While
// set, the match rejects all normal actions.
type mulliganState struct {
	done [2]bool
}

// Match is a 1v1 game. All mutations hold mu.
type Match struct {
	ID       string
	mu       sync.Mutex
	players  [2]Sender
	state    [2]*playerState
	turn     int // index into players: whose turn it is
	turnNum  int
	nextUID  int
	over     bool
	rng      *rand.Rand       // seeded per match; drives random-target effects
	log      []protocol.Event // ordered event log for the action in progress
	history  []protocol.Event // rolling recent event log (across actions) for reconnect replay
	pending  *pendingChoice   // a Seek awaiting a choice (blocks other actions)
	mulligan *mulliganState   // opening mulligan phase (blocks all actions until both submit)

	// castMul doubles the damage/healing of the spell or hero power currently
	// resolving (`oracle_velneth`). 0/1 = normal, 2 = doubled. Set around a
	// spell/hero-power applyEffect and cleared after; transient (never snapshotted).
	castMul int

	// observers are spectators, each bound to the seat (0/1) whose point of view
	// they watch. They receive the same per-seat snapshots that player gets (that
	// player's hand revealed, the opponent's hidden) but never act on the match.
	observers map[Sender]int

	// Turn timer: a turn auto-ends after turnDuration. turnGen guards stale timer
	// callbacks (a manual EndTurn reschedules + bumps the generation). turnEndsAt
	// feeds the per-snapshot countdown.
	turnDuration time.Duration
	turnTimer    *time.Timer
	turnEndsAt   time.Time
	turnGen      int

	// Mulligan backstop timer: force-keeps any player who hasn't submitted within
	// mulliganLimit, then begins play. Cleared when play starts (beginPlay).
	mulliganTimer *time.Timer

	// AI opponent (vs-AI matches). aiSeat is the bot's seat index, or -1 for a
	// human-vs-human match. aiRng is a planner-only RNG stream (seeded from the
	// match seed) so simulation never consumes the game's RNG. See ai.go / ai_driver.go.
	aiSeat int
	aiRng  *rand.Rand

	// Ranked stats. class[pi] is each seat's deck class (for per-class W/L); rank[pi]
	// is each seat's ladder position at match start (0 = unranked / AI), shown on the
	// in-game nameplate. ranked marks a competitive (matchmaking-queue) game whose
	// result is persisted; onEnd fires once when a hero dies, with the winning seat,
	// so the transport can record the result. resultDone guards against a double-fire.
	class      [2]cards.Class
	rank       [2]int
	ranked     bool
	onEnd      func(winnerSeat int)
	resultDone bool

	// startGame[pi] is the "Start of Game" card that actually fired for seat pi
	// (duskwarden_genmar / lunar_devourer), or nil. Revealed center-stage to both
	// players once the mulligan ends (beginPlay) so they understand the Hero Power change.
	startGame [2]*cards.Card
}

// New creates a match between two players. Player at index 0 goes first. seed
// makes the match's RNG deterministic for tests. deckA/deckB are the players'
// decks; each is shuffled, opening hands are dealt, and the match opens in the
// mulligan phase. A deck shorter than the opening hand still works (fatigue
// arrives sooner).
func New(id string, a, b Sender, seed int64, deckA, deckB []cards.Card) *Match {
	m := &Match{
		ID:      id,
		players: [2]Sender{a, b},
		state: [2]*playerState{
			{heroHP: startHP, deck: append([]cards.Card(nil), deckA...), heroPower: cards.HeroPowerForClass(cards.DeckClass(deckA))},
			{heroHP: startHP, deck: append([]cards.Card(nil), deckB...), heroPower: cards.HeroPowerForClass(cards.DeckClass(deckB))},
		},
		rng:          rand.New(rand.NewSource(seed)),
		mulligan:     &mulliganState{},
		turnDuration: turnLimit,
		observers:    make(map[Sender]int),
		aiSeat:       -1, // human-vs-human unless EnableAI is called
		class:        [2]cards.Class{cards.DeckClass(deckA), cards.DeckClass(deckB)},
	}
	// Start-of-Game passives (deck-construction gated) resolve off the full deck,
	// before it is shuffled and the opening hand is dealt.
	m.applyStartOfGame(0, deckA)
	m.applyStartOfGame(1, deckB)
	m.shuffleDeck(0)
	m.shuffleDeck(1)
	// Deal opening hands directly (not via drawCard — the opener never fatigues
	// and never overdraws). Mana Surge is granted to player 1 once mulligan ends.
	m.dealOpening(0, openingFirst)
	m.dealOpening(1, openingSecond)
	return m
}

// applyStartOfGame resolves the deck-construction "Start of Game" passives for
// player pi from their full deck: duskwarden_genmar (all even-cost → Hero Power
// costs 1) and lunar_devourer (all odd-cost → upgraded Hero Power). The passive
// only fires when its own card is in the deck and the parity holds.
func (m *Match) applyStartOfGame(pi int, deck []cards.Card) {
	find := func(id string) *cards.Card {
		for i := range deck {
			if deck[i].ID == id {
				return &deck[i]
			}
		}
		return nil
	}
	allParity := func(odd bool) bool {
		if len(deck) == 0 {
			return false
		}
		for _, c := range deck {
			if (c.Cost%2 == 1) != odd {
				return false
			}
		}
		return true
	}
	ps := m.state[pi]
	if c := find("duskwarden_genmar"); c != nil && allParity(false) {
		ps.heroPower.Cost = 1
		m.startGame[pi] = c
	}
	if c := find("lunar_devourer"); c != nil && allParity(true) {
		ps.heroPower = cards.UpgradedMageHeroPower()
		m.startGame[pi] = c
	}
}

// emitStartOfGame reveals each seat's fired "Start of Game" card center-stage so
// both players see why a Hero Power changed. Emitted in seat order; the client
// reorders per-viewer (own card first) and plays them one after another with a
// gap. No-op when neither player triggered one.
func (m *Match) emitStartOfGame() {
	for pi := 0; pi < 2; pi++ {
		if c := m.startGame[pi]; c != nil {
			cv := cardView(*c)
			m.emit(protocol.Event{Kind: "startgame", Source: m.pid(pi), Name: c.Name, Card: &cv})
		}
	}
}

// Start sends the initial snapshot to both players. The match opens in the
// mulligan phase (no turn has begun); play starts once both players submit a
// mulligan.
func (m *Match) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, p := range m.players {
		ms := protocol.MatchStart{
			Type:     protocol.TypeMatchStart,
			Players:  []string{m.players[0].ID(), m.players[1].ID()},
			Turn:     m.players[m.turn].ID(),
			You:      p.ID(),
			Self:     m.selfView(i, m.players[i].Name()),
			Opp:      m.oppView(1-i, m.players[1-i].Name()),
			Mulligan: true,
		}
		p.Send(protocol.Marshal(ms))
	}
	m.scheduleMulliganTimer()
}

// scheduleMulliganTimer arms the mulligan backstop. Caller holds m.mu.
func (m *Match) scheduleMulliganTimer() {
	if m.mulligan == nil {
		return
	}
	m.mulliganTimer = time.AfterFunc(mulliganLimit, m.onMulliganTimeout)
}

// stopMulliganTimer cancels the pending mulligan backstop. Caller holds m.mu.
func (m *Match) stopMulliganTimer() {
	if m.mulliganTimer != nil {
		m.mulliganTimer.Stop()
		m.mulliganTimer = nil
	}
}

// onMulliganTimeout force-keeps the hand of anyone who never submitted, then
// starts play. Idempotent: a manual submit that already began play clears
// m.mulligan, so this no-ops.
func (m *Match) onMulliganTimeout() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.over || m.mulligan == nil {
		return
	}
	m.mulligan.done[0] = true
	m.mulligan.done[1] = true
	m.beginPlay()
}

// beginPlay ends the mulligan phase: grants Mana Surge to the second player, runs
// the first player's opening turn (mana + draw), and broadcasts the first
// real snapshot. Caller holds m.mu.
func (m *Match) beginPlay() {
	m.stopMulliganTimer()
	m.mulligan = nil
	if len(m.state[1].hand) < maxHand {
		m.state[1].hand = append(m.state[1].hand, cards.ManaSurge())
	}
	m.resetLog()
	m.emitStartOfGame()
	m.startTurn(m.turn) // first player: ramp mana + draw turn-1 card
	m.finish()          // win-check (covers turn-1 fatigue edge) + snapshot
	if m.isAITurn() {
		go m.runBotTurn(m.turn) // bot holds the opening turn (only if seated first)
	}
}

// EndTurn passes the turn if c is the current player. Returns (false, reason)
// when rejected without state change.
func (m *Match) EndTurn(c Sender) (bool, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.over {
		return false, "game over"
	}
	if m.mulligan != nil {
		return false, "mulligan in progress"
	}
	if m.pending != nil {
		return false, "finish seeking first"
	}
	if m.players[m.turn].ID() != c.ID() {
		return false, "not your turn"
	}
	m.endTurnLocked()
	return true, ""
}

// endTurnLocked passes the turn (thaw → flip → start the next turn → win-check +
// snapshot). Caller holds m.mu and has already validated. Used by both the manual
// EndTurn and the turn-timer auto-end.
func (m *Match) endTurnLocked() {
	m.resetLog()
	m.fireTriggers(m.turn, cards.OnTurnEnd, nil)    // end-of-turn triggers for the ending player
	m.fireTriggers(m.turn, cards.OnAnyTurnEnd, nil) // global end-of-EACH-turn triggers (`cragmaw`); reactors are both boards
	m.returnTempControl(m.turn)                     // `gloom_thrall`: temporarily-controlled minions go home
	m.thawAfterTurn(m.turn)                         // thaw the ending player's frozen characters
	m.clearTempBuffs()                              // expire "this turn" buffs
	m.state[0].immune = false                       // hero immunity (frostward_aegis) lasts only "this turn"
	m.state[1].immune = false
	m.state[0].heroAtkThisTurn = 0 // `valiant_strike`'s hero Attack lasts only "this turn"
	m.state[1].heroAtkThisTurn = 0
	m.state[0].minMinHealth1ThisTurn = false // `rallying_roar`'s damage floor lasts only "this turn"
	m.state[1].minMinHealth1ThisTurn = false
	m.turn = 1 - m.turn
	m.turnNum++
	m.startTurn(m.turn)
	m.finish() // win-check (turn-start fatigue can kill) + snapshot
	if m.isAITurn() {
		go m.runBotTurn(m.turn) // drive the bot's turn off-thread (it re-locks m.mu)
	}
}

// scheduleTurnTimer (re)arms the auto-end timer for the current turn. Caller
// holds m.mu. Bumping turnGen invalidates any previously-scheduled callback.
func (m *Match) scheduleTurnTimer() {
	m.stopTurnTimer()
	d := m.activeTurnDuration()
	if d <= 0 || m.over {
		return
	}
	m.turnGen++
	gen := m.turnGen
	m.turnEndsAt = time.Now().Add(d)
	m.turnTimer = time.AfterFunc(d, func() { m.onTurnTimeout(gen) })
}

// activeTurnDuration is the turn length to schedule: the base turn limit, shortened
// to the smallest TurnSeconds of any in-play, non-silenced minion (`chronlord_zhal` caps
// turns to 15s). Returns 0 (timers disabled) if the base is disabled, so a global
// rule-changer never arms a timer that was intentionally off.
func (m *Match) activeTurnDuration() time.Duration {
	d := m.turnDuration
	if d <= 0 {
		return d
	}
	for pi := 0; pi < 2; pi++ {
		for _, mn := range m.state[pi].board {
			if mn.silenced || mn.card.TurnSeconds <= 0 {
				continue
			}
			if cand := time.Duration(mn.card.TurnSeconds) * time.Second; cand < d {
				d = cand
			}
		}
	}
	return d
}

// stopTurnTimer cancels any pending auto-end and clears the deadline. Caller
// holds m.mu.
func (m *Match) stopTurnTimer() {
	if m.turnTimer != nil {
		m.turnTimer.Stop()
		m.turnTimer = nil
	}
	m.turnEndsAt = time.Time{}
}

// onTurnTimeout auto-ends the active turn when its timer fires. gen guards
// against a stale timer (one whose turn already ended manually). A pending
// Seek is auto-resolved (first option) so the turn can pass.
func (m *Match) onTurnTimeout(gen int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.over || m.mulligan != nil || gen != m.turnGen {
		return
	}
	if m.pending != nil {
		ch := m.pending.options[0]
		ps := m.state[m.pending.player]
		if len(ps.hand) < maxHand {
			ps.hand = append(ps.hand, ch) // auto-pick; not named in the shared log
		}
		m.pending = nil
	}
	m.endTurnLocked()
}

// turnSecondsLeft is the whole seconds remaining in the current turn (0 if no
// turn is active). Caller holds m.mu.
func (m *Match) turnSecondsLeft() int {
	if m.turnEndsAt.IsZero() {
		return 0
	}
	d := time.Until(m.turnEndsAt)
	if d <= 0 {
		return 0
	}
	return int(d.Seconds()) + 1
}
