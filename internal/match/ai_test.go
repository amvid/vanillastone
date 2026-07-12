package match

import (
	"testing"
	"time"

	"github.com/amvid/vanillastone/internal/cards"
)

// aiMatch builds a match with the bot seated at `seat`, its turn active, and a
// clean slate on both seats (no hands/boards) for the test to populate.
func aiMatch(seat int) *Match {
	m, _, _ := newMatch()
	m.EnableAI(seat, 1)
	m.turn = seat
	for pi := 0; pi < 2; pi++ {
		m.state[pi].hand = nil
		m.state[pi].board = nil
		m.state[pi].heroPowerUsed = true // silence the hero power unless a test wants it
		m.state[pi].mana, m.state[pi].maxMana = 0, 0
	}
	return m
}

func bodyMinion(uid string, owner, atk, hp int) *minion {
	return &minion{uid: uid, card: cards.Card{ID: uid, Type: cards.TypeMinion, Attack: atk, Health: hp}, owner: owner, health: hp}
}

// TestPlannerTradesToAvoidLethal is the headline behavior the user asked for: when
// the opponent's board threatens lethal next turn, the bot makes the defensive
// trade (kill the threat) instead of mindlessly attacking face. Encodes WHY: a
// line that leaves the bot dead next turn must score worse than removing the
// threat, even though facing deals more immediate damage.
func TestPlannerTradesToAvoidLethal(t *testing.T) {
	m := aiMatch(1)
	m.state[1].heroHP = 6                                       // bot is in lethal range
	m.state[0].board = []*minion{bodyMinion("threat", 0, 6, 6)} // enemy 6/6 → 6(+hero power) next turn = lethal on 6 HP
	m.state[1].board = []*minion{bodyMinion("mine", 1, 6, 6)}   // bot 6/6 can kill it

	mv, ok := m.planBest(1)
	if !ok {
		t.Fatal("planner found no move; expected a defensive trade")
	}
	if mv.kind != mAttack || mv.target != "threat" {
		t.Fatalf("expected attack into the threat, got %+v", mv)
	}
}

// TestPlannerTakesLethal: with lethal on board, the bot swings the hero, not a
// pointless trade. Encodes WHY: a winning line must dominate every other score.
func TestPlannerTakesLethal(t *testing.T) {
	m := aiMatch(1)
	m.state[0].heroHP = 5
	m.state[1].board = []*minion{bodyMinion("killer", 1, 6, 6)} // 6 ≥ 5 = lethal, no enemy taunt

	mv, ok := m.planBest(1)
	if !ok {
		t.Fatal("planner found no move; expected the lethal swing")
	}
	if mv.kind != mAttack || mv.target != oppHeroTarget {
		t.Fatalf("expected lethal swing at the hero, got %+v", mv)
	}
}

// TestPlannerSwingsWeaponForLethal: with an equipped weapon and the opponent in
// range, the bot swings the hero at the face for the kill. Encodes WHY: the hero
// attack must be an enumerated move — without it the bot equips weapons and never
// uses them (and so misses weapon lethals), the user-reported bug.
func TestPlannerSwingsWeaponForLethal(t *testing.T) {
	m := aiMatch(1)
	m.state[0].heroHP = 3
	m.state[1].weapon = &weaponInst{card: cards.Card{Type: cards.TypeWeapon, Attack: 3, Durability: 2}, attack: 3, durability: 2}

	mv, ok := m.planBest(1)
	if !ok {
		t.Fatal("planner found no move; expected the weapon swing")
	}
	if mv.kind != mAttack || mv.attacker != selfHeroTarget || mv.target != oppHeroTarget {
		t.Fatalf("expected hero weapon swing at the face, got %+v", mv)
	}
}

// TestPlannerDevelopsFromHand: an idle board with mana and a minion in hand should
// develop the minion (board presence beats doing nothing).
func TestPlannerDevelopsFromHand(t *testing.T) {
	m := aiMatch(1)
	m.state[1].mana, m.state[1].maxMana = 3, 3
	m.state[1].hand = []cards.Card{getCard("clay_acolyte")} // 2-mana 3/2

	mv, ok := m.planBest(1)
	if !ok {
		t.Fatal("planner found no move; expected to develop a minion")
	}
	if mv.kind != mPlay || mv.hand != 0 {
		t.Fatalf("expected to play the hand minion, got %+v", mv)
	}
}

