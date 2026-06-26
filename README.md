# Vanillastone

A small, server-authoritative digital collectible card game — 1v1, fast,
playable in the browser with a friend. It is a hobby project built to explore
the mechanics of the genre (minions, spells, battlecries, deathrattles,
keywords) with a clean, authoritative game engine.

**Source:** https://github.com/amvid/vanillastone

> **Original work.** Every card, name, and asset in this project is original or
> uses freely-licensed (CC0 / CC-BY) assets. See **Legal & IP** below.

## What it is

- **1v1** matches, Mage hero only (for now), all cards unlocked.
- **Server-authoritative**: the Go server owns all game truth — rules, RNG,
  damage, draw, and effect resolution all happen server-side. The client is a
  dumb renderer; it never decides game state.
- **Custom card set**: a small set of original minions, spells, secrets, and
  weapons, growing as the engine grows.
- **Deckbuilding**: build and save your own 30-card decks (up to 10 per account),
  with the usual shuffle, opening-hand mulligan, per-turn draw, and fatigue.
- **No accounts beyond a username + password.** No email, no resets, no roles.
- **Animated board**: cards fly to the table, attacks lunge, hero powers fling,
  damage/heal/fatigue pop — driven by the server's ordered event log.
- **Live lobby**: see who's online and who's currently fighting; reconnect after a
  drop within a grace window; a per-turn timer keeps games moving.

## Status

Built deliberately, one phase at a time:

| Phase | Scope                                                                                   | State |
| ----: | --------------------------------------------------------------------------------------- | :---: |
|     1 | Skeleton loop — connect, lobby, matchmaking, turn ping-pong                             |  ✅   |
|     2 | Minions + combat (mana, summon sickness, hero HP, win/lose)                             |  ✅   |
|     3 | Spells + targeting                                                                      |  ✅   |
|     4 | Event bus + triggers (battlecry, deathrattle), event log                                |  ✅   |
|     5 | Keywords wave 1 — taunt, charge, rush, divine shield, freeze                            |  ✅   |
|     6 | Keywords wave 2 — windfury, stealth, poisonous, lifesteal, spell damage, aura, silence  |  ✅   |
|     7 | Secrets + Discover (hidden state, mid-action prompts)                                   |  ✅   |
|     8 | Hero power + weapons                                                                    |  ✅   |
|     9 | Decks + SQLite (deckbuilder, mulligan, fatigue)                                         |  ✅   |
|    10 | Polish — UI overhaul, reconnect, turn timer, animations, live player list               |  ✅   |
|    11 | Spectator mode — watch a live match from a player's point of view                       |  ✅   |
|    12 | Card set — initial core neutrals + Mage spells/secrets/weapons/hero power               |  ✅   |
|    13 | **Expand the card set — fuller Mage toolkit + deeper neutral curve (original designs)** |  ✅   |
|    14 | Play versus AI                                                                          |  ✅   |
|    15 | Hunter — second class + Hunter cards                                                    |  ✅   |
|    16 | Ladder                                                                                  |  ✅   |
|    17 | Warlock — third class + Warlock cards                                                   |  ⏳   |
|    18 | Warrior class + Warrior cards                                                           |  ⏳   |
|   ??? | Art                                                                                     |  ⏳   |

## Stack

- **Server**: Go, authoritative. WebSocket + JSON transport.
- **Client**: React + Vite + TypeScript.
- **Storage**: SQLite (pure-Go `modernc.org/sqlite`) for accounts and saved decks;
  match state is in-memory.
- **Deploy**: a single Go binary serves the embedded client (`go:embed`) and the
  `/ws` endpoint.

## Develop

Requires Docker.

```sh
make dev      # server :8080 + Vite dev server :5173 (hot reload)
make help     # all targets
```

Open **http://localhost:5173** (Vite, with HMR) during development. Port 8080
serves the embedded production build.

To play locally: register two accounts (two browser tabs / windows), log in on
each, and hit **Play** to be matched against each other.

## Legal & IP

This is a **non-commercial, fan-made project** inspired by the digital
collectible card game genre. It is **not affiliated with, endorsed by, or
associated with Blizzard Entertainment**, and it is **not** Hearthstone®.
Hearthstone is a trademark of Blizzard Entertainment, Inc.

Game _mechanics_ are not copyrightable; an engine that implements card-game rules
is fine to build and share. _Names, art, sounds, and flavor text_ are protected,
so this project uses **none** of Blizzard's:

- No Blizzard art, audio, fonts, card/hero names, or flavor text.
- No datamined or extracted game files.
- All card names and art are original; any third-party assets are CC0 / CC-BY and
  tracked in [ASSETS.md](ASSETS.md).

### Card art (AI-generated)

The card art is **generated by an AI image tool from original prompts** — generic
fantasy archetypes (a fire elemental, a frost mage, a murloc-like fish-folk), never
a specific existing character. We deliberately keep every image visually distinct
from any recognizable copyrighted character or logo; prompts that drift toward an
iconic design are reworked. Provenance and the tool's terms are tracked in
[ASSETS.md](ASSETS.md).

Note: purely AI-generated images may not be eligible for copyright, so we make no
ownership claim over the art — treat it as provided "as is" for use within this
project. If you spot any image that resembles an existing character or asset, open
an issue and we'll replace it.

### Self-hosting

You're welcome to run your own server to play with friends. By doing so you agree to
keep it **non-commercial** (no ads, sales, or paid access), not to re-introduce any
real card/hero names, art, or flavor text, and to host at your own risk.

Code is released under the [MIT License](LICENSE). The MIT license covers the
**code only**, not the card art (see above).
