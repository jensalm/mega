// looktime times an adapter's look at one session, as the adopt dialog asks
// for it, from a scratch store the environment names (CLAUDE_CONFIG_DIR,
// PI_CODING_AGENT_DIR, CODEX_HOME). it prints the time and the sizes of what
// came back, never the text.
//
//	go run ./probes/look/looktime HARNESS CWD SESSION
//
// claude runs this checkout's shim (shim/claude) through node; codex runs
// the codex on the PATH.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/harness/claude"
	"github.com/michaelquigley/mega/internal/harness/codex"
	"github.com/michaelquigley/mega/internal/harness/pi"
)

func main() {
	name, cwd, session := os.Args[1], os.Args[2], os.Args[3]
	var looker harness.Looker
	switch name {
	case "claude":
		shim, _ := filepath.Abs("shim/claude")
		looker = claude.New(claude.Config{Node: "node", Shim: shim, Executable: "/bin/true"})
	case "codex":
		looker = codex.New(codex.Config{Binary: "codex"})
	case "pi":
		looker = pi.New(pi.Config{})
	default:
		fmt.Fprintln(os.Stderr, "harness: claude, codex or pi")
		os.Exit(2)
	}
	env := os.Environ()
	start := time.Now()
	look, err := looker.Look(context.Background(), harness.Place{Cwd: cwd, Env: env, Find: env}, session)
	took := time.Since(start).Round(time.Millisecond)
	if err != nil {
		fmt.Printf("%s look: %s, then failed: %v\n", name, took, err)
		os.Exit(1)
	}
	fmt.Printf("%s look: %s; %d prompts; parts of %d, %d and %d bytes\n", name, took, look.Prompts, len(look.FirstPrompt), len(look.LastPrompt), len(look.LastReply))
}
