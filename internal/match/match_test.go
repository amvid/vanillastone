package match

import (
	"encoding/json"
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/protocol"
)

// place builds a minion instance on the given player's board for white-box
// tests, mirroring how summonMinion would set it up. ready=true means it can act
// this turn (not summon-sick); aegis is taken from the card's keyword.
func place(m *Match, owner int, uid, cardID string, attack, health int, ready bool) {
	c := getCard(cardID)
	// Attack/maxHealth are derived from base card stats + enchantments, so express
	// the requested attack/health (maxHealth == health) as a single enchantment
	// delta off the card's printed stats.
	mn := &minion{
		uid: uid, card: c, owner: owner,
		enchants:         []enchant{{atk: attack - c.Attack, hp: health - c.Health}},
		health:           health,
		summonedThisTurn: !ready,
		aegis:            c.Has(cards.KeywordAegis),
		stealthed:        c.Has(cards.KeywordStealth),
	}
	mn.attacksMade = 0
	m.state[owner].board = append(m.state[owner].board, mn)
}

// placeSecret puts an active secret into a player's hidden zone for white-box
// tests, mirroring playSecret without the hand/mana bookkeeping.
func placeSecret(m *Match, owner int, cardID string) {
	c := getCard(cardID)
	m.nextUID++
	m.state[owner].secrets = append(m.state[owner].secrets, &secretInst{uid: uid(m.nextUID), card: c, owner: owner})
}

// fakeSender records every message the match pushes to a player, so tests can
// inspect the authoritative snapshot the server would send.
type fakeSender struct {
	id   string
	msgs [][]byte
}

func (f *fakeSender) ID() string    { return f.id }
func (f *fakeSender) Name() string  { return f.id }
func (f *fakeSender) Send(b []byte) { f.msgs = append(f.msgs, append([]byte(nil), b...)) }

// lastState decodes the most recent "state" (or "match_start") snapshot sent to
// this player. Both carry the same Self/Opp shape.
func lastState(t *testing.T, f *fakeSender) protocol.State {
	t.Helper()
	for i := len(f.msgs) - 1; i >= 0; i-- {
		var env protocol.Envelope
		json.Unmarshal(f.msgs[i], &env)
		if env.Type == protocol.TypeState || env.Type == protocol.TypeMatchStart {
			var st protocol.State
			json.Unmarshal(f.msgs[i], &st)
			return st
		}
	}
	t.Fatal("no snapshot sent")
	return protocol.State{}
}

func gameOverWinner(f *fakeSender) (string, bool) {
	for _, b := range f.msgs {
		var env protocol.Envelope
		json.Unmarshal(b, &env)
		if env.Type == protocol.TypeGameOver {
			var go_ protocol.GameOver
			json.Unmarshal(b, &go_)
			return go_.Winner, true
		}
	}
	return "", false
}

// testHand is the deterministic opening hand the white-box tests rely on (the
// fixed 36-card set used before real decks/mulligan landed). Card indices below
// (idxPebbleImp, idxThicket) point into it.
var testHand = []string{
	"pebble_imp", "cinder_bolt", "spark_adept", "ember_striker", "brood_mother",
	"whetstone", "thicket_stalker", "bog_warden", "volatile_wisp", "cinder_husk",
	"mend", "quake", "clay_acolyte", "granite_watcher", "bastion_golem",
	"swift_raptor", "lurking_stalker", "gilded_sentry", "frost_snap", "permafrost",
	"gale_harrier", "veil_stalker", "toxic_fang", "bloodthorn_knight", "ember_scribe",
	"pack_leader", "hush", "drain_touch", "snare", "mimic", "nullify",
	"arcane_insight", "wild_summons", "ember_cleaver", "quartz_spike", "frost_ward",
}

func newMatch() (*Match, *fakeSender, *fakeSender) {
	a, b := &fakeSender{id: "p1"}, &fakeSender{id: "p2"}
	deck := testDeck(testHand)
	m := New("m1", a, b, 1, deck, deck) // fixed seed: deterministic random-target effects
	// Existing white-box tests predate decks/mulligan: install the fixed opening
	// hand, skip the mulligan phase, and open turn 1 with 1 mana and no draw (a
	// small deck remains so later end-of-turn draws don't immediately fatigue).
	m.mulligan = nil
	for pi := 0; pi < 2; pi++ {
		m.state[pi].hand = testDeck(testHand)
		m.state[pi].deck = testDeck(testHand)
	}
	m.state[0].maxMana, m.state[0].mana = 1, 1
	m.sendStateAll()
	return m, a, b
}

// TestTurnTimer: a turn auto-ends when its timer fires, and a stale timer (from
// a turn that already passed) is ignored. This matters because the timer drives
// an authoritative forfeit-of-turn — it must pass exactly the active turn once,
// never an already-ended one.
func TestTurnTimer(t *testing.T) {
	m, _, _ := newMatch()
	// Real schedule arms the deadline + countdown without firing in-test.
	m.turn = 0
	m.scheduleTurnTimer()
	if s := m.turnSecondsLeft(); s <= 0 || s > 75 {
		t.Fatalf("turnSecondsLeft after schedule = %d, want 1..75", s)
	}
	gen := m.turnGen

	// Drive the callback directly. turnDuration=0 first so the auto-end's
	// startTurn doesn't re-arm a chaining timer during the test.
	m.turnDuration = 0

	// A stale-generation callback must do nothing.
	m.onTurnTimeout(gen - 1)
	if m.turn != 0 || m.over {
		t.Fatalf("stale turn timer must be ignored (turn=%d over=%v)", m.turn, m.over)
	}

	// The current-generation callback passes the turn.
	before := m.turnNum
	m.onTurnTimeout(gen)
	if m.turn != 1 {
		t.Fatalf("turn timer should pass the turn, got turn=%d", m.turn)
	}
	if m.turnNum != before+1 {
		t.Fatalf("turnNum should increment on auto-end, got %d", m.turnNum)
	}
	if m.turnSecondsLeft() != 0 {
		t.Fatalf("no timer should be armed (turnDuration=0), got %d", m.turnSecondsLeft())
	}
}

// TestTurnTimerStop: stopping the timer clears the countdown.
func TestTurnTimerStop(t *testing.T) {
	m, _, _ := newMatch()
	m.scheduleTurnTimer()
	m.stopTurnTimer()
	if got := m.turnSecondsLeft(); got != 0 {
		t.Fatalf("stopTurnTimer should clear the deadline, got %d", got)
	}
}

// handIndex finds the position of cardID in the player's own hand, -1 if absent.
func handIndex(st protocol.State, cardID string) int {
	for i, c := range st.Self.Hand {
		if c.CardID == cardID {
			return i
		}
	}
	return -1
}

// Opening-hand positions used directly. Pebble Imp (1-cost minion) is the
// cheapest play; Thicket Stalker (3-cost minion) is unaffordable on turn 1.
const (
	idxPebbleImp = 0 // pebble_imp, 1 mana
	idxThicket   = 6 // thicket_stalker, 3 mana
)

// TestManaGate: a player can only play what their mana ramp affords. On turn 1
// (1 mana) a 3-cost minion is illegal but the 1-cost is fine. This matters
// because mana is the core tempo constraint — without it any board is instant.
func TestManaGate(t *testing.T) {
	m, a, _ := newMatch()
	if ok, msg := m.PlayCard(a, idxThicket, ""); ok {
		t.Fatalf("3-cost minion should be illegal at 1 mana")
	} else if msg != "not enough mana" {
		t.Fatalf("want 'not enough mana', got %q", msg)
	}
	if ok, msg := m.PlayCard(a, idxPebbleImp, ""); !ok {
		t.Fatalf("1-cost minion should be playable at 1 mana: %s", msg)
	}
	st := lastState(t, a)
	if st.Self.Mana != 0 {
		t.Fatalf("mana should be spent to 0, got %d", st.Self.Mana)
	}
	if len(st.Self.Board) != 1 {
		t.Fatalf("minion should be on board, got %d", len(st.Self.Board))
	}
}

// TestSummonSickness: a minion cannot attack the turn it is summoned, but can on
// the owner's next turn. This is the rule that stops drop-and-swing burst.
func TestSummonSickness(t *testing.T) {
	m, a, b := newMatch()
	m.PlayCard(a, idxPebbleImp, "")
	st := lastState(t, a)
	imp := st.Self.Board[0].InstanceID
	if st.Self.Board[0].CanAttack {
		t.Fatalf("freshly summoned minion must be summon-sick")
	}
	if ok, msg := m.Attack(a, imp, oppHeroTarget); ok {
		t.Fatalf("summon-sick minion should not attack")
	} else if msg != "minion cannot attack" {
		t.Fatalf("want 'minion cannot attack', got %q", msg)
	}
	// Round-trip the turn back to player a.
	m.EndTurn(a)
	m.EndTurn(b)
	if !lastState(t, a).Self.Board[0].CanAttack {
		t.Fatalf("minion should wake on owner's next turn")
	}
}

// TestCombatTrade: attacking exchanges damage simultaneously and both lethal
// minions die. Encodes that combat is mutual, not one-sided.
func TestCombatTrade(t *testing.T) {
	m, a, b := newMatch()
	// a plays a 1/1, ends; b plays a 1/1, ends; now a's 1/1 can attack b's 1/1.
	m.PlayCard(a, idxPebbleImp, "")
	m.EndTurn(a)
	m.PlayCard(b, idxPebbleImp, "")
	m.EndTurn(b)

	aImp := lastState(t, a).Self.Board[0].InstanceID
	bImp := lastState(t, a).Opp.Board[0].InstanceID
	if ok, msg := m.Attack(a, aImp, bImp); !ok {
		t.Fatalf("1/1 vs 1/1 attack should resolve: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Self.Board) != 0 || len(st.Opp.Board) != 0 {
		t.Fatalf("both 1/1s should die in the trade: self=%d opp=%d",
			len(st.Self.Board), len(st.Opp.Board))
	}
}

// TestOffTurnRejected: the off-turn player cannot play cards. Server authority
// over turn order is the whole point of Phase 1/2.
func TestOffTurnRejected(t *testing.T) {
	m, _, b := newMatch()
	if ok, msg := m.PlayCard(b, idxPebbleImp, ""); ok {
		t.Fatalf("off-turn play should be rejected")
	} else if msg != "not your turn" {
		t.Fatalf("want 'not your turn', got %q", msg)
	}
}

// TestBoardCap: no more than seven minions per side. With the board full, a
// play is rejected even though the player has the mana for it. (White-box: the
// opening hand can't legally fill 7 slots in one turn, so we fill the board
// directly and exercise the cap check.)
func TestBoardCap(t *testing.T) {
	m, a, _ := newMatch() // a's turn, 1 mana — enough for the 1-cost imp.
	ps := m.state[0]
	for len(ps.board) < maxBoard {
		ps.board = append(ps.board, &minion{uid: "x"})
	}
	if ok, msg := m.PlayCard(a, idxPebbleImp, ""); ok {
		t.Fatalf("play beyond board cap should be rejected")
	} else if msg != "board full" {
		t.Fatalf("want 'board full', got %q", msg)
	}
}

// --- Phase 3: spells + targeting ---

// TestSpellDamageKillsMinion: a targeted damage spell destroys an enemy minion.
func TestSpellDamageKillsMinion(t *testing.T) {
	m, a, b := newMatch()
	m.EndTurn(a)
	m.PlayCard(b, idxPebbleImp, "") // b summons a 1/1
	m.EndTurn(b)                    // back to a, now 2 mana

	bImp := lastState(t, a).Opp.Board[0].InstanceID
	bolt := handIndex(lastState(t, a), "cinder_bolt")
	if ok, msg := m.PlayCard(a, bolt, bImp); !ok {
		t.Fatalf("cinder bolt on enemy minion should resolve: %s", msg)
	}
	if n := len(lastState(t, a).Opp.Board); n != 0 {
		t.Fatalf("3-dmg bolt should kill the 1/1, board has %d", n)
	}
}

// TestSpellTargetRequired: a targeted spell with no valid target is rejected,
// and crucially does NOT consume mana or the card (validation precedes mutation).
func TestSpellTargetRequired(t *testing.T) {
	m, a, b := newMatch()
	m.EndTurn(a)
	m.EndTurn(b) // a's turn, 2 mana, empty boards

	before := lastState(t, a)
	bolt := handIndex(before, "cinder_bolt")
	if ok, msg := m.PlayCard(a, bolt, ""); ok {
		t.Fatalf("targeted spell with no target should be rejected")
	} else if msg != "no such target" {
		t.Fatalf("want 'no such target', got %q", msg)
	}
	after := lastState(t, a)
	if after.Self.Mana != before.Self.Mana || after.Self.HandCount != before.Self.HandCount {
		t.Fatalf("rejected spell must not spend mana or card: mana %d->%d hand %d->%d",
			before.Self.Mana, after.Self.Mana, before.Self.HandCount, after.Self.HandCount)
	}
}

// TestSpellIllegalTarget: a friendly-minion-only buff cannot hit the enemy hero.
func TestSpellIllegalTarget(t *testing.T) {
	m, a, _ := newMatch() // a's turn, 1 mana; Whetstone costs 1
	whet := handIndex(lastState(t, a), "whetstone")
	if ok, msg := m.PlayCard(a, whet, oppHeroTarget); ok {
		t.Fatalf("friendly buff on enemy hero should be rejected")
	} else if msg != "illegal target" {
		t.Fatalf("want 'illegal target', got %q", msg)
	}
}

// TestBuffRaisesStats: a buff adds attack and health (and max health).
func TestBuffRaisesStats(t *testing.T) {
	m, a, b := newMatch()
	m.PlayCard(a, idxPebbleImp, "") // 1/1 on board, mana now 0
	m.EndTurn(a)
	m.EndTurn(b) // a's turn, 2 mana, imp present

	imp := lastState(t, a).Self.Board[0].InstanceID
	whet := handIndex(lastState(t, a), "whetstone")
	if ok, msg := m.PlayCard(a, whet, imp); !ok {
		t.Fatalf("whetstone on friendly minion should resolve: %s", msg)
	}
	mv := lastState(t, a).Self.Board[0]
	if mv.Attack != 3 || mv.Health != 2 || mv.MaxHealth != 2 {
		t.Fatalf("1/1 + (2/+1) should be 3/2 (max 2), got %d/%d (max %d)", mv.Attack, mv.Health, mv.MaxHealth)
	}
}

// TestHealCapsAtMax: healing never overshoots the cap. White-box-damage the hero
// first, then over-heal and assert it caps.
func TestHealCapsAtMax(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 28 // wounded by 2; Mend heals 4
	mend := handIndex(lastState(t, a), "mend")
	if ok, msg := m.PlayCard(a, mend, selfHeroTarget); !ok {
		t.Fatalf("mend on own hero should resolve: %s", msg)
	}
	if hp := lastState(t, a).Self.HeroHP; hp != heroMaxHP {
		t.Fatalf("heal should cap at %d, got %d", heroMaxHP, hp)
	}
}

// TestQuakeHitsAllEnemyMinions: an untargeted AoE damages every enemy minion and
// kills those that drop to 0, leaving survivors and friendly minions untouched.
func TestQuakeHitsAllEnemyMinions(t *testing.T) {
	m, a, b := newMatch()
	m.EndTurn(a)
	m.PlayCard(b, idxPebbleImp, "") // b: 1/1
	m.EndTurn(b)
	m.EndTurn(a)
	gw := handIndex(lastState(t, b), "granite_watcher")
	m.PlayCard(b, gw, "") // b: + 2/3 (mana 2 on b's 2nd turn)
	m.EndTurn(b)          // a's turn, 3 mana

	quake := handIndex(lastState(t, a), "quake")
	if ok, msg := m.PlayCard(a, quake, ""); !ok {
		t.Fatalf("quake (untargeted) should resolve: %s", msg)
	}
	board := lastState(t, a).Opp.Board
	if len(board) != 1 || board[0].Health != 2 {
		t.Fatalf("quake should kill the 1/1 and leave the 2/3 at 2 health, got %+v", board)
	}
}

// TestSpellToFaceWins: lethal spell damage to the enemy hero ends the match with
// the caster as winner. (White-box: lower the hero to spell range first.)
func TestSpellToFaceWins(t *testing.T) {
	m, a, b := newMatch()
	m.EndTurn(a)
	m.EndTurn(b)          // a's turn, 2 mana
	m.state[1].heroHP = 2 // opponent at 2; Cinder Bolt deals 3
	bolt := handIndex(lastState(t, a), "cinder_bolt")
	if ok, msg := m.PlayCard(a, bolt, oppHeroTarget); !ok {
		t.Fatalf("bolt to face should resolve: %s", msg)
	}
	winner, over := gameOverWinner(a)
	if !over || winner != a.ID() {
		t.Fatalf("lethal spell should win for %s: over=%v winner=%s", a.ID(), over, winner)
	}
	_ = b
}

// TestHeroDamageWins: hammering the opponent hero to 0 with minions ends the
// match with the attacker as winner. This is the core combat win condition.
func TestHeroDamageWins(t *testing.T) {
	m, a, b := newMatch()
	// On each a-turn, summon the first affordable minion and swing all awake
	// minions at the enemy hero. b passes. Deterministic; bounded.
	for half := 0; half < 60; half++ {
		if _, over := gameOverWinner(a); over {
			break
		}
		st := lastState(t, a)
		if st.Turn == a.ID() {
			for i, c := range st.Self.Hand {
				if c.CardType == "minion" && c.Cost <= st.Self.Mana {
					// oppHero is ignored by minions with no targeted onset;
					// for those that have one it just sends the onset face.
					m.PlayCard(a, i, oppHeroTarget)
					break
				}
			}
			for _, mn := range lastState(t, a).Self.Board {
				if mn.CanAttack {
					m.Attack(a, mn.InstanceID, oppHeroTarget)
				}
			}
			if _, over := gameOverWinner(a); over {
				break
			}
			m.EndTurn(a)
		} else {
			m.EndTurn(b)
		}
	}
	winner, over := gameOverWinner(a)
	if !over {
		t.Fatalf("game should end when hero reaches 0; opp HP=%d", lastState(t, a).Opp.HeroHP)
	}
	if winner != a.ID() {
		t.Fatalf("winner should be the attacker %s, got %s", a.ID(), winner)
	}
}

// TestConcede verifies a player can forfeit on the opponent's turn: conceding
// ends the match immediately with the opponent as winner, and further actions
// are rejected. Encodes the intent that concede is a surrender, not a turn play.
func TestConcede(t *testing.T) {
	m, a, b := newMatch()

	// It is a's turn (first player). b concedes off-turn — still valid.
	if ok, msg := m.Concede(b); !ok {
		t.Fatalf("off-turn concede should be allowed: %q", msg)
	}
	winner, over := gameOverWinner(b)
	if !over {
		t.Fatal("concede should end the match")
	}
	if winner != a.ID() {
		t.Fatalf("conceding player loses; winner should be %s, got %s", a.ID(), winner)
	}
	// Match is over: another concede is rejected.
	if ok, _ := m.Concede(a); ok {
		t.Fatal("concede after game over should be rejected")
	}
}

// --- Phase 4: triggers (onset / finalGasp) + event log ---

// TestOnsetHitsFace: a targeted onset deals its damage to the chosen
// character. Encodes that on_play effects resolve like a spell at summon time.
func TestOnsetHitsFace(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2 // afford Spark Adept (2)
	spark := handIndex(lastState(t, a), "spark_adept")
	if ok, msg := m.PlayCard(a, spark, oppHeroTarget); !ok {
		t.Fatalf("spark adept onset to face should resolve: %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.HeroHP != 28 {
		t.Fatalf("onset should deal 2 to enemy hero (30->28), got %d", st.Opp.HeroHP)
	}
	if len(st.Self.Board) != 1 || st.Self.Board[0].Name != "Spark Adept" {
		t.Fatalf("the minion should still be summoned, board=%+v", st.Self.Board)
	}
}

// TestOnsetKillsMinion: a targeted onset can kill an enemy minion, and
// the death resolves in the same action.
func TestOnsetKillsMinion(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	place(m, 1, "victim", "pebble_imp", 1, 1, false) // enemy 1/1
	spark := handIndex(lastState(t, a), "spark_adept")
	if ok, msg := m.PlayCard(a, spark, "victim"); !ok {
		t.Fatalf("spark adept on enemy minion should resolve: %s", msg)
	}
	if n := len(lastState(t, a).Opp.Board); n != 0 {
		t.Fatalf("2-dmg onset should kill the 1/1, board has %d", n)
	}
}

// TestOnsetRequiresTarget: a targeted onset with a legal target
// available but none chosen is rejected, and (validation precedes mutation) does
// not spend mana or the card.
func TestOnsetRequiresTarget(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	before := lastState(t, a)
	spark := handIndex(before, "spark_adept")
	if ok, msg := m.PlayCard(a, spark, ""); ok {
		t.Fatalf("onset needing a target must be rejected with none chosen")
	} else if msg != "no such target" {
		t.Fatalf("want 'no such target', got %q", msg)
	}
	after := lastState(t, a)
	if after.Self.Mana != before.Self.Mana || after.Self.HandCount != before.Self.HandCount {
		t.Fatalf("rejected play must not spend mana/card: mana %d->%d hand %d->%d",
			before.Self.Mana, after.Self.Mana, before.Self.HandCount, after.Self.HandCount)
	}
}

// TestOnsetFizzlesWithNoTarget: an enemy-minion onset with no enemy
// minions still summons the minion; the onset simply does nothing. (This is
// the "play a onset card without a target" case.)
func TestOnsetFizzlesWithNoTarget(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2 // Ember Striker costs 2; empty enemy board
	ember := handIndex(lastState(t, a), "ember_striker")
	if ok, msg := m.PlayCard(a, ember, ""); !ok {
		t.Fatalf("onset with no legal target should still play (fizzle): %s", msg)
	}
	st := lastState(t, a)
	if len(st.Self.Board) != 1 || st.Self.Board[0].Name != "Ember Striker" {
		t.Fatalf("minion should be summoned despite fizzled onset, board=%+v", st.Self.Board)
	}
	if st.Self.Mana != 0 {
		t.Fatalf("mana should be spent (2->0), got %d", st.Self.Mana)
	}
}

// TestOnsetSummon: an untargeted summon onset puts a token alongside the
// played minion.
func TestOnsetSummon(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 3, 3 // Bog Warden costs 3
	bog := handIndex(lastState(t, a), "bog_warden")
	if ok, msg := m.PlayCard(a, bog, ""); !ok {
		t.Fatalf("bog warden onset should resolve: %s", msg)
	}
	board := lastState(t, a).Self.Board
	if len(board) != 2 {
		t.Fatalf("bog warden + summoned bogling = 2 minions, got %d", len(board))
	}
}

// TestOnsetDrawsCard: a onset draw effect pulls a card from the
// controller's deck into hand (the played minion itself having left the hand).
func TestOnsetDrawsCard(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 5, 5
	drake, _ := cards.Get("sapphire_drake") // Onset: Draw a card.
	m.state[0].hand = []cards.Card{drake}
	m.state[0].deck = testDeck([]string{"pebble_imp", "granite_watcher"})
	m.sendStateAll()
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("playing the drake should resolve: %s", msg)
	}
	// Hand: 1 (drake) -> 0 (played) -> 1 (onset draw). Deck: 2 -> 1.
	if got := len(m.state[0].hand); got != 1 {
		t.Fatalf("onset should draw 1 card, hand=%d", got)
	}
	if got := len(m.state[0].deck); got != 1 {
		t.Fatalf("draw should remove 1 from the deck, deck=%d", got)
	}
}

// TestFinalGaspSummon: a minion's death fires its on_death effect (summon a
// token). The token replaces it on the board.
func TestFinalGaspSummon(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "granite_watcher", 2, 2, true) // a's 2/2 attacker, ready
	place(m, 1, "bm", "brood_mother", 2, 1, false)    // enemy Brood Mother at 1 health
	if ok, msg := m.Attack(a, "atk", "bm"); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	board := lastState(t, a).Opp.Board
	if len(board) != 1 || board[0].Name != "Broken Golem" {
		t.Fatalf("finalGasp should summon a Broken Golem, board=%+v", board)
	}
}

