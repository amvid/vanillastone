package match

import (
	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/protocol"
)

// resetLog clears the event log at the start of an action.
func (m *Match) resetLog() { m.log = nil }

// emit appends an event to the in-progress action's log and to the rolling
// history (kept across actions, capped, so a reconnecting player can replay
// what happened — including events that resolved while they were disconnected).
func (m *Match) emit(e protocol.Event) {
	m.log = append(m.log, e)
	m.history = append(m.history, e)
	if len(m.history) > maxHistory {
		m.history = m.history[len(m.history)-maxHistory:]
	}
}

// pid returns the player id for an index (used as the absolute id of a hero in
// event Source/Target).
func (m *Match) pid(i int) string { return m.players[i].ID() }

// needsTarget reports whether a targeting rule requires a player-chosen target
// at play time (vs. resolving server-side).
func needsTarget(rule cards.TargetRule) bool {
	switch rule {
	case cards.TargetAny, cards.TargetMinion, cards.TargetFriendlyMinion, cards.TargetEnemyMinion,
		cards.TargetEnemy, cards.TargetFriendlyHero, cards.TargetHero:
		return true
	}
	return false
}

// summonMinion puts a fresh minion (summon-sick) on owner's board and emits a
// summon event. If the board is already full (maxBoard) the summon is discarded
// and nil is returned — this is the single place the board cap is enforced for
// every summon path (onset, finalGasp, token summon, Mimic copy). Playing
// a minion from hand checks the cap earlier so the play is rejected, not burned.
func (m *Match) summonMinion(owner int, c cards.Card) *minion {
	if len(m.state[owner].board) >= maxBoard {
		return nil
	}
	m.nextUID++
	mn := &minion{
		uid:              uid(m.nextUID),
		card:             c,
		owner:            owner,
		health:           c.Health,
		summonedThisTurn: true, // summon sickness (Charge/Rush bypass it)
		aegis:            c.Has(cards.KeywordAegis),
		stealthed:        c.Has(cards.KeywordStealth),
	}
	m.state[owner].board = append(m.state[owner].board, mn)
	m.emit(protocol.Event{Kind: "summon", Source: m.pid(owner), Target: mn.uid, Name: c.Name})
	// Other friendly minions react to the summon (e.g. a per-summon ping).
	m.fireTriggers(owner, cards.OnFriendlySummon, mn)
	return mn
}

// placeAt moves a just-summoned minion (currently at the end of owner's board)
// to board position pos, so a player can drag a card between minions on the
// table. pos < 0 (or nil from the wire) means "leave it appended". pos is
// clamped to a valid slot. No-op if mn is nil (board was full).
func (m *Match) placeAt(owner int, mn *minion, pos int) {
	if mn == nil || pos < 0 {
		return
	}
	b := m.state[owner].board
	last := len(b) - 1
	if last <= 0 || b[last] != mn {
		return // nothing to reorder (only minion, or not the one we just added)
	}
	if pos >= last {
		return // already at the end
	}
	copy(b[pos+1:], b[pos:last]) // shift [pos,last) right by one
	b[pos] = mn
}

// playSecret puts a secret into the caster's hidden secret zone. Rejected if the
// zone is full or the same secret is already active (standard rules). Caller
// holds m.mu and has already checked turn and mana.
func (m *Match) playSecret(pi, handIndex int, card cards.Card, cost int) (bool, string) {
	ps := m.state[pi]
	if len(ps.secrets) >= maxSecrets {
		return false, "too many secrets"
	}
	for _, s := range ps.secrets {
		if s.card.ID == card.ID {
			return false, "secret already active"
		}
	}
	ps.hand = append(ps.hand[:handIndex], ps.hand[handIndex+1:]...)
	ps.mana -= cost
	ps.nextSecretFree = false // `spellwarden_magus`'s free-secret discount is consumed by this play
	// A secret is a spell, so casting it triggers the opponent's "enemy casts a
	// spell" secrets (e.g. Nullify). If countered, the secret never enters play —
	// the card and mana are still spent.
	if m.triggerSecrets(1-pi, cards.OnEnemyCastSpell, secretCtx{}) {
		// A secret is a spell: casting it fires the controller's cast-synergy even
		// when the secret is itself countered. (A countered secret never enters play,
		// so on-secret-played does NOT fire.)
		m.fireTriggers(pi, cards.OnSpellCast, nil)
		m.fireTriggers(pi, cards.OnPlayCard, nil)
		m.finish()
		return true, ""
	}
	m.nextUID++
	ps.secrets = append(ps.secrets, &secretInst{uid: uid(m.nextUID), card: card, owner: pi})
	// No reveal event for the placed secret: the opponent only sees the count rise.
	m.fireTriggers(pi, cards.OnSpellCast, nil)
	m.fireTriggers(pi, cards.OnSecretPlayed, nil) // global: "whenever a Secret is played"
	m.fireTriggers(pi, cards.OnPlayCard, nil)
	m.finish()
	return true, ""
}

