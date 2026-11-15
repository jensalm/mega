---
title: harness version tolerance
state: building
created: 2026-11-13
tags: [story, spike]
---

Mega works with the harness versions people actually have, within a stated range, and says plainly when it is outside one. Today it names one Claude Code build and one Codex build, and nobody has recorded what happens with any other.

## why

The README asks for Claude Code 2.1.318 at `~/.local/share/claude/versions/2.1.318` and codex-cli 0.166.0. An external operator's Claude Code updates itself, and their Codex is whatever they installed last. So does Michael's, as it turned out: see "what is already on the machine" below. The stage 2 gate crossed one Claude version without trouble, which is the only evidence there is. Without this, the commonest first experience of mega is a failure with no explanation.

## spike

Recorded in the capability report with versions:

- A Claude Code newer than the shim's pinned Agent SDK: what works, what breaks, and whether the SDK or the shim notices.
- Whether the shim can use the Claude Code on `PATH` instead of a versioned directory, and what that costs in certainty.
- A Codex newer and older than 0.166.0 against the app-server calls the adapter makes.
- How each harness reports its version, and how early mega can know it.

## shape to explore

- A tested range per harness, stated in the README and in `mega version`, not one exact build.
- At start-up and at launch, mega compares what it finds with the range. Inside it, nothing is said. Outside it, the launch goes ahead with a plain notice of what was tested, unless the spike finds a version that cannot work, which is refused with the reason.
- Never an upgrade or a downgrade by mega. It reports; the operator decides.

## notes from the clean-machine rehearsal, 2026-11-13

- **No fresh harness was installed.** In the agent's pass, running the downloaded Claude Code installer in the scratch home was refused by the session's permission classifier. Codex through npm into a scratch prefix was not tried, for the same reason. So the rehearsal saw no version other than the pinned ones, only none at all. Michael's WSL2 pass, or an agent pass he approves installs for, is where a fresh version is first seen.
- **A Claude other than 2.1.318 is, to mega, a Claude that is not there, and it says so badly.** The default `claude.executable` is the versioned path. A Claude Code installed today lands under another version's directory, and mega's check passes an absolute path unchecked. The first launch fails with "internal error" on the page and the SDK's "Claude Code native binary not found at …/versions/2.1.318" in the log (`a-missing-claude-is-named`). Until a range exists, the refusal could at least name the version mega expected and the path it tried.
- **An npm-installed `claude` is refused by name**, in code though not observed: a `claude` that is a script is refused with "set claude.executable to a native build, such as ~/.local/share/claude/versions/<version>". External operators who installed Claude Code through npm get that.
- **Codex is taken from `PATH` at any version**, with no check before the launch. What an older or newer one does is the open question above.
- **Nothing names the expected or found versions before a launch.** `mega version` prints go, target and the log path. `/api/version` adds `claude_sdk: 0.3.318`. Neither names Claude Code 2.1.318 or codex-cli 0.166.0, nor what was found.

## carried from the first-run batch

Noted by the planning agent on 2026-11-14. Since that batch a missing Claude Code is named: "'…/versions/2.1.318' does not exist; install Claude Code, or set claude.executable…". That is true and still misleading for the commonest newcomer, who has Claude Code installed at another version. The message tells them to install what they have. This card's answer has to reach that message: when another version is present beside the pinned one, say which was found and what mega was tested against, and do not say "install".

## what is already on the machine

Read on 2026-11-14, after Michael said his harnesses update themselves by their own mechanisms. An earlier line of this card said he held them still; that was wrong.

- **Claude Code.** `~/.local/share/claude/versions/` holds five builds: 2.1.318, 2.1.319, 2.1.321, 2.1.322 and 2.1.325. His `claude` on `PATH` is a link to 2.1.325, installed the same day. Mega has no config file, so it uses its default, the 2.1.318 path. He has been running mega's Claude conversations on a build seven versions behind the one in his terminal, both against the same configuration, all week.
- **So the Claude half of the spike needs no download and no scratch login.** The newer builds are his own, already installed by his own updater and already used against his real configuration.
- **A risk on his own machine.** 2.1.318 is the oldest of the five. If Claude's updater prunes old builds, mega stops launching Claude the day that file goes, with "…/versions/2.1.318 does not exist". How many builds the updater keeps has not been observed.
- **Codex.** `codex` on `PATH` is 0.166.0, the pinned version, from an npm prefix. Mega takes whatever is on `PATH`, so when it updates mega follows with no check. A newer Codex for the spike does need fetching into scratch.

## approval

Michael, 2026-11-14, in two parts. First: an agent may download Claude Code and Codex builds into a scratch directory for this spike, and he wants to be kept in the loop as it proceeds. Then, asked whether foreign builds should be kept away from his real configuration: "builds do run under my own logins... but my harnesses are automatically updated by their individual update mechanisms. making configuration backups sounds smart regardless of how we handle harness versioning." So:

- The builds run as him, against his real configuration and logins. No scratch logins.
- Before any build other than the one mega uses today is run, his harness configuration is backed up. Configuration only: never a credential or token file, and never transcripts. The list of files is shown to him first.
- Before a download: what is fetched, from where, and into which scratch directory, then a stop for his go-ahead. Only Codex needs one.
- His installs and his `claude` and `codex` on `PATH` are never changed by the spike. Nothing is installed outside the scratch directory.
- A running log in the journal as each build is tried, and a stop between the Claude half and the Codex half.

## the spike, 2026-11-14

