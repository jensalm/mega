---
title: harness configuration backups
state: building
created: 2026-11-14
tags: [feature]
---

Mega keeps dated copies of each harness's configuration in its own data directory, away from the places the harnesses and their updaters manage, so a settings file an update rewrote or removed can be put back.

## why

Michael, 2026-11-14, while approving a hand-made backup before the version spike: "it might be worth also building a mechanism in mega to routinely back up configuration data outside of the reach of the typical harness locations." Mega depends on configuration it does not own. Claude Code and Codex update themselves, run migrations over their own files, and clean up after themselves, and nothing keeps a copy from before. The same day showed the shape of the risk: five Claude builds on disk, a cleanup that ran unannounced, and migration markers a different build could rewrite.

## shape

- **What is copied.** Configuration only, per harness: settings, instructions, rules, skills, plugin lists, the migration and version markers, and for Claude `~/.claude.json`. The list the version spike drew up by hand on 2026-11-14, on the card `harness-version-tolerance`, is the starting point.
- **What is never copied.** Credential and token files, transcripts, session stores, histories, databases and caches. The list is of things to include, so a file a new harness version adds is left out until someone decides about it.
- **Where.** Under mega's data directory, one dated folder per snapshot, readable only by the operator. `~/.claude.json` carries account identifiers and a line of prompt text per project, so the folder is treated as private as the store it came from.
- **When.** At start-up when the configuration differs from the newest snapshot, and whenever mega sees a harness report a version it has not seen before, which is the moment a migration may just have run. Never on a timer that copies unchanged files.
- **How many.** A bounded number kept, oldest removed, with the count in the config file. A snapshot from before a version change is kept longer than an ordinary one.
- **Visible.** `mega version` and the version panel say when the last snapshot was taken and where it is.
- **Restore is the operator's.** Mega shows what differs between a snapshot and now, and prints the copy commands. It does not write into a harness's configuration by itself, at least at first.

## to settle

Whether mega's own workspace wants the same treatment. It is much larger and already durable by design, so probably not here.

## checks

Go tests over a scratch home: the include list, a credential file beside the included ones left alone, nothing copied when nothing changed, the bound on how many are kept, and a snapshot forced by a new version.

## the first layer of something larger

Michael widened this the same day: conversations too, and anything short of credentials that would let a workspace be rebuilt on another host. That is the epic `a-workspace-that-can-be-rebuilt-elsewhere`, and this card is its first and smallest layer. It stays configuration only, so it can be built soon and by itself.
