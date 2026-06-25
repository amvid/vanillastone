package match

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/protocol"
)

// hasCharge reports whether a minion effectively has Charge: the keyword, or
// `tideblade_raider`'s conditional Charge granted while its controller has a weapon.
func (m *Match) hasCharge(mn *minion) bool {
	if mn.has(cards.KeywordCharge) {
		return true
	}
	return mn.card.ChargeWithWeapon && !mn.silenced && m.state[mn.owner].weapon != nil
}

// canAttack reports whether a minion is able to attack something this turn:
// positive attack, not yet swung, not frozen, and either not summon-sick or has
// Charge/Rush.
func (m *Match) canAttack(mn *minion) bool {
	return mn.atk() > 0 && mn.attacksMade < mn.attacksPerTurn() && !mn.frozen && !mn.has(cards.KeywordCantAttack) &&
		(!mn.summonedThisTurn || m.hasCharge(mn) || mn.has(cards.KeywordRush))
}

// canAttackHero is canAttack minus the Rush restriction: a Rush minion (without
// Charge) cannot hit heroes the turn it is summoned.
func (m *Match) canAttackHero(mn *minion) bool {
	if !m.canAttack(mn) {
		return false
	}
	if mn.summonedThisTurn && mn.has(cards.KeywordRush) && !m.hasCharge(mn) {
		return false
	}
	return true
}

// hasTaunt reports whether the player controls any Taunt minion the enemy can
// be forced to attack. A Stealthed Taunt is hidden, so it does not compel.
func hasTaunt(ps *playerState) bool {
	for _, mn := range ps.board {
		if mn.has(cards.KeywordTaunt) && !mn.stealthed {
			return true
		}
	}
	return false
}

// resolveDeaths removes dead minions and fires their finalGasps, cascading
// until no minion is left at <=0 health (a finalGasp's damage can kill more).
// Deaths are emitted and finalGasps fire in board order. Heroes are not
// removed here — finish does the win check. Caller holds m.mu.
func (m *Match) resolveDeaths() {
	for {
		var dead []*minion
		for pi := 0; pi < 2; pi++ {
			for _, mn := range m.state[pi].board {
				if mn.health <= 0 {
					dead = append(dead, mn)
				}
			}
		}
		if len(dead) == 0 {
			return
		}
		// Remove all dead first, so finalGasps see an updated board.
		m.state[0].board = removeDead(m.state[0].board)
		m.state[1].board = removeDead(m.state[1].board)
		for _, d := range dead {
			m.emit(protocol.Event{Kind: "death", Target: d.uid, Name: d.card.Name})
			// Record the death for "died this turn" effects (`revenant_priestess`). The base card
			// is stored so a resummon comes back fresh (no buffs).
			m.state[d.owner].diedThisTurn = append(m.state[d.owner].diedThisTurn, d.card)
			// Surviving minions react to the death (any-death and friendly-other-
			// death triggers). These fire regardless of the dead minion's Silence —
			// Silence only suppresses the dead minion's OWN finalGasp (below).
			m.fireTriggers(d.owner, cards.OnFriendlyDeath, d)
			m.fireTriggers(d.owner, cards.OnAnyMinionDeath, d)
			if d.silenced {
				continue // Silence suppresses finalGasps
			}
			for _, eff := range d.card.FinalGasps() {
				m.emit(protocol.Event{Kind: "finalGasp", Source: d.uid, Name: d.card.Name})
				e := eff
				m.applyEffect(d.owner, &e, charRef{}, 0, d.uid)
			}
		}
		// Loop: finalGasp damage may have created new deaths.
	}
}

// finish resolves all pending deaths (with finalGasps), then checks for a dead
// hero. If one died the match ends (winner = survivor); otherwise it just pushes
// the resulting snapshot. Always sends state. Caller holds m.mu.
func (m *Match) finish() {
	m.resolveDeaths()
	m.refreshAuras() // board changed; recompute aura buffs before snapshotting
	dead := -1
	if m.state[0].heroHP <= 0 {
		dead = 0
	}
	if m.state[1].heroHP <= 0 {
		dead = 1
	}
	if dead >= 0 {
		m.over = true
		m.stopTurnTimer()
		m.sendStateAll()
		m.broadcast(protocol.Marshal(protocol.GameOver{
			Type:   protocol.TypeGameOver,
			Winner: m.players[1-dead].ID(),
		}))
		return
	}
	m.sendStateAll()
}

