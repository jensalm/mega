# the plumbing claude.sh and codex.sh share: the scratch root, the port, the
# mega build, and starting and stopping a scratch mega. sourced, not run.
#
# a scratch mega runs with its own data directory and XDG directories under
# the scratch root, and with the operator's own HOME and login: the builds run
# as the operator does, against the real configuration. it listens on the
# port given, never 8470.

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo=$(cd "$here/../.." && pwd)
probe_root=${MEGA_PROBE_DIR:-${TMPDIR:-/tmp}/mega-versions-probe}
mega_bin="$probe_root/mega"
mega_pid=""

say() { printf '[%s] %s\n' "$(date +%H:%M:%S)" "$*"; }
die() { printf 'cannot run: %s\n' "$*" >&2; exit 3; }

# check_port refuses the operator's port, anything not a port, and a port
# something already answers on.
check_port() {
  local port=$1
  [[ $port =~ ^[0-9]+$ ]] && (( port > 1024 && port < 65536 )) || die "'$port' is not a port"
  [[ $port != 8470 ]] || die "port 8470 is the operator's mega; use a scratch port"
  if curl -s -o /dev/null --max-time 2 "http://127.0.0.1:$port/api/workspace"; then
    die "something already answers on port $port; stop it or use another"
  fi
}

# build_mega builds this checkout's mega into the scratch root. the embedded
# UI and the shim's dependencies come from `make`.
build_mega() {
  [[ -f $repo/ui/dist/index.html ]] || die "the UI build is missing; run 'make' first"
  mkdir -p "$probe_root"
  say "building mega from $repo into $mega_bin"
  (cd "$repo" && go build -o "$mega_bin" ./cmd/mega) || die "the mega build failed"
}

# new_run makes a fresh run directory, with its data and XDG directories.
new_run() {
  local dir=$1
  [[ ! -e $dir ]] || die "'$dir' exists already"
  mkdir -p "$dir/data" "$dir/xdg/state" "$dir/xdg/config" "$dir/xdg/data"
}

# start_mega starts mega on a run directory's config, with its scratch XDG
# directories, and waits for it to answer. extra NAME=VALUE pairs go into its
# environment. its output goes to mega.out; its log, since no terminal sees
# it, to xdg/state/mega/mega.log.
start_mega() {
  local dir=$1 port=$2
  shift 2
  env "$@" XDG_STATE_HOME="$dir/xdg/state" XDG_CONFIG_HOME="$dir/xdg/config" XDG_DATA_HOME="$dir/xdg/data" \
    "$mega_bin" --config "$dir/mega.yaml" >"$dir/mega.out" 2>&1 &
  mega_pid=$!
  local i
  for i in $(seq 1 120); do
    if ! kill -0 "$mega_pid" 2>/dev/null; then
      mega_pid=""
      tail -n 20 "$dir/mega.out" "$dir/xdg/state/mega/mega.log" 2>/dev/null
      die "mega exited at start; see $dir"
    fi
    if curl -s -o /dev/null --max-time 2 "http://127.0.0.1:$port/api/workspace"; then
      say "mega $mega_pid answers on 127.0.0.1:$port, data under $dir"
      return 0
    fi
    sleep 0.25
  done
  stop_mega
  die "mega did not answer on port $port within 30 s"
}

# stop_mega asks mega to stop as a terminal's Ctrl+C does, which releases
# every runtime, and kills it after a minute.
stop_mega() {
  [[ -n $mega_pid ]] || return 0
  local pid=$mega_pid i
  mega_pid=""
  kill -INT "$pid" 2>/dev/null || return 0
  for i in $(seq 1 240); do
    kill -0 "$pid" 2>/dev/null || { say "mega $pid stopped"; return 0; }
    sleep 0.25
  done
  say "mega $pid did not stop within 60 s of SIGINT; killing it"
  kill -KILL "$pid" 2>/dev/null || true
}
trap stop_mega EXIT

# run_driver runs a driver against the scratch mega, keeping its output and
# summary in the run directory, and returns its exit code.
run_driver() {
  local dir=$1 port=$2 driver=$3
  shift 3
  set +e
  MEGA_PROBE_URL="http://127.0.0.1:$port" MEGA_PROBE_SUMMARY="$dir/summary.json" \
    node "$here/$driver" "$@" 2>&1 | tee "$dir/driver.out"
  local code=${PIPESTATUS[0]}
  set -e
  return "$code"
}

# summary_field prints one field of a run's summary.
summary_field() {
  node -e 'const s = JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")); const v = s[process.argv[2]]; if (v !== undefined) console.log(typeof v === "string" ? v : JSON.stringify(v))' "$1/summary.json" "$2" 2>/dev/null
}
