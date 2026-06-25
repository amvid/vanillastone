package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/protocol"
	"github.com/amvid/vanillastone/internal/store"
)

// maxDeckName bounds a deck's display name.
const maxDeckName = 40

// deckBody is the JSON payload for creating/updating a deck.
type deckBody struct {
	Name  string   `json:"name"`
	Class string   `json:"class"`
	Cards []string `json:"cards"`
}

// deckJSON is a deck as returned to the client.
type deckJSON struct {
	ID    int64    `json:"id"`
	Name  string   `json:"name"`
	Class string   `json:"class"`
	Cards []string `json:"cards"`
}

// userFromRequest extracts the authenticated username from the Authorization:
// Bearer <token> header. Returns ok=false if missing or invalid.
func (a *Auth) userFromRequest(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	token, found := strings.CutPrefix(h, "Bearer ")
	if !found || token == "" {
		return "", false
	}
	return a.Username(token)
}

// HandleDecks handles the collection endpoints: GET /decks (list) and
// POST /decks (create).
func (a *Auth) HandleDecks(w http.ResponseWriter, r *http.Request) {
	name, ok := a.userFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.listDecks(w, name)
	case http.MethodPost:
		a.createDeck(w, r, name)
	default:
		writeErr(w, http.StatusMethodNotAllowed, "GET or POST only")
	}
}

// HandleDeck handles the item endpoints: PUT /decks/{id} (update) and
// DELETE /decks/{id} (delete).
func (a *Auth) HandleDeck(w http.ResponseWriter, r *http.Request) {
	name, ok := a.userFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad deck id")
		return
	}
	switch r.Method {
	case http.MethodPut:
		a.updateDeck(w, r, name, id)
	case http.MethodDelete:
		a.deleteDeck(w, name, id)
	default:
		writeErr(w, http.StatusMethodNotAllowed, "PUT or DELETE only")
	}
}

// HandlePool handles GET /pool: the buildable card pool plus the deck rules, so
// the deckbuilder can render the collection without hardcoding it. Cards are
// returned as protocol.CardView (the same shape the client sees in hand).
func (a *Auth) HandlePool(w http.ResponseWriter, r *http.Request) {
	ids := cards.DeckPoolIDs()
	pool := make([]protocol.CardView, 0, len(ids))
	for _, id := range ids {
		if c, ok := cards.Get(id); ok {
			pool = append(pool, poolCardView(c))
		}
	}
	classes := make([]string, 0)
	for _, c := range cards.PlayableClasses() {
		classes = append(classes, string(c))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"cards":     pool,
		"deckSize":  cards.DeckSize,
		"maxCopies": cards.MaxCopies,
		"maxDecks":  store.MaxDecksPerUser,
		"classes":   classes,
	})
}

// poolCardView converts a card to the client's CardView shape (mirrors
// match.cardView, kept here to avoid importing the match package).
func poolCardView(c cards.Card) protocol.CardView {
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
	if c.Effect != nil {
		cv.Target = string(c.Effect.Target)
	} else if bc := c.Onset(); bc != nil {
		cv.Target = string(bc.Target)
	}
	return cv
}

func (a *Auth) listDecks(w http.ResponseWriter, username string) {
	decks, err := a.store.ListDecks(username)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "server error")
		return
	}
	out := make([]deckJSON, len(decks))
	for i, d := range decks {
		out[i] = deckJSON{ID: d.ID, Name: d.Name, Class: d.Class, Cards: d.Cards}
	}
	writeJSON(w, http.StatusOK, map[string]any{"decks": out})
}

func (a *Auth) createDeck(w http.ResponseWriter, r *http.Request, username string) {
	body, ok := decodeDeck(w, r)
	if !ok {
		return
	}
	d, err := a.store.CreateDeck(username, body.Name, body.Class, body.Cards)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, deckJSON{ID: d.ID, Name: d.Name, Class: d.Class, Cards: d.Cards})
	case errors.Is(err, store.ErrDeckLimit):
		writeErr(w, http.StatusConflict, "deck limit reached")
	default:
		writeErr(w, http.StatusInternalServerError, "server error")
	}
}

func (a *Auth) updateDeck(w http.ResponseWriter, r *http.Request, username string, id int64) {
	body, ok := decodeDeck(w, r)
	if !ok {
		return
	}
	switch err := a.store.UpdateDeck(username, id, body.Name, body.Class, body.Cards); {
	case err == nil:
		writeJSON(w, http.StatusOK, deckJSON{ID: id, Name: body.Name, Class: body.Class, Cards: body.Cards})
	case errors.Is(err, store.ErrNotFound):
		writeErr(w, http.StatusNotFound, "no such deck")
	default:
		writeErr(w, http.StatusInternalServerError, "server error")
	}
}

func (a *Auth) deleteDeck(w http.ResponseWriter, username string, id int64) {
	switch err := a.store.DeleteDeck(username, id); {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, store.ErrNotFound):
		writeErr(w, http.StatusNotFound, "no such deck")
	default:
		writeErr(w, http.StatusInternalServerError, "server error")
	}
}

// decodeDeck reads and validates a deck body: a non-empty name within bounds and
// a deck that passes the card rules (size, copies, legal cards).
func decodeDeck(w http.ResponseWriter, r *http.Request) (deckBody, bool) {
	var body deckBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return deckBody{}, false
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" || len(body.Name) > maxDeckName {
		writeErr(w, http.StatusBadRequest, "deck name must be 1-40 chars")
		return deckBody{}, false
	}
	if err := cards.ValidateDeck(body.Cards, cards.Class(body.Class)); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return deckBody{}, false
	}
	return body, true
}