func uid(n int) string { return "u" + itoa(n) }

// itoa avoids pulling strconv for a single small positive int.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func findMinion(board []*minion, uid string) *minion {
	for _, mn := range board {
		if mn.uid == uid {
			return mn
		}
	}
	return nil
}

// removeDead returns the board with health<=0 minions removed, preserving order.
func removeDead(board []*minion) []*minion {
	out := board[:0]
	for _, mn := range board {
		if mn.health > 0 {
			out = append(out, mn)
		}
	}
	return out
}

func (m *Match) minionViews(board []*minion) []protocol.MinionView {
	out := make([]protocol.MinionView, len(board))
	for i, mn := range board {
		out[i] = protocol.MinionView{
			InstanceID:    mn.uid,
			CardID:        mn.card.ID,
			Name:          mn.card.Name,
			Class:         string(mn.card.Class),
			Cost:          mn.card.Cost,
			Rarity:        string(mn.card.Rarity),
			Attack:        mn.atk(),
			Health:        mn.health,
			MaxHealth:     mn.maxHP(),
			BaseAttack:    mn.card.Attack,
			BaseHealth:    mn.card.Health,
			CanAttack:     m.canAttack(mn),
			CanAttackHero: m.canAttackHero(mn),
			Taunt:         mn.has(cards.KeywordTaunt),
			Aegis:         mn.aegis,
			Immune:        mn.has(cards.KeywordImmune),
			Frozen:        mn.frozen,
			Twinstrike:    mn.has(cards.KeywordTwinstrike),
			Stealth:       mn.stealthed,
			Poisonous:     mn.has(cards.KeywordPoisonous),
			Lifesteal:     mn.has(cards.KeywordLifesteal),
			FinalGasp:     !mn.silenced && len(mn.card.FinalGasps()) > 0,
			SpellDamage:   spellDamageOf(mn),
			Enraged:       mn.enraged(),
			HasEnrage:     !mn.silenced && (mn.card.Enrage != nil || mn.card.EnrageWeaponAtk > 0 || len(mn.card.EnrageGrant) > 0),
			Silenced:      mn.silenced,
			Elusive:       mn.has(cards.KeywordElusive),
			CantAttack:    mn.has(cards.KeywordCantAttack),
			Tribe:         string(mn.card.Tribe),
			Text:          minionText(mn),
		}
	}
	return out
}

// minionText is the rules text shown for a minion in play: blank once silenced
// (its abilities are gone, so the printed text would lie).
func minionText(mn *minion) string {
	if mn.silenced {
		return ""
	}
	return mn.card.Text
}

// emitPlay records that player pi played card (revealing it to both clients).
// Used for minions/spells/weapons; secrets are intentionally NOT revealed.
func (m *Match) emitPlay(pi int, card cards.Card) {
	cv := cardView(card)
	m.emit(protocol.Event{Kind: "play", Source: m.pid(pi), Name: card.Name, Card: &cv})
}

func cardView(c cards.Card) protocol.CardView {
	cv := protocol.CardView{
		CardID:     c.ID,
		Name:       c.Name,
		CardType:   string(c.Type),
		Class:      string(c.Class),
		Rarity:     string(c.Rarity),
		Cost:       c.Cost,
		Attack:     c.Attack,
		Health:     c.Health,
		Durability: c.Durability,
		Tribe:      string(c.Tribe),
		Text:       c.Text,
	}
	// Target is the targeting rule the client uses to arm targeting: a spell's
	// effect, or a minion's onset (on_play) effect. ReqAttack/ReqTaunt carry any
	// extra target condition so the client highlights only legal targets.
	if eff := c.Effect; eff != nil {
		cv.Target = string(eff.Target)
		cv.ReqAttack, cv.ReqTaunt, cv.ReqTribe = eff.ReqAttack, eff.ReqTaunt, string(eff.ReqTribe)
	} else if bc := c.Onset(); bc != nil {
		cv.Target = string(bc.Target)
		cv.ReqAttack, cv.ReqTaunt, cv.ReqTribe = bc.ReqAttack, bc.ReqTaunt, string(bc.ReqTribe)
	}
	return cv
}

