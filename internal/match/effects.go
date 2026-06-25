package match

import (
	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/protocol"
)

// charRef points at a character a spell can affect: a minion (minion != nil) or
// a hero (minion == nil). owner is the player index that controls it.
type charRef struct {
	minion *minion
	owner  int
}

// resolveTarget maps a targetID, from the perspective of player pi, to a
// character. Returns false if no such character exists.
func (m *Match) resolveTarget(pi int, targetID string) (charRef, bool) {
	switch targetID {
	case selfHeroTarget:
		return charRef{owner: pi}, true
	case oppHeroTarget:
		return charRef{owner: 1 - pi}, true
	}
	if mn := findMinion(m.state[pi].board, targetID); mn != nil {
		return charRef{minion: mn, owner: pi}, true
	}
	if mn := findMinion(m.state[1-pi].board, targetID); mn != nil {
		return charRef{minion: mn, owner: 1 - pi}, true
	}
	return charRef{}, false
}

// validTarget reports whether ref satisfies the spell's targeting rule, from
// the perspective of caster pi.
// spellTargetable reports whether ref may be targeted by a spell or hero power.
// An Elusive minion cannot be (by either player); battlecries and attacks ignore
// this.
func spellTargetable(ref charRef) bool {
	return ref.minion == nil || !ref.minion.has(cards.KeywordElusive)
}

func validTarget(rule cards.TargetRule, ref charRef, pi int) bool {
	// An enemy Stealthed minion can never be targeted by the opponent.
	if ref.minion != nil && ref.owner != pi && ref.minion.stealthed {
		return false
	}
	switch rule {
	case cards.TargetAny:
		return true
	case cards.TargetMinion:
		return ref.minion != nil
	case cards.TargetFriendlyMinion:
		return ref.minion != nil && ref.owner == pi
	case cards.TargetEnemyMinion:
		return ref.minion != nil && ref.owner == 1-pi
	case cards.TargetEnemy:
		return ref.owner == 1-pi // any enemy character (minion or hero)
	case cards.TargetFriendlyHero:
		return ref.minion == nil && ref.owner == pi
	case cards.TargetHero:
		return ref.minion == nil // either hero, not a minion
	default:
		return false
	}
}

// targetCondOK reports whether ref satisfies an effect's extra target conditions
// (Trophy Hunter: Attack >= ReqAttack; Grave Knight: must have Taunt). Heroes
// never satisfy a minion-stat condition. Effects with no condition always pass.
func targetCondOK(eff *cards.Effect, ref charRef) bool {
	if eff.ReqAttack > 0 && (ref.minion == nil || ref.minion.atk() < eff.ReqAttack) {
		return false
	}
	if eff.ReqTaunt && (ref.minion == nil || !ref.minion.has(cards.KeywordTaunt)) {
		return false
	}
	if eff.ReqTribe != cards.TribeNone && (ref.minion == nil || ref.minion.card.Tribe != eff.ReqTribe) {
		return false
	}
	return true
}

// hasLegalTargetFor reports whether any character is a legal target for caster pi
// under the effect's targeting rule AND its extra conditions. Drives onset
// fizzle (a conditional onset with no valid target still plays the minion).
func (m *Match) hasLegalTargetFor(pi int, eff *cards.Effect) bool {
	// Battlecries ignore Elusive (it only guards spells/hero powers), so this does
	// not call spellTargetable — validTarget already bars enemy Stealth.
	check := func(ref charRef) bool {
		return validTarget(eff.Target, ref, pi) && targetCondOK(eff, ref)
	}
	if check(charRef{owner: pi}) || check(charRef{owner: 1 - pi}) {
		return true
	}
	for pj := 0; pj < 2; pj++ {
		for _, mn := range m.state[pj].board {
			if check(charRef{minion: mn, owner: pj}) {
				return true
			}
		}
	}
	return false
}

// copyMinion turns dst into a copy of src: same card and persistent enchantments,
// fresh full health (its buffed max), copied shield/stealth/silence state. dst
// keeps its own uid/owner/board slot and is summon-sick (a freshly-summoned copy).
// Used by `visage_thief`-style battlecries. Caller resolves auras afterwards.
func (m *Match) copyMinion(dst, src *minion) {
	dst.card = src.card
	dst.enchants = append([]enchant(nil), src.enchants...)
	dst.auraAtk = 0
	dst.health = src.maxHP()
	dst.summonedThisTurn = true
	dst.frozen = false
	dst.silenced = src.silenced
	dst.aegis = src.aegis
	dst.stealthed = src.has(cards.KeywordStealth)
	dst.attacksMade = 0
	m.emit(protocol.Event{Kind: "transform", Target: dst.uid, Name: src.card.Name})
}

// transformMinion replaces mn in place with token tok: same uid/owner/board slot,
// fresh stats, all enchantments/keywords/statuses dropped, summon-sick. The
// original finalGasp does NOT fire (it is replaced, not killed). Auras are
// recomputed by the caller's finish(). Used by EffectTransform / EffectTransformRandom.
func (m *Match) transformMinion(mn *minion, tok cards.Card) {
	mn.card = tok
	mn.enchants = nil
	mn.auraAtk = 0
	mn.auraHP = 0
	mn.health = tok.Health
	mn.summonedThisTurn = true
	mn.frozen = false
	mn.silenced = false
	mn.aegis = tok.Has(cards.KeywordAegis)
	mn.stealthed = tok.Has(cards.KeywordStealth)
	mn.attacksMade = 0
	m.emit(protocol.Event{Kind: "transform", Target: mn.uid, Name: tok.Name})
}

