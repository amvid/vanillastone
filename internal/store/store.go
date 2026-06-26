// Package store is the SQLite persistence layer. Phase: accounts only (users).
// Decks come later (Phase 9). Uses modernc.org/sqlite (pure Go, no CGO) so the
// distroless static prod build (CGO_ENABLED=0) keeps working.
package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

// eloK is the Elo K-factor: the max rating a single game can move a player.
// baseRating is every player's hidden starting rating (lazy-created on first
// ranked game). Ratings are internal — the client only ever sees ladder rank.
const (
	eloK       = 24
	baseRating = 1000
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

// Store wraps the SQLite connection pool. ratingMu serializes the ladder's
// read-modify-write (RecordResult) so two games finishing at once can't compute
// Elo off a stale rating.
type Store struct {
	db       *sql.DB
	ratingMu sync.Mutex
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
CREATE TABLE IF NOT EXISTS ratings (
	username TEXT PRIMARY KEY,
	rating   INTEGER NOT NULL DEFAULT 1000
);
CREATE TABLE IF NOT EXISTS results (
	username TEXT NOT NULL,
	class    TEXT NOT NULL,
	wins     INTEGER NOT NULL DEFAULT 0,
	losses   INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (username, class)
);
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

// ClassRecord is a player's win/loss tally for one class.
type ClassRecord struct {
	Class  string
	Wins   int
	Losses int
}

// Profile is a player's ranked-stats view. Rating is the hidden Elo; Rank is the
// ladder position (1 = top). Rank/Ranked are zero/false for a player who has not
// finished a ranked game. Overall W/L is the sum across Classes.
type Profile struct {
	Username string
	Ranked   bool
	Rank     int
	Rating   int
	Wins     int
	Losses   int
	Classes  []ClassRecord
}

// LeaderRow is one leaderboard entry (rank assigned by the caller from order).
type LeaderRow struct {
	Username string
	Wins     int
	Losses   int
}

// RecordResult applies one ranked game to the persistent stats: it updates both
// players' hidden Elo ratings (zero-sum, lazy-created at baseRating) and bumps the
// per-class win/loss tallies. winnerClass/loserClass are the deck classes each
// played. All writes run in one transaction so a crash can't half-apply a result.
func (s *Store) RecordResult(winner, loser, winnerClass, loserClass string) error {
	s.ratingMu.Lock()
	defer s.ratingMu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rw, err := ratingTx(tx, winner)
	if err != nil {
		return err
	}
	rl, err := ratingTx(tx, loser)
	if err != nil {
		return err
	}
	nw, nl := elo(rw, rl)
	if err := setRatingTx(tx, winner, nw); err != nil {
		return err
	}
	if err := setRatingTx(tx, loser, nl); err != nil {
		return err
	}
	if err := bumpResultTx(tx, winner, winnerClass, true); err != nil {
		return err
	}
	if err := bumpResultTx(tx, loser, loserClass, false); err != nil {
		return err
	}
	return tx.Commit()
}

// elo returns the post-game ratings for a winner rw and loser rl. The swing is
// zero-sum (winner +d, loser -d) and at least 1 so every game moves the ladder.
func elo(rw, rl int) (int, int) {
	exp := 1.0 / (1.0 + math.Pow(10, float64(rl-rw)/400.0))
	d := int(math.Round(eloK * (1 - exp)))
	if d < 1 {
		d = 1
	}
	return rw + d, rl - d
}

func ratingTx(tx *sql.Tx, username string) (int, error) {
	var r int
	err := tx.QueryRow(`SELECT rating FROM ratings WHERE username = ?`, username).Scan(&r)
	if errors.Is(err, sql.ErrNoRows) {
		return baseRating, nil
	}
	return r, err
}

func setRatingTx(tx *sql.Tx, username string, rating int) error {
	_, err := tx.Exec(
		`INSERT INTO ratings (username, rating) VALUES (?, ?)
		 ON CONFLICT(username) DO UPDATE SET rating = excluded.rating`,
		username, rating,
	)
	return err
}

func bumpResultTx(tx *sql.Tx, username, class string, win bool) error {
	col := "losses"
	if win {
		col = "wins"
	}
	_, err := tx.Exec(
		`INSERT INTO results (username, class, wins, losses)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(username, class) DO UPDATE SET `+col+` = `+col+` + 1`,
		username, class, b2i(win), b2i(!win),
	)
	return err
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// GetProfile returns a player's ranked stats. A player with no ranked games (or
// an unknown username) comes back unranked with empty class records.
func (s *Store) GetProfile(username string) (Profile, error) {
	p := Profile{Username: username, Rating: baseRating}
	err := s.db.QueryRow(`SELECT rating FROM ratings WHERE username = ?`, username).Scan(&p.Rating)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return p, nil // unranked
	case err != nil:
		return Profile{}, err
	}
	p.Ranked = true
	// Rank = 1 + players strictly ahead in the ladder order (rating desc, then
	// username asc as the tie-break), so every player has a unique rank that matches
	// the leaderboard ordering — no two players ever share a position.
	if err := s.db.QueryRow(
		`SELECT COUNT(*) + 1 FROM ratings WHERE rating > ? OR (rating = ? AND username < ?)`,
		p.Rating, p.Rating, username,
	).Scan(&p.Rank); err != nil {
		return Profile{}, err
	}
	rows, err := s.db.Query(
		`SELECT class, wins, losses FROM results WHERE username = ? ORDER BY class`, username,
	)
	if err != nil {
		return Profile{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var c ClassRecord
		if err := rows.Scan(&c.Class, &c.Wins, &c.Losses); err != nil {
			return Profile{}, err
		}
		p.Classes = append(p.Classes, c)
		p.Wins += c.Wins
		p.Losses += c.Losses
	}
	return p, rows.Err()
}

// TopPlayers returns the top limit players by rating (ties broken by username),
// each with their overall win/loss tally. The caller assigns ranks by order.
func (s *Store) TopPlayers(limit int) ([]LeaderRow, error) {
	rows, err := s.db.Query(
		`SELECT r.username,
		        COALESCE((SELECT SUM(wins)   FROM results WHERE username = r.username), 0),
		        COALESCE((SELECT SUM(losses) FROM results WHERE username = r.username), 0)
		 FROM ratings r
		 ORDER BY r.rating DESC, r.username ASC
		 LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LeaderRow
	for rows.Next() {
		var lr LeaderRow
		if err := rows.Scan(&lr.Username, &lr.Wins, &lr.Losses); err != nil {
			return nil, err
		}
		out = append(out, lr)
	}
	return out, rows.Err()
}

// Rank returns a player's current ladder position (1 = top), or 0 if the player
// has no ranked games yet. Cheap lookup for the in-match nameplate.
func (s *Store) Rank(username string) int {
	var rating int
	if err := s.db.QueryRow(`SELECT rating FROM ratings WHERE username = ?`, username).Scan(&rating); err != nil {
		return 0
	}
	var rank int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) + 1 FROM ratings WHERE rating > ? OR (rating = ? AND username < ?)`,
		rating, rating, username,
	).Scan(&rank); err != nil {
		return 0
	}
	return rank
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
