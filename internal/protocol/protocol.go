// Package protocol defines the JSON messages exchanged between the web client
// and the authoritative server. Every message is a JSON object with a "type"
// field; the server decodes "type" first, then the concrete payload.
package protocol

import "encoding/json"

// --- Client -> server ---

// Envelope is used to peek at the message type before full decode.
type Envelope struct {
	Type string `json:"type"`
}

// Message types (client -> server).
const (
	TypeAuth       = "auth"
	TypeFindMatch  = "find_match"
	TypeEnterLobby = "enter_lobby"
	TypeEndTurn    = "end_turn"
	TypePlayCard   = "play_card"
	TypeAttack     = "attack"
	TypeConcede    = "concede"
	TypeChoose     = "choose"
	TypeHeroPower  = "hero_power"
	TypeMulligan   = "mulligan"

	TypeInvite        = "invite"
	TypeInviteCancel  = "invite_cancel"
	TypeInviteRespond = "invite_respond"

	TypeSpectate = "spectate"

	TypeIntent = "intent"
)

// Intent is an EPHEMERAL, non-authoritative aiming hint sent by the player who is
// acting: which hand card they're holding, which board minion they're inspecting,
// and where they're aiming an in-progress play/attack. The server relays it to the
// viewers who see that player as the opponent (the other seat + its spectators) so
// they get HS-style "what is my opponent about to do" feedback. NEVER stored, never
// logged, never affects game truth. Sent throttled by the client; cleared by sending
// a neutral intent (HoverHand -1, empty strings) on pointer-up / turn end.
type Intent struct {
	Type string `json:"type"`
	// HoverHand is the index of the hand card the player is holding/considering;
	// -1 = none. (No omitempty: index 0 is valid, so the zero value must transmit.)
	HoverHand int `json:"hoverHand"`
	// Hover is a board minion instance id the player is merely inspecting (mouse
	// over, not aiming); "" = none.
	Hover string `json:"hover,omitempty"`
	// AimFrom / AimTo describe an in-progress targeting line. AimFrom is the source:
	// "hand:<i>" (a card being aimed), a minion instance id (an attacker), "heroPower",
	// or a hero player id. AimTo is the character currently under the aim (a minion
	// instance id or hero player id). Empty = no active aim.
	AimFrom string `json:"aimFrom,omitempty"`
	AimTo   string `json:"aimTo,omitempty"`
}

// Spectate asks to watch a player's live match from that player's point of view
// (their hand revealed, the opponent's hidden). Target is the player's username.
// Valid only when the requester is not in a match of their own. Stop spectating
// with EnterLobby.
type Spectate struct {
	Type   string `json:"type"`
	Target string `json:"target"`
}

// Invite challenges another lobby player to a match. Target is their username;
// DeckID is the inviter's chosen deck (locked at send time, 0 = default). Only
// one outstanding invite per player; the server rejects a second.
type Invite struct {
	Type   string `json:"type"`
	Target string `json:"target"`
	DeckID int64  `json:"deckId,omitempty"`
}

// InviteCancel withdraws the inviter's current outstanding invite.
type InviteCancel struct {
	Type string `json:"type"`
}

// InviteRespond answers an incoming invite from From. Accept starts the match;
// DeckID is the responder's chosen deck (0 = default).
type InviteRespond struct {
	Type   string `json:"type"`
	From   string `json:"from"`
	Accept bool   `json:"accept"`
	DeckID int64  `json:"deckId,omitempty"`
}

// Auth authenticates the connection with a session token (from POST /login).
// On success the player lands in the lobby (it does NOT auto-queue; the client
// sends FindMatch to enter matchmaking).
type Auth struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// FindMatch enters the matchmaking queue. Valid only once authenticated and not
// already in an active match. DeckID selects which saved deck to play; 0 means
// "use a default deck" (e.g. the player has built none). When VsAI is set the
// player is matched immediately against an AI opponent (instead of being queued
// against another human). AIDeckID selects one of the player's own saved decks
// for the bot to play; 0 means the bot plays a random prebuilt deck of AIClass.
type FindMatch struct {
	Type     string `json:"type"`
	DeckID   int64  `json:"deckId,omitempty"`
	VsAI     bool   `json:"vsAI,omitempty"`
	AIClass  string `json:"aiClass,omitempty"`
	AIDeckID int64  `json:"aiDeckId,omitempty"`
}

