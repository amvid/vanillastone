import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { readdirSync } from 'node:fs'
import { join } from 'node:path'

// In docker, the Go server is reachable as service "server"; locally it's
// localhost. Set VITE_PROXY_TARGET in compose to point the dev proxy at it.
const proxyTarget = process.env.VITE_PROXY_TARGET ?? 'http://localhost:8080'

// Vite doesn't HMR files under public/ (they're served verbatim, not part of the
// module graph). Rather than full-reload the page (which resets all React state
// and bounces the user to the main screen), we send a custom `art-changed` event
// that the client listens for to cache-bust just the art images in place.
// chokidar's fs watcher catches edits to existing files, but on the docker-on-macOS
// bind mount it never fires `add`/`unlink` for *new*/removed files (the FUSE layer
// doesn't surface new directory entries to the poller) — which is exactly the
// case for dropping in new art. So we also poll the public/ file set ourselves
// and notify when the set changes.
function listPublicFiles(dir: string): Set<string> {
  const out = new Set<string>()
  const walk = (d: string) => {
    for (const e of readdirSync(d, { withFileTypes: true })) {
      const p = join(d, e.name)
      if (e.isDirectory()) walk(p)
      else out.add(p)
    }
  }
  try { walk(dir) } catch { /* public/ may not exist */ }
  return out
}

function reloadOnPublicChange() {
  return {
    name: 'reload-on-public-change',
    configureServer(server: any) {
      const publicDir: string = server.config.publicDir
      const notify = () => server.ws.send({ type: 'custom', event: 'art-changed' })

      // Edits to existing public files (chokidar catches these reliably).
      server.watcher.add(publicDir)
      server.watcher.on('change', (file: string) => {
        if (file.replace(/\\/g, '/').startsWith(publicDir.replace(/\\/g, '/'))) notify()
      })

      // New/removed files: poll the set, since the bind-mount watcher misses them.
      let prev = listPublicFiles(publicDir)
      const timer = setInterval(() => {
        const next = listPublicFiles(publicDir)
        if (next.size !== prev.size || [...next].some((f) => !prev.has(f))) notify()
        prev = next
      }, 1000)
      server.httpServer?.on('close', () => clearInterval(timer))
    },
  }
}

export default defineConfig({
  plugins: [react(), reloadOnPublicChange()],
  server: {
    host: true, // bind 0.0.0.0 so the container is reachable from the host
    port: 5173,
    // Bind-mounted source on macOS docker doesn't emit fs events reliably.
    watch: { usePolling: true },
    proxy: {
      '/ws': { target: proxyTarget, ws: true },
      '/register': { target: proxyTarget },
      '/login': { target: proxyTarget },
      '/pool': { target: proxyTarget },
      '/decks': { target: proxyTarget },
      '/profile': { target: proxyTarget },
      '/leaderboard': { target: proxyTarget },
    },
  },
  // Build output goes into static/, which the Go server embeds (go:embed) and
  // serves in prod. Stable (unhashed) filenames keep the committed build from
  // churning on every rebuild; cache-busting can return when it matters.
  build: {
    outDir: 'static',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        entryFileNames: 'assets/app.js',
        chunkFileNames: 'assets/[name].js',
        assetFileNames: 'assets/app.[ext]',
      },
    },
  },
})
