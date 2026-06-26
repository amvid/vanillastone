// Command server is the Vanillastone authoritative game server: it serves the web
// client (embedded) and the /ws WebSocket endpoint from a single binary.
package main

import (
	"log"
	"net/http"
	"os"

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

	mux := http.NewServeMux()
	mux.HandleFunc("/register", au.HandleRegister)
	mux.HandleFunc("/login", au.HandleLogin)
	mux.HandleFunc("/pool", au.HandlePool)
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
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}
