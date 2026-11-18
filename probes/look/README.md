# The look's cost

What the adopt dialog's look at a session costs on a long one, for each harness. The look is read when a row is picked, through the same read an import makes: Claude Code's shim reads the whole transcript (`getSessionMessages`), Codex's app-server reads the whole thread (`thread/read`), and pi's adapter reads the whole session file. So it grows with the session.

`run.sh [TURNS]`, from the repository's root, writes one synthetic session per harness of TURNS turns (2000 by default) into a scratch directory. Each turn is a prompt, a tool call, its result and a reply of filler text. It times each adapter's `Look` three times, through `looktime`, and removes the scratch. Nothing touches a real store, and nothing logs in or calls a model: Claude Code's runs this checkout's shim with `CLAUDE_CONFIG_DIR` in scratch, Codex's runs the `codex` on the `PATH` with `CODEX_HOME` in scratch, and pi's reads its file with `PI_CODING_AGENT_DIR` in scratch. It prints times and sizes, never the text.

- `gen.mjs SCRATCH CWD TURNS` writes the three sessions and prints their ids. Each harness's format is written by a writer beside its fake, so the protocols stay with the adapters and the shim: `shim/claude/fake/transcript.mjs`, `internal/harness/codex/fake/rollout.mjs` and `internal/harness/pi/fake/session.mjs`.
- `looktime HARNESS CWD SESSION` times one look.

## 2026-11-18, 2000 turns

Claude Code 2.1.325's shim with Agent SDK 0.3.318, codex-cli 0.166.0, and pi's adapter for 1.5.2. The sessions were 5.8 MB, 5.5 MB and 4.5 MB.

| Harness | Look | Of which |
| --- | --- | --- |
| Claude Code | about 0.33 s | 0.2 s for the shim's read, the rest mapping 8,000 messages |
| Codex | 0.22 to 0.27 s | a short-lived app-server, started for the read |
| pi | about 0.1 s | the file read and its chain followed |
