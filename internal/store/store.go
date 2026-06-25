// Package store is the SQLite persistence layer. Phase: accounts only (users).
// Decks come later (Phase 9). Uses modernc.org/sqlite (pure Go, no CGO) so the
// distroless static prod build (CGO_ENABLED=0) keeps working.
package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	_ "modernc.org/sqlite"
)

// MaxDecksPerUser caps how many decks a single account may save (Phase 9).
const MaxDecksPerUser = 15

// ErrUsernameTaken is returned when a username already exists.
var ErrUsernameTaken = errors.New("username taken")

// ErrNotFound is returned when a user (or deck) does not exist.
var ErrNotFound = errors.New("not found")

// ErrDeckLimit is returned when a user already has MaxDecksPerUser decks.
var ErrDeckLimit = errors.New("deck limit reached")

// User is a stored account. PasswordHash is a bcrypt hash, never plaintext.
type User struct {
	ID           int64
	Username     string
	PasswordHash string
}

// Store wraps the SQLite connection pool.
type Store struct {
	db *sql.DB
}

// Deck is a saved deck: a named list of card ids owned by a user. Cards holds
// the card ids (duplicates allowed); deck-legality is validated by the cards
// package before persistence, not here.
type Deck struct {
	ID       int64
	Username string
	Name     string
	Class    string
	Cards    []string
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	username      TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS decks (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	username   TEXT NOT NULL,
	name       TEXT NOT NULL,
	class      TEXT NOT NULL DEFAULT 'mage',
	cards      TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_decks_username ON decks (username);
`

// Open opens (or creates) the database at path and runs migrations. Use a file
// path for persistence; ":memory:" works for tests if MaxOpenConns is 1.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// SQLite handles one writer at a time; keep the pool small and predictable.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	// Migrate older databases whose decks table predates the class column.
	// ADD COLUMN errors if it already exists; that error is expected and ignored.
	if _, err := db.Exec(`ALTER TABLE decks ADD COLUMN class TEXT NOT NULL DEFAULT 'mage'`); err != nil &&
		!strings.Contains(err.Error(), "duplicate column name") {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// CreateUser inserts a new user. Returns ErrUsernameTaken on a unique conflict.
func (s *Store) CreateUser(username, passwordHash string) error {
	_, err := s.db.Exec(
		`INSERT INTO users (username, password_hash) VALUES (?, ?)`,
		username, passwordHash,
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint") {
		return ErrUsernameTaken
	}
	return err
}

// GetUser looks up a user by username. Returns ErrNotFound if absent.
func (s *Store) GetUser(username string) (User, error) {
	var u User
	err := s.db.QueryRow(
		`SELECT id, username, password_hash FROM users WHERE username = ?`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

// ListDecks returns all decks owned by username, oldest first.
func (s *Store) ListDecks(username string) ([]Deck, error) {
	rows, err := s.db.Query(
		`SELECT id, username, name, class, cards FROM decks WHERE username = ? ORDER BY id`,
		username,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var decks []Deck
	for rows.Next() {
		d, err := scanDeck(rows)
		if err != nil {
			return nil, err
		}
		decks = append(decks, d)
	}
	return decks, rows.Err()
}

// GetDeck returns one deck by id, scoped to username so a user cannot read
// another's deck. Returns ErrNotFound if absent.
func (s *Store) GetDeck(username string, id int64) (Deck, error) {
	row := s.db.QueryRow(
		`SELECT id, username, name, class, cards FROM decks WHERE id = ? AND username = ?`,
		id, username,
	)
	d, err := scanDeck(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Deck{}, ErrNotFound
	}
	return d, err
}

// CreateDeck saves a new deck for username. Returns ErrDeckLimit if the user
// already has MaxDecksPerUser decks. Card-legality must be validated by the
// caller before this point.
func (s *Store) CreateDeck(username, name, class string, cardIDs []string) (Deck, error) {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM decks WHERE username = ?`, username).Scan(&n); err != nil {
		return Deck{}, err
	}
	if n >= MaxDecksPerUser {
		return Deck{}, ErrDeckLimit
	}
	cardsJSON, err := json.Marshal(cardIDs)
	if err != nil {
		return Deck{}, err
	}
	res, err := s.db.Exec(
		`INSERT INTO decks (username, name, class, cards) VALUES (?, ?, ?, ?)`,
		username, name, class, string(cardsJSON),
	)
	if err != nil {
		return Deck{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Deck{}, err
	}
	return Deck{ID: id, Username: username, Name: name, Class: class, Cards: cardIDs}, nil
}

// UpdateDeck overwrites an existing deck's name and cards, scoped to username.
// Returns ErrNotFound if the deck does not exist for that user.
func (s *Store) UpdateDeck(username string, id int64, name, class string, cardIDs []string) error {
	cardsJSON, err := json.Marshal(cardIDs)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(
		`UPDATE decks SET name = ?, class = ?, cards = ?, updated_at = datetime('now') WHERE id = ? AND username = ?`,
		name, class, string(cardsJSON), id, username,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteDeck removes a deck, scoped to username. Returns ErrNotFound if absent.
func (s *Store) DeleteDeck(username string, id int64) error {
	res, err := s.db.Exec(`DELETE FROM decks WHERE id = ? AND username = ?`, id, username)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// scanDeck reads a deck row (id, username, name, cards-json) into a Deck.
func scanDeck(row interface{ Scan(...any) error }) (Deck, error) {
	var d Deck
	var cardsJSON string
	if err := row.Scan(&d.ID, &d.Username, &d.Name, &d.Class, &cardsJSON); err != nil {
		return Deck{}, err
	}
	if err := json.Unmarshal([]byte(cardsJSON), &d.Cards); err != nil {
		return Deck{}, err
	}
	return d, nil
}
