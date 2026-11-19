---
title: a turn is inferred from a message
state: horizon
created: 2026-11-19
tags: [enhancement, spike]
---

A turn begins and ends because the harness said so, in its own terms, and not because the service saw a message and guessed. A turn that goes quiet is shown as quiet, and an interrupt that gets no answer says so.

## why

Pull request 3 fixed one case of a general fault. Claude Code answers a model switch on an idle runtime with a user message outside any turn. Mega took it for the start of a turn, no end ever came, and the conversation showed as working for fifty minutes with Interrupt doing nothing. The follow-up on 2026-11-19 found that every Claude Code build from 2.1.322 on does this, so it had been broken inside the tested range all along, and asked the wider question.

The answer: the service opens a turn whenever a runtime reports a user message, or any assistant or tool output, while no turn is open. Only the harness's own turn end closes it. Nothing bounds it. So any message a harness sends outside a turn can do what the echo did. Candidates the implementer named, none yet seen: for Claude, a compaction summary or a hook's message; for Codex, an item arriving after its turn that the adapter does not mark as late; for pi, a message an extension injects outside a run.

## the implementer's recommendation

- **Each adapter reports the start and the end of a turn explicitly,** from its harness's own markers, and the service stops inferring a turn from a message.
- **Never end a turn on a guess.** Show how long an open turn has had no activity, and leave the judgement to the operator.
- **An interrupt that gets no answer at all** shows the turn as "interrupt not answered", and offers Release.

## to settle

What each harness's own markers are, and whether all three have both a start and an end the adapter can trust. What a turn the harness opens by itself looks like, such as Claude waking when a background job finishes, which is a real turn nobody prompted. And how this meets `honest-runtime-state`: a turn shown as running with nothing happening is a claim the service cannot stand behind for long.

## depends on

Nothing. It bears on `loose-ends-around-stops-and-restarts`, whose seventh item is the same question asked of history instead of the live stream.
