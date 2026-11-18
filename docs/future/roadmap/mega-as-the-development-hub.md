---
title: mega as the development hub
state: horizon
created: 2026-11-11
tags: [epic]
---

Grow mega from a workspace for conversations into the hub of the whole development workflow: the roadmap, the review brokers and the practice itself reachable from where the operator's attention already is. Michael set this direction on 2026-11-11.

## why

Two days of building mega through mega showed the tax plainly. Michael carried some forty implementer reports to the planning conversation by hand and carried prompts back; roadmap cards were written in one place and rewritten as prompts in another; terminus verdicts arrived as prose inside a report; ranger sat in a separate window from the conversations it steered. All of it is transport, the same tax mercurius was built to remove between a design agent and a reviewer.

This is consolidation, not invention. Mega is new, but the practice it gathers is more than a year old and maintains on the order of a million lines of code: the roadmap convention is practice-wide, mercurius has about a hundred converged sessions behind it, terminus gates every stage of every project, and the roles and pipeline have been refined in daily use. Mega's part is to be the surface those proven pieces meet on.

The practice has a name, orienteering (the grimoire's `agents/orienteering`): a solo operator keeping a map honest against the ground, with the roadmap, the pipeline, the brokers and the repo disciplines as its parts. Mega is the wedge that refines and streamlines that whole practice into what it can be for a solo operator. The hub is where orienteering is run, and every piece below is one of its named parts arriving on one surface.

## the seven directions, from the presentation of 2026-11-18

Michael presented mega and named where it goes next. Each direction, and where it lives:

1. **The harness orchestration framework is extracted** from mega, mercurius and terminus, as the foundation for orchestration native to the overlay. `the-orchestration-framework-is-extracted`.
2. **Multiple environments.** Any number of orchestration environments, reachable anywhere, in a single pane of glass. `central-application`, with `a-workspace-that-can-be-rebuilt-elsewhere` and the two cards under it.
3. **Sandboxed, ephemeral development environments,** reached over the overlay. `sandboxed-ephemeral-environments`.
4. **Change digestion as a first-class model:** a change presented as a story. `change-digestion`.
5. **Terminus grows reference-model reviews:** quality steered by reference architecture components. Terminus's own roadmap, and the last piece listed below.
6. **Collaborative agent flows.** Work that is moved between agents by hand today may be coordinated by mega. `agents-roles-and-lifecycles-in-a-project`, with `hand-off-between-conversations` as its first step and `re-entry-briefing` beside it.
7. **Ranger grows an extensive metadata model:** epics, punchlists, defects, specs and work orders, and other structured project artifacts. Ranger's own roadmap, where `field-notes-from-megas-roadmap` holds the evidence, and the first piece listed below.

The first three stack: the framework, then environments built on it, then sandboxes among those environments. The fourth and sixth stack too, in his words: "as quality management and change digestion get better, attention scales wider."

## the principle

Mega carries; it never decides. The spec says mega is not a coordinating agent, not a pipeline engine, and holds no competing copy of project truth. Every integration here is something mega shows the operator or moves at his command. Choosing the next agent, advancing a stage, or keeping its own copy of a card or a verdict would make it the thing the spec ruled out.

The sixth direction tests this, and the tension is recorded here and not resolved. Michael, 2026-11-18: "today we manually move work around between agents. future mega potentially auto-coordinates." That is mega choosing the next agent, which this principle rules out. Two readings keep both: coordination that the operator sets up and mega then carries out, so the deciding is still his; or the principle giving way as review and change digestion become good enough to trust. Which it is belongs to the conversation on `agents-roles-and-lifecycles-in-a-project`.

A second rule follows from it: each tool keeps working without mega. Ranger's truth stays in files, the brokers stay usable from a terminal, the practice stays readable as plain documents.

## the pieces Michael named

- **Ranger inside mega, and ranger gaining issue management.** A project page shows its board from the files in its checkout. A conversation can be launched from a card, with the card as its opening prompt, which is what the convention already calls a card. Issues from the forge reach the roadmap with their `source`.
  - Michael, 2026-11-14, a week into daily use: "i suspect the velocity we're developing with mega will result in a substantial overhaul of ranger, and possibly integrating ranger's new functionality directly into mega. or at least making them work together very well." The evidence for what that overhaul needs is being kept as it happens, in ranger's own roadmap, on the card `field-notes-from-megas-roadmap`.
- **Terminus and mercurius as first-class citizens.** Mega understands a review as an object: running or finished, its verdict, its findings one at a time, its log, shown on the conversation and the project instead of scrolling past as tool calls. The brokers stay standalone and are observed, not absorbed. Their restartability, which per-turn mode already waits on, belongs to the same work.
- **A modular grimoire, v2, as the operating system for agents in the mega ecosystem.** The roles, conventions and pipeline in a form a project can adopt and a stranger can read. This is what makes a public mega more than an interface for two harnesses, and it wants its own spec and likely its own project.
- **Terminus referencing complete architectural examples.** A quality points at a running reference such as the console project, whose own instructions call it the standard. To design for: the reviewer's context budget, and pinning the reference so it cannot drift from the rule.

## a first step

An operator-initiated hand-off: send this message or this report to that conversation. Pure transport, inside the spec's lines, and the seed of mercurius living in mega. Then the board on the project page, then a review verdict on its conversation.

## harvested from CodeAF, 2026-11-18

Michael asked for a look at CodeAF (`github.com/Agent-Field/CodeAF`) for ideas. The planning agent read its README on this date; nothing was run, and its benchmark and memory figures are its own claims. The review had a card of its own, closed into this section.

**What it is.** A coding harness for open models, in Go, that calls itself a software factory: one window over every project on the machine, where you describe work and it sizes, splits, runs, checks and merges it, asking only when it must. That is the opposite of mega's principle. CodeAF decides; mega carries and never decides. So what is worth taking is how it shows and moves work, not who is in charge of it.

**Worth taking, roughly in order of fit:**

- **Each piece of work in its own copy of the repository, on its own branch.** Its tasks run where they cannot collide, and land on your branch only when a check passes, never on a main or release branch. For mega this is the concrete answer to running two implementers in one project: a conversation launched in a worktree of its own, by the operator's choice. It bears directly on `agents-roles-and-lifecycles-in-a-project`.
- **A home that answers five questions at once:** what needs you, what is unread, where you were, what is running, and what changed since you left. Mega has the first, second and fourth, and since 2026-11-17 a mark for the third. "Since you left" is `re-entry-briefing` without the inference: a plain list of what ended and what was asked while the page was not being watched.
- **Answering from the list.** A digit answers a question from its row, without opening the conversation. Mega's Allow keys answer the oldest docked approval of the open conversation only; the needs-you list could take an answer too.
- **Files crossing to the machine that owns the work.** Paste a screenshot or drop a file and it lands on the remote machine; click a path and it opens locally. Mega reaches files by path search, which needs the file to be on the service's machine already. A dropped or pasted file landing in a named directory would close that gap, and matters most through a share.
- **Typing that never waits on the network.** Its screen runs where you sit and the conversation where the work is. Michael found mega through a share slower than a local page, enough to change how it comes across. Worth knowing what in mega's page waits on a round trip that need not.
- **What each piece of work cost.** Every task says which models ran it and what it cost against an estimate, with a limit per task and per day. Mega shows no usage at all, though the harnesses report it.
- **Roles as seats.** A worker, a planner and a checker, chosen per task. Michael's practice has the same three as a planning agent, implementers and the review brokers; CodeAF fills the seats itself, where he fills them.
- **Idle sessions are not free.** Its own table puts sixteen idle Claude Code sessions at about 3.5 GB. Unverified, and the reason per-turn runtime mode is on the agents card.

**Not for mega.** The factory deciding what to run and what to merge; a headless mode that answers its own questions; automations that act on a clock; telemetry.