The spike half is done. It is recorded in `docs/research/harness-capabilities.md`, "Follow-up (2026-11-14): harness versions other than the pinned ones", with the running log in the journal of the same day. The configuration backup is at `~/Sandbox/harness-config-backup-2026-11-14/`, with the 2.1.318 build. The scratch under `/tmp/mega-harness-spike/` was removed later the same day, when Michael asked for `/tmp` to be trimmed: the fetched Codex builds, whose SHA-256 sums are in the journal, and with them the spike's driver scripts, which were not in the repository. The journal lists every call the drivers made. The build that follows recreates them as a rerunnable probe under `probes/`, where they should have been.

### what the evidence supports

- **Claude Code: 2.1.318 to 2.1.325, with Agent SDK 0.3.318.** Five builds, every call mega makes, no turn's record wrong.
  - **What breaks outside that:** nothing broke within it. The failure seen in practice is not a version at all: it is a versioned `claude.executable` whose build the updater has pruned. By the updater's documentation, a build is kept only while it is the linked one, one of the two newest, or running. 2.1.318 is the oldest of five on this machine.
  - **Not run:** anything older than 2.1.318 or newer than 2.1.325.
- **Codex: 0.163.0 to 0.169.0, installed whole.** Every call, as npm installs it or as the release's package unpacks.
  - **What breaks outside that:** a lone `codex` binary without the helper programs beside it. At 0.169.0 the shell tool then fails inside the model's turn, the turn ends `success`, and nothing at the protocol level says so.
  - **Not run:** anything older than 0.163.0 or newer than 0.169.0.
- **Builds share their stores safely.** Nearby builds read each other's sessions and threads in both directions.
  - **The one rewrite:** Claude's `migrationVersion` in `~/.claude.json`, which 2.1.318 and 2.1.319 lower to 14 and 2.1.321 on raise to 15. It has been moving daily on Michael's machine, with nothing observed broken.
  - **Codex** migrated and rewrote nothing.
- **Each harness reports its version in about 10 ms** with `--version`: "2.1.325 (Claude Code)" and "codex-cli 0.169.0". Both report it again at launch.

### how mega should find Claude Code

Michael set `claude.executable: claude` on this machine on 2026-11-14, and it is working. The spike ran the bare name too, and it passed everything.

- **A. The installed `claude`, by its bare name, as the default.** mega follows whatever build the updater links, as the operator's terminal does.
  - **For it:** an update never leaves mega on a pruned build, and mega and the terminal run the same build.
  - **Against it:** an update can bring a build nobody has tested, between one launch and the next. A running runtime keeps its build, so two conversations may run different builds for a while. An npm-installed `claude` is a script, and is refused as it is today, with its own message.
- **B. A versioned path, as today, kept from going stale.**
  - **For it:** the build stays fixed while it exists.
  - **Against it:** it breaks the day the updater prunes it, and the operator's terminal runs another build all week anyway. Keeping it alive would mean mega pinning or copying builds, and mega never installs.
- **C. The bare name first, then the newest build under `~/.local/share/claude/versions/`** when no `claude` is on the `PATH` mega uses, as for a launcher-started daemon whose shell lacks `~/.local/bin`. It is the same as A where A works, and it finds a native install that only lacks the link.

**Recommendation: A as the default, with C's fallback, and a version check (below).** An existing config naming a versioned path keeps working while the build exists. When it is gone, the refusal names what is installed beside it, not "install Claude Code". The `carried from the first-run batch` note asks for exactly that.

### a version check

- **For Claude, `claude --version`.** Run it in the start-up check of the harnesses and before each launch, in the project's environment; about 10 ms each time. Compare the number with the tested range.
  - **Inside the range,** nothing is said.
  - **Outside it,** the launch goes ahead with a plain notice, such as "Claude Code 2.1.330 is newer than the builds mega was tested with, 2.1.318 to 2.1.325".
  - **A build the spike shows cannot work** is refused with the reason. None is known today.
  - **The version the launch actually ran** is in the started facts, as today. The conversation's header already shows it; it could be compared with the range too.
- **For Codex, `codex --version`.** The same check against its range: "codex-cli 0.163.0 to 0.169.0".
- **A check that Codex's helper programs sit beside the binary.** The silent failure above makes this the more useful of the two Codex checks. Each package has a manifest, `codex-package.json` (`layoutVersion` 1), beside `bin/`. It names `entrypoint` and, from 0.167.2, `execServer`, which must exist. `bin/codex-code-mode-host` sits beside `bin/codex` in every package seen, though the manifest does not name it.
  - **The check:** find the manifest where the binary's package keeps it. For a `codex.binary` path, that is the parent of `bin/`. For npm's `codex.js`, it is `node_modules/@openai/codex-<platform>/vendor/<target>/`. Check that each program it names, and `codex-code-mode-host`, is there.
  - **A lone binary with no manifest beside it** is reported as missing its helpers: "'codex' at '…' is a lone binary; install Codex with npm or from the release's codex-package archive, whose helper programs it needs".
  - **Like every other missing executable,** this would be a `MissingError`, shown in the same places.

### what mega could say

- **At start-up,** in the start-up check's line per harness: the version found, beside where. For example, "harness 'claude' can run, in the operator's shell environment: claude code 2.1.325 at '/home/…/.local/bin/claude'". A version outside the range adds "(newer than the 2.1.318–2.1.325 mega was tested with)". The first page's list could carry the same.
- **At launch,** nothing inside the range. Outside it, a one-line notice in the history, as the runtime starts, naming the version and the range. It does not refuse.
- **In `mega version` and `/api/version`,** the tested ranges, beside `claude_sdk`.
- **In the README,** the ranges in place of the two exact builds. Claude Code is taken from `claude` on the `PATH`, and Codex must be installed whole.
