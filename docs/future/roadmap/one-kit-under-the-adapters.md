---
title: one kit under the adapters
state: researching
created: 2026-11-19
tags: [story]
---

Four things each adapter does for itself today are written once, in `internal/harness`, and no rule changes. An adapter is then a translation of one harness's protocol.

Ninth in the sequence of `simplify-the-core`.

## why

Four things are written three times and were each fixed three times in the hardening passes: checking that a stream or a file belongs to the session asked for, refusing what would start work once a stop has begun, finding what a runtime left running, and ordering an answer against the harness's own settlement of the same request. Some of it is shared already: `SameSession`, `CarriesWork` and `procs.go`.

It is also the first step of `the-orchestration-framework-is-extracted`, arrived at from the other side.

## the failure that becomes impossible

A fix to one adapter's stop rule, or its session check, can no longer be missing from the other two.

## what changes

The four mechanisms move. The kit holds the order and the decision; each adapter hands it the evidence its harness has.

## what does not change

Every rule. What each adapter does today is the specification, and the three differ in what evidence their harness gives: how left-over processes are found, how each learns that a request is settled, and what each takes as proof that an answer landed. Those differences stay.

One limit stays too, recorded and not engineered around: Codex may have settled a request while its notice is still on the pipe, and an answer written in that moment is recorded as `answered` though Codex ignored it. Michael declined a fourth outcome for it on 2026-11-19.

## how it is shown

- every adapter test from before passes unchanged;
- the round trips for all three harnesses, the version runners under `probes/versions/`, and pi's RPC probe;
- one test added, for a case no test holds today: an answer to a pi dialog after its deadline has passed and before its timer has removed it is refused;
- each of the four mechanisms has one shared implementation, and each adapter supplies only its harness's evidence to it.

## order

After `a-codex-question-is-a-message`, which removes one of the differences: an answer that starts a turn. Smaller on the pi side if `megas-history-may-have-a-gap` is taken first. Lands on `main`.
