// Deno Desktop entrypoint. The UI is the same React app as the web build, but
// compiled with VITE_SERVER_URL baked in so it talks to a remote Go server
// instead of its own origin. This process only serves the static bundle to the
// webview — all game truth still lives on the Go server.
//
// The `static/` dir here is a copy of web/static, embedded into the binary via
// `--include static` at build time (see the `desktop` target in the Makefile).
// `make desktop` handles the copy, icon, and packaging; don't run this by hand.
//
// See README.desktop.md for the full flow.
import { serveDir } from "jsr:@std/http/file-server";

const ROOT = new URL("./static", import.meta.url).pathname;

Deno.serve((req) => {
  return serveDir(req, {
    fsRoot: ROOT,
    quiet: true,
    // SPA fallback: unknown paths render index.html so client-side routing works.
    urlRoot: "",
  }).then((res) => {
    if (res.status === 404 && req.headers.get("sec-fetch-dest") === "document") {
      return serveDir(new Request(new URL("/index.html", req.url), req), {
        fsRoot: ROOT,
        quiet: true,
      });
    }
    return res;
  });
});

// The HTTP server above is a live async task, so closing the window would leave
// the process running. Adopt the auto-opened window (the first BrowserWindow
// constructed takes it over) and quit the whole app when the user closes it.
const win = new Deno.BrowserWindow({});
win.addEventListener("close", () => Deno.exit(0));