// EnterLobby returns the player to the lobby: it leaves the queue / abandons a
// finished match and asks the server to refresh the lobby presence counts.
type EnterLobby struct {
	Type string `json:"type"`
}

// EndTurn passes the turn to the opponent. Only valid for the player whose
// turn it currently is; the server rejects it otherwise.
type EndTurn struct {
	Type string `json:"type"`
}

// PlayCard plays the card at HandIndex. For minions, TargetID is ignored. For
// targeted spells, TargetID is the chosen character ("selfHero", "oppHero", or a
// minion instance id); untargeted spells leave it empty.
type PlayCard struct {
	Type      string `json:"type"`
	HandIndex int    `json:"handIndex"`
	TargetID  string `json:"targetId,omitempty"`
	// Pos is the board position (0-based) to insert a played minion at, letting
	// the player drag a card between minions on the table. nil = append to the
	// end (back-compat / non-minion plays). Clamped server-side.
	Pos *int `json:"pos,omitempty"`
}

// Attack orders the friendly minion AttackerID to attack a target. TargetID is
// "oppHero" for the opponent's hero, otherwise the opponent minion's instance id.
// AttackerID may also be "selfHero" — a weapon-armed hero attack.
type Attack struct {
	Type       string `json:"type"`
	AttackerID string `json:"attackerId"`
	TargetID   string `json:"targetId"`
}

// HeroPower uses the player's hero power against TargetID (a character id for a
// targeted power like Fire Dart; empty for untargeted ones). Once per turn.
type HeroPower struct {
	Type     string `json:"type"`
	TargetID string `json:"targetId,omitempty"`
}

// Concede forfeits the match: the conceding player loses immediately. Valid on
// either player's turn.
type Concede struct {
	Type string `json:"type"`
}

// Choose answers a pending Seek prompt by picking option Index. Valid only
// while the server has sent this player a Seek and is awaiting a choice.
type Choose struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// Mulligan submits the opening-hand replacement during the mulligan phase.
// Indices points at the opening-hand cards to toss; an empty list keeps the
// whole hand. Each player submits once.
type Mulligan struct {
	Type    string `json:"type"`
	Indices []int  `json:"indices"`
}

// --- Server -> client ---

// Message types (server -> client).
const (
	TypeJoined     = "joined"
	TypeLobby      = "lobby"
	TypeWaiting    = "waiting"
	TypeMatchStart = "match_start"
	TypeState      = "state"
	TypeSeek       = "seek"
	TypeOppSeek    = "opp_seek"
	TypeGameOver   = "game_over"
	TypeOppConn    = "opp_conn"
	TypeError      = "error"

	TypeInviteReceived  = "invite_received"
	TypeInviteDeclined  = "invite_declined"
	TypeInviteCancelled = "invite_cancelled"

	TypeSpectateStart = "spectate_start"
	TypeSpectators    = "spectators"

	TypeOppIntent = "opp_intent"
)

// OppIntent relays the acting player's Intent to a viewer who sees them as the
// opponent. Fields mirror Intent; identifiers are match-global (minion instance ids,
// hero player ids), so the receiving client resolves them against its own board
// (mapping a hero player id to self/opp by comparing to its own id).
type OppIntent struct {
	Type      string `json:"type"`
	HoverHand int    `json:"hoverHand"`
	Hover     string `json:"hover,omitempty"`
	AimFrom   string `json:"aimFrom,omitempty"`
	AimTo     string `json:"aimTo,omitempty"`
}

// SpectateStart confirms the requester is now spectating Target's match. It is
// sent just before the first (resync) State snapshot, so the client switches into
// its read-only spectator view before that snapshot arrives.
type SpectateStart struct {
	Type   string `json:"type"`
	Target string `json:"target"`
}

