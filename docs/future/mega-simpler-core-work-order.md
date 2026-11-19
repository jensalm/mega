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

## Decisions

Each changes what mega promises, so each was Michael's. He settled all five as recommended on 2026-11-19: "they sound reasonable. let's consider them settled and see where we land." Each is written below as it was put to him, with what it replaces and what it costs.

1. **A failed write stops mega.** Today a failed write leaves the service running on its last confirmed view. The registry and each event log have an uncertain state of their own that refuses later changes, and each operation has its own rule for what the page may be told when its record fails. Recommended: any failed write to the record, which is the registry and the event logs, is fatal. Mega tells every page that it can no longer record and is stopping, ends every runtime it started, and exits. The next start is an unclean restart, which is a path that has to exist and work anyway. This deletes both uncertain states and every per-operation answer to "what if the record fails".
   - **The cost:** a turn in flight when the write fails is cut off, and what it had done is brought back from the harness's own history on resume. The alternative is to exit and leave the runtimes, as a crash does. The recommendation is to end them, because a harness nobody is watching can raise a request nobody can answer.
2. **Unread means there is output beyond where you have read.** Today output is counted when its turn ends, with exceptions for output after its turn, for a runtime that exits mid-turn, and for imports by how they happened. Recommended: output counts when it is recorded. A conversation with a turn in flight and no pending request reads as running whatever output it holds, so nothing visible changes while it runs. The order of attention is unchanged: a pending request, then unsupervised, then a turn in flight, then an error, then unread. Whether there is unread output is worked out on its own, apart from which of those the mark shows. Only reading moves the reading position. Output brought in from a harness's history counts like any other, however it came: after a resume from an unsupervised state, after a release and a resume, after an adoption. Nothing works out which output is new.
   - **Revised by Michael on 2026-11-19,** on the review's finding `R2-04`. The first draft moved the reading position to the end after an import the operator asked for. That could pass over output he had not seen: a conversation with an unread reply, released and resumed from elsewhere in the workspace, would drop off the needs-you list.
   - **The cost:** a conversation released, worked on in a terminal and resumed shows as unread until it is looked at, and so does a session just adopted. Resumed from its own page, the page reads it at once and the mark is never seen.
3. **What changes about a conversation is kept in its log, and nowhere else.** Today a conversation's runtime state, its request and its unread marker are in the registry, and the events that changed them are in the log. Recommended: the registry holds what a conversation is, how the operator arranged it and how far he has read it: its id, project, harness, name, label, mode and reading position. Everything else is read from the log. This is a one-way change to the data directory's format. The registry as it was is kept beside the new one, so going back is possible at the cost of what changed since.
   - **The native session id moves too.** The log already records it, in the `started` event, and will record it sooner, as "One failure" says. The registry's copy is the second write that a launch can fail on. The cost: the spec promises a registry that names each conversation's session and can be read with the service stopped. After this, that takes the registry and the logs, and a `mega` command that prints the roster from both would restore it.
   - **The reading position stays in the registry.** Michael reversed the first draft's move on 2026-11-19, on the review's fifth finding. It is the operator's fact and changes only when he reads. A read recorded in the log would take a sequence number, be sent to the page, and be acknowledged as something new to read, which is another read. Kept in the registry it is one write, as now, and no record is added to any log.
4. **Start-up reads the end of each log, not all of it.** A consequence of 3, and it absorbs the card `start-up-reads-every-log`. The log carries its own checkpoints, and start-up reads back to the last one.
5. **A stop mega forced does not clear an error.** Carried over from the hardening card, and small. When mega itself stops a runtime because something was wrong, such as a runtime on the wrong session, the stop ends as stopped, and today that clears the error. Recommended: the error keeps showing until a runtime starts. A stop the operator asked for still clears it. It is settled here because the fold is where the rule will be written, once.

## The shape after it

### One fold

A conversation's state is a value: its native session, its runtime state, what it requests, the requests it is waiting on, its newest output, whether a turn is open, its errors and why, when it was last instructed. Today parts of that value are in memory, parts in the registry, and parts are worked out at start-up from other events by code that is separate from the code that decided them. Three passes of hardening went into one such piece of working-out.