// TestPlannerIgnoresHiddenSecrets: the bot must not plan against the opponent's
// hidden secrets. With an enemy secret that punishes minion plays (Echo Glass
// copies it), the bot should STILL develop its hand minion — seeing the secret
// fire in simulation and stalling on hero power is the fog-of-war leak we fixed.
func TestPlannerIgnoresHiddenSecrets(t *testing.T) {
	m := aiMatch(1)
	m.state[1].mana, m.state[1].maxMana = 3, 3
	m.state[1].hand = []cards.Card{getCard("clay_acolyte")} // 2-mana 3/2
	placeSecret(m, 0, "echo_glass")                         // opponent's enemy-play punisher

	mv, ok := m.planBest(1)
	if !ok {
		t.Fatal("planner found no move; the hidden secret must not deter developing")
	}
	if mv.kind != mPlay || mv.hand != 0 {
		t.Fatalf("expected to play the hand minion despite the secret, got %+v", mv)
	}
}

// TestPlannerHoldsIdleAuraMinion is the reported bug: on an empty board the bot
// dropped Fang Alpha ("Adjacent minions have +1 Attack") alone, wasting its aura
// (nobody to buff), when it could have developed an equivalent plain body and saved
// the aura minion for a turn where it has a neighbour. With a comparable alternative
// in hand, the bot must develop the plain body first. Encodes WHY: an aura buffing no
// one is idle — its whole value is the neighbours.
func TestPlannerHoldsIdleAuraMinion(t *testing.T) {
	m := aiMatch(1)
	m.state[0].heroHP = 30
	m.state[1].mana, m.state[1].maxMana = 2, 2
	m.state[1].hand = []cards.Card{
		getCard("fang_alpha"),   // 2/2 adjacency aura — idle on an empty board
		getCard("clay_acolyte"), // 2/3 plain body — develop this first
	}

	mv, ok := m.planBest(1)
	if !ok || mv.kind != mPlay {
		t.Fatalf("expected to develop a body, got %+v ok=%v", mv, ok)
	}
	if m.state[1].hand[mv.hand].ID == "fang_alpha" {
		t.Fatal("bot must not drop a lone aura minion when an equivalent body is available")
	}
}

// TestPlannerSequencesAdjacencyOnsetLast is the reported bug: the bot played
// Bannerguard (Onset: give ADJACENT minions +1/+1 & Taunt) FIRST, into an empty
// flank, so its one-shot battlecry buffed nobody — then developed another minion
// beside it that arrived too late to be buffed. An adjacency-battlecry minion must be
// sequenced AFTER the bodies it should buff (playPhase phaseAdjOnset > phasePlay), so
// it lands between them and its onset lands. Encodes WHY: a battlecry fires once, on
// entry — order is the whole value.
func TestPlannerSequencesAdjacencyOnsetLast(t *testing.T) {
	m := aiMatch(1)
	m.state[1].mana, m.state[1].maxMana = 6, 6
	m.state[1].hand = []cards.Card{
		getCard("bannerguard"),  // adjacency-battlecry minion — must go LAST
		getCard("clay_acolyte"), // a plain body — develop this FIRST
	}

	mv, ok := m.planBest(1)
	if !ok || mv.kind != mPlay {
		t.Fatalf("expected to develop a body first, got %+v ok=%v", mv, ok)
	}
	if m.state[1].hand[mv.hand].ID != "clay_acolyte" {
		t.Fatalf("must play the plain body before the adjacency-onset minion, got %s",
			m.state[1].hand[mv.hand].ID)
	}
}

