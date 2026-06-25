import type { CardView, Event, MinionView, TargetRule } from '../protocol'
import type { CharKind, LogActor, LogEntry, LogTarget } from './types'

// entName resolves an event id to a display name. Player ids (p…) are heroes;
// minion uids (u…) look up the accumulated name map, falling back to the id.
export function entName(
  id: string | undefined,
  names: Record<string, string>,
  me: string | null,
): string {
  if (!id) return '?'
  if (id[0] === 'p') return id === me ? 'Your hero' : 'Enemy hero'
  return names[id] ?? id
}

// formatEvent turns one resolution event into a human log line.
export function formatEvent(e: Event, names: Record<string, string>, me: string | null): string {
  const ent = (id?: string) => entName(id, names, me)
  switch (e.kind) {
    case 'play':
      return `${ent(e.source)} plays ${e.name ?? e.card?.name ?? 'a card'}`
    case 'attack':
      return `${ent(e.source)} attacks ${ent(e.target)}`
    case 'damage':
      return `${ent(e.target)} takes ${e.amount} damage`
    case 'heal':
      return `${ent(e.target)} heals ${e.amount}`
    case 'buff':
      return `${ent(e.target)} is buffed`
    case 'summon':
      return `${ent(e.source)} summons ${e.name ?? ent(e.target)}`
    case 'onset':
      return `${e.name ?? ent(e.source)}: Onset`
    case 'finalGasp':
      return `${e.name ?? ent(e.source)}: Final Gasp`
    case 'trigger':
      return `${e.name ?? ent(e.source)} triggers`
    case 'death':
      return `${e.name ?? ent(e.target)} dies`
    case 'freeze':
      return `${e.name ?? ent(e.target)} is frozen`
    case 'shield':
      return `${e.name ?? ent(e.target)}'s Aegis pops`
    case 'silence':
      return `${e.name ?? ent(e.target)} is silenced`
    case 'secret':
      return e.note
        ? `${e.name} counters ${e.note}`
        : `Secret revealed — ${e.name} triggers`
    case 'equip':
      return `${ent(e.source)} equips ${e.name}`
    case 'weaponBreak':
      return `${ent(e.source)}'s ${e.name} breaks`
    case 'heropower':
      return `${ent(e.source)} uses ${e.name}`
    case 'armor':
      return `${ent(e.target)} gains ${e.amount} armor`
    case 'mana':
      return `${ent(e.target)} gains a Mana Crystal this turn`
    case 'fatigue':
      return `${ent(e.target)} takes ${e.amount} fatigue damage`
    case 'burn':
      return `${ent(e.target)} burns ${e.card ? e.card.name : 'a card'} (hand full)`
    case 'generate':
      return `${ent(e.target)} adds a card to hand`
    case 'destroy':
      return `${e.name ?? ent(e.target)} is destroyed`
    case 'transform':
      return `${ent(e.target)} is transformed into ${e.name}`
    case 'control':
      return `${ent(e.source)} takes control of ${e.name ?? ent(e.target)}`
    case 'bounce':
      return `${e.name ?? ent(e.target)} is returned to hand`
    case 'sethealth':
      return `${ent(e.target)}'s Health is set to ${e.amount}`
    default:
      return e.kind
  }
}

// minionToCardView renders a live minion as a full card (for the log's hover
// popup), using its current cost/attack/health.
export function minionToCardView(mn: MinionView): CardView {
  return {
    cardId: mn.cardId,
    name: mn.name,
    cardType: 'minion',
    class: mn.class,
    rarity: mn.rarity,
    cost: mn.cost,
    attack: mn.attack,
    health: mn.health,
    tribe: mn.tribe,
    text: mn.text,
  }
}

// actorFor resolves an event id to a log participant: a hero portrait (p…) or a
// card (u… minion uid, full CardView from the cache for the hover popup; name +
// cardId fall back to the name map for a minion that has already left the board).
function actorFor(
  id: string | undefined,
  names: Record<string, string>,
  cards: Record<string, CardView>,
  me: string | null,
): LogActor | undefined {
  if (!id) return undefined
  // Minion uids start with 'u' (server `uid()`); every other id is a hero/player
  // id (a human's session id, or the bot's `ai:…`) — never assume a 'p' prefix.
  if (id[0] !== 'u') return { name: id === me ? 'Your hero' : 'Enemy hero', hero: id === me ? 'self' : 'opp' }
  const card = cards[id]
  return { name: card?.name ?? names[id] ?? id, uid: id, cardId: card?.cardId, card }
}

