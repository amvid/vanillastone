// Lightweight DOM animation layer. The server resolves each action instantly and
// sends the settled snapshot plus an ordered event log; we replay the visually
// important events (attacks, damage, heals) as transient WAAPI animations over
// the already-rendered board. Best-effort: if an element is gone (e.g. a minion
// that died in the trade) the animation is skipped. No game state lives here —
// the server is still authoritative.

const el = (cid: string) => document.querySelector<HTMLElement>(`[data-cid="${CSS.escape(cid)}"]`)

const center = (r: DOMRect) => ({ x: r.left + r.width / 2, y: r.top + r.height / 2 })

// lunge translates the attacker partway toward its target and back — the HS
// "charge in, snap back" hit. The target may be a live element (its data-cid) or
// a DOMRect, so a minion that just died in the trade can still be the endpoint.
export function lunge(sourceCid: string, target: string | DOMRect) {
  const s = el(sourceCid)
  if (!s) return
  const tr = typeof target === 'string' ? el(target)?.getBoundingClientRect() : target
  if (!tr) return
  const a = center(s.getBoundingClientRect())
  const b = center(tr)
  const dx = (b.x - a.x) * 0.55
  const dy = (b.y - a.y) * 0.55
  const prevZ = s.style.zIndex
  s.style.zIndex = '40'
  const anim = s.animate(
    [
      { transform: 'translate(0,0)' },
      { transform: `translate(${dx}px, ${dy}px)`, offset: 0.5 },
      { transform: 'translate(0,0)' },
    ],
    { duration: 900, easing: 'cubic-bezier(.5,0,.4,1)' },
  )
  anim.onfinish = () => {
    s.style.zIndex = prevZ
  }
}

// hitFlash shakes and reddens the struck character and floats the damage number.
export function hitFlash(targetCid: string, amount?: number) {
  const t = el(targetCid)
  if (!t) return
  t.animate(
    [
      { filter: 'brightness(1)', transform: 'translate(0,0)' },
      { filter: 'brightness(2.2) drop-shadow(0 0 8px #f55)', transform: 'translate(-3px,0)', offset: 0.2 },
      { transform: 'translate(3px,0)', offset: 0.5 },
      { transform: 'translate(-2px,0)', offset: 0.8 },
      { filter: 'brightness(1)', transform: 'translate(0,0)' },
    ],
    { duration: 650, easing: 'ease-out' },
  )
  if (amount && amount > 0) floatNumber(t, `-${amount}`, '#ff5555')
}

// projectile flings an icon (e.g. a hero power glyph) from a source element to a
// target character, growing on arrival then bursting. The hit flash/number is
// driven separately by the damage event.
export function projectile(fromSelector: string, target: string | DOMRect, icon: string) {
  const src = document.querySelector<HTMLElement>(fromSelector)
  if (!src) return
  const a = center(src.getBoundingClientRect())
  const tr = typeof target === 'string' ? el(target)?.getBoundingClientRect() : target
  if (!tr) return
  const b = center(tr)
  const dx = b.x - a.x
  const dy = b.y - a.y
  const node = document.createElement('div')
  node.className = 'projectile'
  node.textContent = icon
  Object.assign(node.style, {
    position: 'fixed',
    left: `${a.x}px`,
    top: `${a.y}px`,
    zIndex: '85',
    pointerEvents: 'none',
  })
  document.body.appendChild(node)
  const anim = node.animate(
    [
      { transform: 'translate(-50%,-50%) scale(.5)', opacity: 0 },
      { transform: 'translate(-50%,-50%) scale(1.1)', opacity: 1, offset: 0.15 },
      { transform: `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px)) scale(1.1)`, opacity: 1, offset: 0.8 },
      { transform: `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px)) scale(1.8)`, opacity: 0 },
    ],
    { duration: 700, easing: 'cubic-bezier(.4,0,.5,1)' },
  )
  anim.onfinish = () => node.remove()
}