// TestPlannerPlacesAdjacentBufferBetween is the reported bug: a minion whose
// battlecry buffs ADJACENT minions was always dropped at the board edge (the bot
// only ever appended), so only one neighbour got the buff. With two friendly bodies
// on board, the planner must slot the buffer BETWEEN them so BOTH are buffed.
// Encodes WHY: placement is part of a card's value, not a cosmetic — an
// adjacent-buff onset played on the flank wastes half its effect.
func TestPlannerPlacesAdjacentBufferBetween(t *testing.T) {
	m := aiMatch(1)
	m.state[1].mana, m.state[1].maxMana = 4, 4
	m.state[1].board = []*minion{bodyMinion("left", 1, 2, 2), bodyMinion("right", 1, 2, 2)}
	m.state[1].hand = []cards.Card{getCard("bannerguard")} // Onset: adjacent minions +1/+1 & Taunt

	mv, ok := m.planBest(1)
	if !ok || mv.kind != mPlay {
		t.Fatalf("expected to play the adjacent buffer, got %+v ok=%v", mv, ok)
	}
	if mv.pos != 1 {
		t.Fatalf("adjacent buffer must land between the two bodies (pos 1), got pos=%d", mv.pos)
	}

	// And applying it must actually buff BOTH neighbours (2/2 → 3/3).
	mv.applyTo(m, 1) // PlayCardAt takes m.mu itself — don't hold it here
	m.mu.Lock()
	m.refreshAuras()
	var buffed int
	for _, mn := range m.state[1].board {
		if mn.uid == "left" || mn.uid == "right" {
			if mn.atk() == 3 && mn.health == 3 {
				buffed++
			}
		}
	}
	m.mu.Unlock()
	if buffed != 2 {
		t.Fatalf("both flanking minions must be buffed to 3/3, got %d buffed", buffed)
	}
}

// TestPlannerBuffsBeforeAttacking is the reported bug: with a buff spell in hand and
// a minion that can attack, the greedy planner swung the minion FIRST (attacking the
// face scored higher than the buff's static value), then buffed the already-spent
// minion — wasting the buff's Attack for the turn. A value-positive play must be
// chosen ahead of an attack so the buff lands first and its Attack counts. Encodes
// WHY: a spent minion can't re-swing, so buffs belong before combat.
func TestPlannerBuffsBeforeAttacking(t *testing.T) {
	m := aiMatch(1)
	m.state[0].heroHP = 30
	m.state[1].mana, m.state[1].maxMana = 2, 2
	m.state[1].board = []*minion{bodyMinion("beater", 1, 3, 3)} // ready to swing at the face
	m.state[1].hand = []cards.Card{getCard("wild_mark")}        // +2/+2 & Taunt to a minion

	mv, ok := m.planBest(1)
	if !ok {
		t.Fatal("planner found no move; expected the pre-combat buff")
	}
	if mv.kind != mPlay || mv.target != "beater" {
		t.Fatalf("expected to buff the attacker BEFORE swinging, got %+v", mv)
	}
}

// TestPlannerPassesWhenStuck: no mana, no plays, no attackers → end the turn
// (planBest returns found=false).
func TestPlannerPassesWhenStuck(t *testing.T) {
	m := aiMatch(1)
	if _, ok := m.planBest(1); ok {
		t.Fatal("expected no move (should pass the turn)")
	}
}

// TestPlannerKillsBoardWiper is the headline Phase-1 behavior: when the opponent
// has a live ruin_oracle (0/7, "at the start of your turn destroy all minions")
// and the bot is ahead on board, the bot removes the wiper instead of going face.
// Encodes WHY: a pending board wipe erases the bot's board lead next turn, so
// denying it must score above the 7 face damage it gives up to do so.
func TestPlannerKillsBoardWiper(t *testing.T) {
	m := aiMatch(1)
	m.state[0].heroHP = 30
	place(m, 0, "oracle", "ruin_oracle", 0, 7, true)         // enemy turn-start board wipe
	m.state[1].board = []*minion{bodyMinion("axe", 1, 7, 7)} // can kill the 0/7 outright

	mv, ok := m.planBest(1)
	if !ok {
		t.Fatal("planner found no move; expected to kill the board wiper")
	}
	if mv.kind != mAttack || mv.target != "oracle" {
		t.Fatalf("expected attack into the board wiper, got %+v", mv)
	}
}

