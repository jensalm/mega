# mega

metawoo mega: build your constellation. mega is where a solo operator runs an orienteering practice with a gang of agents.

It is a project-oriented workspace for working directly with those agents: Claude Code, Codex and pi conversations, grouped by the checkouts they work in, in one page on your own machine. It is not an autonomous coordinator and not a replacement coding harness. Each agent runs as its own harness would in a terminal, with its own sessions, settings and permissions; mega supervises them, shows what needs you, and lets you hand a conversation to a terminal and back.

One Go binary serves the workspace and its UI on loopback. A small Node bridge (`shim/claude`) talks to Claude Code through its Agent SDK; Codex is driven through `codex app-server`.

## What it needs

- Go (the version in `go.mod`) and Node 24, to build.
- Claude Code, from its native installer: mega runs `claude` on your `PATH`, the installer's link to the build it last installed. With no `claude` on the `PATH`, as a desktop launcher's shell may have, mega runs the newest build in `~/.local/share/claude/versions/`. `claude.executable` names another, which must be a native build, not a script.
- pi (`@earendil-works/pi-coding-agent`), installed with npm. mega runs `pi` on your `PATH`, or wherever `pi.binary` says, in its RPC mode. pi takes its model from its own settings, so a local model it is set up for works under mega as it does in pi's terminal. A project with its own pi extensions must be trusted in pi's own terminal first: run `pi` in the checkout once and accept. mega never trusts a project for pi.
- Codex, installed whole: with npm, or from the release's `codex-package-<target>.tar.gz`. The release's lone `codex-<target>.tar.gz` lacks the helper programs Codex runs a shell command with, so mega refuses it. mega runs `codex` on your `PATH`, or wherever `codex.binary` says.
- Your own Claude and Codex logins. mega uses them as they are and changes nothing in them.

mega was tested with Claude Code 2.1.318 to 2.1.325, with the shim's pinned Agent SDK 0.3.318, with Codex 0.163.0 to 0.169.0, and with pi 1.5.2. These are its tested ranges. A build outside them still runs, and the start-up log, the first page, `mega version` and the version panel say where each build stands. Claude Code updates itself every day or two, so a build newer than the range is the normal state, and is said plainly. A build older than the range is the likelier to fail, since the Agent SDK expects the Claude Code of its own patch or later. It is marked wherever versions are shown, and on a conversation's "Runtime started" line.

## Build and run from source

```bash
git clone https://github.com/michaelquigley/mega.git
cd mega
make          # builds the UI, installs the shim's pinned SDK, installs `mega` into your Go bin directory
mega          # serves http://127.0.0.1:8470
```

The installed binary is stamped with the path of `shim/claude` in the checkout it was built from, so keep the checkout where it is. `make test` runs the full gate, including the UI build and lint and the shim's own tests. `make clean` removes only what the build made: `ui/dist`, `ui/node_modules`, `shim/claude/node_modules`, `checks/node_modules`, what `go clean` removes from the source tree, and the `mega` binary in your Go bin directory. Nothing else in that directory is touched.

A clean `npm ci` in `ui/` and in `shim/claude/` prints no deprecation warning and reports no vulnerabilities. `ui/package.json` pins `minimatch` to 11.0.8 with an `overrides` entry. ESLint and `typescript-eslint` depend on a 10.x release inside the range of GHSA-g3hv-75wg-m249, for which no 10.x fix exists. The linter only reads the project's own source, and nothing from it ships.

The browser checks in `checks/` drive the page with Playwright against a scratch mega and fake harnesses. They are not part of `make test`. Install the browser they use once, then run them:

```bash
make checks-browser   # downloads the Chromium the pinned Playwright drives, into ~/.cache/ms-playwright
make checks           # every suite; make checks SUITE=requests runs one
```

`make checks` fails, saying so, when that browser is missing. `checks/README.md` says what each suite covers and how to add one.

`make e2e` runs the harness round trips in `internal/e2e` against the real Claude Code, Codex and pi, with your own logins and pi's own default model, in three disposable fixture checkouts; `internal/e2e/doc.go` says what each holds. They cost model turns and are run by hand. Each takes its fixture from `MEGA_E2E_CLAUDE_FIXTURE`, `MEGA_E2E_CODEX_FIXTURE` or `MEGA_E2E_PI_FIXTURE` and is skipped when that is unset; the pi round trips also skip when pi's model does not answer. To avoid typing them, put them in `e2e.local.mk` beside the `Makefile`, which git ignores:

