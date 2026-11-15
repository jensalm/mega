---
title: an external operator can install mega
state: evaluating
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

## stories, in order

1. `clean-machine-rehearsal`: follow the README on a machine that has never seen mega, and write down everything that goes wrong. Its findings reorder what follows.
   - The agent's pass ran on 2026-11-13 and produced three cards, built together as the first-run batch before anything else: `a-missing-claude-is-named`, `dialog-focus-and-a-stale-model-list` and `first-run-explains-itself`. Michael's own pass under WSL2 comes after that batch and the version work, so it does not stop at blockers already known.
2. `harness-version-tolerance`: what mega does with a Claude Code or Codex other than the pinned one, and what it should say.
3. `pi-adapter`: the third harness.
4. `the-checks-run-in-a-browser-michael-does-not-use`: the demo runs in LibreWolf and the checks run in Chromium, so this protects what follows.
5. `hotkey-for-allow` and `action-view-improvements`: the polish a live demo shows.
6. `mega-on-macos-and-windows`.
7. `mega-on-twentytwo`: the first tagged release, delivered by `push`, with twentytwo as the proving ground.
8. `readme-for-newcomers`.
9. `the-demo`: a workspace made for showing, with no freeze before it.

Not on the path, and ordered after: `project-mcp-servers-and-hooks-run-without-approval`, which matters more once other people's repositories are opened, and `claude-shim-distribution`.

## done when

An external operator on macOS or Linux, and one on Windows through WSL2, has gone from the README to a first conversation without help, and the README says plainly what is untested.