// cardTypeMatchesAura reports whether a card of type ct is affected by a CostAura
// scoped to type at. A spell-typed aura ("your spells cost less") also covers
// Secrets: a secret is a spell subtype that only carries TypeSecret so the engine
// can route it to the hidden secret zone (see cards/mage.go).
func cardTypeMatchesAura(ct, at cards.Type) bool {
	return ct == at || (at == cards.TypeSpell && ct == cards.TypeSecret)
}

// effectiveCost is the mana cost player pi pays for a hand card right now, after
// every cost modifier: `spellwarden_magus`'s free-next-secret override, the card's intrinsic
// CostRule (e.g. `tidecolossus`'s per-board-minion discount), and every in-play CostAura
// (`mana_leech` / `arcane_adept` / `pocket_conjurer`). Floored at 0.
// Computed fresh on each read — no stored per-card state, no recompute hook.
func (m *Match) effectiveCost(pi int, card cards.Card) int {
	// `spellwarden_magus`: the next Secret this turn is free — overrides all other math.
	if card.Type == cards.TypeSecret && m.state[pi].nextSecretFree {
		return 0
	}
	// `fizzle_sparkmuddle`: this player's spells are free for one specific (future) turn.
	// spellsFreeOnTurn==0 means "never set" (it is only ever assigned turnNum+1 ≥ 1).
	if card.Type == cards.TypeSpell && m.state[pi].spellsFreeOnTurn != 0 && m.state[pi].spellsFreeOnTurn == m.turnNum {
		return 0
	}
	cost := card.Cost
	if r := card.CostRule; r != nil {
		if r.PerBoardMinion != 0 {
			// The card is in hand, so every board minion (either side) counts as "other".
			cost += r.PerBoardMinion * (len(m.state[0].board) + len(m.state[1].board))
		}
		if r.PerOwnWeaponAttack != 0 {
			cost += r.PerOwnWeaponAttack * heroAttackValue(m.state[pi]) // `dread_buccaneer`: -1 per weapon Attack
		}
		if r.PerCardInHand != 0 {
			// `crag_colossus`: per OTHER card in hand. The card itself is in hand while
			// this cost is read, so subtract it.
			if others := len(m.state[pi].hand) - 1; others > 0 {
				cost += r.PerCardInHand * others
			}
		}
		if r.PerMissingHealth != 0 {
			cost += r.PerMissingHealth * (heroMaxHP - m.state[pi].heroHP) // `magma_behemoth`: per missing Health
		}
	}
	for ai := 0; ai < 2; ai++ {
		for _, src := range m.state[ai].board {
			ca := src.card.CostAura
			if ca == nil || src.silenced {
				continue
			}
			if ca.Scope == cards.CostScopeFriendly && ai != pi {
				continue
			}
			if ca.Type != "" && !cardTypeMatchesAura(card.Type, ca.Type) {
				continue
			}
			if ca.FirstMinionEachTurn && (card.Type != cards.TypeMinion || m.state[pi].minionsPlayedThisTurn > 0) {
				continue
			}
			cost += ca.Delta
		}
	}
	if cost < 0 {
		return 0
	}
	return cost
}