// buildLog turns one action's ordered events into grouped HS-style rows. Each row is
// ONE cause (a play / attack / hero power / onset / finalGasp / trigger / secret)
// and EVERY character it affected: an AoE (e.g. an end-of-turn ping that hits five
// minions) collapses to a single row — the source, then all targets with their
// outcomes — instead of one line per hit. A `death` event folds into the target it
// killed (looked up by uid across the whole action) so a kill shows as a ☠ badge on
// that target rather than a separate, often mis-ordered, "dies" line. Melee/hero-power
// impact damage still folds into its cause via the `causeOf` pairing the animator uses.
export function buildLog(
  events: Event[],
  names: Record<string, string>,
  cards: Record<string, CardView>,
  me: string | null,
): LogEntry[] {
  const causeOf: (number | null)[] = events.map(() => null)
  for (let i = 0; i < events.length; i++) {
    const k = events[i].kind
    if (k !== 'attack' && k !== 'heropower') continue
    for (let j = i + 1; j < events.length; j++) {
      if (events[j].kind === 'damage' || events[j].kind === 'shield') causeOf[j] = i
      else break
    }
  }
  const A = (id?: string) => actorFor(id, names, cards, me)
  const rows: LogEntry[] = []
  let group: LogEntry | undefined // the cause currently collecting affected targets
  let playedPending = false // a minion was just played; its own self-summon is redundant with the play row
  // uid → the last damage/destroy/sethealth target object, so a later `death` event
  // (emitted in a batch after all damage) marks the right target as killed — even when
  // it died in an earlier group than the one currently open (e.g. a finalGasp group).
  const lethalTarget: Record<string, LogTarget> = {}

  const open = (g: LogEntry): LogEntry => {
    rows.push(g)
    group = g
    return g
  }
  // Effect events (damage/heal/buff/…) attach to the current cause; a stray one with no
  // open group starts its own bare group keyed on its own source.
  const cur = (e: Event): LogEntry => group ?? open({ kind: e.kind, text: formatEvent(e, names, me), source: A(e.source), targets: [] })
  const add = (e: Event, t: LogTarget): LogTarget => {
    cur(e).targets.push(t)
    return t
  }

  events.forEach((e, i) => {
    const text = formatEvent(e, names, me)
    if (causeOf[i] !== null) return // an impact hit — folded into its attack/hero-power row below
    switch (e.kind) {
      case 'play': {
        const card: LogActor = e.card
          ? { name: e.card.name, cardId: e.card.cardId, card: e.card }
          : (A(e.source) ?? { name: e.name ?? 'a card' })
        playedPending = e.card?.cardType === 'minion' // its self-summon follows; skip that target
        open({ kind: e.kind, text, source: card, targets: [] })
        return
      }
      case 'attack':
      case 'heropower': {
        const g = open({ kind: e.kind, text, source: A(e.source), note: e.kind === 'heropower' ? e.name : undefined, targets: [] })
        // Fold every impact damage/shield this attack/hero-power caused (defender hit +
        // any retaliation on the attacker) into this one row as targets.
        events.forEach((d, j) => {
          if (causeOf[j] !== i) return
          const t: LogTarget = { actor: A(d.target), kind: d.kind, amount: d.amount }
          g.targets.push(t)
          if (d.target && (d.kind === 'damage' || d.kind === 'shield')) lethalTarget[d.target] = t
        })
        return
      }
      case 'onset':
        // The onset of the just-played minion: keep its effects in the play row.
        if (group) group.note = 'Onset'
        else open({ kind: e.kind, text, source: A(e.source), note: 'Onset', targets: [] })
        playedPending = false // later summons are onset tokens — keep them
        return
      case 'finalGasp':
        open({ kind: e.kind, text, source: A(e.source), note: 'Final Gasp', targets: [] })
        return
      case 'trigger':
        open({ kind: e.kind, text, source: A(e.source), note: 'Trigger', targets: [] })
        return
      case 'secret':
        open({
          kind: e.kind,
          text,
          source: e.card ? { name: e.card.name, cardId: e.card.cardId, card: e.card } : { name: e.name ?? 'Secret' },
          note: 'Secret',
          targets: [],
        })
        return
      case 'damage': {
        const t = add(e, { actor: A(e.target), kind: e.kind, amount: e.amount })
        if (e.target) lethalTarget[e.target] = t
        return
      }
      case 'shield': {
        const t = add(e, { actor: A(e.target), kind: e.kind })
        if (e.target) lethalTarget[e.target] = t
        return
      }
      case 'summon':
        if (playedPending) {
          playedPending = false // the played minion entering — the play row already shows it
          return
        }
        add(e, { actor: A(e.target), kind: e.kind })
        return
      case 'heal':
        add(e, { actor: A(e.target), kind: e.kind, amount: e.amount })
        return
      case 'buff': {
        // The cache holds the AFTER stats (it was refreshed from this snapshot before
        // buildLog ran); subtract the delta to recover the BEFORE card so the popup can
        // show before → after (e.g. 5/5 → 7/7), like it did pre-grouping.
        const after = A(e.target)
        const da = e.buffAtk ?? 0
        const dh = e.buffHp ?? 0
        const before: LogActor | undefined =
          after?.card && (da || dh)
            ? { ...after, card: { ...after.card, attack: after.card.attack - da, health: after.card.health - dh } }
            : undefined
        add(e, { actor: after, kind: e.kind, buffAtk: e.buffAtk, buffHp: e.buffHp, before })
        return
      }
      case 'destroy': {
        const t = add(e, { actor: A(e.target), kind: e.kind, note: e.name, died: true })
        if (e.target) lethalTarget[e.target] = t // a following `death` for the same uid folds in
        return
      }
      case 'sethealth': {
        const t = add(e, { actor: A(e.target), kind: e.kind, amount: e.amount })
        if (e.target) lethalTarget[e.target] = t
        return
      }
      case 'freeze':
      case 'silence':
      case 'transform':
      case 'bounce':
      case 'control':
        add(e, { actor: A(e.target), kind: e.kind, note: e.name })
        return
      case 'death': {
        // Fold into the target the killing blow already created; else a bare death row.
        const hit = e.target ? lethalTarget[e.target] : undefined
        if (hit) hit.died = true
        else add(e, { actor: A(e.target), kind: e.kind, died: true })
        return
      }
      case 'startgame':
        return // revealed center-stage (cast showcase), not as a log row
      case 'equip':
      case 'weaponBreak':
        open({ kind: e.kind, text, source: A(e.source), targets: [] })
        return
      case 'burn':
        // Show the hero whose hand burned + the destroyed card's face (now revealed).
        open({
          kind: e.kind,
          text,
          source: A(e.target),
          targets: e.card
            ? [{ actor: { name: e.card.name, cardId: e.card.cardId, card: e.card }, kind: e.kind }]
            : [],
        })
        return
      default:
        // armor / mana / fatigue / burn / generate — hero-centric, one target.
        open({ kind: e.kind, text, source: undefined, targets: [{ actor: A(e.target), kind: e.kind, amount: e.amount }] })
    }
  })
  return rows
}

