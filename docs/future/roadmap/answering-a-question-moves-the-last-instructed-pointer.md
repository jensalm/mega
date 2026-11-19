---
title: answering a question moves the last instructed pointer
state: researching
created: 2026-11-19
tags: [enhancement]
---

when an agent asks a question, answering the question should move the last instructed pointer.

## notes from the planning agent, 2026-11-19

- **This changes a rule Michael set on 2026-11-17** for the marker of the agent last instructed: "i would count a prompt of any kind, and nothing else." The service's test for it is `instructs` in `internal/service/live.go`: a user event mega sent. An answer to a request is not one today.
- **Which answers.** An answer to a question is the case named. To settle: whether Allow or Deny on a permission counts too, or only an answer that says something.
- **One case moves by itself.** After `a-codex-question-is-a-message`, a Codex question asked in a message is answered in the reply box, which is a prompt and already counts.
- **Size.** Small, and on one surface. It can ride between two pieces of `simplify-the-core`.
