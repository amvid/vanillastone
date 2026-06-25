package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/store"
)

func testAuth(t *testing.T) (*Auth, string) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	a := New(st)
	if err := a.Register("alice", "password123"); err != nil {
		t.Fatalf("register: %v", err)
	}
	token, err := a.Login("alice", "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return a, token
}

// req builds a request to HandleDecks (collection) with an optional bearer token.
func decksReq(method, token string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	r := httptest.NewRequest(method, "/decks", &buf)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	return r
}

// TestDecksRequireAuth: the deck endpoints reject a missing/invalid token.
func TestDecksRequireAuth(t *testing.T) {
	a, _ := testAuth(t)
	w := httptest.NewRecorder()
	a.HandleDecks(w, decksReq(http.MethodGet, "", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unauth list should be 401, got %d", w.Code)
	}
}

// TestCreateDeckValidation: an illegal deck is rejected (400); a legal one is
// created (201) and then listed.
func TestCreateDeckValidation(t *testing.T) {
	a, token := testAuth(t)

	// Too few cards -> 400.
	w := httptest.NewRecorder()
	a.HandleDecks(w, decksReq(http.MethodPost, token, map[string]any{"name": "Bad", "cards": []string{"mote"}}))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("illegal deck should be 400, got %d (%s)", w.Code, w.Body)
	}

	// A legal default deck -> 201.
	w = httptest.NewRecorder()
	a.HandleDecks(w, decksReq(http.MethodPost, token, map[string]any{"name": "Good", "class": "mage", "cards": cards.DefaultDeck()}))
	if w.Code != http.StatusCreated {
		t.Fatalf("legal deck should be 201, got %d (%s)", w.Code, w.Body)
	}

	// A deck with no/invalid class -> 400 (class is required and must be playable).
	w = httptest.NewRecorder()
	a.HandleDecks(w, decksReq(http.MethodPost, token, map[string]any{"name": "Classless", "cards": cards.DefaultDeck()}))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("deck without a playable class should be 400, got %d (%s)", w.Code, w.Body)
	}

	// It now appears in the list alongside the starter deck seeded at
	// registration, carrying its class.
	w = httptest.NewRecorder()
	a.HandleDecks(w, decksReq(http.MethodGet, token, nil))
	var got struct {
		Decks []struct {
			Name  string   `json:"name"`
			Class string   `json:"class"`
			Cards []string `json:"cards"`
		} `json:"decks"`
	}
	json.Unmarshal(w.Body.Bytes(), &got)
	if len(got.Decks) != 2 {
		t.Fatalf("expected starter + Good, got: %+v", got.Decks)
	}
	good := got.Decks[1]
	if good.Name != "Good" || good.Class != "mage" || len(good.Cards) != cards.DeckSize {
		t.Fatalf("listed deck mismatch: %+v", got.Decks)
	}
}

// TestRegisterSeedsStarterDeck: a freshly registered account already owns one
// legal, playable Mage deck so it can queue without building anything first.
func TestRegisterSeedsStarterDeck(t *testing.T) {
	a, token := testAuth(t)
	w := httptest.NewRecorder()
	a.HandleDecks(w, decksReq(http.MethodGet, token, nil))
	var got struct {
		Decks []struct {
			Name  string   `json:"name"`
			Class string   `json:"class"`
			Cards []string `json:"cards"`
		} `json:"decks"`
	}
	json.Unmarshal(w.Body.Bytes(), &got)
	if len(got.Decks) != 1 {
		t.Fatalf("new account should own exactly one starter deck, got: %+v", got.Decks)
	}
	d := got.Decks[0]
	if d.Class != "mage" || cards.ValidateDeck(d.Cards, cards.ClassMage) != nil {
		t.Fatalf("starter deck must be a legal Mage deck: %+v", d)
	}
}

// TestDeckLimitHTTP: the 11th deck is rejected with 409.
func TestDeckLimitHTTP(t *testing.T) {
	a, token := testAuth(t)
	deck := cards.DefaultDeck()
	// Registration already seeded one starter deck, leaving room for one fewer.
	for i := 0; i < store.MaxDecksPerUser-1; i++ {
		w := httptest.NewRecorder()
		a.HandleDecks(w, decksReq(http.MethodPost, token, map[string]any{"name": "d", "class": "mage", "cards": deck}))
		if w.Code != http.StatusCreated {
			t.Fatalf("deck %d should be 201, got %d", i, w.Code)
		}
	}
	w := httptest.NewRecorder()
	a.HandleDecks(w, decksReq(http.MethodPost, token, map[string]any{"name": "extra", "class": "mage", "cards": deck}))
	if w.Code != http.StatusConflict {
		t.Fatalf("exceeding the limit should be 409, got %d", w.Code)
	}
}