After this there is one function, `apply(state, event) state`, with no I/O. The service changes a conversation's state in exactly one way: it appends an event to the log, and when the append is confirmed it applies the event. Start-up builds each conversation's state by applying the same function to the same events. What the service decided and what a restart recovers cannot disagree, because they are one piece of code.

Anything the service decides that is not obvious from an event already in the log gets an event of its own. "Session bound", "error cleared", "request filled in" are recorded as decisions, never inferred later.

**What the fold holds.** These are a conversation's durable fields, and each has an event that sets it:

- its native session;
- its runtime state;
- the permission mode, model and effort it requests;
- whether a fresh session may start, which today is read from every `starting` and `not_started` in the log;
- the requests raised and not yet resolved, each with its id, kind and time, and the highest request id given out, which keeps ids unique across every runtime the conversation has had;
- the seq of its newest output, and the time of the output that put it on the needs-you list;
- whether a turn is open;
- its two errors, each with its cause, kept apart as now: a runtime's failure, and a failed read of history. One field for both would let a failed read overwrite a runtime's failure, and a later good read clear it;
- when it was last instructed.

**What the fold does not hold.** Some of what a conversation shows is observed and never recorded: the MCP list, the background jobs, the text still streaming, and the checks made when a conversation is read, such as whether the checkout is there and whether its environment file is current. They stay in memory beside the folded state. Each is unknown the moment supervision is lost, and the page already says so.

**Where a fold over an old log starts.** Two of the durable fields have no event in a log written before this work. The first slice wrote them to the registry alone:

- the native session of an adopted conversation, until its first launch;
- a model or effort changed while no runtime was up.

A fold over such a log starts from a seed: one event holding the durable fields the registry has, which are the native session, the runtime state and the three requests. The seed is the first checkpoint, the same one decision 4 uses. From stage 3 on, each of the two has an event of its own, so no later log needs a seed.

**A seed is what version 1 itself would conclude,** not the registry's row as it lies. The first slice finishes two things from the log at start-up, and the seeding step applies both before it writes:

- **A permission mode request a crash interrupted.** The request is logged first and the registry written second. Where a log's last event is a `mode_requested` naming another mode than the registry holds, the event's mode is the one requested (`settleModeRequestLocked`). A seed from the raw row would put a conversation back on the mode the operator had just left, which could be the looser one.
- **A session the registry never got.** Where the registry holds no session and a `started` event names one, that is the session (`sessionOf`).

These two rules live in the seeding step and nowhere else. A log is seeded once. After its seed, the fold stands: nothing from the registry overrides it.

Whether a conversation is unread is not a folded field. It is the fold's newest output set against the registry's reading position, worked out when asked.

### One record

The registry is `registry.yaml` as now, holding projects as they are and, for each conversation, only what has no history: its id, project, harness, name, label, mode and reading position. It is written when the operator arranges something or reads, and at no other time.

Two rules say which write counts.

- **The registry entry is what makes a conversation exist.** On the way in, for a new conversation or an adopted one, the log's first record is written and synced before the entry is published, and for an adoption that record holds the session's id. On the way out the entry is removed first and the log after. A log with no entry is ignored. So a crash can never leave an entry whose log is empty, which matters because an empty log is the proof that a fresh session may start.
- **The registry's version says whether it is still whole.** At version 1 it holds every field the first slice wrote, so a build from before this work can open the workspace. At version 2 it does not. One atomic replacement of the registry changes the number, and that replacement is the cutover. The loader already refuses a version it does not know, so a build from before the cutover cannot open a workspace after it.

The log is the only record of everything else.

**Seeds and checkpoints are events of a kind of their own.** A log keeps one kind of line. A seed and a checkpoint are events, each with the next sequence number, of a kind the service keeps for itself. The first draft after round 1 made them records without a number, to keep a read from writing forever; that reason went when the reading position stayed in the registry, and round 2 showed the cost: the reader in every build so far takes the last line's number as the newest, so a line without one would have an older build reuse numbers on its next write.

A page leaves them out the way it already leaves out a native event while those are hidden: not shown, counted as received, so nothing reads as a gap. They add no turn to its history. A page following a conversation acknowledges a checkpoint's number like any other, which is one write of the reading position and nothing more, since no read adds anything to a log.

