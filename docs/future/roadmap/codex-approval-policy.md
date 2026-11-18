---
title: codex approval policy
state: researching
created: 2026-11-12
tags: [feature, spike]
---

Let the operator choose a Codex conversation's approval policy and sandbox from mega, at launch and while it runs, the way the permission mode works for Claude. Today a Codex conversation runs with whatever the operator's `config.toml` says and mega only displays it.

## why

The Claude side got this in the permission-mode batch, and daily use showed it matters: approving every operation by hand is the loudest cost in the workspace. Codex has the same range of choices under other names.

## shape

Follow the Claude pattern as it now stands, which changed on 2026-11-13: the conversation's request is concrete from creation, copied once from the operator's own Codex config, passed at every launch and changeable from the control bar. The panel shows the request and each turn's line shows what ran. There is no unset or inherited state. Mega never selects a policy on its own and never offers the policies that remove the sandbox or approvals altogether; the canon's `permissions-never-weakened` is amended the same way it was for Claude, to forbid a policy mega chose and allow one the operator chose.

## spike

Which policies and sandbox modes codex-cli offers at the pinned version, and which of `thread/start`, `thread/resume` and `turn/start` carry them; whether a change mid-conversation applies at the next turn, as model and effort do; what the app-server reports back; and how the per-tool "allow for the session" grants from the MCP approval batch interact with a policy change.
