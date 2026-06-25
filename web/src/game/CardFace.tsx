import { useEffect, useState, type ReactNode } from 'react'
import type { CardView } from '../protocol'
import { cardArtIcon, cardDesc, cardTypeLabel } from './format'
import { artUrl, useArtVersion } from './artVersion'

// CardArt is the dedicated art slot: a real image at /art/<cardId>.png when one
// exists, falling back to the placeholder type glyph until art is added. Keyed by
// cardId at the call site so a card swap remounts it (resetting the error state).
function CardArt({ card }: { card: CardView }) {
  const [failed, setFailed] = useState(false)
  // Bumped when dev art changes; cache-busts the src AND clears a stale failed
  // state so a card that just got its first art swaps from glyph to image live.
  const version = useArtVersion()
  useEffect(() => setFailed(false), [version])
  if (failed) return <span className="card-art-glyph">{cardArtIcon(card.cardType)}</span>
  const src = artUrl(card.cardId, version)
  // Paint the art as a background on a full-bleed layer rather than a sized <img>:
  // a <div> with background-size:cover fills the slot edge-to-edge with no element
  // to mis-size to its intrinsic aspect (which left side gaps). A hidden probe img
  // only detects a missing file and flips to the placeholder glyph.
  return (
    <>
      <div
        className="card-art-img"
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: `url(${src})`,
          backgroundSize: 'cover',
          // Slot is wider than the square art, so cover crops vertically. Anchor the
          // BOTTOM so the crop eats the empty-sky top strip (per the art framing rule),
          // never the subject/feet/tail at the bottom.
          backgroundPosition: 'center bottom',
        }}
      />
      <img src={src} alt="" style={{ display: 'none' }} onError={() => setFailed(true)} />
    </>
  )
}

// Mechanic keywords bolded in the rules box — longest first so "Aegis" /
// "Spell Damage" win over a bare "Shield"/"Damage". Matched case-sensitively (card
// text capitalizes keywords) to avoid bolding the same word used in prose.
const KEYWORDS = [
  'Aegis',
  'Spell Damage',
  'Onset',
  'Final Gasp',
  'Seek',
  'Poisonous',
  'Lifesteal',
  'Twinstrike',
  'Stealth',
  'Silence',
  'Enrage',
  'Frozen',
  'Freeze',
  'Charge',
  'Taunt',
  'Secret',
  'Immune',
  'Rush',
]
// One regex: a {sd:N} green-number marker, or any keyword on a word boundary.
const DESC_RE = new RegExp(`\\{sd:(\\d+)\\}|\\b(${KEYWORDS.join('|')})\\b`, 'g')

// renderDesc turns formatted rules text into nodes: the {sd:N} Spell Damage marker
// (server-injected when a card's damage is boosted in hand) becomes a green number,
// and mechanic keywords are bolded. Plain prose is left as-is.
function renderDesc(desc: string) {
  const nodes: ReactNode[] = []
  let last = 0
  let m: RegExpExecArray | null
  DESC_RE.lastIndex = 0
  while ((m = DESC_RE.exec(desc)) !== null) {
    if (m.index > last) nodes.push(desc.slice(last, m.index))
    if (m[1] !== undefined) {
      nodes.push(
        <span key={m.index} className="sd-buff">
          {m[1]}
        </span>,
      )
    } else {
      nodes.push(
        <strong key={m.index} className="kw">
          {m[2]}
        </strong>,
      )
    }
    last = m.index + m[0].length
  }
  if (nodes.length === 0) return desc
  if (last < desc.length) nodes.push(desc.slice(last))
  return nodes
}

// CardFace is the shared inner content of a card "face" (hand, drag clone, cast
// preview, seek option, mulligan, deckbuilder) so every surface looks the
// same. Layout (top→bottom): cost gem (top-left), name band, dominant ART slot,
// a rules-text box, and a type/tribe band along the bottom; attack/health corner
// circles for minions/weapons. The caller owns the wrapping .card element
// (button/div), its type/state classes, data-cid, click handlers, and tooltip.
export function CardFace({ card }: { card: CardView }) {
  const hasStats = card.cardType === 'minion' || card.cardType === 'weapon'
  const right = card.cardType === 'weapon' ? (card.durability ?? 0) : card.health
  // Colour the mana gem's DIGIT when a cost modifier (aura / cost-rule minion)
  // changed the printed cost: green if cheaper, red if pricier.
  const costCls =
    card.baseCost != null && card.cost !== card.baseCost
      ? card.cost < card.baseCost
        ? ' cost-reduced'
        : ' cost-raised'
      : ''
  // Title font auto-shrinks so the name fits the band cleanly. Two factors: a long
  // single word can't wrap (shrink by longest-word length), AND the cost gem eats
  // the top-left corner so a long *total* name crowds it even when each word is
  // short (e.g. "Arcane Wyrmling") — shrink by total length too. Only the longest
  // names step down; everything else keeps the crisp default size.
  const longestWord = Math.max(...card.name.split(/\s+/).map((w) => w.length))
  const nameLen = card.name.length
  const nameStyle =
    longestWord >= 13 || nameLen >= 20 ? { fontSize: '8px' } : longestWord >= 11 || nameLen >= 14 ? { fontSize: '9px' } : undefined
  // The text box is a FIXED region (every card the same size). To keep the longest
  // texts (e.g. Decoy Ward) from clipping, shrink the rules-text font as the text
  // gets longer — short texts stay crisp, long ones step down to fit the box.
  const desc = card.text ? cardDesc(card.text) : ''
  // Font tier is keyed off the *visible* length: strip the {sd:N} Spell Damage
  // markers (server-injected; rendered as a green number) so a buffed card doesn't
  // shrink its text just because the marker syntax is longer than the digit.
  const descLen = desc.replace(/\{sd:(\d+)\}/g, '$1').length
  const descStyle =
    descLen > 130
      ? { fontSize: '7px' }
      : descLen > 95
        ? { fontSize: '8px' }
        : descLen > 60
          ? { fontSize: '9px' }
          : undefined
  return (
    <>
      {/* Art fills the top region; the cost gem and name plate sit OVER it. */}
      <div className="card-art">
        <CardArt key={card.cardId} card={card} />
        {/* Nameplate ribbon straddles the art's bottom edge (not the top, where it
            covered the art's subject and wrapped badly). Solid plate → 2-row names read clean. */}
        <div className={'name' + (card.rarity ? ` rarity-${card.rarity}` : '')} style={nameStyle}>
          {card.name}
        </div>
      </div>
      <div className={'cost' + costCls}>{card.cost}</div>
      <div className={'card-textbox' + (desc ? '' : ' empty')} style={descStyle}>
        {/* One wrapper so the textbox's flex sees a SINGLE item — without it the
            {sd:N} split turns each text run into its own flex item (scrambles order
            + drops the space before the green number). */}
        <span className="card-desc-body">{renderDesc(desc)}</span>
      </div>
      <div className="type-band">{cardTypeLabel(card)}</div>
      {hasStats && <div className="stat atk">{card.attack}</div>}
      {hasStats && <div className="stat hp">{right}</div>}
    </>
  )
}