// TestFinalGaspDamagesHero: Cinder Husk's finalGasp deals 2 to the enemy
// hero (the enemy of the dying minion's owner) when it dies.
func TestFinalGaspDamagesHero(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "granite_watcher", 2, 2, true)
	place(m, 1, "husk", "cinder_husk", 3, 1, false) // enemy husk at 1 health
	if ok, msg := m.Attack(a, "atk", "husk"); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	if hp := lastState(t, a).Self.HeroHP; hp != 28 {
		t.Fatalf("husk finalGasp should deal 2 to a's hero (30->28), got %d", hp)
	}
}

// TestFinalGaspRandomDamage: Volatile Wisp's finalGasp deals 1 to a random
// enemy character. Whatever the RNG picks, exactly 1 total health is removed
// from the enemy side.
func TestFinalGaspRandomDamage(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "surv", "granite_watcher", 2, 3, false) // a survivor: hero + minion = 2 candidates
	place(m, 1, "wisp", "volatile_wisp", 1, 1, false)
	// a side total health before (hero + minions).
	before := m.state[0].heroHP
	for _, mn := range m.state[0].board {
		before += mn.health
	}
	// Kill the wisp (white-box) and resolve: its finalGasp fires.
	m.state[1].board[0].health = 0
	m.resetLog()
	m.finish()
	after := m.state[0].heroHP
	for _, mn := range m.state[0].board {
		after += mn.health
	}
	if before-after != 1 {
		t.Fatalf("random finalGasp should remove exactly 1 from enemy side, before=%d after=%d", before, after)
	}
}

// TestFinalGaspsFireOnSimultaneousDeaths: an AoE that kills several minions at
// once fires every finalGasp. Quake (1 to all enemy minions) wipes two
// 1-health enemies; both finalGasps resolve.
func TestFinalGaspsFireOnSimultaneousDeaths(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 3, 3        // afford Quake
	place(m, 1, "wisp", "volatile_wisp", 1, 1, false) // dies -> 1 to a (no a minions -> hero)
	place(m, 1, "husk", "cinder_husk", 3, 1, false)   // dies -> 2 to a hero
	quake := handIndex(lastState(t, a), "quake")
	if ok, msg := m.PlayCard(a, quake, ""); !ok {
		t.Fatalf("quake should resolve: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Opp.Board) != 0 {
		t.Fatalf("both 1-health enemies should die, board=%+v", st.Opp.Board)
	}
	if st.Self.HeroHP != 27 { // 30 - 1 (wisp) - 2 (husk)
		t.Fatalf("both finalGasps should hit a's hero (30->27), got %d", st.Self.HeroHP)
	}
}

// TestEventLogEmitted: actions carry an ordered event log. Summoning a minion
// emits a summon event targeting that minion.
func TestEventLogEmitted(t *testing.T) {
	m, a, _ := newMatch()
	if ok, msg := m.PlayCard(a, idxPebbleImp, ""); !ok {
		t.Fatalf("pebble imp should play: %s", msg)
	}
	st := lastState(t, a)
	uid := st.Self.Board[0].InstanceID
	found := false
	for _, e := range st.Events {
		if e.Kind == "summon" && e.Target == uid {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a summon event for %s, got %+v", uid, st.Events)
	}
}

// --- Phase 4 extended: edge triggers (on_spell_cast / on_summon / on_turn_start /
// on_minion_death). Cards react to ongoing game events while in play. ---

// TestSpellCastSelfBuffTrigger: an on_spell_cast self-buff minion gains Attack
// when its controller casts a spell.
func TestSpellCastSelfBuffTrigger(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "wyrm", "arcane_wyrmling", 1, 3, true)
	m.state[0].mana, m.state[0].maxMana = 3, 3
	cb := handIndex(lastState(t, a), "cinder_bolt")
	if ok, msg := m.PlayCard(a, cb, oppHeroTarget); !ok {
		t.Fatalf("cast should resolve: %s", msg)
	}
	if atk := findMinion(m.state[0].board, "wyrm").atk(); atk != 2 {
		t.Fatalf("on-cast buff should raise Attack 1->2, got %d", atk)
	}
}

// TestSilencedTriggerDoesNotFire: Silence removes an edge trigger like any other.
func TestSilencedTriggerDoesNotFire(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "wyrm", "arcane_wyrmling", 1, 3, true)
	m.silence(findMinion(m.state[0].board, "wyrm"))
	m.state[0].mana, m.state[0].maxMana = 3, 3
	cb := handIndex(lastState(t, a), "cinder_bolt")
	m.PlayCard(a, cb, oppHeroTarget)
	if atk := findMinion(m.state[0].board, "wyrm").atk(); atk != 1 {
		t.Fatalf("a silenced minion must not react to the cast, atk=%d", atk)
	}
}

// TestSpellCastDrawTrigger: an on_spell_cast draw minion draws when its controller
// casts a spell.
func TestSpellCastDrawTrigger(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "crier", "bazaar_crier", 4, 4, true)
	m.state[0].mana, m.state[0].maxMana = 3, 3
	m.state[0].deck = testDeck([]string{"pebble_imp", "granite_watcher"})
	cb := handIndex(lastState(t, a), "cinder_bolt")
	if ok, msg := m.PlayCard(a, cb, oppHeroTarget); !ok {
		t.Fatalf("cast should resolve: %s", msg)
	}
	if n := len(m.state[0].deck); n != 1 {
		t.Fatalf("on-cast draw should remove 1 from the deck, deck=%d", n)
	}
}

// TestSpellCastSummonTrigger: an on_spell_cast summon minion summons a token when
// its controller casts a spell.
func TestSpellCastSummonTrigger(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "tutor", "adept_tutor", 3, 5, true)
	m.state[0].mana, m.state[0].maxMana = 3, 3
	cb := handIndex(lastState(t, a), "cinder_bolt")
	if ok, msg := m.PlayCard(a, cb, oppHeroTarget); !ok {
		t.Fatalf("cast should resolve: %s", msg)
	}
	if n := len(m.state[0].board); n != 2 {
		t.Fatalf("on-cast summon should add a token (board 1->2), got %d", n)
	}
}

// TestSummonTriggerPingsEnemy: an after-you-summon minion damages a random enemy
// each time another friendly minion enters play.
func TestSummonTriggerPingsEnemy(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "tosser", "dagger_tosser", 2, 2, true)
	m.state[0].mana, m.state[0].maxMana = 1, 1
	imp := handIndex(lastState(t, a), "pebble_imp")
	if ok, msg := m.PlayCard(a, imp, ""); !ok {
		t.Fatalf("summon should resolve: %s", msg)
	}
	// No enemy minions, so the only enemy target is the hero.
	if hp := lastState(t, a).Opp.HeroHP; hp != 29 {
		t.Fatalf("summoning a minion should ping the enemy hero (30->29), got %d", hp)
	}
}

// TestTurnStartTrigger: an at-the-start-of-your-turn minion fires when its
// controller's turn begins (not the opponent's).
func TestTurnStartTrigger(t *testing.T) {
	m, a, b := newMatch()
	place(m, 0, "siege", "siege_engine", 1, 4, true)
	m.EndTurn(a) // b's turn — siege (player 0's) must NOT fire
	if hp := m.state[1].heroHP; hp != 30 {
		t.Fatalf("siege must not fire on the opponent's turn, enemy hero=%d", hp)
	}
	m.EndTurn(b) // back to a — siege fires, 2 to a random enemy (the hero)
	if hp := lastState(t, a).Opp.HeroHP; hp != 28 {
		t.Fatalf("turn-start trigger should deal 2 to the enemy hero (30->28), got %d", hp)
	}
}

// TestEmberlordEndTurnBurn: Emberlord Vrakgar (Ragnaros clone) can't attack, and
// at the end of ITS controller's turn deals 8 to a random enemy. Proves the
// can't-attack keyword + OnTurnEnd -> TargetRandomEnemy(hero) burn combo.
func TestEmberlordEndTurnBurn(t *testing.T) {
	m, a, b := newMatch()
	place(m, 0, "ember", "emberlord_vrakgar", 8, 8, true)

	// Can't attack despite being ready with 8 Attack.
	if ok, _ := m.Attack(a, "ember", "oppHero"); ok {
		t.Fatal("emberlord must not be able to attack")
	}

	m.EndTurn(b) // not yet a's turn-end; wrong-turn end_turn is a no-op for a's trigger
	if hp := m.state[1].heroHP; hp != 30 {
		t.Fatalf("emberlord must not fire on the opponent's turn, enemy hero=%d", hp)
	}
	m.EndTurn(a) // a's turn ends -> emberlord fires, 8 to the only enemy (the hero)
	if hp := m.state[1].heroHP; hp != 22 {
		t.Fatalf("end-of-turn burn should deal 8 to the enemy hero (30->22), got %d", hp)
	}
}

// TestTurnStartMassDestroy: a 0/7 whose turn-start trigger destroys ALL minions
// clears both boards (itself included) when its controller's turn begins. Proves
// the OnTurnStart -> AreaAllMinions -> EffectDestroy -> finish() path.
func TestTurnStartMassDestroy(t *testing.T) {
	m, a, b := newMatch()
	place(m, 0, "oracle", "ruin_oracle", 0, 7, true)
	place(m, 0, "ally", "pebble_imp", 1, 1, true)
	place(m, 1, "foe", "granite_warden", 1, 7, true)
	m.EndTurn(a) // b's turn — oracle (player 0's) must NOT fire
	if len(m.state[0].board)+len(m.state[1].board) != 3 {
		t.Fatalf("oracle must not fire on the opponent's turn")
	}
	m.EndTurn(b) // back to a — oracle fires at turn start, wipes both boards
	if n := len(m.state[0].board) + len(m.state[1].board); n != 0 {
		t.Fatalf("turn-start destroy-all should clear both boards, got %d minions left", n)
	}
}

// TestEnrageBonusTracksDamage: an Enrage minion gains its Attack bonus only while
// damaged, and loses it again when healed back to full.
func TestEnrageBonusTracksDamage(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "hg", "highland_guardian", 2, 3, true) // 2/3, Enrage +3
	mn := findMinion(m.state[0].board, "hg")
	if mn.atk() != 2 {
		t.Fatalf("undamaged enrage minion should be 2 atk, got %d", mn.atk())
	}
	mn.health = 2 // damaged
	if mn.atk() != 5 {
		t.Fatalf("damaged enrage minion should be 2+3=5 atk, got %d", mn.atk())
	}
	mn.health = 3 // healed to full
	if mn.atk() != 2 {
		t.Fatalf("healed-to-full enrage minion should drop back to 2 atk, got %d", mn.atk())
	}
}

// TestSilenceCancelsEnrage: Silence removes the Enrage ability, so a damaged
// minion no longer gets the bonus.
func TestSilenceCancelsEnrage(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "hg", "highland_guardian", 2, 3, true)
	mn := findMinion(m.state[0].board, "hg")
	mn.health = 2 // damaged -> enraged (5 atk)
	m.silence(mn)
	if mn.atk() != 2 {
		t.Fatalf("silence should cancel enrage (back to base 2 atk), got %d", mn.atk())
	}
}

// TestEnragedAtkUsedInCombat: combat reads the live (enraged) attack, so a damaged
// enrage minion strikes for its boosted value.
func TestEnragedAtkUsedInCombat(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "hg", "highland_guardian", 2, 3, true)
	findMinion(m.state[0].board, "hg").health = 2 // damaged -> 5 atk
	place(m, 1, "wall", "granite_warden", 1, 7, true)
	if ok, msg := m.Attack(a, "hg", "wall"); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	if hp := findMinion(m.state[1].board, "wall").health; hp != 2 {
		t.Fatalf("enraged 5-atk strike should leave the 7-hp wall at 2, got %d", hp)
	}
}

// TestAdjacencyOnsetBuffsNeighbours: an adjacency onset (Bannerguard)
// buffs +1/+1 and grants Taunt to the minions on each side of where it is played,
// and never to itself.
func TestAdjacencyOnsetBuffsNeighbours(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "left", "pebble_imp", 1, 1, true)
	place(m, 0, "right", "pebble_imp", 1, 1, true)
	m.state[0].hand = []cards.Card{getCard("bannerguard")}
	m.state[0].mana, m.state[0].maxMana = 4, 4
	if ok, msg := m.PlayCardAt(a, 0, "", 1); !ok { // drop between left and right
		t.Fatalf("play should resolve: %s", msg)
	}
	for _, uid := range []string{"left", "right"} {
		mn := findMinion(m.state[0].board, uid)
		if mn.atk() != 2 || mn.maxHP() != 2 || !mn.has(cards.KeywordTaunt) {
			t.Fatalf("%s neighbour should be 2/2 with Taunt, got %d/%d taunt=%v",
				uid, mn.atk(), mn.maxHP(), mn.has(cards.KeywordTaunt))
		}
	}
	bg := m.state[0].board[1] // inserted at pos 1, between the neighbours
	if bg.card.ID != "bannerguard" {
		t.Fatalf("bannerguard should sit between its neighbours, got %s", bg.card.ID)
	}
	if bg.maxHP() != 3 || bg.has(cards.KeywordTaunt) {
		t.Fatalf("an adjacency onset must not buff itself, got %d/%d taunt=%v",
			bg.atk(), bg.maxHP(), bg.has(cards.KeywordTaunt))
	}
}

// TestGrantedKeywordStrippedBySilence: a Taunt granted by a buff is removed when
// the recipient is silenced (the grant lives in an enchantment).
func TestGrantedKeywordStrippedBySilence(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "ally", "pebble_imp", 1, 1, true)
	m.state[0].hand = []cards.Card{getCard("wardstone_sentinel")} // BC: adjacent gain Taunt
	m.state[0].mana, m.state[0].maxMana = 2, 2
	if ok, msg := m.PlayCardAt(a, 0, "", 0); !ok { // drop left of ally
		t.Fatalf("play should resolve: %s", msg)
	}
	ally := findMinion(m.state[0].board, "ally")
	if !ally.has(cards.KeywordTaunt) {
		t.Fatal("ally should have gained Taunt from the adjacency onset")
	}
	m.silence(ally)
	if ally.has(cards.KeywordTaunt) {
		t.Fatal("silence should strip a granted Taunt")
	}
}

// TestSplashSpellHitsTargetAndNeighbours: a splash spell (Frostshear) damages and
// freezes the chosen minion plus the minions either side of it, and nothing else.
func TestSplashSpellHitsTargetAndNeighbours(t *testing.T) {
	m, a, _ := newMatch()
	for _, uid := range []string{"l", "mid", "r", "far"} {
		place(m, 1, uid, "granite_warden", 1, 7, true) // 1/7 wall
	}
	m.state[0].hand = []cards.Card{getCard("frostshear")}
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, msg := m.PlayCard(a, 0, "mid"); !ok { // target the middle wall
		t.Fatalf("cast should resolve: %s", msg)
	}
	for _, uid := range []string{"l", "mid", "r"} {
		mn := findMinion(m.state[1].board, uid)
		if mn.health != 6 || !mn.frozen {
			t.Fatalf("%s should take 1 (7->6) and be frozen, got hp=%d frozen=%v", uid, mn.health, mn.frozen)
		}
	}
	if far := findMinion(m.state[1].board, "far"); far.health != 7 || far.frozen {
		t.Fatalf("the non-adjacent minion must be untouched, got hp=%d frozen=%v", far.health, far.frozen)
	}
}

// TestElusiveCantBeSpellTargeted: a spell cannot target an Elusive minion.
func TestElusiveCantBeSpellTargeted(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "drake", "glimmerwing_drake", 3, 2, true) // Elusive
	m.state[0].hand = []cards.Card{getCard("frostshear")}
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, _ := m.PlayCard(a, 0, "drake"); ok {
		t.Fatal("a spell must not be allowed to target an Elusive minion")
	}
	// The card was not spent (rejected before mutation).
	if len(m.state[0].hand) != 1 {
		t.Fatal("a rejected cast must leave the hand untouched")
	}
}

// TestCantAttackMinion: a Can't Attack minion may never attack.
func TestCantAttackMinion(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "watch", "stoneveil_watcher", 4, 5, true) // Can't attack
	if ok, _ := m.Attack(a, "watch", "oppHero"); ok {
		t.Fatal("a Can't Attack minion must not be able to attack")
	}
}

// TestTempBuffExpiresAtEndOfTurn: a "+X Attack this turn" buff vanishes when the
// caster's turn ends.
func TestTempBuffExpiresAtEndOfTurn(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "ally", "pebble_imp", 1, 1, true)
	m.state[0].hand = []cards.Card{getCard("goad_imp")} // BC: give a minion +2 atk this turn
	m.state[0].mana, m.state[0].maxMana = 1, 1
	if ok, msg := m.PlayCard(a, 0, "ally"); !ok {
		t.Fatalf("play should resolve: %s", msg)
	}
	if atk := findMinion(m.state[0].board, "ally").atk(); atk != 3 {
		t.Fatalf("ally should be 1+2=3 atk this turn, got %d", atk)
	}
	m.EndTurn(a)
	if atk := findMinion(m.state[0].board, "ally").atk(); atk != 1 {
		t.Fatalf("temp buff should expire at end of turn (back to 1), got %d", atk)
	}
}

// TestBounceReturnsMinionToHand: a bounce onset removes a friendly minion from
// the board and puts its base card back in hand.
func TestBounceReturnsMinionToHand(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "victim", "pebble_imp", 3, 3, true) // buffed stats; should reset on bounce
	m.state[0].hand = []cards.Card{getCard("tavern_apprentice")}
	m.state[0].mana, m.state[0].maxMana = 2, 2
	if ok, msg := m.PlayCard(a, 0, "victim"); !ok {
		t.Fatalf("play should resolve: %s", msg)
	}
	if findMinion(m.state[0].board, "victim") != nil {
		t.Fatal("the bounced minion should be off the board")
	}
	hand := m.state[0].hand
	if len(hand) != 1 || hand[0].ID != "pebble_imp" {
		t.Fatalf("the bounced minion should return to hand as its base card, got %+v", hand)
	}
}

// TestOnHealTrigger: a "whenever a character is healed" minion grows when any heal
// lands (here, an Earthroot-style onset heals a damaged ally).
func TestOnHealTrigger(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "tender", "dawn_tender", 1, 2, true) // +2 atk on any heal
	place(m, 0, "hurt", "granite_warden", 1, 7, true)
	findMinion(m.state[0].board, "hurt").health = 4 // damaged
	m.state[0].hand = []cards.Card{getCard("earthroot_healer")}
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, msg := m.PlayCard(a, 0, "hurt"); !ok { // heal 3
		t.Fatalf("play should resolve: %s", msg)
	}
	if atk := findMinion(m.state[0].board, "tender").atk(); atk != 3 {
		t.Fatalf("on-heal minion should gain +2 atk (1->3), got %d", atk)
	}
}

// TestOnSecretPlayedTrigger: a "whenever a Secret is played" minion grows when a
// secret is played.
func TestOnSecretPlayedTrigger(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "keeper", "rune_warden", 1, 2, true)
	m.state[0].hand = []cards.Card{getCard("glacial_ward")} // a secret
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("secret should play: %s", msg)
	}
	k := findMinion(m.state[0].board, "keeper")
	if k.atk() != 2 || k.maxHP() != 3 {
		t.Fatalf("on-secret minion should gain +1/+1 (1/2->2/3), got %d/%d", k.atk(), k.maxHP())
	}
}

// TestOnPlayCardTrigger: a "whenever you play a card" minion grows when its
// controller plays another card (but not from its own play).
func TestOnPlayCardTrigger(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "seeker", "relic_seeker", 2, 2, true)
	m.state[0].hand = []cards.Card{getCard("mote")} // a vanilla 1/1
	m.state[0].mana, m.state[0].maxMana = 1, 1
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play should resolve: %s", msg)
	}
	s := findMinion(m.state[0].board, "seeker")
	if s.atk() != 3 || s.maxHP() != 3 {
		t.Fatalf("on-play-card minion should gain +1/+1 (2/2->3/3), got %d/%d", s.atk(), s.maxHP())
	}
}

// TestRandomFriendlyEndOfTurn: an end-of-turn "give another random friendly minion
// +1 Attack" buffs the only other friendly minion (deterministic with one).
func TestRandomFriendlyEndOfTurn(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "forge", "forge_hand", 1, 3, true)
	place(m, 0, "buddy", "pebble_imp", 1, 1, true)
	m.EndTurn(a)
	if atk := findMinion(m.state[0].board, "buddy").atk(); atk != 2 {
		t.Fatalf("the only other friendly minion should get +1 atk (1->2), got %d", atk)
	}
	if atk := findMinion(m.state[0].board, "forge").atk(); atk != 1 {
		t.Fatalf("the source must not buff itself, got %d", atk)
	}
}

// TestAnyDeathTrigger: an on-any-minion-death minion grows when ANY minion dies.
func TestAnyDeathTrigger(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "fiend", "carrion_fiend", 2, 3, true)
	place(m, 1, "victim", "pebble_imp", 1, 1, false)
	m.state[1].board[0].health = 0 // kill the victim (white-box)
	m.resetLog()
	m.finish()
	if atk := findMinion(m.state[0].board, "fiend").atk(); atk != 3 {
		t.Fatalf("a minion death should grow the fiend 2->3, got %d", atk)
	}
}

// TestFriendlyDeathDrawTrigger: an on-friendly-death draw minion draws when one of
// its controller's OTHER minions dies.
func TestFriendlyDeathDrawTrigger(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "overseer", "cabal_overseer", 4, 2, true)
	place(m, 0, "ally", "pebble_imp", 1, 1, true)
	m.state[0].deck = testDeck([]string{"pebble_imp", "granite_watcher"})
	m.state[0].board[1].health = 0 // kill the ally
	m.resetLog()
	m.finish()
	if n := len(m.state[0].deck); n != 1 {
		t.Fatalf("a friendly death should draw 1 (deck 2->1), got %d", n)
	}
}

// --- Batch 2 edge triggers: on_turn_end + condition, card-generate, all-minion AoE ---

// TestTurnEndConditionalBuffFires: an end-of-turn minion with a "control a Secret"
// condition buffs itself when its controller has a secret in play.
func TestTurnEndConditionalBuffFires(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "scholar", "warded_scholar", 3, 3, true)
	placeSecret(m, 0, "cinder_trap")
	m.EndTurn(a) // player 0's turn ends -> scholar fires (secret present)
	sc := findMinion(m.state[0].board, "scholar")
	if sc.atk() != 5 || sc.maxHP() != 5 {
		t.Fatalf("end-of-turn buff with a secret should give +2/+2 (3/3->5/5), got %d/%d", sc.atk(), sc.maxHP())
	}
}

