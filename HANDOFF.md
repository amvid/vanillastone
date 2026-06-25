# HANDOFF — Vanillastone

Living doc for future sessions. **Keep updated every session. Read this first.**

Full session-by-session history (phases 1–10 + every card-clone wave) lives in
`HANDOFF.archive.md` and git history — this file is the lean current-state summary.

Last updated: **2026-06-25** (art in progress — 104/140 collectible art files placed + 2 tokens:
`pyrebolt`, `mana_surge`; collectibles placed = all 32 legendaries + all 13 neutral epics + first
neutral rares `managlutton`, `veiled_assassin`, `bazaar_crier`, `dawnguard_protector`, `vanguard_champion`, `cobalt_loreling`, `grudge_smith`, `mesmer_adept`, `sapphire_drake`, `tollkeeper_brute`, `trampling_brute`, `adept_tutor`, `bannerguard`, `cabal_overseer`, `covert_saboteur`, `runeward_sage`, `brineseer_diviner`, `clockwork_swapbot`, `imp_warden`, `moonfury_stalker`, `relic_seeker`, `runed_golem`, `shadowtail_familiar`, `wounded_duelist`, `venom_serpent`, `siege_engine`, `tidescry_oracle`, `stoneveil_watcher`, `glimmerwing_drake`, `wardstone_sentinel`, `dagger_tosser`, `ashflame_zealot`, `forge_hand`, `spellrage_acolyte`, `mana_leech`, `pocket_conjurer`, `addled_brewer`, `riled_rooster`, `brine_cutter`, `dawn_tender`, `rune_warden`, `acolyte_novice`, `brackish_caller` + all 16 non-legendary Mage collectible art files
(`arcane_adept`, `warded_scholar`, `spellwarden_magus`, `glacial_splinter`, `arcane_wyrmling`,
`frost_tempest`, `codex_of_insight`, `frostshear`, `frostlance`, `pyrecataclysm`, `nullrune`,
`cinder_trap`, `echo_glass`, `glacial_ward`, `decoy_ward`, `frostward_aegis`);
Mage hero/UI art added: `mage_hero`, `fire_dart`;
all legendary art done; all epic art done; **all rare art done** (49/49 verified); common art started high-to-low by cost;
prompt prefix overhauled to style-first + hard negatives + cartoon character-design + top-35% empty
framing; CardFace art slot crops bottom-anchored; **board minions now render real art** (`MinionArt`,
zoomed `auto 160%`/`center bottom` to cut the top sky); hover card-preview delayed **0.7s**. Prev
2026-06-23: art layout finalized — full-bleed ~square slot via `background-cover`, fixed the
`<button>` global-padding inset gotcha, style locked to **cel cartoon**, vite auto-reload on art
drops, all 30 legendary prompts drafted in `.notes/art-prompts.md`. See `web/public/art/README.md`).

---

## What this is
Hearthstone-*inspired* digital card game. **Custom** cards/art/names — NOT Blizzard's IP.
Goal: a fast HS-like 1v1 to play with friends. Solo project.
Repo: `github.com/amvid/vanillastone` (will be public).

---

## ⚠️ HARD RULE — no IP names in code (READ FIRST)
NEVER put Blizzard / Hearthstone / Warcraft card or character names in source code OR
comments OR tests — public repo, git history is forever. Use our original card **ids**
(e.g. `wraithqueen_selvara`) or a generic mechanic description. Real names live ONLY in
**`.notes/classic-mapping.md`** (gitignored). Also a memory: `no-ip-names-in-code`.

---

## Locked decisions

| Topic | Decision |
|---|---|
| Server | Go, **authoritative** (owns all truth, client is a dumb renderer) |
| Transport | WebSocket + JSON |
| Client | **Web** (React + Vite + TS), NOT Godot |
| Hosting | One Go binary serves client (`go:embed` dist) + runs `/ws` |
| Accounts | Username + password, unique username (`users.username UNIQUE`, trimmed, charset `[A-Za-z0-9_]{3,20}`), bcrypt. Register + login. No email/reset/roles |
| Sessions | In-memory `token→username`. POST /login issues a 256-bit token. Dies on restart (re-login) |
| Storage | SQLite via **`modernc.org/sqlite`** (pure Go, no CGO). Match state in-memory |
| Cards | All players have all cards. Custom set |
| Mode | 1v1 only. No AI |
| Hero | Mage only (Hunter reserved, not built) |
| Card design | Faithfully clone the **HS Classic + Hall of Fame** set: match cost/stats/effect, give wholly original names + art. Mechanics aren't copyrightable; names/art/flavor are |
| RNG | Server-side, seeded |
| Match persistence | In-memory (dies on restart) — OK for now; restart-survival is OUT |

