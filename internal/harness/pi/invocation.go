package pi

import (
	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
)

// Invocation reads what a tool call invoked from the toolCall block its
// event recorded: its name and its arguments. bash is a shell command; any
// other tool is its arguments, the path it acts on first.
func (h *Harness) Invocation(ev *model.Event) (harness.Invocation, bool) {
	block, _ := ev.Raw.(map[string]any)
	if block["type"] != "toolCall" {
		return harness.Invocation{}, false
	}
	name, _ := block["name"].(string)
	args, _ := block["arguments"].(map[string]any)
	if name == "bash" {
		return harness.ShellInvocation(name, args, "command"), true
	}
	return harness.Invocation{Tool: name, Fields: harness.InputFields(args, "path", "pattern", "oldText", "newText", "content")}, true
}
