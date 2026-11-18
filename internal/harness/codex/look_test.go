package codex

import (
	"os"
	"testing"

	"github.com/michaelquigley/mega/internal/harness"
)

// a look at a thread, through thread/read as an import reads it: the first
// prompt and the last reply. a thread with one prompt has no last prompt
// apart from it.
func TestLookReadsAThread(t *testing.T) {
	h := fakeHarness(t, "history")
	look, err := h.Look(t.Context(), harness.Place{Cwd: t.TempDir(), Env: os.Environ(), Find: os.Environ()}, "t-1")
	if err != nil {
		t.Fatal(err)
	}
	if look.FirstPrompt != "next" || look.LastPrompt != "" || look.LastReply != "going" || look.Prompts != 1 {
		t.Errorf("look: %+v", look)
	}
}