// TestTurnEndConditionUnmet: the same minion does NOT buff when the condition
// (control a Secret) is false.
func TestTurnEndConditionUnmet(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "scholar", "warded_scholar", 3, 3, true)
	m.EndTurn(a) // no secret -> condition unmet -> no buff
	if atk := findMinion(m.state[0].board, "scholar").atk(); atk != 3 {
		t.Fatalf("end-of-turn trigger must not fire without a secret, atk=%d", atk)
	}
}

// TestGenerateAddsCardToHand: an on-cast generate minion adds the named card to
// its controller's hand when they cast a spell.
func TestGenerateAddsCardToHand(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "magus", "emberforge_magus", 5, 7, true)
	m.state[0].mana, m.state[0].maxMana = 3, 3
	m.state[0].hand = testDeck([]string{"cinder_bolt"}) // room for the generated card
	m.sendStateAll()
	cb := handIndex(lastState(t, a), "cinder_bolt")
	if ok, msg := m.PlayCard(a, cb, oppHeroTarget); !ok {
		t.Fatalf("cast should resolve: %s", msg)
	}
	found := false
	for _, c := range m.state[0].hand {
		if c.ID == "pyrebolt" {
			found = true
		}
	}
	if !found {
		t.Fatalf("on-cast generate should add a pyrebolt to hand")
	}
}

// TestAllMinionsAoEOnCast: an after-you-cast minion deals 1 to ALL minions —
// both boards, including itself.
func TestAllMinionsAoEOnCast(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "zealot", "ashflame_zealot", 3, 2, true)
	place(m, 0, "ally", "granite_watcher", 2, 3, true)
	place(m, 1, "victim", "pebble_imp", 1, 1, false)
	m.state[0].mana, m.state[0].maxMana = 3, 3
	cb := handIndex(lastState(t, a), "cinder_bolt")
	if ok, msg := m.PlayCard(a, cb, oppHeroTarget); !ok { // spell hits the hero, not minions
		t.Fatalf("cast should resolve: %s", msg)
	}
	if findMinion(m.state[1].board, "victim") != nil {
		t.Fatalf("the 1/1 enemy minion should die to the 1-damage AoE")
	}
	if hp := findMinion(m.state[0].board, "ally").health; hp != 2 {
		t.Fatalf("a friendly minion should take the AoE too (3->2), got %d", hp)
	}
	if hp := findMinion(m.state[0].board, "zealot").health; hp != 1 {
		t.Fatalf("the zealot should take its own AoE (2->1), got %d", hp)
	}
}

// --- Phase B: destroy effect + enemy-character / friendly-hero target rules ---

// TestDestroySpellIgnoresAegis: a Destroy effect removes the target minion
// outright, even through a Aegis (HS rule: destroy != damage).
func TestDestroySpellIgnoresAegis(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "shielded", "silver_page", 1, 1, false) // enemy 1/1 with Aegis
	m.state[0].mana, m.state[0].maxMana = 4, 4
	m.state[0].hand = testDeck([]string{"banish_rite"})
	m.sendStateAll()
	br := handIndex(lastState(t, a), "banish_rite")
	if ok, msg := m.PlayCard(a, br, "shielded"); !ok {
		t.Fatalf("banish rite should resolve: %s", msg)
	}
	if findMinion(m.state[1].board, "shielded") != nil {
		t.Fatalf("destroy must remove the minion despite its Aegis")
	}
}

// TestDestroyOnset: a onset that destroys an enemy minion fires on play.
func TestDestroyOnset(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "victim", "tundra_yak", 4, 5, false)
	m.state[0].mana, m.state[0].maxMana = 5, 5
	m.state[0].hand = testDeck([]string{"headsman"})
	m.sendStateAll()
	hs := handIndex(lastState(t, a), "headsman")
	if ok, msg := m.PlayCard(a, hs, "victim"); !ok {
		t.Fatalf("headsman should resolve: %s", msg)
	}
	if findMinion(m.state[1].board, "victim") != nil {
		t.Fatalf("onset should destroy the enemy minion")
	}
}

// TestTargetEnemyRule: an enemy-character effect may hit the enemy hero but not a
// friendly target.
func TestTargetEnemyRule(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "friendly", "tundra_yak", 4, 5, true)
	m.state[0].mana, m.state[0].maxMana = 5, 5
	m.state[0].hand = testDeck([]string{"bombard_captain"})
	m.sendStateAll()
	bc := handIndex(lastState(t, a), "bombard_captain")
	// Aiming at a friendly minion is illegal for an enemy-only effect.
	if ok, _ := m.PlayCard(a, bc, "friendly"); ok {
		t.Fatalf("enemy-target onset must not hit a friendly minion")
	}
	bc = handIndex(lastState(t, a), "bombard_captain")
	if ok, msg := m.PlayCard(a, bc, oppHeroTarget); !ok {
		t.Fatalf("enemy-target onset should hit the enemy hero: %s", msg)
	}
	if hp := m.state[1].heroHP; hp != 28 {
		t.Fatalf("enemy hero should take 2 (30->28), got %d", hp)
	}
}

// TestFriendlyHeroHeal: a friendly-hero heal restores the caster's own hero.
func TestFriendlyHeroHeal(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 20
	m.state[0].mana, m.state[0].maxMana = 2, 2
	m.state[0].hand = testDeck([]string{"tavern_medic_fx"})
	m.sendStateAll()
	tm := handIndex(lastState(t, a), "tavern_medic_fx")
	if ok, msg := m.PlayCard(a, tm, selfHeroTarget); !ok {
		t.Fatalf("tavern medic should resolve: %s", msg)
	}
	if hp := m.state[0].heroHP; hp != 24 {
		t.Fatalf("friendly-hero heal should restore 4 (20->24), got %d", hp)
	}
}

// TestTransformReplacesMinionInPlace: a transform effect swaps the target minion
// for the token in its board slot, resetting stats and dropping enchants/keywords,
// and the original finalGasp does NOT fire (it is replaced, not killed).
func TestTransformReplacesMinionInPlace(t *testing.T) {
	m, a, _ := newMatch()
	// A buffed, shielded finalGasp minion at slot 1 of the enemy board.
	place(m, 1, "ally0", "mote", 1, 1, false)
	place(m, 1, "victim", "reaper_golem", 6, 9, false) // reaper_golem has a summon finalGasp
	findMinion(m.state[1].board, "victim").aegis = true
	m.state[0].mana, m.state[0].maxMana = 4, 4
	m.state[0].hand = testDeck([]string{"hex_bolt"})
	m.sendStateAll()
	wc := handIndex(lastState(t, a), "hex_bolt")
	if ok, msg := m.PlayCard(a, wc, "victim"); !ok {
		t.Fatalf("hex bolt should resolve: %s", msg)
	}
	v := findMinion(m.state[1].board, "victim") // same uid, transformed
	if v == nil {
		t.Fatalf("transform keeps the same uid/slot, minion should still be there")
	}
	if v.atk() != 2 || v.maxHP() != 1 {
		t.Fatalf("transformed minion should be a 2/1, got %d/%d", v.atk(), v.maxHP())
	}
	if v.aegis {
		t.Fatalf("transform should strip Aegis")
	}
	if m.state[1].board[1] != v {
		t.Fatalf("transform must keep the board slot (index 1)")
	}
	// No extra token from a finalGasp: board is still the 2 we placed.
	if n := len(m.state[1].board); n != 2 {
		t.Fatalf("original finalGasp must not fire on transform; board=%d", n)
	}
}

// TestSpellSummonsTokens: a spell whose effect is a summon puts the tokens on the
// caster's board (the summon path was previously only reached by battlecries).
func TestSpellSummonsTokens(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 3, 3
	m.state[0].hand = testDeck([]string{"twin_summons"})
	m.sendStateAll()
	mc := handIndex(lastState(t, a), "twin_summons")
	if ok, msg := m.PlayCard(a, mc, ""); !ok {
		t.Fatalf("twin summons should resolve: %s", msg)
	}
	if n := len(m.state[0].board); n != 2 {
		t.Fatalf("summon spell should put 2 tokens on the board, got %d", n)
	}
}

// --- Phase 5: keywords wave 1 (taunt, charge, rush, aegis, freeze) ---

// TestChargeAttacksImmediately: a Charge minion can attack the enemy hero the
// turn it is played.
func TestChargeAttacksImmediately(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 3, 3
	raptor := handIndex(lastState(t, a), "swift_raptor")
	if ok, msg := m.PlayCard(a, raptor, ""); !ok {
		t.Fatalf("swift raptor should play: %s", msg)
	}
	rid := lastState(t, a).Self.Board[0].InstanceID
	if ok, msg := m.Attack(a, rid, oppHeroTarget); !ok {
		t.Fatalf("charge minion should attack face the turn it's played: %s", msg)
	}
	if hp := lastState(t, a).Opp.HeroHP; hp != 27 {
		t.Fatalf("3-attack charge should hit face for 3 (30->27), got %d", hp)
	}
}

// TestRushCannotHitHeroOnSummonTurn: a Rush minion may attack minions the turn
// it is played, but not the hero.
func TestRushCannotHitHeroOnSummonTurn(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 3, 3
	place(m, 1, "dummy", "granite_watcher", 2, 3, false) // enemy minion to hit
	stalker := handIndex(lastState(t, a), "lurking_stalker")
	if ok, msg := m.PlayCard(a, stalker, ""); !ok {
		t.Fatalf("lurking stalker should play: %s", msg)
	}
	sid := lastState(t, a).Self.Board[0].InstanceID
	if ok, msg := m.Attack(a, sid, oppHeroTarget); ok {
		t.Fatalf("rush minion must not hit the hero on its summon turn")
	} else if msg != "can't attack heroes this turn" {
		t.Fatalf("want hero-restriction message, got %q", msg)
	}
	if ok, msg := m.Attack(a, sid, "dummy"); !ok {
		t.Fatalf("rush minion should attack a minion on its summon turn: %s", msg)
	}
}

// TestTauntForcesAttack: while the defender controls a Taunt minion, the attacker
// may hit neither the hero nor a non-taunt minion.
func TestTauntForcesAttack(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "granite_watcher", 2, 3, true) // my ready attacker
	place(m, 1, "wall", "bastion_golem", 3, 5, false) // enemy taunt
	place(m, 1, "soft", "pebble_imp", 1, 1, false)    // enemy non-taunt

	if ok, msg := m.Attack(a, "atk", oppHeroTarget); ok {
		t.Fatalf("cannot hit hero through a taunt")
	} else if msg != "must attack a Taunt minion" {
		t.Fatalf("want taunt message for hero, got %q", msg)
	}
	if ok, msg := m.Attack(a, "atk", "soft"); ok {
		t.Fatalf("cannot hit a non-taunt minion through a taunt")
	} else if msg != "must attack a Taunt minion" {
		t.Fatalf("want taunt message for non-taunt, got %q", msg)
	}
	if ok, msg := m.Attack(a, "atk", "wall"); !ok {
		t.Fatalf("attacking the taunt minion should be allowed: %s", msg)
	}
}

// TestAegisAbsorbsOneHit: the first damage instance pops the shield and
// deals no damage; the minion keeps full health.
func TestAegisAbsorbsOneHit(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "granite_watcher", 2, 3, true)
	place(m, 1, "sentry", "gilded_sentry", 2, 3, false) // Aegis from keyword
	if ok, msg := m.Attack(a, "atk", "sentry"); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	sv := lastState(t, a).Opp.Board[0]
	if sv.Health != 3 || sv.Aegis {
		t.Fatalf("shield should absorb the hit (full health, shield gone), got %d hp shield=%v", sv.Health, sv.Aegis)
	}
}

// TestFreezePreventsAttackThenThaws: a frozen minion cannot attack on its
// controller's turn, and thaws at the end of that turn (it didn't attack).
func TestFreezePreventsAttackThenThaws(t *testing.T) {
	m, a, b := newMatch()
	place(m, 1, "foe", "granite_watcher", 2, 3, false)
	m.state[0].mana, m.state[0].maxMana = 2, 2
	snap := handIndex(lastState(t, a), "frost_snap")
	if ok, msg := m.PlayCard(a, snap, "foe"); !ok {
		t.Fatalf("frost snap should resolve: %s", msg)
	}
	fv := lastState(t, a).Opp.Board[0]
	if !fv.Frozen || fv.Health != 2 {
		t.Fatalf("frost snap should deal 1 and freeze, got %d hp frozen=%v", fv.Health, fv.Frozen)
	}
	m.EndTurn(a) // b's turn; foe is frozen
	if ok, msg := m.Attack(b, "foe", oppHeroTarget); ok {
		t.Fatalf("frozen minion must not attack")
	} else if msg != "minion is frozen" {
		t.Fatalf("want 'minion is frozen', got %q", msg)
	}
	m.EndTurn(b) // end of b's turn: foe didn't attack -> thaws
	if lastState(t, a).Opp.Board[0].Frozen {
		t.Fatalf("foe should thaw at the end of its controller's turn")
	}
}

// TestPermafrostFreezesAllEnemyMinions: the AoE freezes every enemy minion and
// (0 damage) kills none.
func TestPermafrostFreezesAllEnemyMinions(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "m1", "granite_watcher", 2, 3, false)
	place(m, 1, "m2", "pebble_imp", 1, 1, false)
	m.state[0].mana, m.state[0].maxMana = 3, 3
	pf := handIndex(lastState(t, a), "permafrost")
	if ok, msg := m.PlayCard(a, pf, ""); !ok {
		t.Fatalf("permafrost should resolve: %s", msg)
	}
	board := lastState(t, a).Opp.Board
	if len(board) != 2 {
		t.Fatalf("0-damage freeze should kill nothing, board=%d", len(board))
	}
	for _, mv := range board {
		if !mv.Frozen {
			t.Fatalf("every enemy minion should be frozen, got %+v", mv)
		}
	}
}

// mview finds a minion's view by instance id on the given side of a snapshot.
func mview(t *testing.T, st protocol.State, self bool, uid string) protocol.MinionView {
	t.Helper()
	board := st.Opp.Board
	if self {
		board = st.Self.Board
	}
	for _, mv := range board {
		if mv.InstanceID == uid {
			return mv
		}
	}
	t.Fatalf("minion %s not found in view", uid)
	return protocol.MinionView{}
}

// --- Phase 6: keywords wave 2 ---

// TestTwinstrikeAttacksTwice: a Twinstrike minion may attack twice in one turn, and a
// third attack is rejected. This is the whole point of Twinstrike — without the
// attacks-remaining model it could only swing once.
func TestTwinstrikeAttacksTwice(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "wf", "gale_harrier", 2, 3, true) // 2/3 Twinstrike, ready
	for i := 0; i < 2; i++ {
		if ok, msg := m.Attack(a, "wf", oppHeroTarget); !ok {
			t.Fatalf("twinstrike attack %d should resolve: %s", i+1, msg)
		}
	}
	if ok, _ := m.Attack(a, "wf", oppHeroTarget); ok {
		t.Fatalf("third attack must be rejected (only two per turn)")
	}
	if hp := lastState(t, a).Opp.HeroHP; hp != 26 {
		t.Fatalf("two 2-attacks should bring hero 30->26, got %d", hp)
	}
}

// TestLifestealHealsController: a Lifesteal minion dealing combat damage heals its
// controller's hero by the damage dealt. This is why Lifesteal exists — it must
// restore the attacker's side, not just deal damage.
func TestLifestealHealsController(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 20                             // damaged, so the heal is observable (cap is 30)
	place(m, 0, "ls", "bloodthorn_knight", 3, 4, true) // 3/4 Lifesteal
	place(m, 1, "foe", "granite_watcher", 2, 3, false) // 2/3
	if ok, msg := m.Attack(a, "ls", "foe"); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	if hp := lastState(t, a).Self.HeroHP; hp != 23 {
		t.Fatalf("lifesteal should heal controller 20->23 (3 dealt), got %d", hp)
	}
}

// TestPoisonousDestroys: any damage a Poisonous minion deals to a minion destroys
// it regardless of remaining health. The 1-attack Toxic Fang must kill a 3-health
// minion outright.
func TestPoisonousDestroys(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "tox", "toxic_fang", 1, 3, true)       // 1/3 Poisonous
	place(m, 1, "big", "granite_watcher", 2, 3, false) // 2/3, would survive 1 damage
	if ok, msg := m.Attack(a, "tox", "big"); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Opp.Board) != 0 {
		t.Fatalf("poisonous should destroy the struck minion, board=%d", len(st.Opp.Board))
	}
	if len(st.Self.Board) != 1 || st.Self.Board[0].Health != 1 {
		t.Fatalf("toxic fang should survive retaliation at 1 hp, got %+v", st.Self.Board)
	}
}

// TestAegisBlocksPoison: a Aegis absorbs the hit, so 0 damage is
// dealt and Poisonous does not trigger. The shielded minion survives.
func TestAegisBlocksPoison(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "tox", "toxic_fang", 1, 3, true)
	place(m, 1, "ds", "gilded_sentry", 2, 3, false) // Aegis
	if ok, msg := m.Attack(a, "tox", "ds"); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	sv := lastState(t, a).Opp.Board
	if len(sv) != 1 || sv[0].Health != 3 || sv[0].Aegis {
		t.Fatalf("shield should absorb poison hit (alive, full hp, shield gone), got %+v", sv)
	}
}

// TestStealthUntargetableByEnemy: an enemy may neither cast a targeted spell at a
// Stealthed minion nor attack it. AoE is unaffected (covered elsewhere).
func TestStealthUntargetableByEnemy(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "veil", "veil_stalker", 2, 1, false) // enemy Stealth
	place(m, 0, "atk", "granite_watcher", 2, 3, true)
	m.state[0].mana, m.state[0].maxMana = 2, 2
	bolt := handIndex(lastState(t, a), "cinder_bolt")
	if ok, msg := m.PlayCard(a, bolt, "veil"); ok {
		t.Fatalf("targeting an enemy Stealthed minion must be illegal")
	} else if msg != "illegal target" {
		t.Fatalf("want 'illegal target', got %q", msg)
	}
	if ok, msg := m.Attack(a, "atk", "veil"); ok {
		t.Fatalf("attacking an enemy Stealthed minion must be illegal")
	} else if msg != "can't attack a Stealthed minion" {
		t.Fatalf("want stealth attack message, got %q", msg)
	}
}

// TestStealthLostOnAttack: a Stealthed minion loses Stealth once it attacks, and
// becomes targetable by the enemy afterward.
func TestStealthLostOnAttack(t *testing.T) {
	m, a, b := newMatch()
	m.EndTurn(a) // b's turn
	place(m, 1, "veil", "veil_stalker", 2, 1, true)
	if ok, msg := m.Attack(b, "veil", oppHeroTarget); !ok {
		t.Fatalf("stealth minion should be able to attack: %s", msg)
	}
	if mview(t, lastState(t, b), true, "veil").Stealth {
		t.Fatalf("stealth should drop after attacking")
	}
	m.EndTurn(b) // a's turn
	m.state[0].mana, m.state[0].maxMana = 2, 2
	bolt := handIndex(lastState(t, a), "cinder_bolt")
	if ok, msg := m.PlayCard(a, bolt, "veil"); !ok {
		t.Fatalf("a revealed minion should be targetable: %s", msg)
	}
}

// TestStealthHitByAoE: an untargeted AoE (Quake) still damages a Stealthed minion
// — Stealth blocks targeting, not area effects.
func TestStealthHitByAoE(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "veil", "veil_stalker", 2, 1, false) // 2/1 Stealth
	m.state[0].mana, m.state[0].maxMana = 3, 3
	quake := handIndex(lastState(t, a), "quake")
	if ok, msg := m.PlayCard(a, quake, ""); !ok {
		t.Fatalf("quake should resolve: %s", msg)
	}
	if n := len(lastState(t, a).Opp.Board); n != 0 {
		t.Fatalf("quake should kill the 1-health stealthed minion, board=%d", n)
	}
}

// TestSpellDamageBoostsSpell: a friendly Spell Damage +1 minion adds 1 to a damage
// spell. Cinder Bolt (3) becomes 4.
func TestSpellDamageBoostsSpell(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "scribe", "ember_scribe", 1, 3, true) // Spell Damage +1
	place(m, 1, "tgt", "iron_bulwark", 4, 5, false)   // 4/5
	m.state[0].mana, m.state[0].maxMana = 2, 2
	bolt := handIndex(lastState(t, a), "cinder_bolt")
	if ok, msg := m.PlayCard(a, bolt, "tgt"); !ok {
		t.Fatalf("cinder bolt should resolve: %s", msg)
	}
	if hp := mview(t, lastState(t, a), false, "tgt").Health; hp != 1 {
		t.Fatalf("3+1 spell damage should leave 5->1, got %d", hp)
	}
}

// TestSpellDamagePerTargetAoE: Spell Damage adds to every instance of an AoE, not
// once. Quake (1) becomes 2 to all enemy minions.
func TestSpellDamagePerTargetAoE(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "scribe", "ember_scribe", 1, 3, true)
	place(m, 1, "g", "granite_watcher", 2, 3, false) // 2/3 -> 1 after 2 dmg
	place(m, 1, "p", "pebble_imp", 1, 1, false)      // 1/1 -> dead
	m.state[0].mana, m.state[0].maxMana = 3, 3
	quake := handIndex(lastState(t, a), "quake")
	if ok, msg := m.PlayCard(a, quake, ""); !ok {
		t.Fatalf("quake should resolve: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Opp.Board) != 1 || st.Opp.Board[0].InstanceID != "g" || st.Opp.Board[0].Health != 1 {
		t.Fatalf("boosted quake should deal 2 each (pebble dead, granite at 1), got %+v", st.Opp.Board)
	}
}

// TestAuraBuffsOthersNotSelf: an aura minion grants +1 Attack to the controller's
// other minions but not itself, and the buff disappears when the source leaves.
func TestAuraBuffsOthersNotSelf(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "pl", "pack_leader", 2, 3, true) // aura: other friendlies +1 atk
	place(m, 0, "p", "pebble_imp", 1, 1, true)
	pl, p := m.state[0].board[0], m.state[0].board[1]
	m.refreshAuras()
	if p.atk() != 2 {
		t.Fatalf("aura should give the pebble +1 (1->2), got %d", p.atk())
	}
	if pl.atk() != 2 {
		t.Fatalf("aura must not buff its own source, got %d", pl.atk())
	}
	// Source leaves -> buff gone.
	m.state[0].board = m.state[0].board[1:]
	m.refreshAuras()
	if p.atk() != 1 {
		t.Fatalf("aura should vanish with its source (back to 1), got %d", p.atk())
	}
}

