#!/usr/bin/env bash
# reruns the core pi scenarios against the installed pi, so a new build can be
# compared with the recorded observations. each scenario names the one dialog
# it answers (--expect), by its title and with the answer it gives; any other
# dialog stops the run unanswered.
#
# compared with the recorded observations (NOTES.md, the capability report's
# pi section). it uses the operator's own default model and provider, as pi's
# settings name them (no --model), and never passes --approve: the fixture
# must already be trusted in pi's own terminal, which the operator does.
#
#   probes/pi/rerun.sh            # every scenario
#   probes/pi/rerun.sh p1 r2      # some of them
#
# the fixture is MEGA_E2E_PI_FIXTURE, else Michael's pi-rt1. logs go to
# logs/v<version>/ (gitignored): each scenario's frames (.jsonl), the probe's
# summary (.summary.json) and pi's stderr (.stderr.txt).
#   p1   P1-P3: a new session, the marker, a cat, the ask_user question, and
#        the extension's approval for sleep 3, answered after 4 s
#   p4   P4: the same session reopened by exact id, marker recall, sleep 2
#   r1   R1: abort during a long tool call, then the next turn; the command,
#        `timeout 45 tail -f /dev/null`, says no "sleep", so the fixture's
#        gate does not ask, and nothing is answered to reach the abort
#   r2   R2: abort while the approval dialog is pending, then a late answer
#   r3   R3: a dialog never answered; the run ends by the probe's timeout
#   r4   R4: the probe SIGKILLed mid-turn, then a resume of that session, on
#        an ungated `timeout 30 tail -f /dev/null` for the same reason
#   p7   P7: two processes on one session id at once
set -uo pipefail
cd "$(dirname "$0")"
fixture=${MEGA_E2E_PI_FIXTURE:-/home/michael/Repos/q/research/roundhouse-fixtures/pi-rt1}
version=$(pi --version 2>/dev/null | head -n 1)
logs=logs/v${version:-unknown}
mkdir -p "$logs"
want=("$@")
[[ ${#want[@]} -gt 0 ]] || want=(p1 p4 r1 r2 r3 r4 p7)
ids="$logs/session-ids.txt"

run() {
  local name=$1
  shift
  node probe.mjs run --cwd "$fixture" --log "$logs/$name.jsonl" --turn-timeout-ms 600000 "$@" \
    >"$logs/$name.summary.json" 2>"$logs/$name.stderr.txt"
  echo "[$(date +%H:%M:%S)] $name exit $?"
}
has() { [[ " ${want[*]} " == *" $1 "* ]]; }

echo "pi $version, fixture $fixture, logs $logs"
if has p1 || has p4; then
  s=$(uuidgen); echo "p1 $s" >>"$ids"
  run p1 --session-id "$s" --turns scenarios/rt1.json --answer-delay-ms 4000 \
    --expect '[Cc]olou?r=>Blue|^Approve bash command: sleep 3$=>Allow'
  has p4 && run p4 --session "$s" --turns scenarios/rt1-rpc-return.json --expect '^Approve bash command: sleep 2$=>Allow'
fi
if has r1; then
  s=$(uuidgen); echo "r1 $s" >>"$ids"
  run r1 --session-id "$s" --turns scenarios/int-active-ungated.json --interrupt-after-ms 15000 --interrupt-turn 1
fi
if has r2; then
  s=$(uuidgen); echo "r2 $s" >>"$ids"
  run r2 --session-id "$s" --turns scenarios/int-dialog.json --interrupt-on-dialog --stale-after-cancel \
    --expect '^Approve bash command: sleep 5$=>Allow'
fi
if has r3; then
  s=$(uuidgen); echo "r3 $s" >>"$ids"
  run r3 --session-id "$s" --turns scenarios/int-dialog.json --never-answer --timeout-ms 120000
fi
if has r4; then
  s=$(uuidgen); echo "r4 $s" >>"$ids"
  run r4 --session-id "$s" --turns scenarios/owner-kill-ungated.json --kill-owner-after-ms 20000
  sleep 6
  run r4-recover --session "$s" --turns scenarios/recover.json
fi
if has p7; then
  s=$(uuidgen); echo "p7 $s" >>"$ids"
  run p7-a --session-id "$s" --turns scenarios/shared-a.json &
  sleep 4
  run p7-b --session "$s" --turns scenarios/shared-b.json
  wait
fi