// TestBoardWipeDevaluesBoardLead encodes the WHY behind Phase 1 directly: an enemy
// minion that wipes the board next turn must make the bot's board lead score LOWER
// than an identical but inert enemy body — that gap is what dampens over-committing
// into the wipe. Same stats on both sides, so the only difference is the trigger.
func TestBoardWipeDevaluesBoardLead(t *testing.T) {
	withWipe := aiMatch(1)
	withWipe.state[0].heroHP, withWipe.state[1].heroHP = 30, 30
	place(withWipe, 0, "oracle", "ruin_oracle", 0, 7, true)
	withWipe.state[1].board = []*minion{bodyMinion("big", 1, 8, 8)}

	inert := aiMatch(1)
	inert.state[0].heroHP, inert.state[1].heroHP = 30, 30
	inert.state[0].board = []*minion{bodyMinion("dummy", 0, 0, 7)} // same 0/7, no trigger
	inert.state[1].board = []*minion{bodyMinion("big", 1, 8, 8)}

	if ws, is := withWipe.scoreForPlanner(1), inert.scoreForPlanner(1); ws >= is {
		t.Fatalf("a pending board wipe must devalue the bot's lead: withWipe=%.2f >= inert=%.2f", ws, is)
	}
}

// TestBotTapsForCardsWhenIdle is the Phase-2 behavior: a Warlock bot with nothing
// that improves the board, comfortable HP, and room in hand uses Life Tap to draw.
// Encodes WHY: tapping never improves the board heuristic (it trades HP for a
// hidden card) so the greedy planner ignores it, yet refilling is the right idle
// play — hence the explicit fallback.
func TestBotTapsForCardsWhenIdle(t *testing.T) {
	m := aiMatch(1)
	m.state[0].heroHP, m.state[1].heroHP = 30, 30
	m.state[1].heroPower = getCard("soul_tithe") // Life Tap: draw + take 2
	m.state[1].heroPowerUsed = false
	m.state[1].mana, m.state[1].maxMana = 2, 2

	if _, ok := m.planBest(1); ok {
		t.Fatal("planner should find no value move (nothing to develop)")
	}
	mv, ok := m.botFallbackHeroPower(1)
	if !ok || mv.kind != mPower {
		t.Fatalf("expected idle Life Tap, got %+v ok=%v", mv, ok)
	}
}

// TestBotWontTapAtLowHP is the safety guard on Phase 2: the bot must not spend
// health to draw when it would drop below lifeTapMinHP. Encodes WHY: digging for
// cards is never worth dying for.
func TestBotWontTapAtLowHP(t *testing.T) {
	m := aiMatch(1)
	m.state[0].heroHP = 30
	m.state[1].heroHP = lifeTapMinHP + 1 // tapping (−2) would drop below the floor
	m.state[1].heroPower = getCard("soul_tithe")
	m.state[1].heroPowerUsed = false
	m.state[1].mana, m.state[1].maxMana = 2, 2

	if _, ok := m.botFallbackHeroPower(1); ok {
		t.Fatal("bot must not Life Tap below the HP floor")
	}
}

// TestDeepPlannerHoldsBackIntoBoardWipe is the headline Phase-3 behavior: with the
// opponent's ruin_oracle live (wipes all minions at the opponent's next turn
// start), the bot should NOT develop a minion that will simply be wiped before it
// does anything — it keeps the card. The 1-ply planner can't see this and would
// develop; the lookahead is exactly what holds it back. Encodes WHY: a play the
// opponent's turn erases is worth no more than not playing it (and costs a card).
func TestDeepPlannerHoldsBackIntoBoardWipe(t *testing.T) {
	mk := func() *Match {
		m := aiMatch(1)
		m.state[0].heroHP, m.state[1].heroHP = 30, 30
		m.state[0].hand, m.state[0].deck = nil, nil // opponent does nothing on its reply turn
		place(m, 0, "oracle", "ruin_oracle", 0, 7, true)
		m.state[1].mana, m.state[1].maxMana = 3, 3
		m.state[1].hand = []cards.Card{getCard("clay_acolyte")} // 2-mana 3/2 that would just die to the wipe
		return m
	}

	// Sanity: the shallow planner develops here — so the test proves the lookahead,
	// not some other deterrent.
	if smv, sok := mk().planBest(1); !sok || smv.kind != mPlay {
		t.Fatalf("sanity: shallow planner should develop here, got %+v ok=%v", smv, sok)
	}
	if mv, ok := mk().planBestDeep(1); ok && mv.kind == mPlay {
		t.Fatalf("deep planner must not develop a minion into a guaranteed board wipe, got %+v", mv)
	}
}

