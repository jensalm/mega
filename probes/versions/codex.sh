#!/usr/bin/env bash
# codex.sh BINARY PORT [--discover] [--setup-only]
#
# runs codex-driver.mjs against one Codex build through a scratch mega on
# PORT, in a scratch CODEX_HOME. BINARY is a whole install: bin/codex in an
# unpacked codex-package-<target>.tar.gz, npm's codex.js, or the bare name
# codex for the one on PATH. the lone codex-<target>.tar.gz lacks the helper
# programs and mega refuses it. the fixture is MEGA_E2E_CODEX_FIXTURE.
#
# --discover then starts a second scratch mega, on fresh data and the same
# scratch CODEX_HOME, which discovers the run's thread and adopts it.
#
# the scratch CODEX_HOME is <scratch root>/codex-home, made on the first run
# and shared by every later one, so each build reads what the others wrote.
# it holds copies of ~/.codex's configuration and a symbolic link to
# ~/.codex/auth.json, which is never copied or read. before and after each
# mega the real ~/.codex is listed by name, size and time, and the run stops
# with exit 2 if a name came or went or a configuration file changed.
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

[[ $# -ge 2 ]] || { echo "usage: $0 BINARY PORT [--discover] [--setup-only]" >&2; exit 3; }
binary=$1 port=$2
shift 2
discover=0 setup=()
for arg in "$@"; do
  case $arg in
    --discover) discover=1 ;;
    --setup-only) setup=(--setup-only) ;;
    *) die "unknown option '$arg'" ;;
  esac
