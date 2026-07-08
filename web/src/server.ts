// Server address resolution. Web build ships inside the Go binary, so it talks to
// its own origin (VITE_SERVER_URL unset). The desktop build has no meaningful
// origin (UI runs in a local webview) and must point at a remote Go server, set
// at build time via VITE_SERVER_URL, e.g. https://game.example.com.
const BASE = (import.meta.env.VITE_SERVER_URL ?? "").replace(/\/$/, "");

// serverURL turns an absolute server path ("/pool") into a fetchable URL.
export function serverURL(path: string): string {
  return BASE + path;
}

// wsURL returns the /ws WebSocket endpoint. Falls back to same-origin when no
// remote server is configured (the web build).
export function wsURL(): string {
  if (BASE) return BASE.replace(/^http/, "ws") + "/ws";
  const proto = location.protocol === "https:" ? "wss" : "ws";
  return `${proto}://${location.host}/ws`;
}
