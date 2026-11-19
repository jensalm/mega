---
title: loose ends around stops and restarts
state: researching
created: 2026-11-19
tags: [defect]
---

Six narrow faults in what mega does when a runtime stops, the service restarts or a connection returns. Each needs something to go wrong first, and none has been seen in use. They were found by review, and they are worth fixing because each is a place where mega's own rules are not kept.

## why

A terminus review of the last three batches' files on 2026-11-19 (`915e203707a7`) confirmed the five fixes of `the-page-says-more-than-it-has-seen` and raised five more blocking findings, all in older code. This was the third pass in a row to end that way. Michael's direction on 2026-11-18 was to clear the review before returning to pull request 3; the planning agent's recommendation on reading this pass was to card what is left and return to the pull request, which is the more urgent.

## the six

1. **A failed runtime's output can be counted late.** When a runtime exits in the middle of a turn, output it had produced in that turn is kept aside and never counted, and a later turn's end can then count it as unread. Recommended: count it at the exit. It is output the agent produced, the conversation reads as an error at that moment either way, and nothing is left over to surface later.
2. **Old imports can be counted unread after an unclean restart.** Resuming scans the whole log, so output brought in by an earlier adoption or import can show as unread. The rule is that imported output after a release or an adoption stays read. Only what this resume brings in should count.
3. **An answer can hide a request.** Answering removes the request from the pending set before the answer has been recorded. If the record fails, the request is hidden and still pending in the harness, and the agent waits on something the operator can no longer see. A request should leave the pending set only once its answer is recorded.
4. **A Codex answer can start work after a stop.** Between checking for an interrupt and starting the turn there is a gap, so an answer can start a turn after the interrupt has landed. And an answer is still accepted once a release has begun. An answer should be refused once an interrupt or a release is under way, with no gap between the check and the start.
5. **A reconnect can still show "Live" a moment early.** The page calls itself caught up when a workspace read lands, even if a change notice during that read has already queued another. It is caught up only when no read is waiting.
6. **The shutdown banner can report an old shutdown as under way.** After mega is shut down from a page and started again, the page's reconnect makes the banner say "Shutting down" and that every runtime is being released, before the restarted service has confirmed its workspace. The footer beside it correctly says "Catching up". The banner should wait in the same way.

## decisions inside it

The first and the fourth each hold one: whether a failed runtime's output is counted or cleared, and whether an answer is refused once a release has begun. The recommendations above are the planning agent's.

## a note on the review

The first five came from reviews by path of whole files, which the prompt for the previous batch asked for; that was the planning agent's instruction, not a limit of the tool. On 2026-11-19 the planning agent ran the review the batch is properly held to, the working tree's changes (`6082dde05ee4`). It was not clean either: it raised the same five and the sixth, and nothing against the five fixes themselves. So these are not an accident of reviewing too widely. A review of a change reads the files it touches, and these sit in those files, in lines the change did not write.
