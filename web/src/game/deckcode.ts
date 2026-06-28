// Deck import/export codes: a short, copy-pasteable string that round-trips a
// deck (class + card ids) so players can share lists. Pure client-side — the
// pool the browser already fetched (`/pool`, buildable cards in DeckPoolIDs
// order) is the single source of truth for the card<->token mapping. Import
// feeds the decoded ids straight to createDeck, so the SERVER still validates
// legality (size, copy cap, class); this codec only moves ids around.
//
// Format:  [classChar] [entry...] [checksumChar]   — all lowercase alphanumeric
//   classChar : m=mage h=hunter r=warrior w=warlock
//   entry     : exactly 3 chars = [letter][base36 index][copies 1-9]
//               letter+index = the card's token (first letter of its id + its
//               base36 position among ids sharing that letter, ids sorted).
//   checksum  : 1 base36 char over the decoded class+ids — a stale code (built
//               against a different card pool) or a typo resolves to different
//               ids, fails the checksum, and is REJECTED rather than silently
//               importing the wrong deck.
//
// The token index is pool-relative: adding cards shifts indices, so codes are
// only valid for the pool version that made them. That is intentional and why
// the checksum exists — old codes fail loud, they never decode to garbage.
import type { CardView } from '../protocol'

const CLASS_TO_CHAR: Record<string, string> = {
  mage: 'm',
  hunter: 'h',
  warrior: 'r',
  warlock: 'w',
  priest: 'p',
}
const CHAR_TO_CLASS: Record<string, string> = Object.fromEntries(
  Object.entries(CLASS_TO_CHAR).map(([k, v]) => [v, k]),
)

// buildMaps derives the deterministic id<->token mapping from the pool. Ids are
// sorted, then numbered within their first-letter group; the token is the
// letter + that index in base36. Both directions are returned so encode and
// decode share one definition.
function buildMaps(pool: CardView[]): { idToToken: Record<string, string>; tokenToId: Record<string, string> } {
  const ids = pool.map((c) => c.cardId).sort()
  const idToToken: Record<string, string> = {}
  const tokenToId: Record<string, string> = {}
  const seen: Record<string, number> = {}
  for (const id of ids) {
    const letter = id[0]
    const n = seen[letter] ?? 0
    seen[letter] = n + 1
    if (n >= 36) {
      // A first-letter group outgrew the single base36 index char. The format
      // needs widening before this letter can encode — fail loud, don't truncate.
      throw new Error(`too many cards starting with "${letter}" to encode`)
    }
    const token = letter + n.toString(36)
    idToToken[id] = token
    tokenToId[token] = id
  }
  return { idToToken, tokenToId }
}

// checksum hashes the resolved class + sorted ids into one base36 char. It is
// computed over the semantic deck (pool-independent id strings), so a decode
// against a drifted pool yields different ids and a different checksum.
function checksum(cls: string, ids: string[]): string {
  const s = cls + '|' + [...ids].sort().join(',')
  let h = 0
  for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) >>> 0
  return (h % 36).toString(36)
}

// encodeDeck turns a deck into its share code. Throws if the class or any card
// is not in the pool (a caller bug — decks are built from this same pool).
export function encodeDeck(cls: string, cards: string[], pool: CardView[]): string {
  const classChar = CLASS_TO_CHAR[cls]
  if (!classChar) throw new Error(`unknown class: ${cls}`)
  const { idToToken } = buildMaps(pool)
  const counts: Record<string, number> = {}
  for (const id of cards) counts[id] = (counts[id] ?? 0) + 1
  const entries = Object.entries(counts).map(([id, n]) => {
    const token = idToToken[id]
    if (!token) throw new Error(`card not in pool: ${id}`)
    if (n < 1 || n > 9) throw new Error(`unsupported copy count ${n} for ${id}`)
    return token + n
  })
  entries.sort() // deterministic: same deck always yields the same code
  return classChar + entries.join('') + checksum(cls, cards)
}

// decodeDeck parses a share code back into a class + card-id list, or throws
// with a player-facing message. The returned ids are NOT validated for deck
// legality — the caller hands them to createDeck, which the server validates.
export function decodeDeck(code: string, pool: CardView[]): { class: string; cards: string[] } {
  const c = code.trim().toLowerCase()
  if (c.length < 5) throw new Error('deck code is too short')
  const cls = CHAR_TO_CLASS[c[0]]
  if (!cls) throw new Error('not a valid deck code')
  const sum = c[c.length - 1]
  const body = c.slice(1, -1)
  if (body.length % 3 !== 0) throw new Error('deck code is malformed')
  const { tokenToId } = buildMaps(pool)
  const cards: string[] = []
  for (let i = 0; i < body.length; i += 3) {
    const token = body.slice(i, i + 2)
    const n = parseInt(body[i + 2], 10)
    if (!(n >= 1 && n <= 9)) throw new Error('deck code is malformed')
    const id = tokenToId[token]
    if (!id) throw new Error('deck code is invalid or from an older card set')
    for (let k = 0; k < n; k++) cards.push(id)
  }
  if (checksum(cls, cards) !== sum) throw new Error('deck code is invalid or from an older card set')
  return { class: cls, cards }
}
