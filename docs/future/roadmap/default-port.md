---
title: default port
state: evaluating
created: 2026-11-18
tags: [enhancement]
---

the default port collides with NetFoundry's Agora, so we should pick a different default.

## background

Noted by the planning agent on 2026-11-18. The default is `127.0.0.1:8470`, set in `internal/config/config.go` and written down in about a dozen places: the README, `docs/current/mega.md`, `AGENTS.md`, the tests of the config, the server and the registry lock, the version probe, and the UI's development proxy. The checks and scratch runs use 18470 and nearby, which were chosen to sit beside it.

**What a change touches beyond the code:**

- **Everyone already running mega.** An operator with no `listen` in their config moves to the new port at their next start. Their bookmarks, and any share or tunnel pointed at 8470, stop reaching it. The changelog entry has to say so first.
- **Michael's own machines.** Three megas, a zrok share and the tunnels he reaches them through all name 8470.
- **Browser preferences.** A browser keeps a page's settings by address and port, so panel widths and folds start fresh once, as they did at the rename.
- **A second mega finding the first.** Starting mega while one runs opens the running one, which is found by the port. A new binary beside an old one on the old port would not see it.

**Picking the number.** It should be unregistered, not a common default of anything a developer runs, and easy to remember. One candidate: 6342, which spells MEGA on a telephone keypad. It is not in this machine's services list; it has not been checked against IANA's registry from here. 8472 is to be avoided, since Linux uses it for VXLAN.

**To settle.** The number, and whether the scratch ports the checks use move with it.

## checked against IANA's registry, 2026-11-18

- **8470 is registered,** to Cisco's Address Validation Protocol on TCP, with UDP reserved. So the present default collides with more than Agora.
- **6342 is unassigned.** It is the last port of an unassigned range, 6327 to 6342. Its neighbour 6343 is sFlow.

The prompt for this card was written with 6342, the planning agent's recommendation, and with the scratch ports left where they are: they collide with nothing, and moving them is a sweep through every suite for no gain. Both are Michael's to change.
