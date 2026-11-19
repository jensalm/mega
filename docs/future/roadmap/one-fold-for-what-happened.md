---
title: one fold for what happened
state: researching
created: 2026-11-19
tags: [story]
---

What the service holds about a conversation because of what happened to it is built by one function with no I/O, `apply(state, event)`, used both while it runs and at start-up. What the service decided and what a restart recovers cannot disagree, because they are one piece of code.

Tenth in the sequence of `simplify-the-core`, and the first of the two that are built on the branch.

## why

Some of a conversation's state is decided by what happened to it: its runtime state, whether a fresh session may start, the requests not yet resolved, its newest output, whether a turn is open, its errors, and when it was last instructed. Today parts are in memory, two parts are in the registry, and parts are worked out at start-up by code that is separate from the code that decided them. One such piece of working-out, the rebuild of the unread marker, failed in three hardening passes running, each time by reading another event wrongly.

## the failure that becomes impossible

A restart can no longer recover something different from what the running service had decided.

## what changes

The service changes that state in one way: it appends an event to the log, and when the append is confirmed it applies the event. Start-up applies the same function to the same events. Each piece of start-up code that works one of these facts out separately is removed.

What each field means, and what sets and clears it, is what the code does today. The fold changes four things:

- **A decision is recorded, never inferred later.** Where the service decides something that no event in the log makes plain, the decision gets an event, or a field on one. Who asked for a stop is one. Whether an error ended its runtime is another. The build lists each one it had to add.
- **For a conversation that holds a session, `starting` means possibly running,** until a `not_started`, a stop or an exit is recorded. A crash after `starting` shows it as unsupervised at the next start, which protects a known session from a second runtime. A conversation that holds no session stays in today's session-unknown case: prompts are refused, the page points at "Adopt a session", and the row can be removed. `not_started` keeps the meaning the code gives it (`harness.ErrNotStarted`).
- **Nothing in progress survives a restart.** A log can stop in the middle of a turn, a resume, an adoption's read of history, or a stop. At start-up none of those is under way. An open turn is closed by `unsupervised`, and the passing states mega shows while it works, such as `resuming` and `stopping`, belong to the operation in memory and are not brought back from the log.
- **Where a log does not prove that nothing is running, the conversation is unsupervised.** Logs written before this are read by the same function, and it does not guess.

With it comes one small decision of Michael's, 2026-11-19: **a stop mega forced does not clear an error.** A stop the operator asked for still does.

## what does not change

No new kind of line in any log, and nothing added to a log written before. The registry: its runtime state is still written in this step, so the two can be compared. What a conversation shows that is observed and never recorded stays in memory beside the fold.

## the comparison

At each start the fold's runtime state is set against the registry's and a difference is logged with both values. The less certain of the two stands: if either says a runtime may be alive, the conversation is unsupervised. Some differences are right, such as a crash just after `starting`, and the build's report names each kind it meets.

This is the one place the work can repeat the mistake it is meant to end, working out from other events what the service had decided. So it soaks on Michael's machines, with the registry as the cross-check, before the registry gives that field up.

## how it is shown

- on copies of Michael's real workspace, a one-off tool builds every conversation's state with the fold and with the build from just before it, each from its own untouched copy, and every difference is explained;
- a log that stops in the middle of a turn, replayed, shows its conversation as unsupervised with no turn in flight, and does not hold a shutdown;
- a log whose last event is an adoption's `resuming`, replayed, shows nothing under way and does not hold a shutdown;
- a fresh launch the harness refused with `not_started`, replayed, allows a fresh session on the retry;
- a crash just after a brand-new conversation's first `starting` leaves a row that can be removed;
- a read of history that fails before a launch, then a resume that succeeds, leaves the same error showing as the build from just before does;
- after a week of Michael's own use on at least one machine, the differences logged at start-up are read and each is explained.

## order

Easier after `unread-is-output-beyond-where-you-have-read` and `a-turn-is-inferred-from-a-message`, which leave it less to fold. Built on the branch `simpler-core`. Nothing in the registry or the logs changes form, so a machine can go between the branch and `main`.
