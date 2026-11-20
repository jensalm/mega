---
title: codex request raised under the runtime lock
state: inbox
created: 2026-11-20
tags: [defect]
---

Two paths take the same two locks in opposite orders, so they can deadlock each other. Found while working Terminus pass `5576b118c21c` on piece 1, `a-codex-question-is-a-message`, and not part of that piece.

- **The Codex adapter raises a request while holding its runtime's lock.** In `internal/harness/codex/runtime.go`, `request` and `itemCompleted` hold `r.mu` across `sink.Request`, so that an answer cannot arrive before the request is known. `sink.Request` takes the service's `s.mu` (`internal/service/live.go`).
- **The service asks the runtime for its session while holding its own lock.** `startRuntime` (`internal/service/conversation.go`) holds `s.mu` after a launch and calls `rt.NativeID()`, which takes `r.mu`.
- **Both orders date from the first slice:** stage 2 for the service's, stage 3 for the adapter's.

The window is the end of `startRuntime`, just after a launch, if Codex sends a server request or a message with questions at that moment. It has not been seen.

Either side could give way. The service could read the session before it takes `s.mu`. Or the adapter could record the request without holding `r.mu` across the sink call, with another way to keep an early answer from missing it. Claude and pi should be checked for the same shape.