// TestSilenceStripsKeywordsBuffsAndShield: Silence removes enchantments, keywords,
// and Aegis, and clamps current health to the now-lower max. This is the
// core of Silence — every ongoing effect must be erased at once.
func TestSilenceStripsKeywordsBuffsAndShield(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "wall", "bastion_golem", 5, 7, false) // 3/5 Taunt, buffed to 5/7
	m.state[0].mana, m.state[0].maxMana = 1, 1
	hush := handIndex(lastState(t, a), "hush")
	if ok, msg := m.PlayCard(a, hush, "wall"); !ok {
		t.Fatalf("hush should resolve: %s", msg)
	}
	mv := mview(t, lastState(t, a), false, "wall")
	if mv.Attack != 3 || mv.MaxHealth != 5 || mv.Health != 5 {
		t.Fatalf("silence should revert to base 3/5 (health clamped 7->5), got %d/%d (max %d)", mv.Attack, mv.Health, mv.MaxHealth)
	}
	if mv.Taunt || !mv.Silenced {
		t.Fatalf("silence should strip Taunt and mark silenced, got taunt=%v silenced=%v", mv.Taunt, mv.Silenced)
	}
	if mv.Text != "" {
		t.Fatalf("silenced minion should show no rules text (its abilities are gone), got %q", mv.Text)
	}
}

// TestSilenceSuppressesFinalGasp: a silenced minion's FinalGasp does not fire
// when it dies. Brood Mother summons no Hatchling after Silence.
func TestSilenceSuppressesFinalGasp(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "granite_watcher", 2, 3, true)
	place(m, 1, "bm", "brood_mother", 2, 1, false) // finalGasp: summon a Hatchling
	m.state[0].mana, m.state[0].maxMana = 1, 1
	hush := handIndex(lastState(t, a), "hush")
	if ok, msg := m.PlayCard(a, hush, "bm"); !ok {
		t.Fatalf("hush should resolve: %s", msg)
	}
	if ok, msg := m.Attack(a, "atk", "bm"); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	if n := len(lastState(t, a).Opp.Board); n != 0 {
		t.Fatalf("silenced finalGasp must not summon, enemy board=%d", n)
	}
}

// TestDrainTouchLifestealSpell: a Lifesteal spell heals the caster by the damage
// it deals. Drain Touch deals 2 to the enemy hero and restores 2 to the caster.
func TestDrainTouchLifestealSpell(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 20
	m.state[0].mana, m.state[0].maxMana = 2, 2
	drain := handIndex(lastState(t, a), "drain_touch")
	if ok, msg := m.PlayCard(a, drain, oppHeroTarget); !ok {
		t.Fatalf("drain touch should resolve: %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.HeroHP != 28 {
		t.Fatalf("drain touch should deal 2 to enemy hero (30->28), got %d", st.Opp.HeroHP)
	}
	if st.Self.HeroHP != 22 {
		t.Fatalf("lifesteal should heal caster 20->22, got %d", st.Self.HeroHP)
	}
}

// --- Phase 7: secrets + seek ---

// lastSeek returns the most recent Seek prompt sent to a player.
func lastSeek(t *testing.T, f *fakeSender) protocol.Seek {
	t.Helper()
	for i := len(f.msgs) - 1; i >= 0; i-- {
		var env protocol.Envelope
		json.Unmarshal(f.msgs[i], &env)
		if env.Type == protocol.TypeSeek {
			var d protocol.Seek
			json.Unmarshal(f.msgs[i], &d)
			return d
		}
	}
	t.Fatal("no seek prompt sent")
	return protocol.Seek{}
}

func hasSeek(f *fakeSender) bool {
	for _, b := range f.msgs {
		var env protocol.Envelope
		json.Unmarshal(b, &env)
		if env.Type == protocol.TypeSeek {
			return true
		}
	}
	return false
}

// TestSecretHiddenFromOpponent: the owner sees the secret card; the opponent sees
// only that a secret is active (a count), never which one. This is the whole
// point of a secret — hidden information.
func TestSecretHiddenFromOpponent(t *testing.T) {
	m, a, b := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	snare := handIndex(lastState(t, a), "snare")
	if ok, msg := m.PlayCard(a, snare, ""); !ok {
		t.Fatalf("playing a secret should resolve: %s", msg)
	}
	own := lastState(t, a).Self
	if own.SecretCount != 1 || len(own.Secrets) != 1 || own.Secrets[0].Name != "Snare" {
		t.Fatalf("owner should see the full secret, got count=%d secrets=%+v", own.SecretCount, own.Secrets)
	}
	foe := lastState(t, b).Opp // a's side, as the opponent sees it
	if foe.SecretCount != 1 {
		t.Fatalf("opponent should see secret count 1, got %d", foe.SecretCount)
	}
	if len(foe.Secrets) != 0 {
		t.Fatalf("opponent must NOT see secret identities, got %+v", foe.Secrets)
	}
}

// TestSecretNoDuplicate: a player cannot have two copies of the same secret
// active at once (standard rule).
func TestSecretNoDuplicate(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 5, 5
	placeSecret(m, 0, "snare")
	snare := handIndex(lastState(t, a), "snare")
	if ok, msg := m.PlayCard(a, snare, ""); ok {
		t.Fatalf("playing a duplicate secret should be rejected")
	} else if msg != "secret already active" {
		t.Fatalf("want 'secret already active', got %q", msg)
	}
}

// TestSnareDestroysAttacker: Snare fires when an enemy minion attacks the owner's
// hero — the attacker is destroyed and deals no damage, and the secret is spent.
func TestSnareDestroysAttacker(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "granite_watcher", 2, 3, true) // a's attacker
	placeSecret(m, 1, "snare")                        // b's secret
	if ok, msg := m.Attack(a, "atk", oppHeroTarget); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Self.Board) != 0 {
		t.Fatalf("snare should destroy the attacker, a board=%d", len(st.Self.Board))
	}
	if st.Opp.HeroHP != 30 {
		t.Fatalf("snare should negate hero damage, enemy hero=%d", st.Opp.HeroHP)
	}
	if st.Opp.SecretCount != 0 {
		t.Fatalf("snare should be spent after triggering, count=%d", st.Opp.SecretCount)
	}
}

// TestMimicCopiesPlayedMinion: Mimic fires when the enemy plays a minion,
// summoning a copy on the secret owner's board.
func TestMimicCopiesPlayedMinion(t *testing.T) {
	m, a, _ := newMatch()
	placeSecret(m, 1, "mimic") // b's secret
	// a plays Pebble Imp (1 mana) on its turn.
	if ok, msg := m.PlayCard(a, idxPebbleImp, ""); !ok {
		t.Fatalf("playing a minion should resolve: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Self.Board) != 1 {
		t.Fatalf("attacker should have its minion, a board=%d", len(st.Self.Board))
	}
	if len(st.Opp.Board) != 1 || st.Opp.Board[0].CardID != "pebble_imp" {
		t.Fatalf("mimic should summon a copy for its owner, got %+v", st.Opp.Board)
	}
	if st.Opp.SecretCount != 0 {
		t.Fatalf("mimic should be spent, count=%d", st.Opp.SecretCount)
	}
}

// TestNullifyCountersSpell: Nullify fires when the enemy casts a spell — the
// effect is cancelled, but the card and mana are still spent.
func TestNullifyCountersSpell(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	placeSecret(m, 1, "nullify") // b's secret
	bolt := handIndex(lastState(t, a), "cinder_bolt")
	if ok, msg := m.PlayCard(a, bolt, oppHeroTarget); !ok {
		t.Fatalf("casting should resolve (and be countered): %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.HeroHP != 30 {
		t.Fatalf("nullify should counter the bolt, enemy hero=%d", st.Opp.HeroHP)
	}
	if st.Self.Mana != 0 {
		t.Fatalf("countered spell still costs mana, mana=%d", st.Self.Mana)
	}
	if handIndex(st, "cinder_bolt") != -1 {
		t.Fatalf("countered spell is still discarded from hand")
	}
	if st.Opp.SecretCount != 0 {
		t.Fatalf("nullify should be spent, count=%d", st.Opp.SecretCount)
	}
}

// TestNullifyCountersSecret: a secret is itself a spell, so casting one into an
// enemy Nullify is countered — the secret never enters play, but the card and
// mana are still spent.
func TestNullifyCountersSecret(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 3, 3
	placeSecret(m, 1, "nullify") // b's counter-spell secret
	mimic := handIndex(lastState(t, a), "mimic")
	if ok, msg := m.PlayCard(a, mimic, ""); !ok {
		t.Fatalf("casting a secret should resolve (and be countered): %s", msg)
	}
	st := lastState(t, a)
	if st.Self.SecretCount != 0 {
		t.Fatalf("countered secret must not enter play, owner count=%d", st.Self.SecretCount)
	}
	if st.Opp.SecretCount != 0 {
		t.Fatalf("nullify should be spent, count=%d", st.Opp.SecretCount)
	}
	if st.Self.Mana != 0 {
		t.Fatalf("countered secret still costs mana, mana=%d", st.Self.Mana)
	}
	if handIndex(st, "mimic") != -1 {
		t.Fatalf("countered secret is still discarded from hand")
	}
}

// TestSeekPausesAndResumes: a Seek onset pauses the action — it sends
// a prompt (only to the chooser) and blocks the player's other actions until a
// Choose adds the picked card to hand and resumes.
func TestSeekPausesAndResumes(t *testing.T) {
	m, a, b := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	// newMatch deals an over-cap white-box hand; trim to a normal hand so the
	// seeked card has room to land (an over-full hand would burn it).
	m.state[0].hand = testDeck([]string{"arcane_insight"})
	m.sendStateAll()
	ai := handIndex(lastState(t, a), "arcane_insight") // Onset: Seek a minion
	if ok, msg := m.PlayCard(a, ai, ""); !ok {
		t.Fatalf("playing the seek minion should resolve: %s", msg)
	}
	// Prompt sent to the chooser, not the opponent.
	d := lastSeek(t, a)
	if len(d.Options) != 3 {
		t.Fatalf("seek should offer 3 options, got %d", len(d.Options))
	}
	for _, o := range d.Options {
		if o.CardType != "minion" {
			t.Fatalf("'seek a minion' must offer only minions, got %s", o.CardType)
		}
	}
	if hasSeek(b) {
		t.Fatalf("opponent must not receive the seek prompt")
	}
	// While pending, other actions are blocked.
	if ok, msg := m.EndTurn(a); ok {
		t.Fatalf("actions must be blocked during a seek")
	} else if msg != "finish seeking first" {
		t.Fatalf("want seek-block message, got %q", msg)
	}
	// Wrong player can't choose.
	if ok, _ := m.Choose(b, 0); ok {
		t.Fatalf("only the seeking player may choose")
	}
	want := d.Options[1].CardID
	if ok, msg := m.Choose(a, 1); !ok {
		t.Fatalf("choosing should resolve: %s", msg)
	}
	if handIndex(lastState(t, a), want) == -1 {
		t.Fatalf("chosen card %q should be added to hand", want)
	}
	// Action resumed: a can now end the turn.
	if ok, msg := m.EndTurn(a); !ok {
		t.Fatalf("turn should be playable after the seek resolves: %s", msg)
	}
}

// TestSeekIntoFullHandBurns: seeking with a full hand (10) burns the
// chosen card rather than overfilling.
func TestSeekIntoFullHandBurns(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	ai := handIndex(lastState(t, a), "arcane_insight")
	if ok, msg := m.PlayCard(a, ai, ""); !ok {
		t.Fatalf("playing the seek minion should resolve: %s", msg)
	}
	// Force a full hand before choosing.
	full := make([]cards.Card, maxHand)
	for i := range full {
		full[i] = getCard("pebble_imp")
	}
	m.state[0].hand = full
	if ok, msg := m.Choose(a, 0); !ok {
		t.Fatalf("choosing should resolve: %s", msg)
	}
	if n := lastState(t, a).Self.HandCount; n != maxHand {
		t.Fatalf("seek into a full hand should burn the card, hand=%d", n)
	}
}

// TestSummonDiscardedWhenBoardFull: an effect that summons into a full board
// (here a onset token onto a board already at the cap) silently discards the
// extra minion rather than exceeding maxBoard.
func TestSummonDiscardedWhenBoardFull(t *testing.T) {
	m, a, _ := newMatch()
	// Fill to one below the cap, then play Bog Warden (onset: summon a Bogling).
	for i := 0; i < maxBoard-1; i++ {
		place(m, 0, "f"+itoa(i), "pebble_imp", 1, 1, false)
	}
	m.state[0].mana, m.state[0].maxMana = 3, 3
	warden := handIndex(lastState(t, a), "bog_warden")
	if ok, msg := m.PlayCard(a, warden, ""); !ok {
		t.Fatalf("bog warden should be playable into the last slot: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Self.Board) != maxBoard {
		t.Fatalf("board should be exactly full, got %d", len(st.Self.Board))
	}
	for _, mv := range st.Self.Board {
		if mv.CardID == "broken_golem" {
			t.Fatalf("the onset token should have been discarded (board full)")
		}
	}
}

// --- Phase 8: hero power + weapons ---

// TestHeroPowerDamageOncePerTurn: the hero power deals 1 and costs 2 mana; a second
// use the same turn is rejected; it refreshes next turn.
func TestHeroPowerDamageOncePerTurn(t *testing.T) {
	m, a, b := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	if ok, msg := m.HeroPower(a, oppHeroTarget); !ok {
		t.Fatalf("hero power should resolve: %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.HeroHP != 29 {
		t.Fatalf("hero power should deal 1 (30->29), got %d", st.Opp.HeroHP)
	}
	if st.Self.Mana != 0 {
		t.Fatalf("hero power should cost 2 mana, got %d left", st.Self.Mana)
	}
	if ok, msg := m.HeroPower(a, oppHeroTarget); ok {
		t.Fatalf("hero power is once per turn")
	} else if msg != "hero power already used" {
		t.Fatalf("want 'hero power already used', got %q", msg)
	}
	m.EndTurn(a)
	m.EndTurn(b) // back to a, mana refreshed (ramped to 2)
	if ok, msg := m.HeroPower(a, oppHeroTarget); !ok {
		t.Fatalf("hero power should refresh next turn: %s", msg)
	}
}

// TestHeroPowerNotBoostedBySpellDamage: Spell Damage minions do NOT increase
// hero power damage (standard rule for the hero power).
func TestHeroPowerNotBoostedBySpellDamage(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "scribe", "ember_scribe", 1, 3, true) // Spell Damage +1
	m.state[0].mana, m.state[0].maxMana = 2, 2
	if ok, msg := m.HeroPower(a, oppHeroTarget); !ok {
		t.Fatalf("hero power should resolve: %s", msg)
	}
	if hp := lastState(t, a).Opp.HeroHP; hp != 29 {
		t.Fatalf("hero power must ignore spell damage (30->29), got %d", hp)
	}
}

// TestEquipWeaponGivesHeroAttack: equipping a weapon gives the hero attack and
// durability, and lets it attack this turn.
func TestEquipWeaponGivesHeroAttack(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	cleaver := handIndex(lastState(t, a), "ember_cleaver")
	if ok, msg := m.PlayCard(a, cleaver, ""); !ok {
		t.Fatalf("equipping a weapon should resolve: %s", msg)
	}
	st := lastState(t, a)
	if st.Self.Weapon == nil || st.Self.Weapon.Attack != 3 || st.Self.Weapon.Durability != 2 {
		t.Fatalf("weapon should be 3/2, got %+v", st.Self.Weapon)
	}
	if st.Self.HeroAttack != 3 || !st.Self.HeroCanAttack {
		t.Fatalf("hero should be able to attack for 3, got atk=%d can=%v", st.Self.HeroAttack, st.Self.HeroCanAttack)
	}
}

// TestHeroAttackFaceSpendsDurability: a hero attack on the enemy face deals the
// weapon's attack and spends one durability.
func TestHeroAttackFaceSpendsDurability(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	cleaver := handIndex(lastState(t, a), "ember_cleaver")
	m.PlayCard(a, cleaver, "")
	if ok, msg := m.Attack(a, selfHeroTarget, oppHeroTarget); !ok {
		t.Fatalf("hero attack should resolve: %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.HeroHP != 27 {
		t.Fatalf("hero attack should deal 3 (30->27), got %d", st.Opp.HeroHP)
	}
	if st.Self.Weapon == nil || st.Self.Weapon.Durability != 1 {
		t.Fatalf("weapon should be at 1 durability, got %+v", st.Self.Weapon)
	}
	if ok, _ := m.Attack(a, selfHeroTarget, oppHeroTarget); ok {
		t.Fatalf("hero may attack only once per turn")
	}
}

// TestHeroAttackMinionTakesRetaliation: attacking a minion with a weapon trades —
// the minion takes the weapon damage, the hero takes the minion's attack back.
func TestHeroAttackMinionTakesRetaliation(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	place(m, 1, "foe", "granite_watcher", 2, 3, false) // 2/3
	cleaver := handIndex(lastState(t, a), "ember_cleaver")
	m.PlayCard(a, cleaver, "")
	if ok, msg := m.Attack(a, selfHeroTarget, "foe"); !ok {
		t.Fatalf("hero attack should resolve: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Opp.Board) != 0 {
		t.Fatalf("3-attack weapon should kill the 2/3, board=%d", len(st.Opp.Board))
	}
	if st.Self.HeroHP != 28 {
		t.Fatalf("hero should take 2 retaliation (30->28), got %d", st.Self.HeroHP)
	}
}

// TestWeaponBreaksAtZeroDurability: a weapon is destroyed once its durability
// runs out across attacks.
func TestWeaponBreaksAtZeroDurability(t *testing.T) {
	m, a, b := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	cleaver := handIndex(lastState(t, a), "ember_cleaver") // 3/2
	m.PlayCard(a, cleaver, "")
	m.Attack(a, selfHeroTarget, oppHeroTarget) // dur 2 -> 1
	m.EndTurn(a)
	m.EndTurn(b)                               // back to a, hero attack refreshed
	m.Attack(a, selfHeroTarget, oppHeroTarget) // dur 1 -> 0, breaks
	st := lastState(t, a)
	if st.Self.Weapon != nil {
		t.Fatalf("weapon should break at 0 durability, got %+v", st.Self.Weapon)
	}
	if st.Self.HeroAttack != 0 {
		t.Fatalf("hero attack should be 0 with no weapon, got %d", st.Self.HeroAttack)
	}
}

// TestEquipReplacesOldWeapon: equipping a new weapon destroys the current one.
func TestEquipReplacesOldWeapon(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 5, 5
	m.PlayCard(a, handIndex(lastState(t, a), "ember_cleaver"), "") // 3/2
	m.PlayCard(a, handIndex(lastState(t, a), "quartz_spike"), "")  // 2/3
	w := lastState(t, a).Self.Weapon
	if w == nil || w.Name != "Quartz Spike" || w.Attack != 2 || w.Durability != 3 {
		t.Fatalf("new weapon should replace the old, got %+v", w)
	}
}

// TestFrozenHeroCannotAttack: a frozen hero cannot attack with its weapon.
func TestFrozenHeroCannotAttack(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	m.PlayCard(a, handIndex(lastState(t, a), "ember_cleaver"), "")
	m.state[0].frozen = true
	if ok, msg := m.Attack(a, selfHeroTarget, oppHeroTarget); ok {
		t.Fatalf("a frozen hero must not attack")
	} else if msg != "hero is frozen" {
		t.Fatalf("want 'hero is frozen', got %q", msg)
	}
}

// TestHeroAttackRespectsTaunt: a hero with a weapon must attack a Taunt minion.
func TestHeroAttackRespectsTaunt(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	m.PlayCard(a, handIndex(lastState(t, a), "ember_cleaver"), "")
	place(m, 1, "wall", "bastion_golem", 3, 5, false) // Taunt
	if ok, msg := m.Attack(a, selfHeroTarget, oppHeroTarget); ok {
		t.Fatalf("hero can't go face through a taunt")
	} else if msg != "must attack a Taunt minion" {
		t.Fatalf("want taunt message, got %q", msg)
	}
}

// TestHeroAttackDoesNotTriggerSnare: Snare is minion-specific, so a weapon-armed
// hero attack on the enemy hero does NOT trigger it.
func TestHeroAttackDoesNotTriggerSnare(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	placeSecret(m, 1, "snare") // b's Snare
	m.PlayCard(a, handIndex(lastState(t, a), "ember_cleaver"), "")
	if ok, msg := m.Attack(a, selfHeroTarget, oppHeroTarget); !ok {
		t.Fatalf("hero attack should resolve: %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.HeroHP != 27 {
		t.Fatalf("hero attack should land for 3 (30->27), got %d", st.Opp.HeroHP)
	}
	if st.Opp.SecretCount != 1 {
		t.Fatalf("snare must NOT trigger on a hero attack, count=%d", st.Opp.SecretCount)
	}
	if st.Self.Weapon == nil {
		t.Fatalf("weapon should survive (snare did not fire)")
	}
}

// --- Armor + Frost Ward (`glacial_ward`-style secret) ---

// TestArmorAbsorbsDamage: armor soaks damage before health, regardless of source
// (here a spell). Health is untouched until armor runs out.
func TestArmorAbsorbsDamage(t *testing.T) {
	m, a, _ := newMatch()
	m.state[1].armor = 5
	m.state[0].mana, m.state[0].maxMana = 2, 2
	bolt := handIndex(lastState(t, a), "cinder_bolt") // 3 damage
	if ok, msg := m.PlayCard(a, bolt, oppHeroTarget); !ok {
		t.Fatalf("cinder bolt should resolve: %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.Armor != 2 || st.Opp.HeroHP != 30 {
		t.Fatalf("armor should absorb 3 (5->2), health untouched (30), got armor=%d hp=%d", st.Opp.Armor, st.Opp.HeroHP)
	}
}

// TestFrostWardArmorOnMinionAttack: Frost Ward gains 8 armor when the hero is
// attacked by a minion, before damage, so the hit is absorbed.
func TestFrostWardArmorOnMinionAttack(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "atk", "granite_watcher", 2, 3, true) // 2/3 attacker
	placeSecret(m, 1, "frost_ward")
	if ok, msg := m.Attack(a, "atk", oppHeroTarget); !ok {
		t.Fatalf("attack should resolve: %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.Armor != 6 || st.Opp.HeroHP != 30 {
		t.Fatalf("frost ward should give 8 armor before a 2-damage hit (->6), hp 30, got armor=%d hp=%d", st.Opp.Armor, st.Opp.HeroHP)
	}
	if st.Opp.SecretCount != 0 {
		t.Fatalf("frost ward should be spent, count=%d", st.Opp.SecretCount)
	}
}

// TestFrostWardArmorOnWeaponAttack: Frost Ward also triggers on a weapon-armed
// hero attack (it is "when your hero is attacked", not minion-specific).
func TestFrostWardArmorOnWeaponAttack(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	m.PlayCard(a, handIndex(lastState(t, a), "ember_cleaver"), "") // 3-attack weapon
	placeSecret(m, 1, "frost_ward")
	if ok, msg := m.Attack(a, selfHeroTarget, oppHeroTarget); !ok {
		t.Fatalf("hero attack should resolve: %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.Armor != 5 || st.Opp.HeroHP != 30 {
		t.Fatalf("frost ward 8 armor before a 3-damage swing (->5), hp 30, got armor=%d hp=%d", st.Opp.Armor, st.Opp.HeroHP)
	}
	if st.Opp.SecretCount != 0 {
		t.Fatalf("frost ward should be spent, count=%d", st.Opp.SecretCount)
	}
}

// TestFrostWardNotTriggeredBySpell: spell damage to the hero is not an attack, so
// Frost Ward does not trigger and the damage lands on health.
func TestFrostWardNotTriggeredBySpell(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].mana, m.state[0].maxMana = 2, 2
	placeSecret(m, 1, "frost_ward")
	bolt := handIndex(lastState(t, a), "cinder_bolt") // 3 damage
	if ok, msg := m.PlayCard(a, bolt, oppHeroTarget); !ok {
		t.Fatalf("cinder bolt should resolve: %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.Armor != 0 || st.Opp.HeroHP != 27 {
		t.Fatalf("spell must not trigger frost ward (armor 0, hp 27), got armor=%d hp=%d", st.Opp.Armor, st.Opp.HeroHP)
	}
	if st.Opp.SecretCount != 1 {
		t.Fatalf("frost ward must remain active after a spell, count=%d", st.Opp.SecretCount)
	}
}

// --- Phase 9: decks, draw, mulligan, fatigue, Mana Surge ---

// deck30 builds a 30-card deck of a single card id (shuffle is then a no-op, so
// opening hands and draws are fully deterministic regardless of seed).
func deck30(id string) []cards.Card {
	ids := make([]string, 30)
	for i := range ids {
		ids[i] = id
	}
	return testDeck(ids)
}

// newDeckMatch starts a real Phase-9 match (mulligan phase) from the given decks.
func newDeckMatch(deckA, deckB []cards.Card) (*Match, *fakeSender, *fakeSender) {
	a, b := &fakeSender{id: "p1"}, &fakeSender{id: "p2"}
	m := New("m1", a, b, 1, deckA, deckB)
	m.Start()
	return m, a, b
}

// TestMulliganOpensAndBlocks: a match opens in the mulligan phase, normal actions
// are rejected until both players submit, and play begins only once both have.
func TestMulliganOpensAndBlocks(t *testing.T) {
	m, a, b := newDeckMatch(deck30("pebble_imp"), deck30("pebble_imp"))
	if m.mulligan == nil {
		t.Fatal("match should open in the mulligan phase")
	}
	// Opening hand sizes: first player 3, second player 4 (ManaSurge added later).
	if got := len(m.state[0].hand); got != openingFirst {
		t.Fatalf("first player opening hand = %d, want %d", got, openingFirst)
	}
	if got := len(m.state[1].hand); got != openingSecond {
		t.Fatalf("second player opening hand = %d, want %d", got, openingSecond)
	}
	// A normal action during mulligan is rejected.
	if ok, msg := m.PlayCard(a, 0, ""); ok || msg != "mulligan in progress" {
		t.Fatalf("play during mulligan should be rejected, got ok=%v msg=%q", ok, msg)
	}
	// First player submits; play has not begun (opponent still pending).
	if ok, msg := m.Mulligan(a, nil); !ok {
		t.Fatalf("first mulligan should succeed: %s", msg)
	}
	if m.mulligan == nil {
		t.Fatal("play should not begin until both players mulligan")
	}
	// Second player submits -> play begins.
	if ok, msg := m.Mulligan(b, nil); !ok {
		t.Fatalf("second mulligan should succeed: %s", msg)
	}
	if m.mulligan != nil {
		t.Fatal("play should begin once both have mulliganed")
	}
}

// TestMulliganReplacesWithoutChangingCounts: tossing k cards draws k replacements
// and shuffles the tossed cards back, so hand and deck sizes are preserved.
func TestMulliganReplacesWithoutChangingCounts(t *testing.T) {
	m, a, _ := newDeckMatch(deck30("pebble_imp"), deck30("pebble_imp"))
	handBefore := len(m.state[0].hand) // 3
	deckBefore := len(m.state[0].deck) // 27
	if ok, msg := m.Mulligan(a, []int{0, 1}); !ok {
		t.Fatalf("mulligan should succeed: %s", msg)
	}
	if got := len(m.state[0].hand); got != handBefore {
		t.Fatalf("mulligan must preserve hand size: %d -> %d", handBefore, got)
	}
	if got := len(m.state[0].deck); got != deckBefore {
		t.Fatalf("mulligan must preserve deck size: %d -> %d", deckBefore, got)
	}
}

// TestSecondPlayerGetsManaSurge: once play begins, the player going second holds The
// ManaSurge and the player going first does not.
func TestSecondPlayerGetsManaSurge(t *testing.T) {
	m, a, b := newDeckMatch(deck30("pebble_imp"), deck30("pebble_imp"))
	m.Mulligan(a, nil)
	m.Mulligan(b, nil)
	if hasManaSurge(m.state[0].hand) {
		t.Fatal("first player should not have Mana Surge")
	}
	if !hasManaSurge(m.state[1].hand) {
		t.Fatal("second player should have Mana Surge")
	}
}

func hasManaSurge(hand []cards.Card) bool {
	for _, c := range hand {
		if c.ID == "mana_surge" {
			return true
		}
	}
	return false
}

// TestManaSurgeGivesTemporaryMana: playing Mana Surge grants +1 mana this turn.
func TestManaSurgeGivesTemporaryMana(t *testing.T) {
	m, a, b := newDeckMatch(deck30("pebble_imp"), deck30("pebble_imp"))
	m.Mulligan(a, nil)
	m.Mulligan(b, nil)
	// It is the first player's turn (1 mana). Give them Mana Surge and play it.
	m.state[0].hand = []cards.Card{cards.ManaSurge()}
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("playing Mana Surge should succeed: %s", msg)
	}
	if got := m.state[0].mana; got != 2 {
		t.Fatalf("ManaSurge should give +1 mana (1 -> 2), got %d", got)
	}
	_ = b
}

// TestFatigueEscalates: drawing from an empty deck deals escalating self-damage.
func TestFatigueEscalates(t *testing.T) {
	m, _, _ := newDeckMatch(deck30("pebble_imp"), deck30("pebble_imp"))
	m.state[0].deck = nil // force an empty draw pile
	m.state[0].heroHP = 30
	m.drawCard(0)
	if m.state[0].fatigue != 1 || m.state[0].heroHP != 29 {
		t.Fatalf("first fatigue draw should deal 1 (hp 29), got fatigue=%d hp=%d", m.state[0].fatigue, m.state[0].heroHP)
	}
	m.drawCard(0)
	if m.state[0].fatigue != 2 || m.state[0].heroHP != 27 {
		t.Fatalf("second fatigue draw should deal 2 (hp 27), got fatigue=%d hp=%d", m.state[0].fatigue, m.state[0].heroHP)
	}
}

// --- Phase 10: reconnect ---

// TestReattachResyncsWithHistory verifies that a reconnecting player gets a
// resync snapshot carrying the recent event history (so the client can rebuild
// its log, including events that resolved while the player was away), not just
// bare state. Intent: a dropped player must not lose the record of what happened.
func TestReattachResyncsWithHistory(t *testing.T) {
	m, a, _ := newMatch()
	// a plays a 1/1: this emits a summon event into the rolling history.
	if ok, msg := m.PlayCard(a, idxPebbleImp, ""); !ok {
		t.Fatalf("play pebble_imp: %s", msg)
	}

	// A fresh connection re-adopts a's seat (same id) and reattaches.
	a2 := &fakeSender{id: "p1"}
	if !m.Reattach(a2) {
		t.Fatal("reattach into a live match should succeed")
	}

	st := lastState(t, a2)
	if !st.Resync {
		t.Fatal("reattach snapshot must be flagged Resync so the client replaces its log")
	}
	if len(st.Events) == 0 {
		t.Fatal("resync should carry the event history, got none")
	}
	var sawSummon bool
	for _, e := range st.Events {
		if e.Kind == "summon" {
			sawSummon = true
		}
	}
	if !sawSummon {
		t.Fatalf("history should include the summon that happened before reconnect: %+v", st.Events)
	}
	if len(st.Self.Board) != 1 {
		t.Fatalf("resync state should show the summoned minion on board, got %d", len(st.Self.Board))
	}
}

// TestReattachAfterOverRejected verifies a finished match is not reconnectable.
func TestReattachAfterOverRejected(t *testing.T) {
	m, a, _ := newMatch()
	m.Concede(a) // a forfeits; match is over
	a2 := &fakeSender{id: "p1"}
	if m.Reattach(a2) {
		t.Fatal("reattach into a finished match should be rejected")
	}
}

// TestPlayCardAtPosition: a played minion lands at the requested board slot
// (drag-to-position), the row shifts to make room, pos<0 appends, and an
// out-of-range pos clamps to the end. This encodes the HS placement rule the
// drag UI relies on — the server, not the client, owns final board order.
func TestPlayCardAtPosition(t *testing.T) {
	m, a, _ := newMatch()
	m.turn = 0
	m.state[0].mana, m.state[0].maxMana = 30, 30

	handIdx := func(id string) int {
		for i, c := range m.state[0].hand {
			if c.ID == id {
				return i
			}
		}
		t.Fatalf("card %s not in hand", id)
		return -1
	}
	play := func(id string, pos int) {
		if ok, msg := m.PlayCardAt(a, handIdx(id), "", pos); !ok {
			t.Fatalf("play %s at %d: %s", id, pos, msg)
		}
	}

	play("pebble_imp", -1)     // append      -> [pebble]
	play("thicket_stalker", 0) // front       -> [thicket, pebble]
	play("granite_watcher", 1) // middle      -> [thicket, granite, pebble]
	play("clay_acolyte", 99)   // clamp->end  -> [thicket, granite, pebble, clay]

	want := []string{"thicket_stalker", "granite_watcher", "pebble_imp", "clay_acolyte"}
	got := make([]string, len(m.state[0].board))
	for i, mn := range m.state[0].board {
		got[i] = mn.card.ID
	}
	if len(got) != len(want) {
		t.Fatalf("board size = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("board order = %v, want %v", got, want)
		}
	}
}

// --- Special / conditional battlecries (sub-wave 1) ---

// setHandSolo replaces player pi's hand with the given card ids and gives plenty
// of mana, so a onset can be played at hand index 0 deterministically.
func setHandSolo(m *Match, pi int, ids ...string) {
	m.state[pi].hand = testDeck(ids)
	m.state[pi].mana, m.state[pi].maxMana = 10, 10
}

// Wounded Duelist's onset deals 4 to ITSELF — it survives (4/7 -> 3 hp) and
// stays on the board. WHY: TargetSelf battlecries must anchor on the just-played
// minion, not fizzle for lack of a chosen target.
func TestSelfDamageOnset(t *testing.T) {
	m, a, _ := newMatch()
	setHandSolo(m, 0, "wounded_duelist")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("wounded duelist should play: %s", msg)
	}
	b := lastState(t, a).Self.Board
	if len(b) != 1 {
		t.Fatalf("duelist should be on board, got %d minions", len(b))
	}
	if b[0].Health != 3 {
		t.Fatalf("self-damage should leave 7-4=3 health, got %d", b[0].Health)
	}
}

// Powder Tosser fires three 1-damage missiles among ALL OTHER characters (never
// itself), totalling exactly 3. WHY: the spread must exclude the source and deal
// the full 3 split across the other characters.
func TestMissilesSplitExcludesSource(t *testing.T) {
	m, a, b := newMatch()
	// No other minions: every missile lands on one of the two heroes.
	setHandSolo(m, 0, "powder_tosser")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("powder tosser should play: %s", msg)
	}
	st := lastState(t, a)
	lost := (heroMaxHP - st.Self.HeroHP) + (heroMaxHP - st.Opp.HeroHP)
	if lost != 3 {
		t.Fatalf("3 missiles should deal 3 total to heroes, dealt %d", lost)
	}
	if st.Self.Board[0].Health != 2 {
		t.Fatalf("the tosser must not damage itself, health %d", st.Self.Board[0].Health)
	}
	_ = b
}

// Trampling Brute destroys a random enemy minion with <=2 Attack and leaves
// higher-attack minions alone. WHY: the stat filter must gate which minions are
// eligible for the random destroy.
func TestRandomConditionalDestroy(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "weak", "pebble_imp", 2, 2, false)     // 2 atk: eligible
	place(m, 1, "strong", "iron_bulwark", 5, 5, false) // 5 atk: spared
	setHandSolo(m, 0, "trampling_brute")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("trampling brute should play: %s", msg)
	}
	opp := lastState(t, a).Opp.Board
	if len(opp) != 1 || opp[0].InstanceID != "strong" {
		t.Fatalf("only the 2-attack minion should die, board=%v", opp)
	}
}

// Duskscale Drake gains +1 Health per card left in hand AFTER it is played.
// WHY: PerCardInHand must count the post-play hand (the drake itself excluded).
func TestHandCountBuff(t *testing.T) {
	m, a, _ := newMatch()
	setHandSolo(m, 0, "duskscale_drake", "mote", "mote", "mote") // 3 cards remain after playing the drake
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("duskscale drake should play: %s", msg)
	}
	b := lastState(t, a).Self.Board
	if b[0].Health != 4 || b[0].MaxHealth != 4 {
		t.Fatalf("drake should be 1+3=4 health, got %d/%d", b[0].Health, b[0].MaxHealth)
	}
}

// Covert Saboteur destroys a random enemy Secret. WHY: secret removal must hit
// the opponent's zone and lower their visible secret count.
func TestDestroyEnemySecret(t *testing.T) {
	m, a, _ := newMatch()
	placeSecret(m, 1, "snare")
	setHandSolo(m, 0, "covert_saboteur")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("covert saboteur should play: %s", msg)
	}
	if n := lastState(t, a).Opp.SecretCount; n != 0 {
		t.Fatalf("enemy secret should be destroyed, count=%d", n)
	}
}

// --- Special / conditional battlecries (sub-wave 2: target-condition + copy) ---

// Trophy Hunter destroys only a minion with 7+ Attack. Targeting a weaker minion
// is rejected; with no 7+ minion the onset fizzles and the body still plays.
// WHY: a target CONDITION must gate which minions are legal, and an unmet
// condition must not block playing the minion.
func TestConditionalDestroyMinAttack(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "beefy", "iron_bulwark", 7, 5, false) // 7 atk: legal target
	place(m, 1, "small", "pebble_imp", 3, 3, false)   // 3 atk: not legal
	setHandSolo(m, 0, "trophy_hunter")
	// Targeting the sub-7 minion is illegal.
	if ok, msg := m.PlayCard(a, 0, "small"); ok {
		t.Fatalf("trophy hunter must not target a 3-attack minion: %s", msg)
	}
	// Targeting the 7-attack minion destroys it.
	if ok, msg := m.PlayCard(a, 0, "beefy"); !ok {
		t.Fatalf("trophy hunter on a 7-attack minion should resolve: %s", msg)
	}
	if ids := minionIDs(lastState(t, a).Opp.Board); len(ids) != 1 || ids[0] != "small" {
		t.Fatalf("only the 7-attack minion should die, board=%v", ids)
	}
}

