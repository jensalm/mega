# pi probe

Minimal, dependency-free Node probe that drives the installed `pi` (`@earendil-works/pi-coding-agent` 1.1.0) in `--mode rpc` over stdio JSONL with a persisted session.

## Requirements

- `pi` on `PATH` (tested: 1.1.0 at `/home/michael/local/node/bin/pi`, Node 24.12.0).
- A provider the installed pi can reach. The operator's default (`ollama/qwen3.8-27b`) was not running during the investigation, so the runs below pass `--model openai-codex/gpt-6-astra`, which uses the operator's existing Codex login.
- The fixture checkout `/home/michael/Repos/q/research/roundhouse-fixtures/pi-rt1` (created by hand: `fixture.txt`, `AGENTS.md` with `ROUNDHOUSE-MARKER-7f3a`, and `.pi/extensions/roundhouse.ts`). The extension adds a `tool_call` approval dialog for `bash` commands containing `sleep`, an `ask_user` tool backed by `ctx.ui.select`, and a `/rh-confirm` command. Loading it requires project trust; the probe passes `--approve` (a one-run CLI override) when given `--approve-project`.

No global pi configuration is changed. Logs are written under `logs/` (gitignored).

## Usage

```sh
# new persisted session
node probe.mjs run --cwd ../../../roundhouse-fixtures/pi-rt1 --session-id $(uuidgen) \
  --approve-project --model openai-codex/gpt-6-astra \
  --turns scenarios/rt1.json --log logs/rt1-managed.jsonl \
  --expect '[Cc]olou?r=>Blue|^Approve bash command: sleep 3$=>Allow' --answer-delay-ms 4000

# reopen the exact session (id or path); model is restored from the session file
node probe.mjs run --cwd ../../../roundhouse-fixtures/pi-rt1 --session <id> --approve-project \
  --turns scenarios/rt1-rpc-return.json --log logs/rt1-rpc-return.jsonl --expect '^Approve bash command: sleep 2$=>Allow'

# print a session's messages
node probe.mjs messages --cwd ../../../roundhouse-fixtures/pi-rt1 --session <id> --approve-project
```

Flags:

- `--expect 'TITLE_REGEX=>ANSWER|…'`: the only dialogs the probe answers, each matched by its title, with an answer the dialog must offer: a `select`'s option, `Yes` or `No` for a `confirm`, text for an `input` or `editor`. Any other dialog is left unanswered and stops the run, recorded as `unexpected_dialog`; the probe never picks an answer of its own (2026-11-16, replacing `--answer`, which fell back to the first option).
- `--answer-delay-ms N`: delay before answering, to show that the run waits.
- `--interrupt-on-dialog`: send `abort` while a dialog is pending; `--stale-after-cancel` then answers the dialog 1.5 s after the abort response or its 30 s timeout.
- `--never-answer`: leave dialogs unanswered.
- `--interrupt-after-ms N [--interrupt-turn I]`: send `abort` during an active turn.
- `--kill-owner-after-ms N`: SIGKILL the probe itself while a turn runs.
- `--timeout-ms`, `--turn-timeout-ms`: probe deadlines.

The probe treats the `prompt` command response (`data.disposition`) as "accepted" and the `agent_settled` event as "settled", per the installed `docs/rpc.md`. Frames are split on LF bytes only. `message_update` and `text_delta` frames are logged trimmed.

## Relay (owner) mode, follow-up probes

`--relay SOCK [--hold] [--no-prompt]` turns the probe into the local owner for the reconnect tests (see `../common/README.md`). Every RPC frame is forwarded to connected relay clients, each `extension_ui_request` dialog becomes a relay request that waits for a client answer, and clients may send prompts, `interrupt` (mapped to `abort`) and `close`. Answer shapes: `{"value":"Allow"}` for `select`/`input`/`editor`, `{"confirmed":true}` for `confirm`, `{"cancelled":true}` to dismiss any dialog (the extension receives `undefined`/`false`). A prompt starting with `RAW:` is sent to pi verbatim as a JSON frame (used once to test a late `extension_ui_response`). Keep the socket path short (`/run/user/1000/roundhouse-probe/*.sock`); a path under the long scratchpad directory fails with `listen EINVAL`.

```sh
node probe.mjs run --cwd ../../../roundhouse-fixtures/pi-rt1 --session-id $(uuidgen) --approve-project \
  --model openai-codex/gpt-6-astra --relay /run/user/1000/roundhouse-probe/pi.sock --hold --no-prompt --log logs/x.jsonl
node ../common/client.mjs --socket /run/user/1000/roundhouse-probe/pi.sock --prompt "..." --answer-json '{"value":"Allow"}' --watch-ms 20000
```

Fixture extension addition: a bash command containing the word `timed` (for example `sleep 1 # timed`) gets an approval `select` with `timeout: 8000`; other `sleep` commands wait indefinitely. `fu-suspend.sh` runs the suspension cases detached from the calling shell. Follow-up scenarios are `scenarios/fu-*.json`; session ids are appended to `logs/followup-session-ids.txt`. Results: `NOTES-followup.md`.

## Rerunning against a new build

The probe was written against pi 1.1.0 and rerun against 1.5.2 on 2026-11-16 (capability report, "Follow-up (2026-11-16): pi 1.5.2"). These rerun it on the operator's own default model, as pi's settings name it, never passing `--approve`: the fixture must already be trusted in pi's own terminal.

- `rerun.sh [p1 p4 r1 r2 r3 r4 p7]` reruns the core scenarios into `logs/v<version>/`, with the session ids in `session-ids.txt` there.
- `inspect.mjs <checkout>` prints, with no model turn: the model and thinking level `get_state` reports, the model list, the thinking levels the model allows, the commands, and the MCP servers' status, which pi gives only as the text answering `/mcp`.
- `scenarios/outside.json` (2026-11-16) asks for a file outside the checkout four times, with `~` and absolute, and with a space in its name. Its paths name the scratch folder `roundhouse-fixtures/outside-spike/`, made for the run and removed after: `node probe.mjs run --cwd <pi-rt1> --session-id $(uuidgen) --turns scenarios/outside.json --log logs/v<version>/outside.jsonl`. No dialog is expected, so none is answered.
- `close-with-dialog.mjs . close|abort-then-close`, run in the fixture, leaves the extension's approval dialog pending and measures how pi ends when its input closes, then after one and two SIGTERMs.
