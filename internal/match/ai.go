package match

import (
	"sort"

	"github.com/amvid/vanillastone/internal/cards"
)

// AI opponent: a single-turn greedy planner over a board-evaluation heuristic.
// No game-tree search and no model — on its turn it repeatedly picks the legal
// action that most improves eval() (simulated on a clone), applies it for real,
// and stops when nothing helps, then ends the turn. eval() is threat-aware and a
// "lethal lens" (burstNow / burstNextTurn) re-weights toward racing or defending,
// so the bot trades to survive instead of blindly attacking face. The clean seam
// for a future "hard mode" is 2-ply lookahead on the same clone.

// Eval weights. Hand-tuned, deliberately simple; the comments say what each one
// trades off so they can be retuned without re-reading the math. Scale reference:
// a vanilla 3/4 minion is worth ~atk+health+presence = 8.
const (
	heroHPWeight   = 1.0 // each point of hero health/armor; ~1:1 with a point of face damage
	minionPresence = 1.0 // flat value of having a body on board (tempo / board control)
	handWeight     = 0.5 // each card in hand (card advantage — minor, never dominates board)
	weaponWeight   = 0.5 // per point of weapon attack×durability
	secretValue    = 1.5 // per armed secret (a card that does something on a future trigger)

	// Asymmetric threat: the opponent's total board attack is what can kill us, so
	// it counts EXTRA from our POV (on top of being part of their minions' value).
	// This is what nudges the planner to trade into big attackers rather than race.
	threatWeight = 0.5

	// Keyword bonuses added to a minion's atk+health body value.
	tauntBonus       = 1.5 // protects the hero / other minions
	aegisBonus       = 2.0 // soaks the next hit (roughly a second life)
	twinstrikeFactor = 0.5 // ×atk: a second swing each turn
	poisonousBonus   = 2.0 // trades up into anything
	lifestealBonus   = 1.0 // incidental healing
	stealthBonus     = 1.0 // can't be targeted until it swings

	// Flat worth of having any unsilenced Final Gasp (deathrattle), on top of a
	// summon Final Gasp's token-body value. Skews the bot toward silencing/removing
	// such minions without dying into them.
	finalGaspBaseBonus = 1.0
)

// eval scores the position from seat's point of view: our value minus the
// opponent's, minus an extra threat term for the opponent's board attack. Higher
// is better for seat. The planner maximizes this.
func (m *Match) eval(seat int) float64 {
	m.refreshAuras() // value reads must see current aura buffs
	me := m.sideValue(seat)
	opp := m.sideValue(1 - seat)
	threat := float64(m.boardAttack(1-seat)) * threatWeight
	return me - opp - threat
}

// sideValue is the total heuristic worth of one seat: hero, board, hand, weapon.
func (m *Match) sideValue(seat int) float64 {
	ps := m.state[seat]
	v := float64(ps.heroHP+ps.armor) * heroHPWeight
	for _, mn := range ps.board {
		v += m.minionValue(mn)
	}
	v += float64(min(len(ps.hand), maxHand)) * handWeight
	if ps.weapon != nil {
		v += float64(ps.weapon.attack*ps.weapon.durability) * weaponWeight
	}
	v += float64(len(ps.secrets)) * secretValue
	return v
}

// minionValue is a single minion's worth: current attack + CURRENT health (a
// damaged minion is worth less than a full one) + presence + keyword bonuses.
func (m *Match) minionValue(mn *minion) float64 {
	v := float64(mn.atk()+mn.health) + minionPresence
	if mn.has(cards.KeywordTaunt) {
		v += tauntBonus
	}
	if mn.aegis {
		v += aegisBonus
	}
	if mn.has(cards.KeywordTwinstrike) {
		v += float64(mn.atk()) * twinstrikeFactor
	}
	if mn.has(cards.KeywordPoisonous) {
		v += poisonousBonus
	}
	if mn.has(cards.KeywordLifesteal) {
		v += lifestealBonus
	}
	if mn.stealthed {
		v += stealthBonus
	}
	v += finalGaspValue(mn)
	return v
}

