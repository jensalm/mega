package harness

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/michaelquigley/mega/internal/model"
)

// Look is how a native session began and how it left off, for the operator
// choosing a session to adopt: the first prompt, the last prompt, and the end
// of the last reply. it is read from the operator's own private transcript,
// shown only in the dialog that asked, and never stored, logged or recorded.
// each part is bounded (LookFirst, LookLast, LookReply), and Cut says which
// were cut short.
type Look struct {
	FirstPrompt string
	LastPrompt  string
	LastReply   string
	// Prompts is how many prompts the session holds.
	Prompts int
	// Cut names the parts cut short: "first", "last" or "reply".
	Cut []string
}

// the bounds on a look's parts, in bytes: the first prompt is kept from its
// start, the last prompt from its start, and the last reply from its end.
const (
	LookFirst = 4000
	LookLast  = 2000
	LookReply = 2000
)

// Looker reads a look at a native session (Look) from the harness's own
// store, changing nothing: no import, no runtime, nothing written.
type Looker interface {
	Look(ctx context.Context, place Place, nativeID string) (Look, error)
}

// LookOf is the look at a session's history, as an adapter maps it to events:
// the first and last prompts and the last reply, each bounded.
func LookOf(evs []*model.Event) Look {
	var look Look
	var first, last, reply string
	for _, ev := range evs {
		switch ev.Kind {
		case model.EventUser:
			if strings.TrimSpace(ev.Text) == "" {
				continue
			}
			if look.Prompts == 0 {
				first = ev.Text
			}
			look.Prompts++
			last = ev.Text
		case model.EventAssistant:
			if strings.TrimSpace(ev.Text) != "" {
				reply = ev.Text
			}
		}
	}
	var cut bool
	if look.FirstPrompt, cut = head(first, LookFirst); cut {
		look.Cut = append(look.Cut, "first")
	}
	if look.Prompts > 1 {
		if look.LastPrompt, cut = head(last, LookLast); cut {
			look.Cut = append(look.Cut, "last")
		}
	}
	if look.LastReply, cut = tail(reply, LookReply); cut {
		look.Cut = append(look.Cut, "reply")
	}
	return look
}

// head is s up to n bytes from its start, cut at a rune boundary.
func head(s string, n int) (string, bool) {
	if len(s) <= n {
		return s, false
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n], true
}

// tail is s's last n bytes at most, cut at a rune boundary.
func tail(s string, n int) (string, bool) {
	if len(s) <= n {
		return s, false
	}
	at := len(s) - n
	for at < len(s) && !utf8.RuneStart(s[at]) {
		at++
	}
	return s[at:], true
}