// burnCard shows a drawn-but-discarded card (overdraw) igniting at a player's
// deck pile and burning away. The card identity is hidden, so it's a card back.
export function burnCard(side: 'self' | 'opp') {
  const ref =
    document.querySelector<HTMLElement>(`.deck-pile.${side} .deck-card-back`) ??
    document.querySelector<HTMLElement>(`.deck-pile.${side}`)
  if (!ref) return
  const r = ref.getBoundingClientRect()
  const node = document.createElement('div')
  node.className = 'card-back burn-card'
  Object.assign(node.style, {
    position: 'fixed',
    left: `${r.left}px`,
    top: `${r.top}px`,
    width: `${r.width || 52}px`,
    height: `${r.height || 74}px`,
    zIndex: '80',
    pointerEvents: 'none',
  })
  const flame = document.createElement('div')
  flame.className = 'burn-flame'
  flame.textContent = '🔥'
  node.appendChild(flame)
  document.body.appendChild(node)
  // Fly from the deck pile to screen center, then burn up there.
  const w = r.width || 52
  const h = r.height || 74
  const cx = window.innerWidth / 2 - (r.left + w / 2)
  const cy = window.innerHeight / 2 - (r.top + h / 2)
  const anim = node.animate(
    [
      { transform: 'translate(0,0) scale(1) rotate(0)', opacity: 1, filter: 'brightness(1)' },
      {
        transform: `translate(${cx}px, ${cy}px) scale(1.6) rotate(-4deg)`,
        opacity: 1,
        filter: 'brightness(1.1)',
        offset: 0.5,
      },
      {
        transform: `translate(${cx}px, ${cy}px) scale(1.7) rotate(3deg)`,
        opacity: 1,
        filter: 'brightness(1.5) sepia(.7) hue-rotate(-25deg)',
        offset: 0.75,
      },
      {
        transform: `translate(${cx}px, ${cy - 30}px) scale(1.2) rotate(10deg)`,
        opacity: 0,
        filter: 'brightness(2.4) sepia(1)',
      },
    ],
    { duration: 1500, easing: 'ease-in' },
  )
  anim.onfinish = () => node.remove()
}

// fatigueBurst flies a black skull card from a player's (empty) deck pile to
// screen center, holds, then fades — the deck punishing them. Hero flash follows.
export function fatigueBurst(side: 'self' | 'opp') {
  const ref =
    document.querySelector<HTMLElement>(`.deck-pile.${side} .deck-card-back`) ??
    document.querySelector<HTMLElement>(`.deck-pile.${side}`)
  if (!ref) return
  const r = ref.getBoundingClientRect()
  const w = r.width || 50
  const h = r.height || 68
  const node = document.createElement('div')
  node.className = 'fatigue-card'
  node.innerHTML = '<span class="fatigue-skull">💀</span><span class="fatigue-name">Fatigue</span>'
  Object.assign(node.style, {
    position: 'fixed',
    left: `${r.left}px`,
    top: `${r.top}px`,
    width: `${w}px`,
    height: `${h}px`,
    zIndex: '82',
    pointerEvents: 'none',
  })
  document.body.appendChild(node)
  const cx = window.innerWidth / 2 - (r.left + w / 2)
  const cy = window.innerHeight / 2 - (r.top + h / 2)
  const anim = node.animate(
    [
      { transform: 'translate(0,0) scale(1) rotate(-8deg)', opacity: 0 },
      { transform: `translate(${cx}px, ${cy}px) scale(2.2) rotate(0)`, opacity: 1, offset: 0.45 },
      { transform: `translate(${cx}px, ${cy}px) scale(2.2) rotate(0)`, opacity: 1, offset: 0.8 },
      { transform: `translate(${cx}px, ${cy - 24}px) scale(2.4) rotate(6deg)`, opacity: 0 },
    ],
    { duration: 1400, easing: 'ease-out' },
  )
  anim.onfinish = () => node.remove()
}

// fatiguePop flashes a hero dark and floats the fatigue damage with a skull.
export function fatiguePop(targetCid: string, amount?: number) {
  const t = el(targetCid)
  if (!t) return
  t.animate(
    [
      { filter: 'brightness(1)' },
      { filter: 'brightness(.45) sepia(1) hue-rotate(220deg)', offset: 0.3 },
      { filter: 'brightness(1)' },
    ],
    { duration: 650, easing: 'ease-out' },
  )
  if (amount && amount > 0) floatNumber(t, `💀 ${amount}`, '#b07be0')
}

// healPop floats a green number off a healed character.
export function healPop(targetCid: string, amount?: number) {
  const t = el(targetCid)
  if (t && amount && amount > 0) floatNumber(t, `+${amount}`, '#4ade80')
}

// Minions currently mid entrance (fly-in / summon-pop). While entering, a WAAPI
// transform (translate + scale) is on the node, so getBoundingClientRect returns
// a transformed box, NOT the layout rect. The caller must NOT cache that rect:
// settleShift would later read it as the survivor's pre-collapse position and
// slide the minion in from wherever the entrance transform happened to be.
const entering = new Set<string>()
export const isEntering = (cid: string) => entering.has(cid)

// flyIn animates a freshly-summoned minion sliding/scaling in from its
// controller's hand region (the played card "arriving" on the table). Source is
// a CSS selector for that hand area (.hand for us, .opp-hand for the opponent).
export function flyIn(minionCid: string, fromSelector: string, duration = 700) {
  const m = el(minionCid)
  if (m) slideFrom(m, fromSelector, duration, minionCid)
}

