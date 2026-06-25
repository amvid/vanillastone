import type { CardView, Event, TargetRule } from '../protocol'

export type Phase =
  | 'auth'
  | 'connecting'
  | 'lobby'
  | 'deckbuilder'
  | 'waiting'
  | 'matchfound'
  | 'mulligan'
  | 'playing'

export type CharKind = 'friendlyMinion' | 'enemyMinion' | 'selfHero' | 'oppHero'

// A targeted card awaiting its target. For a minion with a targeted onset,
// pos remembers the board slot it was dragged onto (undefined = append).
export type PendingSpell = {
  handIndex: number
  target: TargetRule
  reqAttack?: number // target minion must have attack >= this
  reqTaunt?: boolean // target minion must have Taunt
  pos?: number
}

export type Counts = { online: number; inGame: number }

// One participant in a log row: a card or a hero portrait. `cardId` drives the
// minimized art chip in the narrow feed; `card` (the full CardView) renders the
// real card face in the on-hover popup.
// `uid` is the board instance id (server `uid()` / CardView.instanceId) — the ONLY
// safe identity for "same character" tests; `cardId` is the card-definition id, so two
// different minions of the same card share it (an attack between same-named minions
// must NOT collapse them).
export type LogActor = { name: string; uid?: string; cardId?: string; hero?: 'self' | 'opp'; card?: CardView }

// One affected character inside a grouped log entry: a card/hero plus what happened
// to it. `kind` is the outcome (damage/heal/buff/freeze/…); `died` marks a target the
// group's effect killed (so a death folds into the row that caused it, not a separate
// line). `note` carries an extra label (transformed-into / returned / mind-controlled).
export type LogTarget = {
  actor?: LogActor
  kind: Event['kind']
  amount?: number
  buffAtk?: number
  buffHp?: number
  before?: LogActor // for a buff: the same card at its pre-buff stats, shown as before → after
  died?: boolean
  note?: string
}

// One event-log entry, HS-style: ONE acting source (card/hero) and every character it
// affected this step, grouped together — so an AoE reads as a single row (the source,
// then all targets with their outcomes incl. deaths) instead of one line per hit. A
// keyword note (Onset/Final Gasp/…) tags the cause. `text` is the plain-language
// headline kept for the hover tip / narrow-feed fallback.
export type LogEntry = {
  kind: Event['kind']
  text: string
  source?: LogActor
  note?: string
  targets: LogTarget[]
}
