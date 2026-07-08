// Command server is the Vanillastone authoritative game server: it serves the web
// client (embedded) and the /ws WebSocket endpoint from a single binary.
package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/amvid/vanillastone/internal/auth"
	"github.com/amvid/vanillastone/internal/store"
	"github.com/amvid/vanillastone/internal/transport"
	"github.com/amvid/vanillastone/web"
)

func main() {
	dbPath := "vanillastone.db"
	if v := os.Getenv("DB_PATH"); v != "" {
		dbPath = v
	}
	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer st.Close()

	au := auth.New(st)
	ts := transport.NewServer(au, st)

	// The web build is served from this binary (same origin). The desktop build
	// runs its UI in a local webview and connects cross-origin, so its origin(s)
	// must be allow-listed via ALLOWED_ORIGINS (comma-separated, e.g.
	// "https://game.example.com", or "*" to allow any). Unset = same-origin only.
	origins := parseOrigins(os.Getenv("ALLOWED_ORIGINS"))
	if len(origins) > 0 {
		ts.SetOriginPatterns(originHosts(origins))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/register", au.HandleRegister)
	mux.HandleFunc("/login", au.HandleLogin)
	mux.HandleFunc("/pool", au.HandlePool)
	mux.HandleFunc("/starters", au.HandleStarters)
	mux.HandleFunc("/decks", au.HandleDecks)
	mux.HandleFunc("/decks/{id}", au.HandleDeck)
	mux.HandleFunc("/profile", au.HandleProfile)
	mux.HandleFunc("/leaderboard", au.HandleLeaderboard)
	mux.HandleFunc("/ws", ts.HandleWS)
	mux.Handle("/", http.FileServer(http.FS(web.FS())))

	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	log.Printf("vanillastone listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux, origins)); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// parseOrigins splits a comma-separated ALLOWED_ORIGINS list, trimming blanks.
func parseOrigins(s string) []string {
	var out []string
	for _, o := range strings.Split(s, ",") {
		if o = strings.TrimSpace(o); o != "" {
			out = append(out, o)
		}
	}
	return out
}

// originHosts reduces full origins to the host patterns coder/websocket wants
// for its Origin check ("https://app.example.com" -> "app.example.com"). "*" is
// passed through to allow any origin.
func originHosts(origins []string) []string {
	hosts := make([]string, 0, len(origins))
	for _, o := range origins {
		if o == "*" {
			hosts = append(hosts, "*")
			continue
		}
		if u, err := url.Parse(o); err == nil && u.Host != "" {
			hosts = append(hosts, u.Host)
		} else {
			hosts = append(hosts, o)
		}
	}
	return hosts
}

// withCORS adds cross-origin headers for the desktop build's HTTP calls when the
// request Origin is allow-listed. With no configured origins it is a no-op, so
// the same-origin web build is unaffected.
func withCORS(next http.Handler, origins []string) http.Handler {
	if len(origins) == 0 {
		return next
	}
	any := false
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		if o == "*" {
			any = true
		}
		allowed[o] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; origin != "" && (any || ok) {
			allow := origin
			if any {
				allow = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", allow)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