// summonPop animates a minion appearing in place — used for tokens summoned by
// a onset/finalGasp (Bogling, Hatchling, a Mimic copy), which don't come
// from a hand. It scales up from small with a slight overshoot at the same speed
// as a played minion's fly-in, so summons read as a distinct "pop onto the board".
export function summonPop(minionCid: string, duration = 700) {
  const m = el(minionCid)
  if (!m) return
  entering.add(minionCid)
  const anim = m.animate(
    [
      { transform: 'scale(.3) rotate(-8deg)', opacity: 0, filter: 'brightness(2)' },
      { transform: 'scale(1.12) rotate(2deg)', opacity: 1, filter: 'brightness(1.4)', offset: 0.65 },
      { transform: 'scale(1) rotate(0)', opacity: 1, filter: 'brightness(1)' },
    ],
    { duration, easing: 'cubic-bezier(.2,.85,.3,1.1)' },
  )
  anim.onfinish = () => entering.delete(minionCid)
}

// flyInEl is flyIn for a node we already hold (no data-cid) — e.g. an opponent's
// face-down hand card sliding in from their deck on a draw.
export function flyInEl(node: HTMLElement, fromSelector: string, duration = 700) {
  slideFrom(node, fromSelector, duration)
}

// dealIn flies a hand card in from `from` after `delay` ms — the opening "deal"
// once the mulligan ends. `from` is a CSS selector (e.g. the deck pile) or a
// fixed point (e.g. the mulligan modal's center, so the cards you just kept fly
// down into the hand). fill:backwards holds the card at the source (opacity 0)
// through the delay so a staggered deal never flashes a card at its final spot.
export function dealIn(cardCid: string, from: string | Point, delay: number, duration = 600) {
  const m = el(cardCid)
  if (m) slideFrom(m, from, duration, undefined, delay)
}

type Point = { x: number; y: number }

// slideFrom animates `node` in from the center of `from` (a CSS selector for a
// deck pile / hand area, or an explicit screen point), scaling and rotating up to
// its final spot. When `cid` is given the minion is flagged as entering for the
// animation's life, so the rect cache skips it (the live box is transformed
// mid-flight, not the layout position). `delay` staggers the start (held at the
// source via fill:backwards).
function slideFrom(node: HTMLElement, from: string | Point, duration: number, cid?: string, delay = 0) {
  const mr = center(node.getBoundingClientRect())
  let sr: Point
  if (typeof from === 'string') {
    const src = document.querySelector<HTMLElement>(from)
    sr = src ? center(src.getBoundingClientRect()) : { x: mr.x, y: mr.y - 200 }
  } else {
    sr = from
  }
  const dx = sr.x - mr.x
  const dy = sr.y - mr.y
  if (cid) entering.add(cid)
  const anim = node.animate(
    [
      { transform: `translate(${dx}px, ${dy}px) scale(.62) rotate(-5deg)`, opacity: 0, offset: 0 },
      // Reach full opacity early so the card is clearly visible for the whole
      // flight (not a faint blur that only resolves once it lands).
      { opacity: 1, offset: 0.22 },
      { transform: 'translate(0,0) scale(1) rotate(0)', opacity: 1, offset: 1 },
    ],
    { duration, delay, fill: 'backwards', easing: 'cubic-bezier(.3,.55,.35,1)' },
  )
  if (cid) anim.onfinish = () => entering.delete(cid)
}

// deathPuff replays a now-removed minion's death where it stood. The board
// already dropped it (server sends settled state), so we re-create a clone from
// its last-render HTML + rect, hold it through `delay` (while the attack/hit
// plays), then fade + shrink it out.
export function deathPuff(html: string, rect: DOMRect, delay = 500) {
  const wrap = document.createElement('div')
  wrap.innerHTML = html
  const node = wrap.firstElementChild as HTMLElement | null
  if (!node) return
  // Keep the stat badges: they're children of the clone, so they fade + shrink
  // with the whole minion and vanish exactly when the body does (the board's real
  // minion is already gone, so without them the stats would blink out early).
  node.querySelectorAll('.tooltip').forEach((n) => n.remove())
  Object.assign(node.style, {
    position: 'fixed',
    left: `${rect.left}px`,
    top: `${rect.top}px`,
    width: `${rect.width}px`,
    height: `${rect.height}px`,
    margin: '0',
    zIndex: '70',
    pointerEvents: 'none',
  })
  document.body.appendChild(node)
  const anim = node.animate(
    [
      { opacity: 1, transform: 'scale(1) rotate(0)', filter: 'brightness(1)' },
      { opacity: 0, transform: 'scale(.5) rotate(10deg)', filter: 'brightness(2)' },
    ],
    { duration: 550, delay, easing: 'ease-in', fill: 'backwards' },
  )
  anim.onfinish = () => node.remove()
}

