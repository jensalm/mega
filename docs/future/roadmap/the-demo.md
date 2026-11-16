---
title: the demo
state: researching
created: 2026-11-13
tags: [story]
---

A live demonstration of mega to its first likely external operators: a workspace made for showing, and a short list of what to show.

## why

Michael demonstrates mega in under a week from 2026-11-13, to the people most likely to become its first external operators, as part of presenting his whole toolset. Work on mega continues until then.

## shape

- **No freeze.** Michael, 2026-11-14: the demo is informal, and changes can land right up to it, as long as mega keeps being used and tested daily until then. The guard is that use, plus `make test` and `make checks` passing on whatever build is shown.
- **A demo workspace.** Its own data directory and port, so nothing private is on screen: two or three small public repositories as projects, conversations with both harnesses, and pi if its adapter has landed.
- **What to show, in order.** Left to Michael; a starting list is the side panel with conversations in different attention states, a prompt with path search, an approval answered from the dock, a mode switch, the turn line, an action group opened, a second project with its own environment, and shutting down when idle.
- **A walk through the list** on the build that will be shown, on the machine and the screen that will be used, whenever suits.
- **A fallback.** Screenshots of each step, in case the network or a harness login fails in the room.

Harness logins are live credentials. The demo workspace uses Michael's own logins and shows no settings pages, tokens or environment values.

## additions to the list, 2026-11-16

Built since this was written and worth showing: the first page saying what mega needs and found, a pi conversation on the local model if its adapter has landed, and `mega snapshots`.

## the plan, 2026-11-16

Michael's decision: the demo runs on his laptop, another Ubuntu 24 machine, with its own mega. Mega's own development stays on `fortyfive` and is reached from the laptop when wanted. On the laptop he works on a handful of other projects for the week, mercurius and terminus among them, with fresh conversations there, and that work comes back through git. Nothing is moved between machines.

Why local: reaching `fortyfive` remotely works but is slower than a local page, and in a demo that slowness would be taken for mega's.

What the laptop needs before the room:

- **Mega built from the README,** which is also the README's first real test on a second machine. Anything that had to be guessed goes on `readme-for-newcomers` or `clean-machine-rehearsal`.
- **Each harness installed and logged in,** and each project's checkout trusted in Claude's own terminal.
- **The projects registered, and their environments approved,** with whatever their activation expects installed.
- **The tools those projects' agents call:** the grimoire, mercurius and terminus servers, and the canon checkout beside them.
- **One conversation per harness already launched and answered once,** so the first launch, the model list and any first-run prompt have happened before anyone is watching.
- **The local model reachable** over the VPN, checked from the room's network, if pi is to be shown.
- **A tab on `fortyfive`'s mega,** if mega working on itself is to be shown. The name in the tab title tells the two apart.

What to keep off the screen is unchanged: settings pages, tokens and environment values.
