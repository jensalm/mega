---
title: a codex question is a message
state: researching
created: 2026-11-19
tags: [story]
---

An answer to a question Codex asks in a message never starts a turn. While the turn that asked is still running, the question keeps its card and the answer is steered into that turn. Once the turn has ended, the question is a message like any other and is answered in the reply box.

First in the sequence of `simplify-the-core`. Its size is waiting on Michael: see "two sizes".

## why

Codex has two ways to ask. One is a protocol request with an id that takes a response. The other is an agent message carrying `questions[]`, which is what Codex 0.166.0 sends in default mode. For the second, mega records a `question` request and then owns its whole life, because Codex sends no settlement for it (`internal/harness/codex/runtime.go`, `question`, `Answer`, `answerAsTurn`):

- an answer while the turn is active is steered into it with `turn/steer`;
- an answer after the turn completed starts the next turn;
- a stop between sending that start and Codex's reply has to interrupt the turn mega just made;
- an interrupt, a release, a failed turn and an exit each abort it.

Much of the hardening work on answers after a stop came from the second and third of these. They are the reason an answer can carry work across a stop at all.

Found by an independent review for simplification on 2026-11-19, which proposed the larger of the two sizes below. Michael the same day: "this sounds like a smart simplification."

## what the research says, found while writing its work order

The planning agent told Michael that Codex usually asks and then ends its turn. That was not checked, and the capability report says otherwise for the case it observed (`docs/research/harness-capabilities.md`, Codex 0.166.0): the model emits the message with its questions and then sleeps in 60 second steps, and "the turn did not end on its own". The round trip in `internal/e2e/codex_test.go` answers such a question while its turn is still running.

That matters because mega refuses a prompt while a turn is in flight, and a conversation with a turn in flight reads as running, not as wanting anything. With no card, a Codex that is asking and sleeping would look like a Codex that is working, and the reply box could not answer it until the operator interrupted.

## two sizes

- **The smaller one, recommended.** Keep the card while the asking turn is running, and keep the steer. Remove the answer that starts the next turn. When the turn ends for any reason the question stops being a request; its message stays in the conversation and the reply box answers it. The operator is still told that Codex is waiting, and the path that carried work across a stop is gone.
- **The larger one, as the review proposed.** No card at all. The question is read in the message and answered in the reply box once the turn ends. It removes the made-up request altogether, and it leaves a question asked by a sleeping turn unseen until the operator looks. It becomes sound once a prompt can be sent into a running turn, which is `queued-operator-messages`.

## the failure that becomes impossible

An answer to a Codex question can no longer start a turn after a stop. In the smaller size this is because no answer starts a turn at all.

## what changes, in the smaller size

- **Removed from the adapter:** `answerAsTurn`, the question's memory of a turn that completed, the count of stops an answer is checked against, and the fall-back from a failed steer to a new turn. A steer that fails means the turn has ended, and the answer is refused as not pending.
- **A question ends with its turn.** A turn that completes aborts its question, as an interrupted or failed one does today.
- **The message says what was asked.** Today the assistant event holds the message's text and the questions travel only in the request. The questions and their choices are added to the recorded message as plain text, so the conversation still reads rightly after the card is gone.
- **`docs/current/mega.md`,** "How the two kinds are mapped", and the HTTP reference.

## what it costs

A question whose turn has ended can no longer be answered from its card. It is answered in the reply box, which starts the next turn, as any reply does.

## what does not change

Every request Codex makes through its protocol. `item/tool/requestUserInput` is still handled if it arrives. The card and the steer while the asking turn runs. The rule that an answer which would carry work is refused once a stop has begun.

## how it is shown

- the Codex round trip as it is today: a question asked in default mode raises its card, and an answer is steered into the running turn;
- a turn that completes with its question unanswered: the request is aborted, the message shows the question and its choices, the conversation is unread, and a reply in the box starts the next turn;
- an answer written just as its turn ends is refused as not pending, and starts nothing;
- the adapter's tests for an answer that starts a turn are deleted with the code, and the others pass unchanged.

## order

Depends on nothing. It comes before `one-kit-under-the-adapters`, because it removes one of the differences that kit would otherwise have to allow. Lands on `main`.