### Legal rules (public repo)
- Card numbers/effects follow genre staples; **names + art are wholly original**.
- **Never commit** Blizzard art/sounds/fonts/names/flavor/logos/datamined files.
- **Safe**: own names + art, CC0/CC-BY assets (Kenney, game-icons.net). Track every 3rd-party
  asset in `ASSETS.md`. `LICENSE` = MIT. Don't name it "Hearthstone". No commercial intent.

---

## Architecture
```
Web client (renderer)  <--WebSocket/JSON-->  Go server (authoritative)
  sends: PlayCard, Attack, HeroPower,          - match engine (state machine)
         EndTurn, Target, Mulligan, Choose     - event bus + effect resolver
  recv:  full GameState snapshot +             - seeded RNG
         ordered event log (for animation)     - lobby / matchmaking · SQLite (decks)
```
Server resolves instantly, sends an **ordered event list** + resulting full snapshot.
Client replays events as animation, settles to the snapshot. Client is swappable
because the server owns truth.

**Effect system:** hybrid — common effects (damage/draw/buff) are data-driven `Effect`
params in card data; weird ones are Go handlers keyed by effect/secret kind.

### Code layout (`internal/`)
- `cards/` — card data (`neutral.go`, `mage.go`) + deck/pool helpers (`cards.go`). Source of truth.
- `match/` — engine, split into `match.go` (types/lifecycle/timer), `actions.go`
  (PlayCard/Attack/HeroPower/Concede/Choose/Mulligan), `state.go` (seats/reconnect/snapshots/
  spectators), `engine.go` (triggers/secrets/summon/seek), `effects.go` (applyEffect/
  damage/heal/buff/silence/auras/targeting), `view.go` (death resolution + snapshot builders).
