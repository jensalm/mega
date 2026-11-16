package pi

import (
	"strings"
	"testing"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
)

// what a tool call invoked comes from the toolCall block its event recorded:
// a bash command whole, past the summary's cap, with the rest of its
// arguments; any other tool as its arguments, its path first; nothing where
// the event holds no block. a result pi marks as an error is a failed
// action.
func TestInvocationReadsTheRecordedToolCall(t *testing.T) {
	long := strings.Repeat("echo one line of a long command && ", 12) + "echo done"
	m := &message{Role: "assistant", Content: []any{
		map[string]any{"type": "toolCall", "id": "call_1", "name": "bash", "arguments": map[string]any{"command": long, "timeout": 30}},
		map[string]any{"type": "toolCall", "id": "call_2", "name": "edit", "arguments": map[string]any{"oldText": "a", "newText": "b", "path": "main.go"}},
	}}
	evs := mapMessage(m, nil, "e1")
	if len(evs) != 2 || evs[0].Kind != model.EventToolUse || len(evs[0].Text) >= len(long) {
		t.Fatalf("mapped %+v", evs)
	}
	h := New(Config{})
	if inv, ok := h.Invocation(evs[0]); !ok || inv.Tool != "bash" || inv.Command != long || len(inv.Fields) != 1 || inv.Fields[0] != (harness.Field{Name: "timeout", Value: "30"}) {
		t.Errorf("bash: %v %+v", ok, inv)
	}
	inv, ok := h.Invocation(evs[1])
	names := []string{}
	for _, f := range inv.Fields {
		names = append(names, f.Name)
	}
	if !ok || inv.Tool != "edit" || inv.Command != "" || strings.Join(names, ",") != "path,oldText,newText" {
		t.Errorf("edit: %v %+v", ok, inv)
	}
	if _, ok := h.Invocation(&model.Event{Kind: model.EventToolUse, Tool: "bash", Text: "bash ls"}); ok {
		t.Error("an event with no payload says what it invoked")
	}
	failed := mapMessage(&message{Role: "toolResult", ToolCallID: "call_1", ToolName: "bash", Content: "exit 1", IsError: true}, nil, "e2")
	ran := mapMessage(&message{Role: "toolResult", ToolCallID: "call_2", ToolName: "edit", Content: "done"}, nil, "e3")
	if failed[0].Status != model.TurnError || ran[0].Status != "" {
		t.Errorf("result statuses: failed %q, ran %q", failed[0].Status, ran[0].Status)
	}
}
