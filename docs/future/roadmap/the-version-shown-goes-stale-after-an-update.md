---
title: the version shown goes stale after an update
state: evaluating
created: 2026-11-15
tags: [defect]
---

After a harness updates while mega is running, the first page and the version panel keep showing the version from before, until mega restarts. They should show what a launch would get now.

## why

Reported by the implementer of `harness-configuration-backups` on 2026-11-15, and older than that batch. The start-up check caches each harness's version by where the program was found. For Claude Code that is the link `~/.local/bin/claude`, which stays the same when the updater points it at a new build. So the cache keeps answering with the old version.

A launch reads the version afresh, so the "Runtime started" line, the version snapshots and the tested-range check at launch are all right. Only the two places an operator looks to see what is installed are wrong, and Claude Code updates every day or two, so on a mega that runs for days they are wrong most of the time.

## shape

- Key the cached version by what the path resolves to, or by the file's identity and modification time, so a moved link is a new entry. Or drop the cache: a version read takes about 10 ms.
- The same for Codex, whose program is replaced in place by an npm update.
- A check in the firstrun suite: the fake's version changes behind the same path, and the first page follows without a restart.