// Trophy Hunter with no 7+ minion on the board fizzles: it plays as a 4/2 and
// destroys nothing. WHY: a conditional onset with no legal target still summons.
func TestConditionalDestroyFizzles(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "small", "pebble_imp", 3, 3, false)
	setHandSolo(m, 0, "trophy_hunter")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("trophy hunter should still play with no legal target: %s", msg)
	}
	if len(lastState(t, a).Opp.Board) != 1 {
		t.Fatalf("fizzled onset must not destroy anything")
	}
	if len(lastState(t, a).Self.Board) != 1 {
		t.Fatalf("trophy hunter body should be on board")
	}
}

// Grave Knight destroys an ENEMY minion with Taunt only. A friendly taunt and an
// enemy non-taunt are both illegal. WHY: the condition (Taunt) and the side
// (enemy) must both be enforced.
func TestConditionalDestroyEnemyTaunt(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "myWall", "granite_warden", 1, 7, false) // friendly taunt: illegal (enemy-only)
	place(m, 1, "imp", "pebble_imp", 2, 2, false)        // enemy non-taunt: illegal (needs taunt)
	place(m, 1, "wall", "granite_warden", 1, 7, false)   // enemy taunt: the only legal target
	setHandSolo(m, 0, "grave_knight")
	if ok, _ := m.PlayCard(a, 0, "myWall"); ok {
		t.Fatal("grave knight must not target a friendly taunt")
	}
	if ok, _ := m.PlayCard(a, 0, "imp"); ok {
		t.Fatal("grave knight must not target an enemy non-taunt")
	}
	if ok, msg := m.PlayCard(a, 0, "wall"); !ok {
		t.Fatalf("grave knight should destroy an enemy taunt: %s", msg)
	}
	if ids := minionIDs(lastState(t, a).Opp.Board); len(ids) != 1 || ids[0] != "imp" {
		t.Fatalf("only the enemy taunt should die, board=%v", ids)
	}
}

// Visage Thief becomes a copy of the chosen minion: it takes that minion's card
// and buffed stats at fresh full health. WHY: a copy onset must replace the
// source's identity/stats, not merely buff it.
func TestCopyOnset(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "model", "clay_acolyte", 5, 6, false) // a buffed 5/6 to copy
	setHandSolo(m, 0, "visage_thief")
	if ok, msg := m.PlayCard(a, 0, "model"); !ok {
		t.Fatalf("visage thief should copy the chosen minion: %s", msg)
	}
	b := lastState(t, a).Self.Board
	if len(b) != 1 {
		t.Fatalf("visage thief should be the only friendly minion, got %d", len(b))
	}
	if b[0].CardID != "clay_acolyte" || b[0].Attack != 5 || b[0].Health != 6 {
		t.Fatalf("copy should be a 5/6 clay_acolyte, got %s %d/%d", b[0].CardID, b[0].Attack, b[0].Health)
	}
}

// minionIDs lists a board's instance ids in order, for compact assertions.
func minionIDs(b []protocol.MinionView) []string {
	out := make([]string, len(b))
	for i, m := range b {
		out[i] = m.InstanceID
	}
	return out
}

// --- Special / conditional battlecries (sub-wave 3: swap-stats + consume-shields) ---

// Addled Brewer swaps a minion's Attack and Health. A 5/2 becomes a 2/5. WHY: the
// swap must set both current and max health from the old attack and the attack
// from the old (current) health.
func TestSwapStats(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "v", "clay_acolyte", 5, 2, false) // a 5/2
	setHandSolo(m, 0, "addled_brewer")
	if ok, msg := m.PlayCard(a, 0, "v"); !ok {
		t.Fatalf("addled brewer should swap: %s", msg)
	}
	got := lastState(t, a).Opp.Board[0]
	if got.Attack != 2 || got.Health != 5 || got.MaxHealth != 5 {
		t.Fatalf("5/2 should swap to 2/5, got %d/%d (max %d)", got.Attack, got.Health, got.MaxHealth)
	}
}

// Swapping a 0-attack minion drops it to 0 Health, so it dies. WHY: the swap is a
// real stat change, and a 0-health result must resolve as a death.
func TestSwapStatsKillsZeroAttack(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "z", "shield_lackey", 0, 4, false) // a 0/4
	setHandSolo(m, 0, "addled_brewer")
	if ok, msg := m.PlayCard(a, 0, "z"); !ok {
		t.Fatalf("addled brewer should resolve on a 0/4: %s", msg)
	}
	if n := len(lastState(t, a).Opp.Board); n != 0 {
		t.Fatalf("a 0/4 swapped to 4/0 should die, board has %d", n)
	}
}

// Crimson Reaver strips every Aegis on both boards and grows +3/+3 per
// shield removed. WHY: it must consume friendly AND enemy shields and scale its
// buff by the count.
func TestConsumeShields(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "f", "silver_page", 1, 1, false) // friendly Aegis
	place(m, 1, "e", "silver_page", 1, 1, false) // enemy Aegis
	setHandSolo(m, 0, "crimson_reaver")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("crimson reaver should play: %s", msg)
	}
	st := lastState(t, a)
	reaver := st.Self.Board[len(st.Self.Board)-1]
	if reaver.CardID != "crimson_reaver" || reaver.Attack != 9 || reaver.Health != 9 {
		t.Fatalf("reaver should be 3/3 + 2 shields*3 = 9/9, got %d/%d", reaver.Attack, reaver.Health)
	}
	for _, mn := range append(append([]protocol.MinionView{}, st.Self.Board...), st.Opp.Board...) {
		if mn.Aegis {
			t.Fatalf("all Aegiss should be stripped, %s still has one", mn.InstanceID)
		}
	}
}

// --- Weapon-manipulation battlecries ---

// equipWeapon arms player pi with a weapon for white-box tests.
func equipWeapon(m *Match, pi int, cardID string, attack, durability int) {
	m.state[pi].weapon = &weaponInst{card: getCard(cardID), attack: attack, durability: durability}
}

// Tidereaver gains Attack equal to the caster's weapon Attack. WHY: the buff must
// read the live weapon Attack and apply to the just-played body.
func TestGainWeaponAttack(t *testing.T) {
	m, a, _ := newMatch()
	equipWeapon(m, 0, "ember_cleaver", 3, 2)
	setHandSolo(m, 0, "tidereaver")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("tidereaver should play: %s", msg)
	}
	if got := lastState(t, a).Self.Board[0].Attack; got != 5 {
		t.Fatalf("2/3 + weapon 3 atk = 5 attack, got %d", got)
	}
}

// Brine Cutter strips 1 Durability from the opponent's weapon (breaking it at 0).
// WHY: weapon chip must target the OPPONENT and destroy the weapon at 0 durability.
func TestChipOpponentWeapon(t *testing.T) {
	m, a, _ := newMatch()
	equipWeapon(m, 1, "quartz_spike", 2, 3)
	setHandSolo(m, 0, "brine_cutter")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("brine cutter should play: %s", msg)
	}
	if w := lastState(t, a).Opp.Weapon; w == nil || w.Durability != 2 {
		t.Fatalf("opponent weapon should drop to 2 durability, got %v", w)
	}
	// One more chip from durability 1 destroys it.
	equipWeapon(m, 1, "ember_cleaver", 3, 1)
	setHandSolo(m, 0, "brine_cutter")
	if ok, _ := m.PlayCard(a, 0, ""); !ok {
		t.Fatal("brine cutter should play again")
	}
	if w := lastState(t, a).Opp.Weapon; w != nil {
		t.Fatalf("a 1-durability weapon should break, got %v", w)
	}
}

// Captain Brackwater gives the caster's weapon +1/+1. WHY: the buff must raise the
// weapon's live attack and durability.
func TestBuffOwnWeapon(t *testing.T) {
	m, a, _ := newMatch()
	equipWeapon(m, 0, "ember_cleaver", 3, 2)
	setHandSolo(m, 0, "captain_brackwater")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("captain brackwater should play: %s", msg)
	}
	w := lastState(t, a).Self.Weapon
	if w == nil || w.Attack != 4 || w.Durability != 3 {
		t.Fatalf("weapon should be 4 atk / 3 dur, got %v", w)
	}
}

// Relic Breaker destroys the opponent's weapon and draws cards equal to its
// Durability. WHY: both the destroy and the durability-scaled draw must happen.
func TestDestroyWeaponDraw(t *testing.T) {
	m, a, _ := newMatch()
	equipWeapon(m, 1, "quartz_spike", 2, 3)
	before := lastState(t, a).Self.DeckCount
	setHandSolo(m, 0, "relic_breaker")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("relic breaker should play: %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.Weapon != nil {
		t.Fatal("opponent weapon should be destroyed")
	}
	if drawn := before - st.Self.DeckCount; drawn != 3 {
		t.Fatalf("should draw 3 (weapon durability), drew %d", drawn)
	}
}

// --- Tribe auras + tribe-synergy minions ---

