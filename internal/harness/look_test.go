package harness

import (
	"strings"
	"testing"

	"github.com/michaelquigley/mega/internal/model"
)

// a look takes the first and last prompts and the last reply, skipping blank
// ones, and bounds each: a prompt kept from its start, a reply from its end,
// each at a rune boundary, and says which it cut.
func TestLookOfBoundsEachPart(t *testing.T) {
	long := strings.Repeat("é", LookFirst)
	reply := "start " + strings.Repeat("ü", LookReply) + " END"
	look := LookOf([]*model.Event{
		{Kind: model.EventUser, Text: long},
		{Kind: model.EventAssistant, Text: "a first reply"},
		{Kind: model.EventUser, Text: "   "},
		{Kind: model.EventToolUse, Text: "Bash ls"},
		{Kind: model.EventUser, Text: "the last prompt"},
		{Kind: model.EventAssistant, Text: reply},
		{Kind: model.EventAssistant, Text: ""},
	})
	if len(look.FirstPrompt) > LookFirst || !strings.HasPrefix(long, look.FirstPrompt) || !strings.HasPrefix(look.FirstPrompt, "é") {
		t.Errorf("first: %d bytes", len(look.FirstPrompt))
	}
	if look.LastPrompt != "the last prompt" || look.Prompts != 2 {
		t.Errorf("last: %q, prompts %d", look.LastPrompt, look.Prompts)
	}
	if len(look.LastReply) > LookReply || !strings.HasSuffix(look.LastReply, " END") || strings.HasPrefix(look.LastReply, "start") {
		t.Errorf("reply: %d bytes, %q…", len(look.LastReply), look.LastReply[:8])
	}
	if strings.Join(look.Cut, ",") != "first,reply" {
		t.Errorf("cut: %v", look.Cut)
	}
	if empty := LookOf(nil); empty.Prompts != 0 || empty.FirstPrompt != "" || len(empty.Cut) != 0 {
		t.Errorf("an empty session: %+v", empty)
	}
}