**A checkpoint is the fold's whole state, written out.** It has no list of its own to fall out of step: whatever the fold holds, the checkpoint has, and a field added to the fold is in the next checkpoint. The one thing it adds is a number that names the fold that wrote it.

One is written at a clean stop, at the end of a turn, and after every so many events. Start-up reads each log backwards to the last checkpoint whose fold it knows and applies what follows. A torn checkpoint, or one from another fold, is passed over for the one before it, and a log with none it can use is read whole.

**What a checkpoint does not carry.** The full read at start-up builds two more things today. Neither is needed to show the workspace, so each is built when it is first wanted, by reading that one log:

- the index of turns that a page's history uses, when a conversation is opened;
- the list of native message ids that stops an import adding a message twice, when an import runs.

### One failure

A write to the record that fails is fatal, as decision 1. There is no state between "recorded" and "not recorded" for any code to handle.

**What the record is.** The registry and the event logs. Nothing else under the data directory is the record, and a failure to write it stays as it is today: a configuration snapshot is best effort, and mega's own log file and the tray's runtime file are diagnostics.

**One latch, and a fence on each runtime.** When a write to the record fails, the service sets a stopped latch before the failing operation lets go of the service's lock. From that moment every command is refused with one error that says mega can no longer record and is stopping, and every page is told once.

The latch alone does not stop a delivery already on its way, because `send` and `Answer` let go of the service's lock before they call the adapter. What stops it is the fence each adapter already has, taken under the runtime's own lock, which is the lock its writes to the harness are made under. For this stop the fence lets nothing through, Deny and Dismiss included, where an ordinary stop lets those two pass. So the order is fixed at the one place a write happens:

- **A prompt or an answer whose write had not begun when the runtime was fenced** is refused and never reaches the harness.
- **One whose write got in first** was delivered. Mega cannot record that, so the next start shows the request as `lost`, as after a crash. Nothing is delivered a second time.

Then each runtime is ended the way a release ends it, by a path that writes nothing, and mega exits.

**What the next start does.** It needs no rule of its own. It uses three that exist today for a crash:

- a conversation last recorded as running is unsupervised;
- a request with no recorded resolution is `lost`, and is never answered again;
- a launch with a `starting` and nothing after it is a session unknown: prompts are refused and the page points at "Adopt a session".

**A session is bound as soon as it is known.** A `starting` event is recorded before a harness is started, as now. The session id is recorded the moment the launch returns it, in the place the registry is written today, as an event of its own. It does not wait for `started`, because a new session can begin its opening prompt before the harness reports that.

### Turns the harness declares

A turn opens in two ways, and both are facts:

- **Mega sent a prompt.** The turn opens when the prompt is accepted, as now. That is what refuses a second prompt and keeps a shutdown from treating the conversation as at rest before the harness has said anything. A prompt the harness refuses closes it.
- **The adapter says the harness started one.** This is a turn nobody prompted, such as Claude waking when a background job finishes.

A turn closes when the harness says it ended, as now, or when its runtime ends. For pi the end is `agent_settled`, which holds through pi's own retries.

What goes is the third way a turn opens today: the service sees any message from a runtime while no turn is open, and takes it for a start. A message outside any turn is recorded and opens nothing. A turn with no activity is shown as quiet, and an interrupt that gets no answer says so. This is the card `a-turn-is-inferred-from-a-message`.

No state is added between idle and running. An accepted prompt is a turn in flight from the moment it is accepted.

### One kit under the adapters

Four things are written three times today and were each fixed three times: checking that a stream or a file belongs to the session asked for, refusing what would start work once a stop has begun, finding what a runtime left running, and ordering an answer against the harness's own resolution of the same request. They move into `internal/harness` as one implementation, built on what is already shared there: `SameSession`, `CarriesWork` and the process inspection in `procs.go`. An adapter is then what it was meant to be: a translation of one harness's protocol, with no policy in it.

The kit decides the order and the rule. Each adapter supplies the evidence its harness has, and the harnesses differ in ways the kit has to allow:

