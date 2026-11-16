package codex

import (
	"strings"
	"testing"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
)

// what a tool call invoked comes from the thread item its event recorded: a
// command execution's whole command, past the summary's cap, and where it
// ran; a file change's files by what was done to each; an MCP tool call's
// arguments; and nothing where the event holds no item.
func TestInvocationReadsTheRecordedItem(t *testing.T) {
	long := "/bin/bash -lc '" + strings.Repeat("printf one-line-of-a-long-command; ", 10) + "'"
	h := New(Config{})
	use := func(raw map[string]any) *model.Event {
		t.Helper()
		evs, err := mapItem(raw)
		if err != nil || len(evs) == 0 || evs[0].Kind != model.EventToolUse {
			t.Fatalf("mapped %v, %v", evs, err)
		}
		return evs[0]
	}
	ev := use(map[string]any{"type": "commandExecution", "id": "c1", "command": long, "cwd": "/w", "status": "completed", "exitCode": 0, "aggregatedOutput": "ok"})
	if len(ev.Text) >= len(long) {
		t.Fatalf("the summary was not capped: %d bytes", len(ev.Text))
	}
	if inv, ok := h.Invocation(ev); !ok || inv.Tool != "shell" || inv.Command != long || len(inv.Fields) != 1 || inv.Fields[0] != (harness.Field{Name: "cwd", Value: "/w"}) {
		t.Errorf("command: %v %+v", ok, inv)
	}
	ev = use(map[string]any{"type": "fileChange", "id": "f1", "status": "completed", "changes": []any{
		map[string]any{"path": "/w/a.go", "kind": map[string]any{"type": "update"}, "diff": "@@"},
		map[string]any{"path": "/w/b.go", "kind": map[string]any{"type": "add"}, "diff": "+"},
	}})
	if inv, ok := h.Invocation(ev); !ok || inv.Tool != "edit" || len(inv.Fields) != 2 || inv.Fields[0] != (harness.Field{Name: "update", Value: "/w/a.go"}) || inv.Fields[1] != (harness.Field{Name: "add", Value: "/w/b.go"}) {
		t.Errorf("file change: %v %+v", ok, inv)
	}
	ev = use(map[string]any{"type": "mcpToolCall", "id": "m1", "server": "fixture", "tool": "echo", "status": "completed", "arguments": map[string]any{"text": "hi", "times": 2}})
	if inv, ok := h.Invocation(ev); !ok || inv.Tool != "fixture.echo" || len(inv.Fields) != 2 || inv.Fields[0] != (harness.Field{Name: "text", Value: "hi"}) || inv.Fields[1] != (harness.Field{Name: "times", Value: "2"}) {
		t.Errorf("mcp: %v %+v", ok, inv)
	}
	if _, ok := h.Invocation(&model.Event{Kind: model.EventToolUse, Tool: "shell", Text: "shell ls"}); ok {
		t.Error("an event with no payload says what it invoked")
	}
}
