package store

import (
	"errors"
	"path/filepath"
	"testing"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// TestDeckCRUD covers the create/list/get/update/delete round-trip and that
// decks are scoped to their owner (a user cannot read another's deck).
func TestDeckCRUD(t *testing.T) {
	s := newStore(t)
	cards := []string{"a", "a", "b"}
	d, err := s.CreateDeck("alice", "Aggro", "mage", cards)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := s.GetDeck("alice", d.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Aggro" || got.Class != "mage" || len(got.Cards) != 3 || got.Cards[0] != "a" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	// Another user cannot see alice's deck.
	if _, err := s.GetDeck("bob", d.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-user get should be ErrNotFound, got %v", err)
	}
	// Update then re-read.
	if err := s.UpdateDeck("alice", d.ID, "Control", "mage", []string{"c"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ = s.GetDeck("alice", d.ID)
	if got.Name != "Control" || len(got.Cards) != 1 || got.Cards[0] != "c" {
		t.Fatalf("update not applied: %+v", got)
	}
	// A non-owner update is a no-op miss.
	if err := s.UpdateDeck("bob", d.ID, "x", "mage", []string{"y"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-user update should be ErrNotFound, got %v", err)
	}
	// Delete, then confirm gone.
	if err := s.DeleteDeck("alice", d.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetDeck("alice", d.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deck should be gone, got %v", err)
	}
}

// TestDeckLimit: a user may save at most MaxDecksPerUser decks.
func TestDeckLimit(t *testing.T) {
	s := newStore(t)
	for i := 0; i < MaxDecksPerUser; i++ {
		if _, err := s.CreateDeck("alice", "deck", "mage", []string{"a"}); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}
	if _, err := s.CreateDeck("alice", "one too many", "mage", []string{"a"}); !errors.Is(err, ErrDeckLimit) {
		t.Fatalf("exceeding the limit should be ErrDeckLimit, got %v", err)
	}
	// A different user is unaffected by alice's count.
	if _, err := s.CreateDeck("bob", "deck", "mage", []string{"a"}); err != nil {
		t.Fatalf("other user create should succeed: %v", err)
	}
}

// TestListDecksScoped: ListDecks returns only the caller's decks.
func TestListDecksScoped(t *testing.T) {
	s := newStore(t)
	s.CreateDeck("alice", "a1", "mage", []string{"a"})
	s.CreateDeck("alice", "a2", "mage", []string{"a"})
	s.CreateDeck("bob", "b1", "mage", []string{"a"})
	decks, err := s.ListDecks("alice")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(decks) != 2 {
		t.Fatalf("alice should have 2 decks, got %d", len(decks))
	}
}

// TestRecordResultStats verifies the ranked-stats round-trip: per-class W/L is
// tallied (overall = sum across classes), hidden Elo moves zero-sum so the winner
// outranks the loser, and a player with no games is unranked. The why: the ladder
// must order by skill (rating), not raw winrate, and the profile must split W/L by
// class without that split affecting the overall total.
func TestRecordResultStats(t *testing.T) {
	s := newStore(t)
	// alice beats bob twice (once as mage, once as hunter); bob beats alice once.
	if err := s.RecordResult("alice", "bob", "mage", "mage"); err != nil {
		t.Fatalf("record 1: %v", err)
	}
	if err := s.RecordResult("alice", "bob", "hunter", "mage"); err != nil {
		t.Fatalf("record 2: %v", err)
	}
	if err := s.RecordResult("bob", "alice", "mage", "mage"); err != nil {
		t.Fatalf("record 3: %v", err)
	}

	ap, err := s.GetProfile("alice")
	if err != nil {
		t.Fatalf("profile alice: %v", err)
	}
	// Overall = sum across classes: 2W (mage+hunter) / 1L (mage).
	if ap.Wins != 2 || ap.Losses != 1 {
		t.Fatalf("alice overall want 2-1, got %d-%d", ap.Wins, ap.Losses)
	}
	if len(ap.Classes) != 2 {
		t.Fatalf("alice should have 2 class rows, got %d", len(ap.Classes))
	}
	// alice won net more, so her hidden rating (and thus rank) must beat bob's.
	bp, _ := s.GetProfile("bob")
	if !(ap.Rating > bp.Rating) {
		t.Fatalf("winner rating should exceed loser: alice=%d bob=%d", ap.Rating, bp.Rating)
	}
	if ap.Rank != 1 || bp.Rank != 2 {
		t.Fatalf("ranks want alice=1 bob=2, got alice=%d bob=%d", ap.Rank, bp.Rank)
	}
	// Elo is zero-sum: what alice gained net, bob lost net (both started at base).
	if (ap.Rating - baseRating) != (baseRating - bp.Rating) {
		t.Fatalf("rating swing not zero-sum: alice=%d bob=%d base=%d", ap.Rating, bp.Rating, baseRating)
	}

	// A user who never played is unranked with no records.
	np, err := s.GetProfile("nobody")
	if err != nil {
		t.Fatalf("profile nobody: %v", err)
	}
	if np.Ranked || np.Rank != 0 || np.Wins != 0 || len(np.Classes) != 0 {
		t.Fatalf("unplayed user should be unranked/0-0, got %+v", np)
	}
}

// TestTopPlayersOrder: the leaderboard orders by hidden rating (skill), so a
// grinder with many net wins outranks a 1-0 player — the whole reason raw winrate
// is not the sort key.
func TestTopPlayersOrder(t *testing.T) {
	s := newStore(t)
	// grinder: many net wins. rookie: a single win (100% winrate, tiny sample).
	for i := 0; i < 10; i++ {
		s.RecordResult("grinder", "punching_bag", "mage", "mage")
	}
	s.RecordResult("rookie", "punching_bag", "mage", "mage")
	top, err := s.TopPlayers(10)
	if err != nil {
		t.Fatalf("top: %v", err)
	}
	if len(top) == 0 || top[0].Username != "grinder" {
		t.Fatalf("grinder should top the ladder over a 1-0 rookie, got %+v", top)
	}
}

// TestRanksAreUnique: two players on the SAME hidden rating must still get
// distinct ranks (tie-broken by username), so no two players ever occupy the same
// ladder position — and that order matches the leaderboard.
func TestRanksAreUnique(t *testing.T) {
	s := newStore(t)
	// alice and bob each win one game from base vs a different opponent, so both
	// land on the identical post-win rating.
	s.RecordResult("alice", "bag1", "mage", "mage")
	s.RecordResult("bob", "bag2", "mage", "mage")
	ap, _ := s.GetProfile("alice")
	bp, _ := s.GetProfile("bob")
	if ap.Rating != bp.Rating {
		t.Fatalf("test setup: expected equal ratings, got alice=%d bob=%d", ap.Rating, bp.Rating)
	}
	if ap.Rank == bp.Rank {
		t.Fatalf("tied ratings must not share a rank: alice=%d bob=%d", ap.Rank, bp.Rank)
	}
	if ap.Rank != 1 || bp.Rank != 2 {
		t.Fatalf("tie-break by username: want alice=1 bob=2, got alice=%d bob=%d", ap.Rank, bp.Rank)
	}
	// Rank order must agree with the leaderboard order.
	top, _ := s.TopPlayers(10)
	if top[0].Username != "alice" || top[1].Username != "bob" {
		t.Fatalf("leaderboard order must match ranks: %+v", top[:2])
	}
}
