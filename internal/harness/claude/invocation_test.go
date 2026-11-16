package claude

import (
	"strings"
	"testing"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
)

// what a tool call invoked comes from the SDK message its event recorded, in
// full: a Bash command whole, past the summary's cap, with its description;
// any other tool as its input, what it acts on first; the block with the
// event's own id where one message holds two; nothing where the payload
// holds none.
func TestInvocationReadsTheRecordedToolUse(t *testing.T) {
	long := "go test ./... && " + strings.Repeat("echo one line of a long command && ", 12) + "echo done"
	raw := map[string]any{
		"type": "assistant", "uuid": "m-1", "session_id": "s", "parent_tool_use_id": nil,
		"message": map[string]any{"role": "assistant", "content": []any{
			map[string]any{"type": "tool_use", "id": "tu-1", "name": "Bash", "input": map[string]any{"command": long, "description": "run the tests"}},
			map[string]any{"type": "tool_use", "id": "tu-2", "name": "Edit", "input": map[string]any{"new_string": "b", "old_string": "a", "file_path": "/x/main.go", "replace_all": false}},
		}},
	}
	evs, err := mapMessage(raw)
	if err != nil || len(evs) != 2 || evs[0].Kind != model.EventToolUse {
		t.Fatalf("mapped %v, %v", evs, err)
	}
	if len(evs[0].Text) >= len(long) {
		t.Fatalf("the summary was not capped: %d bytes", len(evs[0].Text))
	}
	h := New(Config{})
	bash, ok := h.Invocation(evs[0])
	if !ok || bash.Tool != "Bash" || bash.Command != long || len(bash.Fields) != 1 || bash.Fields[0] != (harness.Field{Name: "description", Value: "run the tests"}) {
		t.Errorf("bash: %v %+v", ok, bash)
	}
	edit, ok := h.Invocation(evs[1])
	names := []string{}
	for _, f := range edit.Fields {
		names = append(names, f.Name)
	}
	if !ok || edit.Tool != "Edit" || edit.Command != "" || strings.Join(names, ",") != "file_path,old_string,new_string,replace_all" || edit.Fields[3].Value != "false" {
		t.Errorf("edit: %v %+v", ok, edit)
	}
	if _, ok := h.Invocation(&model.Event{Kind: model.EventToolUse, Tool: "Bash", ToolUseID: "tu-9", Text: "Bash ls"}); ok {
		t.Error("an event with no payload says what it invoked")
	}
}