// In-flight settle slides, keyed by minion id, so a measurement (or a fresh
// settle) can drop a held FLIP offset before reading the element's geometry.
const settleAnims = new Map<string, Animation>()

// clearSettle cancels any in-flight settle slide on a minion, snapping it back to
// its true layout position. The caller MUST call this before measuring a board
// element's rect, otherwise getBoundingClientRect returns the held (translated)
// position and corrupts the next FLIP.
export function clearSettle(cid: string) {
  const a = settleAnims.get(cid)
  if (a) a.cancel()
  settleAnims.delete(cid)
}

// settleShift slides a surviving minion from where it stood last render to its
// new spot — but only AFTER a beat. When a neighbour dies the server sends the
// settled (already-collapsed) board, so the row would otherwise snap shut the
// instant the corpse is still mid-death-puff. This holds each survivor at its old
// position (a FLIP) while the puff plays, then closes the gap once the corpse has
// faded. `hold` ≈ deathPuff's delay+fade so the shift lands as the body vanishes.
// `fromRect` must be the survivor's pre-collapse layout rect and the element must
// be at its true (untransformed) layout position when this is called.
export function settleShift(cid: string, fromRect: DOMRect, hold = 1100, slide = 350) {
  const m = el(cid)
  if (!m) return
  clearSettle(cid) // measure clean: drop any prior held offset first
  const now = m.getBoundingClientRect()
  const dx = fromRect.left - now.left
  const dy = fromRect.top - now.top
  if (Math.abs(dx) < 1 && Math.abs(dy) < 1) return // didn't move → nothing to settle
  const total = hold + slide
  const anim = m.animate(
    [
      { transform: `translate(${dx}px, ${dy}px)` },
      { transform: `translate(${dx}px, ${dy}px)`, offset: hold / total },
      { transform: 'translate(0, 0)' },
    ],
    { duration: total, easing: 'cubic-bezier(.4,0,.3,1)' },
  )
  settleAnims.set(cid, anim)
  anim.onfinish = () => {
    if (settleAnims.get(cid) === anim) settleAnims.delete(cid)
  }
}

// playGhost clones a card being played and flies the copy from the hand up to
// the center of the screen, growing then fading — the HS "card to the table"
// flourish. Used for our own spells/secrets/weapons (minions get flyIn).
export function playGhost(sourceCid: string) {
  const s = el(sourceCid)
  if (!s) return
  const r = s.getBoundingClientRect()
  const clone = s.cloneNode(true) as HTMLElement
  clone.classList.add('play-ghost')
  clone.querySelectorAll('.tooltip').forEach((n) => n.remove())
  flyToCenter(clone, r)
}

// flyToCenter positions `node` (fixed) at `from` and flies it to screen center,
// growing then fading out, then removes it. Shared by both card ghosts.
function flyToCenter(node: HTMLElement, from: DOMRect) {
  Object.assign(node.style, {
    position: 'fixed',
    left: `${from.left}px`,
    top: `${from.top}px`,
    width: `${from.width}px`,
    height: `${from.height}px`,
    margin: '0',
    zIndex: '90',
    pointerEvents: 'none',
  })
  document.body.appendChild(node)
  const cx = window.innerWidth / 2 - (from.left + from.width / 2)
  const cy = window.innerHeight / 2 - (from.top + from.height / 2)
  const anim = node.animate(
    [
      { transform: 'translate(0,0) scale(1)', opacity: 1 },
      { transform: `translate(${cx}px, ${cy}px) scale(1.45)`, opacity: 1, offset: 0.6 },
      { transform: `translate(${cx}px, ${cy - 30}px) scale(1.6)`, opacity: 0 },
    ],
    { duration: 1300, easing: 'ease-out' },
  )
  anim.onfinish = () => node.remove()
}

function floatNumber(target: HTMLElement, text: string, color: string) {
  const r = target.getBoundingClientRect()
  const pop = document.createElement('div')
  pop.className = 'dmg-pop'
  pop.textContent = text
  pop.style.left = `${r.left + r.width / 2}px`
  pop.style.top = `${r.top + r.height / 2}px`
  pop.style.color = color
  document.body.appendChild(pop)
  const anim = pop.animate(
    [
      { transform: 'translate(-50%,-50%) scale(.6)', opacity: 0 },
      { transform: 'translate(-50%,-90%) scale(1.3)', opacity: 1, offset: 0.3 },
      { transform: 'translate(-50%,-160%) scale(1)', opacity: 0 },
    ],
    { duration: 1400, easing: 'ease-out' },
  )
  anim.onfinish = () => pop.remove()
}
