---
title: late codex output during a newer turn
state: horizon
created: 2026-11-18
tags: [defect]
---

A Codex background command can finish after its turn has ended. If a newer turn is running by then, its output is marked unread only when that newer turn ends, not at once.

## background

A terminus review on 2026-11-18 (`47302dd02565`) found this while reviewing the reply box batch, which does not touch the code. In `internal/service/live.go`, at `sink`'s unread accounting, output the adapter marks `AfterTurn` is marked unread at once only while no turn is active. With a newer turn active, it falls through to the active-turn branch and becomes that turn's `turnOutput`. So it reads as unread only when the newer turn ends, and a failed background job stays out of needs-you until then.

Whether that is wrong is a decision. While a turn runs, the conversation is working and the operator is likely watching it, and the newer turn's end covers the late output anyway. The review's suggestion is to handle `AfterTurn` output before the active-turn branch, whatever turn is active.

## the planning agent's reading, 2026-11-18

This changes nothing the operator can see, so it is not a defect. Attention is one value, decided in order: a pending request, an unsupervised runtime, a turn in flight, an error, unread output, idle. A conversation with a turn in flight reads as running whether or not it holds unread output, so it is not in needs-you as unread during that turn either way. Marking the late output unread at once would move a counter and nothing on the page. When the newer turn ends, the late output is unread with the rest, which is the moment it could first show.

A failed background job is the case the review worried about. Its result is in the history as an action marked failed the moment it arrives, and the conversation's mark is the running one. Nothing is hidden; nothing is raised early, because nothing can be while a turn runs.

The canon's quality `bookkeeping-never-creates-attention` gained a line saying so, in its list of what not to flag, so a review does not block on this again. If Michael reads it differently, the fix is the review's: handle output marked as after its turn before the branch for an active turn.
