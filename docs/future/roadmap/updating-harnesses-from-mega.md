---
title: updating harnesses from mega
state: horizon
created: 2026-11-14
tags: [feature, spike]
---

An update of Claude Code, Codex or pi becomes something the operator does from mega, at a moment mega knows is safe, instead of something that happens in the background of whichever terminal was opened last.

## why

Michael, 2026-11-14: "could we have mega also take care of updating the harnesses?" Today each harness updates by its own mechanism and on its own schedule. Mega finds out afterwards, or not at all: that day its Claude conversations were running a build seven versions behind the one in his terminal, against the same configuration, and nobody had chosen that.

## what mega is well placed to do

- It knows when nothing is running: the shutdown gate already computes idle across every conversation.
- It knows which versions it was tested against, once `harness-version-tolerance` gives a range.
- It can take a configuration snapshot first, once `harness-configuration-backups` exists.
- It can say what changed afterwards: the version each harness now reports, and whether it is inside the tested range.

## the planning agent's recommendation

A wrapper around each harness's own updater, started by the operator, and nothing more automatic than that:

- Mega shows that an update is available and whether it is inside the tested range. The operator presses update.
- Mega waits for idle as the shutdown gate does, snapshots the configuration, runs the harness's own update command, and reports the version before and after. It installs nothing by any route the harness does not itself provide.
- Mega never updates on a schedule of its own, and never turns a harness's own updater off unless the operator has asked for exactly that in mega's config. Turning it off is a change to the harness's configuration, which is the operator's.
- The terminal and mega use the same installed build. Mega keeping private builds of its own would split the two again, with both writing one session store.

## why not more

Running installers is running downloaded code with the operator's privileges. Each harness has its own mechanism, they differ by platform and by how the harness was installed, and they change. An external operator's install may be nothing like Michael's. A wrapper that calls the harness's own command inherits whatever that command does and adds a safe moment and a record. A manager that knows how to install each harness is a second package manager to maintain.

## spike

What each harness offers: an update command that can be run on demand, a way to ask what is available without installing it, a way to hold at a version, whether the old build is kept for going back, and what its updater does to running sessions. Recorded in the capability report with versions.

## depends on

`harness-version-tolerance` and `harness-configuration-backups`.

## where Michael thinks this may land

2026-11-14: "we might end up advising users to turn automatic updates off, and let mega manage the update and migration process to make sure it's safe for multiple sessions, etc." That goes past the wrapper recommended above, and the planning agent thinks it is probably right, on conditions:

- **The case for it.** Mega runs many sessions at once, and a harness that updates under them leaves old builds running beside new launches, all on one configuration and one session store. Only mega knows when every session is idle, so only mega can make an update one event: stop, snapshot, update, check, resume.
- **It needs a supported switch.** Each harness has to offer a way to turn its updater off and a way to update on demand. Where one does not, mega can only advise.
- **Turning the updater off is the operator's act.** Mega can notice that it is on, say why that is a risk with several sessions, and offer to change it. It changes nothing until the operator says so, and it records that they did.
- **Mega then owes the operator the updates.** A machine with updates off gets no security fixes by itself. Mega has to say when an update is waiting and for how long, plainly and persistently.
- **Mega has to know what is safe.** Advising people to update through mega means mega's tested range keeps up with releases. That wants the round trips run against each new harness release automatically, or mega either holds people back or blesses builds nobody tried.
- **Migrations are the harness's.** Mega can order the steps and keep a snapshot to go back to. It cannot make a harness's own migration reversible.

## from the version spike, 2026-11-14

Three observations that bear on this card:

- **A lever that touches nothing of the operator's.** The spike kept Claude's updater quiet by setting `DISABLE_AUTOUPDATER=1` and, from 2.1.322, `DISABLE_UPDATES=1` in the environment of the runtimes it launched. Mega could do the same for every runtime it starts. Sessions under mega would then never set off an update, with no change to the harness's own configuration. The operator's terminal would still update as before.
- **Claude's installer keeps a build while a session runs on it,** by its own documentation: the version the launcher points to, any version a session is running, and the two newest. What was observed did not fully match, since two builds with no running session survived a cleanup.
- **A running conversation keeps the build it started on.** So after an update, old and new builds run side by side until each conversation is released. That is the mixed state a managed update would avoid.
- **Two builds on one configuration were quietly undoing each other.** Later the same day the spike found that Claude Code 2.1.318 and 2.1.319 lower `migrationVersion` in `~/.claude.json` from 15 to 14 whenever they start, and the newer builds raise it again. Michael's mega ran 2.1.318 all week beside a terminal on newer builds, so that marker was being moved back and forth, and whatever migration 15 does was presumably being run again each time. Nothing visibly broke. It is the plainest evidence so far that builds should not be mixed on one configuration, and that an update wants to be one event for every session.
- **Codex was gentler.** Four builds from 0.163.0 to 0.169.0 shared one store and none migrated or rewrote anything; the oldest read what the newest wrote.

## where this stands, 2026-11-16

Both things this depends on are built: `harness-version-tolerance` on 2026-11-15 and `harness-configuration-backups` the same day. A version snapshot is already taken whenever mega meets a harness version it has not seen, which is the "snapshot first" step of an update, done after the fact.