// applyEffect mutates game state per the effect, emitting events as it goes.
// caster is the player index from whose perspective sides (enemy/friendly) and
// summons resolve — the spell's owner, or a dying minion's owner for a
// finalGasp. ref is the chosen target for player-targeted effects. srcID is the
// id of the entity dealing the effect (a casting minion's uid, or the caster
// hero's id for hand spells / hero powers) — attached to damage events so the
// client can animate the shooter and the log can read "X hits Y". Caller holds m.mu.
func (m *Match) applyEffect(caster int, eff *cards.Effect, ref charRef, sp int, srcID string) {
	switch eff.Kind {
	case cards.EffectDamage:
		// `frostlance`: per target, deal FrozenDamage if it is ALREADY Frozen, otherwise
		// Freeze it (and deal nothing). Spell Damage boosts the damage branch.
		if eff.FrozenDamage > 0 {
			for _, t := range m.damageTargets(caster, eff, ref) {
				frozen := (t.minion != nil && t.minion.frozen) || (t.minion == nil && m.state[t.owner].frozen)
				if frozen {
					if t.minion != nil {
						m.damageMinion(t.minion, eff.FrozenDamage+sp, srcID)
					} else {
						m.damageHero(t.owner, eff.FrozenDamage+sp, srcID)
					}
				} else if t.minion != nil {
					m.freezeMinion(t.minion)
				} else {
					m.freezeHero(t.owner)
				}
			}
			return
		}
		// Damage and/or Freeze each character the effect addresses. Amount may be
		// 0 (a pure-Freeze effect like Permafrost). Spell Damage (sp) adds to each
		// damage instance that deals damage, never to a 0-damage effect.
		amt := eff.Amount
		if amt > 0 {
			amt += sp
		}
		// `glacial_splinter`: capture the target minion's Frozen state BEFORE damage, so a
		// conditional draw fires off the pre-damage status (the minion may die).
		wasFrozen := eff.DrawIfFrozen > 0 && ref.minion != nil && ref.minion.frozen
		total := 0
		for _, t := range m.damageTargets(caster, eff, ref) {
			if t.minion != nil {
				total += m.damageMinion(t.minion, amt, srcID)
				if eff.Freeze {
					m.freezeMinion(t.minion)
				}
			} else {
				total += m.damageHero(t.owner, amt, srcID)
				if eff.Freeze {
					m.freezeHero(t.owner)
				}
			}
		}
		if eff.Lifesteal {
			m.lifestealHeal(caster, total)
		}
		if wasFrozen {
			for i := 0; i < eff.DrawIfFrozen; i++ {
				m.drawCard(caster)
			}
		}
		for i := 0; i < eff.ThenDraw; i++ {
			m.drawCard(caster) // Anthem of Ambush: draw after the damage
		}
	case cards.EffectSilence:
		if ref.minion != nil {
			m.silence(ref.minion)
		}
	case cards.EffectHeal:
		// Heals share the target resolver, so a single-target heal (default ref) and a
		// mass heal (AreaFriendlyChars: `darkscale_mender`) take the same path.
		for _, t := range m.damageTargets(caster, eff, ref) {
			healed := 0
			if t.minion != nil {
				before := t.minion.health
				t.minion.health = min(t.minion.health+eff.Amount, t.minion.maxHP())
				healed = t.minion.health - before
				m.emit(protocol.Event{Kind: "heal", Target: t.minion.uid, Amount: healed})
			} else {
				before := m.state[t.owner].heroHP
				m.state[t.owner].heroHP = min(m.state[t.owner].heroHP+eff.Amount, heroMaxHP)
				healed = m.state[t.owner].heroHP - before
				m.emit(protocol.Event{Kind: "heal", Target: m.pid(t.owner), Amount: healed})
			}
			if healed > 0 {
				m.fireTriggers(caster, cards.OnHeal, nil) // global: "whenever a character is healed"
			}
		}
	case cards.EffectBuff:
		// Buffs share the target resolver, so they support adjacency
		// (`bannerguard` / `wardstone_sentinel`) and self-target alike. Grant adds keywords.
		// PerCardInHand scales the buff by the caster's current hand size (the played
		// minion has already left the hand, so it doesn't count itself).
		scale := 1
		if eff.PerCardInHand {
			scale = len(m.state[caster].hand)
		}
		if eff.PerOtherFriendlyMinion {
			scale = 0
			for _, mn := range m.state[caster].board {
				if mn != ref.minion { // every friendly minion except the source itself
					scale++
				}
			}
		}
		for _, t := range m.damageTargets(caster, eff, ref) {
			if t.minion == nil {
				continue
			}
			ba, bh := eff.BuffAtk*scale, eff.BuffHP*scale
			t.minion.enchants = append(t.minion.enchants, enchant{atk: ba, hp: bh, spellDamage: eff.GrantSpellDamage, keywords: eff.Grant, temp: eff.Temporary})
			t.minion.health += bh // a +health buff raises current health too
			if eff.DestroyNextTurn {
				t.minion.destroyAtTurnStart = true // Nightmare: dies at the owner's next turn start
			}
			m.emit(protocol.Event{Kind: "buff", Target: t.minion.uid, BuffAtk: ba, BuffHP: bh})
		}
	case cards.EffectMissiles:
		// Fire Count missiles of Amount damage, each at a freshly-chosen random
		// character among all OTHER characters (both heroes + every minion except the
		// source). Re-picked per missile so a kill removes that target from the pool.
		n, amt := eff.Count, eff.Amount
		if n < 1 {
			n = 1
		}
		if amt < 1 {
			amt = 1
		}
		for i := 0; i < n; i++ {
			// AreaEnemyChars scopes the missiles to enemy characters (`arcane_barrage`);
			// the default spreads among all other characters (`powder_tosser`).
			var pool []charRef
			if eff.Area == cards.AreaEnemyChars {
				pool = m.enemyChars(caster)
			} else {
				pool = m.otherChars(ref.minion)
			}
			if len(pool) == 0 {
				break
			}
			t := pool[m.rng.Intn(len(pool))]
			if t.minion != nil {
				m.damageMinion(t.minion, amt, srcID)
			} else {
				m.damageHero(t.owner, amt, srcID)
			}
		}
	case cards.EffectSwapStats:
		// Swap the target's current Attack and Health. Expressed as one enchant so
		// the new totals land on the derived atk()/maxHP(): the minion ends at
		// atk == old health and health == max health == old attack (a 0-attack swap
		// drops it to 0 health and it dies in finish()).
		if ref.minion == nil {
			return
		}
		mn := ref.minion
		curAtk, curHP := mn.atk(), mn.health
		mn.enchants = append(mn.enchants, enchant{atk: curHP - curAtk, hp: curAtk - mn.maxHP()})
		mn.health = curAtk
		m.emit(protocol.Event{Kind: "buff", Target: mn.uid})
	case cards.EffectConsumeShields:
		// Strip every Aegis on both boards (popping each), then buff the
		// source minion by BuffAtk/BuffHP per shield removed. ref is the source
		// (self-anchored at play time).
		count := 0
		for pi := 0; pi < 2; pi++ {
			for _, mn := range m.state[pi].board {
				if mn.aegis {
					mn.aegis = false
					m.emit(protocol.Event{Kind: "shield", Target: mn.uid, Name: mn.card.Name})
					count++
				}
			}
		}
		if count > 0 && ref.minion != nil {
			ba, bh := eff.BuffAtk*count, eff.BuffHP*count
			ref.minion.enchants = append(ref.minion.enchants, enchant{atk: ba, hp: bh})
			ref.minion.health += bh
			m.emit(protocol.Event{Kind: "buff", Target: ref.minion.uid, BuffAtk: ba, BuffHP: bh})
		}
	case cards.EffectGainWeaponAttack:
		// Buff the source minion (self-anchored) by the caster's weapon Attack.
		w := m.state[caster].weapon
		if w == nil || w.attack <= 0 || ref.minion == nil {
			return
		}
		ref.minion.enchants = append(ref.minion.enchants, enchant{atk: w.attack})
		m.emit(protocol.Event{Kind: "buff", Target: ref.minion.uid, BuffAtk: w.attack})
	case cards.EffectChipWeapon:
		// Remove Amount durability from the opponent's weapon; break it at 0.
		n := eff.Amount
		if n < 1 {
			n = 1
		}
		m.chipWeapon(1-caster, n)
	case cards.EffectBuffWeapon:
		// Give the caster's weapon +BuffAtk Attack / +BuffHP Durability.
		if w := m.state[caster].weapon; w != nil {
			w.attack += eff.BuffAtk
			w.durability += eff.BuffHP
		}
	case cards.EffectDestroyWeapon:
		// Destroy the opponent's weapon. With DrawWeaponDurability the caster also
		// draws cards equal to its remaining durability (`relic_breaker`); plain
		// destroy (`corroding_ooze`) draws nothing. Fatigue/over-draw via drawCard.
		opp := 1 - caster
		w := m.state[opp].weapon
		if w == nil {
			return
		}
		draws := 0
		if eff.DrawWeaponDurability {
			draws = w.durability
		}
		m.emit(protocol.Event{Kind: "weaponBreak", Source: m.pid(opp), Name: w.card.Name})
		m.state[opp].weapon = nil
		for i := 0; i < draws; i++ {
			m.drawCard(caster)
		}
	case cards.EffectKillSecret:
		// Destroy one random enemy Secret. Not named in the (shared) event log — the
		// caster must not learn which secret it was.
		opp := 1 - caster
		secs := m.state[opp].secrets
		if len(secs) == 0 {
			return
		}
		i := m.rng.Intn(len(secs))
		m.state[opp].secrets = append(secs[:i], secs[i+1:]...)
		m.emit(protocol.Event{Kind: "destroy", Source: m.pid(opp), Name: "Secret"})
	case cards.EffectBounce:
		// Return the target minion to its owner's hand as its base card (all
		// enchantments/damage reset). Burned if the hand is full. Death/auras
		// resolve in finish().
		if ref.minion == nil {
			return
		}
		m.bounceMinion(ref.minion, ref.owner)
	case cards.EffectSummon:
		tok, ok := cards.Get(eff.Summon)
		if !ok {
			return
		}
		n := eff.Count
		if n < 1 {
			n = 1
		}
		if eff.CountMax > n {
			n += m.rng.Intn(eff.CountMax - n + 1) // Anthem of the Muster: a random Count..CountMax
		}
		side := caster
		if eff.SummonForOpponent {
			side = 1 - caster // e.g. `the_gorehound`: summon for the opponent
		}
		for i := 0; i < n; i++ {
			if m.summonMinion(side, tok) == nil {
				break // board full: remaining tokens are discarded
			}
		}
	case cards.EffectMana:
		// Mana Surge: temporary mana for this turn only (capped at the crystal cap).
		ps := m.state[caster]
		ps.mana = min(ps.mana+eff.Amount, maxMana)
		m.emit(protocol.Event{Kind: "mana", Target: m.pid(caster), Amount: eff.Amount})
	case cards.EffectDraw:
		// The caster (or, for ToOpponent, the opponent) draws Amount cards (default 1).
		// drawCard handles fatigue and over-draw burn; a normal draw emits no event.
		// ReqDeckAllOdd gates the draw on every remaining deck card being odd-cost
		// (`shadowtail_familiar`). `brineseer_diviner` uses two effects (self + ToOpponent) for "each player".
		if eff.ReqDeckAllOdd && !deckAllOdd(m.state[caster].deck) {
			return
		}
		who := caster
		if eff.ToOpponent {
			who = 1 - caster
		}
		n := eff.Amount
		if n == 0 {
			n = 1
		}
		for i := 0; i < n; i++ {
			m.drawCard(who)
		}
	case cards.EffectGenerate:
		// Add Count copies (default 1) of a specific card to a hand — the caster's, or
		// the opponent's for ToOpponent (`grovelord_brakka`). Hidden identity, so the event
		// carries no name; a full hand burns the copy.
		gen, ok := cards.Get(eff.Generate)
		if !ok {
			return
		}
		who := caster
		if eff.ToOpponent {
			who = 1 - caster
		}
		n := eff.Count
		if n < 1 {
			n = 1
		}
		for i := 0; i < n; i++ {
			if len(m.state[who].hand) >= maxHand {
				m.emitBurn(who, gen)
				continue
			}
			m.state[who].hand = append(m.state[who].hand, gen)
			m.emit(protocol.Event{Kind: "generate", Target: m.pid(who)})
		}
	case cards.EffectDestroy:
		// Destroy each target minion outright (ignores Aegis). The death
		// itself (and any finalGasp) resolves in finish() via resolveDeaths.
		for _, t := range m.damageTargets(caster, eff, ref) {
			if t.minion != nil {
				t.minion.health = 0
				m.emit(protocol.Event{Kind: "destroy", Target: t.minion.uid, Name: t.minion.card.Name})
			}
		}
		if eff.DiscardHand {
			m.discardHand(caster) // `voidwyrm_tyrant` also discards the caster's remaining hand
		}
	case cards.EffectResummonDead:
		// Summon the caster's minions that died this turn, as their base cards (no
		// buffs), in death order. Board cap applies via summonMinion. Snapshot the
		// list first — summoning doesn't extend it, but be explicit. (`revenant_priestess`.)
		dead := append([]cards.Card(nil), m.state[caster].diedThisTurn...)
		for _, c := range dead {
			if m.summonMinion(caster, c) == nil {
				break // board full: the rest are discarded
			}
		}
	case cards.EffectTransform:
		// Replace the target minion with a token, in place (same uid/owner/board slot,
		// fresh stats, all enchantments/keywords/statuses dropped). The new minion is
		// summon-sick (no original finalGasp fires — it's replaced, not killed).
		tok, ok := cards.Get(eff.Transform)
		if ref.minion == nil || !ok {
			return
		}
		m.transformMinion(ref.minion, tok)
	case cards.EffectTransformRandom:
		// Transform a random OTHER minion (either board) into a random token from
		// GenIDs. Self-anchored, so the source (ref.minion) is excluded. (`sprocket_tinkerer`.)
		var pool []*minion
		for pi := 0; pi < 2; pi++ {
			for _, mn := range m.state[pi].board {
				if mn != ref.minion {
					pool = append(pool, mn)
				}
			}
		}
		if len(pool) == 0 || len(eff.GenIDs) == 0 {
			return
		}
		tok, ok := cards.Get(eff.GenIDs[m.rng.Intn(len(eff.GenIDs))])
		if !ok {
			return
		}
		m.transformMinion(pool[m.rng.Intn(len(pool))], tok)
	case cards.EffectGenerateRandom:
		// Add a random card to the caster's hand (burn if full). The pool is the
		// explicit GenIDs list (`dreamwarden_ylena`'s Dream cards) or, failing that, the filtered
		// collectible pool (`codex_of_insight`, `gleamwing`). Identity hidden from the opponent.
		ids := eff.GenIDs
		if len(ids) == 0 {
			ids = cards.RandomGenPoolIDs(eff.GenClass, eff.GenType, eff.GenRarity, eff.GenTribe)
		}
		if len(ids) == 0 {
			return
		}
		gen, ok := cards.Get(ids[m.rng.Intn(len(ids))])
		if !ok {
			return
		}
		who := caster
		if eff.ToOpponent {
			who = 1 - caster // Warhorn Chieftain: give the opponent one too
		}
		if len(m.state[who].hand) >= maxHand {
			m.emitBurn(who, gen)
			return
		}
		m.state[who].hand = append(m.state[who].hand, gen)
		m.emit(protocol.Event{Kind: "generate", Target: m.pid(who)})
	case cards.EffectFreeNextSecret:
		// `spellwarden_magus`: the caster's next Secret this turn costs 0 (consumed on play,
		// cleared at the caster's next turn start).
		m.state[caster].nextSecretFree = true
	case cards.EffectEnemySpellsFree:
		// `fizzle_sparkmuddle`: the opponent's spells cost 0 on their next turn. turnNum+1 is the
		// opponent's upcoming turn; the cost check matches that exact turnNum, so no
		// flag-clearing is needed (later turns never match).
		m.state[1-caster].spellsFreeOnTurn = m.turnNum + 1
	case cards.EffectSwapWithHand:
		// `clockwork_swapbot`: swap the source minion (ref, self-anchored) with a random
		// MINION in the caster's hand — the hand minion enters play in the source's
		// slot (no onset, it's a transform-in-place), and the source returns to
		// hand as its base card. Net hand size is unchanged, so no overdraw.
		if ref.minion == nil {
			return
		}
		ps := m.state[caster]
		var idxs []int
		for i, c := range ps.hand {
			if c.Type == cards.TypeMinion {
				idxs = append(idxs, i)
			}
		}
		if len(idxs) == 0 {
			return
		}
		pick := idxs[m.rng.Intn(len(idxs))]
		handCard := ps.hand[pick]
		selfBase, ok := cards.Get(ref.minion.card.ID)
		if !ok {
			selfBase = ref.minion.card
		}
		ps.hand = append(ps.hand[:pick], ps.hand[pick+1:]...)
		m.transformMinion(ref.minion, handCard) // in place: no onset, no finalGasp
		ps.hand = append(ps.hand, selfBase)
	case cards.EffectGiveOppMana:
		// `runed_golem`: give the opponent an empty Mana Crystal (max mana +1, capped;
		// their current mana is unchanged).
		opp := 1 - caster
		if m.state[opp].maxMana < maxMana {
			m.state[opp].maxMana++
			m.emit(protocol.Event{Kind: "mana", Target: m.pid(opp), Amount: 1})
		}
	case cards.EffectSetHealth:
		// Set the target hero's Health to Amount (clamped to the hero max). A direct
		// set — no armor interaction and it does NOT fire heal/damage triggers
		// (`emberqueen_valtha`). Only heroes are valid targets (TargetHero).
		if ref.minion != nil {
			return
		}
		hp := min(eff.Amount, heroMaxHP)
		m.state[ref.owner].heroHP = hp
		m.emit(protocol.Event{Kind: "sethealth", Target: m.pid(ref.owner), Amount: hp})
	case cards.EffectSummonRandom:
		// Summon a random minion onto the caster's board (board cap via summonMinion).
		// The pool is the explicit GenIDs (Gearmaster Cog / Anthem of War) or, failing
		// that, the filtered collectible pool (`wilds_beastcaller`).
		ids := eff.GenIDs
		if len(ids) == 0 {
			ids = cards.RandomGenPoolIDs(eff.GenClass, eff.GenType, eff.GenRarity, eff.GenTribe)
		}
		if len(ids) == 0 {
			return
		}
		if c, ok := cards.Get(ids[m.rng.Intn(len(ids))]); ok {
			m.summonMinion(caster, c)
		}
	case cards.EffectMindControl:
		// Take control of a random enemy minion (`wraithqueen_selvara` finalGasp /
		// `mesmer_adept` onset). ReqOppMinions gates it (`mesmer_adept` needs 4+). Fizzles if the
		// caster's board is full. The stolen minion changes sides and is summon-sick.
		opp := 1 - caster
		if eff.ReqOppMinions > 0 && len(m.state[opp].board) < eff.ReqOppMinions {
			return
		}
		if len(m.state[opp].board) == 0 || len(m.state[caster].board) >= maxBoard {
			return
		}
		i := m.rng.Intn(len(m.state[opp].board))
		mn := m.state[opp].board[i]
		m.state[opp].board = append(m.state[opp].board[:i], m.state[opp].board[i+1:]...)
		mn.owner = caster
		mn.summonedThisTurn = true
		mn.attacksMade = 0
		m.state[caster].board = append(m.state[caster].board, mn)
		m.emit(protocol.Event{Kind: "control", Source: m.pid(caster), Target: mn.uid, Name: mn.card.Name})
	case cards.EffectTutorTribe:
		// Draw a random card of Effect.Tribe from the caster's deck (`corsair_macaw`).
		ps := m.state[caster]
		var idxs []int
		for i, c := range ps.deck {
			if c.Tribe == eff.Tribe {
				idxs = append(idxs, i)
			}
		}
		if len(idxs) == 0 {
			return
		}
		pick := idxs[m.rng.Intn(len(idxs))]
		c := ps.deck[pick]
		ps.deck = append(ps.deck[:pick], ps.deck[pick+1:]...)
		if len(ps.hand) >= maxHand {
			m.emitBurn(caster, c)
			return
		}
		ps.hand = append(ps.hand, c)
		m.emit(protocol.Event{Kind: "generate", Target: m.pid(caster)})
	}
}

