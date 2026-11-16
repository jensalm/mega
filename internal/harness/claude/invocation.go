package claude

import (
	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
)

// Invocation reads what a tool call invoked from the SDK message its event
// recorded: the tool_use block with the event's id, its name and its whole
// input. Bash is a shell command; any other tool is its input, the fields
// that name what it acts on first.
func (h *Harness) Invocation(ev *model.Event) (harness.Invocation, bool) {
	raw, _ := ev.Raw.(map[string]any)
	message, _ := raw["message"].(map[string]any)
	content, _ := message["content"].([]any)
	for _, item := range content {
		block, _ := item.(map[string]any)
		if block["type"] != "tool_use" {
			continue
		}
		if id, _ := block["id"].(string); ev.ToolUseID != "" && id != ev.ToolUseID {
			continue
		}
		name, _ := block["name"].(string)
		input, _ := block["input"].(map[string]any)
		if name == "Bash" {
			return harness.ShellInvocation(name, input, "command"), true
		}
		return harness.Invocation{Tool: name, Fields: harness.InputFields(input, "file_path", "path", "pattern", "url", "query", "old_string", "new_string", "content")}, true
	}
	return harness.Invocation{}, false
}
