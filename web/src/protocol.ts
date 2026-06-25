// Wire types mirroring the Go server's internal/protocol package. The server is
// authoritative; these describe the messages, not any client-side game logic.

export type TargetRule =
  | 'none'
  | 'any'
  | 'minion'
  | 'friendlyMinion'
  | 'enemyMinion'
  | 'enemy'
  | 'friendlyHero'
  | 'hero'

export type CardView = {
  cardId: string
  name: string
  cardType: 'minion' | 'spell' | 'secret' | 'weapon' | 'heroPower'
  class?: 'neutral' | 'mage' // drives card color (mage = blue); absent/neutral = default
  rarity?: 'common' | 'rare' | 'epic' | 'legendary' // drives the rarity gem
  cost: number
  baseCost?: number // hand cards: printed cost (present so the client can colour a cost-modified card)
  attack: number
  health: number
  durability?: number // weapons
  tribe?: string // minion creature type (shown in the type band)
  target?: TargetRule // present for spells
  reqAttack?: number // targeted onset: target minion must have attack >= this
  reqTaunt?: boolean // targeted onset: target minion must have Taunt
  reqTribe?: string // targeted onset: target minion must be of this tribe
  text?: string // rules text for the hover box
}

export type MinionView = {
  rarity?: 'common' | 'rare' | 'epic' | 'legendary'
  instanceId: string
  cardId: string
  name: string
  class?: 'neutral' | 'mage' // drives card color in the hover preview
  cost: number // printed mana cost (for the full-card hover preview)
  attack: number
  health: number
  maxHealth: number
  baseAttack: number // printed attack (for buff/debuff coloring)
  baseHealth: number // printed health (for buff/damage coloring)
  canAttack: boolean // able to attack something this turn
  canAttackHero: boolean // may also go face (false for Rush on its summon turn)
  taunt?: boolean
  aegis?: boolean
  frozen?: boolean
  twinstrike?: boolean
  stealth?: boolean
  poisonous?: boolean
  lifesteal?: boolean
  finalGasp?: boolean
  spellDamage?: number
  enraged?: boolean
  hasEnrage?: boolean
  silenced?: boolean
  elusive?: boolean
  cantAttack?: boolean
  tribe?: string
  text?: string // rules text for the hover box
}

export type PlayerView = {
  name?: string // display username
  heroHP: number
  armor?: number // absorbs damage before health
  frozen?: boolean // hero frozen (blocks weapon attacks)
  immune?: boolean // hero ignores all damage this turn (Ice Block)
  mana: number
  maxMana: number
  board: MinionView[]
  hand?: CardView[] // present only for your own side
  handCount: number
  deckCount: number // cards left in the draw pile (public)
  secrets?: CardView[] // own side only; opponent sees secretCount
  secretCount?: number
  heroPower?: CardView // the hero's reusable ability (public)
  heroPowerUsed?: boolean
  weapon?: WeaponView // equipped weapon (public)
  heroAttack?: number // current hero attack (weapon attack)
  heroCanAttack?: boolean // hero may attack right now
}

export type WeaponView = {
  name: string
  attack: number
  durability: number
  text?: string
}

// Event is one ordered step of an action's resolution (mirrors Go
// protocol.Event). Source/Target are absolute ids: a minion instanceId, or a
// player id (p1/p2) for a hero.
export type Event = {
  kind:
    | 'play'
    | 'onset'
    | 'finalGasp'
    | 'trigger'
    | 'damage'
    | 'heal'
    | 'buff'
    | 'summon'
    | 'death'
    | 'attack'
    | 'freeze'
    | 'shield'
    | 'silence'
    | 'secret'
    | 'equip'
    | 'weaponBreak'
    | 'heropower'
    | 'armor'
    | 'mana'
    | 'fatigue'
    | 'burn'
    | 'generate'
    | 'destroy'
    | 'transform'
    | 'bounce'
    | 'sethealth'
    | 'control'
    | 'startgame'
  source?: string
  target?: string
  amount?: number
  buffAtk?: number // 'buff' event stat delta, for the log's before→after
  buffHp?: number
  name?: string
  note?: string // secret reveal: the card the secret acted on (e.g. countered spell)
  card?: CardView // set on 'play' events (reveals an opponent's cast)
}

export type Snapshot = {
  turn: string
  you: string
  self: PlayerView
  opp: PlayerView
}

export type ClientMessage =
  | { type: 'auth'; token: string }
  | { type: 'find_match'; deckId?: number; vsAI?: boolean; aiClass?: string; aiDeckId?: number }
  | { type: 'enter_lobby' }
  | { type: 'end_turn' }
  | { type: 'play_card'; handIndex: number; targetId?: string; pos?: number }
  | { type: 'attack'; attackerId: string; targetId: string }
  | { type: 'concede' }
  | { type: 'choose'; index: number }
  | { type: 'hero_power'; targetId?: string }
  | { type: 'mulligan'; indices: number[] }
  | { type: 'invite'; target: string; deckId?: number }
  | { type: 'invite_cancel' }
  | { type: 'invite_respond'; from: string; accept: boolean; deckId?: number }
  | { type: 'spectate'; target: string } // watch a player's live match

// PlayerInfo is one online player in the lobby list. status: lobby | waiting |
// in_game. For in_game, vs is the opponent and matchId is reserved for spectating.
export type PlayerInfo = {
  name: string
  status: 'lobby' | 'waiting' | 'in_game'
  vs?: string
  matchId?: string
}

export type ServerMessage =
  | { type: 'joined'; you: string; name: string }
  | { type: 'lobby'; online: number; inGame: number; players?: PlayerInfo[] }
  | { type: 'waiting' }
  | ({ type: 'match_start'; players: string[]; mulligan?: boolean } & Snapshot)
  | ({
      type: 'state'
      turnNum: number
      events: Event[]
      mulligan?: boolean
      resync?: boolean
      turnSecs?: number // seconds left in the current turn
    } & Snapshot)
  | { type: 'seek'; options: CardView[] }
  | { type: 'opp_seek'; count: number } // opponent is choosing a Seek card
  | ({ type: 'opp_intent' } & OppIntent) // ephemeral: what the opponent is hovering / aiming at
  | { type: 'game_over'; winner: string }
  | { type: 'opp_conn'; connected: boolean } // opponent dropped / reconnected
  | { type: 'invite_received'; from: string } // someone challenged you
  | { type: 'invite_declined'; by: string } // your invite was refused / dropped
  | { type: 'invite_cancelled'; from: string } // an incoming invite was withdrawn
  | { type: 'spectate_start'; target: string } // now spectating target's match (read-only)
  | { type: 'spectators'; names: string[] } // who is watching your match (players only)
  | { type: 'error'; msg: string }

// OppIntent is the ephemeral, non-authoritative aiming hint the server relays from
// the acting player (mirrors the Go `protocol.OppIntent`). All identifiers are sent
// from the SENDER's perspective: `hover`/`aimFrom`/`aimTo` use minion instance ids
// for board minions, perspective-free role tokens `self`/`enemy` for heroes,
// `heroPower` for the hero power, and `hand:<i>` for a hand card being aimed. The
// receiver flips `self`/`enemy` (the sender's enemy is the viewer) when rendering.
export type OppIntent = {
  hoverHand: number // index of the hand card the opponent is holding; -1 = none
  hover?: string // a board minion the opponent is inspecting
  aimFrom?: string // source of an in-progress aim
  aimTo?: string // character currently under the aim
}
