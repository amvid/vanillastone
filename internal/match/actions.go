package match

import (
	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/protocol"
)

// PlayCard plays the card at handIndex, appending a played minion to the board.
// Convenience wrapper over PlayCardAt for callers that don't choose a position.
func (m *Match) PlayCard(c Sender, handIndex int, targetID string) (bool, string) {
	return m.PlayCardAt(c, handIndex, targetID, -1)
}

// PlayCardAt plays the card at handIndex for the current player. Minions are
// summoned (inserted at board position pos, or appended when pos < 0); spells
// resolve their effect against targetID (ignored for minions and untargeted
// spells). All validation happens before any mutation so a rejected play leaves
// state untouched.
func (m *Match) PlayCardAt(c Sender, handIndex int, targetID string, pos int) (bool, string) {
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
	pi := m.indexOf(c)
	if pi != m.turn {
		return false, "not your turn"
	}
	ps := m.state[pi]
	if handIndex < 0 || handIndex >= len(ps.hand) {
		return false, "no such card"
	}
	card := ps.hand[handIndex]
	cost := m.effectiveCost(pi, card) // cost modifiers (auras / `spellwarden_magus` / `tidecolossus`)
	if ps.mana < cost {
		return false, "not enough mana"
	}

	if card.Type == cards.TypeSpell {
		m.resetLog()
		return m.playSpell(pi, handIndex, card, targetID, cost)
	}

	if card.Type == cards.TypeSecret {
		m.resetLog()
		return m.playSecret(pi, handIndex, card, cost)
	}

	if card.Type == cards.TypeWeapon {
		m.resetLog()
		ps.hand = append(ps.hand[:handIndex], ps.hand[handIndex+1:]...)
		ps.mana -= cost
		ps.weapon = &weaponInst{card: card, attack: card.Attack, durability: card.Durability} // replaces any current weapon
		m.emitPlay(pi, card)
		m.emit(protocol.Event{Kind: "equip", Source: m.pid(pi), Name: card.Name})
		m.fireTriggers(pi, cards.OnPlayCard, nil)
		m.finish()
		return true, ""
	}

	if len(ps.board) >= maxBoard {
		return false, "board full"
	}

	// Onset (on_play) validation happens before any mutation. A onset
	// that needs a target requires a legal one when targets exist; with no legal
	// targets it fizzles and the minion still plays (standard behavior).
	bc := card.Onset()
	applyBC := bc != nil
	var bcRef charRef
	if bc != nil && needsTarget(bc.Target) {
		if m.hasLegalTargetFor(pi, bc) {
			ref, ok := m.resolveTarget(pi, targetID)
			if !ok {
				return false, "no such target"
			}
			if !validTarget(bc.Target, ref, pi) || !targetCondOK(bc, ref) {
				return false, "illegal target"
			}
			bcRef = ref
		} else {
			applyBC = false // fizzle (no legal target for a conditional onset)
		}
	}

	// Validation complete — mutate.
	m.resetLog()
	ps.hand = append(ps.hand[:handIndex], ps.hand[handIndex+1:]...)
	ps.mana -= cost
	ps.minionsPlayedThisTurn++ // after cost is locked, so `pocket_conjurer` discounts THIS minion
	m.emitPlay(pi, card)
	mn := m.summonMinion(pi, card)
	m.placeAt(pi, mn, pos) // honor a drag-to-position drop (no-op if pos < 0)
	// A onset that refers to the minion itself anchors on the minion just
	// played: adjacency (`bannerguard`/`wardstone_sentinel`) resolves off its final board slot, a
	// TargetSelf effect (self-damage / hand-count self-buff) aims back at it, and
	// EffectMissiles excludes it from the random spread.
	if applyBC && !needsTarget(bc.Target) &&
		(bc.Target == cards.TargetSelf || bc.Kind == cards.EffectMissiles ||
			bc.Kind == cards.EffectConsumeShields || bc.Kind == cards.EffectGainWeaponAttack ||
			bc.Kind == cards.EffectTransformRandom ||
			bc.Area == cards.AreaAdjacent || bc.Area == cards.AreaSplash || bc.Area == cards.AreaFriendlyTribe ||
			bc.Area == cards.AreaOtherMinions) {
		bcRef = charRef{minion: mn, owner: pi}
	}
	// The opponent's "enemy plays a minion" secrets (e.g. Mimic) trigger as the
	// minion enters play, before its onset resolves.
	m.triggerSecrets(1-pi, cards.OnEnemyPlayMinion, secretCtx{minion: mn})
	if applyBC {
		m.emit(protocol.Event{Kind: "onset", Source: mn.uid, Name: card.Name})
		if bc.Kind == cards.EffectSeek {
			// Seek pauses the action: send the prompt and the snapshot so far,
			// then wait for Choose to finish (resolve deaths / win / state).
			m.startSeek(pi, bc.Pool)
			return true, ""
		}
		if bc.Kind == cards.EffectCopy {
			// `visage_thief`-style: the just-played minion becomes a copy of the target.
			// Needs the source minion (mn), so it is handled here rather than in
			// applyEffect (which only knows the target ref).
			if bcRef.minion != nil {
				m.copyMinion(mn, bcRef.minion)
			}
		} else {
			m.applyEffect(pi, bc, bcRef, 0, mn.uid) // battlecries are not boosted by Spell Damage
		}
		// Self-buff rider (`shellback_crab`: destroy a Gilkin AND gain +2/+2): applied to
		// the played minion after the main effect, only when the onset resolved.
		if mn != nil && (bc.SelfBuffAtk > 0 || bc.SelfBuffHP > 0) {
			mn.enchants = append(mn.enchants, enchant{atk: bc.SelfBuffAtk, hp: bc.SelfBuffHP})
			mn.health += bc.SelfBuffHP
			m.emit(protocol.Event{Kind: "buff", Target: mn.uid, BuffAtk: bc.SelfBuffAtk, BuffHP: bc.SelfBuffHP})
		}
		// Additional on_play triggers beyond the primary onset (e.g. a second
		// "each player draws" / "give both players" rider — `brineseer_diviner`,
		// `warhorn_chieftain`). These are untargeted by construction; fire each in
		// declared order. A future targeted extra would need its own resolution.
		for _, extra := range card.TriggersFor(cards.OnPlay)[1:] {
			e := extra
			if needsTarget(e.Target) {
				continue
			}
			m.applyEffect(pi, &e, charRef{}, 0, mn.uid)
		}
	}
	m.fireTriggers(pi, cards.OnPlayCard, mn) // "after you play a card" (the minion itself doesn't count)
	m.finish()
	return true, ""
}

