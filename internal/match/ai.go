package match

import "github.com/amvid/vanillastone/internal/cards"

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

// Planner scoring extremes + the lethal-lens magnitudes.
const (
	winScore       = 1e6  // resulting state kills the opponent — take it above all else
	loseScore      = -1e6 // resulting state kills me — never choose a suicidal line
	dangerWeight   = 30.0 // flat penalty for ending in range of the opponent's next-turn lethal
	overkillWeight = 3.0  // extra penalty per point their next-turn burst exceeds my health
	epsilon        = 0.01 // a move must beat "do nothing" by at least this to be worth it
	maxBotMoves    = 40   // hard per-turn action bound (planner loop backstop)
)

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
