// HTTP auth + deck calls. All hit the Go server (proxied in dev via vite.config).
import type { CardView } from './protocol'

async function post(path: string, body: unknown): Promise<Response> {
  return fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

async function errorMessage(res: Response, fallback: string): Promise<string> {
  try {
    const data = (await res.json()) as { error?: string }
    return data.error ?? fallback
  } catch {
    return fallback
  }
}

export async function register(username: string, password: string): Promise<void> {
  const res = await post('/register', { username, password })
  if (!res.ok) throw new Error(await errorMessage(res, 'register failed'))
}

export async function login(username: string, password: string): Promise<string> {
  const res = await post('/login', { username, password })
  if (!res.ok) throw new Error(await errorMessage(res, 'login failed'))
  const data = (await res.json()) as { token: string }
  return data.token
}

// --- Decks (Phase 9). All require a Bearer session token. ---

export type Deck = { id: number; name: string; class: string; cards: string[] }
export type Pool = {
  cards: CardView[]
  deckSize: number
  maxCopies: number
  maxDecks: number
  classes: string[] // playable deck classes (the rest are "coming soon")
}

function authHeaders(token: string): HeadersInit {
  return { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }
}

// fetchPool returns the buildable card collection and the deck rules.
export async function fetchPool(): Promise<Pool> {
  const res = await fetch('/pool')
  if (!res.ok) throw new Error('failed to load card pool')
  return (await res.json()) as Pool
}

export async function listDecks(token: string): Promise<Deck[]> {
  const res = await fetch('/decks', { headers: authHeaders(token) })
  if (!res.ok) throw new Error(await errorMessage(res, 'failed to load decks'))
  const data = (await res.json()) as { decks: Deck[] | null }
  return data.decks ?? []
}

export async function createDeck(token: string, name: string, cls: string, cards: string[]): Promise<Deck> {
  const res = await fetch('/decks', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ name, class: cls, cards }),
  })
  if (!res.ok) throw new Error(await errorMessage(res, 'failed to save deck'))
  return (await res.json()) as Deck
}

export async function updateDeck(token: string, id: number, name: string, cls: string, cards: string[]): Promise<void> {
  const res = await fetch(`/decks/${id}`, {
    method: 'PUT',
    headers: authHeaders(token),
    body: JSON.stringify({ name, class: cls, cards }),
  })
  if (!res.ok) throw new Error(await errorMessage(res, 'failed to update deck'))
}

export async function deleteDeck(token: string, id: number): Promise<void> {
  const res = await fetch(`/decks/${id}`, { method: 'DELETE', headers: authHeaders(token) })
  if (!res.ok) throw new Error(await errorMessage(res, 'failed to delete deck'))
}

// --- Ranked stats (PvP-queue games only). Public reads, no token needed. ---

export type ClassStat = { class: string; wins: number; losses: number; winrate: number }
export type Profile = {
  username: string
  ranked: boolean
  rank: number // 0 = unranked
  wins: number
  losses: number
  winrate: number
  classes: ClassStat[]
}
export type LeaderRow = { rank: number; username: string; wins: number; losses: number; winrate: number }

// fetchProfile returns a player's ranked stats (rank + overall + per-class W/L).
export async function fetchProfile(user: string): Promise<Profile> {
  const res = await fetch(`/profile?user=${encodeURIComponent(user)}`)
  if (!res.ok) throw new Error(await errorMessage(res, 'failed to load profile'))
  return (await res.json()) as Profile
}

// fetchLeaderboard returns the top players by ladder rank.
export async function fetchLeaderboard(): Promise<LeaderRow[]> {
  const res = await fetch('/leaderboard')
  if (!res.ok) throw new Error(await errorMessage(res, 'failed to load leaderboard'))
  const data = (await res.json()) as { players: LeaderRow[] | null }
  return data.players ?? []
}