// finalGaspValue estimates the standing worth of a minion's unsilenced Final Gasp
// to its controller: the payoff its death will hand them. Added to the body value
// so the bot (a) prices its own deathrattle minions correctly and (b) is willing
// to SILENCE / transform a nasty enemy deathrattle minion instead of only ever
// trading into it. (Trading into one is already priced correctly by simulation —
// the death and its Final Gasp resolve on the clone — so this term only captures
// the value of the as-yet-unfired effect.) Summons are valued by the token's body;
// every other Final Gasp gets a modest flat bonus (deathrattles skew net-positive
// for their owner, and the rare exceptions are corrected by simulation on a kill).
func finalGaspValue(mn *minion) float64 {
	if mn.silenced {
		return 0
	}
	var v float64
	for _, eff := range mn.card.FinalGasps() {
		v += finalGaspBaseBonus
		if eff.Kind == cards.EffectSummon {
			if tok, ok := cards.Get(eff.Summon); ok {
				n := eff.Count
				if n < 1 {
					n = 1
				}
				v += float64(n) * (float64(tok.Attack+tok.Health) + minionPresence)
			}
		}
	}
	return v
}

// boardAttack is the sum of attack across a seat's board (its raw offensive
// potential, used by the threat term and the burst estimates).
func (m *Match) boardAttack(seat int) int {
	total := 0
	for _, mn := range m.state[seat].board {
		total += mn.atk()
	}
	return total
}

// hasTurnStartBoardWipe reports whether a live minion will destroy all minions at
// its controller's next turn start (the ruin_oracle / Doomsayer pattern). Silence
// strips its triggers, so a silenced one is harmless.
func hasTurnStartBoardWipe(mn *minion) bool {
	if mn.silenced {
		return false
	}
	for _, t := range mn.card.Triggers {
		if t.When == cards.OnTurnStart && t.Effect.Kind == cards.EffectDestroy && t.Effect.Area == cards.AreaAllMinions {
			return true
		}
	}
	return false
}

// oppBoardWipePending reports whether the opponent of seat has a live minion that
// wipes all minions at the opponent's next turn start.
func (m *Match) oppBoardWipePending(seat int) bool {
	for _, mn := range m.state[1-seat].board {
		if hasTurnStartBoardWipe(mn) {
			return true
		}
	}
	return false
}

// netBoardValue is seat's board minion value minus the opponent's — the slice of
// eval that a board wipe would erase.
func (m *Match) netBoardValue(seat int) float64 {
	var v float64
	for _, mn := range m.state[seat].board {
		v += m.minionValue(mn)
	}
	for _, mn := range m.state[1-seat].board {
		v -= m.minionValue(mn)
	}
	return v
}

// burstNow is the max face damage seat can deal THIS turn, ignoring enemy Taunt
// (the planner clears Taunt itself before the lethal swing). It sums every
// ready-to-attack minion, the hero's weapon swing, and a damaging hero power the
// seat can currently afford. Used to spot "I have lethal — stop trading, race".
func (m *Match) burstNow(seat int) int {
	m.refreshAuras()
	ps := m.state[seat]
	dmg := 0
	for _, mn := range ps.board {
		if m.canAttackHero(mn) {
			dmg += mn.atk()
		}
	}
	if !ps.frozen && !ps.heroAttacked {
		dmg += heroAttackValue(ps)
	}
	if !ps.heroPowerUsed && ps.mana >= m.effectiveCost(seat, ps.heroPower) {
		dmg += heroPowerFace(ps.heroPower)
	}
	return dmg
}

// burstNextTurn estimates the face damage seat could deal on its NEXT turn from
// the current board: every minion can swing (none is summon-sick next turn), plus
// weapon and an (assumed affordable) damaging hero power. Used to detect incoming
// lethal so the planner defends this turn instead of racing into death.
func (m *Match) burstNextTurn(seat int) int {
	m.refreshAuras()
	ps := m.state[seat]
	dmg := m.boardAttack(seat)
	dmg += heroAttackValue(ps)
	dmg += heroPowerFace(ps.heroPower) // next turn there's almost always mana for it
	return dmg
}

// heroPowerFace is the direct face damage a hero power can throw (0 if it isn't a
// damaging power). Mage's is Deal 1 (2 upgraded).
func heroPowerFace(hp cards.Card) int {
	if hp.Effect != nil && hp.Effect.Kind == cards.EffectDamage {
		return hp.Effect.Amount
	}
	return 0
}

// heroPowerDraws reports whether using the hero power draws a card (the Warlock
// Life Tap pattern, `soul_tithe`: damage self + draw).
func heroPowerDraws(hp cards.Card) bool {
	return hp.Effect != nil && (hp.Effect.ThenDraw > 0 || hp.Effect.Kind == cards.EffectDraw)
}

