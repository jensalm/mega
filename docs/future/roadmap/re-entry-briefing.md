---
title: re-entry briefing
state: horizon
created: 2026-11-06
tags: [epic, spike]
---

Local inference in the environment generating a short reading of where a project and its conversations stand, for Michael arriving back after circuits elsewhere: recent decisions, what is being asked of him, how the conversations relate. Generated at turn end from a bounded recent window plus the previous briefing, cached as derived state, labeled stale on new activity, linked to supporting messages, never able to send or approve anything. First endpoint is the Ollama already configured on `fortyfive`. The spec's attention section carries the design.

## background

Noted by the planning agent on 2026-11-16. This card names an Ollama on `fortyfive` as its first endpoint. Michael's local model of choice is now `qwen3.8-27b` on his host `eleven`, under a custom inference engine for a 5090, which is also what the pi adapter's round trips use. Which endpoint a briefing uses is open again when this is shaped.