// secretCtx carries the context a triggering secret may need: the attacking or
// played minion (kind-dependent), or — for an enemy spell cast on a minion — a
// pointer to the spell's chosen target, which a retarget secret (`decoy_ward`) may
// replace in place so the spell resolves against the summoned decoy instead.
type secretCtx struct {
	minion      *minion
	spellTarget *charRef
	spellName   string   // the enemy spell being cast — named in a Counter Spell reveal
	redirect    *charRef // SecretRetargetAttack (`feint_trap`): the attack's new target is written here
	didRedirect *bool    // set true when a retarget secret fired, so the caller re-aims the blow
}

// secretFires reports whether secret s should fire given the context, beyond its
// trigger event already matching. `decoy_ward` only fires when the enemy spell
// targets one of the owner's (defender's) minions; every other secret always
// fires on its event.
func (m *Match) secretFires(defender int, s *secretInst, ctx secretCtx) bool {
	if s.card.Secret.Kind == cards.SecretRetargetSpell {
		return ctx.spellTarget != nil && ctx.spellTarget.minion != nil && ctx.spellTarget.owner == defender
	}
	return true
}

// triggerSecrets fires player `defender`'s secrets matching event ev (in play
// order), removing each and emitting a reveal. It returns true if the triggering
// action should be cancelled — an interrupt secret fired (Counter Spell or
// Destroy Attacker). Caller holds m.mu.
func (m *Match) triggerSecrets(defender int, ev cards.EventType, ctx secretCtx) bool {
	ps := m.state[defender]
	if len(ps.secrets) == 0 {
		return false
	}
	cancelled := false
	kept := ps.secrets[:0]
	for _, s := range ps.secrets {
		if s.card.Secret == nil || s.card.Secret.Trigger != ev || !m.secretFires(defender, s, ctx) {
			kept = append(kept, s)
			continue
		}
		// A triggered secret is no longer hidden — reveal the card to both players.
		// A Counter Spell names the spell it negated (the log otherwise only shows
		// the secret, not what it stopped).
		scv := cardView(s.card)
		reveal := protocol.Event{Kind: "secret", Source: m.pid(defender), Name: s.card.Name, Card: &scv}
		if s.card.Secret.Kind == cards.SecretCounterSpell {
			reveal.Note = ctx.spellName
		}
		m.emit(reveal)
		// `hawkeye_bow`: the wielder's weapon gains +1 Durability whenever one of their
		// Secrets is revealed. The secret owner (defender) is also the weapon's wielder.
		if w := m.state[defender].weapon; w != nil && w.card.WeaponSecretGain {
			w.durability++
		}
		switch s.card.Secret.Kind {
		case cards.SecretDestroyAttacker:
			if ctx.minion != nil {
				ctx.minion.health = 0 // destroyed (resolved in finish); no hero damage
			}
			cancelled = true
		case cards.SecretCounterSpell:
			cancelled = true
		case cards.SecretGainArmor:
			m.gainArmor(defender, s.card.Secret.Amount) // does not cancel the attack
		case cards.SecretCopyMinion:
			if ctx.minion != nil {
				m.summonMinion(defender, ctx.minion.card) // base copy; nil if board full
			}
		case cards.SecretRetargetSpell:
			// Summon the decoy and redirect the in-flight spell onto it (does not
			// cancel). secretFires already guaranteed a minion target owned by defender.
			if tok, ok := cards.Get(s.card.Secret.Summon); ok {
				if decoy := m.summonMinion(defender, tok); decoy != nil {
					*ctx.spellTarget = charRef{minion: decoy, owner: defender}
				}
			}
		case cards.SecretDamageAll:
			// `blasting_snare`: deal Amount to every enemy character (the attacker's side).
			// Does not cancel; the attacker may die to this before its blow lands.
			amt := s.card.Secret.Amount
			m.damageHero(1-defender, amt, m.pid(defender))
			for _, mn := range append([]*minion(nil), m.state[1-defender].board...) {
				m.damageMinion(mn, amt, m.pid(defender))
			}
		case cards.SecretBounceAttacker:
			// `snaring_trap`: return the attacker to its owner's hand, costing Amount more.
			// Cancels the attack.
			if ctx.minion != nil {
				m.bounceMinionCost(ctx.minion, ctx.minion.owner, s.card.Secret.Amount)
			}
			cancelled = true
		case cards.SecretDamageMinion:
			// `marksman_trap`: deal Amount to the minion the enemy just played. Non-cancel.
			if ctx.minion != nil {
				m.damageMinion(ctx.minion, s.card.Secret.Amount, m.pid(defender))
			}
		case cards.SecretRetargetAttack:
			// `feint_trap`: redirect the attack to a random OTHER character (not the
			// original hero target, not the attacker itself). Non-cancel.
			if ctx.minion != nil && ctx.redirect != nil && ctx.didRedirect != nil {
				if pick, ok := m.randomRetarget(defender, ctx.minion); ok {
					*ctx.redirect = pick
					*ctx.didRedirect = true
				}
			}
		case cards.SecretSummon:
			// `serpent_trap`: summon Amount copies of Summon for the owner. Non-cancel.
			if tok, ok := cards.Get(s.card.Secret.Summon); ok {
				n := s.card.Secret.Amount
				if n < 1 {
					n = 1
				}
				for i := 0; i < n; i++ {
					if m.summonMinion(defender, tok) == nil {
						break
					}
				}
			}
		}
		// secret consumed (not kept)
	}
	ps.secrets = kept
	return cancelled
}