done
[[ -n ${MEGA_E2E_CODEX_FIXTURE:-} ]] || die "MEGA_E2E_CODEX_FIXTURE is not set; it names the disposable codex-rt1 fixture"
[[ -d $MEGA_E2E_CODEX_FIXTURE ]] || die "the fixture '$MEGA_E2E_CODEX_FIXTURE' is not a directory"
[[ -d $repo/shim/claude ]] || die "no shim at '$repo/shim/claude'"
if [[ $binary == */* ]]; then
  binary=$(cd "$(dirname "$binary")" && pwd)/$(basename "$binary")
  [[ -f $binary && -x $binary ]] || die "'$binary' is not an executable file"
  runs=$binary
else
  runs=$(command -v "$binary") || die "'$binary' is not on PATH"
fi
check_port "$port"

real_home="$HOME/.codex"
codex_home="$probe_root/codex-home"
mkdir -p "$probe_root"
[[ $(realpath -m "$codex_home") != "$(realpath -m "$real_home")"* ]] || die "the scratch CODEX_HOME '$codex_home' is inside the real ~/.codex"

# the scratch CODEX_HOME: the configuration copied, the login linked. made
# once; remove it (the auth.json link first) to start from a fresh copy.
if [[ ! -d $codex_home ]]; then
  say "making the scratch CODEX_HOME $codex_home"
  mkdir -m 700 "$codex_home"
  for f in "$real_home"/*.toml "$real_home/AGENTS.md"; do
    if [[ -f $f ]]; then cp -p "$f" "$codex_home/"; fi
  done
  for d in rules skills; do
    if [[ -d $real_home/$d ]]; then cp -rp "$real_home/$d" "$codex_home/"; fi
  done
  if [[ -e $real_home/auth.json ]]; then
    ln -s "$real_home/auth.json" "$codex_home/auth.json"
  else
    say "there is no ~/.codex/auth.json to link; Codex will run without a login"
  fi
else
  say "reusing the scratch CODEX_HOME $codex_home"
fi
say "the scratch CODEX_HOME holds: $(cd "$codex_home" && ls -A | tr '\n' ' ')"

# the real ~/.codex at the top level, by name, type, size and time. nothing
# is opened.
codex_state() {
  find "$real_home" -mindepth 1 -maxdepth 1 -printf '%f\t%y\t%s\t%TY-%Tm-%TdT%TT\n' | sort
}
# compare_state stops the run when a name came or went or a configuration
# file changed, and reports any other change: the operator's own Codex
# processes write their logs, caches and login under ~/.codex as they run.
compare_state() {
  local before=$1 after=$2
  if ! diff <(cut -f1,2 "$before") <(cut -f1,2 "$after") >/dev/null; then
    diff <(cut -f1,2 "$before") <(cut -f1,2 "$after") || true
    say "STOP: a name came or went in the real ~/.codex (above)"
    exit 2
  fi
  if ! diff <(grep -E '^([^\t]*\.toml|AGENTS\.md|rules|skills)\t' "$before" || true) <(grep -E '^([^\t]*\.toml|AGENTS\.md|rules|skills)\t' "$after" || true); then
    say "STOP: a configuration file in the real ~/.codex changed (above)"
    exit 2
  fi
  local changed
  changed=$(diff <(sort "$before") <(sort "$after") | grep '^>' | cut -f1 | sed 's/^> //' | tr '\n' ' ' || true)
  if [[ -n $changed ]]; then
    say "changed in the real ~/.codex by size or time, as the operator's own Codex processes write there: $changed"
  else
    say "the real ~/.codex is unchanged"
  fi
}

# the codex-cli line: under a scratch CODEX_HOME in a temporary directory,
# codex first warns that it will not create its PATH aliases there.
version=$(env CODEX_HOME="$codex_home" "$runs" --version 2>&1) || die "'$runs --version' failed: $version"
version=$(grep -m 1 'codex-cli' <<<"$version" || head -n 1 <<<"$version")
label=$(grep -oE '[0-9]+\.[0-9]+\.[0-9]+' <<<"$version" | head -n 1 || true)
dir="$probe_root/codex-${label:-unknown}-$port-$(date +%Y%m%d-%H%M%S)"
build_mega
new_run "$dir"
say "Codex '$binary' says '$version'"
echo "$version" >"$dir/version.txt"

# a run's mega names the build in codex.binary and no Claude Code. the
# scratch CODEX_HOME is in mega's own environment as well, for its start-up
# version check; a project's environment never takes mega's own, so the
# fixture's override is what puts it there, and the driver's guard checks it.
write_config() {
  cat >"$1/mega.yaml" <<EOF
listen: 127.0.0.1:$port
data: $1/data
claude:
  node: ${MEGA_PROBE_NODE:-node}
  shim: $repo/shim/claude
  executable: /bin/false
codex:
  binary: $binary
EOF
}

# one mega and one driver run, with the real ~/.codex listed before and
# after, and the guard's CODEX_HOME confirmed as the scratch one.
run_once() {
  local run=$1
  shift
  write_config "$run"
  codex_state >"$run/state-before.txt"
  start_mega "$run" "$port" CODEX_HOME="$codex_home"
  local code=0
  run_driver "$run" "$port" codex-driver.mjs "$@" || code=$?
  stop_mega
  codex_state >"$run/state-after.txt"
  compare_state "$run/state-before.txt" "$run/state-after.txt"
  local seen
  seen=$(summary_field "$run" codex_home || true)
  if [[ -n $seen && $seen != "$codex_home" ]]; then
    say "STOP: the run's CODEX_HOME was '$seen', not the scratch '$codex_home'"
    exit 2
  fi
  if [[ -n $seen ]]; then say "the run's CODEX_HOME was the scratch one, $seen"; fi
  return "$code"
}

export MEGA_PROBE_CODEX_HOME=$codex_home
code=0
run_once "$dir" "${setup[@]}" || code=$?
say "the run is in $dir: driver.out, summary.json, mega.out, xdg/state/mega/mega.log"

if (( discover )); then
  thread=$(summary_field "$dir" native_id || true)
  if [[ -z $thread && ${#setup[@]} -eq 0 ]]; then
    say "no thread in the run's summary, so nothing to discover"
    exit 1
  fi
  adopt="$dir/adopt"
  new_run "$adopt"
  say "a second mega, on fresh data and the same scratch CODEX_HOME, discovers and adopts ${thread:-nothing (setup only)}"
  run_once "$adopt" --adopt "${thread:-none}" "${setup[@]}" || code=$?
  say "the adoption run is in $adopt"
fi
exit "$code"
