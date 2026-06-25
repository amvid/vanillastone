# Vanillastone — project instructions

**Read [HANDOFF.md](HANDOFF.md) FIRST.** It holds all locked decisions, architecture,
mechanic taxonomy, build phases, and current state. It is the memory between sessions.

## Rules for this project
- **Keep HANDOFF.md updated.** After any meaningful change, update its "Current state",
  phase progress, "Open/next", and the "Last updated" date.
- **Server is authoritative.** Client never decides game truth. All rules, RNG, damage,
  draw resolve server-side.
- **Custom IP only.** No Blizzard art/names/sounds/flavor text — repo is public. Don't
  name it "Hearthstone". See HANDOFF "Legal rules".
- **Discuss scope before building game logic.** This project is being designed
  deliberately, phase by phase. Don't jump ahead phases.
- Match existing conventions. Surgical changes.

## Stack (locked)
Go authoritative server (WebSocket + JSON) · React + Vite web client · one Go binary
serves client (`go:embed`) + `/ws` · SQLite (decks) · in-memory match state · 1v1,
Mage only, no accounts, all cards unlocked, custom set.

## Dev
```
make dev      # dockerized server, hot reload (wgo)
make help     # all targets
```

## Current phase
Phases 1 (skeleton), 2 (minions + combat), 3 (spells + targeting) done. Next =
**Phase 4**: event bus + triggers (battlecry/deathrattle). See HANDOFF phases.