- `protocol/` — wire types (Go), mirrored in `web/src/protocol.ts`.
- `store/` (SQLite: users, decks), `auth/` (bcrypt + deck HTTP API), `lobby/`, `transport/` (ws).
- `web/` — React client; build → `web/static` (committed, `go:embed`'d). `web/public/art/` = card art.

---

## Mechanic taxonomy (the spec — all covered)
**Card types**: Minion, Spell, Weapon, Hero Power.
**Keywords** (original names — see "IP renames" below): Onset, Final Gasp, Taunt, Charge/Rush,
Aegis, Twinstrike, Spell Damage, Secret, Aura, Freeze, Silence, Stealth, Poisonous, Lifesteal,
Enrage, Seek/Choose (Overload deferred to Shaman).
**Tribes**: Beast, Gilkin, Pirate, Dragon, Mech, Demon, Elemental, Undead, Riftborn.
**Triggers (event bus)**: on_play, on_death, on_damage, on_heal, on_turn_start/end,
on_attack, on_summon, on_spell_cast, on_card_draw, on_fatal_damage, secret triggers.
Engine covers all of these; custom cards are data on top.

---

## Build phases — ALL DONE
1. Skeleton loop · 2. Minions + combat · 3. Spells + targeting · 4. Event bus + triggers ·
5. Keywords wave 1 (taunt/charge/rush/aegis/freeze) · 6. Keywords wave 2 (spell dmg/
aura/silence/stealth/poisonous/lifesteal/twinstrike) · 7. Secrets + Seek · 8. Hero power +
weapons + armor · 9. Decks + SQLite (deckbuilder/mulligan/fatigue) · 10. Polish (reconnect,
animations, turn timer, prod deploy). Detail for each is in `HANDOFF.archive.md`.

---

## Current state (2026-06-24)

**Card set — COMPLETE.** Full HS **Classic (120/120)** + **Hall of Fame** sets cloned as
original-named cards, 1:1 mechanics. Mage 15/15 (full class set). Real pool now:
- **140 collectible** = **123 neutral / 17 mage** (`DeckPoolIDs()`; tokens + hero power excluded).
- Source of truth = **`.notes/classic-mapping.md`** (gitignored — real names only there; our
  id + cost/stats/effect/tribe + status). **Read it before any card work.**
- Coverage guarded by `cards.TestClassicMechanicsHaveCards`. `go test -race ./...` green,
  gofmt/vet clean. The card-clone phase is **done** — next card work = balance or scope expansion.

**Engine** covers the full taxonomy above, incl. (added across the clone waves): transform,
enrage, grant-keyword/spell-damage, adjacency/tribe auras (atk + HP), bounce, mind-control,
random-generate/summon/transform, cost-modification (`effectiveCost`: intrinsic CostRule +
in-play CostAura + free-secret/free-spell flags), set-health, swap-stats/with-hand, weapon
manipulation, Start-of-Game hero-power upgrades, Ice Block immunity, copy-spell, and the live
attack model (`attacksMade < attacksPerTurn()`).

**Deckbuilding** — decks bind to a **class** (server-authoritative `ValidateDeck(ids, class)`;
`PlayableClasses() = {mage}`, Hunter reserved/card-less). SQLite `decks` has a `class` column.
Legendary 1-copy cap; others `MaxCopies=2`; `DeckSize=30`, `MaxDecksPerUser=10`. Curated
`DefaultDeck()` = freeze/tempo Mage. Client: class picker, collection tabs (class/Neutral),
mana + rarity filters, ascending-mana sort, mana-curve histogram.

**Client UI** — HS-style table, fully redesigned (see TASKS.md history):
- Board minions = **art objects** (portrait + rarity/atk/hp gems + name tag; full card on hover).
  Keyword visuals: Taunt shield crest, Divine Shield gold ring, Frozen sheen, ❄ badge; action
  state via `outline` so it's never hidden by a frame/aura. **Portrait art** = `MinionArt`
  (`Board.tsx`) paints `/art/<cardId>.png`; the minion box (132×150) is barely tall, so plain cover
  would show the art's empty top-35% sky band — instead it's zoomed `background-size: auto 160%`,
  `center bottom` to crop the sky and frame the face (tweak the 160% to taste). **Hover preview**
  (`.minion-preview`, `index.css`) is delayed **0.7s** via opacity/visibility + `transition-delay`
  (NOT `display`, which can't delay); hides instantly on leave. **Battlecry token sequencing**
  (`GameScreen.tsx` layout effect + `summonPop`): a minion played from hand flies in first; any
  token it summons (Onset/battlecry, same event batch) `summonPop`s with a delay — `FLY_LAND`
  (650ms, ≈ fly-in) so the played minion lands before the token pops, then `TOKEN_GAP` (320ms)
  staggers multiple tokens. `summonPop(cid, dur, delay)` holds the token hidden via the first
  keyframe + `fill:'backwards'` (no flash at final spot during the delay).
- **CardFace** = full-bleed art + overlaid cost gem + name plate + rules box (font auto-shrinks
  by length) + type band; rarity shown by title colour. Card 164×236. Name band:
  `align-items:center` (title centred in band, not jammed to top), `padding:0 28px`
  (clears the 28px-wide cost gem both sides so a centred title never hides behind it).
  Title font auto-shrinks by longest-word AND total length (`nameLen>=14→9px`,
  `>=20→8px`) so long two-word names (e.g. Arcane Wyrmling) fit + clear the gem.
- **Deckbuilder in-deck hover preview** (`.dc-floating-preview`, `Deckbuilder.tsx`):
  reuses `CardFace`, portaled to `<body>`. Two fixes worth remembering: (1) it escapes
  `.builder-page`'s `scale(--u*--ui-bias)`, so it re-applies that same scale itself
  (`transform-origin:right center` to grow leftward from the panel) or it renders
  smaller than the scaled-up book cards; (2) its `.name` must be in the absolute
  name-overlay selector list or the title falls into normal flow (lands in the textbox).
- **Art wiring** done: `CardArt` paints `/art/<cardId>.png` as a **`background-size:cover`** layer
  filling a **full-bleed** `.card-art` slot (cost gem + name plate overlay its top); a hidden probe
  `<img onError>` flips to the placeholder type glyph for missing art. Drop files into
  `web/public/art/` incrementally. **GOTCHA (cost a long debug):** card faces are `<button>`s and
  the global `button { padding: 6px 10px }` was insetting the whole content box ~10px — fixed with
  `padding: 0` on `.card`/`.book-card`. Without that, no full-bleed/edge alignment will ever work.
  Style = **clean cel-shaded cartoon** (see `web/public/art/README.md` prefix). **Hand cards** use
  `CardArt`'s full-bleed `background-size:cover` anchored **`center bottom`** (crop eats the top sky,
  keeps feet/tail). **104/140 collectible art files placed** + tokens `pyrebolt`, `mana_surge`: all 32 legendaries,
  all 13 neutral epics, all 49 rares (verified complete after `brackish_caller`), plus all non-legendary Mage collectibles:
  `arcane_adept`,
  `warded_scholar`, `spellwarden_magus`, `glacial_splinter`, `arcane_wyrmling`, `frost_tempest`,
  `codex_of_insight`, `frostshear`, `frostlance`, `pyrecataclysm`, `nullrune`, `cinder_trap`,
  `echo_glass`, `glacial_ward`, `decoy_ward`, `frostward_aegis`; rest placeholder. Prompts in
  `.notes/art-prompts.md`. Rare wave DONE (highest to lowest cost): `managlutton` (8),
  `veiled_assassin` (7), `bazaar_crier` (6), `dawnguard_protector` (6), `vanguard_champion`
  (6), `cobalt_loreling` (5), `grudge_smith` (5), `mesmer_adept` (5), `sapphire_drake` (5),
  `tollkeeper_brute` (5), `trampling_brute` (5), `adept_tutor` (4), `bannerguard` (4), and
  `cabal_overseer` (4), `covert_saboteur` (4), `runeward_sage` (4), `brineseer_diviner` (3), and
  `clockwork_swapbot` (3), `imp_warden` (3), `moonfury_stalker` (3), `relic_seeker` (3),
  `runed_golem` (3), `shadowtail_familiar` (3), `wounded_duelist` (3), `venom_serpent` (3),
  `siege_engine` (3), `tidescry_oracle` (3), `stoneveil_watcher` (2), `glimmerwing_drake` (2),
  `wardstone_sentinel` (2), `dagger_tosser` (2), `ashflame_zealot` (2), `forge_hand` (2),
  `spellrage_acolyte` (2), `mana_leech` (2), `pocket_conjurer` (2), `addled_brewer` (2), and
  `riled_rooster` (1), `brine_cutter` (1), `dawn_tender` (1), `rune_warden` (1),
  `acolyte_novice` (1), and `brackish_caller` (1) placed after user approval. Rare coverage check:
  49/49 rare cards have art. Next preview target = `galewing_harpy` (6-cost common). User noted the recent batch skewed boy-heavy; vary upcoming
  humanoid prompts with more women/girls where the card concept allows.
  Updated workflow: generate a raw
  preview and wait for user review; on `continue`, resize/compress/place/update notes, then start
  the next card; on `rework`, regenerate without placing.**
- Hero = framed art slot using `/art/mage_hero.png` (emoji underneath as fallback); hero power =
  framed circle using `/art/fire_dart.png` and still **3D-flips** when used.
- Mana = vertical crystal column right of each deck pile (out of flex flow). Hands hug the edges.
- **Viewport scaling** — one `--u` knob on `App` = `min(1, vw/2056, vh/1329)` floored 0.65
  (2056×1329 = dev display = u-baseline). Chrome screens scale whole-element (`--ui-bias` 1.5);
  the board uses inflate-then-scale on `.game-stage` (`--game-bias` 1.1 — past ~1.1 the board
  clips the hand vertically). Drag clone + targeting arrow stay screen-space.
- **Event log (`buildLog`, `format.ts`)** — **grouped**: one row per cause (play/attack/hero
  power/battlecry/deathrattle/trigger/secret) carrying EVERY character it affected, so an AoE is a
  single row (source + all targets w/ outcomes) not one line per hit. A `death` folds into the
  target it killed (☠ badge, looked up by uid across the whole action — handles a kill whose blow
  was in an earlier group), never a separate mis-ordered "dies" line. Feed shows source chip +
  kind icon + ☠/×N badges; hover popup = source card → wrapped grid of affected cards each w/ its
  verb/amount/☠. Newest-on-top everywhere (live path reverses the action's groups, matching
  resync) so order never reads backwards. `LogEntry` now = `{kind,text,source,note,targets[]}`.
  Feed is Y-centered (`.log` `justify-content:center`), shows the last 25 (`slice(0,25)`). Hover
  popup target grid (`.log-pop-targets`) wraps wide (`max-width:82vw` — vw is real since the popup
  is portaled to `<body>`, not the scaled stage — popup `scale(0.8)`) so a big AoE spreads across a
  couple short rows (up to ~75% screen) instead of a tall off-screen stack. A **Counter Spell**
  reveal now names the spell it negated:
  server sets `Event.Note = spellName` (threaded via `secretCtx.spellName`), popup reads
  "<secret> counters <spell>".
- Spectator mode, direct invites (challenge a lobby player), online-players panel, edge-trigger
  pulse, 75s turn timer all done.
- **Play mode picker** — the lobby **Play** button opens a modal (`.mode-picker`, `App.tsx`):
  **vs AI** (class dropdown → random prebuilt deck), **vs Player** (queue), **Arena** (disabled,
  "Coming soon"). Shared **Your deck** selector lives in the modal. While queuing, Play flips to
  the cancel toggle (modal suppressed). **AI plays / AI deck pickers** now use `FancySelect`
  (generic styled dropdown sharing `.deck-select` chrome, with art icons) instead of raw
  `<select>`s; constrained to the card via `.mode-aiclass .deck-select {flex:1;min-width:0}`.
- **Collection empty-state fix** — `.book-grid` pads invisible `.book-card.placeholder` slots to
  `PER_PAGE` even when **zero** cards match (was guarded by `pageCards.length > 0`), so the `1fr`
  columns no longer collapse and the book keeps full size under the "No cards match" message.

**AI opponent (vs-AI, Phase 14)** — server-authoritative, **no ML model**: a single-turn greedy
planner over a board heuristic. Files in `internal/match/`: `clone.go` (`cloneForSim` — pure deep
copy of match state + no-op `botSender` + fresh RNG + timers off; the state graph is a clone-friendly
tree — no back-pointers, auras are recomputed ints, `owner` is an index), `ai.go` (threat-aware
`eval`, `burstNow`/`burstNextTurn`, broad move enumeration, greedy `planBest` + lethal-lens
`scoreForPlanner`), `ai_driver.go` (`NewVsAI`, `EnableAI`, async `runBotTurn` off the turn-handoff
goroutine, auto-mulligan keep-all, discover handling, `botActionDelay` paces moves — a `var` so tests
zero it). Wiring: `lobby.StartVsAI` → `transport.startVsAI` picks a random `cards.AIDecks(class)` →
`protocol.FindMatch{VsAI, AIClass}`. **3 prebuilt Mage decks** in `cards/ai_decks.go` (default/aggro/
midrange), guarded by `TestAIDecksAreLegal`. Behavior: **threat term + lethal lens** make it trade to
survive (kill a threat) instead of always going face, and race when it has lethal — locked by
`TestPlannerTradesToAvoidLethal`. Tested: clone independence, planner cases, full vs-AI flow under
`-race`. **Dumb-but-legal** by design; the clean seam for a future **hard mode** = 2-ply lookahead on
the same clone. **Class picker is Mage-only** until a second class ships (only Mage has a pool).

**Infra/ops** — `make dev` (docker compose: server :8080 + vite :5173, open **:5173**).
`make prod [PORT=]` = single distroless binary, SQLite in `vs-data` volume. `.githooks/pre-commit`
builds + stages `web/static` (`make hooks`). **nginx in front MUST set `proxy_http_version 1.1`
+ Upgrade/Connection headers** or the ws handshake fails (coder/websocket rejects HTTP/1.0).

---

## Open / next
- **Auth UX + validation hardening (2026-06-25) — DONE.** Server (`internal/auth/auth.go`):
  username trimmed + restricted to `[A-Za-z0-9_]{3,20}` (`usernameRe`), password bounded
  6–72 (bcrypt's 72-byte input limit, rejected up front → clean 400 not 500). Login trims too,
  so a padded signup resolves to the same unique account. New tests: charset/whitespace/long-pass
  rejection + trim-uniqueness. Client (`App.tsx`): single username/password form with **Login**
  (primary) + **Register** (secondary, below) buttons — no mode toggle. Per-field inline validation
  mirroring the server (`usernameError`/`passwordError`), both buttons disabled until valid,
  invalid-field red border (`.auth-input.invalid`), `.field-err` messages. Register **auto-logs-in**
  on success.
- **Gameplay-readability polish (2026-06-25) — DONE (see TASKS.md).** Client-only except #2.
  - **Log same-name bug**: dedup keyed on instance `uid` (was `cardId`), so a minion attacking
    another minion of the same card no longer collapses to an empty popup (`LogActor.uid`).
  - **Opp spell/weapon play**: slow ~1.2s card travel hand→center (`.opp-cast-fly`); minions keep
    board flyIn; hidden secrets have none. Left cast showcase now 4s (was 5s).
  - **Secrets**: center burst ring + "🔮 <name>" on ANY secret fire (`.secret-burst`); a
    Counter-type secret shows BOTH cards paired (negated spell desaturated ✕ secret) — paired
    client-side from the same event batch (`emitPlay` precedes `triggerSecrets`), no protocol change.
  - **#2 opponent intent/aiming**: NEW ephemeral, non-authoritative channel — `protocol.Intent`/
    `OppIntent`, `Match.RelayIntent` (to the other seat + its spectators only; never stored/logged),
    `transport.handleIntent`, guarded by `TestRelayIntentToOpponentOnly`. Wire tokens are
    perspective-free (`self`/`enemy` heroes flipped by the receiver, `heroPower`, `hand:<i>`,
    minion uids). Client streams hover/aim (throttled 45ms) and renders the opp's held hand-card
    lift + inspected/aimed outline + dashed ghost arrow. v1: arrow snaps to targets (no
    free-cursor line); bot emits nothing.
  - **IP renames — DONE (2026-06-25).** Swept ~590 occurrences (display Text + Go identifiers +
    wire json keys + CSS classes + comments + tests) to original names — keywords **Onset / Final
    Gasp / Seek / Aegis / Twinstrike**, tribes **Gilkin / Riftborn**. Wire keys
    (`finalGasp`/`aegis`/`twinstrike`, msg types `seek`/`opp_seek`) + CSS classes kept in sync
    server↔client. `go test -race ./...` + `tsc` + `vite build` all green. Generic keywords
    (Taunt/Charge/Rush/Stealth/Silence/Freeze/Secret/Lifesteal/Poisonous/Enrage/Spell Damage) kept.
    Real-name→our-name map lives ONLY in gitignored `.notes/classic-mapping.md`.
- **Mulligan polish — DONE.** 20s client countdown (`MULLIGAN_SECS`, `App.tsx`) auto-keeps the
  current selection at 0; server backstop `mulliganLimit` (30s, `match.go`) force-keeps a dead
  client so the match can't hang. Mulligan→play now dissolves via a `.play-reveal` curtain
  (blur/dark → clear), then deals the opening hand in from the deck staggered (`dealIn`,
  `GameScreen` `intro` prop); the first player's turn-1 draw flies in a beat later as a
  separate "draw". The normal per-draw `flyIn` is suppressed during the deal (`!intro`).
- **vs-AI seating — random.** `NewVsAI` coin-flips who goes first: human takes seat 0 (first)
  or seat 1 (second, gets Mana Surge), bot on the other. Was always human-first.
- **Board death anim — DONE.** `deathPuff` strips stat badges (no lingering overlapping atk/hp);
  rect cache skips minions mid fly-in/summon (`isEntering`) so `settleShift` no longer slides a
  just-played minion in from the hand on the next death.
- **#7 — real art assets (TASKS.md), IN PROGRESS.** ~140 images, one per card. **Read
  `web/public/art/README.md` first** (style prefix + framing + placement pipeline + dev reload).
  Loop: assistant generates one raw preview (attach/use latest good cartoon art as style context when
  useful) and stops for user review. On `continue`, the assistant copies the approved output into
  `web/public/art/<cardId>.png`, resizes to 512w, compresses <150KB (drop PIL palette `quantize(N)`
  until under), updates notes, then starts the next card preview. On `rework`, regenerate without
  placing. **Do not visually review generated art unless asked**; the user reviews and requests
  changes to save tokens. Decided this
  session: **style = clean cel-shaded CARTOON** (not painterly); slot is **full-bleed ~square**,
  generate **1:1 @ 1024**; **FRAMING — entire top 35% must be empty sky (subject's highest point
  below the 35% line), head ~50-60% height** or the title/cost overlay covers it; **bottom edge must
  be filled** (the card crop is bottom-anchored). Per-card prompts queued in **`.notes/art-prompts.md`**.
  83/140 collectible art files placed — all 32 legendaries DONE:
  `voidwyrm_tyrant` (10);
  `emberwing_matron`, `lunar_devourer`, `chronlord_zhal`, `dreamwarden_ylena`, `emberqueen_valtha`,
  `spelltide_wyrm` (9); `cragmaw`, `emberlord_vrakgar` (8); `cinder_baron`, `emberforge_magus` (7);
  `hornelder_chief`, `snarlmaw`, `the_gorehound`, `grave_knight`, `nightmare_lord`,
  `revenant_priestess`, `gearmaster_cog`, `duskwarden_genmar`, `wraithqueen_selvara` (6).
  Cost 5 DONE: `captain_brackwater` (120KB), `relic_breaker` (149KB), `warhorn_chieftain`
  (139KB), `reckless_vanguard` (131KB), generated/placed without assistant visual review.
  Cost 4 DONE: `brinelord_gorrak` (146KB), generated/placed without assistant visual review.
  Cost 3 DONE: `grovelord_brakka` (144KB), `sprocket_tinkerer` (143KB), generated/placed
  without assistant visual review.
  Cost 2 DONE: `vael_emberscribe` (149KB), `archivist_solenne` (149KB), `fizzle_sparkmuddle`
  (145KB), `lucky_angler` (137KB), `gleamwing` (137KB), generated/placed without assistant
  visual review.
  Epic art started high-cost → low: `magma_behemoth` (20-cost, 138KB), first draft rejected as too
  close to `cinder_baron`, reworked as a low volcanic landform-colossus; `crag_colossus` (12-cost,
  145KB), generated/placed without assistant visual review; `tidecolossus` (10-cost, 132KB),
  generated/placed as a low oceanic colossus with crowded battlefield silhouettes around its base;
  `wilds_beastcaller` (7-cost, 142KB), generated/placed as a forest caller summoning one beast;
  `trophy_hunter` (5-cost, 146KB), generated/placed as a clever monster-slayer snaring one huge target;
  `visage_thief` (5-cost, 142KB), generated/placed as a masked copier stepping through mirror magic;
  `rotgut_horror` (5-cost, 145KB), generated/placed as a squat taunt undead with a green deathrattle shockwave;
  `crimson_reaver` (3-cost, 146KB), generated/placed as a shield-consuming red-armored reaver;
  `reef_warchief` (3-cost, 146KB), generated/placed as an original reef-clan aura leader;
  `tidehook_captain` (3-cost, 148KB), generated/placed as an original sea-corsair crew aura leader.
  Final 3 epics generated from one triptych prompt and cropped into separate assets: `ruin_oracle`
  (2-cost, 149KB), `corsair_macaw` (2-cost, 139KB), `shellback_crab` (1-cost, 136KB).
  **Epic art is complete.** **Rare art is complete** (49/49 verified after `brackish_caller`).
  Rare art high-cost → low: `managlutton` (8-cost, 119KB),
  `veiled_assassin` (7-cost, 122KB), `bazaar_crier` (6-cost, 137KB), `dawnguard_protector`
  (6-cost, 113KB), `vanguard_champion` (6-cost, 105KB), `cobalt_loreling` (5-cost, 121KB),
  `grudge_smith` (5-cost, 150KB), `mesmer_adept` (5-cost, 123KB), `sapphire_drake`
  (5-cost, 144KB), `tollkeeper_brute` (5-cost, 142KB), `trampling_brute` (5-cost, 146KB),
  `adept_tutor` (4-cost, 131KB), `bannerguard` (4-cost, 132KB), `cabal_overseer`
  (4-cost, 136KB), `covert_saboteur` (4-cost, 146KB), `runeward_sage` (4-cost, 138KB),
  `brineseer_diviner` (3-cost, 137KB), `clockwork_swapbot` (3-cost, 146KB), `imp_warden`
  (3-cost, 144KB), `moonfury_stalker` (3-cost, 147KB), `relic_seeker` (3-cost, 146KB),
  `runed_golem` (3-cost, 145KB), `shadowtail_familiar` (3-cost, 137KB), `wounded_duelist`
  (3-cost, 131KB), `venom_serpent` (3-cost, 133KB), `siege_engine` (3-cost, 145KB), and
  `tidescry_oracle` (3-cost, 124KB), `stoneveil_watcher` (2-cost, 144KB),
  `glimmerwing_drake` (2-cost, 130KB), `wardstone_sentinel` (2-cost, 143KB), `dagger_tosser`
  (2-cost, 131KB), `ashflame_zealot` (2-cost, 145KB), `forge_hand` (2-cost, 144KB),
  `spellrage_acolyte` (2-cost, 140KB), `mana_leech` (2-cost, 143KB), `pocket_conjurer` (2-cost,
  143KB), `addled_brewer` (2-cost, 144KB), `riled_rooster` (1-cost, 128KB), `brine_cutter`
  (1-cost, 142KB), `dawn_tender` (1-cost, 136KB), `rune_warden` (1-cost, 143KB),
  `acolyte_novice` (1-cost, 143KB), and `brackish_caller` (1-cost, 139KB), generated raw for
  user review, then placed after approval without assistant visual review. Next preview target:
  `galewing_harpy` (6-cost common). Going forward, vary humanoid prompts with more
  women/girls where suitable.
  Mage review batch 1 DONE: `arcane_adept` (512w, 110KB), `warded_scholar` (123KB),
  `spellwarden_magus` (121KB), `glacial_splinter` (130KB; first version rerolled lower/wider for
  top-band safety). Mage review batch 2 DONE with stronger variety: `arcane_wyrmling` (warm indoor
  creature portrait, 138KB) and `frost_tempest` (cold outdoor AoE spell, 131KB). Mage batch 3 DONE:
  `codex_of_insight` (143KB) and `frostshear` (130KB), no assistant visual review per updated workflow.
  Mage single-card batches DONE: `frostlance` (119KB), `pyrecataclysm` (replaced after user noted
  first version read as AoE; now single-target pyre-comet, processed/not checked), no assistant visual
  review. `nullrune` generated/placed as an abstract counterspell rune seal, processed/not checked.
  `cinder_trap` generated/placed as a single-attacker ember snare, processed/not checked.
  `echo_glass` generated/placed as a mirror-copy secret, processed/not checked.
  `glacial_ward` generated/placed as a defensive frost armor ward, processed/not checked.
  `decoy_ward` generated/placed as a redirect-spell dummy ward, processed/not checked.
  `frostward_aegis` generated/placed as final non-legendary Mage collectible, processed/not checked.
  Token `pyrebolt` generated/placed after several reworks; latest version keeps the bright day-sky
  background but makes the projectile a more dramatic falling fireball. Token `mana_surge` generated/
  placed as pure blue-gold mana energy forming an extra crystal; earlier coin-like drafts were rejected
  for IP/design distance and because the card name is about a surge, not a coin. **Mage collectible art is complete.**
  Non-card UI art DONE: `mage_hero` portrait + `fire_dart` hero power, processed/not checked and
  wired in `Hero.tsx`/`GameScreen.tsx`/`index.css`.
  Cartoon, full-bleed, 512w / <150KB; art slot crops vertically, anchored bottom so the top sky is
  what's lost; busy/high-contrast images may need quantize down to q=80–128. Placing high cost → low;
  next art = continue remaining non-legendary neutral cards high cost → low:
  generate/place/process/update notes, then
  wait for user review before continuing. **Framing rule bumped to top 35% empty** (was 33%) — subject's highest
  point (head/crown/flames/weapon tip) must sit below the 35% line, head ~50-60% height; updated in
  both `.notes/art-prompts.md` + `web/public/art/README.md`. **IP watch:** fire-elemental gens drift
  toward the iconic molten-boss design — keep flame crown (not skull-helm) + red/gold trim + no
  spiral-rune weapon to stay original. `cinder_baron` = raw armorless magma brute (distinct from `emberlord_vrakgar`). **Prompt prefix upgraded** (style-first +
  hard negatives + cartoon character-design + no-crop composition + ban default volcano/lava;
  synced in `.notes/art-prompts.md` + `web/public/art/README.md`) — kills the semi-realistic drift.
- **Manual end-to-end playtest** — Ice Block reveal, mind-control board-cap fizzle, Start-of-Game
  hero-power swap, ETC giving both hands a random anthem. All unit-covered, never seen live.
- **Deferred bug** — rare phantom "opponent played a card" right after mulligan with no action;
  suspect a stale resync/`history` replay across matches. Not reproduced.
- **AI opponent — DONE (dumb-but-legal).** See Current state "AI opponent". Future: **hard mode** =
  2-ply lookahead on the same `cloneForSim` (the seam is already there); eval-weight tuning; bot
  going first. Currently Mage-only.
- **Hunter — second class (not started).** Generalize the hardcoded `MageHeroPower()` first, then
  add per-class pools + hero power + class-scoped deckbuilding. Hero Power: Quick Shot (2 mana,
  2 dmg to enemy hero). Beast/trap/reach cards. **Also unlocks the AI/lobby class picker** (today
  Mage-only) and needs an `AIDecks(hunter)` entry in `cards/ai_decks.go`.
- **Scope:** card-clone phase is complete — next is balance/playtest/polish or a deliberate scope
  expansion (more classes, beyond Classic/HoF). **Discuss before starting** (phase-by-phase rule).

### Standing gotchas / notes
- `validTarget`/`ruleMatches` + `cardView` shape are duplicated in **3 places** (match, auth
  `poolCardView`, client `format.ts`) — server authoritative, client highlight-only; sync by hand.
- Reconnect grace is per-username (60s window); a near-simultaneous takeover-vs-disconnect can
  flip the opponent's connected banner out of order (rare, cosmetic).
- Restart-survival is OUT (in-memory match + sessions) — a server restart drops everyone to login.
- Absolutely-positioned board elements anchor to the **padding-box edge** — adding play-area
  padding does NOT move `right:`-anchored items; change their `right:` values.

---

## Commands
```
make dev          # run server + vite, hot reload (open :5173)
make dev-build    # rebuild after go.mod / Dockerfile change
make build-web    # build client → web/static
make prod [PORT=] # build + run prod image
make logs / down / help
```

## Update protocol
After any meaningful change: update **Current state**, **Open/next**, and the **Last updated**
date. This file is the memory between sessions. Append deep detail to `HANDOFF.archive.md` if
it's worth keeping but not current-state.