// boardMinion returns the minion with the given uid on owner's board (white-box).
func boardMinion(m *Match, owner int, uid string) *minion {
	return findMinion(m.state[owner].board, uid)
}

// A tribe-scoped Attack aura buffs only the controller's OTHER minions of that
// tribe, not itself and not off-tribe minions. WHY: the aura must filter by tribe.
func TestTribeAttackAura(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "fish", "brackish_caller", 1, 2, true) // a Gilkin
	place(m, 0, "wisp", "mote", 1, 1, true)            // not a Gilkin
	place(m, 0, "lord", "reef_warchief", 3, 3, true)   // +2 Atk to other Gilkins
	m.refreshAuras()
	if got := boardMinion(m, 0, "fish").atk(); got != 3 {
		t.Fatalf("gilkin should get +2 atk = 3, got %d", got)
	}
	if got := boardMinion(m, 0, "wisp").atk(); got != 1 {
		t.Fatalf("non-gilkin must not be buffed, got %d", got)
	}
	if got := boardMinion(m, 0, "lord").atk(); got != 3 {
		t.Fatalf("aura source must not buff itself, got %d", got)
	}
}

// An adjacency aura buffs only the immediate neighbours. WHY: positional auras
// must cover idx-1/idx+1 and nothing further.
func TestAdjacencyAura(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "L", "mote", 1, 1, true)
	place(m, 0, "C", "fang_alpha", 2, 2, true) // +1 Atk to adjacent
	place(m, 0, "R", "mote", 1, 1, true)
	place(m, 0, "far", "mote", 1, 1, true)
	m.refreshAuras()
	if boardMinion(m, 0, "L").atk() != 2 || boardMinion(m, 0, "R").atk() != 2 {
		t.Fatal("both neighbours should be +1 atk")
	}
	if boardMinion(m, 0, "far").atk() != 1 {
		t.Fatal("non-adjacent minion must not be buffed")
	}
	if boardMinion(m, 0, "C").atk() != 2 {
		t.Fatal("aura source must not buff itself")
	}
}

// A health aura raises a covered minion's max AND current health on entry; on the
// source leaving, a DAMAGED minion only clamps to the new max (never dies), while
// an undamaged one drops to the lower max. WHY: this is the tricky delta rule.
func TestHealthAuraDelta(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "pir", "brine_cutter", 1, 2, true)     // a Pirate, base 1/2
	place(m, 0, "cap", "tidehook_captain", 3, 3, true) // +1/+1 to other Pirates
	m.refreshAuras()
	pir := boardMinion(m, 0, "pir")
	if pir.atk() != 2 || pir.health != 3 || pir.maxHP() != 3 {
		t.Fatalf("aura'd pirate should be 2/3 (health 3), got %d/%d max %d", pir.atk(), pir.health, pir.maxHP())
	}
	// Damage it to 1, then remove the aura source: must clamp to new max, not die.
	pir.health = 1
	m.state[0].board = m.state[0].board[:1] // drop the captain
	m.refreshAuras()
	if pir.health != 1 || pir.maxHP() != 2 {
		t.Fatalf("damaged pirate should survive at 1 health / max 2, got %d/max %d", pir.health, pir.maxHP())
	}
	if pir.atk() != 1 {
		t.Fatalf("attack aura should be gone, got %d", pir.atk())
	}
}

// Removing a health aura from an UNDAMAGED minion clamps its current health down to
// the lower max. WHY: the clamp branch must apply when current exceeds new max.
func TestHealthAuraClampUndamaged(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "pir", "brine_cutter", 1, 2, true)
	place(m, 0, "cap", "tidehook_captain", 3, 3, true)
	m.refreshAuras()
	pir := boardMinion(m, 0, "pir")
	if pir.health != 3 {
		t.Fatalf("expected full 3 health under aura, got %d", pir.health)
	}
	m.state[0].board = m.state[0].board[:1]
	m.refreshAuras()
	if pir.health != 2 {
		t.Fatalf("undamaged minion should clamp from 3 to 2, got %d", pir.health)
	}
}

// Tidescry Oracle's onset gives the controller's OTHER Gilkins +2 Health, not
// itself and not off-tribe minions. WHY: the tribe-scoped buff area must filter.
func TestTribeOnsetBuff(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "fish", "brackish_caller", 1, 2, true) // Gilkin, base health 2
	place(m, 0, "wisp", "mote", 1, 1, true)            // not a Gilkin
	setHandSolo(m, 0, "tidescry_oracle")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("tidescry oracle should play: %s", msg)
	}
	if got := boardMinion(m, 0, "fish").health; got != 4 {
		t.Fatalf("other gilkin should gain +2 health = 4, got %d", got)
	}
	if got := boardMinion(m, 0, "wisp").health; got != 1 {
		t.Fatalf("non-gilkin must be untouched, got %d", got)
	}
}

// Brackish Caller gains +1 Attack only when a GILKIN is summoned, not other tribes.
// WHY: the summon trigger must gate on the subject minion's tribe.
func TestTribeSummonTrigger(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "caller", "brackish_caller", 1, 2, true)
	caller := boardMinion(m, 0, "caller")
	m.summonMinion(0, getCard("reef_warchief")) // a Gilkin
	if caller.atk() != 2 {
		t.Fatalf("summoning a Gilkin should give +1 atk = 2, got %d", caller.atk())
	}
	m.summonMinion(0, getCard("mote")) // not a Gilkin
	if caller.atk() != 2 {
		t.Fatalf("summoning a non-Gilkin must not buff, got %d", caller.atk())
	}
}

// --- Special legendaries (all-character AoE + summon-for-opponent) ---

// Cinder Baron deals 2 to all OTHER characters at end of turn: both heroes and
// every minion except itself. WHY: a self-anchored AreaOtherCharacters edge
// trigger must hit everyone but the source.
func TestAllOtherCharactersAoE(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "baron", "cinder_baron", 7, 5, true)
	if ok, msg := m.EndTurn(a); !ok {
		t.Fatalf("end turn should work: %s", msg)
	}
	st := lastState(t, a)
	if st.Self.HeroHP != 28 || st.Opp.HeroHP != 28 {
		t.Fatalf("both heroes should take 2, got self=%d opp=%d", st.Self.HeroHP, st.Opp.HeroHP)
	}
	if baron := boardMinion(m, 0, "baron"); baron == nil || baron.health != 5 {
		t.Fatalf("baron must not damage itself, health=%v", baron)
	}
}

// Rotgut Horror's finalGasp deals 2 to ALL characters when it dies. WHY: the
// all-character AoE must fire on death and hit both heroes.
func TestAllCharactersFinalGasp(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "rot", "rotgut_horror", 0, 1, false) // enemy, 0 atk so no retaliation, 1 hp
	place(m, 0, "mote", "mote", 1, 1, true)          // my attacker
	if ok, msg := m.Attack(a, "mote", "rot"); !ok {
		t.Fatalf("attack should kill rotgut: %s", msg)
	}
	st := lastState(t, a)
	if st.Self.HeroHP != 28 || st.Opp.HeroHP != 28 {
		t.Fatalf("finalGasp should deal 2 to both heroes, self=%d opp=%d", st.Self.HeroHP, st.Opp.HeroHP)
	}
	if len(st.Self.Board) != 0 {
		t.Fatalf("the 1/1 attacker should die to the 2-damage AoE, board=%d", len(st.Self.Board))
	}
}

// The Gorehound's finalGasp summons a 3/3 for the OPPONENT. Placed on the enemy
// board, its opponent is me — so the whelp lands on my side. WHY: SummonForOpponent
// must summon on the dying minion's opponent's board.
func TestSummonForOpponent(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "beast", "the_gorehound", 0, 1, false) // enemy gorehound, 0 atk, 1 hp
	place(m, 0, "mote", "mote", 1, 1, true)
	if ok, msg := m.Attack(a, "mote", "beast"); !ok {
		t.Fatalf("attack should kill the gorehound: %s", msg)
	}
	st := lastState(t, a)
	found := false
	for _, mn := range st.Self.Board {
		if mn.CardID == "gorehound_whelp" {
			found = true
			if mn.Attack != 3 || mn.Health != 3 {
				t.Fatalf("whelp should be 3/3, got %d/%d", mn.Attack, mn.Health)
			}
		}
	}
	if !found {
		t.Fatal("a 3/3 should be summoned on MY board (the dying gorehound's opponent)")
	}
}

// Voidwyrm Tyrant's onset destroys every OTHER minion (both boards) and
// discards the caster's remaining hand — but spares itself. WHY: a self-anchored
// destroy-all-other must not kill the played minion, and DiscardHand must empty
// the rest of the hand that played it.
func TestVoidwyrmTyrantDestroyAllAndDiscard(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "ally", "mote", 1, 1, true)            // my other minion
	place(m, 1, "foe1", "granite_warden", 1, 7, false) // enemy minions
	place(m, 1, "foe2", "mote", 1, 1, false)
	m.state[0].hand = []cards.Card{getCard("voidwyrm_tyrant"), getCard("mote"), getCard("mend")}
	m.state[0].mana, m.state[0].maxMana = 10, 10
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play should resolve: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Self.Board) != 1 || st.Self.Board[0].CardID != "voidwyrm_tyrant" {
		t.Fatalf("only the tyrant should remain on my board, got %d: %+v", len(st.Self.Board), st.Self.Board)
	}
	if st.Self.Board[0].Health != 12 || st.Self.Board[0].Attack != 12 {
		t.Fatalf("the tyrant must be unscathed 12/12, got %d/%d", st.Self.Board[0].Attack, st.Self.Board[0].Health)
	}
	if len(st.Opp.Board) != 0 {
		t.Fatalf("all enemy minions should be destroyed, got %d", len(st.Opp.Board))
	}
	if len(st.Self.Hand) != 0 {
		t.Fatalf("the rest of the hand should be discarded, got %d cards", len(st.Self.Hand))
	}
}

// Cragmaw gains +1/+1 at the end of EVERY turn, not just its controller's. WHY:
// the both-turns trigger (OnAnyTurnEnd) must fire globally — one growth on my turn
// end and another on the opponent's.
func TestCragmawGrowsEachTurn(t *testing.T) {
	m, a, b := newMatch()
	place(m, 0, "crag", "cragmaw", 7, 7, true)
	if ok, msg := m.EndTurn(a); !ok { // my turn ends → +1/+1
		t.Fatalf("end my turn: %s", msg)
	}
	if crag := boardMinion(m, 0, "crag"); crag.atk() != 8 || crag.maxHP() != 8 {
		t.Fatalf("after my turn end cragmaw should be 8/8, got %d/%d", crag.atk(), crag.maxHP())
	}
	if ok, msg := m.EndTurn(b); !ok { // opponent's turn ends → +1/+1 again
		t.Fatalf("end opponent turn: %s", msg)
	}
	if crag := boardMinion(m, 0, "crag"); crag.atk() != 9 || crag.maxHP() != 9 {
		t.Fatalf("after the opponent's turn end cragmaw should be 9/9, got %d/%d", crag.atk(), crag.maxHP())
	}
}

// Revenant Priestess's onset resummons the caster's minions that died THIS
// turn — and only friendly ones, as fresh base copies. WHY: died-this-turn must
// track the caster's own deaths and bring them back without buffs.
func TestRevenantPriestessResummonsDeadThisTurn(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "ally", "mote", 1, 1, true) // my 1/1
	place(m, 1, "foe", "mote", 1, 1, false) // enemy 1/1
	if ok, msg := m.Attack(a, "ally", "foe"); !ok {
		t.Fatalf("trade should resolve: %s", msg) // both 1/1s die this turn
	}
	m.state[0].hand = []cards.Card{getCard("revenant_priestess")}
	m.state[0].mana, m.state[0].maxMana = 6, 6
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play priestess: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Self.Board) != 2 { // priestess + the resummoned mote
		t.Fatalf("board should hold priestess + 1 resummoned mote, got %d: %+v", len(st.Self.Board), st.Self.Board)
	}
	motes := 0
	for _, mn := range st.Self.Board {
		if mn.CardID == "mote" {
			motes++
			if mn.Attack != 1 || mn.Health != 1 {
				t.Fatalf("resummoned mote should be a fresh 1/1, got %d/%d", mn.Attack, mn.Health)
			}
		}
	}
	if motes != 1 {
		t.Fatalf("exactly one friendly mote should be resummoned (not the enemy's), got %d", motes)
	}
	if len(st.Opp.Board) != 0 {
		t.Fatalf("the enemy's dead minion must NOT be resummoned for me, opp board=%d", len(st.Opp.Board))
	}
}

// A minion that died on a PREVIOUS turn is not resummoned: the died-this-turn
// window resets each turn start. WHY: "this turn" must be scoped to the current
// turn, else `revenant_priestess` would resurrect stale deaths.
func TestRevenantPriestessIgnoresPriorTurnDeaths(t *testing.T) {
	m, a, b := newMatch()
	place(m, 0, "ally", "mote", 1, 1, true)
	place(m, 1, "foe", "mote", 1, 1, false)
	if ok, msg := m.Attack(a, "ally", "foe"); !ok { // both die this turn
		t.Fatalf("trade: %s", msg)
	}
	if ok, msg := m.EndTurn(a); !ok { // pass to opponent
		t.Fatalf("end my turn: %s", msg)
	}
	if ok, msg := m.EndTurn(b); !ok { // back to me — the death is now last turn's
		t.Fatalf("end opp turn: %s", msg)
	}
	m.state[0].hand = []cards.Card{getCard("revenant_priestess")}
	m.state[0].mana, m.state[0].maxMana = 6, 6
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play priestess: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Self.Board) != 1 || st.Self.Board[0].CardID != "revenant_priestess" {
		t.Fatalf("only the priestess should be on board (no stale resummon), got %d: %+v", len(st.Self.Board), st.Self.Board)
	}
}

// --- Random-generation (filtered random card from the pool) ---

// Codex of Insight adds a random MAGE SPELL to hand. WHY: the random-generate
// pool filter (class + type) must only ever yield a card matching both.
func TestRandomGenerateMageSpell(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = []cards.Card{getCard("codex_of_insight")}
	m.state[0].mana, m.state[0].maxMana = 1, 1
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play codex: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Self.Hand) != 1 {
		t.Fatalf("hand should hold exactly the generated card, got %d", len(st.Self.Hand))
	}
	gen := getCard(st.Self.Hand[0].CardID)
	if gen.Class != cards.ClassMage || gen.Type != cards.TypeSpell {
		t.Fatalf("generated card must be a Mage spell, got %s (%s/%s)", gen.ID, gen.Class, gen.Type)
	}
}

// Gleamwing's onset adds a random LEGENDARY MINION to hand. WHY: the rarity +
// type filter must only yield a legendary minion.
func TestRandomGenerateLegendaryMinion(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = []cards.Card{getCard("gleamwing")}
	m.state[0].mana, m.state[0].maxMana = 2, 2
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play gleamwing: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Self.Hand) != 1 {
		t.Fatalf("hand should hold the generated legendary, got %d", len(st.Self.Hand))
	}
	gen := getCard(st.Self.Hand[0].CardID)
	if gen.Rarity != cards.RarityLegendary || gen.Type != cards.TypeMinion {
		t.Fatalf("generated card must be a legendary minion, got %s (%s/%s)", gen.ID, gen.Rarity, gen.Type)
	}
}

// Wilds Beastcaller summons a random BEAST onto the caster's board. WHY: the
// tribe-filtered random summon must land an actual Beast, not just any minion.
func TestRandomSummonBeast(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = []cards.Card{getCard("wilds_beastcaller")}
	m.state[0].mana, m.state[0].maxMana = 7, 7
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play beastcaller: %s", msg)
	}
	st := lastState(t, a)
	if len(st.Self.Board) != 2 {
		t.Fatalf("board should hold the beastcaller + 1 summoned beast, got %d", len(st.Self.Board))
	}
	summoned := 0
	for _, mn := range st.Self.Board {
		if mn.CardID == "wilds_beastcaller" {
			continue
		}
		summoned++
		if getCard(mn.CardID).Tribe != cards.TribeBeast {
			t.Fatalf("the summoned minion must be a Beast, got %s", mn.CardID)
		}
	}
	if summoned != 1 {
		t.Fatalf("exactly one beast should be summoned, got %d", summoned)
	}
}

// Sprocket Tinkerer transforms ANOTHER random minion into one of its two tokens,
// and never itself. WHY: a self-anchored random transform must pick from the two
// declared outcomes and exclude the source.
func TestRandomTransform(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "victim", "granite_warden", 1, 7, false) // the only other minion
	m.state[0].hand = []cards.Card{getCard("sprocket_tinkerer")}
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play tinkerer: %s", msg)
	}
	v := boardMinion(m, 1, "victim")
	if v == nil {
		t.Fatal("victim should still occupy its slot (transformed in place)")
	}
	if v.card.ID != "thornback_saurian" && v.card.ID != "bramble_squirrel" {
		t.Fatalf("victim must become one of the two tokens, got %s", v.card.ID)
	}
	var tink *minion
	for _, mn := range m.state[0].board {
		if mn.card.ID == "sprocket_tinkerer" {
			tink = mn
		}
	}
	if tink == nil || tink.atk() != 3 || tink.maxHP() != 3 {
		t.Fatalf("the tinkerer must not transform itself, got %+v", tink)
	}
}

// --- Set-hero-Health (`emberqueen_valtha`) ---

// Emberqueen Valtha's onset sets the ENEMY hero's Health to 15 — lowering a
// full hero. WHY: a set-health effect must overwrite the hero's HP regardless of
// its current value, and TargetHero must let the enemy hero be chosen.
func TestSetHeroHealthLowersEnemy(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = []cards.Card{getCard("emberqueen_valtha")}
	m.state[0].mana, m.state[0].maxMana = 9, 9
	if ok, msg := m.PlayCard(a, 0, "oppHero"); !ok {
		t.Fatalf("play valtha on enemy hero: %s", msg)
	}
	st := lastState(t, a)
	if st.Opp.HeroHP != 15 {
		t.Fatalf("enemy hero should be set to 15, got %d", st.Opp.HeroHP)
	}
	if st.Self.HeroHP != 30 {
		t.Fatalf("my hero must be untouched at 30, got %d", st.Self.HeroHP)
	}
}

// Targeting your own wounded hero raises it to 15. WHY: set-health also heals up
// to the set value, and TargetHero must allow the friendly hero.
func TestSetHeroHealthRaisesFriendly(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = 5
	m.state[0].hand = []cards.Card{getCard("emberqueen_valtha")}
	m.state[0].mana, m.state[0].maxMana = 9, 9
	if ok, msg := m.PlayCard(a, 0, "selfHero"); !ok {
		t.Fatalf("play valtha on my hero: %s", msg)
	}
	st := lastState(t, a)
	if st.Self.HeroHP != 15 {
		t.Fatalf("my wounded hero should be raised to 15, got %d", st.Self.HeroHP)
	}
}

// --- Cost modification (Phase F) ---

// Mana Leech makes ALL minions (both players') cost 1 more, but leaves non-minions
// alone. WHY: a CostAura with Scope=all + Type=minion must hit every minion card
// and nothing else.
func TestCostAuraRaisesAllMinions(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "leech", "mana_leech", 2, 2, true)
	imp := getCard("pebble_imp")
	if c := m.effectiveCost(0, imp); c != imp.Cost+1 {
		t.Fatalf("my minion should cost +1, got %d (base %d)", c, imp.Cost)
	}
	if c := m.effectiveCost(1, imp); c != imp.Cost+1 {
		t.Fatalf("the enemy's minion should also cost +1 (scope all), got %d", c)
	}
	bolt := getCard("cinder_bolt")
	if c := m.effectiveCost(0, bolt); c != bolt.Cost {
		t.Fatalf("a spell must be unaffected by a minion cost aura, got %d", c)
	}
}

// Arcane Adept reduces only the controller's spells by 1, floored at 0, leaving
// the enemy's spells alone. WHY: a friendly-scoped, type-restricted CostAura with
// the cost floor.
func TestCostAuraReducesFriendlySpellsFloored(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "adept", "arcane_adept", 3, 2, true)
	mend := getCard("mend") // a 1-cost spell
	if c := m.effectiveCost(0, mend); c != 0 {
		t.Fatalf("my 1-cost spell should floor to 0, got %d", c)
	}
	if c := m.effectiveCost(1, mend); c != mend.Cost {
		t.Fatalf("the enemy's spell must be unaffected (friendly scope), got %d", c)
	}
	if c := m.effectiveCost(0, getCard("pebble_imp")); c != getCard("pebble_imp").Cost {
		t.Fatalf("a minion must be unaffected by a spell cost aura, got %d", c)
	}
	// A Secret is a spell subtype (TypeSecret only routes it to the hidden zone), so
	// "your spells cost (1) less" must discount it too.
	ward := getCard("glacial_ward") // a 3-cost secret
	if c := m.effectiveCost(0, ward); c != ward.Cost-1 {
		t.Fatalf("my secret should get the spell discount, got %d want %d", c, ward.Cost-1)
	}
}

// Tidecolossus costs 1 less per minion on either board, floored at 0. WHY: an
// intrinsic CostRule reads live board state.
func TestSeaGiantCostByBoard(t *testing.T) {
	m, _, _ := newMatch()
	giant := getCard("tidecolossus") // base 10
	if c := m.effectiveCost(0, giant); c != 10 {
		t.Fatalf("with an empty board the giant costs full, got %d", c)
	}
	place(m, 0, "a", "mote", 1, 1, true)
	place(m, 0, "b", "mote", 1, 1, true)
	place(m, 1, "c", "mote", 1, 1, true) // 3 minions total, both boards
	if c := m.effectiveCost(0, giant); c != 7 {
		t.Fatalf("3 minions should drop the giant to 7, got %d", c)
	}
}

// Spellwarden Magus's onset makes the next Secret this turn cost 0, consumed
// by that one play. WHY: a one-shot per-player cost flag that overrides cost and
// clears after use.
func TestSpellwardenMagusFreeSecret(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = []cards.Card{getCard("spellwarden_magus")}
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play magus: %s", msg)
	}
	if !m.state[0].nextSecretFree {
		t.Fatal("onset should set the free-secret flag")
	}
	snare := getCard("snare") // a 2-cost secret
	if c := m.effectiveCost(0, snare); c != 0 {
		t.Fatalf("the next secret should cost 0, got %d", c)
	}
	m.state[0].hand = []cards.Card{getCard("snare")}
	m.state[0].mana = 1 // less than snare's real cost — only payable because it's free
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play free secret: %s", msg)
	}
	if m.state[0].mana != 1 {
		t.Fatalf("a free secret must spend 0 mana, mana=%d", m.state[0].mana)
	}
	if m.state[0].nextSecretFree {
		t.Fatal("the flag must be consumed after one secret")
	}
}

// Pocket Conjurer discounts only the controller's FIRST minion each turn. WHY: a
// conditional CostAura gated on minionsPlayedThisTurn.
func TestPintSizedFirstMinionDiscount(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "conj", "pocket_conjurer", 2, 2, true)
	imp := getCard("pebble_imp") // base 1
	if c := m.effectiveCost(0, imp); c != 0 {
		t.Fatalf("the first minion should be discounted to 0, got %d", c)
	}
	m.state[0].hand = []cards.Card{getCard("pebble_imp"), getCard("pebble_imp")}
	m.state[0].mana, m.state[0].maxMana = 5, 5
	if ok, msg := m.PlayCard(a, 0, ""); !ok { // first minion: free
		t.Fatalf("play first minion: %s", msg)
	}
	if c := m.effectiveCost(0, imp); c != imp.Cost {
		t.Fatalf("the second minion should cost full, got %d", c)
	}
}

