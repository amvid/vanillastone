/// <reference types="vite/client" />
import { useSyncExternalStore } from 'react'

// Art lives under public/ (not the module graph), so Vite can't HMR it. The dev
// plugin in vite.config.ts sends a custom `art-changed` event when files there
// change; we bump a version here and append it as a ?v= query to art URLs so the
// browser refetches the image — without a full page reload that would reset the
// React tree and bounce the player to the main screen.
let version = 0
const listeners = new Set<() => void>()

function bump() {
  version++
  listeners.forEach((l) => l())
}

if (import.meta.hot) {
  import.meta.hot.on('art-changed', bump)
}

// useArtVersion returns the current art version; components re-render on bump.
export function useArtVersion(): number {
  return useSyncExternalStore(
    (l) => {
      listeners.add(l)
      return () => listeners.delete(l)
    },
    () => version,
  )
}

// artUrl builds the art src for a card, cache-busted by the current version.
export function artUrl(cardId: string, version: number): string {
  return version ? `/art/${cardId}.png?v=${version}` : `/art/${cardId}.png`
}