- **Left-over processes.** Claude and pi inspect a process group. Codex has to note the descendants before their parent exits, because its sandbox starts a session of its own.
- **Whether an answer landed.** Claude's shim returns a verdict. Codex acknowledges no reply to a server request. A pi dialog can expire on a timer of its own.

  The kit's rule for an answer is the one each adapter follows today: it refuses an answer once it has read the harness's own settlement of that request, and it writes an answer at most once, under the runtime's lock. For a harness that confirms nothing, `answered` records that mega delivered the operator's answer. It does not say the harness acted on it; what the harness then did is in the conversation's next events.

  One limit stays, recorded and not engineered around. Codex may have settled a request while its notice is still on the pipe, and an answer written in that moment is recorded as `answered` though Codex ignored it. No fourth outcome is added for this beside `answered`, `aborted` and `lost`. Michael declined one on 2026-11-19, on the review's finding `R2-06`: it would be a new state, with its own words on the page and its own handling at a restart, for a window one notice wide.
- **A question answered by a prompt.** A question Codex's adapter makes up from a finished turn can be answered by starting a new turn.

### One state for the page

The page has one value for where it stands, with two parts. It is not one list, because the parts are true at the same time: a page can be catching up while it still holds a notice from the service it was talking to before a restart.

- **What the page knows of the workspace.** The five values `ui/src/stream.ts` already has: nothing yet, while connecting; unreachable, when nothing answers; unreadable, when the service answers and the read fails; current; and stale, when it shows only what it last read.
- **What the service last said about itself, if anything.** Shutdown asked, with the time the gate was armed, which is what ties the notice to one gate; shut down; or stopped because it can no longer record.

One function works the value out from the stream's state, the outcome of the last read and the last notice. A notice stands until a current read shows which gate exists now, as today. The footer, the rail, the banner, the held controls and the read all take the value from that function and from nowhere else.

Most of this exists after the hardening passes, in pieces: three types in `stream.ts` and the notice's handling in `App.tsx`. The stage joins them and does not add a state.

## Where the work happens

Michael, 2026-11-19: "we're also going to want to make sure we do this work on a branch so that we can test it locally and ensure it's solid before merging it to `main` and to the two new downstream users."

- **A branch.** All seven stages land on one branch, `simpler-core`, which Michael creates. He commits each stage to it after its gate, as he commits everything. Agents work in the working tree of that branch and commit nothing.
- **`main` stays as it is.** It is what the external operators build from. It takes nothing from this work until the merge. A fix that cannot wait goes to `main` and Michael merges `main` into the branch.
- **Michael's own machines run the branch,** one at first. That is the test. From stage 4 on, a machine that has run the branch has a version 2 workspace, and going back to `main` on it means putting `registry.v1.yaml` back.
- **The merge is a gate of its own, and Michael's.** Not before stage 4's migration has run on each of his three machines and he has worked on each afterwards. The external operators get the migration only then, with the changelog telling them what it does and where their old registry is kept.

## Stages

Each stage is gated by terminus to `clean` on its working tree, then by Michael. Each ends with its journal entry and `docs/current/` caught up. Each must leave `make test`, all of `make checks` and `make e2e` passing, and `internal/service` no larger than it found it. It is 25 files and about 6,600 lines today, of about 19,400 under `internal`, tests aside.

**Stage 1: unread, made plain.** Decision 2. Remove the buffering of a turn's output until its end, the counting of output flagged as after its turn, the scan a resume makes for what is new, its bounds, and the detection of an open turn in Claude's and pi's histories where it served only that scan. The newest output is taken from the log, as the newest event already is, and the registry's unread marker is no longer read. It is still written at a turn's end until stage 4 removes the field, only so that an older build opened on the same workspace shows what it would have; a failure to write it is logged and nothing depends on it. The flag itself stays through this stage, because it is also what stops late output from opening a turn; it goes in stage 2 with the inference it guards against. The canon's `bookkeeping-never-creates-attention` is amended in the same stage. Gate: the circuit from the spec, by the checks; by hand, an unclean restart of a scratch mega with a turn in flight, resumed, showing unread exactly once; a conversation with an unread reply, released and resumed without being opened, still on the needs-you list; a session just adopted showing as unread until it is opened, and not after.