// Spectators tells both players who is currently watching their match. Sent to
// the two players (not the spectators) whenever the watcher set changes. Names is
// the spectators' usernames, sorted; empty means no one is watching.
type Spectators struct {
	Type  string   `json:"type"`
	Names []string `json:"names"`
}

// InviteReceived prompts the recipient: From challenges you to a match.
type InviteReceived struct {
	Type string `json:"type"`
	From string `json:"from"`
}

// InviteDeclined tells the inviter their invite was refused by By.
type InviteDeclined struct {
	Type string `json:"type"`
	By   string `json:"by"`
}

// InviteCancelled tells the invitee the invite from From is gone (the inviter
// withdrew or became unavailable), so the prompt should disappear.
type InviteCancelled struct {
	Type string `json:"type"`
	From string `json:"from"`
}

// CardView is a card as seen in a player's own hand (or a seek option / hero
// power / weapon). CardType is "minion"/"spell"/"secret"/"weapon"/"heroPower";
// for spells and hero powers, Target is the targeting rule ("none"/"any"/
// "minion"/"friendlyMinion"/"enemyMinion") so the client knows what to highlight.
type CardView struct {
	CardID     string `json:"cardId"`
	Name       string `json:"name"`
	CardType   string `json:"cardType"`
	Class      string `json:"class,omitempty"`  // "neutral"/"mage" — drives the card's color
	Rarity     string `json:"rarity,omitempty"` // "common"/"rare"/"epic"/"legendary" — drives the rarity gem
	Cost       int    `json:"cost"`
	BaseCost   int    `json:"baseCost,omitempty"` // hand cards only: printed cost, so the client can colour a cost-modified card
	Attack     int    `json:"attack"`
	Health     int    `json:"health"`
	Durability int    `json:"durability,omitempty"` // weapons
	Tribe      string `json:"tribe,omitempty"`
	Target     string `json:"target,omitempty"`
	ReqAttack  int    `json:"reqAttack,omitempty"` // targeted onset: target minion must have Attack >= this
	ReqTaunt   bool   `json:"reqTaunt,omitempty"`  // targeted onset: target minion must have Taunt
	ReqTribe   string `json:"reqTribe,omitempty"`  // targeted onset: target minion must be of this tribe
	Text       string `json:"text,omitempty"`      // human-readable rules text for the hover box
}

// MinionView is a minion in play, visible to both players. CanAttack means the
// minion is able to attack something this turn; CanAttackHero is the subset that
// may also go face (false for a Rush minion on its summon turn). Taunt/
// Aegis/Frozen drive badges and targeting (the server is authoritative).
type MinionView struct {
	InstanceID    string `json:"instanceId"`
	CardID        string `json:"cardId"`
	Name          string `json:"name"`
	Class         string `json:"class,omitempty"` // drives card color in the hover preview
	Cost          int    `json:"cost"`            // printed mana cost (for the full-card hover preview)
	Rarity        string `json:"rarity,omitempty"`
	Attack        int    `json:"attack"`
	Health        int    `json:"health"`
	MaxHealth     int    `json:"maxHealth"`
	BaseAttack    int    `json:"baseAttack"` // printed attack — client colors current vs base
	BaseHealth    int    `json:"baseHealth"` // printed health — client colors current vs base
	CanAttack     bool   `json:"canAttack"`
	CanAttackHero bool   `json:"canAttackHero"`
	Taunt         bool   `json:"taunt,omitempty"`
	Aegis         bool   `json:"aegis,omitempty"`
	Frozen        bool   `json:"frozen,omitempty"`
	Twinstrike    bool   `json:"twinstrike,omitempty"`
	Stealth       bool   `json:"stealth,omitempty"`
	Poisonous     bool   `json:"poisonous,omitempty"`
	Lifesteal     bool   `json:"lifesteal,omitempty"`
	FinalGasp     bool   `json:"finalGasp,omitempty"`
	SpellDamage   int    `json:"spellDamage,omitempty"`
	Enraged       bool   `json:"enraged,omitempty"`   // Enrage bonus currently active (damaged)
	HasEnrage     bool   `json:"hasEnrage,omitempty"` // has the Enrage ability at all (badge shows even at full HP)
	Silenced      bool   `json:"silenced,omitempty"`
	Elusive       bool   `json:"elusive,omitempty"`
	CantAttack    bool   `json:"cantAttack,omitempty"`
	Tribe         string `json:"tribe,omitempty"`
	Text          string `json:"text,omitempty"` // human-readable rules text for the hover box
}

