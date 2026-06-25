// Package auth handles account registration, login, and in-memory session
// tokens. Passwords are bcrypt-hashed; plaintext is never stored or logged.
// Sessions live in memory (like match state) and are lost on restart.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// starterDeckName is the name given to the curated Mage deck every new account
// is seeded with at registration. The player can rename or rebuild it later.
const starterDeckName = "Mage Starter"

// Validation bounds (basic, deliberately minimal).
const (
	minUsername = 3
	maxUsername = 20
	minPassword = 6
)

// Errors surfaced to HTTP handlers.
var (
	ErrValidation = errors.New("invalid username or password format")
	ErrTaken      = store.ErrUsernameTaken
	ErrBadCreds   = errors.New("wrong username or password")
)

// Auth owns the user store and the live session map.
type Auth struct {
	store    *store.Store
	mu       sync.Mutex
	sessions map[string]string // token -> username
}

// New returns an Auth backed by the given store.
func New(s *store.Store) *Auth {
	return &Auth{store: s, sessions: make(map[string]string)}
}

// Register validates input, hashes the password, creates the user, and seeds a
// persisted starter Mage deck the player can rebuild later.
// Returns ErrValidation or ErrTaken on failure.
func (a *Auth) Register(username, password string) error {
	if !validUsername(username) || len(password) < minPassword {
		return ErrValidation
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := a.store.CreateUser(username, string(hash)); err != nil {
		return err
	}
	_, err = a.store.CreateDeck(username, starterDeckName, string(cards.ClassMage), cards.DefaultDeck())
	return err
}

// Login verifies credentials and returns a fresh session token. Returns
// ErrBadCreds for both unknown user and wrong password (no user enumeration).
func (a *Auth) Login(username, password string) (string, error) {
	u, err := a.store.GetUser(username)
	if errors.Is(err, store.ErrNotFound) {
		return "", ErrBadCreds
	}
	if err != nil {
		return "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", ErrBadCreds
	}
	token := newToken()
	a.mu.Lock()
	a.sessions[token] = u.Username
	a.mu.Unlock()
	return token, nil
}

// Username returns the account for a session token, or ok=false if invalid.
func (a *Auth) Username(token string) (string, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	name, ok := a.sessions[token]
	return name, ok
}

func validUsername(u string) bool {
	return len(u) >= minUsername && len(u) <= maxUsername
}

// newToken returns a random 256-bit hex session token.
func newToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err) // crypto/rand failure is unrecoverable
	}
	return hex.EncodeToString(b)
}