// damageMinion applies damage to a minion and returns the damage actually dealt.
// A Aegis absorbs the instance (popping, 0 dealt). Zero/negative amounts
// do nothing — so a pure-Freeze effect neither damages nor pops a shield.
func (m *Match) damageMinion(mn *minion, amt int, srcID string) int {
	if amt <= 0 {
		return 0
	}
	if mn.aegis {
		mn.aegis = false
		m.emit(protocol.Event{Kind: "shield", Target: mn.uid, Name: mn.card.Name})
		return 0
	}
	mn.health -= amt
	m.emit(protocol.Event{Kind: "damage", Source: srcID, Target: mn.uid, Amount: amt})
	// The damaged minion reacts to taking damage (a draw-on-damage minion: draw a card). Its
	// own OnDamage triggers fire from its controller's perspective. Silenced minions
	// don't react. Fires even if the hit was lethal (the draw still happens).
	if !mn.silenced {
		for _, t := range mn.card.Triggers {
			if t.When != cards.OnDamage {
				continue
			}
			e := t.Effect
			ref := charRef{owner: mn.owner}
			if e.Target == cards.TargetSelf {
				ref = charRef{minion: mn, owner: mn.owner}
			}
			m.emit(protocol.Event{Kind: "trigger", Source: mn.uid, Name: mn.card.Name})
			m.applyEffect(mn.owner, &e, ref, 0, mn.uid)
		}
	}
	return amt
}

