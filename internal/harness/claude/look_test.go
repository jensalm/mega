package claude

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/michaelquigley/mega/internal/harness"
)

// a look at a recorded session, through the shim's messages command as an
// import reads it (here the fake shim's, serving the recorded transcript):
// the first prompt, the last, and the last reply.
func TestLookReadsARecordedSession(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	shim := t.TempDir()
	fake, err := filepath.Abs(filepath.Join("..", "..", "..", "shim", "claude", "fake", "shim.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(fake, filepath.Join(shim, "shim.mjs")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(shim, "node_modules", "@anthropic-ai", "claude-agent-sdk"), 0o700); err != nil {
		t.Fatal(err)
	}
	recorded, err := os.ReadFile(filepath.Join("testdata", "history.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(shim, "messages-s-1.jsonl"), recorded, 0o600); err != nil {
		t.Fatal(err)
	}
	h := New(Config{Node: node, Shim: shim, Executable: truePath(t)})
	look, err := h.Look(t.Context(), harness.Place{Cwd: t.TempDir(), Env: os.Environ(), Find: os.Environ()}, "s-1")
	if err != nil {
		t.Fatal(err)
	}
	if look.FirstPrompt != "Reply with the project instruction marker followed by ALPHA-1a2b" || look.LastPrompt != "Reply with TERMINAL-MARKER-e9d1" || look.LastReply != "TERMINAL-MARKER-e9d1" || look.Prompts != 2 {
		t.Errorf("look: %+v", look)
	}
}
