---
title: central application
state: horizon
created: 2026-11-06
tags: [epic]
---

The central application on `apps.hq` at `roundhouse.hq.quigley.com`: local username/password login, per-environment enrollment and revocation, environment services dialing out over persistent authenticated connections, relay of commands and state, cross-project views assembled from several environments, and a last-seen cache that is never an authority. Brings `twentytwo` in as the second environment. Transport choice (plain outbound WebSocket through nginx or a private zrok share) is decided here.

## background

Noted by the planning agent on 2026-11-14. This card was written before the rename and before `twentytwo` ran a mega of its own. Two things have moved since. The hostname is no longer settled, since the project is now mega. And `twentytwo` is not waiting to be brought in: it runs its own independent workspace, so this card's first job is joining two workspaces that already exist, which is also what `moving-conversations-between-hosts` wants in a smaller form.

## the second direction, 2026-11-18

From Michael's presentation: "mega grows multiple environment support. today mega is a single service that contains the harness orchestration and the user interface. future mega separates these into any number of harness orchestration environments, reachable anywhere, consolidated into a single pane of glass."

That is this card, in his present words, and it sharpens two things the card of 2026-11-06 left loose:

- **The split is between orchestration and interface.** An environment runs harnesses; the pane of glass shows many environments. Today's mega is both in one process. `the-orchestration-framework-is-extracted` is what makes the first half something that can stand alone.
- **Reachable anywhere means the overlay.** He already reaches a mega from another machine through a private share, and each page already says which mega it is. Those are the single-environment version of this.

What the first card got right and still holds: environments dial out, nothing is published, a login stands in front, and a cached last-seen state is never an authority.