// deckAllOdd reports whether every card remaining in the deck is odd-cost (Black
// Cat). An empty deck counts as satisfying the condition (vacuously true).
func deckAllOdd(deck []cards.Card) bool {
	for _, c := range deck {
		if c.Cost%2 == 0 {
			return false
		}
	}
	return true
}

// damageHero applies damage to a hero (player index h): armor absorbs first, the
// remainder hits health. Emits a damage event and returns the full damage dealt
// (armor-absorbed included, so Lifesteal heals for the whole hit, as in HS).
func (m *Match) damageHero(h, amt int, srcID string) int {
	if amt <= 0 {
		return 0
	}
	ps := m.state[h]
	if ps.immune {
		return 0 // `frostward_aegis`: the hero ignores all damage this turn
	}
	absorbed := min(ps.armor, amt)
	net := amt - absorbed
	// `frostward_aegis`: if this hit would be fatal, an active secret prevents it entirely
	// (no armor spent, no Health lost) and the hero becomes Immune for the turn.
	if net >= ps.heroHP && m.tryIceBlock(h) {
		return 0
	}
	ps.armor -= absorbed
	ps.heroHP -= net
	m.emit(protocol.Event{Kind: "damage", Source: srcID, Target: m.pid(h), Amount: amt})
	return amt
}