// playSpell validates the target, then spends mana, discards the card, applies
// the effect, and resolves deaths/win. Caller holds m.mu and has already
// checked turn and mana.
func (m *Match) playSpell(pi, handIndex int, card cards.Card, targetID string, cost int) (bool, string) {
	eff := card.Effect
	var ref charRef
	if eff.Target != cards.TargetNone {
		var ok bool
		ref, ok = m.resolveTarget(pi, targetID)
		if !ok {
			return false, "no such target"
		}
		if !validTarget(eff.Target, ref, pi) {
			return false, "illegal target"
		}
		if !spellTargetable(ref) {
			return false, "can't target an Elusive minion"
		}
	}
	ps := m.state[pi]
	ps.hand = append(ps.hand[:handIndex], ps.hand[handIndex+1:]...)
	ps.mana -= cost
	m.emitPlay(pi, card)
	// The opponent's "enemy casts a spell" secrets trigger before the effect. A
	// counter cancels it (card + mana still spent); a retarget secret (`decoy_ward`)
	// may replace `ref` with a summoned decoy, so the effect resolves against that.
	if m.triggerSecrets(1-pi, cards.OnEnemyCastSpell, secretCtx{spellTarget: &ref, spellName: card.Name}) {
		// The spell was cast (then countered): cast-synergy still fires.
		m.copySpellToOpponent(pi, card)
		m.fireTriggers(pi, cards.OnSpellCast, nil)
		m.fireTriggers(pi, cards.OnPlayCard, nil)
		m.finish()
		return true, ""
	}
	m.applyEffect(pi, eff, ref, m.spellPower(pi), m.pid(pi))
	m.copySpellToOpponent(pi, card)
	m.fireTriggers(pi, cards.OnSpellCast, nil)
	m.fireTriggers(pi, cards.OnPlayCard, nil)
	m.finish()
	return true, ""
}

