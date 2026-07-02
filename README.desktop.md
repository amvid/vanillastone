# Desktop build (Deno Desktop)

The desktop app is the **same** React client as the web version, packaged into a
native binary via [Deno Desktop](https://deno.com/blog/v2.9#deno-desktop). The Go
server stays authoritative and unchanged — the desktop app is just a webview that
connects to a remote server over WebSocket + HTTP.

## How it differs from the web build

| | Web build | Desktop build |
|---|---|---|
| Client host | Go binary (`go:embed static/`) | Deno webview (`desktop/main.ts`) |
| Server URL | same origin (`/ws`, `/pool`, …) | `VITE_SERVER_URL` (baked at build) |
| Server needs CORS | no | yes (`ALLOWED_ORIGINS`) |

`VITE_SERVER_URL` is read by Vite **at build time**, not at runtime. To change the
target server you rebuild the UI.

## Build

1. Build the UI pointed at your server:
   ```
   cd web && VITE_SERVER_URL=https://game.example.com deno task build && cd ..
   ```
   (unset → same-origin, i.e. the normal web build)

2. Package the desktop binary:
   ```
   cd desktop && deno desktop build main.ts
   ```
   Output: `.dmg`/`.app` (macOS), `.exe`/`.msi` (Windows), `.AppImage`/`.deb` (Linux).

## Run the server for desktop clients

The server must allow the desktop app's origin (its webview runs on a local
`http://localhost:PORT`, not your server's host). Set `ALLOWED_ORIGINS`:

```
ALLOWED_ORIGINS="*" ./server          # allow any origin (simplest)
ALLOWED_ORIGINS="http://localhost:1420,https://game.example.com" ./server
```

- Empty/unset → strict same-origin only (web build; unchanged, secure default).
- Comma-separated full origins, or `*`.
- Applies to both HTTP (CORS headers) and the `/ws` upgrade (Origin check).

The exact webview origin depends on Deno Desktop; start with `*` to confirm the
end-to-end flow, then tighten to the real origin once known.
