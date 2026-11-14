package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/michaelquigley/mega/internal/model"
)

// what the sdk printed when claude code was not installed (the clean-machine
// rehearsal, 2026-11-13), shortened: a message, then a node stack.
const sdkFailure = `exit status 1: ReferenceError: Claude Code native binary not found at /home/x/.local/share/claude/versions/2.1.318. Please ensure Claude Code is installed via native installer or specify a valid path with options.pathToClaudeCodeExecutable.
    at u5 (file:///home/x/mega/shim/claude/node_modules/@anthropic-ai/claude-agent-sdk/sdk.mjs:12:345)
    at async main (file:///home/x/mega/shim/claude/shim.mjs:88:5)`

func TestSummarizeKeepsTheReasonAndDropsTheStack(t *testing.T) {
	got := summarize(sdkFailure)
	if !strings.Contains(got, "Claude Code native binary not found") || strings.Contains(got, "sdk.mjs") || strings.Contains(got, "\n") {
		t.Errorf("summary = %q", got)
	}
	// a stack whose first line is not the error: the error line joins it.
	got = summarize("file:///x/shim.mjs:3\n    throw e\n    ^\n\nTypeError: x is not a function\n    at y (file:///x:1:1)")
	if got != "throw e TypeError: x is not a function" {
		t.Errorf("summary = %q", got)
	}
	if long := summarize(strings.Repeat("x", 1000)); len([]rune(long)) != summarizeLimit+1 {
		t.Errorf("long summary of %d runes", len([]rune(long)))
	}
	if got := summarize("\n  \n"); got != "it gave no reason" {
		t.Errorf("empty = %q", got)
	}
}

// a harness that fails to start is refused with its own reason, coded
// harness_failed, never an internal fault; its stack goes nowhere the page
// reads, the refusal or the history.
func TestAHarnessFailureCarriesItsReason(t *testing.T) {
	r := newRig(t)
	r.fake.FailLaunch(errors.New(sdkFailure))
	conv, err := r.svc.StartConversation(r.project.ID, model.HarnessClaude, "", "one", LaunchChoice{})
	if ErrorCode(err) != CodeHarnessFailed || !strings.Contains(err.Error(), "native binary not found") || strings.Contains(err.Error(), "sdk.mjs") {
		t.Fatalf("refusal = %q (%q)", err, ErrorCode(err))
	}
	for _, ev := range r.events(conv.ID) {
		if strings.Contains(ev.Text, "sdk.mjs") || strings.Contains(ev.Text, "    at ") {
			t.Errorf("the history holds the stack: %q", ev.Text)
		}
	}
}