// --- Conditional spell / tribe-cond / cross-turn cost ---

// Glacial Splinter deals 2 and draws a card ONLY if the target was Frozen. WHY: a
// conditional effect must read the target's pre-damage status.
func TestGlacialSplinterDrawsIfFrozen(t *testing.T) {
	// Frozen target → draw.
	m, a, _ := newMatch()
	place(m, 1, "frosty", "granite_warden", 1, 7, false)
	boardMinion(m, 1, "frosty").frozen = true
	m.state[0].hand = []cards.Card{getCard("glacial_splinter")}
	m.state[0].mana, m.state[0].maxMana = 2, 2
	if ok, msg := m.PlayCard(a, 0, "frosty"); !ok {
		t.Fatalf("cast splinter on frozen: %s", msg)
	}
	if h := boardMinion(m, 1, "frosty"); h == nil || h.health != 5 {
		t.Fatalf("target should take 2 damage (7→5), got %v", h)
	}
	if n := len(m.state[0].hand); n != 1 { // spell left hand (0), then drew 1
		t.Fatalf("a Frozen target should trigger the draw (hand=1), got %d", n)
	}

	// Non-frozen target → no draw.
	m2, a2, _ := newMatch()
	place(m2, 1, "warm", "granite_warden", 1, 7, false)
	m2.state[0].hand = []cards.Card{getCard("glacial_splinter")}
	m2.state[0].mana, m2.state[0].maxMana = 2, 2
	if ok, msg := m2.PlayCard(a2, 0, "warm"); !ok {
		t.Fatalf("cast splinter on unfrozen: %s", msg)
	}
	if n := len(m2.state[0].hand); n != 0 {
		t.Fatalf("an unfrozen target must NOT draw (hand=0), got %d", n)
	}
}

// Shellback Crab destroys a Gilkin and gains +2/+2; with no Gilkin it fizzles
// (plays as a vanilla 1/2, no buff). WHY: ReqTribe gates the target, and the
// self-buff rider only fires when the onset resolved.
func TestShellbackCrabDestroysGilkinAndBuffs(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "murk", "brackish_caller", 1, 2, false) // enemy Gilkin
	m.state[0].hand = []cards.Card{getCard("shellback_crab")}
	m.state[0].mana, m.state[0].maxMana = 1, 1
	if ok, msg := m.PlayCard(a, 0, "murk"); !ok {
		t.Fatalf("play crab on gilkin: %s", msg)
	}
	if boardMinion(m, 1, "murk") != nil {
		t.Fatal("the Gilkin should be destroyed")
	}
	crab := boardMinion(m, 0, "")
	for _, mn := range m.state[0].board {
		if mn.card.ID == "shellback_crab" {
			crab = mn
		}
	}
	if crab == nil || crab.atk() != 3 || crab.maxHP() != 4 {
		t.Fatalf("the crab should be 3/4 after +2/+2, got %v", crab)
	}

	// No Gilkin present → fizzle: vanilla 1/2, no buff.
	m2, a2, _ := newMatch()
	place(m2, 1, "beast", "mote", 1, 1, false) // not a Gilkin
	m2.state[0].hand = []cards.Card{getCard("shellback_crab")}
	m2.state[0].mana, m2.state[0].maxMana = 1, 1
	if ok, msg := m2.PlayCard(a2, 0, ""); !ok {
		t.Fatalf("play crab with no gilkin: %s", msg)
	}
	if boardMinion(m2, 1, "beast") == nil {
		t.Fatal("a non-Gilkin must not be destroyed")
	}
	if mn := m2.state[0].board[0]; mn.atk() != 1 || mn.maxHP() != 2 {
		t.Fatalf("a fizzled crab must stay 1/2, got %d/%d", mn.atk(), mn.maxHP())
	}
}

// Fizzle Sparkmuddle makes the opponent's spells cost 0 on their NEXT turn only.
// WHY: a cross-turn, opponent-side cost flag keyed to a specific turnNum.
func TestFizzleSparkmuddleEnemySpellsFreeNextTurn(t *testing.T) {
	m, a, b := newMatch()
	m.state[0].hand = []cards.Card{getCard("fizzle_sparkmuddle")}
	m.state[0].mana, m.state[0].maxMana = 2, 2
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play fizzle_sparkmuddle: %s", msg)
	}
	mend := getCard("mend") // a 1-cost spell
	if c := m.effectiveCost(1, mend); c != mend.Cost {
		t.Fatalf("the discount must not apply on MY turn yet, got %d", c)
	}
	if ok, msg := m.EndTurn(a); !ok { // → opponent's turn, where their spells are free
		t.Fatalf("end my turn: %s", msg)
	}
	if c := m.effectiveCost(1, mend); c != 0 {
		t.Fatalf("the opponent's spells should be free this turn, got %d", c)
	}
	if ok, msg := m.EndTurn(b); !ok { // a turn later the discount is gone
		t.Fatalf("end opp turn: %s", msg)
	}
	if c := m.effectiveCost(1, mend); c != mend.Cost {
		t.Fatalf("the discount should expire after one turn, got %d", c)
	}
}

// --- `decoy_ward` + misc one-offs ---

// Decoy Ward summons a 1/3 and redirects an enemy spell onto it, sparing the
// original target. WHY: a retarget secret must mutate the in-flight spell's target.
func TestDecoyWardRetargetsSpell(t *testing.T) {
	m, a, _ := newMatch()
	placeSecret(m, 1, "decoy_ward")                        // opponent's secret
	place(m, 1, "mine", "granite_warden", 1, 7, false)     // opponent's minion = the chosen target
	m.state[0].hand = []cards.Card{getCard("cinder_bolt")} // 3 dmg any
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, msg := m.PlayCard(a, 0, "mine"); !ok {
		t.Fatalf("cast bolt at enemy minion: %s", msg)
	}
	if mn := boardMinion(m, 1, "mine"); mn == nil || mn.health != 7 {
		t.Fatalf("the original target must be unharmed (redirected), got %v", mn)
	}
	if len(m.state[1].secrets) != 0 {
		t.Fatal("the secret should be revealed + consumed")
	}
	// The 1/3 decoy soaked the 3-damage bolt and died, so the board has only "mine".
	if got := len(m.state[1].board); got != 1 {
		t.Fatalf("the decoy should have soaked the bolt and died, board=%d", got)
	}
}

// A `decoy_ward`-style secret does NOT fire on an untargeted spell or a spell aimed
// at the hero. WHY: the retarget condition requires a minion target owned by the
// secret's controller.
func TestDecoyWardIgnoresHeroSpell(t *testing.T) {
	m, a, _ := newMatch()
	placeSecret(m, 1, "decoy_ward")
	m.state[0].hand = []cards.Card{getCard("cinder_bolt")}
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, msg := m.PlayCard(a, 0, "oppHero"); !ok { // spell at the hero, not a minion
		t.Fatalf("cast bolt at hero: %s", msg)
	}
	if len(m.state[1].secrets) != 1 {
		t.Fatal("the secret must NOT fire on a hero-targeted spell")
	}
	if len(m.state[1].board) != 0 {
		t.Fatal("no decoy should be summoned")
	}
}

// Nightmare Lord summons a 2/1 Satyr after you play ANOTHER card (not itself).
// WHY: OnPlayCard fires off subsequent cards, never the played minion's own play.
func TestNightmareLordSummonsAfterCard(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "xav", "nightmare_lord", 7, 5, true)
	m.state[0].hand = []cards.Card{getCard("pebble_imp")}
	m.state[0].mana, m.state[0].maxMana = 5, 5
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play a card: %s", msg)
	}
	satyrs := 0
	for _, mn := range m.state[0].board {
		if mn.card.ID == "thornwood_satyr" {
			satyrs++
		}
	}
	if satyrs != 1 {
		t.Fatalf("exactly one Satyr should be summoned after a card, got %d", satyrs)
	}
}

// Imp Warden's two end-of-turn triggers both fire: 1 damage to itself and a 1/1
// Imp. WHY: a minion may carry multiple triggers for the same event.
func TestImpWardenEndTurn(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "warden", "imp_warden", 1, 5, true)
	if ok, msg := m.EndTurn(a); !ok {
		t.Fatalf("end turn: %s", msg)
	}
	if w := boardMinion(m, 0, "warden"); w == nil || w.health != 4 {
		t.Fatalf("warden should take 1 self-damage (5→4), got %v", w)
	}
	imps := 0
	for _, mn := range m.state[0].board {
		if mn.card.ID == "imp_whelp" {
			imps++
		}
	}
	if imps != 1 {
		t.Fatalf("one Imp should be summoned, got %d", imps)
	}
}

// Runed Golem's onset gives the opponent an empty Mana Crystal (max +1),
// capped at 10. WHY: the give-opponent-mana effect raises max mana, not current.
func TestArcaneGolemGivesOppMana(t *testing.T) {
	m, a, _ := newMatch()
	m.state[1].maxMana, m.state[1].mana = 5, 5
	m.state[0].hand = []cards.Card{getCard("runed_golem")}
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play golem: %s", msg)
	}
	if m.state[1].maxMana != 6 {
		t.Fatalf("opponent max mana should rise to 6, got %d", m.state[1].maxMana)
	}
	if m.state[1].mana != 5 {
		t.Fatalf("opponent CURRENT mana must be unchanged, got %d", m.state[1].mana)
	}
}

// --- Grant Spell Damage / give-opp-cards / rng trigger / copy-spell ---

// Runeward Sage grants adjacent minions +1 Spell Damage (not itself), and Silence
// strips it. WHY: granted Spell Damage rides in an enchant, so it adds to
// spellPower and is removed by Silence.
func TestAncientMageGrantsSpellDamage(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "left", "pebble_imp", 1, 1, true)
	place(m, 0, "right", "pebble_imp", 1, 1, true)
	m.state[0].hand = []cards.Card{getCard("runeward_sage")}
	m.state[0].mana, m.state[0].maxMana = 4, 4
	if ok, msg := m.PlayCardAt(a, 0, "", 1); !ok { // drop between the two
		t.Fatalf("play sage: %s", msg)
	}
	left, right := boardMinion(m, 0, "left"), boardMinion(m, 0, "right")
	if spellDamageOf(left) != 1 || spellDamageOf(right) != 1 {
		t.Fatalf("both neighbours should have Spell Damage +1, got %d/%d", spellDamageOf(left), spellDamageOf(right))
	}
	if m.spellPower(0) != 2 {
		t.Fatalf("controller's spell power should be 2, got %d", m.spellPower(0))
	}
	m.silence(left)
	if spellDamageOf(left) != 0 {
		t.Fatalf("Silence must strip the granted Spell Damage, got %d", spellDamageOf(left))
	}
}

// Grovelord Brakka's onset puts 2 gift cards into the OPPONENT's hand. WHY:
// EffectGenerate with ToOpponent + Count adds copies to the other player.
func TestGiveOpponentCards(t *testing.T) {
	m, a, _ := newMatch()
	m.state[1].hand = nil // start the opponent empty for a clean count
	m.state[0].hand = []cards.Card{getCard("grovelord_brakka")}
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play brakka: %s", msg)
	}
	if n := len(m.state[1].hand); n != 2 {
		t.Fatalf("the opponent should receive 2 cards, got %d", n)
	}
	for _, c := range m.state[1].hand {
		if c.ID != "jungle_gift" {
			t.Fatalf("the gifts should be Jungle Gifts, got %s", c.ID)
		}
	}
}

// A 50%-chance turn-start trigger (Lucky Angler) fires probabilistically — neither
// never nor always. WHY: Trigger.Chance must gate the effect on an RNG roll.
func TestChanceTriggerFiresSometimes(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "angler", "lucky_angler", 0, 4, true)
	deckN := 400
	m.state[0].deck = nil
	for i := 0; i < deckN; i++ {
		m.state[0].deck = append(m.state[0].deck, getCard("mote"))
	}
	for i := 0; i < deckN; i++ {
		m.fireTriggers(0, cards.OnTurnStart, nil) // each roll may draw one card
	}
	hits := deckN - len(m.state[0].deck)
	if hits <= deckN/10 || hits >= deckN*9/10 {
		t.Fatalf("a 50%% trigger should fire sometimes but not always, got %d/%d", hits, deckN)
	}
}

// Archivist Solenne (`archivist_solenne`) puts a copy of every cast spell into the
// OTHER player's hand. WHY: a spell cast by either player copies to the non-caster.
func TestArchivistSolenneCopiesSpellToOpponent(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "cho", "archivist_solenne", 0, 4, true)
	m.state[1].hand = nil
	m.state[0].hand = []cards.Card{getCard("cinder_bolt")}
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, msg := m.PlayCard(a, 0, "oppHero"); !ok {
		t.Fatalf("cast bolt: %s", msg)
	}
	if n := len(m.state[1].hand); n != 1 || m.state[1].hand[0].ID != "cinder_bolt" {
		t.Fatalf("the opponent should get a copy of the cast spell, got %+v", m.state[1].hand)
	}
}

// --- More cost auras / enrage-weapon / global turn rule ---

// Tollkeeper Brute raises only the controller's minions by 3. WHY: a friendly,
// minion-typed CostAura with a positive delta.
func TestCostAuraRaisesFriendlyMinions(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "toll", "tollkeeper_brute", 7, 6, true)
	imp := getCard("pebble_imp")
	if c := m.effectiveCost(0, imp); c != imp.Cost+3 {
		t.Fatalf("my minion should cost +3, got %d", c)
	}
	if c := m.effectiveCost(1, imp); c != imp.Cost {
		t.Fatalf("the enemy's minion must be unaffected (friendly scope), got %d", c)
	}
	if c := m.effectiveCost(0, getCard("cinder_bolt")); c != getCard("cinder_bolt").Cost {
		t.Fatalf("a spell must be unaffected, got %d", c)
	}
}

// Dread Buccaneer costs 1 less per point of the caster's weapon Attack, floored.
// WHY: an intrinsic CostRule reading the owner's weapon.
func TestDreadCorsairWeaponCost(t *testing.T) {
	m, _, _ := newMatch()
	corsair := getCard("dread_buccaneer") // base 4
	if c := m.effectiveCost(0, corsair); c != 4 {
		t.Fatalf("with no weapon it costs full, got %d", c)
	}
	m.state[0].weapon = &weaponInst{card: getCard("ember_cleaver"), attack: 3, durability: 2}
	if c := m.effectiveCost(0, corsair); c != 1 {
		t.Fatalf("a 3-attack weapon should drop it to 1, got %d", c)
	}
}

// Grudge Smith gives the controller's weapon +2 Attack while it is damaged, and
// Silence cancels it. WHY: an enrage-conditioned weapon buff read live in
// heroAttackValue.
func TestSpitefulSmithWeaponBuff(t *testing.T) {
	m, _, _ := newMatch()
	m.state[0].weapon = &weaponInst{card: getCard("ember_cleaver"), attack: 3, durability: 2}
	place(m, 0, "smith", "grudge_smith", 4, 6, true) // full health: not damaged
	if v := heroAttackValue(m.state[0]); v != 3 {
		t.Fatalf("an undamaged Smith gives no bonus, got %d", v)
	}
	boardMinion(m, 0, "smith").health = 4 // now damaged (maxHP 6)
	if v := heroAttackValue(m.state[0]); v != 5 {
		t.Fatalf("a damaged Smith should give weapon +2 (3→5), got %d", v)
	}
	m.silence(boardMinion(m, 0, "smith"))
	if v := heroAttackValue(m.state[0]); v != 3 {
		t.Fatalf("Silence must cancel the weapon buff, got %d", v)
	}
}

// Chronlord Zhal caps the turn length to 15s while in play, and Silence reverts it.
// WHY: a global turn-timer rule reads the smallest in-play TurnSeconds.
func TestChronlordZhalTurnLimit(t *testing.T) {
	m, _, _ := newMatch() // base turn limit is 75s
	if d := m.activeTurnDuration(); d.Seconds() != 75 {
		t.Fatalf("base turn length should be 75s, got %v", d)
	}
	place(m, 0, "noz", "chronlord_zhal", 8, 8, true)
	if d := m.activeTurnDuration(); d.Seconds() != 15 {
		t.Fatalf("Chronlord should cap turns to 15s, got %v", d)
	}
	m.silence(boardMinion(m, 0, "noz"))
	if d := m.activeTurnDuration(); d.Seconds() != 75 {
		t.Fatalf("Silence should revert the turn length, got %v", d)
	}
}

// --- Conditional Charge / enrage-Twinstrike ---

// Tideblade Raider has Charge only while its controller holds a weapon — so a
// freshly-summoned one can hit the enemy hero with a weapon equipped, but not
// without, and Silence removes the conditional Charge. WHY: weapon-gated keyword.
func TestTidebladeRaiderConditionalCharge(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "raider", "tideblade_raider", 2, 1, false) // summon-sick
	raider := boardMinion(m, 0, "raider")
	if m.canAttackHero(raider) {
		t.Fatal("with no weapon a summon-sick raider cannot attack")
	}
	m.state[0].weapon = &weaponInst{card: getCard("ember_cleaver"), attack: 3, durability: 2}
	if !m.canAttackHero(raider) {
		t.Fatal("with a weapon the raider should gain Charge and reach the hero")
	}
	m.silence(raider)
	if m.canAttackHero(raider) {
		t.Fatal("Silence must remove the conditional Charge")
	}
}

// Moonfury Stalker gains +1 Attack and Twinstrike while damaged — including a
// SECOND attack earned mid-combat once it drops below max health. WHY: eligibility
// reads attacksPerTurn() live (the attacksMade model), so Twinstrike gained after the
// first swing still grants the extra attack.
func TestRagingWorgenEnrageTwinstrike(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "worg", "moonfury_stalker", 3, 3, true)
	place(m, 1, "big", "granite_warden", 1, 5, false) // 1 atk → retaliates, damaging the worgen
	worg := boardMinion(m, 0, "worg")
	if worg.atk() != 3 || worg.attacksPerTurn() != 1 {
		t.Fatalf("undamaged worgen should be 3-attack, single-attack, got atk=%d per=%d", worg.atk(), worg.attacksPerTurn())
	}
	if ok, msg := m.Attack(a, "worg", "big"); !ok { // worg→big; big retaliates 1 → worg damaged
		t.Fatalf("first attack: %s", msg)
	}
	worg = boardMinion(m, 0, "worg")
	if worg == nil || worg.health != 2 {
		t.Fatalf("worgen should be at 2 health (took 1 retaliation), got %v", worg)
	}
	if worg.atk() != 4 || !worg.has(cards.KeywordTwinstrike) {
		t.Fatalf("a damaged worgen should be 4-attack with Twinstrike, got atk=%d wf=%v", worg.atk(), worg.has(cards.KeywordTwinstrike))
	}
	if ok, msg := m.Attack(a, "worg", "big"); !ok { // the Twinstrike second attack
		t.Fatalf("second (Twinstrike) attack should be allowed: %s", msg)
	}
	if boardMinion(m, 1, "big") != nil {
		t.Fatal("two worgen attacks (3 then 4) should have killed the 5-health minion")
	}
}

// --- `clockwork_swapbot` swap / `dreamwarden_ylena` Dream pool ---

// Clockwork Swapbot swaps itself with a minion in hand at turn start: the hand
// minion enters play in its slot (same uid, no onset), and the bot returns to
// hand. WHY: the swap-with-hand effect is an in-place transform plus a hand swap.
func TestClockworkSwapbotSwapsWithHand(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "bot", "clockwork_swapbot", 0, 3, true)
	m.state[0].hand = []cards.Card{getCard("granite_warden")}
	m.fireTriggers(0, cards.OnTurnStart, nil)
	b := boardMinion(m, 0, "bot")
	if b == nil || b.card.ID != "granite_warden" {
		t.Fatalf("the bot's slot should now hold the hand minion, got %v", b)
	}
	if len(m.state[0].hand) != 1 || m.state[0].hand[0].ID != "clockwork_swapbot" {
		t.Fatalf("the bot should have returned to hand, got %+v", m.state[0].hand)
	}
}

// Dreamwarden Ylena adds a random Dream card to hand at end of turn. WHY: a
// GenIDs-driven random generate draws only from the declared Dream pool.
func TestDreamwardenYlenaAddsDreamCard(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "ylena", "dreamwarden_ylena", 4, 12, true)
	m.state[0].hand = nil
	m.fireTriggers(0, cards.OnTurnEnd, nil)
	if len(m.state[0].hand) != 1 {
		t.Fatalf("a Dream card should be added, got %d", len(m.state[0].hand))
	}
	if id := m.state[0].hand[0].ID; len(id) < 6 || id[:6] != "dream_" {
		t.Fatalf("the added card should be a Dream card, got %s", id)
	}
}

// Waking Nightmare gives +5/+5 now and destroys the minion at the start of its
// owner's next turn. WHY: a delayed-destroy rider scheduled on the buffed minion.
func TestNightmareDelayedDestroy(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "v", "granite_warden", 1, 7, true)
	m.state[0].hand = []cards.Card{getCard("dream_waking_nightmare")}
	m.state[0].mana, m.state[0].maxMana = 1, 1
	if ok, msg := m.PlayCard(a, 0, "v"); !ok {
		t.Fatalf("cast nightmare: %s", msg)
	}
	v := boardMinion(m, 0, "v")
	if v == nil || v.atk() != 6 || v.maxHP() != 12 || !v.destroyAtTurnStart {
		t.Fatalf("the minion should be +5/+5 and flagged to die, got %v", v)
	}
	m.startTurn(0) // owner's next turn begins
	m.finish()     // resolve the scheduled death
	if boardMinion(m, 0, "v") != nil {
		t.Fatal("the buffed minion should be destroyed at the owner's turn start")
	}
}

// Emerald Reckoning deals 5 to all characters EXCEPT Dreamwardens. WHY: the
// ExceptCardID filter spares the named minions from an all-character AoE.
func TestEmeraldReckoningSparesDreamwardens(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "ylena", "dreamwarden_ylena", 4, 12, true)
	place(m, 1, "foe", "granite_warden", 1, 7, false)
	m.state[0].hand = []cards.Card{getCard("dream_emerald_reckoning")}
	m.state[0].mana, m.state[0].maxMana = 2, 2
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("cast reckoning: %s", msg)
	}
	st := lastState(t, a)
	if st.Self.HeroHP != 25 || st.Opp.HeroHP != 25 {
		t.Fatalf("both heroes should take 5, got self=%d opp=%d", st.Self.HeroHP, st.Opp.HeroHP)
	}
	if y := boardMinion(m, 0, "ylena"); y == nil || y.health != 12 {
		t.Fatalf("the Dreamwarden must be spared, got %v", y)
	}
	if f := boardMinion(m, 1, "foe"); f == nil || f.health != 2 {
		t.Fatalf("a non-Dreamwarden minion should take 5 (7→2), got %v", f)
	}
}

// --- Hall of Fame Wave HoF-2/HoF-3 mechanics ---

