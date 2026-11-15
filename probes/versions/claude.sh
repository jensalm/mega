#!/usr/bin/env bash
# claude.sh BUILD PORT [--setup-only]
#
# runs driver.mjs against one Claude Code build through a scratch mega on
# PORT. BUILD is a native build, such as ~/.local/share/claude/versions/2.1.330,
# or the bare name claude, which mega finds on the login shell's PATH. the
# fixture is MEGA_E2E_CLAUDE_FIXTURE. the shim is the checkout's, with its
# pinned Agent SDK, or MEGA_PROBE_SHIM.
#
# before and after the run it records the installed builds and where claude
# on PATH points, and stops with exit 2 if either changed.
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

[[ $# -ge 2 ]] || { echo "usage: $0 BUILD PORT [--setup-only]" >&2; exit 3; }
build=$1 port=$2
shift 2
[[ -n ${MEGA_E2E_CLAUDE_FIXTURE:-} ]] || die "MEGA_E2E_CLAUDE_FIXTURE is not set; it names the disposable claude-rt1 fixture"
[[ -d $MEGA_E2E_CLAUDE_FIXTURE ]] || die "the fixture '$MEGA_E2E_CLAUDE_FIXTURE' is not a directory"
shim=${MEGA_PROBE_SHIM:-$repo/shim/claude}
[[ -d $shim/node_modules ]] || die "the shim at '$shim' has no node_modules; run 'make shim'"
if [[ $build == */* ]]; then
  build=$(cd "$(dirname "$build")" && pwd)/$(basename "$build")
  [[ -f $build && -x $build ]] || die "'$build' is not an executable file"
  runs=$build
else
  runs=$(command -v "$build") || die "'$build' is not on PATH"
fi
check_port "$port"

# what the builds share and the updater changes: the installed builds, by
# name, size and time, and the link on PATH.
claude_state() {
  echo "versions: ~/.local/share/claude/versions"
  ls -l --time-style=+%Y-%m-%dT%H:%M:%S "$HOME/.local/share/claude/versions" 2>&1 | grep -v '^total ' || true
  local link
  link=$(command -v claude || true)
  echo "claude on PATH: ${link:-none} -> $( [[ -n $link ]] && readlink -f "$link" )"
}

version=$(env DISABLE_AUTOUPDATER=1 DISABLE_UPDATES=1 "$runs" --version 2>&1 | head -n 1) || die "'$runs --version' failed: $version"
label=$(grep -oE '[0-9]+\.[0-9]+\.[0-9]+' <<<"$version" | head -n 1 || true)
dir="$probe_root/claude-${label:-unknown}-$port-$(date +%Y%m%d-%H%M%S)"
build_mega
new_run "$dir"
say "Claude Code '$build' says '$version'"
echo "$version" >"$dir/version.txt"
claude_state >"$dir/state-before.txt"

cat >"$dir/mega.yaml" <<EOF
listen: 127.0.0.1:$port
data: $dir/data
claude:
  node: ${MEGA_PROBE_NODE:-node}
  shim: $shim
  executable: $build
codex:
  binary: /bin/false
EOF

start_mega "$dir" "$port"
code=0
run_driver "$dir" "$port" driver.mjs "$@" || code=$?
stop_mega

claude_state >"$dir/state-after.txt"
if ! diff -u "$dir/state-before.txt" "$dir/state-after.txt"; then
  say "STOP: the installed builds or the claude link changed during the run (above); see $dir"
  exit 2
fi
say "the installed builds and the claude link are unchanged"
say "the run is in $dir: driver.out, summary.json, mega.out, xdg/state/mega/mega.log"
exit "$code"
