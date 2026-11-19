package harness

import "github.com/michaelquigley/mega/internal/model"

// TurnTail follows a history as it is read back, message by message, for
// whether its last turn ended: the one rule every adapter that reads a
// transcript with no turn status uses to fill History.Open. a turn begins at
// a prompt, or, where it has none of its own (a wake for a background job),
// just after the turn before it ended; it ends with a message the harness
// marks as ending it. a message that leaves it open, a tool call or a tool's
// result, keeps it open. bookkeeping is never fed to it.
type TurnTail struct {
	from int  // where the last turn begins in the events
	open bool // the last message seen leaves its turn open
}

// Prompt marks a prompt at index at in the events, which begins a turn that
// has had no reply yet.
func (t *TurnTail) Prompt(at int) {
	t.from, t.open = at, true
}

// Working marks output that leaves the turn open: a tool call, a tool's
// result, a pause.
func (t *TurnTail) Working() {
	t.open = true
}

// Ended marks a message that ends the turn; whatever comes next, from index
// next, begins another.
func (t *TurnTail) Ended(next int) {
	t.from, t.open = next, false
}

// Open names the events of a last turn with no end, from where it began, or
// nothing when it ended.
func (t *TurnTail) Open(evs []*model.Event) map[string]bool {
	open := make(map[string]bool)
	if !t.open {
		return open
	}
	for _, ev := range evs[t.from:] {
		if ev.NativeID != "" {
			open[ev.NativeID] = true
		}
	}
	return open
}
