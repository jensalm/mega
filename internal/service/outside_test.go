package service

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/michaelquigley/mega/internal/model"
)

// the @ search beyond the checkout lists only what is under a named
// directory, shown with ~, newest first: a path outside every named
// directory is never listed, whatever is typed; dot-directories are
// skipped; a link out of the named directory is not followed or listed; and
// with no directory named, the search is the checkout's as it always was.
func TestTheSearchBeyondTheCheckout(t *testing.T) {
	r := newRig(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	sandbox := filepath.Join(home, "Sandbox")
	at := time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)
	stamp := func(rel string, minutes int) {
		t.Helper()
		write(t, home, rel, rel)
		when := at.Add(time.Duration(minutes) * time.Minute)
		if err := os.Chtimes(filepath.Join(home, rel), when, when); err != nil {
			t.Fatal(err)
		}
	}
	stamp("Sandbox/old.png", 5)
	stamp("Sandbox/shots/newest.png", 30)
	stamp("Sandbox/middle.txt", 10)
	stamp("Sandbox/.cache/hidden.png", 40)
	stamp("Private/secret.txt", 50)
	if err := os.Symlink(filepath.Join(home, "Private", "secret.txt"), filepath.Join(sandbox, "link-out.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(home, "Private"), filepath.Join(sandbox, "dir-out")); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{filepath.Join(sandbox, "shots"), sandbox} {
		if err := os.Chtimes(dir, at, at); err != nil {
			t.Fatal(err)
		}
	}

	// with no directory named, ~/ and / are a checkout search.
	if _, searched, err := r.svc.OutsidePaths(r.project.ID, "~/"); err != nil || searched {
		t.Fatalf("with nothing named: searched %v, %v", searched, err)
	}
	r.svc.outside = outsideSearch{dirs: []string{sandbox}}
	shown := func(q string) []string {
		t.Helper()
		found, searched, err := r.svc.OutsidePaths(r.project.ID, q)
		if err != nil || !searched {
			t.Fatalf("%q: searched %v, %v", q, searched, err)
		}
		var out []string
		for _, p := range found {
			if p.Full != filepath.Join(home, strings.TrimPrefix(strings.TrimSuffix(p.Shown, "/"), "~"))+map[bool]string{true: "/", false: ""}[strings.HasSuffix(p.Shown, "/")] {
				t.Errorf("%q: %s shown for %s", q, p.Shown, p.Full)
			}
			out = append(out, p.Shown)
		}
		return out
	}
	got := shown("~/")
	want := []string{"~/Sandbox/shots/newest.png", "~/Sandbox/middle.txt", "~/Sandbox/old.png", "~/Sandbox/", "~/Sandbox/shots/"}
	if !slices.Equal(got, want) {
		t.Errorf("~/: newest first, no dot-directory, no link out:\n got %v\nwant %v", got, want)
	}
	if got := shown("~/Sa png"); !slices.Equal(got, []string{"~/Sandbox/shots/newest.png", "~/Sandbox/old.png"}) {
		t.Errorf("~/Sa png: %v", got)
	}
	if got := shown(sandbox + "/mid"); !slices.Equal(got, []string{"~/Sandbox/middle.txt"}) {
		t.Errorf("an absolute query: %v", got)
	}
	// outside every named directory: nothing, typed with ~ or from /.
	for _, q := range []string{"~/Private", "~/Private/secret.txt", "/etc/passwd", filepath.Join(home, "Private"), "~/Sandbox/../Private/secret"} {
		if got := shown(q); len(got) != 0 {
			t.Errorf("%q lists outside the named directories: %v", q, got)
		}
	}
	// a relative query is still the checkout's.
	if _, searched, _ := r.svc.OutsidePaths(r.project.ID, "Sandbox"); searched {
		t.Error("a relative query searched the named directories")
	}
	if _, _, err := r.svc.OutsidePaths("no-such-project", "~/"); ErrorCode(err) != CodeNotFound {
		t.Errorf("an unknown project: %v", err)
	}
}

// what an action invoked is read from its recorded event by the adapter,
// when it is expanded: in full where the payload holds it, the summary alone
// where it holds nothing the adapter can read, and refused for an event that
// is not an action.
func TestAnActionsInvocationIsReadFromItsEvent(t *testing.T) {
	r := newRig(t)
	conv := r.start("one")
	rt := r.fake.Last()
	long := strings.Repeat("echo a long command; ", 20)
	rt.Emit(&model.Event{Kind: model.EventToolUse, Tool: "Bash", ToolUseID: "t1", Text: "Bash echo…", Raw: map[string]any{"command": long, "description": "say it"}})
	rt.Emit(&model.Event{Kind: model.EventToolUse, Tool: "Bash", ToolUseID: "t2", Text: "Bash ls"})
	var uses []int64
	var other int64
	for _, ev := range r.events(conv.ID) {
		switch ev.Kind {
		case model.EventToolUse:
			uses = append(uses, ev.Seq)
		case model.EventUser:
			other = ev.Seq
		}
	}
	if len(uses) != 2 || other == 0 {
		t.Fatalf("events: uses %v, other %d", uses, other)
	}
	inv, err := r.svc.Invocation(conv.ID, uses[0])
	if err != nil || !inv.Recorded || inv.Tool != "Bash" || inv.Command != long || len(inv.Fields) != 1 || inv.Fields[0].Value != "say it" {
		t.Errorf("recorded: %+v, %v", inv, err)
	}
	inv, err = r.svc.Invocation(conv.ID, uses[1])
	if err != nil || inv.Recorded || inv.Summary != "Bash ls" || inv.Tool != "Bash" {
		t.Errorf("no payload: %+v, %v", inv, err)
	}
	for _, seq := range []int64{other, 9999} {
		if _, err := r.svc.Invocation(conv.ID, seq); ErrorCode(err) != CodeNotFound {
			t.Errorf("seq %d: %v", seq, err)
		}
	}
}