// weaponView is the hero's equipped weapon (public), or nil.
func weaponView(w *weaponInst) *protocol.WeaponView {
	if w == nil {
		return nil
	}
	return &protocol.WeaponView{Name: w.card.Name, Attack: w.attack, Durability: w.durability, Text: w.card.Text}
}

// heroPowerView is the hero's power as a card (public).
func heroPowerView(ps *playerState) *protocol.CardView {
	hp := cardView(ps.heroPower)
	return &hp
}

// spellDamageText rewrites a spell's rules text so its boosted damage number
// reflects the caster's current Spell Damage (sp > 0), wrapping the bumped number
// in a {sd:N} marker the client renders green. Only EffectDamage spells are
// boosted at cast time (see applyEffect), so only those are rewritten; the boosted
// number is the effect's Amount, or its FrozenDamage when Amount is 0 (`frostlance`).
// The first whole-number occurrence of that base value in the text is the damage
// figure for our cards (they lead with "Deal N ..."). Returns text unchanged if
// nothing applies.
func spellDamageText(c cards.Card, sp int) string {
	if c.Type != cards.TypeSpell || c.Effect == nil || c.Effect.Kind != cards.EffectDamage {
		return c.Text
	}
	base := c.Effect.Amount
	if base == 0 {
		base = c.Effect.FrozenDamage
	}
	if base <= 0 {
		return c.Text
	}
	done := false
	re := regexp.MustCompile(`\b` + strconv.Itoa(base) + `\b`)
	return re.ReplaceAllStringFunc(c.Text, func(s string) string {
		if done {
			return s
		}
		done = true
		return fmt.Sprintf("{sd:%d}", base+sp)
	})
}

// selfView is the recipient's own side: hand cards and secrets are visible. Hand
// cards carry their effective (cost-modified) mana cost plus the printed BaseCost
// so the client can show + colour a changed cost. pi is the owner's player index.
func (m *Match) selfView(pi int, name string) protocol.PlayerView {
	ps := m.state[pi]
	sp := m.spellPower(pi)
	hand := make([]protocol.CardView, len(ps.hand))
	for i, c := range ps.hand {
		hand[i] = cardView(c)
		hand[i].BaseCost = c.Cost
		hand[i].Cost = m.effectiveCost(pi, c)
		if sp > 0 {
			hand[i].Text = spellDamageText(c, sp)
		}
	}
	secrets := make([]protocol.CardView, len(ps.secrets))
	for i, s := range ps.secrets {
		secrets[i] = cardView(s.card)
	}
	return protocol.PlayerView{
		Name:          name,
		HeroHP:        ps.heroHP,
		Armor:         ps.armor,
		Frozen:        ps.frozen,
		Immune:        ps.immune,
		Mana:          ps.mana,
		MaxMana:       ps.maxMana,
		Board:         m.minionViews(ps.board),
		Hand:          hand,
		HandCount:     len(ps.hand),
		DeckCount:     len(ps.deck),
		Secrets:       secrets,
		SecretCount:   len(ps.secrets),
		HeroPower:     heroPowerView(ps),
		HeroPowerUsed: ps.heroPowerUsed,
		Weapon:        weaponView(ps.weapon),
		HeroAttack:    heroAttackValue(ps),
		HeroCanAttack: heroCanAttack(ps),
	}
}

// oppView hides the opponent's hand and the identity of their secrets, exposing
// only the counts.
func (m *Match) oppView(pi int, name string) protocol.PlayerView {
	ps := m.state[pi]
	return protocol.PlayerView{
		Name:          name,
		HeroHP:        ps.heroHP,
		Armor:         ps.armor,
		Frozen:        ps.frozen,
		Immune:        ps.immune,
		Mana:          ps.mana,
		MaxMana:       ps.maxMana,
		Board:         m.minionViews(ps.board),
		HandCount:     len(ps.hand),
		DeckCount:     len(ps.deck),
		SecretCount:   len(ps.secrets),
		HeroPower:     heroPowerView(ps),
		HeroPowerUsed: ps.heroPowerUsed,
		Weapon:        weaponView(ps.weapon),
		HeroAttack:    heroAttackValue(ps),
		HeroCanAttack: heroCanAttack(ps),
	}
}
