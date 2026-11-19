# An answer never starts a turn: work order

Work order for the card `docs/future/roadmap/a-codex-question-is-a-message.md`, the first piece of `simplify-the-core`, in the smaller of its two sizes, which Michael chose on 2026-11-19. It is one batch in one adapter, on `main`.

## How to read it

What this does not say changes, does not. The code and its tests as they stand are the specification of everything this is silent on. Where a sentence here describes today's behaviour, it is a pointer to the code, and where the two differ the code is right.

## The failure that becomes impossible

An answer to a Codex question can no longer start a turn after a stop, because no answer starts a turn at all.

## What it is about

Codex 0.166.0 in default mode asks by sending an agent message that carries `questions[]`. Codex sends no settlement for such a question, so the adapter keeps its life itself (`internal/harness/codex/runtime.go`, the `question` type and the `questions` map). An answer goes one of two ways today:

- **into the turn that asked,** with `turn/steer`, while that turn is running;
- **as the start of the next turn,** through `answerAsTurn`, when the asking turn completed with the question unanswered.

The second way is the one this removes. It is why a question outlives its turn, why an answer is checked against a count of stops, and why a stop that lands just after has to interrupt a turn mega made.

The first way stays. The capability report records that a turn which asks this way does not end by itself: the model sleeps and waits (`docs/research/harness-capabilities.md`, Codex 0.166.0). The card is how the operator learns that Codex is waiting.

## What changes

All of it is in `internal/harness/codex`.

1. **A question ends with its turn.** When a turn ends, in any status, its questions are aborted. Today a turn that completes keeps them. A question that arrives after its turn has ended is not raised, since nothing could answer it; its message says what was asked.
2. **An answer is steered or refused.** `Answer` steers an answer into the question's turn when that turn is the active one. In every other case, and when the steer fails, it returns `harness.ErrNotPending`. It never sends `turn/start`.
3. **Removed outright:** `answerAsTurn`, the test hook `afterStopCheck`, the question's `completed` field, the runtime's `stops` count, and whatever else is left with no reader once those are gone. The report names each thing removed.
4. **The message says what was asked.** The recorded assistant message holds only the message's text today, and the questions travel only in the request. `mapItem` adds the questions and their choices to the message's text, as plain lines: each question's title, then its options as a list. It is the one place, so a message read back from Codex's history later reads the same.

## What does not change

- Every request Codex makes through its own protocol, and `item/tool/requestUserInput` if it arrives.
- The card for a question, and its payload, while its turn runs.
- When a question is not raised at all, and the refusal of an answer that would carry work once a stop has begun.
- `Prompt`, including how it treats a turn that completes while its start is still being answered.
- The service, the page, and the other two adapters. Nothing outside `internal/harness/codex` is edited but tests, documents and the changelog.

## Tests

- **Deleted with the code they test:** `TestAsyncQuestionOutlivesACompletedTurnAndStartsTheNext`, `TestAnswerInFlightNeverStartsATurnAcrossAnInterrupt`, `TestNoAnswerStartsATurnAfterTheStopIsFenced` and `TestStopDuringAnswerTurnStartInterruptsIt`. If another test's whole subject is an answer that starts a turn, it goes too and the report says which.
- **Added, each seen failing first:**
  - a turn that completes with its question unanswered aborts the question;
  - a message with questions that arrives after its turn has ended raises no request;
  - an answer to a question whose turn has ended is refused as not pending, and nothing is sent to Codex;
  - an answer whose steer fails is refused as not pending, and nothing else is sent;
  - a message with questions is recorded with the questions and their options in its text, and one without is recorded as before.
- **Every other test in the package passes unchanged.**

## Gate

- `make test`, all of `make checks`, and `make e2e` with the Codex fixture. The round trip's question step is not edited and passes as it is: a question asked in default mode raises its card, and the answer is steered into the running turn.
- The report names the Codex version the round trip ran on and anything skipped.
- `internal/harness/codex` is smaller, tests aside, and the report says by how much.
- Terminus to `clean` on the working tree, against the canon. Then Michael.

## Documents

- `docs/current/mega.md`: the bullet on a message carrying `questions[]`, under how Codex's items are mapped.
- `CHANGELOG.md`: one `CHANGE:` entry under `## Unreleased`, saying what an operator will notice: a Codex question whose turn has ended is answered in the reply box.
- `docs/journal/`: the day's entry.
- The spec already says it, in "A simpler core": an answer never starts a turn.

## If it does not fit

If making this work needs a change outside the Codex adapter, a new state, or a rule for a case this does not name, stop and ask. That would mean the piece is not the small one it is meant to be.
