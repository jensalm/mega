---
title: simplify the core
state: researching
created: 2026-11-19
tags: [epic, spike]
---

Make the core of mega smaller where it can be, so that whole kinds of fault have nowhere to live, instead of finding and patching each instance. A design step first, then the changes, then the full-tree review continues on a tree with fewer corners in it.

## why

Michael, 2026-11-19, five passes into `harden-requests-stops-and-restarts`: "part of this work needs to be an effort to simplify the core architecture. rather than patching and covering every possible corner case, we might consider simplifying the code or the model where we can."

The hardening work has fixed about forty-five findings in three days, every one real, and each pass still returns five. A third of each pass is now a hole in an earlier pass's fix. That is what patching a complicated thing looks like. The findings are evidence of where the complication is.

## what the findings say, counted by cause

The planning agent sorted the findings of 2026-11-18 and 2026-11-19 by what made each possible. Roughly:

- **A write that fails partway, about eleven.** A change is recorded in the event log and then applied to the registry, or the other way round, and each place that does so needs its own answer to "what if the second write fails". An answer hiding a request before it is recorded; a settlement that could not be recorded; an unread marker lost when its write failed, and then three passes of trying to rebuild it.
- **The page's sense of being current, about eight.** "Live" before it had caught up, twice; a shutdown banner, twice; reads of output not yet seen; controls live while stale.
- **Work after a stop, and the gap before a start, about seven.** An answer starting a turn after an interrupt; a fence too narrow; a launch finishing after shutdown began.
- **Whose session is this, about seven.** One harness at a time: a file imported by name, a history read unchecked, a check made once, a check made on a different read.
- **What counts as unread across a lost runtime or an import, about seven.** Output from a failed turn; old imports recounted; a turn still open when its history was read.
- **The rest:** what a release left running, tests that answered without looking, and the cases where a message outside a turn opened one.

Four of these six are not about a feature. They are about how the core is put together.

## candidates, largest effect first

1. **One record for each fact, and stop when it cannot be written.** Today a conversation's state is in two places, its event log and the registry, written one after the other. Keep the registry for what a conversation is: its project, its harness, its session. Keep everything that changes in the log alone, and have start-up read where each log left off. Then there is no second write to fail. And when any write does fail, mega says once, on every page, that it can no longer record, and refuses everything until it is restarted. That replaces a rule for each operation with one rule, and would have left the first group with nothing to find.
2. **A coarser, honest unread.** Count output when it is recorded, not when its turn ends: a conversation with a turn in flight reads as running either way, so the buffering buys nothing that can be seen. After a runtime is lost, or history is brought in from outside, do not work out exactly which output is new. Mark the conversation as wanting a look, once. That removes the open-turn detection, the recovery scan and its bounds, and most of the fifth group.
3. **Turns the harness declares.** Already a card, `a-turn-is-inferred-from-a-message`. It removes the cases where a stray message opens a turn nothing closes.
4. **One kit under the three adapters.** Checking a session's identity, fencing a stop, finding what was left running and ordering an answer against its resolution are each written three times and were each fixed three times. An adapter should translate one harness's protocol and nothing else. This is the first step of `the-orchestration-framework-is-extracted`, arrived at from the other side.
5. **One state for the page.** Connecting, catching up, live, lost, shutting down: one value, decided in one place, that the footer, the rail, the banner and the held controls all read.

## how this would run

- **The spec first.** The first two change what mega promises, not only how it is built: what survives a crash, and what "unread" means after one. `AGENTS.md` says a discovery that changes what mega should do goes to the spec before it goes into a batch.
- **Then one candidate at a time,** each a batch of its own that removes code, with the checks and round trips as the proof that nothing an operator relies on changed.
- **Then the full-tree review resumes.** It is the measure: fewer findings per pass on a smaller core.

## the work order, 2026-11-19

Michael, the same day: "we should aggressively pursue the simplification options with the spec and work order." Both are drafted:

- **The spec:** the section "A simpler core" in `docs/future/mega.md`.
- **The work order:** `docs/future/mega-simpler-core-work-order.md`, seven stages, each gated by terminus and then by Michael, each required to leave the service no larger than it found it.

Five decisions were his, and he settled all five as recommended the same day ("let's consider them settled and see where we land"): a failed write stops mega; unread means output beyond where you have read; what changes about a conversation is kept only in its log; start-up reads the end of each log; a stop mega forced does not clear an error.

The candidates above became the stages in a different order, smallest promise first: unread, declared turns, the one fold, the one record and the one failure, the end of the log, the adapter kit, the page's state. The fold was not a candidate. It came out of asking why the unread rebuild failed three times: live code and start-up code were two tellings of one rule.

## what it must not cost

The promises stay: a session is resumed by its exact id; permissions are never weakened; nothing is answered for the operator; a state mega cannot vouch for is never shown as certain. A simpler model that says "I do not know, look at this" more often is acceptable. One that says something untrue is not.
