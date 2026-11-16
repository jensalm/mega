package harness

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/michaelquigley/mega/internal/model"
)

// Invocation is what a tool call invoked, in full: a shell command as the
// command, any other tool as its name and its input. the adapter reads it
// from the tool_use event it recorded, whose payload already holds it, so it
// is never stored a second time.
type Invocation struct {
	// Tool is the tool's name as the harness gave it.
	Tool string
	// Command is a shell command, whole, for a tool that runs one.
	Command string
	// Fields is the rest of the input, each value as text: a string as it
	// is, anything else as indented JSON.
	Fields []Field
}

// Field is one named part of a tool's input.
type Field struct {
	Name  string
	Value string
}

// Invoker says what a recorded tool call invoked. ok is false where the
// event's payload does not hold it, as for an action recorded with no
// payload, and the page then shows what the event has.
type Invoker interface {
	Invocation(ev *model.Event) (Invocation, bool)
}

// InputFields lays a tool's input out to be read: the keys named in first,
// in that order, where present, then the rest by name. a string value is
// kept as it is; anything else is indented JSON.
func InputFields(input map[string]any, first ...string) []Field {
	keys := make([]string, 0, len(input))
	for key := range input {
		if !slices.Contains(first, key) {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)
	var ordered []string
	for _, key := range first {
		if _, ok := input[key]; ok {
			ordered = append(ordered, key)
		}
	}
	ordered = append(ordered, keys...)
	fields := make([]Field, 0, len(ordered))
	for _, key := range ordered {
		fields = append(fields, Field{Name: key, Value: fieldText(input[key])})
	}
	return fields
}

func fieldText(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(data)
}

// ShellInvocation is the invocation of a tool that runs a shell command
// under key: the command, and the rest of its input as fields. a command
// that is missing or not text leaves every key a field.
func ShellInvocation(tool string, input map[string]any, key string) Invocation {
	command, ok := input[key].(string)
	if !ok || command == "" {
		return Invocation{Tool: tool, Fields: InputFields(input)}
	}
	rest := make(map[string]any, len(input))
	for k, v := range input {
		if k != key {
			rest[k] = v
		}
	}
	return Invocation{Tool: tool, Command: command, Fields: InputFields(rest, "description")}
}
