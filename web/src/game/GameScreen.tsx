import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import type { CardView, Event, MinionView, OppIntent, Snapshot } from '../protocol'
import type { CharKind, LogActor, LogEntry, PendingSpell } from './types'
import { cardColorClass, hpIcon, kindIcon } from './format'
import { CardFace } from './CardFace'
import { Board } from './Board'
import { Hero } from './Hero'
import { CardTooltip } from './Tooltip'
import { TargetingArrow } from './TargetingArrow'
import {
  dealIn,
  deathPuff,
  fatigueBurst,
  fatiguePop,
  flyIn,
  flyInEl,
  healPop,
  clearSettle,
  hitFlash,
  isEntering,
  lunge,
  projectile,
  settleShift,
  summonPop,
} from './animate'

export type GameScreenProps = {
  snap: Snapshot
  name: string
  myTurn: boolean
  turnSecs: number
  turnNum: number
  winner: string | null
  // Ladder rank change for a finished ranked game (null when unranked/AI), shown
  // on the win/loss screen.
  rankUpdate?: { oldRank: number; newRank: number } | null
  attacker: string | null
  spell: PendingSpell | null
  heroPowerArmed: boolean
  seek: CardView[] | null
  oppSeek: number | null
  oppIntent?: OppIntent | null
  log: LogEntry[]
  anim: { seq: number; events: Event[] } | null
  oppOnline: boolean
  status: string
  hint: string
  send: (msg: object) => void
  onBackToLobby: () => void
  onChar: (targetId: string, kind: CharKind, m?: MinionView) => void
  targetable: (kind: CharKind, m?: MinionView) => boolean
  onHandCard: (i: number, card: CardView, pos?: number) => void
  onHeroPower: () => void
  // True for ~one beat when play begins from the mulligan: deal the opening hand
  // in (then the turn-1 draw from the deck, if it's our turn).
  intro?: boolean
  // Screen point the opening deal flies FROM — the mulligan modal's center, so
  // the cards you just kept drop into the hand. Falls back to the deck pile.
  introFrom?: { x: number; y: number } | null
  // When set, this is a read-only spectator view of that player's match: all
  // controls are hidden/disabled and a "spectating" banner + Leave button show.
  spectating?: string | null
  // Usernames currently watching this match (shown to players as a badge).
  spectators?: string[]
}

