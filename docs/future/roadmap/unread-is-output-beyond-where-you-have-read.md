---
title: unread is output beyond where you have read
state: researching
created: 2026-11-19
tags: [story]
---

Output counts when it is recorded, and only reading moves the reading position. Mega stops working out which output is new.

Third in the sequence of `simplify-the-core`. Its shape depends on what is decided for `megas-history-may-have-a-gap`, which comes just before it.

## why

Today output is counted when its turn ends, with exceptions for output after its turn, for a runtime that exits mid-turn, and for imports by how they happened. A marker in the registry holds the result, and start-up rebuilds it from the log when the write was lost. About seven of the hardening findings were about what counts as unread after a lost runtime or an import, and one fix, the rebuild, failed in three passes running.

Decided by Michael on 2026-11-19, and revised the same day so that an import never moves the reading position: it could pass over a reply he had not seen.

## the failure that becomes impossible

An unread mark can no longer be lost because a write failed, or miscounted by a rebuild at start-up.

## what changes

- **Removed:** the buffering of a turn's output until its end, the counting of output flagged as after its turn, the scan a resume makes for what is new, and the detection of an open turn in Claude's and pi's histories where it served only that scan.
- **The newest output is taken from the log.** The registry's unread marker is neither read nor written. The field stays in the file until `the-registry-records-no-runtime-state` removes it.
- **One thing the marker's write did is kept:** the workspace notification, when recorded output changes a conversation's mark or its place in the needs-you list.
- **The canon's `bookkeeping-never-creates-attention`** is amended with it.

## what it costs

A conversation released, worked on in a terminal and resumed shows as unread until it is looked at, and so does a session just adopted. If mega's history is allowed a gap, this cost changes shape: there is less imported output to count.

## what does not change

The order of attention. A conversation with a turn in flight still reads as running. The flag for output after its turn stays for now, because it also stops late output from opening a turn; `a-turn-is-inferred-from-a-message` removes it.

## how it is shown

- an unclean restart of a scratch mega with a turn in flight, resumed, shows unread once;
- a conversation with an unread reply, released and resumed without being opened, is still on the needs-you list;
- a session just adopted shows as unread until it is opened;
- two unread conversations, the older then given more output, are in the order of each one's newest output;
- a conversation that is not the one on the page gains its unread mark on the rail from output alone.

## order

Depends on nothing in code. Lands on `main`. A build from before it, opened on the same workspace, counts unread its own way, so marks may differ after going back.
