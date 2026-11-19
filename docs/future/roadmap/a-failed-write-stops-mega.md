---
title: a failed write stops mega
state: researching
created: 2026-11-19
tags: [story]
---

On the first failed write to the record, mega logs why and exits at once, the way a crash would leave things. The record is the registry and the event logs.

Fifth in the sequence of `simplify-the-core`.

## why

Today a failed write leaves the service running on its last confirmed view. The registry and each event log have an uncertain state of their own that refuses later changes (`registry.ErrUncertain`, `events.ErrUncertain`), and each operation has its own rule for what the page may be told when its record fails. About eleven of the hardening findings were a write that failed partway: an answer hiding a request before it was recorded, a settlement that could not be recorded, a mark for a request that "takes no further answer".

An unclean restart is a path that has to exist and work anyway. This makes it the only one.

Decided by Michael on 2026-11-19. Two things were in earlier drafts and were taken out, and should stay out: ending the runtimes on the way out, and telling the page why. Each needed rules of its own in every review round it survived.

## the failure that becomes impossible

Mega can no longer run on while unable to record.

## what changes

- **A failed write to the registry or an event log ends the process.** It logs the reason first. It does not take the orderly stop that `mega serve` and `mega daemon` take today, which releases every runtime.
- **Removed:** both uncertain states, every rule for a record that failed after another was made, and the mark for a settlement that could not be recorded.

## what it costs

Whatever a harness was doing goes on unwatched until its own pipes close or its turn ends, exactly as after a crash. A command already under way may finish what it was sending. The page shows the service as unreachable, and the reason is in mega's log.

It is the right trade for one operator on his own machine, where a full disk is rare and a restart is cheap.

## what does not change

A failure to write anything else under the data directory: a snapshot is best effort, and mega's own log file and the tray's runtime file are diagnostics. What the next start does: it is the start after a crash.

## how it is shown

- an injected write failure, and separately an injected sync failure, each end the process without the orderly stop: a conversation at rest with a runtime is sent nothing;
- the next start shows a conversation that was running as unsupervised and a pending request as lost;
- a failed snapshot stops nothing.

## order

Depends on nothing. Lands on `main`.
