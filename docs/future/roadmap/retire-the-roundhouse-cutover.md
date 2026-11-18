---
title: retire the roundhouse cutover
state: researching
created: 2026-11-18
tags: [enhancement]
---

Remove what was left behind for the rename from roundhouse to mega: the start-up check that refuses an empty workspace beside an unmoved roundhouse one, and the one-off checklist it points at.

## why

The rename was on 2026-11-11. The check and the checklist existed for one operator on a known set of machines, to stop a new binary opening an empty workspace beside the full one under the old name. Michael, 2026-11-18: "no machines have an unmoved roundhouse workspace." No external operator ever had a roundhouse workspace, so none can meet the check. A terminus review of the same day noticed the leftover on its own account: the refusal still points at `docs/future/mega-cutover.md`, a document under `docs/future` describing something finished.

## shape

- Remove `checkMoved`, `ErrNotMoved` and the legacy name from `internal/config/config.go`, with their tests and wherever the error is handled.
- Remove `docs/future/mega-cutover.md` and every reference to it.
- `docs/current/mega.md` and the README lose whatever they say about refusing to start beside a roundhouse workspace.
- The changelog's entry for the rename, which describes the refusal, stays as written, with a new entry saying the check is gone.
- The probes that still call themselves roundhouse in their client names are records of what was run and stay as they are.

## not here

The backup of the old workspace on Michael's machine, `~/.local/share/roundhouse.bak-2026-11-11`, is his to remove.
