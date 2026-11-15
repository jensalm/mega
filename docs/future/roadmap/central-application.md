---
title: central application
state: horizon
created: 2026-11-06
tags: [epic]
---

The central application on `apps.hq` at `roundhouse.hq.quigley.com`: local username/password login, per-environment enrollment and revocation, environment services dialing out over persistent authenticated connections, relay of commands and state, cross-project views assembled from several environments, and a last-seen cache that is never an authority. Brings `twentytwo` in as the second environment. Transport choice (plain outbound WebSocket through nginx or a private zrok share) is decided here.

## background

Noted by the planning agent on 2026-11-14. This card was written before the rename and before `twentytwo` ran a mega of its own. Two things have moved since. The hostname is no longer settled, since the project is now mega. And `twentytwo` is not waiting to be brought in: it runs its own independent workspace, so this card's first job is joining two workspaces that already exist, which is also what `moving-conversations-between-hosts` wants in a smaller form.
