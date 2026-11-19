# mega: a simpler core, work order

Work order for the section "A simpler core" in `docs/future/mega.md`, drafted 2026-11-19. It is a work order to remove things. Its measure is that `internal/service` is smaller at the end of each stage than at the start, and that a full-tree terminus review of the result has less to find.

## Why now

Three days of hardening fixed about forty-five review findings, every one real. Each full pass still returned five, and a third of each pass was a hole in an earlier pass's fix. Sorted by what made each possible, most were not about any feature:

| Cause | About |
| --- | --- |
| A write that fails partway through | 11 |
| The page's sense of whether it is current | 8 |
| Work after a stop, or in the gap before a start | 7 |
| Whose session a stream or a file belongs to | 7 |
| What counts as unread after a lost runtime or an import | 7 |
| The rest | 5 |

Michael, 2026-11-19: "rather than patching and covering every possible corner case, we might consider simplifying the code or the model where we can." And: "we should aggressively pursue the simplification options with the spec and work order."

The card is `docs/future/roadmap/simplify-the-core.md`. The hardening card, `harden-requests-stops-and-restarts`, is paused at the end of its fifth pass and resumes after this.

## Decisions for Michael before stage 1

Each changes what mega promises, so each is his. The planning agent's recommendation is the one stated.

1. **A failed write stops mega.** Today a failed write puts the service in a state where it keeps its last confirmed view, refuses that one change, and carries on, with a rule for each operation about what the page may then be told. Recommended: any failed write to the data directory is fatal. Mega tells every page that it can no longer record and has stopped, releases every runtime, and exits. The next start is an unclean restart, which is a path that has to exist and work anyway. This deletes the degraded mode and every per-operation answer to "what if the record fails".
2. **Unread means there is output beyond where you have read.** Today output is counted when its turn ends, with exceptions for output after its turn, for a runtime that exits mid-turn, and for imports by how they happened. Recommended: output counts when it is recorded. A conversation with a turn in flight reads as running whatever it holds, so nothing visible changes while it runs. After a resume from an unsupervised state, the reading position stays where it was, and whatever lies beyond it is unread. After a release or an adoption, where the operator asked for the import, the reading position moves to the end. Nothing works out which output is new.
3. **What changes about a conversation is kept in its log, and nowhere else.** Today a conversation's runtime state, its request, its reading position and its unread marker are in the registry, and the events that changed them are in the log. Recommended: the registry holds what a conversation is and how the operator arranged it: its project, harness, name, label and star. Everything else is read from the log. This is a one-way change to the data directory's format.
4. **Start-up reads the end of each log, not all of it.** A consequence of 3, and it absorbs the card `start-up-reads-every-log`. The log carries its own checkpoints, and start-up reads back to the last one.

## The shape after it

### One fold

A conversation's state is a value: its native session, its runtime state, what it requests, where it has been read to, its newest output, whether a turn is open, its last error and why, when it was last instructed. Today parts of that value are in memory, parts in the registry, and parts are worked out at start-up from other events by code that is separate from the code that decided them. Three passes of hardening went into one such piece of working-out.

After this there is one function, `apply(state, event) state`, with no I/O. The service changes a conversation's state in exactly one way: it appends an event to the log, and when the append is confirmed it applies the event. Start-up builds each conversation's state by applying the same function to the same events. What the service decided and what a restart recovers cannot disagree, because they are one piece of code.

Anything the service decides that is not obvious from an event already in the log gets an event of its own. "Counted as unread", "error cleared", "request filled in" are recorded as decisions, never inferred later.

### One record

The registry is `registry.yaml` as now, holding projects and, for each conversation, only what has no history: its id, project, harness, name, label and star. It is written when the operator arranges something, and at no other time.

The log is the only record of everything else. A checkpoint is an event that carries a conversation's whole state. One is written at a clean stop, at the end of a turn, and after every so many events. Start-up reads each log backwards to its last checkpoint and applies what follows.

### One failure

A write that fails is fatal, as decision 1. There is no state between "recorded" and "not recorded" for any code to handle.

### Turns the harness declares

Each adapter reports the start and the end of a turn from its harness's own markers. The service no longer opens a turn because it saw a message. A message outside any turn is recorded and opens nothing. A turn with no activity is shown as quiet, and an interrupt that gets no answer says so. This is the card `a-turn-is-inferred-from-a-message`.

### One kit under the adapters

