---
title: a workspace that can be rebuilt elsewhere
state: horizon
created: 2026-11-14
tags: [epic, spike]
---

Everything short of credentials that an operator needs to stand their mega workspace up again, on the same host after a loss or on a different one, is kept by mega in a form it can restore from: the harnesses' configuration, and each conversation with its history and its native session.

## why

Michael, 2026-11-14, widening `harness-configuration-backups`: "it might also be useful to back up conversations and anything short of credentials that would allow a user to reconstruct a harness environment on a different host. this would close a workflow resiliency gap... and it could also serve as the foundation for moving sessions between hosts."

The gap is real. A week of work now lives in three places per conversation: mega's registry, mega's event log, and the harness's own session store, which mega neither owns nor copies. Lose the disk, or a harness migration goes wrong, and the conversations are gone. And the two hosts mega runs on are two islands.

## the unit: a conversation bundle

One conversation, complete and self-describing: its registry entry, its event log, the native session the harness needs to resume it, and the facts needed to make it work somewhere else, such as the checkout path it was recorded under and the harness version that wrote it. The same bundle serves three purposes:

- **Backup.** Every conversation bundled on a schedule, beside the configuration snapshots.
- **Moving between hosts.** One bundle carried from one mega to another, which is `moving-conversations-between-hosts`.
- **Relocating a checkout.** A bundle restored under another path on the same host, which is `relocate-a-projects-checkout`.

## layers, smallest first

1. `harness-configuration-backups`: configuration snapshots. Small, and built soon.
2. The bundle: export and import of one conversation on one host, with the path rewrite. The shared spike.
3. Routine backup of the whole workspace as bundles, to a place the operator names.
4. Restore on another host, and a move between two running megas.

## what makes it hard

- **Size.** Configuration is a few megabytes. Mega's own logs were 185 MB after a week, and the harnesses' session stores are larger. Dated full copies will not do; this wants bundles written once and extended, or copied only when changed.
- **Privacy.** A transcript holds whatever an agent saw: file contents, command output, anything pasted. Excluding credential files does not make a bundle free of secrets. A bundle is as private as the store it came from, stays on the machine by default, and wants encrypting before it leaves.
- **Copying a store that is being written.** Claude's session files are appended to, and Codex keeps threads in databases. A copy has to be taken at a moment that is whole. Mega knows when a conversation is idle and when a turn ends.
- **Paths.** Claude keys its store by the checkout's path. A session restored where the checkout lives elsewhere is not found without a rewrite. The capability report's follow-up of 2026-11-11 recorded this.
- **Versions.** A session written by one harness version and restored under another. `harness-version-tolerance` is finding out how forgiving the stores are.
- **Two writers.** A conversation restored on a second host while still live on the first is the hazard every harness has. A move has to end the conversation where it left.

## what does not travel

- Credentials and logins. The operator logs in on the new host.
- Trust. Mega never sets a trust flag, so the operator accepts each harness's trust for the checkout again.
- Approvals of project environments. An approval is the operator's decision about what runs in one machine's shell; it is given again on the next.

## spike

Per harness, what the smallest native artifact is that resumes a session: for Claude the session file, for Codex whether the rollout file is enough or the databases are needed, for pi the session file and its branches. What must be rewritten for another path. Whether a session resumes under a newer and an older harness build than wrote it. Recorded in the capability report.