**Stage 2: turns declared.** It starts with a probe of each harness for the marker of a turn it starts by itself, recorded with its version under `probes/`. Codex and pi have one. Claude may not, and if it does not, its adapter decides from the first assistant output, where Claude's own messages outside a turn are known and left out. Each adapter then reports a turn's start. The service's inference goes, and the after-its-turn flag with it. The prompt path in `send` is not changed. Gate: the round trips for all three harnesses, including the model switch on an idle runtime, which is the case that started this; a message outside a turn opens none; a second prompt and a shutdown asked in the moment after a prompt is accepted are both held, as now; a turn the harness starts by itself shows as running and holds a shutdown.

**Stage 3: one fold.** Introduce the state value and `apply`. Every state change in the service becomes append, then apply. Adopting a session, and changing a model or effort while idle, each gain the event they lack. Seeds are the first events of the service's own kind, so the page learns to leave that kind out here. Start-up builds state by the same function, still reading whole logs. A log with no seed gets one at its first open, written from the registry's fields. Remove each piece of start-up code that works a fact out from other events.

The service runs on the fold from this stage. The registry stays at version 1 and is written after each event as before, so a build from `main` can still open the workspace. At each start the fold's durable fields are compared with the registry's, for the fields the registry has: the native session, the runtime state and the three requests. A difference is logged with both values and nothing is changed: the fold stands, because every change is logged first, so after a crash the log is the newer of the two. With a correct fold there are none beyond that case, so each one logged is evidence to read. This makes the stage a soak: it runs on Michael's machines with the registry as the cross-check.

Going back and forth between this stage's build and `main` is not engineered beyond that. The one thing known to be lost is a model or effort changed while idle under `main` after a log was seeded: the fold never sees it.

Gate: a comparison against the build as it stood at the end of stage 2, kept for the purpose, because stages 1 and 2 will already have changed what today's code would say. Both builds start on separate copies of the same registry and logs, taken from Michael's real workspace, and a one-off tool compares each conversation's durable fields, with every difference explained. The copies include five made cases: a session adopted and never launched; a model changed while idle; a conversation read and then left, whose reading position the registry still holds and the fold never touches; a log ending in a `mode_requested` the registry never got; and a session named in a `started` event and absent from the registry. A page following a conversation settles after one read, with nothing more written. The build `main` holds opens a workspace this stage has written whose logs end in a seed, shows it with nothing odd on the page, and records a prompt and its events; the log is reopened and its numbers run on without a repeat. And after a week of Michael's own use on at least one machine, the differences logged at start-up are read and each is explained.

**Stage 4: one record, one failure.** Decisions 1 and 3. The registry loses its changing fields. A failed write stops mega. The uncertain states of the registry and the event log go, with every rule for a record that failed after another was made. The registry goes to version 2, and a start on version 1 migrates it, in this order:

1. Every conversation's log that has no seed is given one, by the rule in "Where a fold over an old log starts", and synced. A workspace that ran stage 3 has them already. One coming straight from `main`, as an external operator's will, has none.
2. The version 1 registry is copied beside itself as `registry.v1.yaml`, and synced. An existing copy is replaced.
3. The registry is replaced, atomically, by one at version 2 without the changing fields. This is the commit.

A migration interrupted before step 3 has changed nothing that counts: the registry still says version 1, and the next start does the three steps again, leaving the logs already seeded as they are. Mega never removes `registry.v1.yaml`. The start that migrates says so once in its log and on the page, with the copy's path.

Creation and adoption change their order, and removal keeps its own, as "One record" says.

The page learns one new notice in this stage: mega has stopped because it can no longer record. It is sent as the last frame of the stream, the way a finished shutdown is announced today, and shown the way that is shown, with its own words. Stage 7 later gives it its place in the page's one value. Gate: the migration run on a copy of Michael's workspace, interrupted before, between and after each of its three steps and rerun; the same on a copy that never ran stage 3; the last version 1 build refusing the migrated workspace; a crash between an adoption's first log record and its registry entry leaving nothing an operator can see, and one after the entry resuming the adopted session by its id; an injected write failure stops mega with the page saying so, and the next start shows the running conversations as unsupervised. The failure is also injected in the two windows where native work is already under way: with an answer reserved and its write not yet begun, where the answer never reaches the harness; with an answer already written and its outcome not recorded, where the next start shows the request as lost and nothing is delivered again; with a Deny and a Dismiss after the fence, both refused; and between a new session's creation and the record of its id, where the next start shows the session as unknown and starts no second one. A failed snapshot stops nothing.