// tryIceBlock consumes player h's active `frostward_aegis` secret (if any), revealing it
// and making the hero Immune for the rest of the turn. Returns true if a secret
// fired. Called only when a hit would be fatal (see damageHero).
func (m *Match) tryIceBlock(h int) bool {
	ps := m.state[h]
	for i, s := range ps.secrets {
		if s.card.Secret != nil && s.card.Secret.Kind == cards.SecretIceBlock {
			scv := cardView(s.card)
			m.emit(protocol.Event{Kind: "secret", Source: m.pid(h), Name: s.card.Name, Card: &scv})
			ps.secrets = append(ps.secrets[:i], ps.secrets[i+1:]...)
			ps.immune = true
			return true
		}
	}
	return false
}

// gainArmor adds armor to a hero (no cap) and emits an armor event.
func (m *Match) gainArmor(h, amt int) {
	if amt <= 0 {
		return
	}
	m.state[h].armor += amt
	m.emit(protocol.Event{Kind: "armor", Target: m.pid(h), Amount: amt})
}

// combatStrike resolves one minion dealing its attack to another in combat,
// applying Poisonous (any damage dealt destroys the struck minion) and Lifesteal
// (damage dealt heals the striker's controller). Aegis absorption yields
// 0 dealt, so it suppresses both.
func (m *Match) combatStrike(src, dst *minion) {
	dealt := m.damageMinion(dst, src.atk(), src.uid)
	if dealt > 0 && src.has(cards.KeywordPoisonous) {
		dst.health = 0 // destroyed regardless of remaining health
	}
	if src.has(cards.KeywordLifesteal) {
		m.lifestealHeal(src.owner, dealt)
	}
	if dealt > 0 && src.has(cards.KeywordFreezeOnHit) {
		m.freezeMinion(dst) // `frostfont_elemental`: Freeze anything it damages
	}
}

// lifestealHeal restores the player's hero by amt (capped at the hero max) and
// emits a heal event. No-op for non-positive amounts.
func (m *Match) lifestealHeal(pi, amt int) {
	if amt <= 0 {
		return
	}
	before := m.state[pi].heroHP
	m.state[pi].heroHP = min(m.state[pi].heroHP+amt, heroMaxHP)
	m.emit(protocol.Event{Kind: "heal", Target: m.pid(pi), Amount: m.state[pi].heroHP - before})
}

