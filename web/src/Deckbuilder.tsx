import { Fragment, useEffect, useMemo, useState } from 'react'
import { createPortal } from 'react-dom'
import { createDeck, deleteDeck, fetchPool, listDecks, updateDeck } from './api'
import type { Deck, Pool } from './api'
import type { CardView } from './protocol'
import { CardFace } from './game/CardFace'
import { cardColorClass } from './game/format'
import { decodeDeck, encodeDeck } from './game/deckcode'
import { useIsMobile } from './game/GameScreen'

const PER_PAGE_DESKTOP = 8 // cards per page (4 columns x 2 rows)
const PER_PAGE_MOBILE = 4 // landscape phone: one short row, no scroll
// Landscape phones are wide but very short; show a single readable row there.
const MOBILE_MQ = '(orientation: landscape) and (max-height: 600px)'

// Page size adapts to the viewport: 4 on a landscape phone, 8 elsewhere.
function usePerPage() {
  const [mobile, setMobile] = useState(() =>
    typeof window !== 'undefined' && window.matchMedia(MOBILE_MQ).matches,
  )
  useEffect(() => {
    const mq = window.matchMedia(MOBILE_MQ)
    const on = () => setMobile(mq.matches)
    mq.addEventListener('change', on)
    return () => mq.removeEventListener('change', on)
  }, [])
  return mobile ? PER_PAGE_MOBILE : PER_PAGE_DESKTOP
}

// All hero classes shown in the new-deck picker. Only those the server reports as
// playable are selectable; the rest render as "coming soon" so the roadmap is
// visible without being usable.
const CLASS_OPTIONS: { id: string; label: string }[] = [
  { id: 'mage', label: 'Mage' },
  { id: 'hunter', label: 'Hunter' },
  { id: 'warrior', label: 'Warrior' },
  { id: 'warlock', label: 'Warlock' },
]

const RARITIES = ['common', 'rare', 'epic', 'legendary'] as const
// Mana buckets for the cost filter; 7 means "7 or more".
const MANA_BUCKETS = [0, 1, 2, 3, 4, 5, 6, 7]
// Curve histogram is 0..7 where 7 aggregates everything costing 7+.
const CURVE_SLOTS = [0, 1, 2, 3, 4, 5, 6, 7]

// Tab splits the collection into the deck's class cards, neutral cards, or both.
type Tab = 'all' | 'class' | 'neutral'

// counts turns a card-id list into id -> copy count.
function counts(ids: string[]): Record<string, number> {
  const out: Record<string, number> = {}
  for (const id of ids) out[id] = (out[id] ?? 0) + 1
  return out
}

function cardClass(c: CardView): string {
  return c.class ?? 'neutral'
}