```make
MEGA_E2E_CLAUDE_FIXTURE ?= /path/to/claude-rt1
MEGA_E2E_CODEX_FIXTURE ?= /path/to/codex-rt1
```

A value in the environment still wins over the file.

Configuration is optional, in `~/.config/mega/mega.yaml`:

```yaml
listen: 127.0.0.1:8470       # loopback only; anything else is refused
data: ~/.local/share/mega    # the workspace: registry and conversation logs
claude:
  node: node
  shim: /path/to/mega/shim/claude
  executable: claude         # a bare name is looked up on PATH; a path is used as it is
codex:
  binary: codex
pi:
  binary: pi
snapshots:                   # copies of each harness's configuration, below
  keep: 10                   # ordinary snapshots kept; 0 takes none
  keep_at_version: 20        # snapshots taken when a harness version changed
```

`mega --data DIR` runs against another workspace, `mega --config FILE` reads another config, and `mega --listen 127.0.0.1:PORT` serves on another port. Like `listen` in the config, the flag must be a loopback address.

### Running as a daemon

`mega daemon` runs the same service under a tray icon. The menu has "open mega" to open the page, "shut down when idle" to arm the shutdown gate (it reads "cancel shutdown" while armed), and Quit. Quit goes through the gate: on an idle workspace it releases every agent and stops; with anything in flight it arms the gate and opens the page instead, so a stray click never interrupts an agent. The tray's icon turns gold, with a count in its tooltip, while something needs you.

To launch it from the desktop:

```bash
mega desktop integrate   # installs com.michaelquigley.mega.desktop and its icons under ~/.local/share
mega desktop remove      # removes exactly those files
```

The launcher runs `mega daemon` through your interactive shell (the one `SHELL` names: bash, zsh or fish), so mega itself starts the way a terminal-started mega does. The agents don't depend on it: each project runs in its own environment (below). `integrate` prints the exact `Exec` line it wrote. `mega desktop integrate --plain` writes the direct form if you want it. The launcher runs the binary from wherever it was when you ran `integrate`; run it again if you move the binary. Nothing starts at login.

mega tells you what it is missing before you try to launch anything:
- **At start-up,** the log has one line per harness: where its programs were found and which version, against the tested range, or why it cannot run.
- **On the first page,** with no project registered yet, mega names the control to begin with, and lists each harness as found, with its version, or missing, with the reason.
- **On a project's page,** once you register a project, mega resolves its environment in the background, and its page lists any harness that cannot run there. It does so again at every start.

A harness that cannot run is marked as unable to run in the new-conversation dialog, and its conversations refuse new work with the name of what is missing. When `claude.executable` or `codex.binary` names a path that is not there, the reason names the path, says to install it or set the value, and names the config file mega read. A versioned Claude Code path the installer has since pruned is named with the builds installed beside it and how to follow the installed one instead; it is never told to install what is installed. A Codex without its helper programs is named as a lone binary, with how to install it whole. A harness that fails some other way says so in its own words; the rest goes to the log.

With no terminal, as when the launcher starts it, mega logs to `~/.local/state/mega/mega.log` (or under `XDG_STATE_HOME`); `mega version` prints the path, with the tested ranges and the version of each harness it finds, and so does the panel the version in the side panel's footer opens. At start-up a log past 10 MB becomes `mega.log.1`, replacing an earlier one, and a fresh log begins.

Launching the daemon while mega is already running opens the running mega's page and exits. `mega serve` beside a running mega refuses to start and names it.

Prebuilt binaries are not offered yet: the Claude shim runs from a checkout with its pinned Agent SDK installed beside it, so a binary on its own cannot run Claude conversations.

## Project environments

Every agent in a project runs in that project's own shell environment: the one a terminal opened in the checkout and activated your usual way would have. mega never hands an agent the environment mega itself was started in. mega starts your shell in the checkout from a clean base, as a login and interactive shell so your profile scripts and start-up files both run, and applies the project's activation, at every start of a runtime and for its own helpers.

The activation lives in the repository, in `.mega/environment.yaml`, as one key:

```yaml
# a Go project isolated by devprj: its own GOPATH, its bin first on PATH
activate: devprj
```

