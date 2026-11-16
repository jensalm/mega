package service

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/harness/harnesstest"
	"github.com/michaelquigley/mega/internal/model"
)

func startedFacts(t *testing.T, r *rig, id string) *model.RuntimeFacts {
	t.Helper()
	for _, ev := range r.events(id) {
		if ev.Kind == model.EventRuntime && ev.Status == model.StatusStarted {
			return ev.Facts
		}
	}
	t.Fatalf("no started event for %s", id)
	return nil
}

// a runtime's start records where its version stands against the tested
// range, for the history's started line: a tested version and a newer one,
// the normal state, with no sentence; an older and an unreadable one with
// it; nothing is refused for its version.
func TestTheStartRecordsWhereTheVersionStands(t *testing.T) {
	r := newRig(t)
	tested := harness.Tested[model.HarnessClaude]
	for _, c := range []struct{ version, standing, says string }{
		{"2.1.320", harness.StandingTested, ""},
		{"2.1.999", harness.StandingNewer, ""},
		{"2.1.100", harness.StandingOlder, "expects Claude Code '" + tested.Low + "' or later"},
		{"fake", harness.StandingUnknown, "could not be read"},
	} {
		conv := r.start("one")
		rt := r.fake.Last()
		rt.Emit(&model.Event{Kind: model.EventRuntime, Status: model.StatusStarted, Facts: &model.RuntimeFacts{Session: rt.NativeID(), Version: c.version}})
		f := startedFacts(t, r, conv.ID)
		if f.Standing != c.standing || f.Tested != tested.String() || (c.says == "") != (f.VersionNote == "") || !strings.Contains(f.VersionNote, c.says) {
			t.Errorf("%s: standing %q, tested %q, note %q", c.version, f.Standing, f.Tested, f.VersionNote)
		}
		if view := r.view(conv.ID); view.Facts.Standing != c.standing {
			t.Errorf("%s: the view's facts stand %q", c.version, view.Facts.Standing)
		}
	}
}

// the version read just before a launch stands in when the harness's own
// start reports none; a read that fails leaves the launch to proceed.
func TestTheVersionReadBeforeALaunchFillsAnEmptyReport(t *testing.T) {
	r := newRig(t)
	r.fake.SetVersion("2.1.100", nil)
	conv := r.start("one")
	rt := r.fake.Last()
	rt.Emit(&model.Event{Kind: model.EventRuntime, Status: model.StatusStarted, Facts: &model.RuntimeFacts{Session: rt.NativeID()}})
	if f := startedFacts(t, r, conv.ID); f.Version != "2.1.100" || f.Standing != harness.StandingOlder {
		t.Errorf("filled %q, standing %q", f.Version, f.Standing)
	}
	r.fake.SetVersion("", errors.New("'claude --version' did not answer within 2s"))
	conv = r.start("two")
	rt = r.fake.Last()
	if rt.Spec.Cwd == "" {
		t.Fatal("the launch did not proceed")
	}
	rt.Emit(&model.Event{Kind: model.EventRuntime, Status: model.StatusStarted, Facts: &model.RuntimeFacts{Session: rt.NativeID()}})
	if f := startedFacts(t, r, conv.ID); f.Version != "" || f.Standing != harness.StandingUnknown {
		t.Errorf("an unreadable version: %q, standing %q", f.Version, f.Standing)
	}
}

// the start-up check reads each harness's version, and the workspace's
// report says where it stands; one that cannot be read is unknown, and the
// harness can still run.
func TestTheStartUpCheckReadsVersions(t *testing.T) {
	claude := harnesstest.New(model.HarnessClaude)
	claude.SetVersion("2.1.330", nil)
	codex := harnesstest.New(model.HarnessCodex)
	codex.SetVersion("", errors.New("printed no version"))
	svc := newService(t, filepath.Join(t.TempDir(), "registry.yaml"), allTrusted, claude, codex)
	got := svc.Workspace().Harnesses
	if len(got) != 2 {
		t.Fatalf("harnesses = %+v", got)
	}
	if c := got[0]; !c.Found || c.Version != "2.1.330" || c.Standing != harness.StandingNewer || c.Tested != harness.Tested[model.HarnessClaude].String() || c.Note != "" {
		t.Errorf("claude = %+v", c)
	}
	if x := got[1]; !x.Found || x.Version != "" || x.Standing != harness.StandingUnknown || !strings.Contains(x.Note, "could not be read") {
		t.Errorf("codex = %+v", x)
	}
	if h := svc.Harnesses(); len(h) != 2 || h[0].Version != "2.1.330" {
		t.Errorf("Harnesses() = %+v", h)
	}
}