// heroPowerSelfDamage is the damage a hero power deals to its own hero (Life Tap's
// 2). Zero for powers that don't hit the friendly hero.
func heroPowerSelfDamage(hp cards.Card) int {
	if hp.Effect != nil && hp.Effect.Kind == cards.EffectDamage && hp.Effect.Area == cards.AreaFriendlyHero {
		return hp.Effect.Amount
	}
	return 0
}

// Planner scoring extremes + the lethal-lens magnitudes.
const (
	winScore       = 1e6  // resulting state kills the opponent — take it above all else
	loseScore      = -1e6 // resulting state kills me — never choose a suicidal line
	dangerWeight   = 30.0 // flat penalty for ending in range of the opponent's next-turn lethal
	overkillWeight = 3.0  // extra penalty per point their next-turn burst exceeds my health
	epsilon        = 0.01 // a move must beat "do nothing" by at least this to be worth it
	maxBotMoves    = 40   // hard per-turn action bound (planner loop backstop)

	// A live enemy minion that destroys all minions at its controller's next turn
	// start (the ruin_oracle / Doomsayer pattern) will erase the board before the
	// opponent acts. We discount the net board value it would wipe by this factor —
	// not the full value, because our own minions still get one attack in first.
	wipeDiscount = 0.7

	// Life Tap policy: the lowest hero HP the bot will leave itself at AFTER paying
	// the self-damage to draw a card. Above this it's safe to tap for cards when
	// nothing else improves the board; below it, keep the health.
	lifeTapMinHP = 10

	// 2-ply lookahead: the deep planner only spends its opponent-turn simulation on
	// the top-scoring shallow candidates, to bound cost. Each deep eval runs a full
	// simulated opponent turn, so this caps the work per bot action.
	lookaheadTopK = 6
)

// aiLookahead enables the 2-ply planner (planBestDeep): each candidate is scored
// after a simulated opponent reply, so the bot avoids plays the opponent simply
// punishes (a minion that dies next turn, bodies dumped into a board wipe). A var
// so tests/perf can disable it and fall back to the 1-ply planBest.
var aiLookahead = true

// aiMove kinds.
const (
	mPlay   = "play"
	mAttack = "attack"
	mPower  = "power"
)

// aiMove is one candidate action the planner can simulate and apply. Only the
// fields relevant to its kind are set.
type aiMove struct {
	kind     string
	hand     int    // mPlay: hand index
	attacker string // mAttack: attacker uid
	target   string // play/power/attack target id ("" = no target)
}

// applyTo runs the move through the same authoritative action handlers a human
// uses, on match m as seat. Returns the handler's (ok, reason) — illegal moves
// are simply rejected, which is how the planner filters its broad candidate list.
func (mv aiMove) applyTo(m *Match, seat int) (bool, string) {
	c := m.players[seat]
	switch mv.kind {
	case mPlay:
		return m.PlayCardAt(c, mv.hand, mv.target, -1)
	case mAttack:
		return m.Attack(c, mv.attacker, mv.target)
	case mPower:
		return m.HeroPower(c, mv.target)
	}
	return false, "unknown move"
}

// aiCandidates lists every action worth trying from the current position: play
// each hand card (untargeted, plus against every character if it needs a target),
// attack with each ready minion (into each enemy minion or the hero), and the
// hero power. The list is intentionally over-broad — illegal entries are dropped
// when applyTo rejects them on the clone. Caller holds m.mu.
func (m *Match) aiCandidates(seat int) []aiMove {
	ps, opp := m.state[seat], m.state[1-seat]
	chars := charTargetIDs(m)
	var moves []aiMove

	for i, card := range ps.hand {
		if isManaRamp(card) && !m.manaRampUnlocksPlay(seat, i) {
			continue // don't burn the Coin with nothing to spend the extra mana on
		}
		moves = append(moves, aiMove{kind: mPlay, hand: i})
		if cardNeedsTarget(card) {
			for _, t := range chars {
				moves = append(moves, aiMove{kind: mPlay, hand: i, target: t})
			}
		}
	}

	m.refreshAuras()
	for _, a := range ps.board {
		if !m.canAttack(a) {
			continue
		}
		for _, d := range opp.board {
			moves = append(moves, aiMove{kind: mAttack, attacker: a.uid, target: d.uid})
		}
		moves = append(moves, aiMove{kind: mAttack, attacker: a.uid, target: oppHeroTarget})
	}

	// The hero attacks too when it has a weapon (or granted attack) — into a minion
	// or the face. Without this the bot would equip weapons and never swing them.
	if heroCanAttack(ps) {
		for _, d := range opp.board {
			moves = append(moves, aiMove{kind: mAttack, attacker: selfHeroTarget, target: d.uid})
		}
		moves = append(moves, aiMove{kind: mAttack, attacker: selfHeroTarget, target: oppHeroTarget})
	}

	if !ps.heroPowerUsed && ps.mana >= m.effectiveCost(seat, ps.heroPower) {
		moves = append(moves, aiMove{kind: mPower})
		if ps.heroPower.Effect != nil && ps.heroPower.Effect.Target != cards.TargetNone {
			for _, t := range chars {
				moves = append(moves, aiMove{kind: mPower, target: t})
			}
		}
	}
	return moves
}

