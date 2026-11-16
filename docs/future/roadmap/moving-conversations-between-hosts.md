---
title: moving conversations between hosts
state: horizon
created: 2026-11-14
tags: [feature, spike]
---

it would be useful if mega had the ability to extract conversation details from one host and move it to another host.

## background

Noted by the planning agent on 2026-11-14, with mega now on two hosts. A conversation is three things in three places: mega's registry entry, mega's event log, and the harness's own session in its native store. Claude keys its store by the checkout's path, so a session moved to a host where the checkout lives elsewhere is not found; the capability report's follow-up of 2026-11-11 recorded exactly that for a moved checkout. `relocate-a-projects-checkout` is the same problem on one host and should be worked first or together, since the spike is shared: what each harness needs rewritten for a session to resume under another path.

Further things to settle when this is shaped: whether the move is a file the operator carries or something two megas hand each other, which is `central-application` territory; what happens to the conversation on the host it left, since two writers on one session is the hazard every harness has; and that a session's transcript is private, so an export is handled as carefully as the store it came from.

Later the same day this gained a foundation: the conversation bundle described in the epic `a-workspace-that-can-be-rebuilt-elsewhere`. A move between hosts is one bundle exported from one mega and imported by another. This card is that epic's last layer.

## the first real case, 2026-11-16

Michael wants to install mega on his laptop for the demo and move mega's own development there, while the mega on `fortyfive` stays. That is this card's first real case, a week before anything is built for it. What moving one project's work actually consists of, read on `fortyfive` by name and size only:

- **The conversations.** Eight Claude sessions for this checkout, 180 MB in Claude's store, keyed by the checkout's path. Plus mega's registry entries and event logs for them, and the Codex threads.
- **The agent's memory.** Eight files beside the sessions, outside the repository. A planning conversation started fresh elsewhere arrives without them.
- **What the agents need around them.** The MCP servers the harnesses start (the grimoire, mercurius, terminus), the canon repository beside the checkout, the fixtures with their trust accepted, the project's environment file and its approval, the toolchain the activation expects, and the harness logins.
- **What cannot move.** Logins, trust and approvals are given again. The local model is reachable only where the network allows.

Two routes exist today without this card. Leave the work where it is and reach it through a share, which already works. Or copy a session's store to the same path on the other machine and adopt it there, since adoption imports a native session's history; that needs the conversation stopped where it left, by hand, because nothing stops two machines writing one session.

Michael then said the move would be temporary, to another Ubuntu machine, with a VPN back to the network `fortyfive` is on. That sharpens what this card is for. A temporary move does not want the conversations moved at all, since they would have to come back: two migrations, and twice the chance of two machines writing one session. What it wants is the work to continue elsewhere and return. The repository already carries that, by design: the journal, the cards, the canon and `AGENTS.md` are what a conversation arriving cold reads. So there are three different things an operator may mean by "move", and this card is only the third:

1. Reach the same mega from another machine. Works today, through a share or a tunnel.
2. Continue the work on another machine with fresh conversations, and bring it back through git. Works today; what does not travel is the agent's memory folder, which sits outside the repository.
3. Carry a conversation itself, with its history, to another machine. This card.