Four things are written three times today and were each fixed three times: checking that a stream or a file belongs to the session asked for, refusing what would start work once a stop has begun, finding what a runtime left running, and ordering an answer against the harness's own resolution of the same request. They move into `internal/harness` as one implementation. An adapter is then what it was meant to be: a translation of one harness's protocol, with no policy in it.

### One state for the page

The page has one value for its connection: connecting, catching up, live, lost, shutdown asked, shut down. It is decided in one place. The footer, the rail, the shutdown banner, the held controls and the read all take it from there.

## Stages

Each stage is gated by terminus to `clean` on its working tree, then by Michael. Each ends with its journal entry and `docs/current/` caught up. Each must leave `make test`, all of `make checks` and `make e2e` passing, and `internal/service` no larger than it found it.

**Stage 1: unread, made plain.** Decision 2. Remove the buffering of a turn's output until its end, the flag for output after its turn, the scan a resume makes for what is new, its bounds, and the detection of an open turn in Claude's and pi's histories where it served only that scan. The canon's `bookkeeping-never-creates-attention` is amended in the same stage. Gate: the circuit from the spec, by the checks; and by hand, an unclean restart of a scratch mega with a turn in flight, resumed, showing unread exactly once.

**Stage 2: turns declared.** Each adapter reports turn start and end. The service's inference goes. Gate: the round trips for all three harnesses, including the model switch on an idle runtime, which is the case that started this; a message outside a turn opens none.

**Stage 3: one fold.** Introduce the state value and `apply`. Every state change in the service becomes append, then apply. Start-up builds state by the same function, still reading whole logs. Remove each piece of start-up code that works a fact out from other events. The registry is still written as before in this stage, and no longer read for anything the log can say. Gate: for every log in a copy of Michael's real workspace, the state built at start-up equals the state the old code produced, compared field by field by a one-off tool, with every difference explained.

**Stage 4: one record, one failure.** Decisions 1 and 3. The registry loses its changing fields. A failed write stops mega. The degraded mode, the uncertain-registry state and the two-phase writes go. The data directory gets a format number, and a start on an older one migrates it: for each conversation, a checkpoint written from the old registry's fields, then the fields removed. Gate: the migration run on a copy of Michael's workspace, interrupted at several points and rerun; an injected write failure stops mega with the page saying so, and the next start shows the running conversations as unsupervised.

**Stage 5: the end of the log.** Decision 4. Checkpoints, and a start-up that reads back to the last one. Gate: start-up time on a copy of Michael's workspace, before and after, and unchanged state.

**Stage 6: one kit.** The four shared mechanisms move under the adapters. Gate: the round trips and the version probe for all three harnesses; each adapter's file holds no policy the other two also hold.

**Stage 7: one state for the page.** Gate: the side, shutdown and panel suites; the reconnect and shutdown cases from the hardening passes each pass against the one value.

Then the full-tree review resumes, on `harden-requests-stops-and-restarts`, and its count per pass is compared with five.

## What must not change

- A session is resumed by its exact id, and a new one is started only on proof that none exists.
- Permissions are never weakened, no trust is set for the operator, and nothing is answered for him.
- A state mega cannot vouch for is never shown as certain. A simpler mega may say "look at this" more often. It may not say something untrue.
- A pending request is answered once, by the operator, and never after it stopped being pending.
- Everything an operator sees when nothing has gone wrong. Each stage's checks are the proof.

## Risks and what to watch

- **It is the core.** Every stage touches the files everything else leans on. The checks, about 700 steps in nine suites, and the round trips are what make this safe to attempt. A stage that needs a check weakened is a stage that has changed behaviour, and stops.
- **Michael's data.** Stage 4 changes the format of a workspace he works in every day, on three machines, and that external operators now have too. The migration is proven on copies first, is safe to interrupt, and the changelog says plainly that a workspace opened by the new mega cannot be opened by an older one.
- **The comparison in stage 3 is the keystone.** If the new fold and the old code disagree about a real log, one of them is wrong about something that happened, and that has to be understood before going on.
- **Stopping on a failed write is a change in character.** Mega stops being a service that limps on. It is the right trade for one operator on his own machine, where a full disk is rare and a restart is cheap. It should be revisited if mega ever runs where neither is true.
- **Size is the measure, and it can be gamed.** Moving code out of `internal/service` is not removing it. Each stage reports what was deleted outright.

## What this leaves alone

The harness adapters' protocols, the project environments, the page's layout and controls, snapshots, path search, and everything on the roadmap's horizon. The extraction of the orchestration framework is not attempted here; stage 6 only makes it easier.
