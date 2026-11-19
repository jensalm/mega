package claude

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/michaelquigley/mega/internal/model"
)

// long is well past the old 4000-byte bound, with multibyte runes throughout,
// so a cut anywhere would show.
var long = strings.Repeat("line of agent output, naïve café ✓\n", 400)

func mapOne(t *testing.T, raw map[string]any) []*model.Event {
	t.Helper()
	evs, err := mapMessage(raw)
	if err != nil {
		t.Fatal(err)
	}
	return evs
}

func TestMessageTextAndToolResultsAreWhole(t *testing.T) {
	if len(long) < 3*4000 {
		t.Fatalf("fixture is only %d bytes", len(long))
	}
	cases := map[string]map[string]any{
		"prompt": {"type": "user", "uuid": "u-1", "message": map[string]any{"content": long}},
		"reply": {"type": "assistant", "uuid": "a-1", "message": map[string]any{"content": []any{
			map[string]any{"type": "text", "text": long},
		}}},
		"tool result": {"type": "user", "uuid": "r-1", "message": map[string]any{"content": []any{
			map[string]any{"type": "tool_result", "tool_use_id": "tu-1", "content": long},
		}}},
	}
	for name, raw := range cases {
		evs := mapOne(t, raw)
		if len(evs) != 1 || evs[0].Text != long {
			t.Errorf("%s: got %d events, text of %d bytes, want %d", name, len(evs), len(evs[0].Text), len(long))
		}
	}
}

func TestToolUseSummaryStaysBounded(t *testing.T) {
	evs := mapOne(t, map[string]any{"type": "assistant", "uuid": "a-2", "message": map[string]any{"content": []any{
		map[string]any{"type": "tool_use", "id": "tu-2", "name": "Bash", "input": map[string]any{"command": long}},
	}}})
	text := evs[0].Text
	if evs[0].Kind != model.EventToolUse || !strings.HasPrefix(text, "Bash line of agent output") || !strings.HasSuffix(text, "…") {
		t.Fatalf("summary = %q", text)
	}
	if len(text) > maxTitle+len("…") || !utf8.ValidString(text) {
		t.Errorf("summary is %d bytes (valid utf-8: %v), bound %d", len(text), utf8.ValidString(text), maxTitle)
	}
}

// the shape getSessionMessages synthesizes for a subagent's hand-back (claude
// code 2.1.318, sdk 0.3.318; capability report, 2026-11-10). the live stream
// never delivers it.
const handBack = `{"type":"user","uuid":"meta-1","session_id":"s-1","parent_tool_use_id":null,"is_meta":true,"isQueuedCommand":true,"message":{"role":"user","content":"<agent-message from=\"a1b2\">\nReport on the fetched page.\n</agent-message>"},"origin":{"kind":"peer","handback":true,"name":"web-fetch","from":"a1b2","senderTaskId":"t-1","body":"Report on the fetched page."}}`

func decode(t *testing.T, frame string) map[string]any {
	t.Helper()
	raw := map[string]any{}
	if err := json.Unmarshal([]byte(frame), &raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestMetaUserMessageIsNativeNotAPrompt(t *testing.T) {
	evs := mapOne(t, decode(t, handBack))
	if len(evs) != 1 || evs[0].Kind != model.EventNative || evs[0].Text != "meta: hand-back from web-fetch" || evs[0].NativeID != "meta-1" || evs[0].Raw == nil {
		t.Fatalf("hand-back mapped to %+v", evs)
	}
	// the transcript's own spelling, and no origin at all.
	plain := decode(t, `{"type":"user","uuid":"meta-2","isMeta":true,"message":{"content":"something claude added"}}`)
	if evs := mapOne(t, plain); evs[0].Kind != model.EventNative || evs[0].Text != "user/meta" {
		t.Errorf("meta without origin mapped to %+v", evs[0])
	}
}

// claude code 2.1.327 echoes a model switch's local command output on the live
// stream, outside any turn, as a user message marked isReplay.
func TestReplayedUserMessageIsNativeNotAPrompt(t *testing.T) {
	frame := `{"type":"user","uuid":"r-1","session_id":"s-1","parent_tool_use_id":null,"isReplay":true,"timestamp":"2026-11-18T14:20:16.053Z","message":{"role":"user","content":"<local-command-stdout>Set model to ` + "`fable (claude-fable-5-2)`" + `</local-command-stdout>"}}`
	evs := mapOne(t, decode(t, frame))
	if len(evs) != 1 || evs[0].Kind != model.EventNative || evs[0].Text != "user/replay" || evs[0].NativeID != "r-1" || evs[0].Raw == nil {
		t.Fatalf("replayed user message mapped to %+v", evs)
	}
}

// claude records the /model a model switch runs in its transcript, never
// sending it live, and other local slash commands and their output the same
// way; the sdk's history reader marks those records isCompletedLocalCommand.
// a marked record is native; a prompt is a prompt whatever its text, even
// one made only of the record's tags.
func TestLocalCommandRecordsAreNativeNotPrompts(t *testing.T) {
	raw := map[string]any{"type": "user", "uuid": "c-1", "isCompletedLocalCommand": true, "message": map[string]any{"role": "user", "content": "<local-command-stdout>Set model to sonnet</local-command-stdout>"}}
	evs := mapOne(t, raw)
	if len(evs) != 1 || evs[0].Kind != model.EventNative || evs[0].Text != "user/local-command" || evs[0].NativeID != "c-1" {
		t.Errorf("a marked record mapped to %+v", evs)
	}
	for _, prompt := range []string{
		"<command-name>/model</command-name>",
		"<command-args>Explain why this failed.</command-args>",
		"<local-command-stdout>Set model to sonnet</local-command-stdout>",
	} {
		evs := mapOne(t, map[string]any{"type": "user", "uuid": "p-1", "message": map[string]any{"content": prompt}})
		if len(evs) != 1 || evs[0].Kind != model.EventUser || evs[0].Text != prompt {
			t.Errorf("the unmarked prompt %.40q mapped to %+v", prompt, evs)
		}
	}
}

// a prompt typed ahead while a turn ran is queued but is the operator's own.
func TestQueuedCommandWithoutMetaIsStillAPrompt(t *testing.T) {
	evs := mapOne(t, decode(t, `{"type":"user","uuid":"q-1","isQueuedCommand":true,"message":{"content":"and then run the tests"}}`))
	if len(evs) != 1 || evs[0].Kind != model.EventUser || evs[0].Text != "and then run the tests" {
		t.Errorf("queued prompt mapped to %+v", evs)
	}
}
