---
title: mega's history may have a gap
state: researching
created: 2026-11-19
tags: [story, spike]
---

A product question, for Michael to decide: after work done outside mega, does its own timeline have to hold those messages, or is it enough to resume the exact session and be able to look at what happened?

Second in the sequence of `simplify-the-core`, as a short design to put to him before anything is built. It comes this early because its answer changes the shape of `unread-is-output-beyond-where-you-have-read` and the pi side of `one-kit-under-the-adapters`.

## where Michael stands, 2026-11-19

"so far, i rarely take harnesses out of mega. so this doesn't bite. and even if i did, in practice i'm always looking at the last 10 turns anyway... so the complete history isn't super relevant."

That leans toward the gap. What the design still has to say: what the gap looks like in a conversation, how the harness's own transcript is opened from it, what an adoption still reads, and what a conversation shows after lost supervision.

## the failure that becomes impossible

A message can no longer be shown twice because a later import did not recognise it, and a turn can no longer be reported as failed because its message could not be matched to pi's file.

## why

From the same independent review as `reading-the-workspace-starts-nothing`, 2026-11-19, which called it "the larger product cut" and asked that it be kept as a separate decision. The planning agent read the pi code it points at and has not checked every claim.

Today mega promises a seamless timeline. After a release and a terminal session, after lost supervision, mega reads the harness's own history and brings in what it did not see, and it must not show a message twice. pi shows what that costs. A message pi reports live does not carry the id it will have in pi's session file. So the adapter asks pi for its entries, matches each live message by role, time and call, moves a cursor, retries, falls back to the file, and holds later messages behind one it could not match (`internal/harness/pi/runtime.go`, `entryID`). If the last match fails, a turn can be reported as an error though the model did its work. All of it serves a later import that might otherwise show the message again.

## the alternative

Mega records what it witnessed, marks the interval it did not, and offers the harness's own transcript for that interval when asked. An adoption still takes one snapshot of history, since there is nothing earlier of mega's to reconcile it with. Resuming by exact session id stays, and so does the deliberate recovery after lost supervision.

What goes: the duty to match every live message to a future import, the lists of native ids kept to stop a second copy, and the reads of history a resume makes.

## what it costs

The timeline in mega is no longer continuous across work done elsewhere, and unread is no longer counted message by message for that work. A conversation would say "this much happened outside mega" and not "these four replies are new".

It bears on the spec's terminal handoff, which today ends with the conversation whole again in mega. It does not touch the rule that nothing replaces native continuity with a summary: the session is still the session.

## how it meets the simpler core

`unread-is-output-beyond-where-you-have-read` already stops working out which imported output is new. This would go further and stop importing: that card removes the scan a resume makes, and this would remove the read behind it.