// kindIcon maps an event kind to a glyph for the compact event-log feed.
export function kindIcon(kind: Event['kind']): string {
  switch (kind) {
    case 'play':
      return '🃏'
    case 'attack':
      return '⚔️'
    case 'damage':
      return '💥'
    case 'heal':
      return '➕'
    case 'buff':
      return '⬆️'
    case 'summon':
      return '✨'
    case 'death':
      return '💀'
    case 'onset':
      return '📣'
    case 'finalGasp':
      return '☠️'
    case 'trigger':
      return '⚡'
    case 'freeze':
      return '❄️'
    case 'shield':
      return '🔆'
    case 'silence':
      return '🤫'
    case 'secret':
      return '❓'
    case 'equip':
      return '🗡️'
    case 'weaponBreak':
      return '💔'
    case 'heropower':
      return '✴️'
    case 'armor':
      return '🛡️'
    case 'mana':
      return '💎'
    case 'fatigue':
      return '🩸'
    case 'burn':
      return '🔥'
    case 'generate':
      return '🪄'
    case 'destroy':
      return '💀'
    case 'transform':
      return '🔄'
    case 'control':
      return '🧠'
    case 'bounce':
      return '↩️'
    case 'sethealth':
      return '❤️'
    default:
      return '•'
  }
}

