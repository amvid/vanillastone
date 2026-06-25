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
