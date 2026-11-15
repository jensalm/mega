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
