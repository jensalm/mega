package pi

import (
	"context"

	"github.com/michaelquigley/mega/internal/harness"
)

// Look reads how a native session began and how it left off from the
// harness's own store, through the same read an import makes, and changes
// nothing: no import, no runtime, nothing written. what it returns is bounded
// (harness.LookOf) and goes only to the dialog that asked.
func (h *Harness) Look(ctx context.Context, place harness.Place, nativeID string) (harness.Look, error) {
	history, err := h.History(ctx, place, nativeID)
	if err != nil {
		return harness.Look{}, err
	}
	return harness.LookOf(history.Events), nil
}
