import { useEffect, useMemo, useState } from 'react'
import { createPortal } from 'react-dom'
import { createDeck, deleteDeck, fetchPool, listDecks, updateDeck } from './api'
import type { Deck, Pool } from './api'
import type { CardView } from './protocol'
import { CardFace } from './game/CardFace'
import { cardColorClass } from './game/format'

const PER_PAGE = 8 // cards shown per book page (4 columns x 2 rows)

// All hero classes shown in the new-deck picker. Only those the server reports as
// playable are selectable; the rest render as "coming soon" so the roadmap is
// visible without being usable.
const CLASS_OPTIONS: { id: string; label: string }[] = [
  { id: 'mage', label: 'Mage' },
  { id: 'hunter', label: 'Hunter' },
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
  // Card previewed on the left while hovering an in-deck row, pinned to the
  // cursor's vertical level. Held in state (not pure CSS) because the in-deck list
  // scrolls (overflow:auto), which would clip an absolutely-positioned preview,
  // and because it follows the mouse Y.
  const [preview, setPreview] = useState<{ card: CardView; x: number; y: number } | null>(null)

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
      return true
    })
    cs.sort((a, b) => a.cost - b.cost || a.name.localeCompare(b.name))
    return cs
  }, [pool, tab, activeClass, manaFilter, rarityFilter, editing])

  // Any tab/filter/deck change can shrink the list — snap back to the first page.
  useEffect(() => setPage(0), [tab, manaFilter, rarityFilter, activeClass])

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
  const pageCount = Math.max(1, Math.ceil(filtered.length / PER_PAGE))
  const pageCards = filtered.slice(page * PER_PAGE, page * PER_PAGE + PER_PAGE)

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

  const remove = async (id: number) => {
    setError('')
    try {
      await deleteDeck(token, id)
      await reloadDecks()
      if (editing?.id === id) setEditing(null)
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

  return (
    <div className="builder-page">
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
          {/* Tabs + filters */}
          <div className="collection-controls">
            <div className="col-tabs">
              <button className={tab === 'all' ? 'on' : ''} onClick={() => setTab('all')}>
                All
              </button>
              {(editing ? [editing.class] : pool.classes).map((cls) => (
                <button
                  key={cls}
                  className={tab === 'class' && activeClass === cls ? 'on' : ''}
                  onClick={() => {
                    setBrowseClass(cls)
                    setTab('class')
                  }}
                >
                  {classLabel(cls)}
                </button>
              ))}
              <button className={tab === 'neutral' ? 'on' : ''} onClick={() => setTab('neutral')}>
                Neutral
              </button>
              <span className="col-count">{filtered.length} cards</span>
            </div>
            <div className="col-filters">
              <div className="filter-row">
                {MANA_BUCKETS.map((m) => (
                  <button
                    key={m}
                    className={'mana-pip' + (manaFilter === m ? ' on' : '')}
                    onClick={() => setManaFilter((cur) => (cur === m ? null : m))}
                  >
                    {m === 7 ? '7+' : m}
                  </button>
                ))}
              </div>
              <div className="filter-row">
                {RARITIES.map((r) => (
                  <button
                    key={r}
                    className={'rarity-pip ' + r + (rarityFilter === r ? ' on' : '')}
                    onClick={() => setRarityFilter((cur) => (cur === r ? null : r))}
                  >
                    {r[0].toUpperCase() + r.slice(1)}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div className="book">
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
              {Array.from({ length: PER_PAGE - pageCards.length }, (_, i) => (
                <div key={`pad-${i}`} className="book-card placeholder" aria-hidden="true" />
              ))}
              {pageCards.length === 0 && <p className="empty-collection">No cards match these filters.</p>}
            </div>
          </div>
          <div className="pager">
            <button disabled={page === 0} onClick={() => setPage((p) => Math.max(0, p - 1))}>
              ‹ Prev
            </button>
            <span className="page-no">
              Page {page + 1} / {pageCount}
            </span>
            <button disabled={page >= pageCount - 1} onClick={() => setPage((p) => Math.min(pageCount - 1, p + 1))}>
              Next ›
            </button>
          </div>
        </main>

        {/* Deck panel */}
        <aside className="deck-panel">
          <h3>
            Your decks ({decks.length}/{maxDecks})
          </h3>
          <div className="deck-list">
            {decks.map((d) => (
              <div key={d.id} className="deck-row">
                <button onClick={() => setEditing({ id: d.id, name: d.name, class: d.class, cards: d.cards })}>
                  <span
                    className="deck-row-art"
                    style={{ backgroundImage: `url('/art/${d.class}_hero.png')` }}
                  />
                  <span className="deck-row-text">{d.name}</span>
                </button>
                <button className="del" onClick={() => remove(d.id)} title="Delete">
                  ✕
                </button>
              </div>
            ))}
            <button className="new-deck" disabled={decks.length >= maxDecks} onClick={() => setPicking(true)}>
              + New deck
            </button>
          </div>

          {editing && (
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
          )}
          {!editing && <p className="hint">Pick a deck to edit, or create a new one.</p>}
        </aside>
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
