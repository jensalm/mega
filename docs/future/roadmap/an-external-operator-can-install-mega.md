---
title: an external operator can install mega
state: researching
created: 2026-11-13
tags: [epic]
---

Someone who watched the demo can get mega running on their own machine, from the README, without Michael beside them. Everything that stands between the repository as it is and that outcome is a story under this card.

## why

Michael presents his toolset in under a week from 2026-11-13, with a live demo of mega, and expects to be asked how to get a copy. Whoever takes one is an external operator: someone other than Michael running their own mega, on their own machine, with their own harness logins. Mega has been used daily for a week on two hosts, both his, both Linux, both built from a checkout he keeps in place, with harness versions he pinned by hand. None of that is true of the next person.

## decided, 2026-11-13

- **Building from source is acceptable for the first users.** A downloadable binary is not required, so `claude-shim-distribution` is not on the path to the demo.
- **The first external operators are mostly on Windows and macOS, with a few on Linux.** macOS is a target for this epic. Windows is reached through WSL2 for now, and native Windows waits. `mega-on-macos-and-windows` holds both.
- **pi is slipstreamed in.** The spec deferred it; Michael pulled it forward.
- **The work is sequential.** One story at a time, in the order below. Running two implementers at once on one repository is coordination that falls on Michael, and helping with that is something mega has yet to grow into.

## stories, as of 2026-11-16

**Built, in evaluating**

- `clean-machine-rehearsal`, the agent's pass, and the first-run batch it produced: `a-missing-claude-is-named`, `dialog-focus-and-a-stale-model-list` and `first-run-explains-itself`.
- `harness-version-tolerance`: Claude Code is found on the `PATH`, each harness has a tested range, and a lone Codex binary is named.
- `harness-configuration-backups`: configuration snapshots under mega's data directory, with `mega snapshots` to list and compare them.

- `pi-adapter`: the third harness, built against pi 1.5.2 and the local model.

**Next, in order**

1. The conversation polish a live demo shows: `the-version-shown-goes-stale-after-an-update`, `hotkey-for-allow`, `action-view-improvements`, `conversation-timestamps` and `path-search-beyond-the-checkout`.
2. `readme-for-newcomers`.
3. `the-demo`: a workspace made for showing, with no freeze before it.

Michael, 2026-11-16: with everything up to `the-demo` handled, mega is in good shape for the demo.

**After the demo**

- `agent-hand-off-turns` and `remote-access-through-zrok`.
- `clean-machine-rehearsal`, Michael's own pass under WSL2, which is also the Windows half of `mega-on-macos-and-windows`.
- `mega-on-macos-and-windows`, the macOS half, when a volunteer has a Mac.
- `mega-on-twentytwo`: the first tagged release, delivered by `push`.
- `the-checks-run-in-a-browser-michael-does-not-use`.

Not on the path, and ordered after: `project-mcp-servers-and-hooks-run-without-approval`, which matters more once other people's repositories are opened, and `claude-shim-distribution`.

## done when

An external operator on macOS or Linux, and one on Windows through WSL2, has gone from the README to a first conversation without help, and the README says plainly what is untested.
