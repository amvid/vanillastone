import { useCallback, useEffect, useRef, useState } from 'react'
import { listDecks, login, register } from './api'
import type { Deck } from './api'
import { Deckbuilder } from './Deckbuilder'
import type { CardView, Event, MinionView, OppIntent, PlayerInfo, ServerMessage, Snapshot } from './protocol'
import type { CharKind, Counts, LogEntry, PendingSpell, Phase } from './game/types'
import { buildLog, cardColorClass, condMet, minionToCardView, ruleMatches } from './game/format'
import { playGhost } from './game/animate'
import { CardFace } from './game/CardFace'
import { GameScreen } from './game/GameScreen'

const TOKEN_KEY = 'vs_token'

// Opening-mulligan time limit (seconds). At 0 the client auto-keeps the current
// selection; the server backstops a hair later (see mulliganLimit).
const MULLIGAN_SECS = 20

/** Custom deck dropdown — options show the class art icon like the deck-builder rows. */
function DeckSelect({
  options,
  value,
  onChange,
}: {
  options: { id: number; name: string; class: string }[]
  value: number
  onChange: (id: number) => void
}) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const close = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', close)
    return () => document.removeEventListener('mousedown', close)
  }, [open])
  const sel = options.find((o) => o.id === value) ?? options[0]
  if (!sel) return null // no decks loaded yet
  return (
    <div className={'deck-select' + (open ? ' open' : '')} ref={ref}>
      <button type="button" className="deck-select-trigger" onClick={() => setOpen((v) => !v)}>
        <span className="deck-row-art" style={{ backgroundImage: `url('/art/${sel.class}_hero.png')` }} />
        <span className="deck-row-text">{sel.name}</span>
        <span className="deck-select-caret">▾</span>
      </button>
      {open && (
        <div className="deck-select-menu">
          {options.map((o) => (
            <button
              key={o.id}
              type="button"
              className={'deck-select-opt' + (o.id === value ? ' active' : '')}
              onClick={() => {
                onChange(o.id)
                setOpen(false)
              }}
            >
              <span className="deck-row-art" style={{ backgroundImage: `url('/art/${o.class}_hero.png')` }} />
              <span className="deck-row-text">{o.name}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

export function App() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [phase, setPhase] = useState<Phase>('auth')
  const [status, setStatus] = useState('not logged in')
  const [name, setName] = useState('')
  const [counts, setCounts] = useState<Counts>({ online: 0, inGame: 0 })
  const [players, setPlayers] = useState<PlayerInfo[]>([])
  // Free-text filter for the online-players panel (kept usable with many players).
  const [playerFilter, setPlayerFilter] = useState('')
  // Direct invites: who we've challenged (one at a time, null = none), and the
  // queue of players who have challenged us (each can be accepted/declined).
  const [invitedName, setInvitedName] = useState<string | null>(null)
  const [incomingInvites, setIncomingInvites] = useState<string[]>([])
  // Deck chosen in the incoming-invite prompt.
  const [inviteDeck, setInviteDeck] = useState<number>(0)
  const [snap, setSnap] = useState<Snapshot | null>(null)
  const [myTurn, setMyTurn] = useState(false)
  const [turnSecs, setTurnSecs] = useState(0)
  const [turnNum, setTurnNum] = useState(0)
  const [winner, setWinner] = useState<string | null>(null)
  // Exactly one of these is active at a time: a minion picked to attack, or a
  // targeted spell awaiting its target.
  const [attacker, setAttacker] = useState<string | null>(null)
  const [spell, setSpell] = useState<PendingSpell | null>(null)
  const [heroPowerArmed, setHeroPowerArmed] = useState(false)
  const [seek, setSeek] = useState<CardView[] | null>(null)
  // Number of cards the opponent is currently seeking (null = not). Drives a
  // non-blocking indicator near their hand; cleared by the next state snapshot.
  const [oppSeek, setOppSeek] = useState<number | null>(null)
  // The opponent's latest ephemeral aiming hint (what they're hovering / aiming at),
  // shown only during their turn. Cleared on every resolved snapshot.
  const [oppIntent, setOppIntent] = useState<OppIntent | null>(null)
  const [log, setLog] = useState<LogEntry[]>([])
  // The most recent action's event list + a monotonic seq, handed to GameScreen
  // so it can replay attack/damage/heal as animations (settled state already in
  // `snap`). Skipped for match_start / resync (no live action to animate).
  const [anim, setAnim] = useState<{ seq: number; events: Event[] } | null>(null)
  const animSeqRef = useRef(0)
  // Decks: the player's saved decks and which one is selected for the next game.
  const [decks, setDecks] = useState<Deck[]>([])
  const [selectedDeck, setSelectedDeck] = useState<number>(0) // set to first saved deck once loaded
  // Class the AI opponent plays (a random prebuilt deck of this class). Mage only
  // for now — the rest are reserved until their card pools land.
  const [aiClass, setAiClass] = useState('mage')
  // Which deck the AI opponent plays: 0 = a random prebuilt deck, or one of the
  // player's own saved deck ids.
  const [aiDeck, setAiDeck] = useState<number>(0)
  // The "how to play" mode picker (vs AI / vs Player / Arena) opened from Play.
  const [playModal, setPlayModal] = useState(false)
  // Mirror of selectedDeck readable from the ws closure (which has no deps).
  const selectedDeckRef = useRef(0)
  // Mulligan phase: which opening-hand indices are toggled for replacement, and
  // whether this player has already submitted (awaiting the opponent).
  const [mulliganPicks, setMulliganPicks] = useState<Set<number>>(new Set())
  const [mulliganSubmitted, setMulliganSubmitted] = useState(false)
  // Seconds left on the opening mulligan; auto-keeps at 0. The server backstops
  // this in case the client never submits (disconnect).
  const [mulliganLeft, setMulliganLeft] = useState(MULLIGAN_SECS)
  // True briefly when play begins from the mulligan, to play a reveal animation
  // (the blurred mulligan look dissolving into the live board) instead of a cut.
  const [introPlay, setIntroPlay] = useState(false)
  // Screen point the opening deal flies from (the mulligan modal's center).
  const [introFrom, setIntroFrom] = useState<{ x: number; y: number } | null>(null)
  const introTimerRef = useRef<number | undefined>(undefined)

  // True while the opponent is connected; false during their disconnect grace
  // window (drives a "waiting for opponent" banner).
  const [oppOnline, setOppOnline] = useState(true)
  // When set, we're watching this player's match read-only (their POV). spectatingRef
  // mirrors it for the dep-less ws closure and the action-gating handlers.
  const [spectating, setSpectating] = useState<string | null>(null)
  const spectatingRef = useRef<string | null>(null)
  // Usernames watching our match (we're a player). Drives the "being watched" badge.
  const [spectators, setSpectators] = useState<string[]>([])
  const wsRef = useRef<WebSocket | null>(null)
  const tokenRef = useRef<string | null>(null)
  const meRef = useRef<string | null>(null)
  // The id whose point of view the board is rendered from: our own when playing,
  // the watched player's when spectating. Equal to each snapshot's `you` field.
  const povRef = useRef<string | null>(null)
  // Set when the server kicks us for logging in elsewhere, so the imminent
  // onclose doesn't overwrite the explanatory status.
  const kickedRef = useRef(false)
  // Auto-reconnect bookkeeping: giveUp suppresses retries (logout / dead token),
  // retries counts attempts, and reconnectTimer holds the pending retry.
  const giveUpRef = useRef(false)
  const retriesRef = useRef(0)
  const reconnectTimerRef = useRef<number | undefined>(undefined)
  // Accumulated minion uid -> name, so log lines can name minions that have
  // already left the board (e.g. a minion that just died).
  const namesRef = useRef<Record<string, string>>({})
  // Accumulated minion uid -> full CardView, so the log can render a card's art
  // chip + on-hover card face even after it has left the board (a dead minion's
  // finalGasp, a bounced minion).
  const cardsRef = useRef<Record<string, CardView>>({})
  // Holds the "Match found!" → mulligan transition timer.
  const matchFoundTimerRef = useRef<number | undefined>(undefined)
  // Last committed phase, readable from the ws closure (which has no deps), so a
  // state message can tell whether it is the first snapshot leaving the mulligan.
  const phaseRef = useRef<Phase>('auth')

  useEffect(() => {
    phaseRef.current = phase
  }, [phase])

  useEffect(() => {
    selectedDeckRef.current = selectedDeck
  }, [selectedDeck])

  // Global UI scale knob. Every screen sizes its elements off --u (via
  // calc(px * var(--u)) tokens), so one viewport-driven value keeps the whole app
  // proportional: ~1 at the 2560x1440 design viewport, scaling down on smaller
  // screens (a 1080p display lands ~0.75 — the manual zoom friends were doing).
  // Capped at 1 so larger screens never upscale past the design.
  useEffect(() => {
    // Design viewport = the display the game was tuned on (looks great at u=1).
    // Smaller screens scale down from here; larger ones are capped at 1 so the
    // tuned look is never upscaled.
    const DESIGN_W = 2056
    const DESIGN_H = 1329
    const apply = () => {
      const u = Math.min(1, window.innerWidth / DESIGN_W, window.innerHeight / DESIGN_H)
      document.documentElement.style.setProperty('--u', String(Math.max(0.65, u)))
    }
    apply()
    window.addEventListener('resize', apply)
    return () => window.removeEventListener('resize', apply)
  }, [])

  const send = useCallback((msg: object) => wsRef.current?.send(JSON.stringify(msg)), [])

  // Mulligan countdown: tick while the mulligan UI is up and unsubmitted, and
  // auto-keep the current selection at 0 (the server backstops a dead client).
  useEffect(() => {
    if (phase !== 'mulligan' || mulliganSubmitted || spectating) return
    setMulliganLeft(MULLIGAN_SECS)
    const id = window.setInterval(() => setMulliganLeft((l) => (l > 0 ? l - 1 : 0)), 1000)
    return () => window.clearInterval(id)
  }, [phase, mulliganSubmitted, spectating])

  useEffect(() => {
    if (phase !== 'mulligan' || mulliganSubmitted || spectating || mulliganLeft > 0) return
    send({ type: 'mulligan', indices: [...mulliganPicks] })
    setMulliganSubmitted(true)
    setStatus('waiting for opponent…')
  }, [phase, mulliganSubmitted, spectating, mulliganLeft, mulliganPicks, send])

  // Tear down the current socket and detach its handlers so its onclose can't
  // fire stale reconnect logic. Used before an explicit (re)login so a lingering
  // OPEN socket (e.g. one left after a bad-token bounce) doesn't make connect()
  // early-return and hang on "Connecting".
  const closeSocket = useCallback(() => {
    const ws = wsRef.current
    if (ws) {
      ws.onopen = null
      ws.onmessage = null
      ws.onclose = null
      ws.close()
      wsRef.current = null
    }
  }, [])

  const connect = useCallback((token: string) => {
    // Guard against opening a second socket. React StrictMode double-invokes the
    // reconnect effect in dev; without this the two connections would race and
    // the second would kick the first (same account) — logging the page out on
    // reload.
    const existing = wsRef.current
    if (existing && existing.readyState <= WebSocket.OPEN) return
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current)
      reconnectTimerRef.current = undefined
    }
    tokenRef.current = token
    meRef.current = null
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const ws = new WebSocket(`${proto}://${location.host}/ws`)
    wsRef.current = ws

    ws.onopen = () => ws.send(JSON.stringify({ type: 'auth', token }))
    ws.onclose = () => {
      // A kick already set the phase + an explanatory status; don't clobber it.
      if (kickedRef.current) {
        kickedRef.current = false
        return
      }
      // Transient drop while we still hold a session: auto-reconnect. The server
      // keeps our match seat open for a grace window, so a quick reconnect
      // resumes the game. Give up after a few tries and fall back to login.
      const tok = tokenRef.current
      if (tok && !giveUpRef.current && retriesRef.current < 6) {
        retriesRef.current++
        setPhase('connecting')
        setStatus(`connection lost — reconnecting… (${retriesRef.current})`)
        reconnectTimerRef.current = window.setTimeout(() => connect(tok), 1500)
        return
      }
      // If we never authenticated, the connection itself failed.
      setStatus(meRef.current ? 'disconnected — log in again' : 'connection closed')
      setPhase('auth')
    }
    ws.onmessage = (e) => {
      const msg: ServerMessage = JSON.parse(e.data)
      switch (msg.type) {
        case 'joined':
          meRef.current = msg.you
          retriesRef.current = 0 // successful (re)connect resets the backoff
          setName(msg.name)
          setPhase('lobby')
          setStatus(`logged in as ${msg.name}`)
          break
        case 'lobby':
          setCounts({ online: msg.online, inGame: msg.inGame })
          setPlayers(msg.players ?? [])
          break
        case 'waiting':
          setPhase('waiting')
          setStatus('waiting for an opponent')
          break
        case 'match_start':
        case 'state': {
          // Render from the snapshot's POV: our own id when playing, the watched
          // player's id when spectating (the server sets `you` to that seat). Using
          // msg.you makes "whose turn" and the event log read from that POV.
          povRef.current = msg.you
          const mine = msg.turn === msg.you
          // Keep the uid->name map current and turn this action's events into
          // log lines (match_start carries no events).
          const names = namesRef.current
          const cards = cardsRef.current
          for (const mn of [...msg.self.board, ...msg.opp.board]) {
            names[mn.instanceId] = mn.name
            cards[mn.instanceId] = minionToCardView(mn) // persistent uid→card map for the log (survives death)
          }
          if (msg.type === 'match_start') {
            setLog([])
            setAnim(null) // drop any leftover action animation from a prior match (else it replays on the new board)
            setOppOnline(true)
            setInvitedName(null)
            setIncomingInvites([])
            setSpectators([])
          } else {
            for (const e of msg.events ?? []) {
              if (e.name && e.source?.[0] === 'u') names[e.source] = e.name
              if (e.name && e.target?.[0] === 'u') names[e.target] = e.name
            }
            const entries: LogEntry[] = buildLog(msg.events ?? [], names, cards, povRef.current)
            if (msg.resync) {
              // Reconnect: Events is the full recent history (chronological).
              // Replace the log, newest on top, recovering what happened while away.
              setLog([...entries].reverse().slice(0, 40))
            } else if (entries.length) {
              // Newest on top, globally: reverse this action's groups so its latest step
              // sits above its earlier ones (matches the resync path) — no zigzag where a
              // death reads as happening before the hit that caused it.
              setLog((prev) => [...entries.slice().reverse(), ...prev].slice(0, 40))
            }
            // Drive animations off this action's events (not a bulk resync replay).
            if (!msg.resync && (msg.events?.length ?? 0) > 0) {
              animSeqRef.current++
              setAnim({ seq: animSeqRef.current, events: msg.events })
            }
          }
          setSnap(msg)
          setTurnSecs(msg.type === 'state' ? (msg.turnSecs ?? 0) : 0)
          setTurnNum(msg.type === 'state' ? msg.turnNum : 0)
          setWinner(null)
          setAttacker(null)
          setSpell(null)
          setHeroPowerArmed(false)
          setSeek(null) // any new snapshot means a paused seek resolved
          setOppSeek(null) // ...including the opponent's, so clear the indicator
          setOppIntent(null) // a resolved action invalidates the opponent's old aim hint
          if (msg.mulligan && !spectatingRef.current) {
            setMyTurn(false)
            setStatus('Mulligan — replace any cards, then keep')
            if (msg.type === 'match_start') {
              // Opening: show a brief "Match found!" splash, then the mulligan.
              setMulliganPicks(new Set())
              setMulliganSubmitted(false)
              setPhase('matchfound')
              window.clearTimeout(matchFoundTimerRef.current)
              matchFoundTimerRef.current = window.setTimeout(() => setPhase('mulligan'), 2000)
            } else {
              // Reconnect / resubmit mid-mulligan: go straight to the mulligan UI.
              setPhase('mulligan')
            }
          } else {
            // First live snapshot after the mulligan: dissolve the blurred
            // mulligan view into the board rather than cutting to it.
            if (phaseRef.current === 'mulligan' || phaseRef.current === 'matchfound') {
              // Capture the mulligan modal's center NOW (DOM still shows it) so the
              // kept cards fly from there into the hand once we render the board.
              const mr = document.querySelector('.mulligan-modal')?.getBoundingClientRect()
              setIntroFrom(mr ? { x: mr.left + mr.width / 2, y: mr.top + mr.height / 2 } : null)
              setIntroPlay(true)
              window.clearTimeout(introTimerRef.current)
              introTimerRef.current = window.setTimeout(() => setIntroPlay(false), 1700)
            }
            setMulliganSubmitted(false)
            setMyTurn(mine)
            setPhase('playing')
            if (spectatingRef.current) {
              setStatus(`spectating ${spectatingRef.current}`)
            } else {
              setStatus(mine ? 'YOUR TURN' : "opponent's turn")
            }
          }
          break
        }
        case 'spectate_start':
          // Switch into the read-only spectator view; the watched player's POV
          // snapshot follows immediately and drives the board render.
          setSpectating(msg.target)
          spectatingRef.current = msg.target
          setLog([])
          setWinner(null)
          break
        case 'spectators':
          setSpectators(msg.names ?? [])
          break
        case 'seek':
          setSeek(msg.options)
          setStatus('Seek — pick a card')
          break
        case 'opp_seek':
          setOppSeek(msg.count)
          break
        case 'opp_intent':
          setOppIntent({ hoverHand: msg.hoverHand, hover: msg.hover, aimFrom: msg.aimFrom, aimTo: msg.aimTo })
          break
        case 'opp_conn':
          setOppOnline(msg.connected)
          break
        case 'invite_received':
          setIncomingInvites((prev) => (prev.includes(msg.from) ? prev : [...prev, msg.from]))
          setInviteDeck(selectedDeckRef.current)
          break
        case 'invite_declined':
          setInvitedName(null)
          setStatus(`${msg.by} declined your invite`)
          break
        case 'invite_cancelled':
          setIncomingInvites((prev) => prev.filter((n) => n !== msg.from))
          break
        case 'game_over':
          setSeek(null)
          setOppSeek(null)
          setOppOnline(true)
          // From the rendered POV: our own when playing, the watched player's when
          // spectating ('you' then means the player we're watching).
          setWinner(msg.winner === povRef.current ? 'you' : 'opponent')
          break
        case 'error':
          // Kicked because the same account logged in elsewhere: this window
          // logs out (clears its token so a reload won't reconnect and bounce
          // the new session).
          if (msg.msg === 'logged in elsewhere') {
            kickedRef.current = true
            localStorage.removeItem(TOKEN_KEY)
            meRef.current = null
            setStatus('logged in from another window')
            setPhase('auth')
            wsRef.current?.close()
            break
          }
          // An error before we ever joined means the (stored) token is bad.
          if (!meRef.current) {
            giveUpRef.current = true // dead token: don't auto-reconnect in a loop
            localStorage.removeItem(TOKEN_KEY)
            tokenRef.current = null
            closeSocket() // drop the dead socket so a fresh login can connect
            setStatus('session expired — log in again')
            setPhase('auth')
          } else {
            // A failed spectate (e.g. the match just ended) leaves us in the lobby:
            // clear the pending spectator state so the lobby renders normally.
            if (spectatingRef.current) {
              spectatingRef.current = null
              setSpectating(null)
            }
            setStatus('error: ' + msg.msg)
          }
          break
      }
    }
  }, [])

  // On load, resume an existing session if we have a stored token.
  useEffect(() => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (token) {
      setPhase('connecting')
      setStatus('reconnecting…')
      connect(token)
    }
  }, [connect])

  const onRegister = async () => {
    try {
      await register(username, password)
      setStatus('registered, now log in')
    } catch (err) {
      setStatus('register failed: ' + (err as Error).message)
    }
  }

  const onLogin = async () => {
    try {
      const token = await login(username, password)
      localStorage.setItem(TOKEN_KEY, token)
      giveUpRef.current = false // fresh session: re-enable auto-reconnect
      retriesRef.current = 0
      closeSocket() // ensure a fresh connection (no lingering socket to early-return on)
      setPhase('connecting')
      setStatus('connecting…')
      connect(token)
    } catch (err) {
      setStatus('login failed: ' + (err as Error).message)
    }
  }

  const onLogout = () => {
    giveUpRef.current = true // intentional disconnect: suppress auto-reconnect
    if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current)
    localStorage.removeItem(TOKEN_KEY)
    tokenRef.current = null
    meRef.current = null
    wsRef.current?.close()
    setPhase('auth')
    setStatus('logged out')
  }

  const onPlay = () => {
    send({ type: 'find_match', deckId: selectedDeck })
  }

  // Play immediately against the AI, which plays a random prebuilt deck of aiClass.
  const onPlayAI = () => {
    send({ type: 'find_match', deckId: selectedDeck, vsAI: true, aiClass, aiDeckId: aiDeck })
  }

  // Direct invites. Only one outgoing at a time; cancel before inviting another.
  const onInvite = (target: string) => {
    setInvitedName(target)
    send({ type: 'invite', target, deckId: selectedDeck })
  }
  const onCancelInvite = () => {
    setInvitedName(null)
    send({ type: 'invite_cancel' })
  }
  const onRespondInvite = (from: string, accept: boolean) => {
    send({ type: 'invite_respond', from, accept, deckId: accept ? inviteDeck : undefined })
    setIncomingInvites((prev) => prev.filter((n) => n !== from))
  }

  // Load the player's saved decks whenever they land in the lobby.
  useEffect(() => {
    if (phase !== 'lobby' || !tokenRef.current) return
    listDecks(tokenRef.current)
      .then((ds) => {
        setDecks(ds)
        // Keep the selection valid; fall back to the first saved deck.
        setSelectedDeck((cur) => (ds.some((d) => d.id === cur) ? cur : (ds[0]?.id ?? 0)))
      })
      .catch(() => setDecks([]))
  }, [phase])

  const onBackToLobby = () => {
    window.clearTimeout(matchFoundTimerRef.current)
    send({ type: 'enter_lobby' })
    setSpectating(null)
    spectatingRef.current = null
    setSpectators([])
    setSnap(null)
    setWinner(null)
    setAttacker(null)
    setSpell(null)
    setHeroPowerArmed(false)
    setSeek(null)
    setOppSeek(null)
    setMulliganPicks(new Set())
    setMulliganSubmitted(false)
    setInvitedName(null)
    setIncomingInvites([])
    setStatus(name ? `logged in as ${name}` : 'in lobby')
    setPhase('lobby')
  }

  // Start spectating a player's live match (read-only). The server replies with
  // spectate_start + the watched player's POV snapshot.
  const onSpectate = (target: string) => send({ type: 'spectate', target })

  // hasLegalTarget mirrors the server's hasLegalTargetFor: is there a character the
  // card can hit right now under its rule AND any extra condition (reqAttack /
  // reqTaunt)? Heroes never satisfy a minion condition.
  const hasLegalTarget = (card: CardView): boolean => {
    if (!snap || !card.target) return false
    const rule = card.target
    const ok = (kind: CharKind, m?: MinionView): boolean => {
      if (kind === 'enemyMinion' && m?.stealth) return false // enemy Stealth untargetable
      return ruleMatches(rule, kind) && condMet(card, m)
    }
    if (ok('selfHero') || ok('oppHero')) return true
    if (snap.self.board.some((m) => ok('friendlyMinion', m))) return true
    if (snap.opp.board.some((m) => ok('enemyMinion', m))) return true
    return false
  }

  // Clicking a hand card. A targeted card (a spell, or a minion with a targeted
  // onset) arms targeting when a legal target exists; click it again to
  // cancel. With no legal target, a minion onset still plays (it fizzles)
  // while a spell cannot be cast. Everything else plays immediately.
  // pos (optional) is the board slot a dragged minion was dropped onto; undefined
  // appends. Ignored for non-minions.
  const onHandCard = (i: number, card: CardView, pos?: number) => {
    if (!myTurn || winner || spectatingRef.current) return
    const minionPos = card.cardType === 'minion' ? pos : undefined
    const rule = card.target
    if (rule && rule !== 'none') {
      if (spell?.handIndex === i) {
        setSpell(null)
        return
      }
      if (hasLegalTarget(card)) {
        setSpell({
          handIndex: i,
          target: rule,
          reqAttack: card.reqAttack,
          reqTaunt: card.reqTaunt,
          pos: minionPos,
        })
        setAttacker(null)
        setHeroPowerArmed(false)
        return
      }
      if (card.cardType === 'minion') {
        send({ type: 'play_card', handIndex: i, pos: minionPos }) // onset fizzles
      } else {
        setStatus('no valid target')
      }
      return
    }
    // Non-minion cards (spell/secret/weapon) fly to the table; minions arrive via
    // the board fly-in instead.
    if (card.cardType !== 'minion') playGhost(`hand-${i}`)
    send({ type: 'play_card', handIndex: i, pos: minionPos })
  }

  // Clicking the hero power. An untargeted power fires immediately; a targeted one
  // (Fire Dart = any character) arms targeting — click again to cancel.
  const onHeroPower = () => {
    if (!myTurn || winner || !snap || spectatingRef.current) return
    const hp = snap.self.heroPower
    if (!hp || snap.self.heroPowerUsed || hp.cost > snap.self.mana) return
    const rule = hp.target
    if (rule && rule !== 'none') {
      setHeroPowerArmed((on) => !on)
      setSpell(null)
      setAttacker(null)
      return
    }
    send({ type: 'hero_power' })
  }

  // True if a character is a legal click target right now. For attacks, Taunt
  // (must hit a taunt minion) and the attacker's hero-reach (Rush) are honored —
  // the server is still authoritative; this only drives highlighting/clicks.
  const targetable = (kind: CharKind, m?: MinionView): boolean => {
    if (!myTurn || winner || !snap || spectatingRef.current) return false
    if (heroPowerArmed) {
      const rule = snap.self.heroPower?.target
      if (!rule) return false
      if (kind === 'enemyMinion' && m?.stealth) return false
      if (m?.elusive) return false // Elusive: untargetable by spells/hero powers
      return ruleMatches(rule, kind)
    }
    if (spell) {
      if (kind === 'enemyMinion' && m?.stealth) return false // enemy Stealth untargetable
      if (m?.elusive) return false // Elusive: untargetable by spells/hero powers
      return ruleMatches(spell.target, kind) && condMet(spell, m)
    }
    if (attacker) {
      // A Stealthed Taunt is hidden, so it does not compel attacks.
      const enemyTaunt = snap.opp.board.some((x) => x.taunt && !x.stealth)
      if (attacker === 'selfHero') {
        // Hero (weapon) attack: may go face unless a taunt is up; no Rush limits.
        if (kind === 'oppHero') return !enemyTaunt
        if (kind === 'enemyMinion') {
          if (m?.stealth) return false
          return enemyTaunt ? !!m?.taunt : true
        }
        return false
      }
      const atk = snap.self.board.find((x) => x.instanceId === attacker)
      if (kind === 'oppHero') return !!atk?.canAttackHero && !enemyTaunt
      if (kind === 'enemyMinion') {
        if (m?.stealth) return false // can't attack a Stealthed minion
        return enemyTaunt ? !!m?.taunt : true
      }
      return false
    }
    return false
  }

  // Clicking a character resolves the active hero power, spell, or attack.
  const onChar = (targetId: string, kind: CharKind, m?: MinionView) => {
    if (!myTurn || winner || spectatingRef.current) return
    if (heroPowerArmed) {
      if (targetable(kind, m)) {
        send({ type: 'hero_power', targetId })
        setHeroPowerArmed(false)
      }
      return
    }
    if (spell) {
      if (targetable(kind, m)) {
        const played = snap?.self.hand?.[spell.handIndex]
        if (played && played.cardType !== 'minion') playGhost(`hand-${spell.handIndex}`)
        send({ type: 'play_card', handIndex: spell.handIndex, targetId, pos: spell.pos })
        setSpell(null)
      }
      return
    }
    if (kind === 'friendlyMinion') {
      if (m && m.canAttack && m.attack > 0) {
        setAttacker((cur) => (cur === targetId ? null : targetId))
      }
      return
    }
    if (kind === 'selfHero') {
      // Select the weapon-armed hero as the attacker (click again to deselect).
      if (snap?.self.heroCanAttack) {
        setAttacker((cur) => (cur === 'selfHero' ? null : 'selfHero'))
      }
      return
    }
    if ((kind === 'enemyMinion' || kind === 'oppHero') && attacker) {
      if (!targetable(kind, m)) return // illegal (taunt / no hero reach)
      send({ type: 'attack', attackerId: attacker, targetId })
      setAttacker(null)
    }
  }

  const ghLink = (
    <a
      className="gh-link"
      href="https://github.com/amvid/vanillastone"
      target="_blank"
      rel="noopener noreferrer"
      title="View source on GitHub"
    >
      <svg viewBox="0 0 16 16" width="20" height="20" aria-hidden="true">
        <path
          fill="currentColor"
          d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82a7.6 7.6 0 014 0c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0016 8c0-4.42-3.58-8-8-8z"
        />
      </svg>
    </a>
  )

  if (phase === 'auth' || phase === 'connecting') {
    const busy = phase === 'connecting'
    return (
      <div className="lobby-screen">
        <div className="lobby-card">
          <h1 className="logo">Vanillastone</h1>
          <p className="welcome">Register once, then log in.</p>

          <input
            className="auth-input"
            placeholder="username"
            value={username}
            disabled={busy}
            onChange={(e) => setUsername(e.target.value)}
          />
          <input
            className="auth-input"
            type="password"
            placeholder="password"
            value={password}
            disabled={busy}
            onChange={(e) => setPassword(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !busy) onLogin()
            }}
          />

          <button className={'play-btn' + (busy ? ' waiting' : '')} onClick={onLogin} disabled={busy}>
            {busy ? (
              <>
                <span className="play-spinner" />
                Connecting
              </>
            ) : (
              '▶ Login'
            )}
          </button>
          <button className="build-btn" onClick={onRegister} disabled={busy}>
            Register
          </button>

          <p className="lobby-status">{status}</p>
          {ghLink}
        </div>
      </div>
    )
  }

  if (phase === 'lobby' || phase === 'waiting') {
    const waiting = phase === 'waiting'
    return (
      <div className="lobby-screen">
        <div className="lobby-card">
          <h1 className="logo">Vanillastone</h1>
          <p className="welcome">
            Welcome, <strong>{name}</strong>
          </p>
          <p className="presence">
            <span className="dot" /> {counts.online} online · {counts.inGame} in game
          </p>

          {/* Play opens the mode picker; while searching (queued vs a human) it
              becomes the cancel toggle back to the lobby. */}
          <button
            className={'play-btn' + (waiting ? ' waiting' : '')}
            onClick={waiting ? onBackToLobby : () => setPlayModal(true)}
          >
            {waiting ? (
              <>
                <span className="play-spinner" />
                Finding opponent
              </>
            ) : (
              '▶ Play'
            )}
          </button>
          {waiting && <p className="cancel-hint">click again to cancel</p>}

          <button className="build-btn" onClick={() => setPhase('deckbuilder')} disabled={waiting}>
            📖 Build deck
          </button>

          <button className="logout" onClick={onLogout} disabled={waiting}>
            Log out
          </button>
          <p className="lobby-status">{status}</p>
          {ghLink}
        </div>

        {/* Mode picker opened by Play: vs AI (pick its class), vs a live player
            (queue), or Arena (reserved). */}
        {playModal && !waiting && (
          <div className="overlay" onClick={() => setPlayModal(false)}>
            <div className="mode-picker" onClick={(e) => e.stopPropagation()}>
              <h2>Choose how to play</h2>

              {/* Your deck — shared by every mode. */}
              <label className="mode-deck">
                <span>Your deck</span>
                <DeckSelect
                  value={selectedDeck}
                  onChange={setSelectedDeck}
                  options={decks}
                />
              </label>

              <div className="mode-grid">
                <div className="mode-card">
                  <span className="mode-icon">🤖</span>
                  <span className="mode-name">Play vs AI</span>
                  <span className="mode-desc">Practice against the computer.</span>
                  <label className="mode-aiclass">
                    <span>AI plays</span>
                    <select value={aiClass} onChange={(e) => setAiClass(e.target.value)}>
                      <option value="mage">Mage</option>
                    </select>
                  </label>
                  <label className="mode-aiclass">
                    <span>AI deck</span>
                    <select value={aiDeck} onChange={(e) => setAiDeck(Number(e.target.value))}>
                      <option value={0}>Random deck</option>
                      {decks.map((d) => (
                        <option key={d.id} value={d.id}>
                          {d.name}
                        </option>
                      ))}
                    </select>
                  </label>
                  <button
                    className="mode-go"
                    onClick={() => {
                      setPlayModal(false)
                      onPlayAI()
                    }}
                  >
                    Start
                  </button>
                </div>

                <div className="mode-card">
                  <span className="mode-icon">⚔️</span>
                  <span className="mode-name">Play vs Player</span>
                  <span className="mode-desc">Queue for a live opponent.</span>
                  <button
                    className="mode-go"
                    onClick={() => {
                      setPlayModal(false)
                      onPlay()
                    }}
                  >
                    Find match
                  </button>
                </div>

                <div className="mode-card disabled">
                  <span className="mode-icon">🏟️</span>
                  <span className="mode-name">Arena</span>
                  <span className="mode-desc">Draft a deck, climb a run.</span>
                  <span className="mode-soon">Coming soon</span>
                </div>
              </div>
              <button className="mode-cancel" onClick={() => setPlayModal(false)}>
                Cancel
              </button>
            </div>
          </div>
        )}

        {/* Online players live in their own fixed, full-height, internally
            scrolling panel — independent of the centered lobby card — so a long
            list (15+ players) never pushes the card's controls off-screen. */}
        {players.length > 0 &&
          (() => {
            const q = playerFilter.trim().toLowerCase()
            const shown = q ? players.filter((p) => p.name.toLowerCase().includes(q)) : players
            return (
              <aside className="player-panel">
                <div className="player-panel-head">
                  <span className="pp-title">Players online</span>
                  <span className="pp-count">{players.length}</span>
                </div>
                <input
                  className="player-search"
                  placeholder="Search players…"
                  value={playerFilter}
                  onChange={(e) => setPlayerFilter(e.target.value)}
                />
                <ul className="player-list">
                  {shown.length === 0 ? (
                    <li className="pl-empty">No players match “{playerFilter}”.</li>
                  ) : (
                    shown.map((p) => (
                      <li key={p.name} className={'pl-row pl-' + p.status}>
                        <span className={'pl-dot ' + p.status} />
                        <span className="pl-name">
                          {p.name}
                          {p.name === name && ' (you)'}
                        </span>
                        {p.name !== name && p.status === 'lobby' ? (
                          invitedName === p.name ? (
                            <button className="pl-invite invited" onClick={onCancelInvite} title="Cancel invite">
                              invited… ✕
                            </button>
                          ) : (
                            <button
                              className="pl-invite"
                              onClick={() => onInvite(p.name)}
                              disabled={invitedName !== null || waiting}
                              title="Invite to a match"
                            >
                              ⚔️ Invite
                            </button>
                          )
                        ) : p.status === 'in_game' ? (
                          <span className="pl-status">
                            ⚔️ vs {p.vs}
                            <button
                              className="pl-invite pl-spectate"
                              onClick={() => onSpectate(p.name)}
                              disabled={waiting}
                              title={`Spectate ${p.name}'s match`}
                            >
                              👁 Watch
                            </button>
                          </span>
                        ) : (
                          <span className="pl-status">{p.status === 'waiting' ? 'searching…' : 'in lobby'}</span>
                        )}
                      </li>
                    ))
                  )}
                </ul>
              </aside>
            )
          })()}

        {incomingInvites.length > 0 && (
          <div className="invite-overlay">
            <div className="invite-modal">
              <h2>⚔️ {incomingInvites.length > 1 ? 'Challenges!' : 'Challenge!'}</h2>
              <label className="deck-pick">
                <span>Your deck</span>
                <select value={inviteDeck} onChange={(e) => setInviteDeck(Number(e.target.value))}>
                  {decks.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.name}
                    </option>
                  ))}
                </select>
              </label>
              <ul className="invite-queue">
                {incomingInvites.map((from) => (
                  <li key={from} className="invite-queue-row">
                    <span className="invite-from">
                      <strong>{from}</strong> invites you
                    </span>
                    <div className="invite-actions">
                      <button className="accept" onClick={() => onRespondInvite(from, true)}>
                        Accept
                      </button>
                      <button className="decline" onClick={() => onRespondInvite(from, false)}>
                        Decline
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        )}
      </div>
    )
  }

  if (phase === 'deckbuilder') {
    return <Deckbuilder token={tokenRef.current ?? ''} onBack={() => setPhase('lobby')} />
  }

  if (!snap) {
    return (
      <div>
        <h1>Vanillastone</h1>
        <p>{status}</p>
        <button onClick={onBackToLobby}>Back to lobby</button>
      </div>
    )
  }

  const hint =
    heroPowerArmed || spell
      ? ' — pick a target'
      : attacker
        ? ' — pick an attack target'
        : ''

  // The game board. Rendered on its own during play, and (blurred) behind the
  // mulligan overlay so the table is visible while you mulligan.
  const board = (
    <GameScreen
      snap={snap}
      name={name}
      myTurn={myTurn}
      turnSecs={turnSecs}
      turnNum={turnNum}
      winner={winner}
      attacker={attacker}
      spell={spell}
      heroPowerArmed={heroPowerArmed}
      seek={seek}
      oppSeek={oppSeek}
      oppIntent={oppIntent}
      log={log}
      anim={anim}
      oppOnline={oppOnline}
      status={status}
      hint={hint}
      send={send}
      onBackToLobby={onBackToLobby}
      onChar={onChar}
      targetable={targetable}
      onHandCard={onHandCard}
      onHeroPower={onHeroPower}
      intro={introPlay}
      introFrom={introFrom}
      spectating={spectating}
      spectators={spectators}
    />
  )

  if (phase === 'matchfound') {
    return (
      <div className="lobby-screen">
        <div className="matchfound">
          <div className="mf-swords">⚔️</div>
          <div className="mf-title">Match found!</div>
          <div className="mf-vs">
            <strong>{name}</strong> vs <strong>{snap?.opp.name || 'Opponent'}</strong>
          </div>
          <div className="play-spinner mf-spinner" />
        </div>
      </div>
    )
  }

  if (phase === 'mulligan') {
    const toggle = (i: number) =>
      setMulliganPicks((prev) => {
        const next = new Set(prev)
        if (next.has(i)) next.delete(i)
        else next.add(i)
        return next
      })
    const submit = () => {
      send({ type: 'mulligan', indices: [...mulliganPicks] })
      setMulliganSubmitted(true)
      setStatus('waiting for opponent…')
    }
    const hand = snap.self.hand ?? []
    return (
      <>
        <div className="mulligan-bg">{board}</div>
        <div className="overlay">
          <div className="mulligan-modal">
            <div className="seek-title">Mulligan — replace any cards, then keep</div>
            {!oppOnline && (
              <div className="banner warn">⚠ Opponent disconnected — waiting for them to reconnect…</div>
            )}
            {mulliganSubmitted ? (
              <p>Cards locked in. Waiting for your opponent…</p>
            ) : (
              <>
                <div className="hand mulligan">
                  {hand.map((c, i) => (
                    <button
                      key={i}
                      className={'card' + cardColorClass(c) + (mulliganPicks.has(i) ? ' tossed' : '')}
                      onClick={() => toggle(i)}
                    >
                      <CardFace card={c} />
                    </button>
                  ))}
                </div>
                <div className={'mulligan-timer' + (mulliganLeft <= 5 ? ' low' : '')}>⏳ {mulliganLeft}s</div>
                <button className="keep-btn" onClick={submit}>
                  {mulliganPicks.size === 0 ? 'Keep all' : `Replace ${mulliganPicks.size}`}
                </button>
              </>
            )}
          </div>
        </div>
      </>
    )
  }

  return board
}
