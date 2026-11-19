---
title: a model switch on an idle runtime is untested
state: researching
created: 2026-11-18
tags: [enhancement]
---

Switching the model of an idle conversation is exercised for real, in a round trip and in the version probe, so the next harness build that changes how it answers is caught by a run and not by an operator.

## why

Pull request 3, from `plorenz`, 2026-11-18: with Claude Code 2.1.327, switching an idle conversation's model left it showing as working for fifty minutes, with Interrupt doing nothing. That build answers the switch with a user message on the live stream, marked as a replay, outside any turn. Mega took it for the start of a turn that never ended. The fix maps such a message to a native event.

Two things let it through:

- **No round trip switches the model on an idle runtime.** The model round trip launches with a model and switches during use. The pull request says so itself: one would have caught this on the version bump.
- **2.1.327 is outside the tested range,** which stops at 2.1.325. Michael's own install was at 2.1.326 on this date. The range is extended by running `probes/versions/` against a new build, and the probe has no idle model switch in it either.

## shape

- An e2e round trip: start a runtime with no prompt, switch its model, and assert the conversation is idle afterwards, takes a prompt, and that the turn ran on the new model.
- The same step in `probes/versions/driver.mjs`, and the equivalent for Codex and pi where a switch while idle is possible.
- Run the probe against the installed Claude Code, record what it sees in the capability report, and extend `harness.Tested` to the builds it passed.
- The service-level question the fix does not ask: a runtime reporting a user message while no turn is open starts a turn. Whether anything should bound a turn that a runtime opens and never ends, or say so on the page, is worth a look while here.
