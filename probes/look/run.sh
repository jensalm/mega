#!/bin/bash
# times the adopt dialog's look at a long session, for each harness, on
# synthetic sessions written in a scratch directory: never a real store,
# and no login or model call. it prints times and sizes, never the text.
#
#   probes/look/run.sh [TURNS]      # from the repository's root; 2000 by default
set -euo pipefail
turns=${1:-2000}
scratch=$(mktemp -d /tmp/mega-look-probe.XXXX)
trap 'rm -rf "$scratch"' EXIT
mkdir -p "$scratch/cwd"
ids=$(node probes/look/gen.mjs "$scratch" "$scratch/cwd" "$turns")
id() { echo "$ids" | node -e "let s='';process.stdin.on('data',d=>s+=d).on('end',()=>console.log(JSON.parse(s)['$1']))"; }
du -sh "$scratch/claude" "$scratch/codex" "$scratch/pi"
go build -o "$scratch/looktime" ./probes/look/looktime
for i in 1 2 3; do CLAUDE_CONFIG_DIR="$scratch/claude" "$scratch/looktime" claude "$scratch/cwd" "$(id claude)"; done
for i in 1 2 3; do CODEX_HOME="$scratch/codex" "$scratch/looktime" codex "$scratch/cwd" "$(id codex)"; done
for i in 1 2 3; do PI_CODING_AGENT_DIR="$scratch/pi" "$scratch/looktime" pi "$scratch/cwd" "$(id pi)"; done