// Deckbuilder lets a player create, edit, and delete decks. A deck binds to one
// hero class (picked up front for a new deck); the collection is then filtered to
// that class + neutral, with tabs, mana/rarity filters, and a mana curve. The
// server validates every save (size, copy cap, legal class) — this UI is
// convenience only.
export function Deckbuilder(props: { token: string; onBack: () => void }) {
  const { token, onBack } = props
  const [pool, setPool] = useState<Pool | null>(null)
  const [decks, setDecks] = useState<Deck[]>([])
  const [error, setError] = useState('')
  const [page, setPage] = useState(0)
  // The deck currently being edited: its id (null = a new, unsaved deck), name,
  // class, and card-id list.
  const [editing, setEditing] = useState<{ id: number | null; name: string; class: string; cards: string[] } | null>(
    null,
  )
  // True while the new-deck class picker is open.
  const [picking, setPicking] = useState(false)
  const [tab, setTab] = useState<Tab>('all')
  // Which class the "class" tab filters to while just browsing (no deck open).
  // When editing a deck the deck's own class wins. Null = the first playable class.
  const [browseClass, setBrowseClass] = useState<string | null>(null)
  const [manaFilter, setManaFilter] = useState<number | null>(null)
  const [rarityFilter, setRarityFilter] = useState<string | null>(null)
  const [tribeFilter, setTribeFilter] = useState<string | null>(null)
  // Card previewed on the left while hovering an in-deck row, pinned to the
  // cursor's vertical level. Held in state (not pure CSS) because the in-deck list
  // scrolls (overflow:auto), which would clip an absolutely-positioned preview,
  // and because it follows the mouse Y.
  const [preview, setPreview] = useState<{ card: CardView; x: number; y: number } | null>(null)
  // Deck-code sharing: id of the deck whose code was just copied (transient
  // "Copied" tag), and the import box's text while it is open (null = closed).
  const [copiedId, setCopiedId] = useState<number | null>(null)
  const [importText, setImportText] = useState<string | null>(null)
  // Mobile: the mana/rarity/tribe filters live in a modal opened from a funnel
  // button on the rail (the inline rows are hidden to reclaim vertical space).
  const isMobile = useIsMobile()
  const [filtersOpen, setFiltersOpen] = useState(false)

  const reloadDecks = async () => setDecks(await listDecks(token))

  useEffect(() => {
    fetchPool().then(setPool).catch((e) => setError((e as Error).message))
    reloadDecks().catch((e) => setError((e as Error).message))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const byId = useMemo(() => {
    const m: Record<string, CardView> = {}
    for (const c of pool?.cards ?? []) m[c.cardId] = c
    return m
  }, [pool])

  // The class context the collection is filtered against: the editing deck's
  // class, or the browse-selected class (first playable by default) when no deck
  // is open. The class tab(s) below switch `browseClass` while browsing.
  const activeClass = editing?.class ?? browseClass ?? pool?.classes[0] ?? 'mage'

  // Filtered + sorted (ascending mana) collection driven by the tab and filters.
  const filtered = useMemo(() => {
    const cs = (pool?.cards ?? []).filter((c) => {
      const cl = cardClass(c)
      if (tab === 'class' && cl !== activeClass) return false
      if (tab === 'neutral' && cl !== 'neutral') return false
      // Editing a deck: "all" = the legal pool (neutral + the deck's class).
      // Just browsing: "all" means every class.
      if (tab === 'all' && editing && cl !== 'neutral' && cl !== activeClass) return false
      if (manaFilter !== null && (manaFilter === 7 ? c.cost < 7 : c.cost !== manaFilter)) return false
      if (rarityFilter && c.rarity !== rarityFilter) return false
      if (tribeFilter && c.tribe !== tribeFilter) return false
      return true
    })
    cs.sort((a, b) => a.cost - b.cost || a.name.localeCompare(b.name))
    return cs
  }, [pool, tab, activeClass, manaFilter, rarityFilter, tribeFilter, editing])

  // Distinct minion tribes present in the whole pool, alphabetised — drives the
  // tribe filter bar at the bottom of the book.
  const tribes = useMemo(() => {
    const set = new Set<string>()
    for (const c of pool?.cards ?? []) if (c.tribe) set.add(c.tribe)
    return [...set].sort()
  }, [pool])

  const perPage = usePerPage()
  // Any tab/filter/deck/page-size change can shrink the list — snap to page one.
  useEffect(() => setPage(0), [tab, manaFilter, rarityFilter, tribeFilter, activeClass, perPage])

  if (!pool) {
    return (
      <div className="builder-page">
        <h1>Vanillastone</h1>
        <p>{error || 'loading…'}</p>
        <button onClick={onBack}>Back to lobby</button>
      </div>
    )
  }

  const { deckSize, maxCopies, maxDecks } = pool
  const editCounts = editing ? counts(editing.cards) : {}
  const pageCount = Math.max(1, Math.ceil(filtered.length / perPage))
  const pageCards = filtered.slice(page * perPage, page * perPage + perPage)

  // Copy cap for a card: legendaries are limited to one (HS rule), others to
  // maxCopies. Server enforces this too.
  const capFor = (c: CardView) => (c.rarity === 'legendary' ? 1 : maxCopies)

  const addCard = (id: string) => {
    if (!editing) return
    const c = byId[id]
    if (!c) return
    // A deck only holds its class + neutral cards (server enforces this too).
    if (cardClass(c) !== 'neutral' && cardClass(c) !== editing.class) return
    if (editing.cards.length >= deckSize) return
    if ((editCounts[id] ?? 0) >= capFor(c)) return
    setEditing({ ...editing, cards: [...editing.cards, id] })
  }

  const removeCard = (id: string) => {
    if (!editing) return
    const i = editing.cards.indexOf(id)
    if (i < 0) return
    const next = editing.cards.slice()
    next.splice(i, 1)
    setEditing({ ...editing, cards: next })
  }

  const startNewDeck = (cls: string) => {
    setPicking(false)
    setTab('all')
    setManaFilter(null)
    setRarityFilter(null)
    setTribeFilter(null)
    setEditing({ id: null, name: 'New Deck', class: cls, cards: [] })
  }

  const save = async () => {
    if (!editing) return
    setError('')
    try {
      if (editing.id === null) await createDeck(token, editing.name.trim(), editing.class, editing.cards)
      else await updateDeck(token, editing.id, editing.name.trim(), editing.class, editing.cards)
      await reloadDecks()
      setEditing(null)
    } catch (e) {
      setError((e as Error).message)
    }
  }

  const remove = async (id: number, name: string) => {
    if (!confirm(`Delete deck "${name}"?`)) return
    setError('')
    try {
      await deleteDeck(token, id)
      await reloadDecks()
      if (editing?.id === id) setEditing(null)
    } catch (e) {
      setError((e as Error).message)
    }
  }

  // copyCode writes a deck's share code to the clipboard and flags it copied.
  const copyCode = async (d: Deck) => {
    setError('')
    try {
      await navigator.clipboard.writeText(encodeDeck(d.class, d.cards, pool.cards))
      setCopiedId(d.id)
      setTimeout(() => setCopiedId((cur) => (cur === d.id ? null : cur)), 1500)
    } catch (e) {
      setError((e as Error).message)
    }
  }

  // importCode decodes a pasted share code and saves it as a new deck. The
  // codec only resolves ids → the server still validates legality on create.
  const importCode = async () => {
    if (!importText?.trim()) return
    setError('')
    try {
      const { class: cls, cards } = decodeDeck(importText, pool.cards)
      await createDeck(token, 'Imported Deck', cls, cards)
      await reloadDecks()
      setImportText(null)
    } catch (e) {
      setError((e as Error).message)
    }
  }

  // In-deck entries, grouped + sorted by mana cost for the editor list.
  const deckEntries = Object.entries(editCounts).sort(
    (a, b) => (byId[a[0]]?.cost ?? 0) - (byId[b[0]]?.cost ?? 0),
  )

  // Mana curve of the deck being edited: a count per cost slot (7 = 7+).
  const curve = CURVE_SLOTS.map(() => 0)
  if (editing) {
    for (const id of editing.cards) {
      const c = byId[id]
      if (c) curve[Math.min(7, c.cost)]++
    }
  }
  const curveMax = Math.max(1, ...curve)
  const classLabel = (id: string) => CLASS_OPTIONS.find((o) => o.id === id)?.label ?? id

  // Filter pip groups, shared between the inline rows (desktop) and the mobile
  // filters modal so there's a single source of truth.
  const manaPips = MANA_BUCKETS.map((m) => (
    <button
      key={m}
      className={'mana-pip' + (manaFilter === m ? ' on' : '')}
      onClick={() => setManaFilter((cur) => (cur === m ? null : m))}
    >
      {m === 7 ? '7+' : m}
    </button>
  ))
  const rarityPips = RARITIES.map((r) => (
    <button
      key={r}
      className={'rarity-pip ' + r + (rarityFilter === r ? ' on' : '')}
      onClick={() => setRarityFilter((cur) => (cur === r ? null : r))}
    >
      {r[0].toUpperCase() + r.slice(1)}
    </button>
  ))
  const tribePips = tribes.map((t) => (
    <button
      key={t}
      className={'tribe-pip' + (tribeFilter === t ? ' on' : '')}
      onClick={() => setTribeFilter((cur) => (cur === t ? null : t))}
    >
      {t[0].toUpperCase() + t.slice(1)}
    </button>
  ))

  // The edit panel renders inline directly under the deck row being edited (or,
  // for a brand-new deck whose id is still null, under the New/Import actions).
  const deckEditPanel = editing && (
    <div className="deck-edit">
      <div className="edit-head">
        <input value={editing.name} onChange={(e) => setEditing({ ...editing, name: e.target.value })} />
        <span className={'meter' + (editing.cards.length === deckSize ? ' full' : '')}>
          {editing.cards.length}/{deckSize}
        </span>
      </div>
      <div className="edit-class">{classLabel(editing.class)} deck</div>

      {/* Mana curve */}
      <div className="curve" title="Mana curve">
        {CURVE_SLOTS.map((slot, i) => (
          <div key={slot} className="curve-col">
            <span className="curve-n">{curve[i] || ''}</span>
            <div className="curve-bar" style={{ height: `${(curve[i] / curveMax) * 100}%` }} />
            <span className="curve-cost">{slot === 7 ? '7+' : slot}</span>
          </div>
        ))}
      </div>

      <div className="edit-actions">
        <button disabled={editing.cards.length !== deckSize || !editing.name.trim()} onClick={save}>
          Save
        </button>
        <button onClick={() => setEditing(null)}>Cancel</button>
      </div>

      <div className="in-deck">
        {deckEntries.length === 0 && <p className="empty">Empty — click cards to add.</p>}
        {deckEntries.map(([id, n]) => {
          const card = byId[id]
          const r = card?.rarity
          return (
            <button
              key={id}
              className={'deck-card' + (card ? cardColorClass(card) : '')}
              onClick={() => removeCard(id)}
              onMouseEnter={(e) =>
                card &&
                setPreview({
                  card,
                  x: (e.currentTarget.closest('.deck-panel') ?? e.currentTarget).getBoundingClientRect().left,
                  y: e.clientY,
                })
              }
              onMouseMove={(e) => setPreview((p) => (p ? { ...p, y: e.clientY } : p))}
              onMouseLeave={() => setPreview(null)}
              title="Remove one"
            >
              <span className="cost">{card?.cost}</span>
              <span className={'dc-name' + (r ? ` rarity-${r}` : '')}>{card?.name}</span>
              {n > 1 && <span className="dc-count">×{n}</span>}
            </button>
          )
        })}
      </div>
    </div>
  )

  return (
    <div className="builder-page">
      <div className="builder-inner">
      <header className="builder-head">
        <button className="back-btn" onClick={onBack}>
          ‹ Lobby
        </button>
        <h1>Collection</h1>
        {error && <span className="error">{error}</span>}
      </header>

      <div className="builder">
        {/* Collection book */}
        <main className="collection">
          {/* Mana + rarity filters — one compact row above the book. */}
          <div className="col-filters">
            <div className="filter-row">{manaPips}</div>
            <div className="filter-row">{rarityPips}</div>
            <span className="col-count">{filtered.length} cards</span>
          </div>

          {/* Vertical class rail on the left + the book. Classes show their hero
             portrait; All/Neutral are text tiles. */}
          <div className="book-area">
            <div className="book-tabs">
              <button
                className={'tab-tile' + (tab === 'all' ? ' on' : '')}
                title="All"
                aria-label="All"
                onClick={() => setTab('all')}
              >
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor" aria-hidden="true">
                  <rect x="3" y="3" width="7" height="7" rx="1.5" />
                  <rect x="14" y="3" width="7" height="7" rx="1.5" />
                  <rect x="3" y="14" width="7" height="7" rx="1.5" />
                  <rect x="14" y="14" width="7" height="7" rx="1.5" />
                </svg>
              </button>
              {(editing ? [editing.class] : pool.classes).map((cls) => (
                <button
                  key={cls}
                  className={'tab-tile tab-portrait' + (tab === 'class' && activeClass === cls ? ' on' : '')}
                  style={{ backgroundImage: `url('/art/${cls}_hero.png')` }}
                  title={classLabel(cls)}
                  aria-label={classLabel(cls)}
                  onClick={() => {
                    setBrowseClass(cls)
                    setTab('class')
                  }}
                />
              ))}
              <button
                className={'tab-tile' + (tab === 'neutral' ? ' on' : '')}
                title="Neutral"
                aria-label="Neutral"
                onClick={() => setTab('neutral')}
              >
                <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                  <path d="M12 2.5 20 7 V17 L12 21.5 4 17 V7 Z" strokeLinejoin="round" />
                </svg>
              </button>
              {/* Page + card count ride the rail's right end on a landscape phone
                  (hidden on desktop, where they live in their own rows). */}
              <span className="rail-meta">
                {/* Mobile: funnel button opens the filters modal (sits before the page). */}
                <button className="filter-btn" onClick={() => setFiltersOpen(true)} aria-label="Filters">
                  <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor" aria-hidden="true">
                    <path d="M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z" />
                  </svg>
                  {(manaFilter !== null || rarityFilter !== null || tribeFilter !== null) && (
                    <span className="filter-dot" aria-hidden="true" />
                  )}
                </button>
                <span className="rail-page">{page + 1} / {pageCount}</span>
                <span className="rail-count">{filtered.length} cards</span>
              </span>
            </div>

            <div className="book">
              {/* Page nav lives on the book itself (full-height left/right edges)
                  instead of a separate button row below — saves vertical space,
                  same on desktop. */}
              <button
                className="book-nav prev"
                disabled={page === 0}
                onClick={() => setPage((p) => Math.max(0, p - 1))}
                aria-label="Previous page"
              >
                ‹
              </button>
              <button
                className="book-nav next"
                disabled={page >= pageCount - 1}
                onClick={() => setPage((p) => Math.min(pageCount - 1, p + 1))}
                aria-label="Next page"
              >
                ›
              </button>
              <div className="book-grid">
              {pageCards.map((c) => {
                const n = editCounts[c.cardId] ?? 0
                const maxed = !editing || n >= capFor(c) || editing.cards.length >= deckSize
                return (
                  <button
                    key={c.cardId}
                    className={'card book-card' + cardColorClass(c)}
                    disabled={maxed}
                    onClick={() => addCard(c.cardId)}
                  >
                    <CardFace card={c} />
                    {n > 0 && <div className="owned">×{n}</div>}
                  </button>
                )
              })}
              {/* Pad a short (or empty) page with invisible slots so the grid keeps a
                 full page's width and height (never shrinks/reflows) — including when
                 zero cards match, where these hold the columns up under the message. */}
              {Array.from({ length: perPage - pageCards.length }, (_, i) => (
                <div key={`pad-${i}`} className="book-card placeholder" aria-hidden="true" />
              ))}
              {pageCards.length === 0 && <p className="empty-collection">No cards match these filters.</p>}
            </div>
            {/* Tribe filter — pinned to the bottom edge of the book. */}
            {tribes.length > 0 && <div className="tribe-bar">{tribePips}</div>}
            </div>
          </div>
          <div className="book-page">
            {page + 1} / {pageCount}
          </div>
        </main>

        {/* Deck panel */}
        <aside className="deck-panel">
          <h3>
            Your decks ({decks.length}/{maxDecks})
          </h3>
          <div className="deck-list">
            {/* New / Import sit at the top of the list. */}
            <div className="deck-actions">
              <button className="new-deck" disabled={decks.length >= maxDecks} onClick={() => setPicking(true)}>
                {isMobile ? 'New' : '+ New deck'}
              </button>
              <button
                className="new-deck"
                disabled={decks.length >= maxDecks}
                onClick={() => setImportText((t) => (t === null ? '' : null))}
              >
                {isMobile ? 'Import' : 'Import code'}
              </button>
            </div>
            {importText !== null && (
              <div className="import-row">
                <input
                  value={importText}
                  autoFocus
                  placeholder="Paste deck code…"
                  onChange={(e) => setImportText(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && importCode()}
                />
                <button disabled={!importText.trim()} onClick={importCode}>
                  Import
                </button>
              </div>
            )}
            {/* New deck (no existing row to attach to): panel opens under the actions. */}
            {editing?.id === null && deckEditPanel}

            {decks.map((d) => (
              <Fragment key={d.id}>
                <div className="deck-row">
                  <button onClick={() => setEditing({ id: d.id, name: d.name, class: d.class, cards: d.cards })}>
                    <span
                      className="deck-row-art"
                      style={{ backgroundImage: `url('/art/${d.class}_hero.png')` }}
                    />
                    <span className="deck-row-text">{d.name}</span>
                  </button>
                  <button className="copy-code" onClick={() => copyCode(d)} title="Copy deck code">
                    {copiedId === d.id ? '✓' : '⧉'}
                  </button>
                  <button className="del" onClick={() => remove(d.id, d.name)} title="Delete">
                    ✕
                  </button>
                </div>
                {/* Edit panel opens inline, right under the deck being edited. */}
                {editing?.id === d.id && deckEditPanel}
              </Fragment>
            ))}
          </div>

          {!editing && <p className="hint">Pick a deck to edit, or create a new one.</p>}
        </aside>
      </div>
      </div>

      {/* New-deck class picker. Portaled to <body> so the fixed overlay escapes
         .builder-page's scale() transform — otherwise the backdrop resolves
         against the scaled box and the page bg leaks at the edges. */}
      {picking &&
        createPortal(
          <div className="overlay" onClick={() => setPicking(false)}>
            <div className="class-picker" onClick={(e) => e.stopPropagation()}>
              <h2>Choose a class</h2>
              <div className="class-grid">
                {CLASS_OPTIONS.map((o) => {
                  const enabled = pool.classes.includes(o.id)
                  return (
                    <button
                      key={o.id}
                      className={'class-card' + (enabled ? '' : ' soon')}
                      disabled={!enabled}
                      onClick={() => enabled && startNewDeck(o.id)}
                    >
                      {enabled && (
                        <span
                          className="class-card-art"
                          style={{ backgroundImage: `url('/art/${o.id}_hero.png')` }}
                        />
                      )}
                      <span className="class-name">{o.label}</span>
                      {!enabled && <span className="class-soon">Coming soon</span>}
                    </button>
                  )
                })}
              </div>
              <button className="class-cancel" onClick={() => setPicking(false)}>
                Cancel
              </button>
            </div>
          </div>,
          document.body,
        )}

      {/* Mobile filters modal — mana / rarity / tribe pips moved off the cramped
         book chrome. Portaled to body to escape .builder-page's scale transform. */}
      {isMobile &&
        filtersOpen &&
        createPortal(
          <div className="overlay" onClick={() => setFiltersOpen(false)}>
            <div className="filters-modal" onClick={(e) => e.stopPropagation()}>
              <div className="fm-head">
                <h2>Filters</h2>
                <button className="mode-close" onClick={() => setFiltersOpen(false)} aria-label="Close">
                  ✕
                </button>
              </div>
              <div className="fm-group">
                <span className="fm-label">Mana</span>
                <div className="filter-row">{manaPips}</div>
              </div>
              <div className="fm-group">
                <span className="fm-label">Rarity</span>
                <div className="filter-row">{rarityPips}</div>
              </div>
              {tribes.length > 0 && (
                <div className="fm-group">
                  <span className="fm-label">Tribe</span>
                  <div className="tribe-bar">{tribePips}</div>
                </div>
              )}
              <div className="fm-actions">
                <button
                  className="fm-clear"
                  onClick={() => {
                    setManaFilter(null)
                    setRarityFilter(null)
                    setTribeFilter(null)
                  }}
                >
                  Clear all
                </button>
                <button className="fm-done" onClick={() => setFiltersOpen(false)}>
                  Done
                </button>
              </div>
            </div>
          </div>,
          document.body,
        )}

      {/* In-deck hover preview — portaled to body so position:fixed uses true
         viewport coords (the .builder-page transform:scale would otherwise become
         its containing block, scaling + offsetting it). Pinned left at cursor Y. */}
      {editing &&
        preview &&
        createPortal(
          // Stuck to the deck panel's left edge (panel left X minus card width +
          // gap), at cursor Y. Clamp Y by the scaled-up half-height (~1.5x the 236px
          // card) so it never clips off the top/bottom of the screen.
          <div
            className="dc-floating-preview"
            style={{
              left: preview.x - 164 - 12,
              top: Math.min(Math.max(preview.y, 180), window.innerHeight - 180),
            }}
          >
            <div className={'card' + cardColorClass(preview.card)}>
              <CardFace card={preview.card} />
            </div>
          </div>,
          document.body,
        )}
    </div>
  )
}