// TestDeepPlannerStillDevelops guards against the lookahead making the bot
// uselessly passive: with no enemy answer in sight, developing a minion is still
// the right play and must beat passing.
func TestDeepPlannerStillDevelops(t *testing.T) {
	m := aiMatch(1)
	m.state[0].heroHP, m.state[1].heroHP = 30, 30
	m.state[0].hand, m.state[0].deck = nil, nil // opponent can't punish the play
	m.state[1].mana, m.state[1].maxMana = 3, 3
	m.state[1].hand = []cards.Card{getCard("clay_acolyte")} // 2-mana 3/2

	mv, ok := m.planBestDeep(1)
	if !ok || mv.kind != mPlay || mv.hand != 0 {
		t.Fatalf("deep planner should still develop a free minion, got %+v ok=%v", mv, ok)
	}
}

// TestDeepPlannerDevelopsIntoAnswerableBoard reproduces the reported bug this fix
// targets: the bot floated mana and hoarded its hand (worst on slow control decks)
// because the 2-ply lookahead demanded a play beat the post-opponent-reply "pass"
// baseline — and a healthy minion the opponent can later trade with nets ≈ neutral
// after the reply, so passing won. Here the opponent has a body that can trade into
// the developed minion (but NO board wipe), so developing is correct tempo and must
// still be chosen. Encodes WHY: a card-for-card trade is not a reason to sit on the
// turn — only a play the opponent's turn erases for free (a wipe) is.
func TestDeepPlannerDevelopsIntoAnswerableBoard(t *testing.T) {
	m := aiMatch(1)
	m.state[0].heroHP, m.state[1].heroHP = 30, 30
	m.state[0].board = []*minion{bodyMinion("answer", 0, 3, 3)} // can trade into the play
	m.state[0].hand, m.state[0].deck = nil, nil                 // no wipe, no other punish
	m.state[1].mana, m.state[1].maxMana = 3, 3
	m.state[1].hand = []cards.Card{getCard("clay_acolyte")} // 2-mana 3/2

	mv, ok := m.planBestDeep(1)
	if !ok || mv.kind != mPlay || mv.hand != 0 {
		t.Fatalf("deep planner must develop even when the opponent can trade, got %+v ok=%v", mv, ok)
	}
}

// TestBuffGatedOnSurvivalButDevelopIsnt reproduces a reported bug and the asymmetry
// that fixes it: the bot played a minion and poured a buff onto it, then the opponent
// killed it — a 2-for-1 (buff card + minion for one removal). A buff spell leaves no
// body of its own, so its whole value rides on the target surviving: the deep planner
// must decline a buff whose post-reply score doesn't beat simply passing
// (`deepEligible` false). A plain minion development must NOT be gated the same way —
// a body that merely trades is a fair 1-for-1, and gating it is exactly what made the
// bot hoard its hand and float mana. This encodes both halves of the rule.
func TestBuffGatedOnSurvivalButDevelopIsnt(t *testing.T) {
	const pass = 5.0

	buffM := aiMatch(1)
	buffM.state[1].hand = []cards.Card{getCard("wild_mark")} // a buff spell
	buff := aiMove{kind: mPlay, hand: 0, target: "runt"}
	if !buffM.isEnhancementPlay(1, buff) {
		t.Fatal("wild_mark is a buff spell — must be classed as an enhancement play")
	}
	if buffM.deepEligible(1, buff, pass, pass) { // score == pass baseline → didn't pay off
		t.Fatal("a buff that doesn't beat passing (target dies) must be ineligible")
	}
	if !buffM.deepEligible(1, buff, pass+1, pass) { // beats pass → target survived, worth it
		t.Fatal("a buff that beats passing (target survives) must be eligible")
	}

	devM := aiMatch(1)
	devM.state[1].hand = []cards.Card{getCard("clay_acolyte")} // a minion (a body)
	dev := aiMove{kind: mPlay, hand: 0}
	if devM.isEnhancementPlay(1, dev) {
		t.Fatal("a minion play leaves a body — it is not an enhancement play")
	}
	if !devM.deepEligible(1, dev, pass, pass) { // even at the baseline, develop must stay eligible
		t.Fatal("a plain development must not be gated on beating pass (no hoarding)")
	}
}