// fireTriggers fires the in-play minions' edge triggers reacting to event `when`.
// controller is the player whose action raised the event; subject is the minion
// the event concerns (the one summoned or that died), or nil. "Other"-scoped
// events (friendly summon/death) skip the subject itself. Each reacting minion's
// effect resolves from its own perspective (caster = its owner); a TargetSelf
// effect (e.g. a self-buff) is aimed back at that minion. Silenced minions do
// not react. Iterates a snapshot so summons created mid-dispatch don't re-fire.
func (m *Match) fireTriggers(controller int, when cards.EventType, subject *minion) {
	// A triggered minion effect is never a spell/hero-power cast, so the cast-output
	// doubler (`oracle_velneth`) must not apply to it. Suspend castMul for the dispatch
	// and restore it afterwards (the in-flight spell's Then chain still doubles).
	savedMul := m.castMul
	m.castMul = 0
	defer func() { m.castMul = savedMul }()
	var reactors []*minion
	switch when {
	case cards.OnAnyMinionDeath, cards.OnHeal, cards.OnMinionHealed, cards.OnSecretPlayed, cards.OnAnyTurnEnd, cards.OnAnyMinionDamage:
		reactors = append(reactors, m.state[0].board...) // global events: both boards react
		reactors = append(reactors, m.state[1].board...)
	default:
		reactors = append(reactors, m.state[controller].board...)
	}
	for _, mn := range reactors {
		if mn == subject || mn.silenced {
			continue
		}
		for _, t := range mn.card.Triggers {
			if t.When != when || !m.condMet(t.Condition, mn.owner) {
				continue
			}
			// Tribe-gated summon/death triggers (e.g. "when you summon a Gilkin")
			// fire only when the subject minion is of the named tribe.
			if t.SubjectTribe != cards.TribeNone && (subject == nil || subject.card.Tribe != t.SubjectTribe) {
				continue
			}
			// Attack-gated summon triggers (`battle_marshal`: "a minion with 3 or less
			// Attack") fire only when the subject's Attack is at or below the cap.
			if t.SubjectMaxAttack > 0 && (subject == nil || subject.atk() > t.SubjectMaxAttack) {
				continue
			}
			// Probabilistic triggers (`lucky_angler`'s 50% draw): roll before firing.
			if t.Chance > 0 && m.rng.Intn(100) >= t.Chance {
				continue
			}
			e := t.Effect
			var ref charRef
			switch e.Target {
			case cards.TargetSelf:
				ref = charRef{minion: mn, owner: mn.owner}
			case cards.TargetSubject:
				if subject == nil {
					continue // `battle_marshal`: grant the keyword to the summoned minion
				}
				ref = charRef{minion: subject, owner: subject.owner}
			case cards.TargetRandomFriendly:
				pick := m.randomFriendlyExcept(mn.owner, mn)
				if pick == nil {
					continue // no other friendly minion to affect
				}
				ref = charRef{minion: pick, owner: mn.owner}
			case cards.TargetFriendlyHero:
				ref = charRef{owner: mn.owner} // controller's own hero (e.g. cog_mender end-of-turn heal)
			}
			// Self-anchored area effects (e.g. `cinder_baron`'s "all OTHER characters")
			// resolve their neighbours/exclusion off the reacting minion.
			if e.Area == cards.AreaAdjacent || e.Area == cards.AreaSplash || e.Area == cards.AreaOtherCharacters {
				ref = charRef{minion: mn, owner: mn.owner}
			}
			// Mark which minion reacted (so the client can pulse it + log the cause)
			// before its effect's own events resolve.
			m.emit(protocol.Event{Kind: "trigger", Source: mn.uid, Name: mn.card.Name})
			m.applyEffect(mn.owner, &e, ref, 0, mn.uid)
		}
	}
}