// copySpellToOpponent implements `archivist_solenne`: for each non-silenced minion in
// play with CopiesSpells, add a copy of the just-cast spell to the NON-caster's
// hand (burn if full). Identity is hidden, so the event carries no name.
func (m *Match) copySpellToOpponent(caster int, card cards.Card) {
	opp := 1 - caster
	for pi := 0; pi < 2; pi++ {
		for _, mn := range m.state[pi].board {
			if mn.silenced || !mn.card.CopiesSpells {
				continue
			}
			if len(m.state[opp].hand) >= maxHand {
				m.emit(protocol.Event{Kind: "burn", Target: m.pid(opp)})
				continue
			}
			m.state[opp].hand = append(m.state[opp].hand, card)
			m.emit(protocol.Event{Kind: "generate", Target: m.pid(opp)})
		}
	}
}

// Attack orders attackerID to attack targetID ("hero" or an opponent minion).
func (m *Match) Attack(c Sender, attackerID, targetID string) (bool, string) {
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
	pi := m.indexOf(c)
	if pi != m.turn {
		return false, "not your turn"
	}
	ps, opp := m.state[pi], m.state[1-pi]
	m.refreshAuras() // attack/eligibility reads must see current aura buffs

	if attackerID == selfHeroTarget {
		return m.heroAttack(pi, targetID)
	}

	atk := findMinion(ps.board, attackerID)
	if atk == nil {
		return false, "no such attacker"
	}
	if atk.atk() <= 0 {
		return false, "minion has no attack"
	}
	if atk.frozen {
		return false, "minion is frozen"
	}
	if !m.canAttack(atk) {
		return false, "minion cannot attack"
	}

	if targetID == oppHeroTarget {
		if !m.canAttackHero(atk) {
			return false, "can't attack heroes this turn"
		}
		if hasTaunt(opp) {
			return false, "must attack a Taunt minion"
		}
		m.resetLog()
		m.emit(protocol.Event{Kind: "attack", Source: atk.uid, Target: m.pid(1 - pi)})
		// Hero-attack secrets fire before damage. Snare ("enemy minion attacks your
		// hero") may destroy the attacker and cancel; Frost Ward ("when your hero is
		// attacked") gains armor and never cancels. Both fire on a minion attack.
		cancelled := m.triggerSecrets(1-pi, cards.OnEnemyAttackHero, secretCtx{minion: atk})
		m.triggerSecrets(1-pi, cards.OnHeroAttacked, secretCtx{minion: atk})
		if cancelled {
			atk.attacksMade++
			atk.stealthed = false
			m.finish()
			return true, ""
		}
		dealt := m.damageHero(1-pi, atk.atk(), atk.uid)
		if atk.has(cards.KeywordLifesteal) {
			m.lifestealHeal(pi, dealt)
		}
		if dealt > 0 && atk.has(cards.KeywordFreezeOnHit) {
			m.freezeHero(1 - pi) // `frostfont_elemental`: Freeze the hero it damages
		}
	} else {
		tgt := findMinion(opp.board, targetID)
		if tgt == nil {
			return false, "no such target"
		}
		if tgt.stealthed {
			return false, "can't attack a Stealthed minion"
		}
		if hasTaunt(opp) && !tgt.has(cards.KeywordTaunt) {
			return false, "must attack a Taunt minion"
		}
		m.resetLog()
		m.emit(protocol.Event{Kind: "attack", Source: atk.uid, Target: tgt.uid})
		// Simultaneous damage exchange (Aegis handled in damageMinion).
		m.combatStrike(atk, tgt) // attacker -> defender
		if tgt.atk() > 0 {
			m.combatStrike(tgt, atk) // defender retaliates
		}
	}

	atk.attacksMade++
	atk.stealthed = false // attacking breaks Stealth
	m.finish()
	return true, ""
}

