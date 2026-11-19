---
title: queued operator messages
state: researching
created: 2026-11-19
tags: [feature, spike]
---

the native harnesses all support typing and sending a message that will be "slip-streamed" to the agent when an opportunity arises in the flow. we need to reproduce that functionality in the mega ui.

## notes from the planning agent, 2026-11-19

- **What mega does today.** A prompt sent while a turn is in flight is refused. That refusal is also what keeps a shutdown from treating a conversation as at rest.
- **A spike first.** Each harness does this its own way, and the capability report has not looked at it: how Claude Code takes input during a turn, what Codex's `turn/steer` does and when it refuses, and what pi does with a prompt during a run. A probe for each, recorded with its version under `probes/`.
- **Where the message waits.** In the harness, if it has a queue of its own, or in mega until the turn ends. The second is simpler and the same for all three; the first is what the terminals do.
- **It meets two pieces of `simplify-the-core`.** `a-turn-is-inferred-from-a-message` moves the check that no turn is open to where a prompt is recorded, which is the place this would change. And `a-codex-question-is-a-message` removes mega's one present use of `turn/steer`, for an answer; this would bring steering back as a thing of its own.
- **Order.** After `a-turn-is-inferred-from-a-message`, so it is built on the rule and not beside it.