export function GameScreen(props: GameScreenProps) {
  const {
    snap,
    name,
    myTurn,
    turnSecs,
    turnNum,
    winner,
    rankUpdate,
    attacker,
    spell,
    heroPowerArmed,
    seek,
    oppSeek,
    oppIntent,
    log,
    anim,
    oppOnline,
    hint,
    send,
    onBackToLobby,
    onChar,
    targetable,
    onHandCard,
    onHeroPower,
    intro,
    introFrom,
    spectating,
    spectators,
  } = props
  const spectator = !!spectating
  const watchers = spectators ?? []

  const hp = snap.self.heroPower

  // The DOM id of whatever is currently aiming: a selected attacker (minion uid
  // or selfHero), an armed spell (its hand card), or an armed hero power. Drives
  // the targeting arrow.
  const sourceId = attacker ?? (spell ? `hand-${spell.handIndex}` : heroPowerArmed ? 'heroPower' : null)
  const [pointer, setPointer] = useState<{ x: number; y: number } | null>(null)
  // Hovered log row → a body-portaled detail popup, positioned at the row's Y and
  // clamped to the viewport (so it tracks the mouse but never goes off-screen).
  const [logPop, setLogPop] = useState<{ entry: LogEntry; rightPx: number; topPx: number } | null>(null)
  const [confirmConcede, setConfirmConcede] = useState(false)
  // The game-over modal appears 2s after the result lands, so the lethal action's
  // animation can play out first instead of being cut off by the overlay.
  const [winnerShown, setWinnerShown] = useState(false)
  useEffect(() => {
    if (!winner) {
      setWinnerShown(false)
      return
    }
    const id = window.setTimeout(() => setWinnerShown(true), 2000)
    return () => window.clearTimeout(id)
  }, [winner])
  const [lowTime, setLowTime] = useState(false) // ≤15s left on your turn → flash action outlines red
  // A card being dragged out of hand. hold-drag (mouse down + move + release) and
  // click-drag-click (click to lift → sticky, click again to drop) share this.
  // pos is the friendly-board slot for a minion (null = not over the table).
  const [drag, setDrag] = useState<{
    handIndex: number
    card: CardView
    sticky: boolean
    x: number
    y: number
    pos: number | null
  } | null>(null)
  const dragRef = useRef(drag)
  dragRef.current = drag
  // A press that hasn't yet decided hold-drag vs click; promoted on move/up.
  const pressRef = useRef<{ handIndex: number; card: CardView; x: number; y: number } | null>(null)
  // Friendly board minion centers captured at drag start so the insertion slot is
  // computed against a stable layout (the live row shifts to open the gap).
  const centersRef = useRef<number[]>([])
  // Set while aiming a targeted spell grabbed from hand (the card lifts toward the
  // table and the aim line follows the cursor); a pointerup over a legal target
  // casts it, otherwise it stays armed for a follow-up click on the target.
  const aimRef = useRef<{ handIndex: number } | null>(null)
  // Stable mirrors so the window pointer listeners always see the latest handlers
  // / snapshot without rebinding (they capture once via the [onHandCard] effect).
  const onCharRef = useRef(onChar)
  onCharRef.current = onChar
  const snapRef = useRef(snap)
  snapRef.current = snap
  // The opponent's most recent cast, previewed on the left for a few seconds
  // (swapped if they play again). `key` forces a fresh entrance per play. A real
  // card for minion/spell/weapon; `secret` for a hidden secret (never revealed).
  const [cast, setCast] = useState<{
    key: number
    card?: CardView
    counteredCard?: CardView // the spell a Counter-type secret negated (shown paired)
    secret?: boolean
    fatigue?: boolean
    triggered?: boolean
    label?: string
  } | null>(null)
  const castTimer = useRef<number | undefined>(undefined)
  const castSeq = useRef(0)
  // Center-screen burst when a secret fires, so a silent interrupt (kill-the-attacker,
  // counter) reads as an event, not "the minion/spell just vanished".
  const [secretFx, setSecretFx] = useState<{ key: number; name: string } | null>(null)
  const secretFxTimer = useRef<number | undefined>(undefined)
  const secretFxSeq = useRef(0)
  const fireSecretFx = useCallback((name: string) => {
    secretFxSeq.current++
    setSecretFx({ key: secretFxSeq.current, name })
    window.clearTimeout(secretFxTimer.current)
    secretFxTimer.current = window.setTimeout(() => setSecretFx(null), 3200)
  }, [])
  // A burnt card's face flies from its owner's deck pile to center, growing small->full
  // with a flame, so both players see what was destroyed, then fades. bx/by = the deck
  // pile's offset from screen center (the fly-from point). Last burn wins (batch overwrites).
  const [burnFace, setBurnFace] = useState<{
    key: number
    card: CardView
    bx: number
    by: number
  } | null>(null)
  const burnFaceTimer = useRef<number | undefined>(undefined)
  const burnFaceSeq = useRef(0)
  const showBurnFace = useCallback((card: CardView, side: 'self' | 'opp') => {
    const ref =
      document.querySelector<HTMLElement>(`.deck-pile.${side} .deck-card-back`) ??
      document.querySelector<HTMLElement>(`.deck-pile.${side}`)
    let bx = 0
    let by = 0
    if (ref) {
      const r = ref.getBoundingClientRect()
      bx = r.left + r.width / 2 - window.innerWidth / 2
      by = r.top + r.height / 2 - window.innerHeight / 2
    }
    burnFaceSeq.current++
    setBurnFace({ key: burnFaceSeq.current, card, bx, by })
    window.clearTimeout(burnFaceTimer.current)
    burnFaceTimer.current = window.setTimeout(() => setBurnFace(null), 2000)
  }, [])
  // Opponent's secret count last render: a rise means they played a secret.
  const prevOppSecrets = useRef<number | null>(null)
  const showReveal = useCallback(
    (payload: {
      card?: CardView
      counteredCard?: CardView
      secret?: boolean
      fatigue?: boolean
      triggered?: boolean
      label?: string
    }) => {
      castSeq.current++
      setCast({ key: castSeq.current, ...payload })
      window.clearTimeout(castTimer.current)
      castTimer.current = window.setTimeout(() => setCast(null), 4000)
    },
    [],
  )
  // --- #2 opponent-intent emit (the acting player streams what they're hovering /
  // aiming at; the server relays it to the opponent for HS-style "what are they about
  // to do" feedback). Ephemeral, non-authoritative. Mirrors so the throttled emitter
  // always reads the latest without re-binding. ---
  const sourceIdRef = useRef<string | null>(sourceId)
  sourceIdRef.current = sourceId
  const pointerRef = useRef<{ x: number; y: number } | null>(pointer)
  pointerRef.current = pointer
  const hoverHandRef = useRef(-1) // hand card the player is hovering (not yet aiming)
  const hoverBoardRef = useRef('') // board minion the player is inspecting
  const lastIntentRef = useRef('') // last payload sent, to drop duplicates
  const intentSentAt = useRef(0)
  // charUnder resolves the character at a screen point to its cid (for aim targets).
  const charUnder = useCallback((x: number, y: number): string | null => {
    const node = document.elementFromPoint(x, y)?.closest('[data-cid]')
    const cid = node?.getAttribute('data-cid')
    if (!cid) return null
    if (cid === 'oppHero' || cid === 'selfHero') return cid
    const s = snapRef.current
    if (s.opp.board.some((mn) => mn.instanceId === cid) || s.self.board.some((mn) => mn.instanceId === cid)) return cid
    return null
  }, [])
  // intentToken maps a local cid to a perspective-free wire token: heroes become
  // self/enemy roles (flipped by the receiver), a hand card becomes hand:<i>, a minion
  // keeps its (match-global) uid.
  const intentToken = (cid: string | null | undefined): string => {
    if (!cid) return ''
    if (cid === 'oppHero') return 'enemy' // my opponent's hero = the watcher
    if (cid === 'selfHero') return 'self'
    if (cid === 'heroPower') return 'heroPower'
    if (cid.startsWith('hand-')) return 'hand:' + cid.slice(5)
    return cid
  }
  // emitIntent derives the player's whole current aim from live refs and sends it
  // (deduped + lightly throttled, but a CLEAR always goes through immediately).
  const emitIntent = useCallback(() => {
    if (spectator) return
    let hoverHand = -1
    let hover = ''
    let aimFrom = ''
    let aimTo = ''
    // Emit regardless of whose turn it is — a WAITING player hovering / inspecting should
    // still telegraph to the active opponent (HS-style "picking up a card"). Aim/attack
    // (aimFrom) can only arise on your own turn anyway (the UI gates arming/attacking).
    if (!winner) {
      hoverHand = hoverHandRef.current
      hover = hoverBoardRef.current
      const d = dragRef.current
      const src = sourceIdRef.current
      if (d) {
        aimFrom = 'hand:' + d.handIndex
        hoverHand = d.handIndex
        const tgt = charUnder(d.x, d.y)
        if (tgt) aimTo = intentToken(tgt)
      } else if (src) {
        aimFrom = intentToken(src)
        if (src.startsWith('hand-')) hoverHand = +src.slice(5)
        const p = pointerRef.current
        const tgt = p && charUnder(p.x, p.y)
        if (tgt) aimTo = intentToken(tgt)
      }
    }
    const msg = { type: 'intent', hoverHand, hover, aimFrom, aimTo }
    const key = JSON.stringify(msg)
    if (key === lastIntentRef.current) return
    const now = performance.now()
    // Throttle ONLY the high-frequency aim motion (drag / targeting cursor). A hand-hover
    // or board-inspect change is discrete — send it immediately or a quick focus-switch
    // (card A → card B) gets swallowed and the opponent stays stuck on the old card.
    if (aimFrom && now - intentSentAt.current < 45) return
    lastIntentRef.current = key
    intentSentAt.current = now
    send(msg)
  }, [send, spectator, winner, charUnder])

  // Board minion ids from the previous render, to detect freshly-summoned minions
  // and fly them in. Null on the first render so an opening board doesn't animate.
  const prevCids = useRef<Set<string> | null>(null)
  // Opponent hand count from the previous render: a drop with no new enemy minion
  // means they played a non-minion card (spell/secret/weapon) → fly a card back; a
  // rise means they drew → fly a card back in from their deck.
  const prevOppHand = useRef<number | null>(null)
  // Our own hand size last render: a rise means we drew → fly the new card in.
  const prevSelfHand = useRef<number | null>(null)
  // Last-known screen rect of every character by data-cid, so an attack whose
  // target died this render can still lunge to where the target stood.
  const rectCache = useRef<Record<string, DOMRect>>({})
  // Last-render rect + HTML of each minion, to replay a death where it stood
  // (the board already removed it from the settled snapshot).
  const minionCache = useRef<Record<string, { rect: DOMRect; html: string }>>({})
  // Displayed-health overrides by cid (minion instanceId or selfHero/oppHero): a
  // damage target holds its PRE-hit health here until the hit animation connects,
  // then the entry is dropped so the snapshot value shows. Board/Hero read it.
  const [heldHp, setHeldHp] = useState<Record<string, number>>({})
  // Health every character showed at the end of the last animated action, so the
  // next action knows the "before" value to hold while its hit plays.
  const prevHpRef = useRef<Record<string, number>>({})
  // Latest action's events, mirrored to a ref so the [snap] layout effect (which
  // animates newly-summoned minions) can read them without re-running on `anim`.
  const animRef = useRef(anim)
  animRef.current = anim

  // Clear the arrow's anchor point whenever aiming stops, so a stale line never
  // flashes when a new target is picked. Re-emit intent so the opponent's view of our
  // aim updates when we arm/disarm a spell, attacker, or hero power.
  useEffect(() => {
    if (!sourceId) setPointer(null)
    emitIntent()
  }, [sourceId, emitIntent])

  // Turn changed (we ended ours, or theirs began): clear any lingering aim hint we
  // were broadcasting so the opponent doesn't see a stale arrow on their turn.
  useEffect(() => {
    emitIntent()
  }, [myTurn, emitIntent])

  // Inspecting a board minion (mouse over it, not aiming): tell the opponent which one
  // we're reading, so they get HS-style "they're eyeing this minion" feedback. A
  // delegated mouseover keeps Board.tsx untouched; moving onto empty space clears it.
  useEffect(() => {
    const onOver = (e: MouseEvent) => {
      if (spectator) return // inspect telegraphs on either turn (planning/threat-reading)
      const node = (e.target as HTMLElement)?.closest?.('[data-cid]') as HTMLElement | null
      const cid = node?.getAttribute('data-cid') ?? ''
      const s = snapRef.current
      const isMinion =
        cid[0] === 'u' && (s.opp.board.some((m) => m.instanceId === cid) || s.self.board.some((m) => m.instanceId === cid))
      const next = isMinion ? cid : ''
      if (next !== hoverBoardRef.current) {
        hoverBoardRef.current = next
        emitIntent()
      }
    }
    window.addEventListener('mouseover', onOver)
    return () => window.removeEventListener('mouseover', onOver)
  }, [emitIntent, spectator])

  // --- #2 opponent-intent RENDER (we are the viewer; the acting player is our
  // opponent). Shown only on their turn. Wire tokens are flipped to our POV:
  // sender 'enemy' = us = selfHero, sender 'self' = them = oppHero. ---
  // Show the opponent's hint whether or not it's our turn: on their turn it's their aim,
  // on our turn it's the waiting opponent telegraphing which card they're eyeing.
  const showOppIntent = !!oppIntent
  // Outline the minion(s)/hero the opponent is inspecting or aiming at. A class toggle
  // (vs. an overlay) keeps it correct as the board re-renders mid-turn; re-applied on
  // every relayed intent. Hand-card tokens have no data-cid (shown via the hand lift).
  useEffect(() => {
    if (!showOppIntent || !oppIntent) return
    const marks: { el: HTMLElement; cls: string }[] = []
    const add = (token: string | undefined, cls: string) => {
      if (!token || token.startsWith('hand:')) return
      const cid = oppTokenCid(token)
      const el = document.querySelector<HTMLElement>(`[data-cid="${CSS.escape(cid)}"]`)
      if (el) {
        el.classList.add(cls)
        marks.push({ el, cls })
      }
    }
    add(oppIntent.hover, 'opp-inspect')
    add(oppIntent.aimTo, 'opp-aim-target')
    return () => marks.forEach(({ el, cls }) => el.classList.remove(cls))
  }, [showOppIntent, oppIntent])
  // The opponent hand card to lift: an explicit hover, or the card they're aiming.
  const oppLift = !showOppIntent
    ? -1
    : oppIntent!.hoverHand >= 0
      ? oppIntent!.hoverHand
      : oppIntent!.aimFrom?.startsWith('hand:')
        ? Number(oppIntent!.aimFrom.slice(5))
        : -1

  // Clear the cast-showcase + secret-burst timers on unmount (leaving the match).
  useEffect(
    () => () => {
      window.clearTimeout(castTimer.current)
      window.clearTimeout(secretFxTimer.current)
      window.clearTimeout(burnFaceTimer.current)
    },
    [],
  )

  // When our turn begins, drop any lingering reveal (opponent-played / secret-triggered
  // showcase) from the opponent's turn — by then it's stale and shouldn't bleed into our
  // turn. A secret WE trigger mid-turn still shows (myTurn is already true, no edge here).
  useEffect(() => {
    if (myTurn) {
      window.clearTimeout(castTimer.current)
      setCast(null)
    }
  }, [myTurn])

  // A secret play emits no 'play' event (it stays hidden), but the opponent's
  // secret count is public — a rise means they set a secret. Preview it generically.
  useEffect(() => {
    const n = snap.opp.secretCount ?? 0
    const p = prevOppSecrets.current
    if (p !== null && n > p) showReveal({ secret: true })
    prevOppSecrets.current = n
  }, [snap, showReveal])

  // Hold each damage target's health number at its PRE-hit value until the hit
  // animation connects (the [anim] effect drops the hold then). Runs pre-paint so
  // the new (lower) number never flashes first. Only drops are held; heals/buffs
  // and unknown "before" values pass through unchanged.
  useLayoutEffect(() => {
    if (!anim) return
    const cidFor = (id?: string): string | null =>
      !id ? null : id[0] === 'u' ? id : id === snap.you ? 'selfHero' : 'oppHero'
    const newHp: Record<string, number> = {}
    for (const m of snap.self.board) newHp[m.instanceId] = m.health
    for (const m of snap.opp.board) newHp[m.instanceId] = m.health
    newHp.selfHero = snap.self.heroHP
    newHp.oppHero = snap.opp.heroHP
    const old = prevHpRef.current
    const holds: Record<string, number> = {}
    for (const e of anim.events)
      if (e.kind === 'damage' && e.target) {
        const c = cidFor(e.target)
        if (c && old[c] !== undefined && newHp[c] !== undefined && newHp[c] < old[c]) holds[c] = old[c]
      }
    prevHpRef.current = newHp
    // Replace (not merge) so any leftover hold from a prior action whose release
    // was canceled on re-run is cleared. Keep the empty ref stable to avoid a
    // needless re-render when there's nothing held and nothing to clear.
    setHeldHp((h) => (Object.keys(holds).length || Object.keys(h).length ? holds : h))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [anim])

  // When play begins from the mulligan, deal the opening hand in from the deck,
  // staggered left-to-right. If it's our turn the last card is the turn-1 draw —
  // fly it a beat later so it reads as a fresh draw, not part of the opening deal.
  useLayoutEffect(() => {
    if (!intro || spectator) return
    const hand = snap.self.hand ?? []
    const n = hand.length
    if (!n) return
    const STAGGER = 130
    const drawLast = !!myTurn // first player draws on turn 1; that card is hand[n-1]
    const openCount = drawLast ? n - 1 : n
    // Opening cards fly from the mulligan modal (where you just kept them); the
    // turn-1 draw still flies from the deck, a beat later, as a real draw. If the
    // modal rect wasn't captured, fall back to mid-screen (still a center → hand
    // fly, not a deck-corner pop).
    const from = introFrom ?? { x: window.innerWidth / 2, y: window.innerHeight * 0.42 }
    for (let i = 0; i < openCount; i++) dealIn(`hand-${i}`, from, i * STAGGER, 700)
    if (drawLast) dealIn(`hand-${n - 1}`, '.deck-pile.self', openCount * STAGGER + 420, 850)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [intro])

  // Fly in any minion that wasn't on the board last render (a play/summon),
  // from its controller's hand region. Layout effect: measure + start the WAAPI
  // before paint so the minion never flashes at its final spot first.
  useLayoutEffect(() => {
    const cur = new Set<string>()
    for (const mn of snap.self.board) cur.add(mn.instanceId)
    for (const mn of snap.opp.board) cur.add(mn.instanceId)
    const prev = prevCids.current
    // Survivors to FLIP after a death, mapped to their pre-collapse layout rect.
    // Filled below, applied only once the cache is refreshed to the new layout.
    const settleFrom: Record<string, DOMRect> = {}
    if (prev) {
      // Tokens summoned by a onset/finalGasp (Bogling, Hatchling, a Mimic
      // copy) "pop" in place; the minion actually played from hand flies in from
      // the hand. Distinguish them via this action's events: every summon emits a
      // `summon` event, but only a hand-played minion also has a `play` event
      // naming it — that one (and only the first such match) flies in.
      const ev = animRef.current?.events ?? []
      const summonIds = new Set<string>()
      for (const e of ev) if (e.kind === 'summon' && e.target) summonIds.add(e.target)
      const playEv = ev.find((e) => e.kind === 'play')
      const playedName = playEv?.card?.name ?? playEv?.name
      // A minion played from hand flies in; any token it then summons (e.g. an
      // Onset/battlecry) should pop only AFTER it lands, so the cause reads before
      // the effect. Tokens wait FLY_LAND for the played minion, then stagger so
      // multiple tokens pop one-by-one rather than all at once.
      const FLY_LAND = 650 // ≈ flyIn duration — the played minion has settled
      const TOKEN_GAP = 320 // between successive token pops
      const newMinions = [...snap.self.board, ...snap.opp.board].filter((mn) => !prev.has(mn.instanceId))
      const hasPlayedMinion = newMinions.some(
        (mn) => playedName && mn.name === playedName && summonIds.has(mn.instanceId),
      )
      let flewPlayed = false
      let tokenIdx = 0
      const animateNew = (mn: MinionView, from: string) => {
        if (!flewPlayed && playedName && mn.name === playedName && summonIds.has(mn.instanceId)) {
          flewPlayed = true
          flyIn(mn.instanceId, from)
        } else if (summonIds.has(mn.instanceId)) {
          const delay = (hasPlayedMinion ? FLY_LAND : 0) + tokenIdx * TOKEN_GAP
          tokenIdx++
          summonPop(mn.instanceId, 700, delay)
        } else {
          flyIn(mn.instanceId, from)
        }
      }
      for (const mn of snap.self.board) if (!prev.has(mn.instanceId)) animateNew(mn, '.hand')
      for (const mn of snap.opp.board) if (!prev.has(mn.instanceId)) animateNew(mn, '.opp-hand')
      // Minions that were on the board last render but are gone now died: replay
      // the death where they stood, held briefly so the attack/hit plays first.
      let died = false
      for (const cid of prev)
        if (!cur.has(cid)) {
          died = true
          const c = minionCache.current[cid]
          if (c) deathPuff(c.html, c.rect)
        }
      // A death collapses the settled board immediately. To hold survivors at
      // their old spots and slide them in once the corpse fades, snapshot their
      // pre-collapse rects from the still-old cache NOW; the actual FLIP runs
      // after the cache is refreshed to the new layout (below), so settleShift
      // measures the true collapsed position rather than a held transform.
      if (died)
        for (const mn of [...snap.self.board, ...snap.opp.board])
          if (prev.has(mn.instanceId)) {
            const r = rectCache.current[mn.instanceId]
            if (r) settleFrom[mn.instanceId] = r
          }
    }
    // Opponent's hand grew: they drew a card from their deck — fly a card back in.
    // (A shrink is a play, now revealed via the 'play' event / cast showcase.)
    // (Suppressed during the opening deal: the mulligan→play hand jump is the
    // deal, not a single draw — see the intro effect above.)
    const ph = prevOppHand.current
    if (!intro && ph !== null && snap.opp.handCount > ph) {
      const backs = document.querySelectorAll<HTMLElement>('.opp-hand .card-back')
      const last = backs[backs.length - 1]
      if (last) flyInEl(last, '.deck-pile.opp', 1100)
    }
    // Our hand grew (a draw): fly the newest card in from our deck pile.
    const sh = snap.self.hand?.length ?? 0
    const psh = prevSelfHand.current
    if (!intro && psh !== null && sh > psh) flyIn(`hand-${sh - 1}`, '.deck-pile.self', 1100)

    // Refresh the caches from the live DOM (merge: a character that left the board
    // keeps its last-known rect/HTML for one more action's animations).
    const cache = rectCache.current
    const mcache = minionCache.current
    document.querySelectorAll<HTMLElement>('[data-cid]').forEach((node) => {
      const id = node.getAttribute('data-cid')
      if (!id) return
      clearSettle(id) // drop any held FLIP offset so we measure the true layout
      // A minion mid fly-in/summon carries an entrance transform, so its live box
      // is wherever the animation is, not its layout spot. Skip it: caching that
      // would later make settleShift slide it in from the hand on the next death.
      if (isEntering(id)) return
      const rect = node.getBoundingClientRect()
      cache[id] = rect
      if (id[0] === 'u') mcache[id] = { rect, html: node.outerHTML }
    })

    // Now the cache holds the new (collapsed) layout: hold each survivor at its
    // old spot and slide it into the gap once the corpse's death puff has faded.
    for (const id in settleFrom) settleShift(id, settleFrom[id])

    prevCids.current = cur
    prevOppHand.current = snap.opp.handCount
    prevSelfHand.current = sh
  }, [snap])

  // Replay this action's events as animations over the settled board. Event ids
  // are absolute: minion uids start with "u"; a hero id is the player's id, which
  // we map to selfHero/oppHero relative to us (snap.you).
  useEffect(() => {
    if (!anim) return
    const events = anim.events
    const cidFor = (id?: string): string | null =>
      !id ? null : id[0] === 'u' ? id : id === snap.you ? 'selfHero' : 'oppHero'
    // Target as a live element id (preferred) or its last-known rect if it died.
    const dstOf = (cid: string): string | DOMRect | null =>
      document.querySelector(`[data-cid="${CSS.escape(cid)}"]`) ? cid : (rectCache.current[cid] ?? null)
    // Drop a held health number so it falls back to the snapshot value — called
    // when that target's hit connects, so the number drops in sync with the flash.
    const releaseHp = (cid: string) =>
      setHeldHp((h) => {
        if (!(cid in h)) return h
        const n = { ...h }
        delete n[cid]
        return n
      })
    // A damage event that is the impact of an attack or hero power flashes when that
    // cause's animation connects — it gets no beat and no fire projectile of its own.
    // causeOf[j] = index of the attack/heropower this damage belongs to, else null.
    const causeOf: (number | null)[] = events.map(() => null)
    for (let i = 0; i < events.length; i++) {
      const k = events[i].kind
      if (k !== 'attack' && k !== 'heropower') continue
      for (let j = i + 1; j < events.length; j++) {
        if (events[j].kind === 'damage' || events[j].kind === 'shield') causeOf[j] = i
        else break
      }
    }
    // Reveal the opponent's latest cast center-stage (their hand is hidden, so the
    // server sends the card on 'play' events). Last play this action wins; it
    // swaps any showing card and restarts the ~2.5s timer.
    const oppPlay = [...events].reverse().find((e) => e.kind === 'play' && e.source !== snap.you && e.card)
    if (oppPlay?.card) showReveal({ card: oppPlay.card })

    // Play this action's events on a timeline so they read one-after-another
    // (HS-style): the player can SEE each cause — two edge triggers fire in turn,
    // a multi-shot card lands its hits as a volley. `t` is the running offset; each
    // handled kind advances it by a beat. Tunables (ms):
    const BEAT = 400 // gap between discrete events (a trigger, a heal, a secret)
    const RANGED_BEAT = 260 // shorter, so a multi-target shooter still reads as one volley
    const LUNGE_HIT = 220 // when a melee lunge connects with its target
    const FLY_HIT = 560 // when a thrown projectile arrives (projectile() runs 700ms)
    const timers: number[] = []
    const at = (ms: number, fn: () => void) => void timers.push(window.setTimeout(fn, ms))

    // Start-of-Game reveals (after the mulligan): show each fired card center-stage,
    // the viewer's own first, then the opponent's — one after another with a gap so
    // both players understand a changed Hero Power. Returns early; these are the only
    // events in this snapshot's reveal beat.
    const startGames = events.filter((e) => e.kind === 'startgame' && e.card)
    if (startGames.length) {
      const SG_GAP = 2600 // ms each Start-of-Game card holds before the next
      startGames
        .sort((a, b) => (a.source === snap.you ? 0 : 1) - (b.source === snap.you ? 0 : 1))
        .forEach((e, k) => {
          const mine = e.source === snap.you
          at(k * SG_GAP, () =>
            showReveal({ card: e.card, label: `🔮 ${mine ? 'Your' : "Opponent's"} Start of Game` }),
          )
        })
    }

    let t = 0
    events.forEach((e, i) => {
      if (causeOf[i] !== null) return // an impact hit — flashed by its cause below
      if (e.kind === 'attack') {
        const s = cidFor(e.source)
        const t0 = t
        if (s) {
          const tdst = cidFor(e.target)
          const dst = tdst ? dstOf(tdst) : null
          if (dst) at(t0, () => lunge(s, dst))
        }
        // Flash each hit this attack dealt (attacker→defender, defender→attacker)
        // on the lunge's connect.
        events.forEach((d, j) => {
          if (causeOf[j] !== i || d.kind !== 'damage') return
          const dt = cidFor(d.target)
          if (dt)
            at(t0 + LUNGE_HIT, () => {
              hitFlash(dt, d.amount)
              releaseHp(dt)
            })
        })
        t += BEAT
      } else if (e.kind === 'heropower') {
        // Fling the power's icon from the caster's button to its target (the paired
        // damage event), then the damage flash plays on arrival.
        const from = e.source === snap.you ? '[data-cid="heroPower"]' : '[data-cid="oppHeroPower"]'
        const t0 = t
        events.forEach((d, j) => {
          if (causeOf[j] !== i || d.kind !== 'damage') return
          const dt = cidFor(d.target)
          const dst = dt ? dstOf(dt) : null
          if (dst && dt) {
            at(t0, () => projectile(from, dst, hpIcon(e.name ?? '')))
            at(t0 + FLY_HIT, () => {
              hitFlash(dt, d.amount)
              releaseHp(dt)
            })
          }
        })
        t += BEAT
      } else if (e.kind === 'damage') {
        // A ranged hit (spell / hero-power-less trigger / missile / finalGasp):
        // fling 🔥 from the shooter to the target, flash on arrival. The shooter id
        // is set server-side (minion uid, or the caster hero for a hand spell).
        const t0 = t
        const tdt = cidFor(e.target)
        const dst = tdt ? dstOf(tdt) : null
        const s = cidFor(e.source)
        const fromSel = s ? `[data-cid="${CSS.escape(s)}"]` : null
        if (dst && fromSel && document.querySelector(fromSel)) {
          at(t0, () => projectile(fromSel, dst, '🔥'))
          if (tdt)
            at(t0 + FLY_HIT, () => {
              hitFlash(tdt, e.amount)
              releaseHp(tdt)
            })
        } else if (tdt) {
          at(t0 + 120, () => {
            hitFlash(tdt, e.amount)
            releaseHp(tdt)
          })
        }
        t += RANGED_BEAT
      } else if (e.kind === 'heal') {
        const tg = cidFor(e.target)
        if (tg) at(t, () => healPop(tg, e.amount))
        t += BEAT
      } else if (e.kind === 'fatigue') {
        const tg = cidFor(e.target)
        const mine = e.target === snap.you
        const t0 = t
        at(t0, () => fatigueBurst(mine ? 'self' : 'opp'))
        if (!mine) showReveal({ fatigue: true }) // preview the opponent's fatigue on the left
        if (tg) at(t0 + 750, () => fatiguePop(tg, e.amount)) // hero flash + number after the card reaches center
        t += BEAT
      } else if (e.kind === 'burn') {
        const card = e.card
        const side = e.target === snap.you ? 'self' : 'opp'
        // The card face flies from the deck pile to center, growing + flaming, then fades.
        if (card) at(t, () => showBurnFace(card, side))
        t += BEAT
      } else if (e.kind === 'secret' && e.card) {
        const card = e.card
        // A Counter-type secret carries the negated spell's name in `note`; that spell's
        // own `play` event is in this same batch (emitPlay runs before triggerSecrets),
        // so pair its CardView to show "<secret> counters <spell>" with both cards.
        const counteredCard = e.note
          ? events.find((p) => p.kind === 'play' && p.card?.cardType === 'spell' && p.card?.name === e.note)?.card
          : undefined
        const label = e.note ? `🔮 ${card.name} counters ${e.note}` : undefined
        at(t, () => {
          showReveal({ card, counteredCard, triggered: true, label }) // a secret just triggered — reveal it
          fireSecretFx(card.name)
        })
        t += BEAT
      } else if (e.kind === 'trigger') {
        // An in-play minion's edge trigger fired — pulse it so the player sees the
        // cause; its effect (damage/summon/buff) animates a beat later.
        const s = cidFor(e.source)
        if (s)
          at(t, () => {
            const el = document.querySelector(`[data-cid="${CSS.escape(s)}"]`)
            if (el) {
              el.classList.add('trigger-pulse')
              window.setTimeout(() => el.classList.remove('trigger-pulse'), 1200)
            }
          })
        t += BEAT
      }
      // Other kinds (play/summon/death/buff/…) are driven by the snapshot-diff
      // effect or need no animation; they don't advance the timeline.
    })
    return () => timers.forEach((id) => window.clearTimeout(id))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [anim])

  // Cancel an in-flight drag if the turn passes or the game ends.
  useEffect(() => {
    if (!myTurn || winner) {
      pressRef.current = null
      setDrag(null)
    }
  }, [myTurn, winner])

  // Global pointer wiring for drag-to-play. Lives on window so a drag tracks the
  // cursor anywhere and a release/second-click lands wherever it ends.
  useEffect(() => {
    const THRESHOLD = 6 // px before a press becomes a hold-drag

    const captureCenters = () => {
      const centers: number[] = []
      document.querySelectorAll<HTMLElement>('.zone.bottom .board [data-cid]').forEach((n) => {
        const r = n.getBoundingClientRect()
        centers.push(r.left + r.width / 2)
      })
      centersRef.current = centers
    }
    const slotFor = (x: number) => {
      let i = 0
      for (const c of centersRef.current) if (x > c) i++
      return i
    }
    // A valid drop = over the play area but not back over the hand.
    const overTable = (x: number, y: number): boolean => {
      const hand = document.querySelector('.hand')?.getBoundingClientRect()
      if (hand && y >= hand.top && x >= hand.left && x <= hand.right) return false
      const area = document.querySelector('.play-area')?.getBoundingClientRect()
      return !!area && x >= area.left && x <= area.right && y >= area.top && y <= area.bottom
    }
    const posFor = (card: CardView, x: number, y: number): number | null =>
      card.cardType === 'minion' && overTable(x, y) ? slotFor(x) : null

    const commit = (d: NonNullable<typeof drag>, x: number, y: number) => {
      setDrag(null)
      dragRef.current = null
      pressRef.current = null
      emitIntent() // drag ended → clear our aim hint for the opponent
      if (!overTable(x, y)) return // off the table → cancel, card stays in hand
      onHandCard(d.handIndex, d.card, d.card.cardType === 'minion' ? slotFor(x) : undefined)
    }

    const onMove = (e: PointerEvent) => {
      const d = dragRef.current
      if (d) {
        setDrag({ ...d, x: e.clientX, y: e.clientY, pos: posFor(d.card, e.clientX, e.clientY) })
        dragRef.current = { ...d, x: e.clientX, y: e.clientY } // so emitIntent reads the live cursor
        emitIntent() // stream the drag's aim to the opponent
        return
      }
      const p = pressRef.current
      if (p && Math.hypot(e.clientX - p.x, e.clientY - p.y) > THRESHOLD) {
        captureCenters()
        setDrag({
          handIndex: p.handIndex,
          card: p.card,
          sticky: false,
          x: e.clientX,
          y: e.clientY,
          pos: posFor(p.card, e.clientX, e.clientY),
        })
      }
    }
    // Resolve the character under a screen point (for release-to-cast aiming).
    const targetFromPoint = (
      x: number,
      y: number,
    ): { targetId: string; kind: CharKind; m?: MinionView } | null => {
      const node = document.elementFromPoint(x, y)?.closest('[data-cid]')
      const cid = node?.getAttribute('data-cid')
      if (!cid) return null
      if (cid === 'oppHero') return { targetId: cid, kind: 'oppHero' }
      if (cid === 'selfHero') return { targetId: cid, kind: 'selfHero' }
      const s = snapRef.current
      const em = s.opp.board.find((mn) => mn.instanceId === cid)
      if (em) return { targetId: cid, kind: 'enemyMinion', m: em }
      const fm = s.self.board.find((mn) => mn.instanceId === cid)
      if (fm) return { targetId: cid, kind: 'friendlyMinion', m: fm }
      return null
    }

    const onUp = (e: PointerEvent) => {
      // Aiming a targeted spell: release over a legal target casts it; releasing
      // elsewhere leaves it armed for a follow-up click on the target.
      if (aimRef.current) {
        aimRef.current = null
        const hit = targetFromPoint(e.clientX, e.clientY)
        if (hit) onCharRef.current(hit.targetId, hit.kind, hit.m)
        emitIntent() // aim released → refresh/clear the hint
        return
      }
      const d = dragRef.current
      if (d && !d.sticky) {
        commit(d, e.clientX, e.clientY) // hold-drag release
        return
      }
      const p = pressRef.current
      if (p && !d) {
        // A click with no movement → lift into sticky mode (click again to drop).
        captureCenters()
        setDrag({
          handIndex: p.handIndex,
          card: p.card,
          sticky: true,
          x: e.clientX,
          y: e.clientY,
          pos: posFor(p.card, e.clientX, e.clientY),
        })
        pressRef.current = null
      }
    }
    // While sticky, the next press anywhere drops — captured before the hand
    // card's own handler so it doesn't start a fresh press.
    const onDownCapture = (e: PointerEvent) => {
      const d = dragRef.current
      if (d?.sticky) {
        e.preventDefault()
        e.stopPropagation()
        commit(d, e.clientX, e.clientY)
      }
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        pressRef.current = null
        setDrag(null)
        dragRef.current = null
        emitIntent() // cancelled → clear our aim hint
      }
    }

    window.addEventListener('pointermove', onMove)
    window.addEventListener('pointerup', onUp)
    window.addEventListener('pointerdown', onDownCapture, true)
    window.addEventListener('keydown', onKey)
    return () => {
      window.removeEventListener('pointermove', onMove)
      window.removeEventListener('pointerup', onUp)
      window.removeEventListener('pointerdown', onDownCapture, true)
      window.removeEventListener('keydown', onKey)
    }
  }, [onHandCard])

  // A hand card press: start a potential drag (unless a sticky drag is already
  // active — the window capture handles dropping that one).
  const onCardPointerDown = (i: number, card: CardView, e: React.PointerEvent) => {
    // Affordability is guarded here (not via a `disabled` attribute) so the card still
    // receives hover events — the opponent must see ANY card we focus, not just playable
    // ones. An unaffordable / off-turn card simply can't be pressed.
    if (!myTurn || winner || dragRef.current || spectator || card.cost > snap.self.mana) return
    e.preventDefault()
    // A targeted card (a spell, or a minion with a targeted onset) doesn't drag
    // onto the table — it lifts toward the board and arms the aim line. onHandCard
    // toggles targeting; if it armed (wasn't already armed on this card), a release
    // over a target casts/plays it (else click the target). Click again to cancel.
    // Targeted-onset minions append (the aim flow doesn't pick a board slot).
    if (card.target && card.target !== 'none') {
      const wasArmed = spell?.handIndex === i
      onHandCard(i, card)
      aimRef.current = wasArmed ? null : { handIndex: i }
      return
    }
    pressRef.current = { handIndex: i, card, x: e.clientX, y: e.clientY }
  }

  return (
    <div
      className="game"
      onMouseMove={(e) => {
        if (sourceId) {
          setPointer({ x: e.clientX, y: e.clientY })
          pointerRef.current = { x: e.clientX, y: e.clientY }
          emitIntent() // stream the attacker/hero-power/spell aim to the opponent
        }
      }}
    >
      <div className="game-stage">
      <div className={'play-area' + (!spectator && myTurn && !winner && lowTime ? ' low-time' : '')}>
        <div className="banners">
          {!winner && !oppOnline && (
            <div className="banner warn">
              ⚠ Opponent disconnected — waiting for them to reconnect…
            </div>
          )}
        </div>

        {winner && winnerShown && (
          <div className="overlay gameover-overlay">
            <div className={'gameover-modal ' + (winner === 'you' ? 'win' : 'lose')}>
              <div className="go-result">
                {spectator
                  ? `${(winner === 'you' ? snap.self.name : snap.opp.name) || 'Player'} wins`
                  : winner === 'you'
                    ? 'Victory!'
                    : 'Defeat'}
              </div>
              {!spectator && <div className="go-sub">vs {snap.opp.name || 'Opponent'}</div>}
              {/* Ranked ladder change (ranked PvP only). Lower number = better:
                  green when you climbed (or placed for the first time), red when you
                  dropped, white when unchanged. */}
              {!spectator && rankUpdate
                ? (() => {
                    const { oldRank, newRank } = rankUpdate
                    const improved = oldRank === 0 || newRank < oldRank
                    const worse = oldRank !== 0 && newRank > oldRank
                    const cls = improved ? 'up' : worse ? 'down' : 'same'
                    const arrow = improved ? '▲' : worse ? '▼' : '—'
                    return (
                      <div className={'go-rank ' + cls}>
                        <span className="go-rank-num">Rank #{newRank}</span>
                        <span className="go-rank-delta">
                          {arrow} {oldRank === 0 ? 'New rank' : `from #${oldRank}`}
                        </span>
                      </div>
                    )
                  })()
                : <div className="go-stats" />}
              <button className="go-exit" onClick={onBackToLobby}>
                Back to lobby
              </button>
            </div>
          </div>
        )}

        {seek && (
          <div className="overlay">
            <div className="seek">
              <div className="seek-title">Seek — pick a card</div>
              <div className="seek-options">
                {seek.map((c, i) => (
                  <button
                    key={i}
                    className={'card' + cardColorClass(c)}
                    onClick={() => send({ type: 'choose', index: i })}
                  >
                    <CardFace card={c} />
                  </button>
                ))}
              </div>
            </div>
          </div>
        )}

        {cast && (
          <div className="cast-show" key={cast.key}>
            <div className="cast-label">
              {cast.label
                ? cast.label
                : cast.triggered
                  ? '🔮 Secret triggered!'
                  : cast.fatigue
                    ? 'Opponent fatigued'
                    : 'Opponent played'}
            </div>
            {cast.fatigue ? (
              <>
                <div className="card cast-card fatigue-preview">
                  <div className="name">Fatigue</div>
                  <div className="art">💀</div>
                  <div className="stats">—</div>
                </div>
                <div className="cast-text">No cards left to draw — takes increasing damage each turn.</div>
              </>
            ) : cast.secret ? (
              <div className="card cast-card secret">
                <div className="name">Secret</div>
                <div className="art">❓</div>
              </div>
            ) : (
              cast.card && (
                <div className="cast-pair">
                  {cast.counteredCard && (
                    <>
                      <div className={'card cast-card countered' + cardColorClass(cast.counteredCard)}>
                        <CardFace card={cast.counteredCard} />
                      </div>
                      <span className="cast-x">✕</span>
                    </>
                  )}
                  <div className={'card cast-card' + cardColorClass(cast.card)}>
                    <CardFace card={cast.card} />
                  </div>
                </div>
              )
            )}
          </div>
        )}

        {burnFace && (
          <div
            className="burn-show"
            key={burnFace.key}
            style={{ '--bx': `${burnFace.bx}px`, '--by': `${burnFace.by}px` } as React.CSSProperties}
          >
            <div className={'card burn-show-card' + cardColorClass(burnFace.card)}>
              <CardFace card={burnFace.card} />
            </div>
            <div className="burn-show-flame">🔥</div>
          </div>
        )}

        {secretFx && (
          <div className="secret-burst" key={secretFx.key}>
            <div className="secret-burst-ring" />
            <div className="secret-burst-label">🔮 {secretFx.name}</div>
          </div>
        )}

        {!spectator && watchers.length > 0 && (
          <div className="watchers">
            👁 {watchers.length}
            <div className="watchers-list">
              <div className="watchers-title">Watching</div>
              {watchers.map((w) => (
                <div key={w} className="watchers-name">
                  {w}
                </div>
              ))}
            </div>
          </div>
        )}

        <DeckPile side="opp" count={snap.opp.deckCount} />
        <DeckPile side="self" count={snap.self.deckCount} />

        <div className="player-name opp">
          {snap.opp.rank ? <span className="pn-rank">#{snap.opp.rank}</span> : null}
          {snap.opp.name || 'Opponent'}
        </div>
        <div className="player-name self">
          {snap.self.rank ? <span className="pn-rank">#{snap.self.rank}</span> : null}
          {snap.self.name || name}
        </div>

        {spectator ? (
          <div className="spectate-banner">
            👁 Spectating <strong>{spectating}</strong>
            <button className="spectate-leave" onClick={onBackToLobby}>
              Leave
            </button>
          </div>
        ) : (
          <>
            {myTurn && !winner && (
              <button className="end-turn" onClick={() => send({ type: 'end_turn' })}>
                End Turn
              </button>
            )}
            {!winner && (
              <button className="concede" onClick={() => setConfirmConcede(true)}>
                Concede
              </button>
            )}
          </>
        )}

        {confirmConcede && (
          <div className="overlay">
            <div className="confirm-modal">
              <div className="confirm-title">Concede this match?</div>
              <div className="confirm-actions">
                <button
                  className="confirm-yes"
                  onClick={() => {
                    send({ type: 'concede' })
                    setConfirmConcede(false)
                  }}
                >
                  Concede
                </button>
                <button className="confirm-no" onClick={() => setConfirmConcede(false)}>
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Opponent (mirrors your side: mana on top, then hand backs). */}
        <div className="zone top">
          <ManaBar side="opp" mana={snap.opp.mana} max={snap.opp.maxMana} />
          <CardBacks count={snap.opp.handCount} lift={oppLift} />
          <div className="self-row">
            {snap.opp.heroPower && (
              <div
                className={'hp-button static' + (snap.opp.heroPowerUsed ? ' used' : '')}
                data-cid="oppHeroPower"
                title={snap.opp.heroPower.text}
              >
                <span className="hp-cost">{snap.opp.heroPower.cost}</span>
                <span className="hp-flip">
                  <span className="hp-face front">
                    <span
                      className="hp-art"
                      style={{ backgroundImage: `url(/art/${snap.opp.heroPower.cardId}.png)` }}
                    >
                      {hpIcon(snap.opp.heroPower.name)}
                    </span>
                  </span>
                  <span className="hp-face back" aria-hidden="true" />
                </span>
                <CardTooltip
                  name={snap.opp.heroPower.name}
                  kind="heroPower"
                  cost={snap.opp.heroPower.cost}
                  text={snap.opp.heroPower.text}
                />
              </div>
            )}
            <Hero
              side="opp"
              p={snap.opp}
              targetable={targetable('oppHero')}
              selected={false}
              ready={false}
              hpOverride={heldHp.oppHero}
              onClick={() => onChar('oppHero', 'oppHero')}
            />
          </div>
          <Board minions={snap.opp.board} enemy attacker={attacker} held={heldHp} targetable={targetable} onChar={onChar} />
        </div>

        <div className="midline">
          <div className="turn-chip">
            <span className={'turn-flag' + (myTurn ? ' mine' : '')}>
              {winner
                ? 'Game over'
                : spectator
                  ? `${(myTurn ? snap.self.name : snap.opp.name) || 'Player'}'s turn`
                  : myTurn
                    ? 'YOUR TURN'
                    : "Opponent's turn"}
              {hint}
            </span>
            {!winner && <TurnTimer key={turnNum} secs={turnSecs} onLow={setLowTime} />}
          </div>
        </div>

        {/* Me */}
        <div className="zone bottom">
          <Board
            minions={snap.self.board}
            myTurn={!spectator && myTurn && !winner}
            attacker={attacker}
            dropIndex={drag && drag.card.cardType === 'minion' ? drag.pos : null}
            held={heldHp}
            targetable={targetable}
            onChar={onChar}
          />

          <div className="self-row">
            {hp && (
              <button
                data-cid="heroPower"
                className={
                  'hp-button' +
                  (heroPowerArmed ? ' selected' : '') +
                  (snap.self.heroPowerUsed ? ' used' : '') +
                  // Green "usable now" outline (like a playable card), unless it's already armed.
                  (!spectator && myTurn && !winner && !snap.self.heroPowerUsed && hp.cost <= snap.self.mana && !heroPowerArmed
                    ? ' ready'
                    : '')
                }
                disabled={spectator || !myTurn || !!winner || snap.self.heroPowerUsed || hp.cost > snap.self.mana}
                onClick={onHeroPower}
              >
                <span className="hp-cost">{hp.cost}</span>
                <span className="hp-flip">
                  <span className="hp-face front">
                    <span className="hp-art" style={{ backgroundImage: `url(/art/${hp.cardId}.png)` }}>
                      {hpIcon(hp.name)}
                    </span>
                  </span>
                  <span className="hp-face back" aria-hidden="true" />
                </span>
                <CardTooltip name={hp.name} kind="heroPower" cost={hp.cost} text={hp.text} />
              </button>
            )}
            <Hero
              side="self"
              p={snap.self}
              targetable={targetable('selfHero')}
              selected={attacker === 'selfHero'}
              ready={!spectator && myTurn && !winner && !!snap.self.heroCanAttack}
              hpOverride={heldHp.selfHero}
              onClick={() => onChar('selfHero', 'selfHero')}
            />
          </div>

        {/* Your hand */}
        <div className="hand">
          {(snap.self.hand ?? []).map((c, i) => {
            const affordable = myTurn && !winner && c.cost <= snap.self.mana
            const armed = spell?.handIndex === i
            return (
              <button
                key={i}
                data-cid={`hand-${i}`}
                className={
                  'card' +
                  cardColorClass(c) +
                  (affordable ? ' playable' : '') +
                  (armed ? ' selected aiming' : '') +
                  (drag?.handIndex === i ? ' dragging' : '')
                }
                onPointerDown={(e) => onCardPointerDown(i, c, e)}
                onMouseEnter={() => {
                  // Tell the opponent which hand card we're considering (HS lifts it).
                  hoverHandRef.current = i
                  emitIntent()
                }}
                onMouseLeave={() => {
                  if (hoverHandRef.current === i) {
                    hoverHandRef.current = -1
                    emitIntent()
                  }
                }}
              >
                <CardFace card={c} />
              </button>
            )
          })}
          </div>
          <ManaBar side="self" mana={snap.self.mana} max={snap.self.maxMana} />
        </div>
      </div>

      <aside className="log">
        <div className="log-title">Log</div>
        {log.length === 0 && <div className="log-empty">—</div>}
        <div className="log-feed">
          {log.slice(0, 25).map((e, i) => {
            const primary = e.source ?? e.targets[0]?.actor
            const died = e.targets.some((t) => t.died)
            return (
              <div
                key={i}
                className={'log-row k-' + e.kind}
                onMouseEnter={(ev) => {
                  // Anchor the popup just left of the row, at its Y, clamped on-screen.
                  const r = (ev.currentTarget as HTMLElement).getBoundingClientRect()
                  const H = 200 // approx popup height for the clamp
                  const top = Math.min(Math.max(r.top + r.height / 2, 8 + H / 2), window.innerHeight - 8 - H / 2)
                  setLogPop({ entry: e, rightPx: window.innerWidth - r.left + 10, topPx: top })
                }}
                onMouseLeave={() => setLogPop(null)}
              >
                {primary && <LogChip actor={primary} />}
                <span className="log-rel-icon">{kindIcon(e.kind)}</span>
                {/* A grouped multi-target hit shows how many were affected; a death shows ☠. */}
                {died && <span className="log-died">☠</span>}
                {e.targets.length > 1 && <span className="log-count">×{e.targets.length}</span>}
              </div>
            )
          })}
        </div>
      </aside>

      {/* Log detail popup — portaled to <body> so it escapes the game-stage transform
          and can be placed with true viewport coordinates (tracks the hovered row's Y,
          clamped on-screen). Shows the real cards: source → action → affected, + a
          plain-language caption so card-less events (mana/armor/fatigue) read clearly. */}
      {logPop &&
        createPortal(
          <LogDetail entry={logPop.entry} rightPx={logPop.rightPx} topPx={logPop.topPx} />,
          document.body,
        )}

      {/* Opponent is seeking: a non-blocking peek near their hand, faces hidden. */}
      {oppSeek != null && (
        <div className="opp-seek">
          <div className="opp-seek-label">Opponent is seeking…</div>
          <div className="opp-seek-cards">
            {Array.from({ length: oppSeek }, (_, i) => (
              <span key={i} className="card-back opp-seek-card" />
            ))}
          </div>
        </div>
      )}
      </div>

      {/* The card following the cursor while dragging from hand. */}
      {drag && (
        <div
          className={'card drag-card' + cardColorClass(drag.card)}
          style={{ left: drag.x, top: drag.y }}
        >
          <CardFace card={drag.card} />
        </div>
      )}

      {sourceId && <TargetingArrow sourceId={sourceId} pointer={pointer} />}
      {showOppIntent && oppIntent?.aimFrom && oppIntent?.aimTo && (
        <OppAimArrow from={oppTokenCenter(oppIntent.aimFrom)} to={oppTokenCenter(oppIntent.aimTo)} />
      )}
    </div>
  )
}

// TurnTimer counts down locally from the seconds the last snapshot reported,
// resetting whenever a new snapshot arrives (each action/turn re-syncs it).
// LogChip renders one log participant: a hero portrait (round, side-coloured) or a
// card art chip (`/art/<cardId>.png`). A card with no/missing art still shows its
// name on hover; the dark chip is the fallback.
function LogChip({ actor }: { actor: LogActor }) {
  if (actor.hero) {
    return <span className={'log-chip hero ' + actor.hero} style={{ backgroundImage: 'url(/art/mage_hero.png)' }} />
  }
  return (
    <span
      className="log-chip"
      style={actor.cardId ? { backgroundImage: `url(/art/${actor.cardId}.png)` } : undefined}
    />
  )
}

// LogActorBig renders a log participant at full size for the on-hover popup: the
// real card face for a card, a large round portrait for a hero.
function LogActorBig({ actor }: { actor: LogActor }) {
  if (actor.hero) {
    return (
      <div className={'log-pop-hero ' + actor.hero} style={{ backgroundImage: 'url(/art/mage_hero.png)' }}>
        <span>{actor.name}</span>
      </div>
    )
  }
  if (actor.card) {
    // CardFace returns absolutely-positioned pieces; it MUST sit in a sized `.card`
    // wrapper (same as the hand / cast preview) or it collapses to a text stack.
    return (
      <div className={'card log-pop-card' + cardColorClass(actor.card)}>
        <CardFace card={actor.card} />
      </div>
    )
  }
  return <div className="log-pop-fallback">{actor.name}</div>
}

// LogDetail is the body-portaled hover popup for one log row: the source card on the
// left, then every affected character (each with its own outcome verb/amount + a ☠ if
// the effect killed it). A self-targeting source (a minion's own play, a self-buff)
// shows once. A keyword note (Onset/Final Gasp/…) tags the cause.
function LogDetail({ entry: e, rightPx, topPx }: { entry: LogEntry; rightPx: number; topPx: number }) {
  // Drop a target that is just the source card again at unchanged stats (a minion's own
  // play). A buff keeps its target (it shows before → after, never a plain dupe).
  const targets = e.targets.filter(
    (t) => !(t.actor?.uid && t.actor.uid === e.source?.uid && t.kind !== 'buff'),
  )
  // A self-only effect (one target, same instance as the source — e.g. a self-buff) reads
  // as just that card's outcome; hide the redundant source so a buff shows clean before →
  // after. Keyed on instance uid: a minion attacking a same-named minion is NOT self.
  const selfOnly =
    !!e.source?.uid && targets.length === 1 && targets[0].actor?.uid === e.source.uid
  return (
    <div className="log-pop" style={{ right: rightPx, top: topPx }}>
      <div className="log-pop-cards">
        {e.source && !selfOnly && <LogActorBig actor={e.source} />}
        {e.source && !selfOnly && (
          <span className="log-pop-rel">
            <span className="log-pop-verb">{kindIcon(e.kind)}</span>
            {e.note ? <span className="log-note">{e.note}</span> : null}
          </span>
        )}
        {targets.length > 0 && (
          <div className="log-pop-targets">
            {targets.map((t, i) => (
              <div className="log-pop-target" key={i}>
                {/* A buff shows the card before → after; everything else shows the one card. */}
                <div className="log-pop-target-cards">
                  {t.before && <LogActorBig actor={t.before} />}
                  {t.before && <span className="log-pop-verb">→</span>}
                  {t.actor && <LogActorBig actor={t.actor} />}
                </div>
                <span className="log-pop-outcome">
                  <span className="log-pop-verb">{kindIcon(t.kind)}</span>
                  {t.amount ? <span className={'log-amt' + (t.kind === 'heal' ? ' heal' : '')}>{t.amount}</span> : null}
                  {t.buffAtk || t.buffHp ? <span className="log-amt buff">+{t.buffAtk ?? 0}/+{t.buffHp ?? 0}</span> : null}
                  {t.died && <span className="log-died">☠</span>}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
      <div className="log-pop-text">{e.text}</div>
    </div>
  )
}

function TurnTimer({ secs, onLow }: { secs: number; onLow?: (low: boolean) => void }) {
  const [left, setLeft] = useState(secs)
  useEffect(() => {
    setLeft(secs)
    if (secs <= 0) return
    const id = setInterval(() => setLeft((l) => (l > 0 ? l - 1 : 0)), 1000)
    return () => clearInterval(id)
  }, [secs])
  const low = left > 0 && left <= 15
  // Surface low-time to the parent so it can flash the board's action outlines red.
  useEffect(() => onLow?.(low), [low, onLow])
  useEffect(() => () => onLow?.(false), [onLow]) // clear when the timer unmounts
  if (secs <= 0) return null
  return <span className={'turn-timer' + (low ? ' low' : '')}>⏳ {left}s</span>
}

// ManaBar renders 10 fixed crystal slots in a corner (so growing mana never
// shifts the layout): filled = available, owned-but-spent = dim, beyond the
// player's max = locked.
function ManaBar({ side, mana, max }: { side: 'self' | 'opp'; mana: number; max: number }) {
  return (
    <div className={'mana-bar ' + side} title={`${mana}/${max} mana`}>
      {Array.from({ length: 10 }, (_, i) => {
        let cls = 'crystal'
        if (i < mana) cls += ' on'
        else if (i < max) cls += ' spent'
        else cls += ' locked'
        return <span key={i} className={cls} />
      })}
      <span className="crystal-count">
        {mana}/{max}
      </span>
    </div>
  )
}

// DeckPile is the draw pile: a stack of face-down cards anchored on the right
// (near End Turn, HS-style). Count shows in a hover tooltip.
function DeckPile({ side, count }: { side: 'self' | 'opp'; count: number }) {
  // Stack height reflects how full the deck is (30 = 100%): 3 layers >66%, 2
  // >33%, 1 above empty, none at 0. An outline always marks the deck's spot.
  const layers = count >= 20 ? 3 : count >= 10 ? 2 : count >= 1 ? 1 : 0
  return (
    <div className={'deck-pile ' + side}>
      {Array.from({ length: layers }, (_, i) => (
        <span key={i} className="deck-card-back" />
      ))}
      <span className="deck-outline" />
      {count === 0 && <span className="deck-empty">∅</span>}
      <span className="deck-tip">{count} cards</span>
    </div>
  )
}

// CardBacks renders the opponent's hidden hand as face-down card backs. `lift` is the
// index of the card the opponent is currently holding/aiming (-1 = none) — it rises and
// glows so we see which card they're considering, HS-style.
function CardBacks({ count, lift = -1 }: { count: number; lift?: number }) {
  return (
    <div className="opp-hand">
      {Array.from({ length: count }, (_, i) => (
        <span key={i} className={'card-back' + (i === lift ? ' lifted' : '')} />
      ))}
      <span className="opp-hand-count">{count} cards</span>
    </div>
  )
}

// oppTokenCid maps an opponent-intent wire token to a local data-cid (viewer POV): the
// sender's 'enemy' hero is us, their 'self' hero is the opponent, their hero power is
// the opponent's. Minion uids are match-global. (A hand:<i> token has no data-cid.)
function oppTokenCid(token: string): string {
  if (token === 'enemy') return 'selfHero'
  if (token === 'self') return 'oppHero'
  if (token === 'heroPower') return 'oppHeroPower'
  return token
}

// oppTokenCenter resolves a wire token to a viewport center point for the ghost arrow.
function oppTokenCenter(token?: string): { x: number; y: number } | null {
  if (!token) return null
  let el: HTMLElement | null
  if (token.startsWith('hand:')) {
    el = document.querySelectorAll<HTMLElement>('.opp-hand .card-back')[Number(token.slice(5))] ?? null
  } else {
    el = document.querySelector<HTMLElement>(`[data-cid="${CSS.escape(oppTokenCid(token))}"]`)
  }
  if (!el) return null
  const r = el.getBoundingClientRect()
  return { x: r.left + r.width / 2, y: r.top + r.height / 2 }
}

// OppAimArrow draws the opponent's in-progress aim as a dimmed dashed line between two
// resolved character centers (source → current target). Snaps target-to-target (no live
// cursor — coordinates wouldn't translate across the flipped board). Null if either end
// can't be resolved this frame.
function OppAimArrow({ from, to }: { from: { x: number; y: number } | null; to: { x: number; y: number } | null }) {
  if (!from || !to) return null
  return (
    <svg className="opp-aim-arrow">
      <defs>
        <marker id="opp-arrowhead" markerWidth="8" markerHeight="8" refX="5" refY="4" orient="auto">
          <path d="M0,0 L8,4 L0,8 Z" fill="#b07cff" />
        </marker>
      </defs>
      <line
        x1={from.x}
        y1={from.y}
        x2={to.x}
        y2={to.y}
        stroke="#b07cff"
        strokeWidth="4"
        strokeDasharray="9 7"
        strokeLinecap="round"
        markerEnd="url(#opp-arrowhead)"
      />
      <circle cx={from.x} cy={from.y} r="6" fill="#b07cff" />
    </svg>
  )
}