// charTargetIDs is every targetable character id: both heroes (relative to the
// acting seat) and all minion uids on both boards. resolveTarget rejects the
// illegal ones for a given effect.
func charTargetIDs(m *Match) []string {
	ids := []string{selfHeroTarget, oppHeroTarget}
	for pi := 0; pi < 2; pi++ {
		for _, mn := range m.state[pi].board {
			ids = append(ids, mn.uid)
		}
	}
	return ids
}

// isManaRamp reports whether a card's whole job is to grant the caster temporary
// mana this turn (the Coin, `mana_surge`). Playing it is pure ramp — worthless on
// its own (and any incidental spell-cast synergy isn't worth the wasted mana),
// so it's only worth playing when the extra mana buys another play.
func isManaRamp(c cards.Card) bool {
	return c.Type == cards.TypeSpell && c.Effect != nil && c.Effect.Kind == cards.EffectMana
}

// manaRampUnlocksPlay reports whether playing the mana-ramp card at hand index
// rampIdx would let the seat afford a DIFFERENT card in hand that it can't afford
// at its current mana — i.e., the ramp actually buys a play this turn rather than
// burning the card for nothing. Caller holds m.mu.
func (m *Match) manaRampUnlocksPlay(seat, rampIdx int) bool {
	ps := m.state[seat]
	ramp := ps.hand[rampIdx]
	cur := ps.mana
	after := cur + ramp.Effect.Amount
	for i, c := range ps.hand {
		if i == rampIdx {
			continue
		}
		if cost := m.effectiveCost(seat, c); cost > cur && cost <= after {
			return true // c becomes affordable only thanks to the ramp
		}
	}
	return false
}

// cardNeedsTarget reports whether playing the card requires a target id: a
// targeted spell, or a minion with a targeted onset.
func cardNeedsTarget(c cards.Card) bool {
	if c.Type == cards.TypeSpell && c.Effect != nil && needsTarget(c.Effect.Target) {
		return true
	}
	if bc := c.Onset(); bc != nil && needsTarget(bc.Target) {
		return true
	}
	return false
}

// planBestDeep is the 2-ply planner the live bot uses. It shortlists the most
// promising moves by the cheap 1-ply score, then for each looks a full opponent
// turn ahead (deepScore) and keeps the move with the best position AFTER the
// opponent's reply. That reply lens is what lets the bot decline plays the
// opponent simply punishes — a minion it'll just kill next turn, or bodies dumped
// into a board wipe that clears them. Beats the deep "do nothing" baseline by
// epsilon or it ends the turn. Falls back to the 1-ply planBest when lookahead is
// off. Caller holds m.mu; simulation never touches the live match.
func (m *Match) planBestDeep(seat int) (aiMove, bool) {
	if !aiLookahead {
		return m.planBest(seat)
	}
	// Survival override: when the opponent already has lethal on us next turn, the
	// opponent-reply sim concludes "I die in every line" and flattens all deepScores
	// to a loss — so nothing beats passing and the bot gives up (no trade, no hero
	// power). The shallow planner's lethal lens instead fights: it rewards trading
	// down the threat (shrinking the opponent's burst) and racing. Defer to it.
	if m.facingLethalNextTurn(seat) {
		return m.planBest(seat)
	}
	// One shared seed for EVERY deep sim this turn (baseline + each candidate). The
	// opponent's reply — its draw, its random effects — is then identical across
	// candidates, so a score difference reflects the candidate move, not draw luck.
	// With independent seeds the reply's draw variance (several eval points) swamps
	// small clean signals (e.g. a hero power worth +2 face), and the bot passes on a
	// coin flip — the "does nothing" bug.
	seed := m.aiRng.Int63()
	best := m.deepScoreAfter(seat, aiMove{}, seed, true) // baseline: pass and let them reply
	var chosen aiMove
	found := false
	for _, mv := range m.topCandidates(seat, lookaheadTopK, seed) {
		if sc := m.deepScoreAfter(seat, mv, seed, false); sc > best+epsilon {
			best, chosen, found = sc, mv, true
		}
	}
	return chosen, found
}

