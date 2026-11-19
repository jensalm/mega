---
title: reading the workspace starts nothing
state: researching
created: 2026-11-19
tags: [story]
---

Reading the workspace reports what mega last observed and starts no work. An environment is resolved when something needs it or when the operator asks, and never because a page or the tray looked.

Sixth in the sequence of `simplify-the-core`. Michael, 2026-11-19: "that sounds like a clearer design."

## why

From an independent review for simplification, by an Astra 6.2 agent Michael asked on 2026-11-19, read against `main` at `c19f72a`. The planning agent checked its claim in the code and it holds.

`Service.Workspace()` ends by calling `resolveStaleLocked`, which starts a background resolution for every project whose environment may run and whose latest resolution is missing or out of date. A resolution runs the project's approved activation in a login shell. The tray calls `Workspace()` each time the service announces a change, and a finished resolution announces a change. There are guards and a limit on how many run at once, so it does not loop for ever. But the reviewer's sentence is true as it stands: "updating the tray icon can cause a project's approved activation to run."

Only an approved activation ever runs, so no promise is broken. What is wrong is the shape: a read that has effects, background jobs a read can start, and a shutdown that has to wait for them.

## two sizes

- **The small one.** Take the call out of `Workspace()`. Resolution still happens at start-up, when a project is registered, when the operator approves or changes an activation, when a conversation starts, and on Check. A project whose file changed on disk shows as changed and not checked until one of those.
- **The reviewer's.** Also drop the resolution of every project at start-up. The workspace reports the last observation, with "not checked yet" as an ordinary answer, and says when it was checked. Approval and freshness are still checked where mega actually runs something. That removes the start-up jobs and their coordination with shutdown as well. The reviewer's point for an external operator: "checked at 10:42" is a clearer claim than an indicator whose freshness depends on which background job has run.

## the failure that becomes impossible

A read of the workspace can no longer run anything.

## order

Depends on nothing. It does not touch the fold or the registry. Lands on `main`. To settle when it is specified: which of the two sizes.