// heroAttack resolves a weapon-armed hero attack by player pi against targetID.
// Caller holds m.mu, has verified turn/pending, and refreshed auras. A hero
// attack does NOT trigger hero-attack secrets (Snare is minion-specific). The
// hero takes retaliation from a struck minion; the weapon loses one durability.
func (m *Match) heroAttack(pi int, targetID string) (bool, string) {
	ps, opp := m.state[pi], m.state[1-pi]
	if heroAttackValue(ps) <= 0 {
		return false, "no weapon to attack with"
	}
	if ps.frozen {
		return false, "hero is frozen"
	}
	if ps.heroAttacked {
		return false, "hero already attacked"
	}
	atkVal := heroAttackValue(ps)
	if targetID == oppHeroTarget {
		if hasTaunt(opp) {
			return false, "must attack a Taunt minion"
		}
		m.resetLog()
		m.emit(protocol.Event{Kind: "attack", Source: m.pid(pi), Target: m.pid(1 - pi)})
		// "When your hero is attacked" secrets (e.g. Frost Ward) fire on a weapon
		// attack too — before damage, so the armor absorbs this hit. Snare does NOT
		// (it is minion-specific).
		m.triggerSecrets(1-pi, cards.OnHeroAttacked, secretCtx{})
		m.damageHero(1-pi, atkVal, m.pid(pi))
	} else {
		tgt := findMinion(opp.board, targetID)
		if tgt == nil {
			return false, "no such target"
		}
		if tgt.stealthed {
			return false, "can't attack a Stealthed minion"
		}
		if hasTaunt(opp) && !tgt.has(cards.KeywordTaunt) {
			return false, "must attack a Taunt minion"
		}
		m.resetLog()
		m.emit(protocol.Event{Kind: "attack", Source: m.pid(pi), Target: tgt.uid})
		m.damageMinion(tgt, atkVal, m.pid(pi)) // Aegis handled in damageMinion
		if tgt.atk() > 0 {
			m.damageHero(pi, tgt.atk(), tgt.uid) // the struck minion hits back at my hero
		}
	}
	ps.heroAttacked = true
	m.useWeaponDurability(pi)
	m.finish()
	return true, ""
}

// useWeaponDurability spends one durability of player pi's weapon, destroying it
// at zero.
func (m *Match) useWeaponDurability(pi int) {
	w := m.state[pi].weapon
	if w == nil {
		return
	}
	w.durability--
	if w.durability <= 0 {
		m.emit(protocol.Event{Kind: "weaponBreak", Source: m.pid(pi), Name: w.card.Name})
		m.state[pi].weapon = nil
	}
}

// HeroPower uses player c's hero power (Mage Fire Dart: 2 mana, 1 damage) against
// targetID, once per turn. Hero power damage is not boosted by Spell Damage.
func (m *Match) HeroPower(c Sender, targetID string) (bool, string) {
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
	pi := m.indexOf(c)
	if pi != m.turn {
		return false, "not your turn"
	}
	ps := m.state[pi]
	if ps.heroPowerUsed {
		return false, "hero power already used"
	}
	hp := ps.heroPower
	if ps.mana < hp.Cost {
		return false, "not enough mana"
	}
	eff := hp.Effect
	var ref charRef
	if eff.Target != cards.TargetNone {
		var ok bool
		ref, ok = m.resolveTarget(pi, targetID)
		if !ok {
			return false, "no such target"
		}
		if !validTarget(eff.Target, ref, pi) {
			return false, "illegal target"
		}
		if !spellTargetable(ref) {
			return false, "can't target an Elusive minion"
		}
	}
	m.resetLog()
	ps.mana -= hp.Cost
	ps.heroPowerUsed = true
	m.emit(protocol.Event{Kind: "heropower", Source: m.pid(pi), Name: hp.Name})
	m.applyEffect(pi, eff, ref, 0, m.pid(pi)) // hero power is not boosted by Spell Damage
	m.finish()
	return true, ""
}

