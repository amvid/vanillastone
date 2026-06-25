// Package web embeds the web client so the single server binary can serve it.
// Phase 1 ships a static test page; a real React (Vite) build under static/
// replaces it in a later step.
package web

import (
	"embed"
	"io/fs"
)

//go:embed static
var embedded embed.FS

// FS returns the client file tree rooted at the static dir.
func FS() fs.FS {
	sub, err := fs.Sub(embedded, "static")
	if err != nil {
		panic(err) // embed is compile-time fixed; failure is a build bug
	}
	return sub
}