// silence strips a minion's enchantments, keywords (Taunt/Charge/Rush/Divine
// Shield/Twinstrike/Stealth/Poisonous/Lifesteal/Spell Damage), aura, and triggers
// (onset already fired; finalGasps are suppressed via the silenced flag).
// Frozen is a status, not an enchantment, so it is NOT removed. Current health is
// clamped to the (now lower) max. Caller resolves deaths/auras afterwards.
func (m *Match) silence(mn *minion) {
	mn.silenced = true
	mn.enchants = nil
	mn.aegis = false
	mn.stealthed = false
	// No attack-count fixup needed: eligibility reads attacksPerTurn() live, so losing
	// Twinstrike simply caps this minion to one attack (its attacksMade already counts up).
	if mn.health > mn.maxHP() {
		mn.health = mn.maxHP()
	}
	m.emit(protocol.Event{Kind: "silence", Target: mn.uid, Name: mn.card.Name})
}

// refreshAuras recomputes every minion's aura-granted Attack and max-Health from
// scratch. Each non-silenced aura source contributes to the controller's other
// minions it covers (all, a tribe, or its neighbours). Health auras track a
// current-health delta: gaining max also raises current health; losing it only
// clamps current down to the new max (never subtracts a damaged minion to death).
// Called whenever the board may have changed.
func (m *Match) refreshAuras() {
	newAtk := map[*minion]int{}
	newHP := map[*minion]int{}
	for pi := 0; pi < 2; pi++ {
		for _, src := range m.state[pi].board {
			if src.silenced {
				continue
			}
			if a := src.card.Aura; a != nil {
				for _, mn := range m.auraTargets(pi, src, a) {
					newAtk[mn] += a.Atk
					newHP[mn] += a.HP
				}
			}
			// SelfCountAtk: +Atk per OTHER in-play minion of a tribe (e.g.
			// `brinelord_gorrak`). Self-only, silence-cancelled (the guard above).
			if sc := src.card.SelfCountAtk; sc != nil {
				newAtk[src] += sc.Atk * m.countTribe(sc.Tribe, src)
			}
		}
	}
	for pi := 0; pi < 2; pi++ {
		for _, mn := range m.state[pi].board {
			mn.auraAtk = newAtk[mn]
			if delta := newHP[mn] - mn.auraHP; delta > 0 {
				mn.health += delta // gaining max health also heals
			}
			mn.auraHP = newHP[mn]
			if mn.health > mn.maxHP() {
				mn.health = mn.maxHP() // losing max health clamps current down
			}
		}
	}
}