// heroAttackValue is the hero's current attack: its weapon's attack plus any
// damaged-minion bonus (`grudge_smith`: +2 weapon Attack while it is damaged).
// Zero with no weapon (the weapon bonus is inert without a weapon).
func heroAttackValue(ps *playerState) int {
	if ps.weapon == nil {
		return 0
	}
	atk := ps.weapon.attack
	for _, mn := range ps.board {
		if !mn.silenced && mn.card.EnrageWeaponAtk > 0 && mn.health < mn.maxHP() {
			atk += mn.card.EnrageWeaponAtk
		}
	}
	return atk
}

// heroCanAttack reports whether the hero is able to attack right now: it has a
// weapon with attack, has not attacked this turn, and is not frozen.
func heroCanAttack(ps *playerState) bool {
	return heroAttackValue(ps) > 0 && !ps.heroAttacked && !ps.frozen
}

// Concede forfeits the match for c: its hero is set to 0 and the opponent wins.
// Valid on either player's turn. Returns (false, reason) when rejected.
func (m *Match) Concede(c Sender) (bool, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.over {
		return false, "game over"
	}
	m.resetLog()
	m.pending = nil  // abandon any in-flight Seek
	m.mulligan = nil // abandon the mulligan phase if a player bails during it
	m.state[m.indexOf(c)].heroHP = 0
	m.finish()
	return true, ""
}

// Choose resolves a pending Seek for player c by picking option index. The
// chosen card is added to the player's hand (or burned if the hand is full),
// then the paused action finishes. Returns (false, reason) when rejected.
func (m *Match) Choose(c Sender, index int) (bool, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.over {
		return false, "game over"
	}
	if m.pending == nil {
		return false, "nothing to choose"
	}
	pi := m.indexOf(c)
	if m.pending.player != pi {
		return false, "not your choice"
	}
	if index < 0 || index >= len(m.pending.options) {
		return false, "no such option"
	}
	chosen := m.pending.options[index]
	m.pending = nil
	m.resetLog()
	ps := m.state[pi]
	if len(ps.hand) < maxHand {
		ps.hand = append(ps.hand, chosen)
	}
	// Note: the chosen card is intentionally NOT named in the (shared) event log —
	// the opponent must not learn which card was seeked.
	m.finish()
	return true, ""
}

// Mulligan submits player c's opening-hand replacement choices: indices points
// at the opening-hand cards to toss. Tossed cards are replaced with fresh draws,
// then shuffled back into the deck. Both players mulligan independently; play
// begins once both have submitted. Returns (false, reason) when rejected.
func (m *Match) Mulligan(c Sender, indices []int) (bool, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.over {
		return false, "game over"
	}
	if m.mulligan == nil {
		return false, "not in mulligan"
	}
	pi := m.indexOf(c)
	if m.mulligan.done[pi] {
		return false, "already mulliganed"
	}
	ps := m.state[pi]
	toss := make(map[int]bool, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(ps.hand) {
			return false, "invalid mulligan index"
		}
		if toss[idx] {
			return false, "duplicate mulligan index"
		}
		toss[idx] = true
	}
	var kept, tossed []cards.Card
	for i, card := range ps.hand {
		if toss[i] {
			tossed = append(tossed, card)
		} else {
			kept = append(kept, card)
		}
	}
	// Draw replacements off the top BEFORE returning the tossed cards, so a
	// tossed card cannot be immediately redrawn; then shuffle them back in.
	n := len(tossed)
	if n > len(ps.deck) {
		n = len(ps.deck)
	}
	kept = append(kept, ps.deck[:n]...)
	ps.deck = append(ps.deck[n:], tossed...)
	m.shuffleDeck(pi)
	ps.hand = kept
	m.mulligan.done[pi] = true
	if m.mulligan.done[0] && m.mulligan.done[1] {
		m.beginPlay()
	} else {
		m.resetLog()
		m.sendStateTo(pi) // confirm to the submitter; opponent learns nothing
	}
	return true, ""
}
