---
title: harden requests, stops and restarts
state: researching
created: 2026-11-19
tags: [enhancement, spike]
---

Review the full tree with terminus and work through every finding until it is clean, so that a review of ordinary work comes back clean unless the work itself is not. Most of what it will find is in three places: the life of a request, what happens when a runtime is stopped, and what is counted when one is lost and resumed.

## why

Between 2026-11-18 and 2026-11-19, five reviews of five different batches each confirmed the batch and then blocked on five or six older findings in the files it touched. Every finding was real. Twenty-odd were fixed, in two rounds built for nothing else, and the sixth review returned six more. Michael, 2026-11-19: "we need to figure out how to clean up these recurring terminus findings... should we do full-tree terminus until clean?"

Three things explain the pattern:

- **The faults cluster.** Nearly all sit in `internal/service/conversation.go` and `live.go`, the adapters' stop paths, and the page's sense of whether it is current. Each concerns one of a few sequences: answering a request, stopping a runtime, resuming after a loss.
- **Those sequences are written out by hand each time.** Record, then change memory, then tell readers; check for a stop, then start work. Each place that does it can get the order wrong in its own way, so fixing one finding leaves its siblings.
- **A reviewer samples.** Each pass reads what it has room for and reports what it sees. Fixing after every pass means never seeing the whole list.

The second is the one to act on. It has a precedent here: stage 2 went four rounds without a clean review until every runtime state change was put through one helper, `Service.transition`, and then it converged.

## the method: the full tree, to clean

Michael, 2026-11-19, on the planning agent's first answer, which argued against this: "terminus always eventually converges, in every other project? by working through all of the findings and either adjusting the code or the qualities, eventually we get to `clean`. i don't understand why we wouldn't just do that here?"

He is right, and the first answer was a prediction against the evidence. Terminus has converged on every other project of his, and on this one: each of the four stages of the first slice reached clean, the second after a helper was built in its fifth round. The planning agent's worry, that fixing after each pass has no reason to end, is not what has happened anywhere.

A clean full tree is also the direct cure for what prompted the question. Reviews of ordinary batches have been blocking on older findings in the files a batch touches. With nothing older left to find, a batch's review is clean unless the batch itself is not.

So: review the full tree, work through every finding, and repeat until it is clean.

- **Each finding ends one of two ways:** the code changes, or the quality does. A finding that is a fault is fixed. A finding that is a matter of judgement goes to Michael as a proposed change to the canon, with the reason, and the canon is his to amend.
- **Nothing is listed and left.** The rule of the last few rounds, that an older finding is reported once and not touched, does not apply to this work.
- **A fix is proven before it is kept:** a test that fails first, then `make test` and `make checks`.

What survives from the first answer is advice on how, not whether: where several findings share one sequence of steps, fix the sequence in one place, which is what made stage 2 converge.

## where one change may settle several findings

- **A request's life in one place.** Pending, being answered, answered, aborted, lost. One function decides each move, records it, and only then changes what readers see. Four of the fixed findings and three of the six below are about this order.
- **One fence for a stop.** Once an interrupt or a release has begun, one check decides what may still be delivered, for every harness. Fixed for answers on 2026-11-19; a Codex answer already accepted still gets through.
- **One sense of "current" in the page.** The footer, the rail, the shutdown banner and the held controls each ask in their own words whether the page has caught up. Three findings were the same question answered differently.
- **Turns that the harness declares.** `a-turn-is-inferred-from-a-message` is this card's fourth candidate, already written.

## the seed list: six older findings from the review of 2026-11-19

Listed once by the implementer of `loose-ends-around-stops-and-restarts`, and left:

1. Release does not fence a Codex answer already accepted, which can start a turn just after Release returns; the adapter then interrupts it.
2. A harness's own resolution of a request that cannot be recorded leaves the request answerable.
3. An answer the adapter rejects without settling it, such as a choice pi did not offer, is recorded as aborted, removing a dialog pi still waits on.
4. pi's history never marks an open turn, as Claude's did not until that round.
5. A dropped connection after a shutdown request reads as a completed shutdown, in the banner, the footer and the rail.
6. Resume logs a failure to save the unread marker, then clears "unsupervised" anyway.

## also for terminus

A card in terminus's own roadmap proposes a baseline of known findings, so a carded finding stops coming back as new. It was written before Michael's answer above. A clean tree needs no baseline, so it matters only for a project that chooses not to go to clean.

## after two passes, 2026-11-19

The seed list and two full passes are worked: five of the six seeds and ten findings fixed, each with a test seen to fail first. Two changes of shape came out of it, as hoped: every move of a request out of the pending list goes through one function that records first, and every runtime has one stop fence, latched in all three adapters, that refuses anything but Deny or Dismiss. The fence came back once because its first version was too narrow, which is the one finding that recurred.

The count went five, five, five. The implementer's reading, which the planning agent shares: these are not the same findings returning. Each pass reads a different part of a large tree, and the three areas still hold more. It should fall as they are worked through.

Decisions, as the planning agent recommended them to Michael:

- **The shutdown banner.** Fix it without changing what an ordinary shutdown says: the service tells the page when a shutdown has finished, as its last word before it closes. A page that heard it says mega was shut down. A page that asked and then lost contact without hearing it says a shutdown was asked for and contact was lost.
- **A request whose settlement could not be recorded** stays on the page as last confirmed, unmarked, with further answers refused. The canon's `single-authority` gained a line saying so, so two reviews do not disagree about it again.
- **Go on,** reporting after each pass.

One thing noticed on the way, not yet a finding: when a write has failed and the service refuses further writes, nothing on the page says so until the operator tries something. A sign that mega can no longer record would say it once, for everything.