// Anguished Scribe draws a card whenever it takes damage (and not when shielded
// or silenced). WHY: an OnDamage trigger fires inside damageMinion on the
// damaged minion, from its controller's perspective.
func TestAnguishedScribeDrawsOnDamage(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "scribe", "anguished_scribe", 1, 4, true)
	m.state[0].hand = nil // room to draw
	m.damageMinion(boardMinion(m, 0, "scribe"), 2, "")
	if got := len(m.state[0].hand); got != 1 {
		t.Fatalf("taking damage should draw 1, hand=%d", got)
	}
	// A silenced scribe does not react.
	m2, _, _ := newMatch()
	place(m2, 0, "mute", "anguished_scribe", 1, 4, true)
	boardMinion(m2, 0, "mute").silenced = true
	m2.state[0].hand = nil
	m2.damageMinion(boardMinion(m2, 0, "mute"), 2, "")
	if got := len(m2.state[0].hand); got != 0 {
		t.Fatalf("a silenced scribe must not draw, hand=%d", got)
	}
}

// Mesmer Adept's onset steals a random enemy minion only when the opponent
// has 4+ minions; with 3 it fizzles. WHY: EffectMindControl with ReqOppMinions=4
// gates the steal, and a successful steal changes sides (board counts shift).
func TestMesmerAdeptMindControlGate(t *testing.T) {
	// 4 enemy minions → steal one.
	m, a, _ := newMatch()
	for _, id := range []string{"m1", "m2", "m3", "m4"} {
		place(m, 1, id, "mote", 1, 1, false)
	}
	m.state[0].hand = []cards.Card{getCard("mesmer_adept")}
	m.state[0].mana, m.state[0].maxMana = 5, 5
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play mesmer_adept: %s", msg)
	}
	if got := len(m.state[1].board); got != 3 {
		t.Fatalf("the opponent should be down to 3 minions after a steal, got %d", got)
	}
	if got := len(m.state[0].board); got != 2 { // adept + stolen minion
		t.Fatalf("the caster should have adept + stolen = 2, got %d", got)
	}

	// 3 enemy minions → no steal (ReqOppMinions not met); adept plays alone.
	m2, a2, _ := newMatch()
	for _, id := range []string{"n1", "n2", "n3"} {
		place(m2, 1, id, "mote", 1, 1, false)
	}
	m2.state[0].hand = []cards.Card{getCard("mesmer_adept")}
	m2.state[0].mana, m2.state[0].maxMana = 5, 5
	if ok, msg := m2.PlayCard(a2, 0, ""); !ok {
		t.Fatalf("play mesmer_adept (3 enemies): %s", msg)
	}
	if got := len(m2.state[1].board); got != 3 {
		t.Fatalf("with only 3 enemy minions nothing is stolen, got %d", got)
	}
	if got := len(m2.state[0].board); got != 1 { // adept only
		t.Fatalf("the caster should have only the adept, got %d", got)
	}
}

// A full caster board blocks the steal even when the gate is met. WHY: mind
// control needs a free slot — no slot means the effect fizzles, nobody moves.
func TestMindControlFizzlesOnFullBoard(t *testing.T) {
	m, _, _ := newMatch()
	for i := 0; i < maxBoard; i++ {
		place(m, 0, uid(100+i), "mote", 1, 1, true)
	}
	for _, id := range []string{"e1", "e2", "e3", "e4"} {
		place(m, 1, id, "mote", 1, 1, false)
	}
	// Resolve the finalGasp steal directly (a full board can't also play adept).
	m.applyEffect(0, &cards.Effect{Kind: cards.EffectMindControl, Target: cards.TargetNone}, charRef{}, 0, "")
	if got := len(m.state[1].board); got != 4 {
		t.Fatalf("a full caster board must block the steal, enemy=%d", got)
	}
}

// Frostward Aegis (Ice Block) prevents an otherwise-fatal hit, leaving the hero
// at its current HP and Immune for the turn; immunity clears at end of turn.
// WHY: tryIceBlock fires only on lethal, sets immune, and immune resets per turn.
func TestIceBlockPreventsLethalThenClears(t *testing.T) {
	m, a, _ := newMatch()
	m.turn = 0
	placeSecret(m, 0, "frostward_aegis")
	m.state[0].heroHP = 5
	if dealt := m.damageHero(0, 12, ""); dealt != 0 {
		t.Fatalf("a lethal hit must be fully prevented, dealt=%d", dealt)
	}
	if m.state[0].heroHP != 5 {
		t.Fatalf("the hero must keep its Health, got %d", m.state[0].heroHP)
	}
	if !m.state[0].immune {
		t.Fatal("the hero should be Immune after Ice Block fires")
	}
	if len(m.state[0].secrets) != 0 {
		t.Fatal("the secret should be consumed")
	}
	// A second hit this turn does nothing while Immune.
	if dealt := m.damageHero(0, 3, ""); dealt != 0 || m.state[0].heroHP != 5 {
		t.Fatalf("Immune must ignore further damage, dealt=%d hp=%d", dealt, m.state[0].heroHP)
	}
	// Ending the turn clears immunity.
	m.EndTurn(a)
	if m.state[0].immune {
		t.Fatal("immunity must clear at end of turn")
	}
}

// Frostlance deals 4 damage to a Frozen target instead of freezing it. WHY: the
// FrozenDamage branch must read the target's pre-cast Frozen state.
func TestFrostlanceDamagesFrozen(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 1, "ice", "granite_warden", 1, 7, false)
	boardMinion(m, 1, "ice").frozen = true
	m.state[0].hand = []cards.Card{getCard("frostlance")}
	m.state[0].mana, m.state[0].maxMana = 1, 1
	if ok, msg := m.PlayCard(a, 0, "ice"); !ok {
		t.Fatalf("cast frostlance on frozen: %s", msg)
	}
	if h := boardMinion(m, 1, "ice"); h == nil || h.health != 3 {
		t.Fatalf("a Frozen target should take 4 (7→3), got %v", h)
	}
}

// Brineseer Diviner's onset draws 2 cards for EACH player. WHY: EffectDraw
// with ToOpponent on a second trigger must fill both hands.
func TestBrineseerDrawsBoth(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = []cards.Card{getCard("brineseer_diviner")}
	m.state[1].hand = nil // opponent must have room to draw (the fixed hand starts full)
	m.state[0].mana, m.state[0].maxMana = 3, 3
	beforeOpp := len(m.state[1].hand)
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play brineseer: %s", msg)
	}
	if got := len(m.state[0].hand); got != 2 { // diviner left hand (0), drew 2
		t.Fatalf("the caster should draw 2, hand=%d", got)
	}
	if got := len(m.state[1].hand) - beforeOpp; got != 2 {
		t.Fatalf("the opponent should also draw 2, delta=%d", got)
	}
}

// Magma Behemoth costs 1 less per missing point of the caster's hero Health,
// floored at 0. WHY: a PerMissingHealth CostRule reads live hero HP.
func TestMagmaBehemothCostByMissingHealth(t *testing.T) {
	m, _, _ := newMatch()
	beast := getCard("magma_behemoth") // base 20
	if c := m.effectiveCost(0, beast); c != 20 {
		t.Fatalf("at full Health the behemoth costs full, got %d", c)
	}
	m.state[0].heroHP = 5 // 25 missing
	if c := m.effectiveCost(0, beast); c != 0 {
		t.Fatalf("25 missing Health should floor the cost to 0, got %d", c)
	}
}

// Brinelord Gorrak gains +1 Attack per OTHER Gilkin in play and not from
// itself or non-Gilkins. WHY: SelfCountAtk filters by tribe, excluding self.
func TestBrinelordSelfCountAttack(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "gorrak", "brinelord_gorrak", 2, 4, true)
	if g := boardMinion(m, 0, "gorrak"); g.atk() != 2 {
		t.Fatalf("alone, gorrak is a 2-attack, got %d", g.atk())
	}
	place(m, 0, "fish1", "brackish_caller", 1, 2, true) // a Gilkin
	place(m, 1, "fish2", "brackish_caller", 1, 2, true) // an enemy Gilkin still counts (in play)
	place(m, 0, "wisp", "mote", 1, 1, true)             // not a Gilkin
	m.refreshAuras()
	if g := boardMinion(m, 0, "gorrak"); g.atk() != 4 {
		t.Fatalf("gorrak should be 2 + 2 other Gilkins = 4, got %d", g.atk())
	}
}

// Corsair Macaw draws a random Pirate from the caster's deck. WHY: EffectTutorTribe
// pulls a tribe-matching card out of the deck into hand.
func TestCorsairMacawTutorsPirate(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].deck = testDeck([]string{"clay_acolyte", "tidereaver", "mote"}) // one Pirate
	m.state[0].hand = []cards.Card{getCard("corsair_macaw")}
	m.state[0].mana, m.state[0].maxMana = 2, 2
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play macaw: %s", msg)
	}
	drewPirate := false
	for _, c := range m.state[0].hand {
		if c.ID == "tidereaver" {
			drewPirate = true
		}
	}
	if !drewPirate {
		t.Fatal("the macaw should have tutored the Pirate into hand")
	}
	for _, c := range m.state[0].deck {
		if c.ID == "tidereaver" {
			t.Fatal("the tutored Pirate must be removed from the deck")
		}
	}
}

// Start of Game: an all-even deck holding Duskwarden Genmar makes the starting
// Hero Power cost 1; an all-odd deck holding Lunar Devourer upgrades it. WHY: the
// deck-parity passives fire from applyStartOfGame off the full deck.
func TestStartOfGameDeckParity(t *testing.T) {
	// All-even deck with the even-passive legendary → Hero Power costs 1.
	m, _, _ := newMatch()
	evenDeck := []cards.Card{getCard("duskwarden_genmar"), getCard("mote")} // costs 6, 2
	m.applyStartOfGame(0, evenDeck)
	if m.state[0].heroPower.Cost != 1 {
		t.Fatalf("an all-even deck should make the Hero Power cost 1, got %d", m.state[0].heroPower.Cost)
	}

	// All-odd deck with the odd-passive legendary → upgraded Hero Power (deal 2).
	m2, _, _ := newMatch()
	oddDeck := []cards.Card{getCard("lunar_devourer"), getCard("pebble_imp")} // costs 9, 1
	m2.applyStartOfGame(0, oddDeck)
	if hp := m2.state[0].heroPower; hp.Effect == nil || hp.Effect.Amount != 2 {
		t.Fatalf("an all-odd deck should upgrade the Hero Power to deal 2, got %+v", hp.Effect)
	}

	// A mixed deck triggers neither passive.
	m3, _, _ := newMatch()
	m3.applyStartOfGame(0, []cards.Card{getCard("duskwarden_genmar"), getCard("pebble_imp")}) // 6 (even) + 1 (odd)
	if m3.state[0].heroPower.Cost == 1 {
		t.Fatal("a mixed-parity deck must not trigger the even passive")
	}
}

// Crag Colossus costs 1 less per OTHER card in hand, floored at 0. WHY: a
// PerCardInHand CostRule reads the caster's live hand size, excluding itself.
func TestCragColossusCostByHand(t *testing.T) {
	m, _, _ := newMatch()
	giant := getCard("crag_colossus") // base 12
	m.state[0].hand = []cards.Card{giant}
	if c := m.effectiveCost(0, giant); c != 12 {
		t.Fatalf("alone in hand the giant costs full, got %d", c)
	}
	m.state[0].hand = []cards.Card{giant, getCard("mote"), getCard("mote"), getCard("mote")} // 3 others
	if c := m.effectiveCost(0, giant); c != 9 {
		t.Fatalf("3 other cards should drop it to 9, got %d", c)
	}
}

// Warhorn Chieftain's onset gives BOTH players a random Warhorn Anthem card
// in hand. WHY: the two on_play triggers (self + ToOpponent) must each generate.
func TestWarhornChieftainGivesBothAnthems(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = []cards.Card{getCard("warhorn_chieftain")}
	m.state[1].hand = nil // room to receive
	m.state[0].mana, m.state[0].maxMana = 5, 5
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play chieftain: %s", msg)
	}
	anthems := map[string]bool{"anthem_muster": true, "anthem_warsong": true, "anthem_ambush": true}
	selfGot := false
	for _, c := range m.state[0].hand {
		if anthems[c.ID] {
			selfGot = true
		}
	}
	if !selfGot {
		t.Fatal("the caster should receive a random anthem")
	}
	if len(m.state[1].hand) != 1 || !anthems[m.state[1].hand[0].ID] {
		t.Fatalf("the opponent should also receive a random anthem, got %v", m.state[1].hand)
	}
}

// Gearmaster Cog's onset summons a random Contraption onto the caster's
// board. WHY: EffectSummonRandom with a GenIDs pool puts a token into play.
func TestGearmasterCogSummonsContraption(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].hand = []cards.Card{getCard("gearmaster_cog")}
	m.state[0].mana, m.state[0].maxMana = 6, 6
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play cog: %s", msg)
	}
	contraptions := map[string]bool{"cog_emboldener": true, "cog_beacon": true, "cog_polymorpher": true, "cog_mender": true}
	got := false
	for _, mn := range m.state[0].board {
		if contraptions[mn.card.ID] {
			got = true
		}
	}
	if !got {
		t.Fatalf("the cog should summon a random Contraption, board=%v", m.state[0].board)
	}
}

// Wraithqueen Selvara's finalGasp steals a random enemy minion to the owner's
// side. WHY: an OnDeath EffectMindControl moves a minion across boards on death.
func TestSelvaraFinalGaspSteals(t *testing.T) {
	m, _, _ := newMatch()
	place(m, 0, "selvara", "wraithqueen_selvara", 5, 5, true)
	place(m, 1, "prey", "mote", 1, 1, false)
	boardMinion(m, 0, "selvara").health = 0 // kill it → finalGasp fires
	m.finish()
	if boardMinion(m, 1, "prey") != nil {
		t.Fatal("the enemy minion should have been stolen away")
	}
	if boardMinion(m, 0, "prey") == nil {
		t.Fatal("the stolen minion should be on the owner's board")
	}
}

// Shadowtail Familiar's onset draws only when the caster's deck is all-odd.
// WHY: a ReqDeckAllOdd gate on the draw effect.
func TestShadowtailOddDeckDraw(t *testing.T) {
	// All-odd deck → draw.
	m, a, _ := newMatch()
	m.state[0].deck = testDeck([]string{"pebble_imp", "thicket_stalker"}) // costs 1, 3 — all odd
	m.state[0].hand = []cards.Card{getCard("shadowtail_familiar")}
	m.state[0].mana, m.state[0].maxMana = 3, 3
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("play familiar (odd deck): %s", msg)
	}
	if got := len(m.state[0].hand); got != 1 { // familiar left (0), drew 1
		t.Fatalf("an all-odd deck should draw 1, hand=%d", got)
	}

	// Mixed deck → no draw.
	m2, a2, _ := newMatch()
	m2.state[0].deck = testDeck([]string{"pebble_imp", "mote"}) // 1 (odd), 2 (even)
	m2.state[0].hand = []cards.Card{getCard("shadowtail_familiar")}
	m2.state[0].mana, m2.state[0].maxMana = 3, 3
	if ok, msg := m2.PlayCard(a2, 0, ""); !ok {
		t.Fatalf("play familiar (mixed deck): %s", msg)
	}
	if got := len(m2.state[0].hand); got != 0 {
		t.Fatalf("a mixed-parity deck must not draw, hand=%d", got)
	}
}

// TestRelayIntentToOpponentOnly locks #2's invariant: an acting player's ephemeral
// aiming hint reaches ONLY the viewers who see them as the opponent (the other seat),
// never echoes back to the sender, and is dropped for a non-seated caller. It must
// never touch authoritative state — verified by the relay carrying its own message
// kind, decoupled from the snapshot/event log.
func TestRelayIntentToOpponentOnly(t *testing.T) {
	m, a, b := newMatch()
	aMsgs, bMsgs := len(a.msgs), len(b.msgs)

	m.RelayIntent(a, protocol.OppIntent{Type: protocol.TypeOppIntent, HoverHand: 2, AimFrom: "hand:2", AimTo: "oppHero"})

	if len(a.msgs) != aMsgs {
		t.Fatalf("intent must not echo to the sender, got %d new msgs", len(a.msgs)-aMsgs)
	}
	if len(b.msgs) != bMsgs+1 {
		t.Fatalf("opponent should receive exactly one intent, got %d", len(b.msgs)-bMsgs)
	}
	var oi protocol.OppIntent
	if err := json.Unmarshal(b.msgs[len(b.msgs)-1], &oi); err != nil {
		t.Fatalf("decode opp_intent: %v", err)
	}
	if oi.Type != protocol.TypeOppIntent || oi.HoverHand != 2 || oi.AimFrom != "hand:2" || oi.AimTo != "oppHero" {
		t.Fatalf("relayed intent mismatch: %+v", oi)
	}

	// A caller who is not a seated player (a stray/ spectator Sender) is ignored.
	stranger := &fakeSender{id: "ghost"}
	bMsgs = len(b.msgs)
	m.RelayIntent(stranger, protocol.OppIntent{Type: protocol.TypeOppIntent, HoverHand: 0})
	if len(b.msgs) != bMsgs {
		t.Fatalf("a non-seated caller's intent must be dropped, got %d", len(b.msgs)-bMsgs)
	}

	// Over-length id strings are dropped (cosmetic guard, no real id is that long).
	bMsgs = len(b.msgs)
	m.RelayIntent(a, protocol.OppIntent{Type: protocol.TypeOppIntent, HoverHand: -1, AimTo: string(make([]byte, 65))})
	if len(b.msgs) != bMsgs {
		t.Fatalf("over-length intent must be dropped, got %d", len(b.msgs)-bMsgs)
	}
}

// Arcane Barrage fires three 1-damage missiles among ENEMY characters only,
// never the caster's own board. WHY: an enemy-scoped missile spell must spare
// the caster's hero and minions even when they are valid "any character" targets
// for the all-chars variant. With no enemy minions, all three hit the enemy hero.
func TestEnemyMissilesHitOnlyEnemies(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "ally", "granite_warden", 1, 7, true) // a friendly minion that must be spared
	setHandSolo(m, 0, "arcane_barrage")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("arcane barrage should play: %s", msg)
	}
	st := lastState(t, a)
	if got := heroMaxHP - st.Opp.HeroHP; got != 3 {
		t.Fatalf("all 3 missiles should hit the enemy hero, dealt %d", got)
	}
	if st.Self.HeroHP != heroMaxHP {
		t.Fatalf("the caster's hero must not be hit, hp=%d", st.Self.HeroHP)
	}
	if boardMinion(m, 0, "ally").health != 7 {
		t.Fatalf("the caster's own minion must not be hit, hp=%d", boardMinion(m, 0, "ally").health)
	}
}

// Darkscale Mender restores 2 Health to every friendly character (hero + minions)
// and touches nothing on the enemy side. WHY: the mass-heal area is friendly-only,
// so a damaged enemy minion must stay damaged while all friendly characters mend.
func TestMassHealFriendlyCharacters(t *testing.T) {
	m, a, _ := newMatch()
	m.state[0].heroHP = heroMaxHP - 5
	place(m, 0, "ally", "granite_warden", 1, 7, true)
	boardMinion(m, 0, "ally").health = 3
	place(m, 1, "foe", "granite_warden", 1, 7, false)
	boardMinion(m, 1, "foe").health = 3
	setHandSolo(m, 0, "darkscale_mender")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("darkscale mender should play: %s", msg)
	}
	if m.state[0].heroHP != heroMaxHP-3 {
		t.Fatalf("friendly hero should heal 2 (25->27), got %d", m.state[0].heroHP)
	}
	if boardMinion(m, 0, "ally").health != 5 {
		t.Fatalf("friendly minion should heal 2 (3->5), got %d", boardMinion(m, 0, "ally").health)
	}
	if boardMinion(m, 1, "foe").health != 3 {
		t.Fatalf("enemy minion must not be healed, got %d", boardMinion(m, 1, "foe").health)
	}
}

// Frostpaw Warlord's onset grants +1/+1 for each OTHER friendly minion. WHY: the
// scaler counts the board minus the warlord itself, so with two other minions it
// becomes a 6/6 — not 4/4 (uncounted) and not 8/8 (counting itself).
func TestBuffPerOtherFriendlyMinion(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "m1", "frostpaw_grunt", 2, 2, true)
	place(m, 0, "m2", "frostpaw_grunt", 2, 2, true)
	setHandSolo(m, 0, "frostpaw_warlord")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("frostpaw warlord should play: %s", msg)
	}
	board := lastState(t, a).Self.Board
	var warlord *protocol.MinionView
	for i := range board {
		if board[i].Name == "Frostpaw Warlord" {
			warlord = &board[i]
		}
	}
	if warlord == nil {
		t.Fatal("warlord should be on board")
	}
	if warlord.Attack != 6 || warlord.Health != 6 {
		t.Fatalf("warlord should be 6/6 with two other minions, got %d/%d", warlord.Attack, warlord.Health)
	}
}

// Frostfont Elemental Freezes whatever it deals combat damage to — a struck
// minion and the enemy hero alike — while never freezing itself. WHY: the
// freeze-on-hit keyword fires off the damage the elemental DEALS, on both the
// minion-trade and face-attack paths.
func TestFreezeOnHitCombat(t *testing.T) {
	m, a, _ := newMatch()
	place(m, 0, "ele", "frostfont_elemental", 3, 6, true)
	place(m, 1, "wall", "marsh_snapjaw", 2, 7, false) // non-Taunt body that survives the hit
	if ok, msg := m.Attack(a, "ele", "wall"); !ok {
		t.Fatalf("elemental should attack the wall: %s", msg)
	}
	wall := boardMinion(m, 1, "wall")
	if wall.health != 4 || !wall.frozen {
		t.Fatalf("struck minion should be 7-3=4 and frozen, got hp=%d frozen=%v", wall.health, wall.frozen)
	}
	if ele := boardMinion(m, 0, "ele"); ele.frozen {
		t.Fatal("the elemental must not freeze itself")
	}

	// Face attack: a second elemental Freezes the enemy hero it hits.
	place(m, 0, "ele2", "frostfont_elemental", 3, 6, true)
	if ok, msg := m.Attack(a, "ele2", oppHeroTarget); !ok {
		t.Fatalf("elemental should attack the hero: %s", msg)
	}
	if !m.state[1].frozen {
		t.Fatal("enemy hero should be frozen after taking the elemental's hit")
	}
}

// Corroding Ooze destroys the opponent's weapon and draws NOTHING — unlike
// Relic Breaker, which draws cards equal to the broken weapon's Durability. WHY:
// the plain weapon-destroy is the common case; only the opt-in flag adds the draw,
// so the Ooze must leave the caster's hand untouched.
func TestDestroyWeaponNoDraw(t *testing.T) {
	m, a, _ := newMatch()
	cleaver := getCard("ember_cleaver")
	m.state[1].weapon = &weaponInst{card: cleaver, attack: cleaver.Attack, durability: cleaver.Durability}
	setHandSolo(m, 0, "corroding_ooze")
	if ok, msg := m.PlayCard(a, 0, ""); !ok {
		t.Fatalf("corroding ooze should play: %s", msg)
	}
	if m.state[1].weapon != nil {
		t.Fatal("the opponent's weapon should be destroyed")
	}
	if n := len(m.state[0].hand); n != 0 {
		t.Fatalf("plain weapon-destroy must not draw, hand=%d", n)
	}
}