```yaml
# a project that uses direnv
activate: eval "$(direnv export bash)"
```

A project with no file gets your shell's own environment in the checkout and nothing more, which suits a project that needs no activation.

**mega will not run a repository's activation until you approve it.** The project page shows the file and an Approve control. If the file changes after you approve it, nothing starts in that project until you have compared the two versions and approved the new one. The approved version is found in the checkout's git history; mega stores only its SHA-256 digest. An approval made before mega stored the full digest needs approving once more. A local override on the project page lets you ignore the file, or run another activation on this machine. A file that cannot be read, an activation that fails, and one that takes longer than ten seconds each refuse to start anything, and say why.

**Approving the file approves its text, nothing more.** An activation that reads other files in the repository is trusting them. `devprj` reads `.devprj`, and a snippet may source a script, and an agent can usually change such a file without asking. Two things limit that:
- **mega finds its own programs before the activation.** `node`, `git` and the harness binaries come from your shell's environment as it was before the activation ran. An activation can change what an agent's commands see, but never which program mega itself runs.
- **The approval remembers the names of the variables the activation changed.** If a later start finds it changing a different set, a new `LD_PRELOAD` say, nothing starts until you have looked and approved again. The project page shows the added and removed names, never their values.

mega records which environment each runtime started in by names and a fingerprint, never values. The control bar at the foot of a conversation marks a running agent that started in an older environment than the project's current one; the header keeps the fingerprint. The project page shows the variables the activation changes, and their values only when you ask.

## Safety posture

- **Loopback only.** The listener must be a loopback address. Requests are refused unless their `Host` is loopback and any `Origin` matches it, which keeps out web pages in your own browser as well as the network.
- **No login.** Anyone who can reach the port on your machine can drive your agents, which is why it never listens anywhere else.
- **Agents run with your permissions.** Each harness runs as you, in your checkout, with your settings and your permission rules. mega never relaxes a harness's permissions, never sets a trust flag on your behalf, and never approves anything for you: every approval it is asked for waits for your answer.
- **pi asks for nothing unless an extension does.** pi has no permission model: it runs its tools without asking, in mega as in its own terminal, and the launch dialog and the conversation say so.
  - **Every dialog is a question.** Nothing marks a pi extension's dialog as an approval, so none is shown as one or offered for the session.
  - **No trust from mega.** mega never passes `--approve` and never writes pi's trust, so a project you have not trusted in pi's terminal runs without its extensions.
  - **Dismiss is yours.** It sends pi's own cancel, as Escape does in pi's terminal, and only when you press it. mega never answers a dialog itself.
  - **Release ends pi.** A pending dialog holds pi open, so releasing ends its process, and the command the dialog gates does not run; the release dialog says so first.
  - **No lock.** pi has none, so do not open a session in a terminal while mega runs it.
- **Your harness configuration is copied, and never written.** mega keeps snapshots of each harness's configuration under its data directory, in `snapshots/`, readable only by you.
  - **What is copied:** an include list per harness, of settings, instructions, rules, skills and agents you keep, plugin lists, the version and migration markers, and `~/.claude.json`.
  - **What never is:** a credential or token file, a transcript, a session store, a history, a database or a cache, nor a file the list does not name. A link is kept as a link and never followed.
  - **When:** a snapshot is taken in the background at start-up when the configuration changed, and when a harness reports a new version.
  - **Restoring is yours.** `mega snapshots` lists them, and `mega snapshots diff NAME` names the files that differ from your live configuration and prints the copy commands. It runs none of them.
  - **Treat them as private.** Configuration can hold secrets, such as an `env` block in a settings file, so the snapshots deserve the same care as the originals.
- **A conversation's model, effort and permission mode are its own.** A new conversation copies them once from the harness's settings, and mega passes them on every launch; only you change them after that. A bypass mode in your settings is never copied: the conversation starts in the mode that asks. The control bar shows what you request, and each turn's line shows what the harness ran with.

## More

- `docs/current/mega.md` describes the system as built.
- `docs/future/mega.md` is the design it grows toward, and `docs/future/roadmap/` holds the items on the way.
- `docs/research/harness-capabilities.md` records what each harness was observed to do, with versions.

## License

MIT; see `LICENSE`.

The Claude Code, Codex and pi marks on the control bar say which harness a conversation runs on; they belong to their owners and imply no endorsement.