// TestDeepPlannerStillFightsWhenLosing reproduces a reported bug: facing lethal
// next turn, the 2-ply lookahead concluded "I die in every line" and passed —
// declining an available trade (and its hero power). When in lethal danger the bot
// must defer to the shallow lethal-lens planner and still fight (trade down the
// threat) rather than give up. Encodes WHY: a losing position is not a reason to
// stop playing — chip the opponent's burst and play for the out.
func TestDeepPlannerStillFightsWhenLosing(t *testing.T) {
	m := aiMatch(1)
	m.state[1].heroHP = 17
	m.state[1].board = []*minion{bodyMinion("lion", 1, 6, 5)} // bot's only body
	// Opponent has lethal on board next turn (8+8+8+2 = 26 attack vs 17 HP).
	m.state[0].board = []*minion{
		bodyMinion("crag", 0, 8, 6), // killable by the 6/5 (6 ≥ 6) → removes 8 burst
		bodyMinion("magma1", 0, 8, 8),
		bodyMinion("magma2", 0, 8, 8),
		bodyMinion("ward", 0, 2, 1),
	}

	mv, ok := m.planBestDeep(1)
	if !ok {
		t.Fatal("bot must not pass when it can still trade down a lethal threat")
	}
	if mv.kind != mAttack {
		t.Fatalf("expected a defensive trade, got %+v", mv)
	}
}

// TestDeepPlannerUsesFreeHeroPower reproduces a reported bug: while behind (but not
// facing lethal), the deep planner passed its whole turn — not even firing a free,
// Taunt-ignoring face hero power (Hunter's Quick Shot, deal 2). Cause: each
// candidate's opponent-reply sim drew different cards, and that variance buried the
// clean +2 signal so "pass" won by luck. With one shared seed per turn the reply is
// identical across candidates, so the hero power's +2 wins deterministically.
// Encodes WHY: a guaranteed free 2 to the face is strictly progress — never a pass.
func TestDeepPlannerUsesFreeHeroPower(t *testing.T) {
	m := aiMatch(1)
	m.state[0].heroHP, m.state[1].heroHP = 20, 25
	m.state[1].heroPower = getCard("quick_shot") // deal 2 to the enemy hero
	m.state[1].heroPowerUsed = false
	m.state[1].mana, m.state[1].maxMana = 2, 2
	m.state[0].board = []*minion{bodyMinion("ogre", 0, 6, 7)} // behind, but 6 < 25 = not lethal

	mv, ok := m.planBestDeep(1)
	if !ok {
		t.Fatal("bot must not pass when a free face hero power is available")
	}
	if mv.kind != mPower {
		t.Fatalf("expected the free hero power, got %+v", mv)
	}
}

// TestDeathrattleValuedInEval encodes Phase 4: a minion with a summon Final Gasp is
// worth more to its controller than the same body silenced, because its death
// hands the owner a token. That standing value is what makes the bot prefer to
// SILENCE / transform such a minion rather than only ever trade into it. (Trading
// into a deathrattle is already priced by simulation — the death fires on the
// clone — so this term is about the un-fired effect.)
func TestDeathrattleValuedInEval(t *testing.T) {
	m, _, _ := newMatch()
	live := &minion{uid: "dr", card: getCard("reaper_golem"), owner: 0, health: 3} // FG: summon a 2/1
	gagged := &minion{uid: "dr2", card: getCard("reaper_golem"), owner: 0, health: 3, silenced: true}

	if lv, gv := m.minionValue(live), m.minionValue(gagged); lv <= gv {
		t.Fatalf("a live deathrattle minion must be worth more than a silenced one: live=%.2f silenced=%.2f", lv, gv)
	}
}

