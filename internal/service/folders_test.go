package service

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// the register dialog's folder picker lists folders only, by name, inside
// the projects root: dot-folders left out, a link leading out of the root
// not followed or listed, a git checkout and a registered folder marked, the
// way up stopping at the root, anything outside it refused however it is
// reached, and the listing bounded and saying so. a path typed in full is
// still registered wherever it is, ~ included.
func TestTheFolderPicker(t *testing.T) {
	r := newRig(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, "Repos")
	outside := filepath.Join(home, "Private")
	for _, dir := range []string{"Repos/alpha/.git", "Repos/beta", "Repos/.hidden", "Repos/alpha/inner", "Private/secret"} {
		if err := os.MkdirAll(filepath.Join(home, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write(t, root, "notes.txt", "a file, not a folder")
	for link, target := range map[string]string{"link-in": filepath.Join(root, "alpha"), "link-out": outside} {
		if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := r.svc.RegisterProject("Beta", "~/Repos/beta"); err != nil {
		t.Fatalf("registering with ~: %v", err)
	}
	r.svc.projectsRoot = root

	top, err := r.svc.Folders("")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range top.Entries {
		names = append(names, fmt.Sprintf("%s git=%v reg=%q", f.Name, f.Git, f.Registered))
	}
	want := []string{`alpha git=true reg=""`, `beta git=false reg="Beta"`, `link-in git=true reg=""`}
	if fmt.Sprint(names) != fmt.Sprint(want) || top.Path != "~/Repos" || top.Root != "~/Repos" || top.Parent != "" || top.Stopped {
		t.Errorf("the root:\n got %v %+v\nwant %v", names, top, want)
	}
	in, err := r.svc.Folders("~/Repos/alpha")
	if err != nil || in.Parent != "~/Repos" || len(in.Entries) != 1 || in.Entries[0].Name != "inner" {
		t.Errorf("a folder inside: %+v, %v", in, err)
	}
	// outside the root, however it is reached: refused.
	for _, path := range []string{home, outside, "~/Private", filepath.Join(root, "link-out"), filepath.Join(root, "alpha", "..", "..", "Private"), "/etc", "relative/path", "~/Repos/nowhere"} {
		got, err := r.svc.Folders(path)
		if err != nil || got.Note == "" || len(got.Entries) != 0 || got.Parent != "" {
			t.Errorf("%q lists something or fails: %+v, %v", path, got, err)
		}
	}
	// a path typed in full is registered wherever it is.
	if _, err := r.svc.RegisterProject("", outside); err != nil {
		t.Errorf("a path outside the root, typed in full: %v", err)
	}
	// with no root configured, the home directory is the root.
	r.svc.projectsRoot = ""
	if mine, err := r.svc.Folders(""); err != nil || mine.Path != "~" || mine.Parent != "" {
		t.Errorf("the default root: %+v, %v", mine, err)
	}

	// the bound: a folder with more folders than it lists says it stopped.
	many := filepath.Join(root, "many")
	for i := 0; i <= foldersListed; i++ {
		if err := os.MkdirAll(filepath.Join(many, fmt.Sprintf("f%04d", i)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	r.svc.projectsRoot = root
	bounded, err := r.svc.Folders(many)
	if err != nil || !bounded.Stopped || len(bounded.Entries) != foldersListed {
		t.Errorf("the bound: %d entries, stopped %v, %v", len(bounded.Entries), bounded.Stopped, err)
	}
}