// deepScoreAfter applies mv on a clone seeded with `seed` (or applies nothing when
// pass is true) and deep-scores the result. Threading one seed through every call
// in a turn makes the simulated opponent reply identical across candidates, so the
// scores are comparable. Returns loseScore for a move that's illegal on the clone.
func (m *Match) deepScoreAfter(seat int, mv aiMove, seed int64, pass bool) float64 {
	sim := m.cloneForSim(seed)
	sim.state[1-seat].secrets = nil // fog of war: hide the opponent's secrets (see planBest)
	if !pass {
		if ok, _ := mv.applyTo(sim, seat); !ok {
			return loseScore
		}
		sim.autoResolveSeek(seat)
	}
	return sim.deepScore(seat)
}

// topCandidates returns up to k candidate moves ranked by their cheap 1-ply score,
// best first — the shortlist the deep planner spends its lookahead budget on, so a
// full opponent turn is simulated only a bounded number of times per action. Uses
// the shared turn seed so the shortlist is deterministic alongside the deep eval.
func (m *Match) topCandidates(seat, k int, seed int64) []aiMove {
	type scored struct {
		mv aiMove
		sc float64
	}
	var ranked []scored
	for _, mv := range m.aiCandidates(seat) {
		sim := m.cloneForSim(seed)
		sim.state[1-seat].secrets = nil
		if ok, _ := mv.applyTo(sim, seat); !ok {
			continue
		}
		sim.autoResolveSeek(seat)
		ranked = append(ranked, scored{mv, sim.scoreForPlanner(seat)})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].sc > ranked[j].sc })
	if len(ranked) > k {
		ranked = ranked[:k]
	}
	out := make([]aiMove, len(ranked))
	for i, s := range ranked {
		out[i] = s.mv
	}
	return out
}

// facingLethalNextTurn reports whether the opponent's current board (plus weapon
// and hero power) could kill seat's hero next turn — the trigger for the deep
// planner to hand off to the shallow survival planner.
func (m *Match) facingLethalNextTurn(seat int) bool {
	return m.burstNextTurn(1-seat) >= m.state[seat].heroHP+m.state[seat].armor
}

// deepScore evaluates a position one ply ahead from seat's POV: it hands the turn
// to the opponent (whose turn-start triggers — e.g. a board wipe — fire on the
// handoff), lets a shallow greedy opponent play out its whole turn, then scores
// the result. So a body the opponent just kills, or minions wiped before they do
// anything, are worth no more than not having played them. `m` is a sim clone with
// the bot's candidate already applied (and any Seek resolved).
func (m *Match) deepScore(seat int) float64 {
	opp := 1 - seat
	// Already decided by the candidate itself — no opponent turn to simulate.
	if m.over || m.state[opp].heroHP <= 0 || m.state[seat].heroHP <= 0 {
		return m.scoreForPlanner(seat)
	}
	m.endTurnLocked()     // hand off; opponent's turn-start triggers fire here
	m.runShallowTurn(opp) // opponent takes a greedy turn
	return m.scoreForPlanner(seat)
}

// runShallowTurn plays a full greedy turn for seat on a simulation clone using the
// cheap 1-ply planner (NOT planBestDeep — that would recurse), resolving any Seek
// the turn opens. Used by deepScore to model the opponent's reply. `m` is a sim;
// callers do not hold m.mu (the clone is single-threaded).
func (m *Match) runShallowTurn(seat int) {
	for i := 0; i < maxBotMoves; i++ {
		if m.over || m.turn != seat {
			return
		}
		if m.pending != nil && m.pending.player == seat {
			m.Choose(m.players[seat], bestSeekIndex(m.pending.options))
			continue
		}
		mv, ok := m.planBest(seat)
		if !ok {
			return
		}
		if applied, _ := mv.applyTo(m, seat); !applied {
			return
		}
	}
}