// cardArtIcon is the placeholder art glyph shown in a card's center, by type.
export function cardArtIcon(cardType: string): string {
  switch (cardType) {
    case 'minion':
      return '🐾'
    case 'spell':
      return '🔮'
    case 'secret':
      return '❓'
    case 'weapon':
      return '⚔️'
    default:
      return '◆'
  }
}

// hpIcon picks a glyph for a hero power by name (Mage = Fire Dart for now).
export function hpIcon(name: string): string {
  if (/fire/i.test(name)) return '🔥'
  return '✴️'
}

// ruleMatches reports whether a character of the given kind is a legal target
// for a spell with the given rule (mirrors the server's validTarget).
export function ruleMatches(rule: TargetRule, kind: CharKind): boolean {
  switch (rule) {
    case 'any':
      return true
    case 'minion':
      return kind === 'friendlyMinion' || kind === 'enemyMinion'
    case 'friendlyMinion':
      return kind === 'friendlyMinion'
    case 'enemyMinion':
      return kind === 'enemyMinion'
    case 'enemy':
      return kind === 'enemyMinion' || kind === 'oppHero'
    case 'friendlyHero':
      return kind === 'selfHero'
    case 'hero':
      return kind === 'selfHero' || kind === 'oppHero'
    default:
      return false
  }
}

// condMet reports whether a minion satisfies a targeted onset's extra
// conditions (reqAttack: attack >= N; reqTaunt: has Taunt). Heroes never satisfy
// a minion condition. Mirrors the server's targetCondOK.
export function condMet(
  cond: { reqAttack?: number; reqTaunt?: boolean; reqTribe?: string },
  m?: MinionView,
): boolean {
  if (cond.reqAttack && (!m || m.attack < cond.reqAttack)) return false
  if (cond.reqTaunt && (!m || !m.taunt)) return false
  if (cond.reqTribe && (!m || m.tribe !== cond.reqTribe)) return false
  return true
}

// cardDesc formats rules text for the card body: drops periods, turning each
// sentence break into a line break (e.g. "Deal 2 damage to a character. Lifesteal."
// → "Deal 2 damage to a character⏎Lifesteal").
export function cardDesc(text: string): string {
  return text
    .replace(/\.\s+/g, '\n')
    .replace(/\.\s*$/, '')
    .trim()
}

// tribeLabel capitalizes a tribe id ("beast" -> "Beast"), or "" if none.
export function tribeLabel(tribe?: string): string {
  if (!tribe) return ''
  return tribe.charAt(0).toUpperCase() + tribe.slice(1)
}

// cardTypeLabel is the type/tribe shown in a card's bottom band. A minion with a
// tribe shows its tribe in place of "Minion" (e.g. "Beast", "Dragon"), matching
// the genre convention; an untribed minion shows "Minion".
export function cardTypeLabel(c: CardView): string {
  switch (c.cardType) {
    case 'minion':
      return tribeLabel(c.tribe) || 'Minion'
    case 'spell':
      return 'Spell'
    case 'secret':
      return 'Secret'
    case 'weapon':
      return 'Weapon'
    case 'heroPower':
      return 'Hero Power'
    default:
      return c.cardType
  }
}

// cardColorClass returns the CSS modifier that colors a card by CLASS (not by
// type): Mage cards are blue, neutral cards keep the default parchment frame.
export function cardColorClass(c: CardView): string {
  return c.class === 'mage' ? ' mage' : ''
}

// cardStats is the small stat string shown on a card: minions a/h, weapons
// a/durability, otherwise the card type.
export function cardStats(c: CardView): string {
  if (c.cardType === 'minion') return `${c.attack}/${c.health}`
  if (c.cardType === 'weapon') return `${c.attack}/${c.durability ?? 0}`
  return c.cardType
}
