// Package lobby is the single global matchmaking queue. When two players are
// queued it pairs them into a match. In-memory only; restart wipes it.
package lobby

import (
	"strconv"
	"sync"
	"time"

	"github.com/amvid/vanillastone/internal/cards"
	"github.com/amvid/vanillastone/internal/match"
)

// queued is a player waiting for an opponent, with the deck they will play.
type queued struct {
	sender match.Sender
	deck   []cards.Card
}

// Lobby holds players waiting for an opponent.
type Lobby struct {
	mu      sync.Mutex
	queue   []queued
	matchID int
}

// New returns an empty lobby.
func New() *Lobby {
	return &Lobby{}
}

// Join adds c (playing deck) to the queue. If another player is already waiting,
// it pairs them and returns a started match (with c as second player, so the
// waiter acts first). Otherwise returns nil and c waits.
func (l *Lobby) Join(c match.Sender, deck []cards.Card) *match.Match {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.queue) == 0 {
		l.queue = append(l.queue, queued{sender: c, deck: deck})
		return nil
	}
	first := l.queue[0]
	l.queue = l.queue[1:]
	l.matchID++
	m := match.New("m"+strconv.Itoa(l.matchID), first.sender, c, time.Now().UnixNano(), first.deck, deck)
	m.Start()
	return m
}

// StartMatch pairs two specific players (a direct invite, bypassing the queue)
// into a started match. a acts first. Uses the same match-id counter as the
// queue so ids never collide.
func (l *Lobby) StartMatch(a match.Sender, deckA []cards.Card, b match.Sender, deckB []cards.Card) *match.Match {
	l.mu.Lock()
	l.matchID++
	id := "m" + strconv.Itoa(l.matchID)
	l.mu.Unlock()
	m := match.New(id, a, b, time.Now().UnixNano(), deckA, deckB)
	m.Start()
	return m
}

// StartVsAI creates a started match where the human plays seat 0 (first) against
// an AI opponent on seat 1 playing botDeck (the caller picks which prebuilt deck).
// botName is the opponent's display name. The bot's opening mulligan is driven
// automatically so play begins without a second human.
func (l *Lobby) StartVsAI(human match.Sender, humanDeck, botDeck []cards.Card, botName string) *match.Match {
	l.mu.Lock()
	l.matchID++
	id := "m" + strconv.Itoa(l.matchID)
	l.mu.Unlock()
	m := match.NewVsAI(id, human, time.Now().UnixNano(), humanDeck, botDeck, botName)
	m.Start()
	m.DriveBotMulligan()
	return m
}

// Remove drops c from the queue (on disconnect before matching).
func (l *Lobby) Remove(c match.Sender) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, q := range l.queue {
		if q.sender.ID() == c.ID() {
			l.queue = append(l.queue[:i], l.queue[i+1:]...)
			return
		}
	}
}
