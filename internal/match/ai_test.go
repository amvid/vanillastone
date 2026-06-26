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

// TestPlannerPassesWhenStuck: no mana, no plays, no attackers → end the turn
// (planBest returns found=false).
func TestPlannerPassesWhenStuck(t *testing.T) {
	m := aiMatch(1)
	if _, ok := m.planBest(1); ok {
		t.Fatal("expected no move (should pass the turn)")
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