// PlayerView is one side of the board. Hand/Secrets are populated only in the
// recipient's own view; the opponent sees HandCount/SecretCount only (hidden).
// Frozen marks the hero as frozen (blocks weapon attacks). HeroPower and Weapon
// are public.
type PlayerView struct {
	Name          string       `json:"name,omitempty"` // player's display username
	HeroHP        int          `json:"heroHP"`
	Armor         int          `json:"armor,omitempty"` // absorbs damage before health
	Frozen        bool         `json:"frozen,omitempty"`
	Immune        bool         `json:"immune,omitempty"` // hero ignores all damage this turn (frostward_aegis)
	Mana          int          `json:"mana"`
	MaxMana       int          `json:"maxMana"`
	Board         []MinionView `json:"board"`
	Hand          []CardView   `json:"hand,omitempty"`
	HandCount     int          `json:"handCount"`
	DeckCount     int          `json:"deckCount"`         // cards left in the draw pile (public)
	Secrets       []CardView   `json:"secrets,omitempty"` // own view only; opponent sees SecretCount
	SecretCount   int          `json:"secretCount,omitempty"`
	HeroPower     *CardView    `json:"heroPower,omitempty"`     // the hero's reusable ability (public)
	HeroPowerUsed bool         `json:"heroPowerUsed,omitempty"` // already used this turn
	Weapon        *WeaponView  `json:"weapon,omitempty"`        // equipped weapon, if any (public)
	HeroAttack    int          `json:"heroAttack,omitempty"`    // current hero attack (weapon attack)
	HeroCanAttack bool         `json:"heroCanAttack,omitempty"` // hero may attack right now
}

// WeaponView is the hero's equipped weapon, visible to both players.
type WeaponView struct {
	Name       string `json:"name"`
	Attack     int    `json:"attack"`
	Durability int    `json:"durability"`
	Text       string `json:"text,omitempty"`
}

// Joined confirms authentication; returns the connection's player id and the
// authenticated username.
type Joined struct {
	Type string `json:"type"`
	You  string `json:"you"`  // connection player id (turn identity)
	Name string `json:"name"` // authenticated username (display)
}

// Lobby reports live presence counts, shown while a player sits in the lobby.
// InGame counts players currently in an active (not finished) match.
type Lobby struct {
	Type    string       `json:"type"`
	Online  int          `json:"online"`
	InGame  int          `json:"inGame"`
	Players []PlayerInfo `json:"players,omitempty"`
}

// PlayerInfo is one entry in the lobby's online-player list. Status is
// "lobby" | "waiting" | "in_game". For in_game players, Vs is the opponent's
// name and MatchID identifies the match (reserved for spectator mode).
type PlayerInfo struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Vs      string `json:"vs,omitempty"`
	MatchID string `json:"matchId,omitempty"`
}

// Waiting means the player is queued, awaiting a second player.
type Waiting struct {
	Type string `json:"type"`
}

// OppConn tells a player whether their opponent's connection is live. Sent when
// the opponent drops mid-match (Connected false, a grace window before forfeit)
// and again if they reconnect in time (Connected true). Purely informational.
type OppConn struct {
	Type      string `json:"type"`
	Connected bool   `json:"connected"`
}

