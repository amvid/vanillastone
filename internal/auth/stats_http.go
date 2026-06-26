package auth

import (
	"net/http"
	"strings"
)

// leaderboardSize is how many top players the leaderboard returns.
const leaderboardSize = 10

// classStat is one class's win/loss/winrate row in a profile.
type classStat struct {
	Class   string `json:"class"`
	Wins    int    `json:"wins"`
	Losses  int    `json:"losses"`
	Winrate int    `json:"winrate"` // whole-percent, 0 when no games
}

// profileJSON is a player's public stats. Rating (the hidden Elo) is never
// exposed — only the ladder rank derived from it.
type profileJSON struct {
	Username string      `json:"username"`
	Ranked   bool        `json:"ranked"`
	Rank     int         `json:"rank"` // 0 when unranked
	Wins     int         `json:"wins"`
	Losses   int         `json:"losses"`
	Winrate  int         `json:"winrate"`
	Classes  []classStat `json:"classes"`
}

// leaderRowJSON is one leaderboard entry.
type leaderRowJSON struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
	Winrate  int    `json:"winrate"`
}

// winrate returns the whole-percent win rate, 0 for a player with no games.
func winrate(wins, losses int) int {
	games := wins + losses
	if games == 0 {
		return 0
	}
	return wins * 100 / games
}

// HandleProfile handles GET /profile?user=<name>: anyone may view any player's
// ranked stats (no auth required — same as the public card pool).
func (a *Auth) HandleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET only")
		return
	}
	user := strings.TrimSpace(r.URL.Query().Get("user"))
	if user == "" {
		writeErr(w, http.StatusBadRequest, "user required")
		return
	}
	p, err := a.store.GetProfile(user)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "server error")
		return
	}
	classes := make([]classStat, len(p.Classes))
	for i, c := range p.Classes {
		classes[i] = classStat{Class: c.Class, Wins: c.Wins, Losses: c.Losses, Winrate: winrate(c.Wins, c.Losses)}
	}
	writeJSON(w, http.StatusOK, profileJSON{
		Username: p.Username,
		Ranked:   p.Ranked,
		Rank:     p.Rank,
		Wins:     p.Wins,
		Losses:   p.Losses,
		Winrate:  winrate(p.Wins, p.Losses),
		Classes:  classes,
	})
}

// HandleLeaderboard handles GET /leaderboard: the top players by hidden rating,
// ranked 1..N. Public, no auth.
func (a *Auth) HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET only")
		return
	}
	rows, err := a.store.TopPlayers(leaderboardSize)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "server error")
		return
	}
	out := make([]leaderRowJSON, len(rows))
	for i, lr := range rows {
		out[i] = leaderRowJSON{
			Rank:     i + 1,
			Username: lr.Username,
			Wins:     lr.Wins,
			Losses:   lr.Losses,
			Winrate:  winrate(lr.Wins, lr.Losses),
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"players": out})
}
