---
title: the registry records no runtime state
state: researching
created: 2026-11-19
tags: [story]
---

The registry loses the two fields that recorded what happened to a conversation: its runtime state and its unread marker. The log already says both. Everything else in the registry stays where it is.

Last in the sequence of `simplify-the-core`, on the branch, after `one-fold-for-what-happened` has soaked.

## why

A change to a conversation's runtime state is recorded in its log and then written to the registry, and each place that does so needs its own answer to what happens when the second write fails. With the fold reading the runtime state from the log, the registry's copy is only a second thing to keep in step.

Decided by Michael on 2026-11-19, and narrowed the same day. An earlier draft moved the session id and the operator's choices out of the registry too. Old logs never recorded those, so that needed a seed and rules for it. They stay in the registry, and the spec's promise holds: the registry names every conversation's session and can be read with the service stopped.

## the failure that becomes impossible

The registry and a log can no longer disagree about whether a runtime is running.

## what changes

The registry goes to version 2 without the two fields. A start on version 1 cuts over, in this order:

1. For each conversation the registry holds as running or unsupervised, where the log does not already say so, an `unsupervised` event is recorded. This carries the fold's "less certain stands" across the cutover.
2. The version 1 registry is copied beside itself as `registry.v1.yaml`, and the copy and the directory that names it are synced, the way the registry's own writes are. An existing copy is replaced.
3. The registry is replaced, atomically, by one at version 2. This is the commit.

An interruption before step 3 has changed nothing that counts, and the next start does the three steps again. The loader already refuses a version it does not know, so a build from before cannot open a workspace after the cutover. Mega never removes `registry.v1.yaml`, and the start that cuts over logs the copy's path.

## going back

Kept simple on purpose. Michael, 2026-11-19: "we don't need to over-engineer the rollback part of migration. i would assume that worst case scenario an agent could be removed from the workspace and re-adopted fresh and the conversation would remain intact." No log changes form, so there are two ways back and neither is tested as a gate. The first: stop mega, put `registry.v1.yaml` back as `registry.yaml`, start the older build. The second, for a conversation that does not come back right: remove it from the workspace and adopt its session again. Its history is the harness's own and comes back whole; what only mega's log held does not.

## how it is shown

- the cutover on a copy of Michael's workspace, interrupted before, between and after each of its three steps, and rerun;
- the same on a copy that never ran the branch, as an external operator's will not have;
- a build from before refuses the workspace afterwards.

## order

After the fold has run on Michael's machines for a week with nothing unexplained. Then the branch is merged, which is a gate of its own and Michael's: not before the cutover has run on each of his three machines and he has worked on each afterwards. At least two external operators have workspaces this will change.