// planBest simulates every candidate on a fresh clone and returns the one whose
// resulting position scores highest, provided it beats doing nothing. Returns
// found=false when no action improves the position (→ end the turn). Caller holds
// m.mu. Simulation never touches the live match.
func (m *Match) planBest(seat int) (aiMove, bool) {
	best := m.scoreForPlanner(seat) // "do nothing" baseline
	var chosen aiMove
	found := false
	for _, mv := range m.aiCandidates(seat) {
		sim := m.cloneForSim(m.aiRng.Int63())
		// Fog of war: the opponent's secrets are hidden. Strip them from the planning
		// clone so the bot can't "see" them by watching them fire in simulation — else
		// it dodges every move that would trip one (won't attack into a punisher secret,
		// won't play minions into an enemy-play secret) and stalls on hero power alone.
		sim.state[1-seat].secrets = nil
		if ok, _ := mv.applyTo(sim, seat); !ok {
			continue
		}
		sim.autoResolveSeek(seat) // settle any seek the move opened, so the state is scorable
		if sc := sim.scoreForPlanner(seat); sc > best+epsilon {
			best, chosen, found = sc, mv, true
		}
	}
	return chosen, found
}

// botFallbackHeroPower returns a hero-power move to use when the planner found no
// value-improving action — currently the Warlock Life Tap (draw + self damage).
// Tapping never improves the board heuristic (it trades 2 HP for a hidden card),
// so the planner never picks it, yet refilling the hand is the right idle play.
// Only taps when safe: HP stays at/above lifeTapMinHP after the self-damage, the
// hand has room for the drawn card, and we're not already in next-turn lethal
// range. Returns found=false otherwise. Caller holds m.mu.
func (m *Match) botFallbackHeroPower(seat int) (aiMove, bool) {
	ps := m.state[seat]
	if ps.heroPowerUsed || ps.mana < m.effectiveCost(seat, ps.heroPower) {
		return aiMove{}, false
	}
	if !heroPowerDraws(ps.heroPower) {
		return aiMove{}, false
	}
	if ps.heroHP-heroPowerSelfDamage(ps.heroPower) < lifeTapMinHP {
		return aiMove{}, false
	}
	if len(ps.hand) >= maxHand {
		return aiMove{}, false // the drawn card would burn
	}
	if m.burstNextTurn(1-seat) >= ps.heroHP+ps.armor {
		return aiMove{}, false // don't spend health into incoming lethal
	}
	return aiMove{kind: mPower}, true
}

// scoreForPlanner is eval() plus the lethal lens: terminal scores for a decided
// game, a big penalty for ending in range of the opponent's next-turn lethal
// (drives defensive trades), evaluated from seat's POV on this (possibly clone)
// state.
func (m *Match) scoreForPlanner(seat int) float64 {
	opp := 1 - seat
	if m.state[opp].heroHP <= 0 {
		return winScore
	}
	if m.state[seat].heroHP <= 0 {
		return loseScore
	}
	s := m.eval(seat)
	// If I can kill the opponent THIS turn, race — their next turn never comes, so
	// don't let the danger term pull me into defensive trades instead of the kill.
	myHP := m.state[seat].heroHP + m.state[seat].armor
	oppHP := m.state[opp].heroHP + m.state[opp].armor
	if m.burstNow(seat) >= oppHP {
		return s
	}
	// A pending enemy turn-start board wipe (ruin_oracle) erases all minions before
	// the opponent acts. Discount the net board value it would destroy: when I'm
	// ahead on board the wipe hurts me, so this dampens over-committing bodies into
	// it and rewards killing the wiper; when I'm behind the wipe favors me, so I hold
	// minions and won't waste removal on it.
	if m.oppBoardWipePending(seat) {
		s -= wipeDiscount * m.netBoardValue(seat)
	}
	// Else, if the opponent could kill me next turn, this line is dangerous. The
	// penalty is GRADED in their overkill (not a flat all-or-nothing): each
	// defensive trade that shaves their burst improves the score, so the planner
	// chains trades down to safety instead of stopping at the first one and facing.
	if next := m.burstNextTurn(opp); next >= myHP {
		s -= dangerWeight + float64(next-myHP+1)*overkillWeight
	}
	return s
}

// autoResolveSeek settles a Seek that a simulated onset opened on the
// clone (it blocks further play), picking the first option so the position can be
// scored. The live bot turn resolves seeks deliberately (botChoose).
func (m *Match) autoResolveSeek(seat int) {
	if m.pending != nil && m.pending.player == seat {
		m.Choose(m.players[seat], 0)
	}
}
