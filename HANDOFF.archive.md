# HANDOFF — Vanillastone

Living doc for future sessions. **Keep updated every session.** Read this first.

**Session 2026-06-23 — UI Sessions B + C (card redesign) DONE; #7 art assets LEFT. Client-only
except a small MinionView add; `static/` re-embedded each step.**

- **Card-visual polish:** all gem numbers (card cost/atk/hp + board minion atk/hp) now **white**
  with a dark text-shadow (were near-black, unreadable). Hand cost gem recolours the **DIGIT**
  green (cheaper) / red (pricier) instead of the gem background. **Frozen** minions get a clear
  ❄ badge pinned top-left (`.frost-badge`) — the blue tint alone was too subtle; top-left avoids
  the taunt crest (top-centre) and rarity gem (top-right).
- **Session B — board layout (#5/#6):** mana is now a **vertical crystal column** anchored just
  to the RIGHT of each deck pile and centred on it (`.mana-bar` absolute, `right:36px`,
  `bottom/top: calc(24% + 87px)` + translate; deck pile moved to `right:84px` to open the gap;
  out of the flex flow so growing mana never shifts the board). Play-area side padding `12→44px`.
  End-turn/spectate moved to `right:36px`. **NOTE the gotcha:** absolutely-positioned elements
  anchor to the **padding-box edge**, so adding play-area padding does NOT move `right:`-anchored
  items — you must change their `right:` values. Hands hug the top/bottom edges (mana left the flow).
- **Session C — CardFace full redesign (layout A, then iterated to dev's spec):**
  `web/src/game/CardFace.tsx` rewritten. Layout top→bottom: **full-bleed ART region** with the
  **cost gem + translucent name plate overlaid on it** (name = absolute, dark→clear gradient +
  text-shadow, `pointer-events:none`); a **rules-text box**; a **type band**; atk/hp gems in the
  bottom corners. art:text = **flex 5:4**. Rules font **auto-shrinks by text length** (inline
  style in CardFace: ≤60→10px, ≤95→9px, ≤130→8px, else 7px) so the longest card (Decoy Ward)
  never clips. **Seamless interior** — art + text panels have NO background (just 1px divider
  lines); only the card body shows through. Card size **164×236** (hand/cast/drag/discover/
  mulligan/preview); book-card **143×197**, book widened (builder-page `max-width:1280`,
  grid rows 197). Shared card rule uses **`align-items: stretch`** (was `center`, which left the
  panels narrower than the card — the long-hunted "side gap" bug). Card border thinned 2→1px,
  radius 10→8, inner paddings tightened. Rarity title-colour rules made global (`.name.rarity-*`).
- **C2 — art wiring:** CardFace `<CardArt>` loads `/art/<cardId>.png`, `onError` → placeholder
  type glyph (so missing art = current look; add incrementally, no per-card code). Files live in
  **`web/public/art/`** (Vite copies → `static/` → embedded). See `web/public/art/README.md`.
  **No real art yet — that's #7 (dev generating in Midjourney: 4:3, PNG, named `<cardId>.png`,
  downscale ~420–512px / <150KB, one fixed style prefix).**
- **C3 — board minion hover = full card:** hovering a board minion shows the exact card via
  `CardFace` (`.minion-preview`) instead of the old mini-tooltip; pops above (below for the enemy
  top row); `pointer-events:none` so clicks still hit the minion; **hidden while hovering a keyword
  badge** (`.m-badges:hover ~ .minion-preview`) so the badge's own tooltip stays readable. Needed
  a server add: **`protocol.MinionView` gained `Cost` + `Class`** (set in `view.go minionViews`
  from `mn.card`); `protocol.ts` MinionView mirrors them; `Board.tsx previewCard()` builds a
  `CardView` from the minion (baseCost=cost so no recolour). `go test ./internal/match
  ./internal/protocol` green, gofmt/vet clean.
- **Hero-power fixes:** cost gem was painted behind the 3D flip layer → `z-index:3` + white digit.
  Opponent power tooltip opens **downward** (`.hp-button.static .tooltip`, it's the top row).
  `perspective` on `.hp-button` makes it a stacking context that trapped the tooltip below the
  minions' z-6 gems → `.hp-button:hover { z-index:300 }` lifts the whole button.
- **LEFT:** **#7 real art** (dev/Midjourney). Manual end-to-end playtest (Ice Block reveal,
  mind-control board-cap fizzle, Start-of-Game HP swap, ETC both-hands anthem — all unit-covered,
  never seen live). Deferred bug: phantom "opponent played a card" after mulligan. **This session's
  work is UNCOMMITTED.**

Last updated: 2026-06-23 (**UI tasks (TASKS.md): #1 collection-box fixed size, #4 hero-power flip,
#2 board minions → art-objects, #3 hero/hero-power art slots — all DONE & `static/` re-embedded;
see "UI redesign — TASKS.md" entry below. Remaining UI = Session B (#5 layout / #6 card size) +
Session C (#7 art assets), to be done in a fresh session.** Prior: **HoF finishing session: client
wiring for `control` event + `Immune`
badge done & `static/` re-embedded; 9+ HoF tests added; mapping doc flipped to all-IN; fixed 2
prior-session overclaims (SelfCountAtk never wired, 2nd on_play trigger never fired) + added the
missing `anguished_scribe` — see "HoF finishing session" entry below; pool 139→140.** Phase 10 in
progress: reconnect done; **client UI overhaul + turn
timer + prod deploy done**. **Animations first cut done**. **Direct invites done** (TASKS #1 —
challenge a lobby player, yes/no prompt with deck pick; see "Direct invites"). **Gameplay-feel polish done** — see
"Gameplay-feel polish" below: HS drag-to-play with positional board insert, deck pile sized like
hand cards, opponent-discovering peek, summon-pop for tokens. **Spectator mode done** (watch a
player's POV read-only; see "Spectator mode"). **Card-gem bleed fix + 5s cast reveal done.**
**Card pool rebuilt** (class + rarity fields, color by class not type, renamed staple set, demo
cards purged, hero power → Fire Dart; see "Card pool rebuild"). **Edge triggers Batch 1 + Batch 2
done** — event bus extended (on_turn_start/end, on_spell_cast, on_friendly_summon/death,
on_any_minion_death) + card-generate + all-minion AoE + trigger conditions. **Card pool expanded**
**Classic pivot**: scope locked to clone the HS **Classic set** — 105 neutral + 15 Mage = 120
collectible (Mage-only game; original names/art, 1:1 mechanics). Mage trimmed to its 9 in-Classic
clones (6 remaining are mechanic-gated). **Neutral audited** to the Classic-105: trimmed to **36
faithful clones** of 105. **Wave 1 done (2026-06-22)**: cloned the 4 buildable-now Classic
neutrals (Doomsayer→ruin_oracle, Hogger→snarlmaw, Onyxia→emberwing_matron, Arcane
Devourer→managlutton) + 2 stat FIXes (veiled_assassin 5/4→7/5, hornelder_heir token 4/5→5/5);
pure data, no engine change. **Phase D — enrage done (2026-06-22)**: `Card.Enrage *Aura`, atk
bonus while `health < maxHP()` (computed live in `atk()`, no recompute hook; silence-cancelled),
💢 client badge; 3 cards. **Phase E — grant-keyword + adjacency done (2026-06-22)**:
`enchant.keywords` + `Effect.Grant`, multi-target `EffectBuff`, `AreaAdjacent`/`AreaSplash`
+ `adjacentRefs` (battlecry self-anchors on the played slot, splash spells on the target);
3 cards (wardstone_sentinel, bannerguard, frostshear); no web change. **Card-blitz waves W1-W3
(2026-06-22)** toward finishing the set: tribe tags (`Card.Tribe` + protocol/view + client
badge), elusive + can't-attack keywords, temp "this turn" buffs (`enchant.temp`, cleared at
turn end), bounce (`EffectBounce`), and 3 new triggers (`OnHeal`/`OnSecretPlayed`/`OnPlayCard`)
+ `TargetRandomFriendly`. 12 neutral cards added; web rebuilt (elusive targeting, bounce event,
badges). Pool now **67** (57 neutral / 10 mage). **Refactor (2026-06-22):** `match.go` (2160 lines)
split into 6 files, same package, zero behavior change — `match.go` (consts/types/Match/lifecycle),
`actions.go` (PlayCard/Attack/HeroPower/Concede/Choose/Mulligan), `state.go` (seats/reconnect/deck/
snapshot-send/spectators), `engine.go` (triggers/secrets/summon/discover), `effects.go` (applyEffect/
damage/heal/buff/silence/auras/targeting-resolution), `view.go` (death resolution + snapshot builders).
**UI polish (2026-06-22):** cards +25% (in-game + collection + mulligan native size), rarity gem
moved to top-right corner + now shown on board minions, tribe shown as the card's type line
(`tribeLabel`; 15 minions tagged), stat gems recolor the NUMBER not the background. **Known bug
(deferred):** rare phantom "opponent played a card" right after mulligan with no action — suspect a
stale resync/`history` replay across matches (client only clears log on `match_start`); not yet
reproduced. **Special-BC sub-wave 1 done (2026-06-22):** 5 server-only special/conditional
battlecries cloned — `EffectMissiles` (random split → powder_tosser), `EffectKillSecret`
(→ covert_saboteur), `AreaRandomEnemyMinion`+`Effect.MaxAttack` (random cond-destroy →
trampling_brute), `Effect.PerCardInHand` (hand-count buff → duskscale_drake), and a TargetSelf
battlecry self-anchor in `PlayCardAt` (self-dmg → wounded_duelist). No client/web change
(untargeted/random; `static/` untouched). Pool 67→72. **Special-BC sub-wave 2 done (2026-06-22):**
target-CONDITION mechanic — `Effect.ReqAttack`/`ReqTaunt` + `targetCondOK` + `hasLegalTargetFor`
(replaced the now-removed `anyLegalTarget`/`targetableEnemyMinions`), condition fields on `CardView`,
client `condMet` highlight — plus `EffectCopy` (`copyMinion`). 3 cloned: trophy_hunter (destroy 7+
atk), grave_knight (destroy enemy Taunt), visage_thief (copy a minion). **Client rebuilt + `static/`
re-embedded.** Costs use real HS-Classic (BGH 5, Black Knight 6; mapping doc had 4 — fixed). Pool
72→75 (65 neutral / 10 mage). **Special-BC sub-wave 3 done (2026-06-22):** `EffectSwapStats` (swap
a minion's atk/health via a derived-stat enchant delta — 0-attack swap kills) + `EffectConsumeShields`
(strip all Divine Shields both boards, self-buff +3/+3 per shield, self-anchored). 2 cloned:
addled_brewer (Crazed Alchemist), crimson_reaver (Blood Knight). Server-only, no client/web change.
Pool 75→77 (67 neutral / 10 mage). Special/conditional-battlecry bucket complete bar 2 tribe/
random-pool-gated cards. **Weapon-manip done (2026-06-22):** `weaponInst` gained a mutable instance
`attack` (heroAttackValue/weaponView now read it, not the card); 4 untargeted battlecry effects —
`EffectGainWeaponAttack` (self-anchored buff = weapon atk), `EffectChipWeapon` (opp durability −N,
break at 0), `EffectBuffWeapon` (own weapon +atk/+dur), `EffectDestroyWeapon` (destroy opp weapon,
draw = its durability). 4 cloned: tidereaver, brine_cutter, captain_brackwater, relic_breaker.
Server-only, no client/web change. Pool 77→81 (71 neutral / 10 mage). **Tribe-auras done
(2026-06-22):** aura engine extended — `auraTargets` (tribe + adjacency filters), **HP auras** via
`minion.auraHP` + maxHP + a current-health delta in `refreshAuras` (gaining max heals, losing it only
clamps current down — a damaged minion never dies), `AreaFriendlyTribe` battlecry-buff area +
`Effect.Tribe`, `Trigger.SubjectTribe` summon gate. 5 cloned: reef_warchief (Murloc Warleader, +2
atk), tidehook_captain (Southsea Captain, +1/+1 HP-aura), fang_alpha (Dire Wolf Alpha, adjacency),
tidescry_oracle (Coldlight Seer, BC tribe-buff), brackish_caller (Murloc Tidecaller, summon trigger).
No client/web change (auras render from snapshot). **Special-legendary AoE batch done (2026-06-23):**
`AreaAllCharacters` + `AreaOtherCharacters` (fireTriggers now self-anchors area edge triggers) +
`Effect.SummonForOpponent`. 3 cloned: cinder_baron (Baron Geddon, end-turn 2-dmg all-other), rotgut_horror
(Abomination, DR 2-dmg all), the_gorehound (The Beast, DR summon 3/3 for opponent; +gorehound_whelp token).
No client/web change. **Accounting:** code `DeckPoolIDs()` = 89 collectible, **88 fill Classic slots**
(78 neutral + 10 mage); `sapphire_drake` is a Phase-A invented extra, not a Classic-105 card. **88/120
done, 32 left** (27 neutral + 5 mage GATE); count Classic progress from mapping IN-rows, not raw pool
size. NEXT = cost-mod Phase F (big, invasive — own session), remaining special legendaries (Alexstrasza
set-hp needs a hero-target rule + client; Deathwing destroy-all+discard; Gruul both-turns; Whitemane
died-this-turn; Nozdormu global; Ysera/Brightwing random-gen), Mage leftovers, + 2 tribe/random-pool
special BCs — toward 105 neutral + 15 mage. See TASKS.md + .notes/classic-mapping.md.)

**Waves 1–3 done (2026-06-23):** (1) special legendaries — voidwyrm_tyrant (Deathwing:
`AreaOtherMinions` + `Effect.DiscardHand` + `discardHand`), cragmaw (Gruul: global `OnAnyTurnEnd`
fired alongside `OnTurnEnd` in `endTurnLocked`; reactors = both boards), revenant_priestess
(Whitemane: `playerState.diedThisTurn` recorded per death in `resolveDeaths`, cleared both sides in
`startTurn`, `EffectResummonDead` summons base cards). (2) random-generation — `cards.RandomGenPoolIDs(
class,type,rarity,tribe)` filter over the collectible set + `EffectGenerateRandom` / `EffectSummonRandom`
/ `EffectTransformRandom` (transform a random OTHER minion into a random token from `Effect.GenIDs`);
`transformMinion` helper extracted from `EffectTransform`. Cards: codex_of_insight (mage Tome), gleamwing
(Brightwing), wilds_beastcaller (Barrens Stablehand), sprocket_tinkerer (Tinkmaster, +thornback_saurian
/bramble_squirrel tokens). **Ysera deferred** (Dream sub-pool needs delayed-destroy + self-excluded AoE
clones; can't invent substitute Dream cards under the 1:1 rule). (3) set-hp — emberqueen_valtha
(Alexstrasza): new `TargetHero` rule (either hero, not minions) + `EffectSetHealth` (direct set, clamped
to 30, no armor/heal-trigger interaction) + `sethealth` event; client got `'hero'` in
TargetRule/ruleMatches + a sethealth log line + icon, **`static/` rebuilt + re-embedded**. Classic
**88→96** (85 neutral + 11 mage), pool `DeckPoolIDs()` 89→97; 24 Classic left (20 neutral + 4 mage).
10 tests added; `go test -race ./...` green, gofmt/vet clean. **NEXT = cost-mod Phase F (own session,
invasive — plan first).**

**Wave 4 — cost-mod Phase F done (2026-06-23):** per-card hand-cost layer. `m.effectiveCost(pi, card)`
= base + intrinsic `Card.CostRule` (Sea Giant: −1 per board minion) + every in-play `Card.CostAura`
(`Scope` friendly/all, `Type` filter, `FirstMinionEachTurn` gated on `playerState.minionsPlayedThisTurn`)
+ Kirin Tor's `playerState.nextSecretFree` override (Secret → 0), floored at 0. Computed fresh on each
read — no stored per-card state. Threaded through `PlayCardAt` (single `cost := m.effectiveCost(...)`
used by every mana gate + spend) → `playSpell`/`playSecret` now take a `cost` param; minion play
increments `minionsPlayedThisTurn` AFTER cost is locked (so Pint-Sized discounts the minion being
played); `playSecret` consumes `nextSecretFree`; both flags reset in `startTurn`. `EffectFreeNextSecret`
sets the flag. Snapshot: `selfView` is now a method `m.selfView(pi, name)` so hand cards carry the
effective `Cost` + printed `BaseCost`. Client: affordability + CardFace already read `cost` (now the
effective value); added `baseCost` to the wire type + a cost-gem recolour (green cheaper / red pricier),
`static/` rebuilt + re-embedded. 5 cards: mana_leech (Mana Wraith), pocket_conjurer (Pint-Sized
Summoner), tidecolossus (Sea Giant), arcane_adept (Sorcerer's Apprentice), spellwarden_magus (Kirin
Tor). **Millhouse deferred** (enemy spells cost 0 NEXT turn — cross-turn opponent-side flag). Classic
**96→101** (88 neutral + 13 mage), pool 97→102; **19 Classic left** (17 neutral + 2 mage: Icicle
conditional, Spellbender retarget-secret). 5 tests + `TestClassicMechanicsHaveCards` cost-mod assert.
`go test -race ./...` green, gofmt/vet clean.

**Wave 5 — conditional / tribe-cond / cross-turn cost done (2026-06-23):** 3 cards.
(1) `Effect.DrawIfFrozen` — EffectDamage captures the target minion's Frozen state BEFORE damage, then
draws after (the minion may die) → glacial_splinter (Icicle, mage: 2 dmg a minion, if Frozen draw).
(2) `Effect.ReqTribe` target-condition (added to `targetCondOK` + `CardView`/client `condMet`) +
`Effect.SelfBuffAtk/SelfBuffHP` battlecry rider (applied to the played minion `mn` after the battlecry
resolves in `PlayCardAt`, only when it didn't fizzle) → shellback_crab (Hungry Crab: destroy a Murloc,
gain +2/+2). (3) `EffectEnemySpellsFree` + `playerState.spellsFreeOnTurn` — Millhouse sets the
opponent's flag to `turnNum+1`; `effectiveCost` returns 0 for that player's spells when
`spellsFreeOnTurn == turnNum` (guarded by `!= 0`, since 0 is the unset default and turn 0 is real — a
collision caught in test) → fizzle_sparkmuddle (Millhouse). Client: `reqTribe` on the wire type +
`condMet` tribe check, `static/` rebuilt + re-embedded. Classic **101→104** (90 neutral + 14 mage),
pool 102→105; **16 Classic left** (15 neutral + 1 mage: Spellbender, retarget-secret). 3 tests.
`go test -race ./...` green, gofmt/vet clean.

**Wave 6 — Spellbender + misc one-offs done (2026-06-23): MAGE SET COMPLETE (15/15).** 4 cards.
`SecretRetargetSpell` + `SecretDef.Summon` + `secretCtx.spellTarget` (a `*charRef` the secret rewrites
in place) + `secretFires(defender, s, ctx)` precondition (fires only when the enemy spell targets one of
the owner's minions); `playSpell` now passes `secretCtx{spellTarget: &ref}` so the in-flight spell
redirects onto the summoned decoy and resolves against it (not a cancel) → decoy_ward (Spellbender,
+conjured_decoy 1/3 token). `EffectGiveOppMana` (opponent maxMana +1, capped at 10, current mana
unchanged) → runed_golem (Arcane Golem, Charge). Pure-data, no engine change: nightmare_lord (Xavius:
`OnPlayCard` → summon thornwood_satyr 2/1) + imp_warden (Imp Master: TWO `OnTurnEnd` triggers on one
card — self-damage `TargetSelf` + summon imp_whelp; `fireTriggers` already runs every matching trigger).
Tokens: conjured_decoy, thornwood_satyr, imp_whelp. **Server-only — no client change.** Classic
**104→108** (93 neutral + 15 mage), pool 105→109; **12 Classic left (all neutral)**. 5 tests.
`go test -race ./...` green, gofmt/vet clean.

**Wave 7 — grant-SpellDmg / give-opp-cards / rng-trigger / copy-spell done (2026-06-23):** 4 neutral.
`enchant.spellDamage` field + `spellDamageOf` now sums enchant grants + `Effect.GrantSpellDamage`
(EffectBuff rider, rides in an enchant so Silence strips it and it adds to `spellPower`) → runeward_sage
(Ancient Mage: adjacency grant). `Effect.ToOpponent` + `Count` on EffectGenerate (loop, opp hand, burn
if full) → grovelord_brakka (King Mukla, +jungle_gift 0/1-cost +1/+1 token). `Trigger.Chance` (percent;
`fireTriggers` rolls `rng.Intn(100) >= chance` to skip) → lucky_angler (Nat Pagle: 50% turn-start extra
draw). `Card.CopiesSpells` + `m.copySpellToOpponent(caster, card)` called in playSpell on BOTH the normal
and countered paths (every non-silenced copier in play adds a copy of the cast spell to the non-caster's
hand) → archivist_solenne (Lorewalker Cho). **Server-only — no client change.** Classic **108→112**
(97 neutral + 15 mage), pool 109→113; **8 Classic left (all neutral)**. 4 tests. `go test -race ./...`
green, gofmt/vet clean.

**Waves 8–10 — CLASSIC SET COMPLETE 120/120 (2026-06-23).** All server-only.
**Wave 8** (4 cards): `CostAura{friendly,+3}` → tollkeeper_brute (Venture Co); `CostRule.PerOwnWeaponAttack`
→ dread_buccaneer (Dread Corsair); `Card.EnrageWeaponAtk` summed in `heroAttackValue` (damaged minion →
weapon +N) → grudge_smith (Spiteful Smith); `Card.TurnSeconds` + `m.activeTurnDuration()` (scheduleTurnTimer
uses the smallest in-play TurnSeconds) → chronlord_zhal (Nozdormu).
**Wave 9** (2 cards) — **attack-model refactor:** `minion.attacksLeft` (remaining) → `attacksMade` (used);
eligibility is now `attacksMade < attacksPerTurn()` read LIVE, so Windfury gained mid-turn grants the 2nd
attack (the silence windfury-clamp is gone, no longer needed). `Card.EnrageGrant []Keyword` checked in
`has()` while `enraged()` → moonfury_stalker (Raging Worgen: +1 atk & Windfury while damaged).
`Card.ChargeWithWeapon` + `m.hasCharge(mn)` → tideblade_raider (Southsea Deckhand). To read the weapon,
`canAttack`/`canAttackHero`/`minionViews`/`oppView` were converted from free funcs to **Match methods**
(callers updated in actions/view/match/state).
**Wave 10** (2 cards + 5 Dream tokens): `EffectSwapWithHand` (swap source minion with a random hand minion,
in-place transform, no battlecry; bot returns to hand) → clockwork_swapbot (Alarm-o-Bot). `EffectGenerateRandom`
now also accepts an explicit `GenIDs` pool → dreamwarden_ylena (Ysera) adds a random **Dream token**. Dream
pool needed two new bits: `Effect.DestroyNextTurn` (EffectBuff rider sets `minion.destroyAtTurnStart`, killed
in `startTurn`) for Waking Nightmare, and `Effect.ExceptCardID` (skips named minions in an all-character AoE)
for Emerald Reckoning (spares Yseras). Classic **112→120** (105 neutral + 15 mage — **WHOLE SET DONE**),
pool 113→121. 14 tests. `go test -race ./...` green, gofmt/vet clean. **The card-clone phase is complete;
next work is balance/playtest/polish or a scope expansion — discuss before starting.**

---

**Deckbuilding — class-bound decks + filters + curve done (2026-06-23):** decks now bind to a
**class** (server-authoritative). `cards.ValidateDeck(ids, class)` gained a class param: rejects a
non-playable class (`PlayableClasses()` = `{mage}` for now; `ClassHunter` reserved/card-less) and any
card that is neither neutral nor that class. SQLite `decks` gained a `class TEXT NOT NULL DEFAULT
'mage'` column (CREATE + guarded `ALTER TABLE … ADD COLUMN` migration in `store.Open`, ignoring the
"duplicate column name" error); `store.CreateDeck`/`UpdateDeck`/`Deck`/`scanDeck` + all SELECTs carry
class. `decks_http` deckBody/deckJSON carry `class`; `decodeDeck` passes it to ValidateDeck; transport
deck-pick (`deckFor`) validates with `d.Class`. `/pool` now also returns `classes` (playable list). 3
tests added (`TestValidateDeckClass`, store/http class round-trips); existing deck tests updated for the
new signatures. Client: new-deck flow opens a **class picker** (Mage enabled, Hunter "coming soon" from
`pool.classes`); collection filtered to the deck's class + neutral with **tabs** (All / class /
Neutral), single-select **mana** (0–6, 7+) and **rarity** filters, **ascending-mana sort**, and a
**mana curve** histogram in the deck-edit panel. `api.ts` Deck/Pool types + create/update calls carry
class. `go test -race` green (store/cards/auth), gofmt/vet clean, web tsc+vite clean, `static/`
re-embedded.

**Deckbuilder polish + legendary cap + viewport sizing done (2026-06-23):**
- **Legendary 1-copy** (HS rule): `ValidateDeck` caps legendaries at 1 copy (others at `MaxCopies`);
  `DefaultDeck` rebuilt to respect it (it doubled every card before — would now be illegal). Client
  `capFor(card)` gates add + book-card disable. Test `TestValidateDeckLegendaryCap`.
- **Deckbuilder UI**: filtered **card count** beside the tabs; owned badge shows **×N** (bigger, raised
  z-index); in-deck list names are **rarity-coloured** (rare blue / epic purple / legendary gold),
  matching the card faces.
- **Rarity on card faces**: rarity gem REMOVED from `CardFace` (it overlapped the selected-card
  outline); rarity now shown by **title colour** instead (`.name.rarity-*`). Board minions
  (`Board.tsx`) keep their own gem — separate surface, untouched.
- **Long-title fix**: a long single-word name overflowed under the cost gem. Fixed in `CardFace` by
  shrinking the title font from the longest word's length (≥13 chars → 9px, 12 → 10px, else default);
  multi-word names wrap at spaces. No mid-word breaks. Scales on every surface (cast preview ×1.4 etc.).
- **Gem centering**: `.card .cost`/`.stat` + `.mana-pip` switched from `line-height` to flexbox
  centering (digits were off-centre).
- **Edge-trigger pulse**: recoloured gold→**blue ring + glow**, swelling low→high→back twice (~1.2s),
  raised z-index. Already wired server-side (`fireTriggers` emits a `trigger` event with the reacting
  minion as source); client adds/removes `.trigger-pulse`.
- **Viewport sizing system** — one `--u` knob set by `App` from the viewport
  (`min(1, vw/2056, vh/1329)`, floored 0.65; **2056×1329 = the dev's display, the u=1 baseline**).
  Two bias knobs on `:root`: `--ui-bias` (1.5) for the centred **chrome** screens (login / lobby /
  invite modal / deckbuilder), applied as `transform: scale(calc(var(--u) * var(--ui-bias)))` — a
  centred card/page has no void so a whole-element scale is faithful; `--game-bias` (1.1) for the
  **game board**, which is full-bleed and would void/letterbox under a fixed-canvas scale (tried,
  reverted). The board instead uses **inflate-then-scale** on a `.game-stage` wrapper:
  `width/height: calc(100% / (var(--u) * var(--game-bias)))` + `transform: scale(...)` → fills the
  viewport exactly, gives the board more layout room on small screens, scales uniformly. Drag clone +
  targeting arrow stay screen-space (pointer math untouched); drag clone scaled by the same factor.
  **Tuning: bump `--game-bias` past ~1.1 and the dense board overflows vertically (clips the hand).**
  No Go change; web tsc+vite clean, `static/` re-embedded.

**Curated default deck + online-players panel done (2026-06-23):**
- **Default deck** (`cards.DefaultDeck`) is now a hand-curated, smooth-curve **freeze/tempo
  Mage** list (`defaultMageDeck` slice) instead of "first 15 pool ids ×2": cheap spell-synergy
  bodies + removal early, sticky 4-drops (Divine Shield / Taunt), Frost Tempest board clear up
  top, one legendary (Emberforge Magus). Cap-legal (`TestDefaultDeckIsLegal` still green).
  `DefaultDeck()` returns a defensive copy. Designed to be per-class later: comment notes adding
  one curated default per class and picking by class. Client labels the option **"Default Mage
  Deck"** (lobby deck-pick + invite prompt) so the class is visible — future classes get their own.
- **Online-players panel**: moved the player list OUT of the centered, transform-scaled
  `.lobby-card` into a dedicated **fixed top-right `.player-panel`** (`<aside>`): own header with
  live count, a **search filter** (`playerFilter` state, name `includes`), and a flex
  internally-scrolling `.player-list` (`max-height: calc(100vh - 32px)`, `flex:1; overflow-y`).
  Scales with `--u` from `transform-origin: top right`. A 15+ player list now scrolls inside its
  own full-height column and never pushes the lobby card's controls off-screen. Empty-filter row
  shown when no name matches. web tsc+vite clean, `static/` re-embedded.

**Hall of Fame scope-add + Wave HoF-1 done (2026-06-23):** the "Classic set" we cloned omitted
the Classic cards later rotated to **Hall of Fame** (Sylvanas, Ragnaros, Leeroy, Azure Drake,
Mountain/Molten Giant, Mind Control Tech, Ice Block, …). Catalogued all **16 HoF neutrals + 2 Mage
+ Black Cat** in `.notes/classic-mapping.md` (new "HALL OF FAME" section) with original ids +
buildability status + a 3-wave plan. **Wave HoF-1 (pure data, no engine change):** 4 cards added to
`neutral.go` — `hexbreaker_warden` (Spellbreaker, 4/3 BC silence), `cobalt_loreling` (Azure Drake,
4/4 Dragon SpellDmg+1 + BC draw), `reckless_vanguard` (Leeroy, 6/2 Charge + BC summon two
`emberwing_whelp` 1/1 for the OPPONENT), `emberlord_vrakgar` (Ragnaros, 8/8 CantAttack + OnTurnEnd 8
dmg to a random enemy). 1 test (`TestEmberlordEndTurnBurn`). Also fixed a latent fragility in
`TestValidateDeckClass` (it picked a Mage card via nondeterministic map iteration then prepended it
ignoring the copy cap — the new curated default deck reliably tripped it; rebuilt to construct a
deterministic legal deck via `DeckPoolIDs`). Server-only — no client/web change.

**Hall of Fame Waves HoF-2 + HoF-3 — ALL CARDS ADDED, server logic done (2026-06-23):** the rest of
the Hall of Fame set is in. Builds clean, `gofmt`/`vet` clean, `go test ./internal/cards ./internal/match`
green. All server-side; client renders via existing snapshot/event handling (two new wire bits not yet
read by the client — see "left" below).

Cards added (ids; real names ONLY in `.notes/classic-mapping.md`):
- **HoF-2:** `brineseer_diviner` (each player draws 2 — `EffectDraw` gained `ToOpponent`, used as two
  battlecries), `crag_colossus` / `magma_behemoth` (`CostRule.PerCardInHand` / `PerMissingHealth`),
  `frostlance` (`Effect.FrozenDamage`: frozen target → 4 dmg instead of freeze), `anguished_scribe`
  (new `OnDamage` trigger fired inside `damageMinion` on the damaged minion), `brinelord_gorrak`
  (`Card.SelfCountAtk{Tribe,Atk}` summed in `refreshAuras` — +1 atk per other Murloc, silence-cancelled).
- **HoF-3:** `mesmer_adept` + `wraithqueen_selvara` (`EffectMindControl` — move a random enemy minion
  across boards, summon-sick; `Effect.ReqOppMinions` gates the 4+ condition), `corsair_macaw`
  (`EffectTutorTribe` — draw a random card of a tribe from deck), `shadowtail_familiar`
  (`Effect.ReqDeckAllOdd` gate on `EffectDraw` + `deckAllOdd` helper; SpellDamage+1),
  `duskwarden_genmar` + `lunar_devourer` (Start-of-Game via `m.applyStartOfGame(pi, deck)` in `New`:
  all-even deck → hero power Cost 1; all-odd deck → `cards.UpgradedMageHeroPower()` deal-2),
  `frostward_aegis` (`SecretKind` `SecretIceBlock` + `playerState.immune`; `damageHero` hook
  `tryIceBlock` prevents a fatal hit and sets immune; `OnFatalDamage` event exists only so the secret
  is NEVER dispatched via `triggerSecrets`; immune cleared both sides in `endTurnLocked`),
  `warhorn_chieftain` (ETC — two `EffectGenerateRandom` battlecries self+`ToOpponent`, pool
  `warhornAnthems`), `gearmaster_cog` (Mekkatorque — `EffectSummonRandom` now honors `GenIDs`, pool
  `cogContraptions`).
- **Tokens added:** `anthem_muster`/`anthem_warsong`/`anthem_ambush` (+ `Effect.CountMax` for random
  3–5 summon, `Effect.ThenDraw` for damage-then-draw), `tide_recruit`, `warsong_grunt`/`warsong_reaver`,
  `cog_emboldener`/`cog_beacon`/`cog_polymorpher`/`cog_mender` (+ `cog_chick`). `cog_mender` needed a
  `TargetFriendlyHero` anchor case in `fireTriggers` (heals its controller's hero).

**HoF finishing session — LEFT-TO-DO cleared + 2 engine bugs fixed + 1 missing card added (2026-06-23):**
1. **Client rendering (web/) — DONE.** `control` event kind added to `format.ts` (formatEvent line
   "takes control of …" + 🧠 icon) and its kind added to the `Event` union in `protocol.ts`.
   `PlayerView.Immune` added to `protocol.ts`; `Hero.tsx` shows a 🛡️ `.immune-badge` + gold-glow
   `.heropanel.immune` ring; CSS added. `pnpm build` clean, **`static/` rebuilt + re-embedded**.
2. **Tests — DONE** (9 new in `match_test.go`): mind-control + ReqOppMinions gate
   (`TestMesmerAdeptMindControlGate`), full-board fizzle (`TestMindControlFizzlesOnFullBoard`),
   Ice Block prevent-lethal + immune + clears-at-end-of-turn (`TestIceBlockPreventsLethalThenClears`),
   FrozenDamage (`TestFrostlanceDamagesFrozen`), draw-both (`TestBrineseerDrawsBoth`), PerMissingHealth
   cost (`TestMagmaBehemothCostByMissingHealth`), SelfCountAtk (`TestBrinelordSelfCountAttack`),
   tutor-by-tribe (`TestCorsairMacawTutorsPirate`), start-of-game parity (`TestStartOfGameDeckParity`),
   + OnDamage draw (`TestAnguishedScribeDrawsOnDamage`). `go test -race ./...` green, gofmt/vet clean.
3. **Mapping doc — DONE.** All HoF rows flipped to **IN** in `.notes/classic-mapping.md` (+ a status note).
4. **⚠️ Two prior-session OVERCLAIMS found while writing tests (now fixed):**
   - **`SelfCountAtk` was never wired** — declared in card data, claimed "summed in `refreshAuras`",
     but the engine never read it (`brinelord_gorrak` was a vanilla 2/4). Now summed in `refreshAuras`
     via a new `m.countTribe(tribe, except)` helper (counts OTHER in-play minions of the tribe on BOTH
     boards), silence-cancelled by the existing source guard.
   - **A minion's 2nd `on_play` trigger never fired** — `PlayCardAt` ran only `card.Battlecry()` (the
     first OnPlay). So `brineseer_diviner` drew only for the caster and `warhorn_chieftain` gave only
     itself an anthem. `PlayCardAt` now also fires the extra untargeted OnPlay triggers
     (`card.TriggersFor(OnPlay)[1:]`, skipping any that need a target).
   - **`anguished_scribe` was never actually added** — the `on_damage` engine path existed
     (`damageMinion` fires the damaged minion's OnDamage triggers) but no card used it. Added as a
     3-cost 1/4 with an OnDamage draw trigger (pure data). Collectible pool 139→**140**.
5. **Full HoF test sweep (follow-up, same day):** added 5 more tests covering the HoF cards that had
   data but no test, to guard against further silent gaps — `TestCragColossusCostByHand` (PerCardInHand),
   `TestWarhornChieftainGivesBothAnthems` (both hands get a random anthem — exercises the multi-on_play
   fix), `TestGearmasterCogSummonsContraption` (random Contraption token), `TestSelvaraDeathrattleSteals`
   (OnDeath mind-control move), `TestShadowtailOddDeckDraw` (ReqDeckAllOdd gate, both branches). **Every
   HoF card now has explicit test coverage** (HoF-1's four reuse already-tested mechanics; emberlord has
   `TestEmberlordEndTurnBurn`). `go test -race ./...` green, gofmt/vet clean.
6. **Still NOT done — manual end-to-end playtest** (item 3 of the old list): Ice Block reveal animation,
   mind-control board-cap fizzle in a real match, Start-of-Game hero-power swap, ETC giving both hands a
   random anthem. All are unit-covered but not yet observed in a live two-client game.

**UI redesign — TASKS.md (2026-06-23):** TASKS.md was split into reviewable slices. Client-only
(React/CSS), no Go change; `go build`/web `tsc`+`vite` clean, `static/` re-embedded each step.
- **#1 collection box** (`Deckbuilder.tsx` + CSS) — a short last page no longer shrinks/reflows the
  book: `.book-grid` reserves 2 full rows (`grid-template-rows: repeat(2,190px)`), short pages are
  padded with invisible `.book-card.placeholder` cells, the no-match message spans the full height.
- **#4 hero-power flip** (`GameScreen.tsx` + CSS) — the power is a 3D flip (`.hp-flip` inner with
  `.hp-face.front`/`.back`, `backface-visibility:hidden`); `.hp-button.used` (driven by server
  `heroPowerUsed`) rotates to a face-down "↺" exhausted back, flips back when the power refreshes on
  your turn. Both heroes. Cost gem stays put (outside the flip). Old ✓ badge removed.
- **#2 board minions = art objects** (`Board.tsx` + `.minion` CSS, full rewrite) — portrait
  (placeholder `cardArtIcon`), rarity gem, atk/hp gems, small rarity-coloured `.m-nametag`; NO
  description body, NO cost. Full card info on hover via `CardTooltip` (`.minion.enemy .tooltip`
  opens downward for the top row so it doesn't clip; tooltip `z-index:200`). Keyword visuals after
  several feedback rounds: **Taunt** = a metal SHIELD CREST pinned to the top edge (`.taunt-frame`,
  `clip-path` pentagon) + a steel inner rim on `.m-portrait` (object stays a rect); **Divine Shield**
  = a crisp shimmering gold ring (`.ds-aura`); **Frozen** = icy sheen on the portrait. Action state
  (`ready`/`selected`/`targetable`) uses `outline`+`outline-offset` so the green/red ring is never
  hidden by a frame/aura; gems/title/rarity-gem are `z-index:6` (above frame z3 / aura z4). Earlier
  attempts (full stone frame; shield-shaped via `border-radius`) were rejected for hiding the action
  outline / the numbers — DON'T reshape the object box; keep cues as overlays + outline.
- **#3 hero & hero-power art slots** — `.portrait` is now a bigger (92px) double-ring framed art slot
  (placeholder 🧙); hero power is a framed art circle (see #4). Ready for real art.
- **LEFT (own session):** **Session B** = #5 (mana vertical bottom-right + opponent mirror; hand
  spacing to top/bottom edges) + #6 (bigger cards to fit art — Decoy Ward has the longest text, size
  so art+text fit); touches the whole board layout — do together, after eyeballing A. **Session C** =
  #7 real art assets (CC0 / generated) + a wiring system to drop them into the portrait/art slots.
  All current art is placeholder icons.

**⚠️ HARD RULE FOR ALL CODE (new session, READ FIRST):** NEVER put Blizzard / Hearthstone / Warcraft
card or character names in source code OR comments OR tests — public repo, git history is forever. Use
our original card **ids** (e.g. `wraithqueen_selvara`) or a generic mechanic description in comments.
Real names live ONLY in `.notes/classic-mapping.md` (gitignored). This session a prior pass had leaked
real names into Go comments; they were scrubbed to backticked ids — keep it that way. (This is also a
memory: see `no-ip-names-in-code`.)

## What this is
Hearthstone-*inspired* digital card game. **Custom** cards/art/names — NOT Blizzard's
IP. Goal: build a HS-like fast, play 1v1 with friends. Solo project.

Repo: `github.com/amvid/vanillastone`. Will be public on GitHub.

---

## Locked decisions

| Topic | Decision |
|---|---|
| Server | Go, **authoritative** (owns all truth, client is dumb renderer) |
| Transport | WebSocket + JSON |
| Client | **Web** (React + Vite), NOT Godot |
| Hosting | One Go binary serves client (`go:embed` dist) + runs `/ws` |
| Accounts | **Basic accounts** — username + password. Unique username. bcrypt hash. Register + login. No email/reset/roles. (revised 2026-06-21 from "no accounts") |
| Sessions | In-memory `token -> username` map. POST /login issues random 256-bit token. ws auths with it. Dies on restart (re-login) |
| Storage | SQLite via **`modernc.org/sqlite`** (pure Go, no CGO — keeps distroless static prod build). Users now; decks later. Match state in-memory |
| Cards | All players have all cards. Custom set |
| Mode | 1v1 only. No AI |
| Hero | Mage only (for now) |
| Scope | Cover **all mechanic TYPES**, skip vanilla HS cards |
| RNG | Server-side, seeded |
| Match persistence | In-memory (dies on restart) — OK for now |

### Repo layout (target)
```
/cmd/server     Go entrypoint (exists, stub)
/server | /internal   engine, ws, http, sqlite (TBD)
/web            React + Vite client -> builds to /web/dist (NOT scaffolded yet)
/cards          shared JSON card defs, server source of truth (TBD)
Dockerfile      dev (wgo) + build + prod (distroless) targets
docker-compose.yml
Makefile
```

---

## Legal rules (public repo)
Mechanics NOT copyrightable — engine is fine. Assets + names ARE.
- **Card design:** card numbers/effects faithfully follow well-established genre staples;
  names + art are wholly original. We rebuild a familiar card pool rather than invent balance.
  Mechanics aren't copyrightable; names/art/flavor are — so never ship any third-party ones.
- **Never commit**: Blizzard art, sounds, fonts, card/hero names, flavor text, logos,
  datamined files. Git history is forever.
- **Safe**: own card names + art, CC0/CC-BY assets (Kenney, game-icons.net).
- Don't name it "Hearthstone". `LICENSE` (MIT) + `ASSETS.md` (track every
  3rd-party asset + license) done. No commercial intent (personal/hobby).

---

## Architecture
```
Web client (renderer)  <--WebSocket/JSON-->  Go server (authoritative)
  sends: PlayCard, Attack, HeroPower,          - match engine (state machine)
         EndTurn, Target, Mulligan             - event bus + effect resolver
  recv:  GameState snapshot + ordered          - seeded RNG
         event log (for animation)             - lobby / matchmaking
                                               - SQLite (decks)
```
Key: server resolves instantly, sends **ordered event list** + resulting state.
Client replays events as animation, settles to state. Full snapshot per action now;
delta optimization later. Client swappable because server owns truth.

### Effect system
Hybrid: common effects (damage/draw/buff) data-driven JSON params; weird ones = Go
handler by ID. Start code-handlers, extract data DSL when patterns repeat.

---

## Mechanic taxonomy ("all types" = the real spec)
**Card types**: Minion (attack/health), Spell (one-shot), Weapon (durability),
Hero Power (Mage Fireblast: 2 mana 1 dmg, reusable/turn).

**Keywords**: Battlecry, Deathrattle, Taunt, Charge/Rush, Divine Shield, Windfury,
Spell Damage, Secret, Aura, Freeze, Silence, Stealth, Poisonous, Lifesteal, Overload,
Discover/Choose.

**Triggers (event bus)**: on_play, on_death, on_damage, on_heal, on_turn_start,
on_turn_end, on_attack, on_summon, on_spell_cast, on_card_draw.

Cover these = engine complete. Custom cards = data on top.

---

## Build phases (order)
1. **Skeleton loop** — connect, username, lobby, 2 players matched, EndTurn ping-pong. Proves transport + sync. ✅ DONE
2. Minions + combat (mana, attack, summon sickness, hero hp, win/lose). ✅ DONE
3. Spells + targeting (validation). ✅ DONE
4. Event bus + triggers (battlecry, deathrattle) — the big one. ✅ DONE
5. Keywords wave 1 (taunt, charge/rush, divine shield, freeze). ✅ DONE
6. Keywords wave 2 (spell dmg, aura, silence, stealth, poisonous, lifesteal, windfury). ✅ DONE
7. Secrets + Discover (hidden state, mid-action prompts). ✅ DONE
8. Hero power + weapons. ✅ DONE (Overload deferred to Shaman)
9. Decks + SQLite (deckbuilder, mulligan, fatigue). ✅ DONE
10. **Polish ← IN PROGRESS** — reconnect ✅ DONE; animation timing + edge triggers TODO.

---

## Current state (2026-06-21)
Renamed project Freestone → **Vanillastone** (module `github.com/amvid/vanillastone`).

**Infra (done):**
- Go module `github.com/amvid/vanillastone`, go 1.26.4
- Dockerfile: `dev` (wgo hot reload) + `build` + `prod` (distroless static) targets
- docker-compose.yml: live source mount, persisted go caches, port 8080
- Makefile, .dockerignore, .gitignore, CLAUDE.md

**Phase 1 (done) — skeleton loop, real code:**
- `cmd/server/main.go`: long-running `http.Server`; routes `/register`, `/login`, `/ws`, `/` static (embed). Opens SQLite (`DB_PATH` env, default `vanillastone.db`). Container stays up, hot reload live.
- Transport: `github.com/coder/websocket`. Per-conn reader + writer goroutine; outbound via buffered `send` chan (drops if full).
- `internal/protocol`: JSON message types. C→S: `auth`, `end_turn`, `play_card` (+`targetId` for targeted spells), `attack` (targetId `oppHero`/minion uid). S→C: `joined`, `waiting`, `match_start` (now + initial snapshot), `state` (full per-player snapshot), `game_over`, `error`. `CardView` carries `cardType` + `target` (spell targeting rule).

**Accounts (done) — added 2026-06-21:**
- `internal/store`: SQLite (`modernc.org/sqlite`). `users(id, username UNIQUE, password_hash, created_at)`. `CreateUser` (ErrUsernameTaken on dup), `GetUser`. Pool `MaxOpenConns(1)`.
- `internal/auth`: bcrypt (`golang.org/x/crypto/bcrypt`). `Register` (validates username 3-20, password 6+), `Login` (opaque ErrBadCreds, no user enumeration), in-memory session map, `Username(token)`. HTTP handlers in `auth/http.go`: POST `/register` (201/409/400), POST `/login` (200+token / 401).
- ws flow: client POSTs /register then /login → token → ws connect → `{type:auth,token}` → server validates → `joined`. Bad token = error.
- Verified live: 201/409/400/200/401 status codes correct; tests pass (`internal/auth`, `internal/transport`).
- `internal/lobby`: single global queue, pairs 2 players → match. First to queue acts first. In-memory.
- `internal/match`: 1v1, owns turn truth (mutex). `EndTurn` validates current player, flips turn, increments `turnNum`, broadcasts `state`. Off-turn end_turn rejected.
- `internal/transport`: ws upgrade + message routing. `Client` implements `match.Sender` (ID/Send).
- Verified: `go build ./...` + `go vet` green, integration test `internal/transport/transport_test.go` passes (auth → match → authoritative turn order → off-turn reject → ping-pong → turnNum; bad token rejected).

**Web client (done) — React + Vite + TypeScript, dockerized:**
- `web/`: Vite app. `src/App.tsx` (auth form → ws → turn UI), `src/api.ts` (register/login fetch), `src/protocol.ts` (TS wire types mirroring Go `internal/protocol`), `src/main.tsx`, `src/index.css`. Entry `web/index.html`.
- Package manager **pnpm** (via corepack; pinned `pnpm@9.15.0` in package.json, lockfile committed).
- `web/web.go` embeds `web/static` (Vite build output). Build uses **stable filenames** (`assets/app.js`, `app.css`) so committed `static/` doesn't churn. `static/` IS committed so `go build` works on clean checkout.
- Dockerized dev: `web/Dockerfile` (node:24-alpine + corepack pnpm), compose `web` service on :5173, HMR (`usePolling` for mac docker), proxies `/ws` `/register` `/login` → `server:8080` via `VITE_PROXY_TARGET` env.
- Prod: root `Dockerfile` has `web-build` stage (pnpm build → `web/static`) feeding the Go `build` stage, so the single distroless binary embeds the real client.
- Verified live (both containers up): vite serves app :5173, register/login proxy → 201/token, Go serves embedded client + `assets/app.js` 200 at :8080. **ws proxy not browser-verified** (config `ws:true`; Go ws path covered by Go test).

**Dev workflow now:** `make dev` = `docker compose up` → server :8080 + vite :5173. **Open http://localhost:5173** (vite, HMR). :8080 serves the embedded prod build.

**Phase 2 (done) — minions + combat, real code:**
- `internal/cards`: server source of truth. `Card{ID,Name,Cost,Attack,Health}`, hardcoded `set` of 6 vanilla minions (custom names: Pebble Imp 1/1, Clay Acolyte 3/2, Granite Watcher 2/3, Thicket Stalker 3/3, Iron Bulwark 4/5, Stormbound Colossus 6/6). `OpeningHand()` = fixed 7-card hand (both players identical). **No deck/draw/fatigue yet** (Phase 9).
- `internal/match` rewritten: owns full game truth under one mutex. `playerState{heroHP=30, mana, maxMana, hand, board}`. `minion{uid, card, attack, health, maxHealth, canAttack}`.
  - Mana ramps +1/turn to **10 cap**, refills to max at turn start. Board cap **7**. Hero 30 HP.
  - `PlayCard(c, handIndex)`: validates turn/index/mana/board-cap; summons with `canAttack=false` (**summon sickness**). `Attack(c, attackerID, targetID)`: `targetID="hero"` or opp minion uid; simultaneous damage exchange, dead (`health<=0`) removed. `startTurn` wakes all owner minions (`canAttack=true`). All action methods return `(ok bool, msg string)`.
  - Win: opp heroHP<=0 → `over=true`, sends final `state` then `game_over{winner}`. Further actions rejected ("game over"). **No socket/match teardown yet** (zombie match still possible).
  - Per-player snapshots: `selfView` includes hand cards, `oppView` exposes **handCount only** (hidden hand). `match_start` now carries initial snapshot (so no turnNum-0 `state`, keeps Phase 1 test green).
- `internal/transport`: routes `play_card`, `attack`; `EndTurn`/`PlayCard`/`Attack` errors surfaced as `error{msg}`.
- `internal/match/match_test.go`: mana gate, summon sickness (+wake next turn), simultaneous combat trade, off-turn reject, board cap, hero-damage→win/game_over. All green; Phase 1 transport test still green.
- Web client: `protocol.ts` mirrors new wire types (CardView/MinionView/PlayerView/Snapshot, play_card/attack/game_over). `App.tsx` renders both heroes (HP/mana/handcount), both boards, own hand (cost/atk/hp, disabled if unaffordable/off-turn). Click hand card = play; click friendly ready minion = select attacker, then click enemy minion or enemy hero = attack. game_over banner. tsc + vite build clean, `static/` regenerated + embedded.

**Phase 2 design notes / open issues:**
- No draw — hand only shrinks. Running out of cards = no plays (no fatigue damage yet). Fine for combat demo.
- Charge not implemented (Phase 5): ALL minions summon-sick. `canAttack` conflates sickness + already-attacked (no windfury yet).
- Win condition exists but match/socket NOT torn down on game_over or disconnect — zombie match persists. Do teardown alongside reconnect (Phase 10) or sooner.
- **Not browser-verified end-to-end** (Go engine fully unit-tested; client compiles). Next session: two tabs at :5173, play through a kill.

**Phase 3 (done) — spells + targeting, real code:**
- `internal/cards`: added `Type` (minion/spell) + data-driven `Effect{Kind,Amount,BuffAtk,BuffHP,Target,Area}`. Enums: EffectKind (damage/heal/buff), TargetRule (none/any/friendlyMinion/enemyMinion), AreaRule (enemyMinions). 4 spells: **Cinder Bolt** (2c, 3 dmg, any char), **Mend** (1c, heal 4 capped, any char), **Whetstone** (1c, +2/+1, friendly minion), **Quake** (3c, 1 dmg to all enemy minions, untargeted). New 8-card opening hand mixing minions+spells (pebble_imp, cinder_bolt, clay_acolyte, whetstone, granite_watcher, mend, thicket_stalker, quake). iron_bulwark + colossus still in set, not dealt.
- `internal/match`: `PlayCard(c, handIndex, targetID)` branches on `card.Type`. `playSpell` validates target **before any mutation** (reject leaves state untouched — tested). `charRef{minion,owner}` + `resolveTarget(pi,targetID)` maps `selfHero`/`oppHero`/uid → character. `validTarget(rule,ref,pi)` enforces targeting rule. `applyEffect`: damage (single or AoE enemyMinions), heal (caps at maxHealth / heroMaxHP=30), buff (attack+health+maxHealth). Win-check centralized: **`resolveDeathsAndWin`** clears dead both boards + ends match if any hero ≤0 (winner = survivor) — used by both `Attack` and spells, so spell-to-face wins. Combat hero target migrated `hero`→`oppHero`.
- `internal/transport`: `play_card` passes `TargetID`.
- Tests: 7 new in `match_test.go` (spell kills minion, target-required no-op, illegal target, buff stats, heal cap, Quake AoE survivors, spell-to-face win) + phase-2 tests updated for new signature/hand. All green.
- Web client: `protocol.ts` adds `TargetRule`, CardView `cardType`/`target`, `play_card.targetId`. `App.tsx` unified targeting: one of {attacker, pendingSpell} active. Targeted spell → arm (click card again to cancel) → click highlighted legal char. Untargeted spell / minion play instantly. `ruleMatches` mirrors server `validTarget`. Self-hero clickable (heal). Spell cards + targetable heroes styled (plain CSS — Tailwind deferred, UI too small to justify). tsc + vite build clean, `static/` embedded.

**Pre-Phase-4 UX (done) — 2026-06-21:**
- **Card text + hover box.** `cards.Card` gained `Text string` (human rules text;
  vanilla minions empty). Surfaced in `protocol.CardView` + `MinionView` (`text`
  omitempty), set in `match.cardView`/`minionViews`. Client renders a CSS-only
  hover tooltip (`CardTooltip` in `App.tsx`, `.tooltip` in index.css) over hand
  cards and board minions showing name · type · cost/stats · text. Spell texts:
  Cinder Bolt "Deal 3 damage to any character." / Mend "Restore 4 Health to any
  character." / Whetstone "Give a friendly minion +2/+1." / Quake "Deal 1 damage
  to all enemy minions."
- **Lobby + persisted session (auth decoupled from queue).** Auth no longer
  auto-queues: `joined` now means "in lobby". New C→S msgs `find_match` (Play
  button → enter queue) and `enter_lobby` (back to lobby; drops queue/abandons a
  finished match + refreshes counts). Client persists the login token in
  `localStorage` (`vs_token`); on load it auto-reconnects (phase `connecting`).
  An `error` before `joined` = stale token → clear it, show login. After
  `game_over` the player returns to the lobby (no re-login); ws stays open.
  Server sessions still die on restart (token then invalid → re-login). Lobby UI:
  Play + Build deck (disabled stub) + Log out.
- **Presence counts.** `transport.Server` tracks an authed-client set
  (`clients map[*Client]struct{}`, guarded by `s.mu`). New S→C `lobby` msg
  `{online, inGame}` broadcast on auth/find_match/enter_lobby/disconnect. inGame
  counts clients whose `match != nil && !match.Over()` (new `Match.Over()`
  method). `Client.match` is now `atomic.Pointer[match.Match]` because presence
  reads it across goroutines (race-tested clean).
- **Presence accuracy.** Counts are by **distinct username**, not connection, so
  a reload (stale socket briefly lingers) or a 2nd tab per account doesn't
  inflate online/inGame. Plus a **heartbeat** per connection (`heartbeat` in
  transport): `conn.Ping` every 15s with a 10s pong timeout; on failure it
  cancels the conn ctx so the dropped client is deregistered within ~25s instead
  of lingering until TCP notices. (Also doubles as a proxy keepalive.)
- **Concede.** New C→S `concede` msg → `Match.Concede(c)` (valid on either
  turn): sets conceder heroHP=0, reuses `resolveDeathsAndWin` → opponent wins +
  game_over. Client "Concede" button next to End Turn (enabled regardless of
  turn, disabled once a winner exists). Test `TestConcede` (off-turn forfeit,
  winner is opponent, post-over rejected).
- Tests: transport test updated to send `find_match` after auth; `expect` skip
  cap raised to 20 (presence chatter). `go test -race ./...` green.

**Phase 4 (done) — event bus + triggers, real code:**
- `internal/cards`: `EffectKind` gained **summon** (`Summon` tokenID + `Count`).
  `TargetRule` gained **randomEnemy** (server-picked, RNG). `AreaRule` gained
  **enemyHero**. New `EventType` (`on_play`, `on_death`) + `Trigger{When,Effect}`;
  `Card.Triggers []Trigger` with helpers `Battlecry()` (first on_play) /
  `Deathrattles()` (all on_death). 6 trigger minions: **Spark Adept** (2/2,
  battlecry deal 2 any), **Ember Striker** (2/1, battlecry deal 1 enemy minion —
  fizzles if none), **Bog Warden** (2/2, battlecry summon 1/1 Bogling),
  **Brood Mother** (2/2, deathrattle summon 1/1 Hatchling), **Cinder Husk**
  (3/2, deathrattle deal 2 enemy hero), **Volatile Wisp** (1/1, deathrattle deal
  1 random enemy). Tokens **bogling**/**hatchling** (summon-only, not in hand).
  Opening hand now 14 cards (pebble_imp still idx 0, thicket_stalker idx 6).
- `internal/match`: now **event-driven**. `minion` gained `owner` (drives
  deathrattle sides/summons). `Match` gained seeded `rng *rand.Rand` and a
  per-action event buffer `log []protocol.Event`. **`New` signature changed**:
  `New(id, a, b, seed int64)` (lobby passes `time.Now().UnixNano()`; tests pass a
  fixed seed). Every action `resetLog()`s, mutates via emitters, then calls
  **`finish()`** = `resolveDeaths()` (cascade: remove all ≤0, fire deathrattles
  in board order, loop until none — deathrattle damage can kill more) + hero
  win-check + `sendStateAll`. `applyEffect(caster,…)` extended with summon /
  AreaEnemyHero / randomEnemy and now **emits events** (damage/heal/buff/summon)
  via `damageMinion`/`damageHero`/`damageRandomEnemy`. Battlecry resolves in
  `PlayCard` **before mutation** (validate target if a legal one exists; **fizzle**
  = play minion, skip battlecry, if none). `resolveDeathsAndWin` replaced by
  `resolveDeaths`+`finish`.
- `internal/protocol`: new **`Event{Kind,Source,Target,Amount,Name}`** (kinds:
  battlecry/deathrattle/damage/heal/buff/summon/death/attack). `State` gained
  `Events []Event` (ordered resolution log; ids are match-global — minion uid or
  player id p1/p2 for heroes — so the list is identical for both players, no
  hidden info). `MatchStart` carries no events (opening has none).
- **Match teardown:** transport now forfeits on disconnect — if a client drops
  mid-match (`match != nil && !Over()`) it calls `Concede`, so the opponent wins
  and the match ends instead of lingering as a zombie; ref then cleared.
- Web client: `protocol.ts` adds `Event` + `state.events`. `App.tsx` keeps a
  uid→name map (`namesRef`, so dead minions still name correctly) and renders an
  **Event log** panel (`formatEvent`/`entName`, `.log` CSS) appending each
  action's events (last 40). Heroes shown as Your/Enemy hero.
- Tests: 10 new in `match_test.go` (battlecry hit-face / kill-minion /
  requires-target / fizzle-no-target / summon; deathrattle summon / hero-damage /
  random-damage; simultaneous-death deathrattles; event-log emitted) via a
  white-box `place()` board helper. All Phase 1-3 tests updated for the `New`
  seed param + `oppHeroTarget` in the hero-win loop. `go test -race ./...` green,
  `go vet` + `gofmt` clean, web `tsc`+`vite` build clean, `static/` re-embedded.

**Phase 5 (done) — keywords wave 1, real code:**
- `internal/cards`: `Keyword` type + `Card.Keywords []Keyword` (taunt/charge/rush/
  divineShield) with `Card.Has(k)`. `Effect` gained `Freeze bool` (applies freeze
  to each character the effect hits; combinable with damage). New minions:
  **Bastion Golem** 3/5 Taunt, **Swift Raptor** 3/2 Charge, **Lurking Stalker**
  3/3 Rush, **Gilded Sentry** 2/3 Divine Shield. New spells: **Frost Snap** (2c, 1
  dmg + freeze a character), **Permafrost** (3c, freeze all enemy minions; 0 dmg).
  Opening hand now 20 cards (idx 0 = pebble_imp, idx 6 = thicket_stalker still).
- `internal/match`: **attack model split** — `minion.canAttack` replaced by
  `summonedThisTurn` (sickness) + `attacked` (swung this turn) + `frozen` +
  `divineShield`. `startTurn` clears summonedThisTurn/attacked (does NOT thaw).
  `canAttack(mn)` = atk>0 && !attacked && !frozen && (!summonedThisTurn || charge
  || rush); `canAttackHero(mn)` additionally bars Rush-on-summon-turn from face.
  `Attack` enforces frozen, sickness, **Taunt** (`hasTaunt(opp)` → must hit a
  taunt; blocks hero + non-taunt), and Rush hero-restriction. Combat damage now
  goes through `damageMinion` which **pops Divine Shield** (absorbs one instance;
  0/neg amounts no-op so pure-freeze doesn't pop). **Freeze thaw**:
  `thawAfterTurn(pi)` runs in `EndTurn` for the ending player — unfreezes its
  characters that didn't attack (so a freeze applied on your turn lasts exactly
  one enemy turn). Hero freeze stored (`playerState.frozen`) but inert (no hero
  attack until Phase 8). `applyEffect` damage path rebuilt around `damageTargets`
  (+ `randomEnemy` returns a `charRef`) so damage and freeze share target
  resolution. New events: **shield** (pop), **freeze**.
- `internal/protocol`: `MinionView` gained `canAttackHero`, `taunt`,
  `divineShield`, `frozen`; `PlayerView` gained `frozen` (hero).
- Web client: `protocol.ts` mirrors the new view/Event fields. `App.tsx`
  targeting honors Taunt (must click a taunt; hero/non-taunt not targetable while
  enemy has taunt) and hero-reach (`canAttackHero`, Rush); minion badges
  (Taunt/Divine Shield/Frozen) + state CSS; event log lines for freeze + shield.
- Tests: 7 new in `match_test.go` (charge hits face, rush no-face + can hit
  minion, taunt forces target, divine shield absorbs one hit, freeze prevents
  attack then thaws, permafrost freezes all). `place()` helper updated for the new
  minion fields. `go test -race ./...` green; web `tsc`+`vite` clean, `static/`
  re-embedded.

**Phase 6 (done) — keywords wave 2, real code:**
- **Stat model refactor (own step):** `minion` no longer stores flat
  `attack`/`maxHealth`. It now holds `enchants []enchant{atk,hp}` (persistent
  buffs) + `auraAtk` (recomputed) + `health` (current only). `atk()` = base +
  enchants + aura (floored 0); `maxHP()` = base + enchants. Whetstone buff is now
  an enchantment. This is the foundation Silence (strip enchants) and Aura
  (recomputed layer) need. No behavior change at that step; all prior tests green.
- `internal/cards`: `EffectKind` gained **silence**. `TargetRule` gained
  **minion** (any minion, either side). `Effect` gained **`Lifesteal bool`**
  (damage dealt heals caster). `Keyword` gained **windfury/stealth/poisonous/
  lifesteal**. New **`Aura{Atk,HP}`** type (HP reserved; cards atk-only) +
  `Card.Aura *Aura`. New **`Card.SpellDamage int`**. New minions: **Gale Harrier**
  2/3 Windfury, **Veil Stalker** 2/1 Stealth, **Toxic Fang** 1/3 Poisonous,
  **Bloodthorn Knight** 3/4 Lifesteal, **Ember Scribe** 1/3 Spell Damage +1,
  **Pack Leader** 2/3 (aura: your other minions +1 Attack). New spells: **Hush**
  (1c, silence a minion), **Drain Touch** (2c, 2 dmg + Lifesteal). Opening hand
  now 28 (idx 0 pebble_imp / 6 thicket_stalker unchanged; new cards appended 20-27).
- `internal/match` (1:1 HS semantics):
  - **Windfury:** `attacked bool` → `attacksLeft int`. `attacksPerTurn()` = 2 w/
    windfury else 1; set at summon + each `startTurn`. `canAttack` checks
    `attacksLeft>0`; each attack does `attacksLeft--`. `hasAttacked()` =
    `attacksLeft < perTurn` (drives freeze thaw).
  - **Lifesteal:** `damageMinion`/`damageHero` now **return damage dealt**.
    `combatStrike(src,dst)` applies combat damage + lifesteal heal + poisonous.
    `lifestealHeal(pi,amt)` heals hero (cap 30). Spell lifesteal via
    `eff.Lifesteal` heals caster by total dealt (shield→0 dealt→0 heal).
  - **Poisonous:** if a poisonous source deals >0 to a minion → `health=0`
    (destroyed regardless of hp). Divine Shield (0 dealt) suppresses it. Heroes
    immune. Combat-only (no poisonous spell exists).
  - **Stealth:** per-instance `stealthed` bool (init from keyword). Enemy can't
    target it (spell/battlecry via `validTarget` guard; attack rejected
    "can't attack a Stealthed minion"). AoE/random still hit. `hasTaunt` ignores
    stealthed taunts. Stealth cleared when the minion attacks.
  - **Spell Damage:** `spellPower(pi)` = sum of non-silenced minions' SpellDamage.
    `applyEffect` gained `sp int`; spells pass `spellPower`, battlecry/deathrattle
    pass 0. Added to each damage instance with amount>0 (per-target on AoE; never
    boosts 0-damage/heal).
  - **Aura:** `refreshAuras()` zeroes then re-sums `auraAtk` from non-silenced
    aura sources onto the controller's *other* minions. Called in `finish()`
    (post-death) + at the top of `Attack` (so combat reads current attack).
    Attack-only for now (HP-aura needs current-health delta tracking — deferred).
  - **Silence:** `silence(mn)` sets `silenced=true`, clears enchants/divineShield/
    stealthed, clamps `attacksLeft` (lost windfury) and `health` to new `maxHP()`
    (HS clamp rule). Frozen is a status, NOT removed. `silenced` gates `has()`
    (keywords), `spellDamageOf`, aura emission, and deathrattle firing (suppressed
    in `resolveDeaths`).
- `internal/protocol`: `MinionView` gained windfury/stealth/poisonous/lifesteal/
  spellDamage/silenced; Event kinds gained **silence**.
- Web client: `protocol.ts` mirrors new view/Event fields + `TargetRule 'minion'`.
  `App.tsx`: `ruleMatches`/`hasLegalTarget`/`targetable` handle `minion` rule +
  enemy-Stealth untargetability + stealthed-taunt-doesn't-compel; new minion
  badges (Windfury/Stealth/Poisonous/Lifesteal/Spell Dmg/Silenced) + stealthed/
  silenced CSS; silence log line.
- Tests: 13 new in `match_test.go` (windfury two-attacks, lifesteal combat heal,
  poisonous destroys, divine-shield-blocks-poison, stealth untargetable, stealth
  lost on attack, stealth hit by AoE, spell damage single + per-target AoE, aura
  buffs-others-not-self + vanishes with source, silence strips keywords/buff/
  shield + clamps hp, silence suppresses deathrattle, drain-touch spell lifesteal).
  `place()` helper updated (enchant-delta stats + attacksLeft + stealthed).
  `go test -race ./...` green, `go vet`+`gofmt` clean, web `tsc`+`vite` clean,
  `static/` re-embedded.

**Phase 7 (done) — secrets + discover, real code:**
- **Caps centralized first:** `summonMinion` now returns nil and discards when the
  board is at `maxBoard` (7) — the single guard for every summon path (token,
  battlecry, deathrattle, Mimic). Playing a minion from hand still checks earlier
  so the play is rejected (not burned). New `maxHand` (10): Discover into a full
  hand burns the card. New `maxSecrets` (5).
- `internal/cards`: new `TypeSecret`. New `EffectDiscover` + `DiscoverPool`
  (spell/minion) on `Effect.Pool`. New secret-trigger `EventType`s
  (`on_enemy_attack_hero`/`on_enemy_play_minion`/`on_enemy_cast_spell`). Secrets
  are **Go-handled** (HANDOFF "weird = handler by ID"): `SecretKind`
  (destroyAttacker/copyMinion/counterSpell) + `Card.Secret *SecretDef{Trigger,
  Kind}`. New `Card.Token bool` (bogling/hatchling flagged; excluded from
  Discover). `DiscoverPoolIDs(pool)` = sorted non-token cards of the type.
  Cards: secrets **Snare** (2c, enemy minion attacks your hero → destroy it),
  **Mimic** (3c, enemy plays a minion → summon a copy), **Nullify** (3c, enemy
  casts a spell → counter it); Discover minions **Arcane Insight** (2/2, discover
  a spell), **Wild Summons** (3/2, discover a minion). Opening hand now 33 (new
  idx 28-32; earlier indices unchanged).
- `internal/match`:
  - **Secret zone:** `playerState.secrets []*secretInst{uid,card,owner}`.
    `playSecret` (cap + no-duplicate-active rules; spends mana; no reveal event).
    `triggerSecrets(defender, ev, ctx)` fires the defender's matching secrets in
    play order, emits a `secret` reveal event, removes them, and returns
    `cancelled` (true for counterSpell / destroyAttacker — interrupts). Wired:
    `PlayCard` minion → `OnEnemyPlayMinion` (after summon, before battlecry);
    `playSpell` → `OnEnemyCastSpell` (after mana/discard, before effect; counter
    skips effect, card+mana still spent); `playSecret` → `OnEnemyCastSpell` too
    (a secret IS a spell, so it can be Nullified — countered secret never enters
    play, card+mana still spent); `Attack` hero → `OnEnemyAttackHero` (before
    damage; destroy attacker + negate damage). NOTE: the Discover cards are
    *minions* (battlecry), not spells, so they are NOT counterable by Nullify —
    only their summon triggers Mimic. Add a discover *spell* if a counterable one
    is wanted.
  - **Discover (blocking protocol):** `Match.pending *pendingChoice{player,
    options}`. A discover battlecry calls `startDiscover` — picks 3 distinct from
    the pool (`pickDistinct` via match RNG), sets `pending`, sends the current
    snapshot + a `discover` prompt **to the chooser only**, and returns WITHOUT
    finishing. `EndTurn`/`PlayCard`/`Attack` reject while `pending != nil`
    ("finish discovering first"). New `Choose(c, index)` validates owner+range,
    adds the card to hand (or burns if full), clears pending, finishes. `Concede`
    clears pending (abandon in-flight discover). The chosen card is deliberately
    NOT named in the shared event log (no hidden-info leak).
- `internal/protocol`: C→S `choose{index}`; S→C `discover{options[]CardView}`.
  `PlayerView` gained `secrets []CardView` (own only) + `secretCount` (both).
  Event kinds gained `secret`. `selfView` exposes own secrets; `oppView` exposes
  only the count.
- `internal/transport`: routes `choose` → `Match.Choose`.
- Web client: `protocol.ts` mirrors all of the above (CardView `'secret'`,
  PlayerView secrets/secretCount, `choose`/`discover` messages, `secret` event).
  `App.tsx`: a `discover` modal (overlay; click an option → `choose`); own
  secrets shown as 🔒 Name near your hero, opponent's as 🔒 count; `secret` log
  line; secret hand-card styling. CSS for overlay/discover/secret.
- Tests: 10 new in `match_test.go` (secret hidden from opponent, no-duplicate,
  Snare destroys attacker + negates damage, Mimic copies played minion, Nullify
  counters spell, Nullify counters a cast secret, both keep card/mana spent,
  Discover pauses + blocks + resumes + pool-filtered + chooser-only prompt,
  Discover into full hand burns, summon discarded when board full). New
  `placeSecret`/`lastDiscover` helpers.
  `go test -race ./...` green, `go vet`+`gofmt` clean, web `tsc`+`vite` clean,
  `static/` re-embedded.

**Phase 8 (done) — hero power + weapons, real code:**
- `internal/cards`: new `TypeWeapon` + `TypeHeroPower`. `Card.Durability` (weapons).
  Cards: **Fireblast** hero power (2 mana, 1 dmg any — `MageHeroPower()`, not in
  any hand); weapons **Ember Cleaver** (2c, 3/2) + **Quartz Spike** (3c, 2/3).
  Opening hand now 35 (weapons idx 33-34; earlier indices unchanged).
- `internal/match`:
  - `playerState` gained `heroPower cards.Card` (set to Fireblast in `New`),
    `heroPowerUsed`, `weapon *weaponInst{card,durability}`, `heroAttacked`. Both
    per-turn flags reset in `startTurn`.
  - **Hero power:** `HeroPower(c, targetID)` — once/turn, costs mana, resolves
    target via the shared `resolveTarget`/`validTarget`, applies the effect with
    `sp=0` (**not** boosted by Spell Damage — HS rule for Fireblast).
  - **Weapons:** equipping is a `TypeWeapon` branch in `PlayCard` (spend mana,
    replace any current weapon, emit `equip`). **Hero attack** reuses the `attack`
    message with `attackerId == "selfHero"` → `heroAttack(pi, targetID)`: requires
    a weapon (attack>0), not frozen, not already attacked; deals weapon attack to
    the target (Divine Shield applies), takes the struck minion's attack back to
    your hero, respects Taunt/Stealth, spends 1 durability (`useWeaponDurability`,
    breaks at 0 → emit `weaponBreak`). The dormant hero `frozen` flag now blocks
    attacking. Hero attacks do NOT trigger Snare (minion-specific).
- `internal/protocol`: C→S `hero_power{targetId}`. `CardView.Durability`.
  `PlayerView` gained `heroPower *CardView` + `heroPowerUsed` + `weapon
  *WeaponView` + `heroAttack` + `heroCanAttack` (all public, both views). New
  `WeaponView{name,attack,durability,text}`. Event kinds `equip`/`weaponBreak`/
  `heropower`.
- `internal/transport`: routes `hero_power` → `Match.HeroPower`.
- Web client: `protocol.ts` mirrors all of it (CardView weapon/heroPower types +
  durability, PlayerView hero-power/weapon fields, WeaponView, `hero_power`
  message, new event kinds). `App.tsx`: a Hero Power button (arms targeting for
  Fireblast like a spell, or fires if untargeted); clicking your own hero (when
  `heroCanAttack`) selects it as the attacker (`attacker === "selfHero"`) → click
  an enemy to swing; weapon shown by both heroes (⚔ atk/dur); equip/weaponBreak/
  heropower log lines; `cardStats` helper renders weapon a/durability. CSS for
  weapon/hero-power/ready-hero.
- Tests: 11 new in `match_test.go` (hero power damage + once/turn + refresh, not
  spell-damage-boosted; equip gives attack, hero attack face spends durability +
  once/turn, minion retaliation, weapon breaks at 0, equip replaces old, frozen
  hero blocked, taunt respected, hero attack doesn't trigger Snare).
  `go test -race ./...` green, `go vet`+`gofmt` clean, web `tsc`+`vite` clean,
  `static/` re-embedded.

**Armor + Frost Ward (added with Phase 8):**
- `playerState.armor int` (no cap). **`damageHero` now absorbs via armor first**,
  remainder to health — applies to ALL damage (spells, hero power, attacks, combat
  retaliation). Returns the full hit for Lifesteal (armor-absorbed counts as damage
  dealt, as in HS). `gainArmor(h, n)` emits an `armor` event.
- New event type `OnHeroAttacked` (the owner's hero is attacked by a minion OR a
  weapon) — distinct from `OnEnemyAttackHero` (Snare, **minion-only**). New
  `SecretKind` **gainArmor** + `SecretDef.Amount`. Card **Frost Ward** (3c secret,
  "When your hero is attacked, gain 8 Armor"). Opening hand now 36 (idx 35).
- Wiring: the minion-attacks-hero path fires BOTH `OnEnemyAttackHero` (Snare) and
  `OnHeroAttacked` (Frost Ward); the weapon-attacks-hero path fires only
  `OnHeroAttacked`. Both fire BEFORE damage, so armor absorbs the triggering hit.
  Spells / hero power damage the hero via `damageHero` directly (no attack event),
  so they do NOT trigger Frost Ward — but armor still absorbs them.
- `protocol.PlayerView.Armor` (both views, public); event kind `armor`. Client
  shows 🛡 N by each hero + an armor log line.
- Tests: 4 more (armor absorbs spell damage; Frost Ward armor on minion attack;
  on weapon attack; NOT triggered by a spell).

**Phase 9 (done) — decks + SQLite, real code:**
- `internal/cards`: new **`coin`** token spell (0-cost, `EffectMana` +1, `Token:true`) +
  `EffectMana` kind. Deck rules: `DeckSize=30`, `MaxCopies=2`. Helpers
  `DeckPoolIDs()` (sorted buildable ids = non-token, non-heroPower of
  minion/spell/secret/weapon), `ValidateDeck(ids)` (size/copies/pool), `DefaultDeck()`
  (first 15 pool ids ×2 — a legal 30), `Deck(ids)→[]Card`, `Coin()`. **Removed the
  fixed `OpeningHand()`/`openingHand`** (replaced by real decks; the 36-id list now
  lives only in `match_test.go` as `testHand` for the legacy white-box tests).
- `internal/store`: **`decks`** table `(id, username, name, cards TEXT json, created_at,
  updated_at)` + `idx_decks_username`. CRUD `ListDecks`/`GetDeck`/`CreateDeck`/
  `UpdateDeck`/`DeleteDeck`, **all scoped by username** (no cross-user read/write).
  `CreateDeck` enforces **`MaxDecksPerUser=10`** → `ErrDeckLimit`. `ErrNotFound`
  reused (message now "not found"). Card-legality validated by the caller (HTTP
  layer), not the store.
- `internal/auth`: deck **HTTP API** (`decks_http.go`), Bearer-token authed via
  `userFromRequest`. `GET/POST /decks`, `PUT/DELETE /decks/{id}` (Go 1.22 method+
  `{id}` patterns), `GET /pool` (public; buildable cards as **protocol.CardView** +
  deckSize/maxCopies/maxDecks). `decodeDeck` validates name (1-40) + `cards.ValidateDeck`
  → 400; create over cap → 409; cross-user update/delete → 404. `poolCardView` mirrors
  `match.cardView` (kept in auth to avoid importing match).
- `internal/match`: `playerState` gained **`deck []cards.Card`** + **`fatigue int`**.
  **`New(id,a,b,seed,deckA,deckB)`** shuffles each deck (match RNG), deals opening
  hands (**3** first / **4** second) via `dealOpening`, and opens in the **mulligan
  phase** (`m.mulligan *mulliganState{done [2]bool}`, no turn started). `Start()` sends
  `match_start{mulligan:true}`. **`Mulligan(c, indices)`**: validates indices, tosses
  them, draws replacements off the top BEFORE shuffling the tossed cards back (no
  instant redraw), marks done; when both done **`beginPlay()`** grants **The Coin** to
  player 2, runs the first player's opening turn, win-checks. Mulligan **blocks**
  EndTurn/PlayCard/Attack/HeroPower ("mulligan in progress"); `Concede` abandons it.
  **`startTurn` now draws** via `drawCard` (empty deck → escalating **fatigue**,
  armor-absorbed; full hand → **burn** the draw). `EndTurn` now ends with `finish()`
  (turn-start fatigue can kill). `applyEffect` handles **`EffectMana`** (Coin: +1 mana
  this turn, capped at 10). Per-player views gained **`deckCount`**; mid-mulligan a
  submitter gets a private confirm via `sendStateTo`.
- `internal/protocol`: C→S **`mulligan{indices}`**, `FindMatch.DeckID`. S→C
  `MatchStart.Mulligan` + `State.Mulligan` + `PlayerView.DeckCount`. Event kinds
  `mana|fatigue|burn` added.
- `internal/lobby`: `Join(c, deck)` carries each queued player's deck; pairs into
  `match.New(...,deckA,deckB)`.
- `internal/transport`: `NewServer(auth, store)`. `find_match` parses `DeckID` and
  `deckFor(username, id)` loads+validates the chosen deck (fallback to `DefaultDeck`
  if 0/missing/invalid — queuing never fails). Routes `mulligan`. `main.go` wires
  `/pool`, `/decks`, `/decks/{id}`.
- Web client: `protocol.ts` (deckCount, mulligan flags, event kinds, `find_match.deckId`,
  `mulligan` msg). `api.ts` (fetchPool/listDecks/createDeck/updateDeck/deleteDeck, Bearer).
  New **`Deckbuilder.tsx`** (saved-deck list + delete, edit name, 30/30 meter, collection
  grid with per-card copy counter, server-validated save). `App.tsx`: lobby **deck
  picker** (`<select>`, default + saved) + Build-deck button → deckbuilder phase;
  **mulligan phase** UI (toggle cards to replace → Keep all/Replace N → "waiting for
  opponent"); deck counts by both heroes; mana/fatigue/burn log lines. Plain CSS.
- Tests: `cards` (DefaultDeck legal, ValidateDeck size/copies/token/heroPower), `store`
  (deck CRUD scoped, 10-cap, list scoping), `match` (mulligan opens+blocks+begins,
  replace preserves counts, 2nd player gets Coin, Coin +mana, fatigue escalates),
  `auth` (deck HTTP: auth required, validation 400, legal 201+list, 10-cap 409).
  Legacy white-box `newMatch()` now installs `testHand` + skips mulligan. `transport`
  ping-pong does the mulligan handshake (`waitPlay`). `go test -race ./...` green,
  `go vet`+`gofmt` clean, web `tsc`+`vite` clean, `static/` re-embedded. Live smoke:
  `/pool` + deck create/update/delete (201/200/204) verified.

**Phase 10 (in progress) — reconnect, real code:**
- **Scope locked this session:** reconnect only (resume = grace-window rejoin, not
  restart-survival, which is impossible — match state is in-memory). Animation timing
  (Q4) is the next session. Edge triggers (on_damage/on_turn_*/on_summon/on_spell_cast)
  **deferred** — no current card needs them (speculative; revisit when such cards arrive).
- **Core mechanism:** the match identifies players purely by `Sender.ID()` and only ever
  calls `ID()`/`Send()` on the Sender. So reconnect = the new socket **re-adopts the old
  player-slot id** and the match **swaps the Sender pointer** in place — turn identity is
  unchanged for BOTH sides, no id remap anywhere else.
- `internal/match`: `Seats(c Sender) bool` (pointer-identity: is c still the live occupant
  of a slot? — distinguishes a dropped client from one already replaced by a takeover).
  `Reattach(c Sender) bool` (c carries the adopted slot id → `indexOf` finds the slot,
  swap `m.players[i]=c`, push a fresh full snapshot via `sendResyncTo`, and re-send a
  pending Discover prompt if that player had one — the one-shot prompt was lost with the
  old socket). `sendDiscoverTo(pi)` extracted from `startDiscover` so both can reuse it.
  All under `m.mu`; returns false if `m.over`.
- **Event-log replay across reconnect:** the match keeps a rolling `history []Event`
  (`emit` appends to both the per-action `log` and `history`, capped at `maxHistory`=60).
  `sendResyncTo` sends the resync snapshot with `Events=history` and a new `State.Resync`
  flag; the client REPLACES its log from it (newest-on-top) instead of appending. Recovers
  the log on a page-reload reconnect AND fills in events that resolved while the player was
  disconnected (their live `Send`s were dropped to a dead socket). Without it the log panel
  was empty after reconnect.
- `internal/transport`: `Server` gained `active map[username]activeSeat{m,slot}` (the live
  match + slot each user occupies) + `grace map[username]*time.Timer` (pending forfeit) +
  per-Server `graceWindow` (default 60s, `var`-free — a field so tests shorten it without a
  shared global). `attachMatch` records both seats. **Disconnect cleanup** no longer
  forfeits immediately: if the dropping client still `Seats` its slot, `startGrace` holds
  it open and `notifyOpp(false)` tells the opponent; the timer `Concede`s only if the seat
  is still c's after the window (so a reconnect — which swaps the Sender pointer — silently
  cancels it; the timer is also `Stop`ped on reconnect). **`handleAuth`** calls
  `tryReconnect` first: if the user has a live (`!Over()`) seat, it cancels grace, kicks any
  other live tab, **adopts `c.id = seat.slot`**, sends `Joined` (BEFORE Reattach, so the
  client records its id before the resync `state` arrives — else the late `joined` would
  bounce it back to lobby), registers presence, `Reattach`es, and `notifyOpp(true)`. A
  finished/absent seat falls through to the normal lobby flow. `find_match`/`enter_lobby`
  clear the user's `active` entry (fresh start / left a finished match). `clearActive(m)`
  drops both seats when the grace timer forfeits.
- `internal/protocol`: S→C **`opp_conn{connected bool}`** (`TypeOppConn`) — opponent dropped
  (false, grace begun) / returned (true). Purely informational.
- Web client: `protocol.ts` adds `opp_conn`. `App.tsx`: **light auto-reconnect** — on an
  unexpected socket close while a token is held, retry `connect` (≤6 tries, 1.5s apart,
  status "connection lost — reconnecting…"); the server keeps the seat, so a transient
  blip resumes the game seamlessly. `giveUpRef` suppresses retries on logout / dead-token
  (avoids a reconnect loop); `retriesRef` resets on a successful `joined`. **Opponent
  banner** (`oppOnline` state + `.banner.warn` CSS) shown in the playing AND mulligan
  phases while the opponent is in their grace window. (Page-reload reconnect already worked
  via the on-load token effect; this adds same-page transient-drop recovery + the opponent
  notice.)
- Tests: `internal/transport` `TestReconnectResumesMatch` (drop → opponent told
  disconnected; reconnect re-adopts the SAME seat id, gets a resync snapshot showing it is
  still that player's turn, opponent told reconnected; match continues — end_turn flips
  turn, proving no forfeit; resync flagged `Resync`) + `TestGraceForfeitsIfNoReconnect`
  (shortened window; drop with no return → opponent wins via game_over). `internal/match`
  `TestReattachResyncsWithHistory` (resync carries the pre-drop summon event + board) +
  `TestReattachAfterOverRejected`. `newServer` helper now also returns the `*Server` (so a
  test can set `graceWindow`); existing call sites updated. `go test -race ./...` green (×3
  on transport), `go vet`+`gofmt` clean, web `tsc`+`vite` clean, `static/` re-embedded.

**Phase 10 — client UI overhaul (done) — 2026-06-21:**
- **Goal:** replace the plain dev UI with an HS-like table. Pure client work — NO engine
  change except one tiny additive field (opponent username, below). Server still authoritative.
- **File split.** `web/src/App.tsx` (was 865 lines) is now the *brain only*: phase state,
  ws connect/reconnect, message routing, all handlers (onChar/targetable/onHandCard/
  onHeroPower). Presentational + helper code extracted to **`web/src/game/`**:
  - `types.ts` — `Phase`, `CharKind`, `PendingSpell`, `Counts`, `LogEntry`.
  - `format.ts` — `entName`, `formatEvent`, `ruleMatches`, `cardStats`, **`kindIcon`** (event→emoji).
  - `Tooltip.tsx` — `CardTooltip` (hover info box).
  - `Board.tsx` — minion row (cards with corner atk/hp circles, taunt bar, keyword badges;
    `data-cid={instanceId}` for the targeting arrow).
  - `Hero.tsx` — portrait panel: round portrait + HP orb / armor orb / weapon orb, secret "?"
    gems arced over the portrait, nameplate, weapon line. (`data-cid` = selfHero/oppHero.)
  - `GameScreen.tsx` — the whole in-game HS layout (props bundle from App). Contains the
    inline `ManaBar`, `CardBacks` (opp hand), `DeckPile` sub-components + discover modal + log.
  - `TargetingArrow.tsx` — SVG aiming line.
- **HS table layout** (in-game = `position:fixed; inset:0`, breaks out of the 720px `#root`
  that still constrains auth/lobby/deckbuilder/mulligan). `.play-area` is a column with a
  warm wooden-board radial bg; `.zone.top`/`.zone.bottom` each `flex:1` + `space-between` so
  the two boards hug the center `.midline` (short attack travel), heroes pushed to the edges.
  `.log` sidebar on the right.
- **Pieces:** mana = 10 fixed crystal slots (on/spent/locked) centered **below** each hand
  (fixed width → ramping never shifts layout). Opp hand = face-down `.card-back` row sized like
  our hand. Decks = stacked card-backs anchored right edge near End Turn (opp upper, self lower),
  count on a custom hover tooltip. End Turn = big button, center-right, **shown only on your
  turn**; Concede = small, bottom-left. Disconnect/game-over banners float (absolute).
- **Event log** = narrow right column of action **icons** (one per row, `kindIcon`), newest
  first (capped 15 shown / 40 in state), custom hover tooltip to the LEFT with the text
  (the sidebar can't use `overflow:auto` — it clips the tooltip; uses `overflow:visible`).
- **Targeting arrow:** selecting an attacker / arming a spell / arming the hero power draws a
  red SVG line + arrowhead from the source (`data-cid` lookup) to the cursor. `GameScreen`
  tracks the pointer on `mousemove` only while aiming; `TargetingArrow` reads the source
  element's `getBoundingClientRect()`. Overlay is `position:fixed; pointer-events:none`.
- **Opponent username (only server change):** `protocol.PlayerView` gained `Name`; the match
  `Sender` interface gained **`Name() string`** (transport `Client.Name()` returns the authed
  username; test `fakeSender.Name()` returns its id). `selfView`/`oppView` take a `name` param;
  the 3 call sites (Start/sendStateTo/sendResyncTo) pass `m.players[i].Name()`. Client shows the
  real names on both nameplates (falls back to "Opponent"/own name). `go test ./internal/...` green.
- All builds green: `pnpm build` (tsc + vite) + `go build ./...` (re-embeds `web/static`).
- **Still plain emoji icons** (event log, orbs) — the planned game-icons.net / Kenney asset pass
  ("Move 2") is NOT done; current look is CSS + emoji only.
- **Not browser-verified by the assistant** — user is eyeballing live in two tabs and iterating.

**Phase 10 — UI iteration round 2 + turn timer + prod deploy (done) — 2026-06-21:**
This was a long live-iteration session on top of the overhaul above. State of the client now:
- **Card art / icons (emoji placeholders, NOT the asset pass yet):** every card shows a
  centered type glyph via `cardArtIcon` (🐾 minion / 🔮 spell / ❓ secret / ⚔️ weapon) in
  `.art`/`.m-art`. Keyword badges are icon chips (`Board.tsx`): 🛡️ Taunt, ✨ Divine Shield,
  ❄️ Frozen, 🌀 Windfury, 🌫️ Stealth, 🧪 Poisonous, 🩸 Lifesteal, 🔮+N Spell Dmg, 🔇 Silenced
  (bottom-center, hover = name). Event-log icons via `kindIcon`. Hero power icon via `hpIcon`
  (🔥 Fireblast). Hero portrait shows 🧙 (Mage).
- **Card colors by type** everywhere (hand/mulligan/discover/book): minion=parchment,
  spell=blue, secret=purple, weapon=amber (`.card.spell/.secret/.weapon-card`).
- **Card back** = magical arcane CSS (scalable SVG rune sigil + violet radial + gold frame),
  shared by deck pile + opponent hand.
- **Board minions bigger** (96×120) so names fit; corner atk/hp circles; taunt bar.
- **Hero panel:** round portrait (HP orb, armor orb = number-only, **weapon chip** 🗡️ atk/dur
  bottom-left, secret "?" gems arced on top). Names moved to corners: opponent top-left,
  you bottom-left (`.player-name`). Real usernames (server `PlayerView.Name`).
- **Mana** = 10 fixed crystal slots centered BELOW each hand (opp mirrors on top); never shifts.
- **Decks** = stacked card-backs on the right near End Turn (opp upper / you lower), count on hover.
- **Opponent hand** = row of face-down card backs (= handCount), sized like your hand.
- **End Turn** = big button center-right, shown only on your turn. **Concede** = small bottom-left,
  now behind a **confirm modal** (`confirmConcede` state in GameScreen).
- **Hero power** moved left of portrait, smaller (46px); opponent's hero power shown (static).
- **Event log** = narrow right column of icons, vertically centered, hover tip on the left.
- **Targeting arrow** (`TargetingArrow.tsx`): red SVG line+arrowhead from source `data-cid` to cursor.
- **Mulligan** plays over the blurred/dimmed live board (`.mulligan-bg` + modal); full cards + tooltips.
- **Lobby** (`App.tsx` lobby branch) = centered card panel, gold-gradient logo, online dot, deck
  picker, big ▶ Play, 📖 Build deck (`.lobby-screen`/`.lobby-card`).
- **Deckbuilder = HS "book"** (`Deckbuilder.tsx` rewritten): collection paged 8/page (4×2 grid,
  `PER_PAGE`) with Prev/Next + "Page X/Y"; right sidebar = saved decks + editor (name, 30/30
  meter, Save/Cancel, in-deck list sorted by cost, click to remove). Reuses `cardArtIcon`/
  `cardStats`/`CardTooltip`. Server still validates every save.

- **Turn timer (server-authoritative, done):** `match.go` `turnLimit = 75s`. `time.AfterFunc`
  per turn (`scheduleTurnTimer` in `startTurn`); `onTurnTimeout` auto-ends via shared
  `endTurnLocked` (manual `EndTurn` also calls it). `turnGen` guards stale timers; a pending
  Discover at expiry auto-picks option 0; `stopTurnTimer` on game over. Snapshot carries
  `State.TurnSecs` (whole seconds left, `turnSecondsLeft()`); resync includes it. Client: `App`
  tracks `turnSecs` + `turnNum`; `GameScreen` `<TurnTimer key={turnNum} secs={turnSecs}>` —
  **keyed by turnNum so it remounts/resets each turn** (the bug fixed: it had reset only on
  `[secs]`, which is identical 75→75 across turns). Counts down locally, red+pulse ≤10s, shown
  for both turns. Tests: `TestTurnTimer`, `TestTurnTimerStop`.

- **Ops / deploy (done):**
  - `make build-web` builds the client → `web/static`. `.githooks/pre-commit` auto-builds +
    stages `web/static` on commit; enable with `make hooks` (sets `core.hooksPath`).
  - `make prod [PORT=8083]` = `git pull --ff-only` + `docker compose -f docker-compose.prod.yml
    up -d --build`. Single distroless binary serves embedded client + `/ws`; SQLite in named
    volume `vs-data` (`DB_PATH=/data/...`), `restart: unless-stopped`. `make prod-image` = build only.
  - One-liner deploy: `git clone … && cd vanillastone && make prod PORT=8083`.
  - **nginx in front MUST set `proxy_http_version 1.1` + Upgrade/Connection headers** or the ws
    handshake fails with "handshake request must be at least HTTP/1.1: HTTP/1.0" (coder/websocket
    rejects HTTP/1.0). Full config given to the user (single `location /` → `127.0.0.1:PORT`,
    `map $http_upgrade $connection_upgrade`). Not committed to the repo.

**Phase 10 design notes / open issues:**
- **Reconnect grace is per-username** (the `active`/`grace` maps), so it cooperates with the
  single-session kick: a takeover login (`kickExisting`) and a reconnect both resolve to
  swapping the seat's Sender pointer; the displaced client's later disconnect sees
  `Seats(old)=false` and does nothing. A near-simultaneous takeover-vs-disconnect can flip
  the opponent's banner false↔true out of order (rare; cosmetic only).
- **Restart-survival is out** (locked: in-memory match + sessions). A server restart still
  drops everyone (token invalid → re-login → no live seat → lobby).
- **Mid-mulligan reconnect after already submitting:** the resync `state` still carries
  `mulligan:true`, and the client shows the mulligan UI fresh (`mulliganSubmitted=false`),
  so a re-submit is possible — the server rejects it ("already mulliganed", harmless error
  toast). Minor wart; the common case (reconnect during play) is clean. Could add a
  per-player "already mulliganed" flag to `State` if it ever matters.
- **Animation timing (Q4) is the planned next session:** the client already receives the
  ordered event log per action; the work is to *replay* it with pacing + per-action
  highlight (and ideally CC0/CC-BY assets — game-icons.net / Kenney) so actions don't look
  instant. Keep engine changes minimal so this drops in on top.
- **Not browser-verified** at handoff (engine race-tested; client builds). Verify in two
  tabs: start a match, reload one tab mid-turn → it rejoins the same game (not the lobby),
  the other tab shows "Opponent disconnected…" then clears; kill a tab and don't return →
  after ~60s the other tab wins; reconnect mid-Discover re-shows the prompt.

**Phase 9 design notes / open issues:**
- **Pool is ~38 distinct cards** → a 30-deck with the 2-copy cap is feasible but tight
  (default = first 15 pool ids ×2). Plenty of room before it's a real constraint.
- **Fatigue is armor-absorbed** (HS rule) and emitted as a distinct `fatigue` event;
  drawn-card identity is hidden, so draws emit no event and overdraw emits a nameless
  `burn`.
- **The Coin** is added to player 2's hand *after* mulligan (so it can't be mulliganed),
  capped at 10 mana when played (a wasted Coin at 10 crystals, as in HS).
- **Mulligan is the second blocking phase** (after Discover) but uses its own
  `m.mulligan` state (simultaneous, multi-select), not `pending`/`Choose`.
- **No reconnect** — a mid-mulligan disconnect still forfeits (Concede clears mulligan).
- `validTarget`/`ruleMatches` + `cardView` shape now duplicated in **three** places
  (match, auth `poolCardView`, client) — keep in sync by hand.
- **Not browser-verified** at handoff (engine race-tested, HTTP smoke-tested, client
  builds). Verify in two tabs: build a deck (30/30, copy cap, save/rename/delete, 10
  cap), pick it in the lobby, queue → mulligan (toss some, opponent sees nothing until
  both keep), draw each turn, 2nd player has The Coin (+1 mana), deck count ticks down,
  fatigue when the deck empties.

**Phase 8 design notes / open issues:**
- **Overload deferred to Shaman** (Shaman-flavored; pointless while Mage-only).
- Hero power is hardcoded Mage Fireblast (`MageHeroPower()`); generalize when more
  heroes/classes exist.
- Weapons are plain (attack + durability). Weapon keywords (lifesteal/poisonous/
  windfury on the hero, deathrattle weapons) NOT modeled — `heroAttack` would need
  to honor them; add when such a weapon exists.
- Hero attack reuses the `attack` message (attackerId "selfHero") rather than a
  new message — keeps the client's attacker-selection flow uniform.
- Only Snare among secrets cares about attacks, and it's minion-specific, so hero
  attacks intentionally skip secret triggers. An Explosive-Trap-style "when your
  hero is attacked" secret would need hero attacks (and minion attacks) to fire it.
- **Not browser-verified** at handoff (engine race-tested; client builds). Verify
  in two tabs: Fireblast (button, target, once/turn, kills a 1-hp minion); equip a
  weapon, swing face + into a minion (retaliation), weapon breaks, re-equip
  replaces; frozen hero can't swing (Frost Snap your own? — only enemy freeze, so
  verify via an enemy freezing your hero); taunt blocks hero face.

**Phase 7 design notes / open issues:**
- **Secret cap (5) is untestable** with only 3 secret cards (no-duplicate blocks
  reaching 5 distinct); the guard exists, no unit test. Revisit when >5 secrets.
- **Mimic timing**: triggers right after the minion is summoned, BEFORE its
  battlecry (so a Discover battlecry's pause doesn't interleave with the secret).
  HS fires Mirror Entity slightly later; immaterial for current cards — note if a
  battlecry-vs-mirror ordering case ever matters.
- **One secret per event per side** is effectively all we have (each trigger has
  a single card + no-dup). `triggerSecrets` fires ALL matching anyway, accumulating
  `cancelled`, so multiple-same-trigger secrets would work if added.
- **Discover is the first blocking action.** Mulligan (Phase 9) can reuse the
  `pending`/`Choose` machinery. Note: only ONE pending choice at a time — a
  chain that needs nested choices isn't modeled.
- Counterspell-countered spells still **fire enemy on-cast secrets first**? No —
  only the defender's secrets exist; a spell triggers the OPPONENT's secrets, and
  the caster has none relevant. Fine for now.
- **Not browser-verified** at handoff (engine race-tested; client builds). Verify
  in two tabs: play a secret (opponent sees only 🔒 count), trigger each
  (Snare/Mimic/Nullify reveal + effect), Discover modal (3 filtered options, pick
  → card to hand, other actions blocked until picked), board-full summon discard.

**Phase 6 design notes / open issues:**
- **Aura is attack-only.** Pack Leader grants +1 Attack. A +Health aura needs
  current-health delta tracking on recompute (raise current when max rises, clamp
  when it drops) — deferred until such a card exists; `Aura.HP` field reserved.
- **Battlecry/deathrattle damage isn't lifesteal/poisonous-aware.** Those keywords
  apply to a minion's *combat* damage and to *spell* lifesteal only. No current
  card has a lifesteal/poisonous battlecry; revisit when one does (applyEffect
  doesn't know the source minion for triggers).
- **Silence health clamp uses the HS rule** (current health → min(current, new
  max)), not a damage-memory model — verified by the buff-then-silence test.
- `validTarget`/`ruleMatches` (+ stealth/taunt) still duplicated Go+TS — client is
  highlight-only, server authoritative. Keep in sync by hand.
- **Not browser-verified** at handoff (engine race-tested; client builds). Verify
  in two tabs: windfury double-swing, lifesteal heal, poisonous trade, divine
  shield blocks poison, stealth (untargetable then revealed on attack, AoE hits),
  spell damage on Cinder Bolt/Quake, Pack Leader aura, Hush silence (strips
  taunt/shield/buff/deathrattle), Drain Touch.

**Phase 5 design notes / open issues:**
- Windfury (multiple attacks/turn) deferred to wave 2 — `attacked bool` suffices
  now; becomes `attacksLeft int` then.
- No **Silence** yet (wave 2) — nothing strips keywords/shields except gameplay.
- Taunt+Stealth interaction N/A (Stealth is wave 2).
- Divine Shield re-grant (e.g. a buff that re-shields) not modeled — `divineShield`
  is a plain bool, fine until such a card exists.
- Hero freeze is inert (no weapons) — verify again when Phase 8 lands.
- **Not browser-verified** at handoff (engine race-tested; client builds). Verify
  in two tabs: charge/rush behavior, taunt blocking, divine shield pop, freeze +
  thaw across turns, Permafrost.

**Phase 4 design notes / open issues:**
- Only **on_play/on_death** wired. The other taxonomy events (on_damage,
  on_turn_start/end, on_summon, on_spell_cast…) are not yet emitted to triggers —
  add as data when their cards arrive (Phase 5/6).
- Battlecry targeting (`needsTarget`/`anyLegalTarget`) is evaluated **pre-summon**,
  so a friendlyMinion battlecry can't target the minion playing it (HS "another"
  semantics) — none of the current cards exercise this.
- Event log is **textual only**; the architecture's animation replay is still
  deferred to Phase 10. Events already carry enough (ordered, absolute ids).
- A true **deathrattle chain** (deathrattle damage deterministically killing
  another deathrattle minion) isn't unit-tested — no current deathrattle does
  fixed minion damage (husk=hero, wisp=random). The cascade loop is covered by
  the simultaneous-death test; revisit when such a card exists.
- **Not browser-verified** at handoff (engine fully unit/race-tested; client
  builds). Verify in two tabs: battlecries (Spark/Ember/Bog), deathrattles
  (Brood/Husk/Wisp), event log lines, disconnect → opponent wins.

**Phase 3 design notes / open issues:**
- Spells have **no spell-damage keyword** (Phase 6) and **no triggers** (no on-cast/secret) — plain numbers, resolve instantly.
- `validTarget`/`ruleMatches` duplicated server (Go) + client (TS) — client is convenience/highlight only, server authoritative. Keep in sync by hand for now.
- Spell discard = just removed from hand (no graveyard/deathrattle yet, Phase 4).
- Same zombie-match issue as Phase 2 (no teardown on game_over/disconnect).
- **Not browser-verified end-to-end** at handoff time (engine fully unit-tested; client compiles). Verify 4 spells in two tabs at :5173.

**Phase 1 design notes / open issues:**
- Turn identity = connection id (`p1,p2,...`, atomic counter), NOT username. Client compares `state.turn` to its own `joined.you`. Username is display only. Match ids `m1,...`.
- **Single session per account** (added 2026-06-21): on auth the server sends any
  earlier connection with the same username a `error{"logged in elsewhere"}`; that
  client clears its token, closes the socket, and returns to login (its disconnect
  forfeits any active match). Notify-not-force-close so the message reliably lands
  before the socket drops. Test `TestSingleSessionKicksPrevious`.
- Match cleanup on disconnect NOT handled yet (only queue removal). A dropped player leaves a zombie match — revisit when adding win/lose.
- No reconnect, no match teardown, no concurrency stress beyond test.
- Dev db `vanillastone.db` written at repo root (gitignored `*.db`).

**NOT done**: engine (minions/spells), decks table, cards, LICENSE, ASSETS.md, README.

---

## Commands
```
make dev          # run server, hot reload (docker compose up)
make dev-build    # rebuild after go.mod / Dockerfile change
make logs         # tail
make down         # stop
make prod         # build prod image
make help         # list targets
# web (after scaffold): make web-install ; make web
```

---

## Animations first cut (done) — 2026-06-21
Lightweight DOM/WAAPI layer over the already-settled snapshot — NOT the full buffer/replay
the architecture envisions (states still apply instantly; we animate *on top* of the new
state). Engine untouched.
- **Lobby "Finding opponent" state** (`App.tsx`): the `waiting` phase now reuses the lobby
  card instead of the old plain screen. The Play button turns **yellow + pulsing**, shows a
  spinner + animated ellipsis, and **doubles as cancel** (click again → `onBackToLobby` →
  `enter_lobby`). Deck picker / Build / Log out disabled while searching. CSS: `.play-btn.waiting`,
  `.play-spinner`, `.ellipsis`, `.cancel-hint` + keyframes `waitpulse/spin/ellipsis`.
- **Card movement (HS-style "to the table")** (`animate.ts`): minions **fly in** to the board
  from their controller's hand region — `GameScreen` `useLayoutEffect` diffs board minion ids vs
  the previous render (`prevCids` ref, null on first render so opening boards don't animate) and
  calls `flyIn(cid, '.hand'|'.opp-hand')` for any new uid (covers plays, tokens, summon
  battlecries/deathrattles). Non-minion plays (spell/secret/weapon) use `playGhost('hand-i')`:
  clones the clicked card and flies the copy up to screen center, growing + fading (called from
  `App.onHandCard` + `App.onChar` right before the `play_card` send, while the card is still in
  the DOM). CSS: `.play-ghost`.
- **In-game action animations** (`web/src/game/animate.ts`, driven by `GameScreen`): `App`
  forwards each `state` action's `events` (skipping match_start/resync) as `anim={seq,events}`.
  `GameScreen` effect maps event ids → `data-cid` (minion uid starts "u"; hero pid → selfHero/
  oppHero via `snap.you`) and plays: **attack** → `lunge` (attacker translates 55% toward target
  and back), **damage** → `hitFlash` (shake + brighten, 160ms delay to hit at lunge apex) +
  floating red `-N`, **heal** → floating green `+N`. Best-effort: skips if the element is gone
  (dead minion). CSS: `.dmg-pop`.
- **Iteration 2 (2026-06-21):** timings slowed (lunge 900ms, summon fly-in 700, draw 1100,
  card-to-table 1300, hit flash 650 / 360ms delay, numbers 1400). **Dead-target lunge fixed** —
  `GameScreen` caches each character's last screen rect (`rectCache`), so an attacker whose target
  died this action still lunges to where it stood (was bailing out → minion-vs-minion looked
  static). **Death fade** — `minionCache` (rect + outerHTML per minion per render) lets `deathPuff`
  re-create a removed minion where it stood and hold it ~500ms (so the hit plays) then fade/shrink.
  **Card-draw animation** — hand growing flies the new card in from the deck pile (real card for us,
  face-down back for the opponent). **Ready-glow bug fixed** — `.ready` (green "can attack" glow)
  now gated on `myTurn` (Board got a `myTurn` prop); your minions no longer look actionable on the
  opponent's turn.
- **Iteration 3 (2026-06-21) — hero power + cast reveal + login screen:**
  - **Hero power animation** — on a `heropower` event the power's glyph (🔥) flies from the
    caster's button to the target (`projectile` in animate.ts); paired damage flash delayed to
    560ms to land on arrival. Opp hero-power button got `data-cid="oppHeroPower"` as a source.
  - **Cast reveal (only engine change this round):** opponents' hands are hidden, so the server now
    emits a **`play` event carrying the played `CardView`** (`Event.Card`, `emitPlay` in match.go) for
    minion/spell/weapon plays — **NOT secrets** (stay hidden). Client shows the opponent's latest
    cast **center-stage for ~2.5s** (`cast` state in `GameScreen`, swaps + restarts timer on a new
    play; `.cast-show`/`.cast-card` CSS). Own plays aren't showcased (we have the play-ghost). This
    **replaced the old face-down `oppCardGhost` heuristic** (removed). `play` added to
    `format.ts` (log line + 🃏 icon) and `protocol` Event kind. All Go tests green.
  - **Login/register screen** restyled to match the lobby card (`.lobby-screen`/`.lobby-card`,
    gold logo, `.auth-input` fields, ▶ Login primary + Register secondary, Enter submits, spinner
    while connecting).
- **Iteration 4 (2026-06-22):** more polish + online player list.
  - **Burn/fatigue → center:** overdraw burns a card back flying deck→center; fatigue flies a **black
    skull card** deck→center then the hero flash. Opponent fatigue also **previews on the left** (cast
    slot, "Opponent fatigued"). Deck pile now **scales by fullness** (30=100%: 3 layers >66%, 2 >33%,
    1 above empty, 0 = none) with an always-visible dashed **outline** marking its spot.
  - **Keyword tooltips** are now a **custom instant CSS tooltip** (`.kw-tip`) — replaced the slow
    native `title`; each badge shows "Label — explanation". Card tooltip meta dropped the "·".
  - **Lobby/flow:** waiting button = spinner only (no dots); a **"Match found!"** splash (⚔️ + you-vs-opp
    + spinner, `phase 'matchfound'`, `matchFoundTimerRef`) shows ~2s before the mulligan. Login screen
    restyled to the lobby card. **Favicon** = card-back sigil (`web/public/favicon.svg`).
  - **Online player list (new):** `protocol.Lobby` gained `Players []PlayerInfo{Name,Status,Vs,MatchID}`
    (status lobby|waiting|in_game). `transport` tracks a `queued` set (set/cleared in find_match/
    enter_lobby/disconnect) and builds the list in `broadcastPresence` (dedup by username; in_game →
    opponent via `Match.Players()` + `Match.ID`). Client renders it in the lobby card (`.player-list`,
    status dots + "⚔️ vs X"). **`MatchID` is in the payload to prep spectator mode** (not used yet).
  - **Spectator mode = NOT built** — next step. Needs: server `matches` lookup by id, `Match` observers
    slice + spectator snapshot (hands shown vs hidden — decide), `spectate{matchId}` routing + observer
    disconnect cleanup. The player list (with matchId) is the prerequisite and is done.
- **Limitations / next:** dead attackers (not targets) can't animate (element already removed); the
  settled snapshot); numbers update instantly under the flash; only attack/damage/heal covered —
  summon/death/buff/freeze/shield/secret/equip not yet. A real buffered replay (queue snapshots,
  pace events, then settle) is still the bigger follow-up; this is the "see how it looks" cut.

## Gameplay-feel polish (done) — 2026-06-22
Four TASKS.md items, mostly client; one additive server field + one new server msg.
- **Drag-to-play (HS-style), TASKS #1.** Hand cards are now dragged onto the board
  instead of click-to-play. Two interaction modes share one drag state in
  `GameScreen` (`drag`/`dragRef`/`pressRef`): **hold-drag** (pointerdown → move past
  a 6px threshold → release) and **click-drag-click** (a click with no movement lifts
  the card "sticky"; the next click anywhere drops it). Window-level pointer listeners
  (`pointermove`/`pointerup` + a capture-phase `pointerdown` that drops a sticky card
  before the hand handler can start a new press); Esc or a drop off the play area
  cancels (card stays in hand). A floating `.drag-card` clone follows the cursor.
  - **Positional board insert.** While dragging a **minion**, the friendly row opens a
    gap at the target slot: `Board` got a `dropIndex` prop that splices a `.drop-slot`
    placeholder in. The slot index is computed from the cursor x against minion centers
    **captured at drag start** (`centersRef`) so the live row shifting to make room
    doesn't cause jitter. On drop, `onHandCard(i, card, pos)` sends `play_card` with
    `pos`. Spells/weapons/secrets ignore pos; a targeted card dropped on the table arms
    the existing targeting flow (a targeted minion battlecry carries its `pos` through
    `PendingSpell.pos` so it lands at the chosen slot after the target is picked).
  - **Server (authoritative placement):** `protocol.PlayCard` gained `Pos *int`
    (omitempty; nil = append, back-compat). `Match.PlayCard` is now a wrapper over
    **`PlayCardAt(c, handIndex, targetID, pos)`** (pos<0 = append). New `placeAt` moves
    the just-summoned minion (always appended by `summonMinion`) to the clamped slot.
    Transport passes `p.Pos` (nil → -1). Test `TestPlayCardAtPosition` (front/middle/
    append/clamp ordering). **Tokens still append** (only the hand-played minion is
    positioned).
- **Deck pile sized like hand cards, TASKS #2.** `.deck-card-back`/`.deck-outline` are
  now 88×116 (was 50×68) to match hand cards; `.deck-pile` box + stack offsets bumped,
  `right` 40→24px. Pure CSS.
- **Opponent-discovering peek, TASKS #3.** While a player is mid-Discover the opponent
  used to see a frozen board. New S→C **`opp_discover{count}`** (`TypeOppDiscover`):
  `startDiscover` sends it to the non-chooser; `Reattach` re-sends it if you reconnect
  while the opponent is choosing. Client (`App` `oppDiscover` state, cleared on any
  `state`/`game_over`/lobby) renders `.opp-discover` near the opponent's hand — a small
  row of face-down `card-back`s with an "Opponent is discovering…" label, **no overlay,
  no darkening** (the chooser's own modal is unchanged).
- **Summon-pop for tokens, TASKS #4.** Minions summoned by a battlecry/deathrattle
  (Bogling, Hatchling, a Mimic copy) now animate with **`summonPop`** (scale-in with a
  slight overshoot, 700ms = same speed as a played minion's fly-in) instead of flying
  in from the hand region. `GameScreen`'s board-diff layout effect distinguishes them
  via this action's events (read through `animRef`): every summon emits a `summon`
  event, but only the hand-played minion also has a `play` event naming it — that one
  (first match) flies in from `.hand`/`.opp-hand`, the rest pop in place.
- Builds green: `go build ./...`, `go test -race ./...`, `pnpm build` (re-embedded
  `web/static`), `gofmt` clean. **Not browser-verified by the assistant** — eyeball in
  two tabs: drag a minion between two others (gap opens, lands there), click-lift then
  click-drop, drag a spell onto an enemy (arms target), opponent Discover shows the peek
  (no darkening), play Bog Warden / kill a Brood Mother (token pops), deck pile is
  card-sized.

### Follow-up refinements (same session, browser-iterated with the user)
- **Floating-clone styling bug (recurring).** Any card cloned/rendered outside `.hand`
  loses the `.hand .card` look (size/flex/fonts/color) — it only had `.drag-card`/
  `.play-ghost`'s own few props. Fix pattern: add the class to the shared card-style
  groups (`.hand .card, .cast-card, .drag-card, .play-ghost { … }` + the `.name`/`.art`/
  `.stats` groups). Applied to the **drag clone** AND the **play-to-center ghost**
  (`playGhost`). Watch for this on any future detached card visual.
- **Targeted-card aiming (spells + targeted-battlecry minions).** Grabbing a card whose
  `target` is set lifts it toward the table (`.card.aiming`: translateY up + red glow)
  and arms the aim line immediately — no drop-on-table step. Release over a legal target
  (`elementFromPoint` → `data-cid` → `onChar`) casts; release elsewhere keeps it armed
  (click the target). Click the card again = cancel. Handled in `GameScreen`
  `onCardPointerDown` (branch on `card.target && !== 'none'`) + `aimRef` in the window
  `pointerup`. Targeted minions **append** (this flow doesn't pick a slot); untargeted
  minions keep drag-to-position.
- **Original card lifts out of hand while dragging** (`.card.dragging { visibility:
  hidden }`) — leaves a gap, returns on cancel/failed drop (was dimmed-in-place).
- **Triggered secret reveal.** The `secret` event now carries the card (`Card` in
  `triggerSecrets` — public once triggered); the client reveals it center-left via the
  cast preview with a "🔮 Secret triggered!" label, so BOTH players see which secret
  fired.
- **Opponent-discovering peek** repositioned to `top:21%` (band between opp hand and
  hero, on plain table bg) and sized to deck/hand cards (88×116).
- **Instant tooltips.** Own hero secret gems use a custom `.secret-tip` (was the slow
  native `title`). Hovered board minion gets `z-index:50` so its keyword/card tooltips
  aren't clipped behind the right-hand neighbor.
- **Turn-timeout during Discover** (pre-existing, confirmed): `onTurnTimeout`
  auto-picks option 0 into hand (burns if full), clears `pending`, ends the turn — no
  soft-lock. NOTE: that path skips `finish()`; harmless for current Discover cards (no
  deaths), revisit if a Discover battlecry can kill mid-resolve.

### Card-face redesign (browser-iterated)
- **Shared `web/src/game/CardFace.tsx`** renders the inner face for EVERY card surface
  (hand, drag clone, cast preview, discover, mulligan, deckbuilder) so they're identical:
  cost gem (top-left), name **title banner** (top, padded to clear the cost gem so long
  single-line names don't hide behind it), body, type/tribe band (bottom), and atk/hp
  gems in the bottom corners. **Board minions** (`Board.tsx`, still `.minion`/`.m-*`
  classes) match the same look via shared CSS groups (parchment frame, banner, gems,
  `.m-type` band) — a minion looks the same in hand and on the table.
- **Numbers are INSIDE the frame** now (HS-style): cost/atk/hp gems at the inner corners
  (`top/left/right/bottom: 4px`) with a dark outline ring; **20px** gems. Cards enlarged
  to **100×132** (deck pile / opp hand / discover / opp-discover peek all synced).
- **Card body shows rules text** (`CardFace` art area) instead of the type emoji — vanilla
  minions with no text still show the icon. `format.ts` `cardDesc()` strips periods and
  turns each sentence break into a line break ("Deal 2 damage to a character⏎Lifesteal");
  `.card-desc` uses `white-space: pre-line`. Full text still in the hover `CardTooltip`.
- The recurring **detached-clone styling bug** (above) applies here too: any new card
  surface must join the shared `.hand .card, .cast-card, .drag-card, .play-ghost { … }`
  groups (and the `.name`/`.art`/`.type-band`/`.stat`/`.card-desc` groups) or it renders
  unstyled. `book-card` and `.minion` are wired in by hand.
- **Forward-compat for tribes:** `cardTypeLabel()` returns the broad type now; when minions
  gain a race (Beast/Dragon with the Hunter/class system) return it there and the bottom
  band shows it. The board `.m-type` is hardcoded "Minion" — swap to the race then.

## Direct invites (done) — 2026-06-22
TASKS #1: challenge a specific lobby player instead of the global queue. Lobby-only,
no engine change.
- **Protocol** (`internal/protocol`): C→S `invite{target,deckId}` / `invite_cancel` /
  `invite_respond{from,accept,deckId}`; S→C `invite_received{from}` /
  `invite_declined{by}` (to the inviter — refused OR target became unavailable) /
  `invite_cancelled{from}` (to the invitee — inviter withdrew / left).
- **Server** (`internal/transport`): `Server.invites map[string]inviteRec` (inviter
  username → `{target, deckID}`), guarded by `s.mu`. **One outstanding invite per
  inviter** (a second is rejected "cancel your current invite first"). Inviter's deck is
  **locked at send time** (inviteRec.deckID); invitee's deck rides `invite_respond`.
  Helpers: `byNameLocked`/`byName` (one client per username — single-session kick
  guarantees that), `availableLocked` (online, no live match, not queued).
  `handleInvite` validates both sides free + target online, records the invite, sends
  `invite_received`. `handleInviteRespond`: decline → `invite_declined` to inviter;
  accept → re-check both free, `deckFor` each side, `clearInvites` both, drop stale
  active/queued, `lobby.StartMatch(inviter,deckA,invitee,deckB)` (**inviter acts
  first**), `attachMatch`, broadcast. `clearInvites(name)` drops every invite name is
  part of and notifies the other side (sent to its target = `invite_cancelled`; aimed at
  name = `invite_declined` to the inviters). Wired into disconnect cleanup,
  `find_match`, and `enter_lobby` so leaving the lobby tears down pending invites.
- **Lobby** (`internal/lobby`): new `StartMatch(a,deckA,b,deckB)` — direct pairing that
  bypasses the queue but reuses the same `matchID` counter (ids never collide).
- **Client** (`App.tsx`, `protocol.ts`, `index.css`): each lobby-status player row (not
  you) shows a `⚔️ Invite` button; while you have an outgoing invite that row becomes
  `invited… ✕` (cancel) and every other invite button is disabled (one at a time).
  Incoming invite → a centered modal ("X challenges you" + **deck `<select>`** +
  Accept/Decline). State: `invitedName` (outgoing), `incomingInvite` (newest incoming),
  `inviteDeck`. `selectedDeckRef` mirrors the deck picker so the dep-less ws closure can
  default the prompt's deck. Cleared on `match_start` / back-to-lobby. Server allows
  multiple inviters targeting one player; the client tracks only the newest incoming
  (newest wins) — acceptable wart.
- Tests: `internal/transport` `TestInviteAcceptStartsMatch` (prompt + inviter-first
  match), `TestInviteDeclined`, `TestInviteOneOutstanding` (second rejected, cancel
  frees + notifies). `go test -race ./...` green, `go vet`+`gofmt` clean, web
  `tsc`+`vite` clean, `static/` re-embedded. **Not browser-verified** — eyeball in two
  tabs: invite, accept (match starts, inviter first), decline, cancel-then-reinvite,
  inviter leaves (invitee prompt clears), pick a deck in the prompt.

## Card-gem bleed fix + longer cast reveal (done) — 2026-06-22
- **Hand gem bleed.** Fanned hand cards overlap (`.hand .card { margin: 0 -14px }`). The
  atk/hp corner gems carry `z-index: 2`, so as positioned positive-z elements they painted
  ABOVE every sibling card regardless of DOM order — each card's hp gem (bottom-right) bled
  onto the neighbour, landing by its atk gem ("hp + atk in the same place"). Fix: one line,
  `.hand .card { isolation: isolate }` — each card gets its own stacking context so the
  gem stays local and later cards cover earlier ones via normal paint order. `CardFace.tsx`
  stat logic was already correct (spells/weapons unaffected). Trade-off: at rest an
  overlapped card's hp sits under its neighbour (HS fan behaviour); hover lifts it
  (`z-index:30`) to read full stats.
- **Opponent cast reveal** (`GameScreen.tsx` `showReveal`) bumped 2500ms → **5000ms**. A new
  cast before it lapses swaps + resets the timer (no pileup).

## Spectator mode (done) — 2026-06-22
Watch a live match from a chosen **player's** point of view (their hand revealed, opponent's
hidden), read-only. A spectator is a **seat-bound observer** that receives byte-identical
snapshots to the watched player — the client renders `Self`/`Opp` from the snapshot, so it
"just works", only inputs are disabled.
- **Protocol**: C→S `spectate{target}` (username). S→C `spectate_start{target}` (sent just
  before the first snapshot so the client flips to its read-only view first). State /
  game_over / opp_discover reuse the existing types. Stop = `enter_lobby`.
- **Match** (`internal/match`): `observers map[Sender]int` (observer → watched seat).
  `SeatOf(c)`, `AddObserver(c, seat)` (pushes a resync snapshot + an opp_discover indicator
  if mid-Discover; false if over), `RemoveObserver(c)`. Per-seat sends route through
  `fanout(seat, b)`: `sendStateTo`/`sendResyncTo` now build via shared `stateFor(i, …)` and
  mirror to observers of seat i; `broadcast` (game_over) hits all observers; `startDiscover`
  sends spectators of BOTH seats the hidden-face `opp_discover` (never the real Discover —
  only the chooser may pick, and a modal would block the spectator). Hidden info is safe:
  spectators only ever see the watched seat's `selfView` + the opponent's `oppView`.
- **Transport** (`internal/transport`): `Client.spectating atomic.Pointer[Match]`.
  `handleSpectate` validates (not in own match, target online + seated), sends
  `spectate_start`, then `AddObserver`. `stopSpectating` deregisters on `enter_lobby`,
  `find_match`, and disconnect.
- **Client** (`App.tsx`, `GameScreen.tsx`, `protocol.ts`, `index.css`): lobby in-game rows
  get a `👁 Watch` button (`onSpectate`). `spectating` state + `spectatingRef` gate the four
  action handlers (`onHandCard`/`onHeroPower`/`onChar`/`targetable`) so nothing is sent.
  `povRef = msg.you` drives "whose turn" + event-log POV from the watched seat (identical to
  `meRef` for real players). `GameScreen` takes `spectating`: shows a "👁 Spectating X +
  Leave" banner (replaces End Turn/Concede), disables hero power / hero-attack / hand drag,
  and the turn flag reads from the watched player's name. Mulligan UI is skipped for
  spectators.
- **"Being watched" badge**: S→C `spectators{names []string}` is pushed to the two PLAYERS
  (not the spectators) whenever the watcher set changes — `Match.notifySpectators()` (sorted
  `Sender.Name()`s) called at the end of `AddObserver`/`RemoveObserver`. Client shows a
  `👁 N` badge top-left of the board (`.watchers`) whose hover panel lists the usernames.
  Reset on `match_start` and back-to-lobby.
- **Wart**: spectator sees a face-down `opp_discover` indicator during the watched player's
  Discover (not the 3 options), to avoid a blocking modal. Live-only (no replay); leaving or
  the match ending returns to the lobby.
- Test: `internal/transport` `TestSpectateMirrorsPlayerPOV` (POV `You`, hand revealed vs
  opponent hidden, follows turn pass, game_over delivered). `go test -race ./...` green, vet
  + gofmt clean, web `tsc`+`vite` clean, `static/` re-embedded. **Not browser-verified** —
  eyeball in three tabs: start a match in two, `👁 Watch` from a third, confirm it tracks
  plays/turns, hands hidden correctly, Leave returns to lobby.

## Card pool rebuild — class + rarity, renamed staple set (done) — 2026-06-22
First real cut of the renamed card pool. **Source pool = a familiar reference set's Mage class
(15) + neutral (~105) cards**; only the ones whose mechanics the engine already supports were
imported, the rest deferred to keyword waves.
- **Card model:** `cards.Card` gained `Class` (`neutral`/`mage`) + `Rarity`
  (`common`/`rare`/`epic`/`legendary`). **Color is now driven by CLASS, not type** — neutral =
  default parchment, Mage = blue. The old per-type coloring (spell/secret/weapon) is gone; secrets
  keep `TypeSecret` for the engine but render blue (Mage) like any spell. Both fields flow through
  `protocol.CardView` → client. Client adds a **rarity gem** on every card face (`CardFace`,
  `.rarity-gem`), one `cardColorClass()` helper replaced 6 hardcoded per-type class blocks
  (GameScreen/App/Deckbuilder/mulligan/discover/cast/drag). Hand cards now zoom on hover even when
  disabled (`.hand .card:hover`).
- **Card files split:** `cards.go` keeps types/helpers and assembles `set` in `init()` from
  registries; **`neutral.go`** (25 minions + 3 tokens + The Coin) and **`mage.go`** (5 collectible
  + the Mage hero power) hold the data. **All prior demo cards purged** from the shipping pool.
- **Hero power renamed** (the old name was third-party IP) → **Fire Dart** (`fire_dart`, 2 mana,
  1 dmg). `MageHeroPower()` updated.
- **Buildable now (33 collectible):** Mage 5 — 4 secrets (gain-armor / copy-minion / counter-spell
  / destroy-attacker) + 1 AoE-freeze spell; Neutral 25 — vanilla + single-keyword bodies (taunt /
  divine shield / stealth / windfury / poisonous / charge+shield / taunt+shield), battlecry
  (heal / silence / summon / freeze), deathrattle (face dmg / summon), spell-damage +5 legendary; 3 card-draw minions (EffectDraw).
- **Tests decoupled:** engine white-box tests no longer depend on the shipping pool — old demo
  cards live as `match/fixtures_test.go` `testCards` (resolved via `getCard`/`testDeck`); summon
  fixtures point at the real `broken_golem` token; the Discover fixture uses the minion pool (the
  shipping spell pool is too small to offer 3). `go test -race`, `vet`, `gofmt`, client build all
  green; `static/` re-embedded.
- **IP hygiene:** no third-party product/card names anywhere in **code or comments** (scrubbed
  existing `Fireblast`/"Hearthstone" mentions too). Reference real names only in private notes,
  never in-repo.

### Missing-keyword backlog (skipped cards, by mechanic — pick waves highest-count-first)
Most of the reference set is skipped pending engine work:
- **Edge triggers** (~22) — on_turn_end / on_spell_cast / on_summon / on_any-death. Biggest unlock;
  already noted deferred.
- **Draw** — ✅ `EffectDraw` added (caster draws Amount, fatigue/burn via `drawCard`, no event =
  hidden). Imported the 3 pure-draw cards (Pilfer Imp DR-draw, Sapphire Drake BC-draw + spell dmg,
  Vael Emberscribe DR-draw + spell dmg). The other ~5 draw-tagged cards still need triggers
  (Cult Master, Auctioneer = on_spell_cast / on_death), opponent-draw, or destroy — still skipped.
- **Cost modification** (~9) — spells/minions cheaper or pricier (hand-cost aura).
- **Destroy a minion** (~6). **Enrage** (~5). **Adjacency / grant-keyword buffs** (~5).
- Scattered singles: transform, bounce-to-hand, tribe synergy, temp "this turn" buff, conditional
  stats, random card generation. Two trivial near-misses: an **enemy-character** target rule and a
  **friendly-hero heal** target each unlock one more card.

## Edge triggers — event bus + Batch 1 (done) — 2026-06-22
First edge-trigger wave (the "Phase 4 extended" work that was deferred). Pool now **40** collectible
(34 neutral / 6 mage).
- **Event bus:** new `cards.EventType`s `on_turn_start` / `on_spell_cast` / `on_friendly_summon` /
  `on_friendly_death` / `on_any_minion_death`, plus a `TargetSelf` rule (effect aims at its own
  source minion). `Card.TriggersFor(when)` generalizes `Battlecry()`/`Deathrattles()`.
- **Dispatch:** `match.fireTriggers(controller, when, subject)` — iterates a snapshot of the
  relevant board(s), fires each non-silenced minion's matching trigger via `applyEffect` from that
  minion's perspective; "other"-scoped events skip the subject. Wired into `startTurn`
  (turn-start), `summonMinion` (friendly summon), `playSpell` + **`playSecret`** (cast — secrets
  count as spells, fires even when countered), and `resolveDeaths` (friendly + any death, fired for
  survivors regardless of the dead minion's Silence). `on_turn_end` is defined-but-unwired (no
  Batch-1 card uses it; wire in Batch 2).
- **Batch 1 cards (7):** Arcane Wyrmling (**mage**, cast→+1 atk), Bazaar Crier (cast→draw),
  Adept Tutor (cast→summon 1/1 Pupil token), Dagger Tosser (friendly summon→1 dmg random enemy),
  Siege Engine (turn-start→2 dmg random enemy), Carrion Fiend (any death→+1 atk), Cabal Overseer
  (your other minion dies→draw). 8 new white-box tests (incl. silence-suppresses-trigger,
  opponent-turn-doesn't-fire). `go test -race`, `vet`, `gofmt`, client build all green.
- **Deferred to Batch 2:** the marquee Mage legendary (cast→add a spell to hand = card-generate) +
  Ethereal Arcanist (`on_turn_end` + "control a Secret" condition) + Wild-Pyromancer-style
  all-minion AoE + temp "this turn" buffs + random-friendly targeting + on_heal. Needs:
  card-generate effect, a trigger condition field, an all-minions area, a random-friendly target.

## Edge triggers — Batch 2 (done) — 2026-06-22
Second edge-trigger wave (Mage-focused). Pool now **44** collectible (35 neutral / 9 mage).
- **Engine primitives:**
  - New `cards.EventType` **`on_turn_end`** — wired in `match.endTurnLocked` (fires the ending
    player's triggers BEFORE the thaw/flip; deaths resolve in the trailing `finish()`).
  - New `EffectGenerate` + `Effect.Generate` (card id) — `applyEffect` adds that card to the
    caster's hand (burns if hand full, emits a nameless **`generate`** event — identity hidden,
    like a draw).
  - New `AreaAllMinions` — `damageTargets` returns every minion on BOTH boards (incl. the source).
  - New `Trigger.Condition` (`cards.TriggerCondition`, currently **`controlSecret`**) — gates an
    edge trigger; `fireTriggers` now iterates `card.Triggers` (not `TriggersFor`) and checks
    `m.condMet(cond, owner)` before firing. Battlecry/deathrattle paths still use the effect-only
    helpers (no conditional battlecries yet).
- **Cards (4):** **Pyrebolt** (mage, 4c spell, 6 dmg any — the generated card + fills the burn gap),
  **Emberforge Magus** (mage legendary, 7c 5/7, cast→add a Pyrebolt to hand), **Warded Scholar**
  (mage, 4c 3/3, on_turn_end if you control a Secret→+2/+2 self), **Ashflame Zealot** (neutral, 3c
  3/2, after you cast a spell→1 dmg to all minions). All follow well-worn staples, original names.
- **Client:** `protocol.ts` + `format.ts` add the `generate` event kind (log line + 🪄 icon).
  Pyrebolt is `TargetAny` (already supported); the new areas/effects are untargeted (no client
  targeting change).
- Tests: 4 new white-box (`match_test.go`): turn-end conditional buff fires WITH a secret /
  doesn't WITHOUT, on-cast generate adds the card, all-minion AoE hits both boards + self.
  `go test -race ./...` green, `go vet`+`gofmt` clean, web `tsc`+`vite` clean, `static/` re-embedded.
- **Still deferred to a later batch:** temp "this turn" buffs, random-friendly targeting, on_heal,
  cost-modification — no Batch-2 card needed them.

## Card pool expansion — vanilla curve + Mage toolkit (done) — 2026-06-22
Pure data, **no engine change** (all existing mechanics). Pool now **57** collectible (43 neutral /
14 mage). Follows well-worn staple stat-lines/effects; all names original (see "Legal rules").
- **Neutral vanilla bodies (8)** — fill the curve, no abilities except where noted: Crag Raptor
  (2c 3/2), Reedbank Snapper (2c 2/3), **Frostpeak Sentry** (2c 2/2 Taunt), Emberfang Brute
  (3c 5/1), Tundra Yak (4c 4/5), Dune Tortoise (4c 2/7), Crag Ogre (6c 6/7), War Colossus (7c 7/7).
- **Mage toolkit (5 + 1 token):** Frost Lance (2c, 3 dmg + Freeze a character), Arcane Burst
  (2c, 1 dmg all enemy minions), Flame Wave (7c, 4 dmg all enemy minions), **Mirror Conjuring**
  (1c, summon two 0/2 Taunt `conjured_mirror` tokens — first SPELL that summons; the path existed
  for battlecries, now exercised by a spell too), Tower Archmage (6c 4/7, Spell Damage +1).
- **No client change** — no new event kinds or target rules (Frost Lance reuses TargetAny+Freeze,
  the AoEs are untargeted, the summon is untargeted). Cards render from the CardView snapshot, so
  `static/` did not need a rebuild.
- Tests: `TestSpellSummonsTokens` (the new spell-summon path). `go test -race ./...` green,
  `vet`+`gofmt` clean. **Not browser-verified** — eyeball: cast Mirror Conjuring (two taunts),
  Flame Wave / Arcane Burst board clears, Frost Lance freeze, vanilla bodies in the deckbuilder.

## Open / next
- **PLAN — support the full Mage + neutral set: see [TASKS.md](TASKS.md).** Phased cheapest→hardest
  (A pure-data purge-gap fill → B destroy → C transform → D enrage → E grant-keyword/adjacency →
  F cost-mod (big) → G bounce → H singles). Hunter still OUT of scope.
- **Phase A DONE (2026-06-22):** re-added renamed Rush/Lifesteal/Aura/Discover/Weapon staples +
  Charge/Poisonous pad + a 10/10 burn (12 cards). Every engine-supported mechanic now has ≥1
  collectible card, guarded by `cards.TestSupportedMechanicsHaveCards`. Pure data.
- **Phase B DONE (2026-06-22):** `EffectDestroy` (sets health=0, ignores Divine Shield; death +
  deathrattle resolve in `finish()`; emits `destroy` event) + `TargetEnemy` (enemy minion/hero) +
  `TargetFriendlyHero` rules. Cards: Banish Rite, Headsman, Bombard Captain, Tavern Medic. Pool
  69→73 (56 neutral / 17 mage). Client mirrors the `destroy` event + 2 rules; `static/` re-embedded.
- **Phase C DONE (2026-06-22):** `EffectTransform` (+ `Effect.Transform`) replaces the target minion
  in place (same uid/slot, fresh stats, drops enchants/keywords/statuses, summon-sick, no
  deathrattle). Client mirrors the `transform` event (log + 🔄).
- **CLASSIC PIVOT (2026-06-22) — scope locked: clone the HS Classic set.** Target = **105 neutral +
  15 Mage = 120** collectible (Mage-only game; original names/art, 1:1 mechanics). See TASKS.md.
  - Mage trimmed to its **9 in-Classic clones** (Mana Wyrm/Ice Barrier/Mirror Entity/Counterspell/
    Vaporize/Ethereal Arcanist/Blizzard/Pyroblast/Antonidas). Removed Basic-set + invented Mage
    cards (Frostbolt/Arcane Explosion/Flamestrike/Mirror Image/Polymorph/Archmage clones +
    pyrebolt→token for Antonidas). The 6 remaining Mage cards are mechanic-gated (cost-reduction,
    adjacency, conditional, random-generate, retarget secret — see TASKS).
  - Rush/Lifesteal/Discover cards removed earlier too (those keywords postdate Classic); the
    ENGINE keeps all these features (transform, rush, lifesteal, discover) — cards return if scope
    ever expands past Classic. Transform/spell-summon engine tests moved to fixtures
    (`hex_bolt`/`twin_summons`). Coverage test → `TestClassicMechanicsHaveCards`.
  - Pool **61** after the Mage trim.
- **Neutral audit DONE (2026-06-22):** mapped the current neutral pool to the real Classic-105
  (full set captured from the wiki). Trimmed 16 (Basic-set vanilla bodies + invented Phase-A/B
  cards), fixed 6 stat/cost mismatches → **36 faithful Classic neutral** of 105. Removed
  weapon/aura/destroy/enemy-target/friendly-hero demo cards; those ENGINE features stay (now
  card-less, cards return with their gated wave). 4 engine tests moved to fixtures
  (`banish_rite`/`headsman`/`bombard_captain`/`tavern_medic_fx`). Coverage test trimmed to
  Classic mechanics with cards. Pool now **45** (36 neutral / 9 mage). See TASKS.md for the
  neutral tracker (36/105; ~6 buildable-now left + ~63 mechanic-gated).
  - **NEXT:** Wave 1 = clone the remaining buildable-now Classic neutrals (data), then build the
    gated mechanics in waves (enrage → tribes → cost-mod → …) toward 105 neutral + finish Mage to 15.
  - **SOURCE OF TRUTH for the card set: `.notes/classic-mapping.md`** (gitignored, private — real
    names live ONLY there, never in the repo). Holds all 120 (15 Mage + 105 neutral) with our id +
    cost/stats/effect/tribe + status (IN / FIX / BUILD / GATE:`<mechanic>`). **New session: read it
    before doing any card work.** Wave 1 batch from it: BUILD Doomsayer / Hogger / Onyxia / Arcane
    Devourer; FIX Cairne(hornelder_chief) token→5/5 and Ravenholdt(veiled_assassin)→7/5. Current
    pool **45** (36 neutral / 9 mage).
- **Batch 2 edge triggers: DONE** (see "Edge triggers — Batch 2"). card-generate + trigger-condition
  + on_turn_end + all-minion AoE landed with the Archmage/Ethereal-Arcanist/Pyromancer equivalents.
- **Batch 3 edge triggers / mechanics (next):** temp "this turn" buffs, random-friendly targeting,
  on_heal trigger, cost-modification (~9) — pick the highest-count unlock.
- **Card pool expansion: DONE** (see "Card pool expansion"). Mage now 14, core neutrals 43, pool 57.
- **Expand the card set further (optional):** more removal/tempo, enrage (~5), destroy (~6) need
  engine work (see backlog). Original art still TODO (emoji only).
- **Expand the card set (older note) — Mage is INCOMPLETE + core neutrals are thin.**
  **DESIGN RULE (locked): faithfully follow well-established genre card designs — match cost,
  stats, and effect — and give each a wholly original name + art.** We rebuild a familiar card
  pool (a Mage class set + a neutral "core" set), not invent balance. Mechanics aren't
  copyrightable; names/art are, so every card gets an original name + original/CC0 art (repo is
  public — see "Legal rules"). When adding a card, match the numbers/text of the staple it
  follows; rename it. Never ship any third-party names/flavor.
  The current pool (`internal/cards/cards.go`) is an early slice, not a full roster: ~6 vanilla
  neutrals, ~6 battlecry/deathrattle, ~10 keyword minions, 2 weapons, 8 Mage spells, 4 secrets,
  2 Discover minions, Fireblast. Still missing most of the classic Mage toolkit (more burn/
  removal, tempo, board clears, card draw, signature spells) and most of the neutral "core" curve
  (bodies at every cost, vanilla stat-lines, simple keyword neutrals). Add as **data** in
  `cards.go` (the effect/keyword/trigger system already covers damage/heal/buff/summon/freeze/
  silence/aura/spell-damage/secrets/discover/lifesteal/poison/etc. — most new cards need no Go;
  add new Go only when a card needs an effect/keyword the engine lacks). Update `DefaultDeck()`/
  `DeckPoolIDs()` + `ValidateDeck` as the pool grows. Do this BEFORE Hunter so the neutral curve
  is solid before the class split.
- **Hunter — second class (TODO, not started).** Currently Mage-only with one shared
  card pool. Add a real class system: per-class card pools + per-class hero power, deck
  building scoped to a class, the engine choosing/serving the right hero power. Hunter is
  the planned second class — **Hero Power: Quick Shot** (2 mana, 2 damage to the enemy
  hero) — with beast/trap/reach cards (Beast tribe tag, trap-style secrets, a "draw a
  card" spell). Was previously sketched in README; design the class abstraction before
  adding cards. (Mage hero power is hardcoded `MageHeroPower()` — generalize first.)
- **Phase 10 reconnect: DONE**. Dropped player rejoins within a 60s grace window; opponent
  sees a banner; client auto-reconnects on transient drops.
- **Phase 10 client UI overhaul + iteration round 2 + turn timer + prod deploy: DONE** (see the
  two "Phase 10 — …" sections). HS-style table, deckbuilder book, lobby, 75s turn timer, `make
  prod` + nginx.
- **NEXT (own session) — ANIMATIONS.** The game resolves instantly and looks "unplayable"
  (user's word). The client already receives the ordered per-action event log (`State.Events`,
  kinds: attack/damage/heal/buff/summon/death/battlecry/deathrattle/freeze/shield/silence/
  secret/equip/weaponBreak/heropower/armor/mana/fatigue/burn). Build a **replay layer**: queue
  incoming snapshots+events, play events with pacing (attack lunge, damage flash + number popup,
  death fade, summon pop, heal/buff/freeze/shield effects), THEN settle to the snapshot state.
  Currently `App` applies each `state` immediately. Plan needed before building (how to buffer
  states without breaking reconnect/resync + the turn timer; whether to drive from `Events` with
  element lookup by `data-cid`, already present on minions/heroes). Keep the engine untouched.
- **Asset pass (lower priority): game-icons.net / Kenney** — replace emoji placeholders (card art,
  keyword badges, event-log icons, hero power) with CC-BY/CC0 SVG/pixel art + a pixel font.
  Track licenses in `ASSETS.md`. Not started.
- **Phase 10 next session — animation timing (Q4):** replay the per-action event log the
  client already receives, with pacing + highlight so actions don't look instant (and check
  the log only for detail). Hunt CC0/CC-BY assets (game-icons.net / Kenney) for card/board
  art. User explicitly flagged this as a separate session. Engine kept minimal so it layers on.
- **Phase 10 edge triggers: DEFERRED** — on_damage/on_turn_*/on_summon/on_spell_cast aren't
  wired; no current card needs them. Add as data when such cards arrive.
- Browser-verify Phase 10 reconnect: two tabs — reload one mid-turn (rejoins the game, not
  the lobby; other tab shows "Opponent disconnected…" then clears), kill a tab and wait
  ~60s (other tab wins), reconnect mid-Discover (prompt re-shown).
- Browser-verify Phase 9: two tabs at :5173 — deckbuilder (build a legal 30/30, copy
  cap, save/rename/delete, 10-deck cap), pick a deck in the lobby, queue → mulligan
  (toss cards, opponent learns nothing until both keep), per-turn draw, 2nd player's
  Coin (+1 mana this turn), deck count, fatigue on empty deck.
- Browser-verify Phase 8: two tabs at :5173 — Fireblast (hero power button, target,
  once/turn, kills a 1-hp minion, ignores spell damage); equip a weapon and swing
  face + into a minion (retaliation to your hero), weapon breaks at 0 durability,
  re-equip replaces; taunt blocks the hero going face; frozen hero can't swing.
- Browser-verify Phase 7: two tabs at :5173 — play a secret (opponent sees only a
  🔒 count, not which), trigger each (Snare destroys an attacker hitting your face,
  Mimic copies a minion the enemy plays, Nullify counters an enemy spell — all
  reveal + log), Discover (Arcane Insight / Wild Summons → modal of 3 filtered
  options, pick one → card to hand, other actions blocked until you pick),
  board-full summon discard.
- Browser-verify Phase 6: two tabs at :5173 — Gale Harrier (windfury double-swing),
  Bloodthorn Knight (lifesteal heals you), Toxic Fang (poisonous destroys + divine
  shield blocks it), Veil Stalker (stealth: untargetable, revealed on attack, AoE
  hits), Ember Scribe (spell damage on Cinder Bolt + per-target on Quake), Pack
  Leader (aura +1 atk to others), Hush (silence strips taunt/shield/buff/
  deathrattle), Drain Touch (spell lifesteal), new badges + log lines.
- Browser-verify Phase 3/4/5 (still unverified at handoff): spells, battlecries,
  deathrattles, charge/rush/taunt/divine-shield/freeze, event log, single-session
  kick, reload keeps session, disconnect→opponent wins.
- **Overload still DEFERRED** to whenever Shaman is added (Shaman-flavored;
  pointless while Mage-only) — revisit then.
- Only on_play/on_death + the three secret triggers are wired — add other event
  hooks (on_damage/on_turn_*/on_summon) as their cards arrive.
- Disconnect handling (Phase 10): a mid-match drop now starts a 60s reconnect grace window
  (the dropping player can rejoin their seat) and forfeits only if no one returns — no longer
  an immediate forfeit. A *finished* match object still lingers until both clients
  `enter_lobby` drop their refs (and the `active` seat map entry is cleared); then GC'd.
- Only on_play/on_death wired in the event bus — add on_damage/on_turn_*/on_summon/etc. as their cards arrive.
- `validTarget`/`ruleMatches` + battlecry-target + taunt/hero-reach/stealth logic duplicated Go+TS — revisit if it drifts.
- Aura is attack-only; +Health aura deferred (needs current-health delta tracking). Battlecry/deathrattle damage isn't lifesteal/poisonous-aware (combat + spell only).
- Public-push legal triad **done** (2026-06-21): `README.md` (description + IP
  disclaimer: not affiliated with Blizzard, original IP only), `LICENSE` (MIT,
  © amvid), `ASSETS.md` (empty table; no 3rd-party assets shipped yet). Code is
  MIT; no commercial intent (personal/hobby — owner + friends), so MIT is fine.

## Update protocol for future sessions
After meaningful change: update "Current state", "Build phases" progress, "Open/next",
and the "Last updated" date. This file is the memory between sessions.