// condMet reports whether an edge trigger's condition holds for controller pi.
func (m *Match) condMet(cond cards.TriggerCondition, pi int) bool {
	switch cond {
	case cards.CondControlSecret:
		return len(m.state[pi].secrets) > 0
	default: // CondNone
		return true
	}
}

// startSeek presents three cards for player pi to pick one, records the pending
// choice, pushes the snapshot so far, and sends the Seek prompt. The pool is the
// whole collection of eff.Pool's type, OR — for eff.FromDeck (`tracking`) — the
// top 3 cards of the caster's own deck (removed now; the two unpicked are
// discarded). Returns false without pausing when there is nothing to offer (e.g.
// an empty deck); the caller must then finish the action itself. The action
// otherwise resumes (and finishes) when Choose arrives. Caller holds m.mu.
func (m *Match) startSeek(pi int, eff *cards.Effect) bool {
	var opts []cards.Card
	if eff.FromDeck {
		ps := m.state[pi]
		n := 3
		if n > len(ps.deck) {
			n = len(ps.deck)
		}
		opts = append([]cards.Card(nil), ps.deck[:n]...)
		ps.deck = ps.deck[n:] // the 3 leave the deck; the unpicked are discarded
	} else {
		ids := m.pickDistinct(cards.SeekPoolIDs(eff.Pool), 3)
		for _, id := range ids {
			if c, ok := cards.Get(id); ok {
				opts = append(opts, c)
			}
		}
	}
	if len(opts) == 0 {
		return false
	}
	m.pending = &pendingChoice{player: pi, options: opts}
	m.sendStateAll() // reflect the summon/onset resolved so far
	m.sendSeekTo(pi)
	// Tell the opponent the player is seeking so their client can show the
	// pick happening (hidden faces, no full-screen modal) instead of a frozen board.
	od := protocol.Marshal(protocol.OppSeek{Type: protocol.TypeOppSeek, Count: len(opts)})
	m.players[1-pi].Send(od)
	// Spectators of EITHER seat see the hidden-face indicator too — never the real
	// Seek prompt (only the choosing player may pick).
	m.fanout(pi, od)
	m.fanout(1-pi, od)
	return true
}

// sendSeekTo sends player pi the prompt for the current pending Seek.
// Caller holds m.mu and must have m.pending set for pi. Split out so a
// reconnecting player (Reattach) can be re-shown a Seek they had paused on.
func (m *Match) sendSeekTo(pi int) {
	views := make([]protocol.CardView, len(m.pending.options))
	for i, c := range m.pending.options {
		views[i] = cardView(c)
	}
	m.players[pi].Send(protocol.Marshal(protocol.Seek{Type: protocol.TypeSeek, Options: views}))
}

// pickDistinct returns up to n distinct ids chosen at random (via the match RNG)
// from ids, without mutating the input.
func (m *Match) pickDistinct(ids []string, n int) []string {
	cp := append([]string(nil), ids...)
	m.rng.Shuffle(len(cp), func(i, j int) { cp[i], cp[j] = cp[j], cp[i] })
	if n > len(cp) {
		n = len(cp)
	}
	return cp[:n]
}

// broadcast sends the same bytes to both players and every spectator (only for
// hidden-info-free messages like game_over).
func (m *Match) broadcast(b []byte) {
	for _, p := range m.players {
		p.Send(b)
	}
	for obs := range m.observers {
		obs.Send(b)
	}
}
