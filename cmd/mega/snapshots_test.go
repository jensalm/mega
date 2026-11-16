package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/snapshot"
)

// the listing names each snapshot by name, time, reason and versions, and
// diff names what differs from the live configuration and prints the copy
// commands, writing nothing; all over a made-up configuration.
func TestSnapshotsListAndDiff(t *testing.T) {
	home := t.TempDir()
	settings := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settings, []byte(`{"effortLevel":"medium"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	store := snapshot.NewStore(snapshot.Options{Dir: filepath.Join(t.TempDir(), "snapshots"), Keep: 5, KeepAtVersion: 5,
		Now: func() time.Time { return time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC) }})
	var out bytes.Buffer
	if err := listSnapshots(&out, store); err != nil || !strings.Contains(out.String(), "no configuration snapshots in '"+store.Dir()+"' yet") {
		t.Errorf("empty: %q %v", out.String(), err)
	}
	roots := []harness.ConfigRoot{{Harness: "claude", Label: "config", Dir: filepath.Join(home, ".claude"), Include: []string{"settings.json"}}}
	res, err := store.Take(roots, map[string]string{"claude": "2.1.325"}, false)
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := listSnapshots(&out, store); err != nil || !strings.Contains(out.String(), res.Manifest.Name+"  ") || !strings.Contains(out.String(), "start-up, 1 files, claude 2.1.325") {
		t.Errorf("listing: %q %v", out.String(), err)
	}
	out.Reset()
	if err := diffSnapshot(&out, store, "newest"); err != nil || !strings.Contains(out.String(), "nothing differs") {
		t.Errorf("unchanged: %q %v", out.String(), err)
	}
	if err := os.WriteFile(settings, []byte(`{"effortLevel":"high"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := diffSnapshot(&out, store, res.Manifest.Name); err != nil {
		t.Fatal(err)
	}
	want := "install -D -m 0644 '" + filepath.Join(store.Dir(), res.Manifest.Name, "claude", "config", "settings.json") + "' '" + settings + "'"
	if !strings.Contains(out.String(), "changed") || !strings.Contains(out.String(), settings) || !strings.Contains(out.String(), want) || !strings.Contains(out.String(), "mega runs none of them") {
		t.Errorf("diff:\n%s\nwant the command %s", out.String(), want)
	}
	if strings.Contains(out.String(), "effortLevel") {
		t.Errorf("diff printed a file's content:\n%s", out.String())
	}
	if body, _ := os.ReadFile(settings); string(body) != `{"effortLevel":"high"}` {
		t.Errorf("diff wrote into the configuration: %s", body)
	}
	if err := diffSnapshot(&out, store, "nope"); err == nil || !strings.Contains(err.Error(), "no snapshot named 'nope'") {
		t.Errorf("an unknown name: %v", err)
	}
}
