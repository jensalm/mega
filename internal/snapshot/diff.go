package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/michaelquigley/mega/internal/harness"
)

// changes between a snapshot and the live configuration.
const (
	ChangeChanged = "changed"
	ChangeGone    = "only in the snapshot"
	ChangeAdded   = "only live"
)

// Difference is one path that differs between a snapshot and the live
// configuration, by name: Live is where it is, or would be, and Restore the
// command that would put the snapshot's copy back, empty when there is
// nothing to put back. Restore is printed for the operator; mega never runs
// it.
type Difference struct {
	Entry   Entry
	Change  string
	Live    string
	Restore string
}

// Diff compares a snapshot with the live configuration at the places it was
// taken from, under the include lists it was taken with. it reads the live
// configuration and writes nothing.
func (s *Store) Diff(m *Manifest) ([]Difference, error) {
	kept := map[string]Entry{}
	for _, e := range m.Files {
		kept[e.key()] = e
	}
	live := map[string]Entry{}
	livePath := map[string]string{}
	for _, r := range m.Roots {
		root := harness.ConfigRoot{Harness: r.Harness, Label: r.Label, Dir: r.Dir, Include: r.Include, Exclude: r.Exclude}
		found, err := scan(root)
		if err != nil {
			return nil, err
		}
		for _, f := range found {
			e, err := liveEntry(root, f)
			if err != nil {
				return nil, err
			}
			live[e.key()] = e
		}
		for _, e := range m.Files {
			if e.Harness == r.Harness && e.Root == r.Label {
				livePath[e.key()] = filepath.Join(r.Dir, filepath.FromSlash(e.Path))
			}
		}
		for _, f := range found {
			livePath[r.Harness+"/"+r.Label+"/"+f.rel] = filepath.Join(r.Dir, filepath.FromSlash(f.rel))
		}
	}
	var out []Difference
	for key, e := range kept {
		l, ok := live[key]
		switch {
		case !ok:
			out = append(out, Difference{Entry: e, Change: ChangeGone, Live: livePath[key], Restore: s.restore(m.Name, e, livePath[key])})
		case !e.same(l):
			out = append(out, Difference{Entry: e, Change: ChangeChanged, Live: livePath[key], Restore: s.restore(m.Name, e, livePath[key])})
		}
	}
	for key, l := range live {
		if _, ok := kept[key]; !ok {
			out = append(out, Difference{Entry: l, Change: ChangeAdded, Live: livePath[key]})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Entry.key() < out[j].Entry.key() })
	return out, nil
}

// restore is the command that puts a snapshot's copy of e back at live, with
// its mode, or empty when the snapshot holds no copy of it.
func (s *Store) restore(name string, e Entry, live string) string {
	switch {
	case e.Skipped != "":
		return ""
	case e.Kind == KindLink:
		return fmt.Sprintf("mkdir -p %s && ln -sfn %s %s", quote(filepath.Dir(live)), quote(e.Target), quote(live))
	default:
		return fmt.Sprintf("install -D -m %04o %s %s", e.Mode, quote(s.path(name, e)), quote(live))
	}
}

// quote is a word the shell reads back as s.
func quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// liveEntry is what a live path holds, as a snapshot would record it.
func liveEntry(root harness.ConfigRoot, f found) (Entry, error) {
	e := Entry{Harness: root.Harness, Root: root.Label, Path: f.rel}
	dir, err := filepath.EvalSymlinks(root.Dir)
	if err != nil {
		return e, err
	}
	src := filepath.Join(dir, filepath.FromSlash(f.rel))
	if f.info.Mode()&fs.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		e.Kind, e.Target = KindLink, target
		return e, err
	}
	e.Kind, e.Mode = KindFile, uint32(f.info.Mode().Perm())
	if f.info.Size() > maxFile {
		e.Skipped = fmt.Sprintf("larger than %d MB", maxFile>>20)
		return e, nil
	}
	in, err := os.OpenFile(src, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return e, err
	}
	defer in.Close()
	h := sha256.New()
	n, err := io.Copy(h, in)
	e.SHA256, e.Size = hex.EncodeToString(h.Sum(nil)), n
	return e, err
}
