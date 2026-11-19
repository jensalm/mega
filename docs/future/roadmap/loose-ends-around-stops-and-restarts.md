---
title: loose ends around stops and restarts
state: evaluating
created: 2026-11-19
tags: [defect]
---

Eight narrow faults and one small follow-up, most of them in what mega does when a runtime stops, the service restarts or a connection returns. Each needs something to go wrong first, and none has been seen in use. They were found by review, and they are worth fixing because each is a place where mega's own rules are not kept.

## why

A terminus review of the last three batches' files on 2026-11-19 (`915e203707a7`) confirmed the five fixes of `the-page-says-more-than-it-has-seen` and raised five more blocking findings, all in older code. This was the third pass in a row to end that way. Michael's direction on 2026-11-18 was to clear the review before returning to pull request 3; the planning agent's recommendation on reading this pass was to card what is left and return to the pull request, which is the more urgent.

## the eight

1. **A failed runtime's output can be counted late.** When a runtime exits in the middle of a turn, output it had produced in that turn is kept aside and never counted, and a later turn's end can then count it as unread. Recommended: count it at the exit. It is output the agent produced, the conversation reads as an error at that moment either way, and nothing is left over to surface later.
2. **Old imports can be counted unread after an unclean restart.** Resuming scans the whole log, so output brought in by an earlier adoption or import can show as unread. The rule is that imported output after a release or an adoption stays read. Only what this resume brings in should count.
3. **An answer can hide a request.** Answering removes the request from the pending set before the answer has been recorded. If the record fails, the request is hidden and still pending in the harness, and the agent waits on something the operator can no longer see. A request should leave the pending set only once its answer is recorded.
4. **A Codex answer can start work after a stop.** Between checking for an interrupt and starting the turn there is a gap, so an answer can start a turn after the interrupt has landed. And an answer is still accepted once a release has begun. An answer should be refused once an interrupt or a release is under way, with no gap between the check and the start.
5. **A reconnect can still show "Live" a moment early.** The page calls itself caught up when a workspace read lands, even if a change notice during that read has already queued another. It is caught up only when no read is waiting.
6. **The shutdown banner can report an old shutdown as under way.** After mega is shut down from a page and started again, the page's reconnect makes the banner say "Shutting down" and that every runtime is being released, before the restarted service has confirmed its workspace. The footer beside it correctly says "Catching up". The banner should wait in the same way.

7. **Claude's history does not say that a turn is still open.** Reading a session back, the adapter never marks a turn as unfinished. After an unclean restart, the harness process mega lost may still be working. A resume then imports what that turn has produced so far and counts it as a finished turn's output. Found by the review of the pull request 3 follow-up on 2026-11-19, raised twice, in older lines.
8. **The tested range claims too much.** Its definition, in `internal/harness/versions.go`, the glossary and `docs/current/mega.md`, says the builds "whose every call mega makes was seen to work". The probe covers a fixed list of steps, which until 2026-11-19 did not include a model switch on an idle runtime, and that was broken inside the range the whole time. It should say what the probe ran.

9. **The pi test's hosted model.** Not a fault, and it rides with this batch. The idle model switch needs a second model, so that one pi round trip leaves Michael's local model for a hosted one. He approved it on 2026-11-19 and asked for `gpt-6-sol`, which pi lists as `openai-codex/gpt-6-sol`. The test's built-in default, and the documents that name it, still say `gpt-6-luna`.

## decisions inside it

The first and the fourth each hold one: whether a failed runtime's output is counted or cleared, and whether an answer is refused once a release has begun. The recommendations above are the planning agent's. On the first, the review's own suggestion was to clear the output; counting it is recommended here because a turn that ends by its runtime exiting has still ended, and its output is what the operator will want to see when they open the error.

## the fourth, sharpened on 2026-11-19

The implementer found that the fourth item, read literally, breaks pi. pi holds an interrupt until its pending dialog is dismissed, and Dismiss is an answer. Refusing every answer once an interrupt has begun would leave a pi interrupt waiting for ever, with Release as the only way out.

The rule that keeps the intent: once an interrupt or a release has begun, refuse what lets work go on, and accept what settles toward the stop.

- **Refused:** Allow, Allow for the session, and a chosen or written answer to a question. A pi dialog is always a question and mega cannot tell which of its choices would let work run, so every chosen answer to one is refused.
- **Accepted:** Deny, and pi's Dismiss.
- **The refusal says why,** that the turn is stopping, and names what can still be pressed. The request is still pending, so the refusal is a conflict, not "no longer pending".

## a note on the review

The first five came from reviews by path of whole files, which the prompt for the previous batch asked for; that was the planning agent's instruction, not a limit of the tool. On 2026-11-19 the planning agent ran the review the batch is properly held to, the working tree's changes (`6082dde05ee4`). It was not clean either: it raised the same five and the sixth, and nothing against the five fixes themselves. So these are not an accident of reviewing too widely. A review of a change reads the files it touches, and these sit in those files, in lines the change did not write.