// TestPlannerDoesntWasteCoin reproduces a reported bug: with Arcane Wyrmling out
// (gains +1 Attack on any spell cast) and only the Coin in hand, the bot cast the
// Coin just to trigger the buff — wasting the ramp with nothing to spend it on. A
// mana-ramp card must not be a candidate unless its mana unlocks another play.
func TestPlannerDoesntWasteCoin(t *testing.T) {
	m := aiMatch(1)
	m.state[1].mana, m.state[1].maxMana = 1, 1
	place(m, 1, "wyrm", "arcane_wyrmling", 1, 3, true) // casting any spell would buff it
	m.state[1].hand = []cards.Card{getCard("mana_surge")}

	for _, mv := range m.aiCandidates(1) {
		if mv.kind == mPlay && m.state[1].hand[mv.hand].ID == "mana_surge" {
			t.Fatal("Coin must not be a candidate when its mana unlocks no play")
		}
	}
}

// TestPlannerCoinsToEnablePlay is the complement: when the extra mana makes an
// otherwise-unaffordable card playable, the Coin IS offered.
func TestPlannerCoinsToEnablePlay(t *testing.T) {
	m := aiMatch(1)
	m.state[1].mana, m.state[1].maxMana = 1, 1
	// clay_acolyte costs 2; at 1 mana it's unaffordable, the Coin (→2) unlocks it.
	m.state[1].hand = []cards.Card{getCard("mana_surge"), getCard("clay_acolyte")}

	found := false
	for _, mv := range m.aiCandidates(1) {
		if mv.kind == mPlay && m.state[1].hand[mv.hand].ID == "mana_surge" {
			found = true
		}
	}
	if !found {
		t.Fatal("Coin should be a candidate when its mana unlocks an otherwise-unaffordable card")
	}
}

// waitFor polls cond (under the match lock) until true or the deadline, failing
// the test on timeout. Used to await the async bot turn without a fixed sleep.
func waitFor(t *testing.T, m *Match, what string, cond func() bool) {
	t.Helper()
	for i := 0; i < 400; i++ { // ~2s at 5ms
		m.mu.Lock()
		ok := cond()
		m.mu.Unlock()
		if ok {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for: %s", what)
}

// TestVsAIFullFlow drives a real vs-AI match end to end: bot auto-mulligans, and
// after the human passes turn 1 the bot takes its turn off-thread — developing a
// minion and handing the turn back — with no deadlock or panic. Run under -race
// to exercise the goroutine + match-mutex path.
func TestVsAIFullFlow(t *testing.T) {
	botActionDelay = 0 // no pacing in tests
	defer func() { botActionDelay = 700 * time.Millisecond }()

	human := &fakeSender{id: "human"}
	bot := botSender{id: "bot", name: "AI"}
	// 1-cost minions so the bot can develop on its very first turn (1 mana).
	m := New("ai1", human, bot, 1, deck30("pebble_imp"), deck30("pebble_imp"))
	m.EnableAI(1, 1)
	m.Start()
	m.driveBotMulligan() // bot keeps its hand (async)
	if ok, _ := m.Mulligan(human, nil); !ok {
		t.Fatal("human mulligan failed")
	}
	waitFor(t, m, "play to begin", func() bool { return m.mulligan == nil })

	// Turn 1 is the human's — just pass.
	waitFor(t, m, "human's turn", func() bool { return m.turn == 0 })
	if ok, msg := m.EndTurn(human); !ok {
		t.Fatalf("human EndTurn failed: %s", msg)
	}

	// Bot now takes its first turn off-thread: it should develop a 1-drop and pass.
	waitFor(t, m, "turn back to human", func() bool { return m.turn == 0 && !m.over })
	m.mu.Lock()
	botBoard := len(m.state[1].board)
	m.mu.Unlock()
	if botBoard < 1 {
		t.Fatalf("bot should have developed at least one minion, board=%d", botBoard)
	}
}
