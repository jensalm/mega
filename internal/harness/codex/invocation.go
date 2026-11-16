package codex

import (
	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
)

// Invocation reads what a tool call invoked from the thread item its event
// recorded: a command execution's whole command and where it ran, a file
// change's files by what was done to each, an MCP tool call's server, tool
// and arguments, and a web search's query.
func (h *Harness) Invocation(ev *model.Event) (harness.Invocation, bool) {
	raw, _ := ev.Raw.(map[string]any)
	if raw == nil {
		return harness.Invocation{}, false
	}
	switch raw["type"] {
	case "commandExecution":
		command, _ := raw["command"].(string)
		inv := harness.Invocation{Tool: "shell", Command: command}
		if cwd, _ := raw["cwd"].(string); cwd != "" {
			inv.Fields = []harness.Field{{Name: "cwd", Value: cwd}}
		}
		return inv, command != ""
	case "fileChange":
		inv := harness.Invocation{Tool: "edit"}
		changes, _ := raw["changes"].([]any)
		for _, item := range changes {
			change, _ := item.(map[string]any)
			path, _ := change["path"].(string)
			what := "change"
			if kind, _ := change["kind"].(map[string]any); kind != nil {
				if t, _ := kind["type"].(string); t != "" {
					what = t
				}
			}
			inv.Fields = append(inv.Fields, harness.Field{Name: what, Value: path})
		}
		return inv, len(inv.Fields) > 0
	case "mcpToolCall":
		server, _ := raw["server"].(string)
		tool, _ := raw["tool"].(string)
		inv := harness.Invocation{Tool: server + "." + tool}
		if args, ok := raw["arguments"].(map[string]any); ok {
			inv.Fields = harness.InputFields(args)
		} else if raw["arguments"] != nil {
			inv.Fields = harness.InputFields(map[string]any{"arguments": raw["arguments"]})
		}
		return inv, true
	case "webSearch":
		query, _ := raw["query"].(string)
		return harness.Invocation{Tool: "web search", Fields: []harness.Field{{Name: "query", Value: query}}}, query != ""
	}
	return harness.Invocation{}, false
}