// Event is one ordered step of an action's resolution (a onset firing, a
// minion taking damage, a death, a finalGasp, a summon). The server emits them
// in resolution order alongside the resulting snapshot so the client can show a
// log now and animate later. Source/Target are absolute ids: a minion instance
// id, or a player id (p1/p2) for a hero. Identifiers are match-global, so the
// list is the same for both players (no hidden info here).
type Event struct {
	Kind   string `json:"kind"`             // play|onset|finalGasp|trigger|damage|heal|buff|summon|death|attack|freeze|shield|silence|secret|equip|weaponBreak|heropower|armor|mana|fatigue|burn
	Source string `json:"source,omitempty"` // acting entity (minion uid or player id)
	Target string `json:"target,omitempty"` // affected entity (minion uid or player id)
	Amount int    `json:"amount,omitempty"`
	// BuffAtk/BuffHP carry a "buff" event's stat delta so the log can show the
	// before→after (e.g. 5/5 → 7/7). Omitted (0) for keyword-only / swap buffs.
	BuffAtk int    `json:"buffAtk,omitempty"`
	BuffHP  int    `json:"buffHp,omitempty"`
	Name    string `json:"name,omitempty"` // card name, when useful for display (summon/death/triggers)
	// Note is a short contextual subtitle for the log popup. Set on a "secret"
	// reveal to name the card the secret acted on (e.g. the spell a Counter Spell
	// negated), which the secret's own Name/Card can't convey.
	Note string `json:"note,omitempty"`
	// Card is the played card, set on "play" events so the client can reveal an
	// opponent's cast (the opponent's hand is otherwise hidden). Omitted for
	// secrets (hidden) and all other event kinds.
	Card *CardView `json:"card,omitempty"`
}

// MatchStart announces a match was made. It carries the initial snapshot so the
// client can render the opening board; it is sent per-player (Self/Opp differ).
type MatchStart struct {
	Type     string     `json:"type"`
	Players  []string   `json:"players"`            // player ids, both sides
	Turn     string     `json:"turn"`               // id of player to act
	You      string     `json:"you"`                // recipient's player id
	Self     PlayerView `json:"self"`               // recipient's side (hand visible)
	Opp      PlayerView `json:"opp"`                // opponent's side (hand hidden)
	Mulligan bool       `json:"mulligan,omitempty"` // true: match opens in the mulligan phase
}

// State is the full game snapshot, sent per-player after each action. Events is
// the ordered resolution log for the action that produced this snapshot.
type State struct {
	Type     string     `json:"type"`
	Turn     string     `json:"turn"`               // id of player to act
	TurnNum  int        `json:"turnNum"`            // increments each end_turn
	You      string     `json:"you"`                // recipient's player id
	Self     PlayerView `json:"self"`               // recipient's side (hand visible)
	Opp      PlayerView `json:"opp"`                // opponent's side (hand hidden)
	Events   []Event    `json:"events"`             // ordered resolution log (may be empty)
	Mulligan bool       `json:"mulligan,omitempty"` // still in the mulligan phase (this player submitted, awaiting opponent)
	Resync   bool       `json:"resync,omitempty"`   // reconnect snapshot: Events is the full recent history to REPLACE the client log (not append)
	TurnSecs int        `json:"turnSecs,omitempty"` // whole seconds left in the current turn (0 if none)
}

// Seek prompts the recipient to pick one of Options (added to their hand).
// The server pauses the triggering action and ignores the player's other actions
// until a Choose arrives. Sent only to the choosing player (the options are not
// hidden info, but only they may pick).
type Seek struct {
	Type    string     `json:"type"`
	Options []CardView `json:"options"`
}

// OppSeek tells the waiting player that their opponent is choosing a Seek
// card right now. Count is how many options the opponent is picking from (the
// faces stay hidden). Sent when the opponent's Seek begins; the next State
// (after they choose) clears it on the client.
type OppSeek struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// GameOver ends the match. Winner is the surviving player's id.
type GameOver struct {
	Type   string `json:"type"`
	Winner string `json:"winner"`
}

// Error reports a rejected action.
type Error struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

// Marshal is a small helper that ignores marshal errors (our own structs
// never fail to marshal).
func Marshal(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
