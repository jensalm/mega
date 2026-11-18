---
title: the orchestration framework is extracted
state: horizon
created: 2026-11-18
tags: [epic, spike]
---

The part of mega that drives harnesses, and the part of mercurius and terminus that does the same, become one framework that stands apart from any of the three. It is the foundation for harness orchestration that is native to the overlay.

## why

From Michael's presentation of mega, 2026-11-18, the first of seven future directions: "the harness orchestration framework in mega and mercurius/terminus is extracted. this becomes the foundation for an overlay-native harness orchestration solution."

Three tools now start, prompt, watch and stop coding harnesses, each in its own way. Mega has adapters for Claude Code, Codex and pi behind one interface, with a service that owns their lifecycle, their requests and their logs. Mercurius and terminus each run a reviewer of their own. What mega learned the hard way, exact-id resume, a pending request never answered for the operator, a release that reports what it could not stop, version tolerance, is exactly what the other two need and do not share.

## what is already shaped like a framework

- The adapter interface in `internal/harness`, with three implementations and a fake for each.
- The capability report and its rerunnable probes, which are the evidence any adapter is written from.
- The canon's qualities for this work: the protocol confined to adapters, honest runtime state, permissions never weakened.
- The service's rules for one authority and durable records.

## what the extraction has to settle

- **Where the line falls.** Adapters alone, or adapters with the lifecycle, the request model and the event log. The second is what makes a second consumer cheap, and is much more to pull out.
- **What "overlay-native" asks of it.** An environment that runs harnesses and is reached over the overlay, not over a local port. That is the next direction, `central-application`, and this card is what it stands on.
- **The Claude bridge.** One adapter runs through a Node program beside the binary. A framework others build on has to carry that, which `first-tagged-release` meets first.
- **Mercurius and terminus as its first other users.** Their reviewers become harness runs like any other, restartable and observable, which is also what the hub epic asks of them.

## depends on

Nothing built yet. It is the first of the directions because the next two rest on it.