**Stage 5: the end of the log.** Decision 4. Checkpoints as "One record" describes them, a start-up that reads back to the last one it knows, and the turn index and the id list built when first wanted. Gate: start-up time on a copy of Michael's workspace, before and after; for every log in that copy, a checkpoint taken at every position and the log read from there, giving at every position the same state as a full read, which covers a checkpoint that falls while a request is pending, or between a failed read and a good one, and the cases nobody has listed; a page's replay across a checkpoint, and its history, the same as from a full read; an import that overlaps the part of a log before the checkpoint adding nothing twice; a log whose last checkpoint is torn, and one whose last checkpoint names another fold, each opening to the same state by the longer read.

**Stage 6: one kit.** The four shared mechanisms move under the adapters. Gate: the round trips for all three harnesses, the version runners under `probes/versions/` for Claude and Codex, and pi's RPC probe as their counterpart; each of the differences "One kit under the adapters" names, shown still to hold, by a test of the adapter it belongs to; each adapter's file holds no policy the other two also hold.

**Stage 7: one state for the page.** Gate: the side, shutdown and panel suites; the reconnect and shutdown cases from the hardening passes each pass against the one value; a page that reconnects to a replacement service already armed for shutdown shows that gate and not the one it remembered; a workspace read that fails on a connected stream, on a page that already shows a workspace from before, reads "Read failed, retrying" and not only "Catching up", holds the controls and acknowledges no read.

Then the full-tree review resumes, on `harden-requests-stops-and-restarts`, and its count per pass is compared with five.

## What must not change

- A session is resumed by its exact id, and a new one is started only on proof that none exists.
- Permissions are never weakened, no trust is set for the operator, and nothing is answered for him.
- A state mega cannot vouch for is never shown as certain. A simpler mega may say "look at this" more often. It may not say something untrue.
- A pending request is answered once, by the operator, and never after it stopped being pending.
- Everything an operator sees when nothing has gone wrong. Each stage's checks are the proof.

## Risks and what to watch

- **It is the core.** Every stage touches the files everything else leans on. The checks, about 550 steps in nine suites, and the round trips are what make this safe to attempt. A stage that needs a check weakened is a stage that has changed behaviour, and stops.
- **Michael's data.** Stage 4 changes the format of a workspace he works in every day, on three machines, and that at least two external operators now have too. The migration is proven on copies first, is safe to interrupt, and keeps the old registry beside the new one. The changelog says plainly that a workspace opened by the new mega cannot be opened by an older one.

  Going back is kept simple on purpose. Michael, 2026-11-19: "we don't need to over-engineer the rollback part of migration. i would assume that worst case scenario an agent could be removed from the workspace and re-adopted fresh and the conversation would remain intact." So there are two ways back and neither has a gate of its own. The first: stop mega, put `registry.v1.yaml` back as `registry.yaml`, start the older build; requested models and runtime states then read as they stood on the day of the migration. The second, for a conversation that does not come back right: remove it from the workspace and adopt its session again. Its history is the harness's own and comes back whole. What only mega's log held does not: how each request was answered, and what each turn was asked to run with.
- **The comparison in stage 3 is the keystone.** If the new fold and the old code disagree about a real log, one of them is wrong about something that happened, and that has to be understood before going on.
- **Stopping on a failed write is a change in character.** Mega stops being a service that limps on. It is the right trade for one operator on his own machine, where a full disk is rare and a restart is cheap. It should be revisited if mega ever runs where neither is true.
- **Size is the measure, and it can be gamed.** Moving code out of `internal/service` is not removing it. Each stage reports what was deleted outright.

## What this leaves alone

The harness adapters' protocols, the project environments, the page's layout and controls, snapshots, path search, and everything on the roadmap's horizon. The extraction of the orchestration framework is not attempted here; stage 6 only makes it easier.
