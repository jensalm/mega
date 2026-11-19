package claude

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
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

// a model switch's /model record, read back through Harness.History and the
// shim's messages command (here the fake shim's, serving it as the sdk gave
// it, marked isCompletedLocalCommand; recorded 2026-11-19 from the
// claude-rt1 fixture, claude code 2.1.326), beside a prompt made only of the
// same tags and unmarked: the record is native, the prompt the operator's.
func TestHistoryKeepsAModelSwitchRecordNative(t *testing.T) {
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
	recorded, err := os.ReadFile(filepath.Join("testdata", "model-switch-history.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(shim, "messages-s-2.jsonl"), recorded, 0o600); err != nil {
		t.Fatal(err)
	}
	h := New(Config{Node: node, Shim: shim, Executable: truePath(t)})
	history, err := h.History(t.Context(), harness.Place{Cwd: t.TempDir(), Env: os.Environ(), Find: os.Environ()}, "s-2")
	if err != nil {
		t.Fatal(err)
	}
	evs := history.Events
	if len(evs) != 2 {
		t.Fatalf("history = %+v", evs)
	}
	if evs[0].Kind != model.EventNative || evs[0].Text != "user/local-command" || evs[0].NativeID != "b1f8d962-ce9a-4033-b49a-8fa15c9d0f9b" {
		t.Errorf("the /model record imports as %+v", evs[0])
	}
	if evs[1].Kind != model.EventUser || evs[1].Text != "<command-args>Explain why this failed.</command-args>" {
		t.Errorf("the prompt made of tags imports as %+v", evs[1])
	}
}