// auraTargets returns the OTHER friendly minions an aura source covers: all of the
// controller's other minions, or — when the aura is tribe-scoped or adjacent —
// only those matching. owner is src's controller.
func (m *Match) auraTargets(owner int, src *minion, a *cards.Aura) []*minion {
	board := m.state[owner].board
	if a.Adjacent {
		idx := -1
		for i, mn := range board {
			if mn == src {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil
		}
		var out []*minion
		if idx-1 >= 0 {
			out = append(out, board[idx-1])
		}
		if idx+1 < len(board) {
			out = append(out, board[idx+1])
		}
		return out
	}
	var out []*minion
	for _, mn := range board {
		if mn == src {
			continue // "other" minions only
		}
		if a.Tribe != cards.TribeNone && mn.card.Tribe != a.Tribe {
			continue
		}
		out = append(out, mn)
	}
	return out
}

// countTribe counts minions of the given tribe in play on either board, excluding
// the source minion itself (used by SelfCountAtk, e.g. `brinelord_gorrak`).
func (m *Match) countTribe(tribe cards.Tribe, except *minion) int {
	n := 0
	for pi := 0; pi < 2; pi++ {
		for _, mn := range m.state[pi].board {
			if mn != except && mn.card.Tribe == tribe {
				n++
			}
		}
	}
	return n
}

// spellPower is the total spell damage bonus for player pi: the sum of its
// non-silenced minions' SpellDamage.
func (m *Match) spellPower(pi int) int {
	sp := 0
	for _, mn := range m.state[pi].board {
		sp += spellDamageOf(mn)
	}
	return sp
}

// freezeMinion / freezeHero mark a character frozen and emit a freeze event.
func (m *Match) freezeMinion(mn *minion) {
	mn.frozen = true
	m.emit(protocol.Event{Kind: "freeze", Target: mn.uid, Name: mn.card.Name})
}

func (m *Match) freezeHero(h int) {
	m.state[h].frozen = true
	m.emit(protocol.Event{Kind: "freeze", Target: m.pid(h)})
}

// damageTargets resolves the set of characters a damage/freeze effect addresses,
// from caster's perspective.
func (m *Match) damageTargets(caster int, eff *cards.Effect, ref charRef) []charRef {
	switch {
	case eff.Area == cards.AreaEnemyMinions:
		opp := 1 - caster
		out := make([]charRef, 0, len(m.state[opp].board))
		for _, mn := range m.state[opp].board {
			out = append(out, charRef{minion: mn, owner: opp})
		}
		return out
	case eff.Area == cards.AreaAllMinions:
		out := make([]charRef, 0, len(m.state[0].board)+len(m.state[1].board))
		for pi := 0; pi < 2; pi++ {
			for _, mn := range m.state[pi].board {
				out = append(out, charRef{minion: mn, owner: pi})
			}
		}
		return out
	case eff.Area == cards.AreaEnemyHero:
		return []charRef{{owner: 1 - caster}}
	case eff.Area == cards.AreaAllCharacters || eff.Area == cards.AreaOtherCharacters:
		// Both heroes and every minion; AreaOtherCharacters omits the anchor minion.
		out := []charRef{{owner: 0}, {owner: 1}}
		for pi := 0; pi < 2; pi++ {
			for _, mn := range m.state[pi].board {
				if eff.Area == cards.AreaOtherCharacters && mn == ref.minion {
					continue
				}
				if eff.ExceptCardID != "" && mn.card.ID == eff.ExceptCardID {
					continue // `dream_emerald_reckoning` spares Dreamwardens
				}
				out = append(out, charRef{minion: mn, owner: pi})
			}
		}
		return out
	case eff.Area == cards.AreaOtherMinions:
		// Every minion on both boards except the anchor (self-anchored, e.g. `voidwyrm_tyrant`).
		out := make([]charRef, 0, len(m.state[0].board)+len(m.state[1].board))
		for pi := 0; pi < 2; pi++ {
			for _, mn := range m.state[pi].board {
				if mn == ref.minion {
					continue
				}
				out = append(out, charRef{minion: mn, owner: pi})
			}
		}
		return out
	case eff.Area == cards.AreaAdjacent:
		return m.adjacentRefs(ref, false)
	case eff.Area == cards.AreaSplash:
		return m.adjacentRefs(ref, true)
	case eff.Area == cards.AreaRandomEnemyMinion:
		opp := 1 - caster
		var pool []*minion
		for _, mn := range m.state[opp].board {
			if eff.MaxAttack > 0 && mn.atk() > eff.MaxAttack {
				continue
			}
			pool = append(pool, mn)
		}
		if len(pool) == 0 {
			return nil
		}
		return []charRef{{minion: pool[m.rng.Intn(len(pool))], owner: opp}}
	case eff.Area == cards.AreaFriendlyChars:
		// The caster's hero and every friendly minion (`darkscale_mender` mass heal).
		out := []charRef{{owner: caster}}
		for _, mn := range m.state[caster].board {
			out = append(out, charRef{minion: mn, owner: caster})
		}
		return out
	case eff.Area == cards.AreaFriendlyTribe:
		var out []charRef
		for _, mn := range m.state[caster].board {
			if mn != ref.minion && mn.card.Tribe == eff.Tribe { // OTHER friendly minions of the tribe
				out = append(out, charRef{minion: mn, owner: caster})
			}
		}
		return out
	case eff.Target == cards.TargetRandomEnemy:
		return []charRef{m.randomEnemy(caster)}
	default:
		return []charRef{ref} // chosen minion or hero
	}
}

// adjacentRefs returns the minions positionally next to the anchor on its own
// board, in board order. inclusive adds the anchor itself (between its
// neighbours), so AreaSplash hits left-anchor-right. Returns nil if the anchor is
// a hero or no longer on the board.
func (m *Match) adjacentRefs(anchor charRef, inclusive bool) []charRef {
	if anchor.minion == nil {
		return nil
	}
	board := m.state[anchor.owner].board
	idx := -1
	for i, mn := range board {
		if mn == anchor.minion {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}
	var out []charRef
	if idx-1 >= 0 {
		out = append(out, charRef{minion: board[idx-1], owner: anchor.owner})
	}
	if inclusive {
		out = append(out, anchor)
	}
	if idx+1 < len(board) {
		out = append(out, charRef{minion: board[idx+1], owner: anchor.owner})
	}
	return out
}

// randomFriendlyExcept picks a random minion on owner's board other than except,
// or nil if there is none. Used by random-friendly triggers (e.g. end-of-turn
// "give another friendly minion +1").
func (m *Match) randomFriendlyExcept(owner int, except *minion) *minion {
	var pool []*minion
	for _, mn := range m.state[owner].board {
		if mn != except {
			pool = append(pool, mn)
		}
	}
	if len(pool) == 0 {
		return nil
	}
	return pool[m.rng.Intn(len(pool))]
}

// discardHand discards every card in player pi's hand. Identities are hidden, so
// each discard is logged only as a burn (no name). Used by `voidwyrm_tyrant`'s onset.
func (m *Match) discardHand(pi int) {
	ps := m.state[pi]
	for _, c := range ps.hand {
		m.emitBurn(pi, c)
	}
	ps.hand = nil
}

// chipWeapon removes n durability from player h's weapon, breaking it at 0.
func (m *Match) chipWeapon(h, n int) {
	w := m.state[h].weapon
	if w == nil {
		return
	}
	w.durability -= n
	if w.durability <= 0 {
		m.emit(protocol.Event{Kind: "weaponBreak", Source: m.pid(h), Name: w.card.Name})
		m.state[h].weapon = nil
	}
}

// otherChars returns every living character except src: both heroes and every
// minion (on either board) with health > 0 other than src. Used by EffectMissiles
// to spread random damage among "all other characters".
func (m *Match) otherChars(src *minion) []charRef {
	var out []charRef
	for pi := 0; pi < 2; pi++ {
		out = append(out, charRef{owner: pi}) // hero
		for _, mn := range m.state[pi].board {
			if mn != src && mn.health > 0 {
				out = append(out, charRef{minion: mn, owner: pi})
			}
		}
	}
	return out
}

// enemyChars returns every enemy character (the opponent's hero plus their live
// minions) from caster's perspective. Used by enemy-scoped missiles (`arcane_barrage`).
func (m *Match) enemyChars(caster int) []charRef {
	opp := 1 - caster
	out := []charRef{{owner: opp}} // enemy hero is always a candidate
	for _, mn := range m.state[opp].board {
		if mn.health > 0 {
			out = append(out, charRef{minion: mn, owner: opp})
		}
	}
	return out
}

// randomEnemy picks a random enemy character (a minion or the hero) of caster's
// opponent. The hero is always a candidate, so this never returns nothing.
func (m *Match) randomEnemy(caster int) charRef {
	opp := 1 - caster
	n := len(m.state[opp].board)
	if pick := m.rng.Intn(n + 1); pick < n { // 0..n-1 = minion, n = hero
		return charRef{minion: m.state[opp].board[pick], owner: opp}
	}
	return charRef{owner: opp}
}
