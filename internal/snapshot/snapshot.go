// Package snapshot keeps dated copies of each harness's configuration under
// mega's data directory, away from the places the harnesses and their
// updaters manage, so a file an update rewrote or removed can be put back by
// the operator. it only ever reads a harness's configuration.
//
// the store, <data>/snapshots, readable only by the operator:
//
//	objects/<2 hex>/<sha256>          each content once, read-only
//	<name>/manifest.json              what the snapshot holds, by name
//	<name>/<harness>/<label>/<path>   each file a hard link to its object
//
// a file unchanged since an earlier snapshot is the same object, so a
// snapshot whose files are mostly unchanged costs only its directory entries
// and the files that did change. a symbolic link is recorded as a link with
// its target, never followed. a snapshot is written beside the store and
// renamed into place, so a reader never sees half of one.
package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/michaelquigley/df/dd"

	"github.com/michaelquigley/mega/internal/harness"
)

// reasons a snapshot was taken.
const (
	// ReasonStartUp is a snapshot taken at start-up because the
	// configuration differed from the newest one.
	ReasonStartUp = "start-up"
	// ReasonVersion is a snapshot taken because a harness reported a
	// version no snapshot had recorded, the moment a migration may just
	// have run. it is kept longer.
	ReasonVersion = "version"
)

// maxFile bounds one copied file. configuration is small; a larger file in
// an included tree (a skill's bundled asset, say) is recorded as skipped.
const maxFile = 16 << 20

// Options configure a store.
type Options struct {
	// Dir is the store, <data>/snapshots.
	Dir string
	// Keep is how many ordinary snapshots are kept, and KeepAtVersion how
	// many taken at a version change; the newest is always kept.
	Keep          int
	KeepAtVersion int
	// Now is the clock, for tests.
	Now func() time.Time
}

// Store is the snapshot store. it is safe for concurrent use within one
// process.
type Store struct {
	opt Options
	mu  sync.Mutex
}

// NewStore returns the store in opt.Dir; nothing is created until a snapshot is
// taken.
func NewStore(opt Options) *Store {
	if opt.Now == nil {
		opt.Now = time.Now
	}
	return &Store{opt: opt}
}

// Dir is where the store keeps its snapshots.
func (s *Store) Dir() string { return s.opt.Dir }

// Entry is one thing a snapshot holds: a file, by its content's digest, or a
// symbolic link, by its target; or one it skipped, saying why.
type Entry struct {
	Harness string
	Root    string
	Path    string
	Kind    string
	SHA256  string `dd:",+omitempty"`
	Size    int64  `dd:",+omitempty"`
	Mode    uint32 `dd:",+omitempty"`
	Target  string `dd:",+omitempty"`
	Skipped string `dd:",+omitempty"`
}

// kinds of entry.
const (
	KindFile = "file"
	KindLink = "link"
)

func (e Entry) key() string { return e.Harness + "/" + e.Root + "/" + e.Path }

// same reports whether two entries hold the same thing.
func (e Entry) same(o Entry) bool {
	return e.Kind == o.Kind && e.SHA256 == o.SHA256 && e.Mode == o.Mode && e.Target == o.Target && e.Skipped == o.Skipped
}

// RootRecord is where a snapshot's root came from, and what of it was
// included, so it can be compared with the live configuration later.
type RootRecord struct {
	Harness string
	Label   string
	Dir     string
	Include []string
	Exclude []string `dd:",+omitempty"`
}

// Manifest is what a snapshot holds, by name: never a file's content.
type Manifest struct {
	Name     string
	Taken    time.Time
	Reason   string
	Versions map[string]string `dd:",+omitempty"`
	Roots    []RootRecord
	Files    []Entry
}

// path is where an entry's copy sits in its snapshot.
func (s *Store) path(name string, e Entry) string {
	return filepath.Join(s.opt.Dir, name, e.Harness, e.Root, filepath.FromSlash(e.Path))
}

// Path is where an entry's copy sits in the snapshot named name.
func (s *Store) Path(name string, e Entry) string { return s.path(name, e) }

// List is every snapshot, newest first. a folder without a readable
// manifest is not a snapshot and is passed over.
func (s *Store) List() ([]*Manifest, error) {
	entries, err := os.ReadDir(s.opt.Dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []*Manifest
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "objects" || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		m := &Manifest{}
		if err := dd.BindJSONFile(m, filepath.Join(s.opt.Dir, e.Name(), "manifest.json")); err != nil {
			continue
		}
		m.Name = e.Name()
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Taken.Equal(out[j].Taken) {
			return out[i].Taken.After(out[j].Taken)
		}
		return out[i].Name > out[j].Name
	})
	return out, nil
}

// Find is the snapshot named name, or the newest for "newest".
func (s *Store) Find(name string) (*Manifest, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}
	for _, m := range all {
		if m.Name == name || (name == "newest" && m == all[0]) {
			return m, nil
		}
	}
	return nil, fmt.Errorf("no snapshot named '%s' in '%s'", name, s.opt.Dir)
}

// NewVersions is the harnesses whose version in versions no snapshot has
// recorded, among those an earlier snapshot recorded a version for: a
// harness seen for the first time is not a change.
func (s *Store) NewVersions(versions map[string]string) ([]string, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}
	return newVersions(all, versions), nil
}

func newVersions(all []*Manifest, versions map[string]string) []string {
	seen := map[string]map[string]bool{}
	for _, m := range all {
		for h, v := range m.Versions {
			if seen[h] == nil {
				seen[h] = map[string]bool{}
			}
			seen[h][v] = true
		}
	}
	var out []string
	for h, v := range versions {
		if v != "" && seen[h] != nil && !seen[h][v] {
			out = append(out, h)
		}
	}
	sort.Strings(out)
	return out
}

// Result is what Take did: the snapshot it took, if any, and how many of its
// files and bytes were new to the store.
type Result struct {
	Manifest *Manifest
	NewFiles int
	NewBytes int64
}

// Take snapshots the configuration under roots, recording versions, when it
// differs from the newest snapshot, or when a harness reports a version no
// snapshot has recorded (which is taken even if nothing differs). with
// onlyNewVersion, only the second: an ordinary difference waits for the next
// start-up. it returns a nil Manifest when nothing was taken. it reads the
// configuration and writes only under the store.
func (s *Store) Take(roots []harness.ConfigRoot, versions map[string]string, onlyNewVersion bool) (Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all, err := s.List()
	if err != nil {
		return Result{}, err
	}
	reason := ReasonStartUp
	if len(newVersions(all, versions)) > 0 {
		reason = ReasonVersion
	} else if onlyNewVersion {
		return Result{}, nil
	}
	if err := os.MkdirAll(filepath.Join(s.opt.Dir, "objects"), 0o700); err != nil {
		return Result{}, err
	}
	if err := os.Chmod(s.opt.Dir, 0o700); err != nil {
		return Result{}, err
	}
	stage, err := os.MkdirTemp(s.opt.Dir, ".taking-")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(stage)

	m := &Manifest{Taken: s.opt.Now().UTC(), Reason: reason, Versions: map[string]string{}}
	for h, v := range versions {
		if v != "" {
			m.Versions[h] = v
		}
	}
	res := Result{Manifest: m}
	for _, root := range roots {
		m.Roots = append(m.Roots, RootRecord{Harness: root.Harness, Label: root.Label, Dir: root.Dir, Include: root.Include, Exclude: root.Exclude})
		entries, err := scan(root)
		if err != nil {
			return Result{}, err
		}
		for _, e := range entries {
			e, fresh, size, err := s.store(stage, root, e)
			if err != nil {
				return Result{}, err
			}
			if fresh {
				res.NewFiles++
				res.NewBytes += size
			}
			m.Files = append(m.Files, e)
		}
	}
	// a configuration with nothing in it has nothing to keep; and one the
	// same as the newest snapshot's is that snapshot.
	if len(m.Files) == 0 || (reason == ReasonStartUp && len(all) > 0 && sameFiles(all[0].Files, m.Files)) {
		return Result{}, nil
	}
	m.Name = s.name(m)
	if err := dd.UnbindJSONFile(m, filepath.Join(stage, "manifest.json")); err != nil {
		return Result{}, err
	}
	if err := os.Chmod(filepath.Join(stage, "manifest.json"), 0o600); err != nil {
		return Result{}, err
	}
	if err := os.Rename(stage, filepath.Join(s.opt.Dir, m.Name)); err != nil {
		return Result{}, err
	}
	if err := s.prune(); err != nil {
		return res, err
	}
	return res, nil
}

// name is a snapshot's folder: its time in UTC and its reason, made unique.
func (s *Store) name(m *Manifest) string {
	base := m.Taken.Format("20060102T150405Z") + "-" + m.Reason
	name := base
	for i := 2; ; i++ {
		if _, err := os.Lstat(filepath.Join(s.opt.Dir, name)); errors.Is(err, fs.ErrNotExist) {
			return name
		}
		name = fmt.Sprintf("%s-%d", base, i)
	}
}

func sameFiles(a, b []Entry) bool {
	if len(a) != len(b) {
		return false
	}
	byKey := make(map[string]Entry, len(a))
	for _, e := range a {
		byKey[e.key()] = e
	}
	for _, e := range b {
		if o, ok := byKey[e.key()]; !ok || !o.same(e) {
			return false
		}
	}
	return true
}

// found is one path under a root an include list names, before it is read.
type found struct {
	rel  string
	info fs.FileInfo
}

// scan is what the include list names under root, in path order: regular
// files and symbolic links, never followed. the root itself is resolved
// first, so a configuration directory that is itself a link is read where
// it points.
func scan(root harness.ConfigRoot) ([]found, error) {
	dir, err := filepath.EvalSymlinks(root.Dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	excluded := func(rel string) bool {
		for _, x := range root.Exclude {
			if rel == x || strings.HasPrefix(rel, x+"/") {
				return true
			}
		}
		return false
	}
	seen := map[string]bool{}
	var out []found
	add := func(rel string, info fs.FileInfo) {
		if seen[rel] || excluded(rel) {
			return
		}
		if info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
			seen[rel] = true
			out = append(out, found{rel: rel, info: info})
		}
	}
	for _, include := range root.Include {
		var matches []string
		switch {
		case strings.ContainsAny(include, "*?["):
			matches, _ = filepath.Glob(filepath.Join(dir, filepath.FromSlash(include)))
		default:
			matches = []string{filepath.Join(dir, filepath.FromSlash(strings.TrimSuffix(include, "/")))}
		}
		for _, match := range matches {
			info, err := os.Lstat(match)
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, err
			}
			rel, err := filepath.Rel(dir, match)
			if err != nil {
				return nil, err
			}
			if !info.IsDir() {
				add(filepath.ToSlash(rel), info)
				continue
			}
			// a tree: WalkDir reports a link as a link and never descends
			// into one.
			err = filepath.WalkDir(match, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				r, _ := filepath.Rel(dir, path)
				r = filepath.ToSlash(r)
				if d.IsDir() {
					if excluded(r) {
						return filepath.SkipDir
					}
					return nil
				}
				info, err := d.Info()
				if err != nil {
					return err
				}
				add(r, info)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].rel < out[j].rel })
	return out, nil
}

// store puts one found path into the snapshot being staged: a link as a
// link, a file as a hard link to its object, the object written if the
// store does not hold that content yet. fresh says it was written.
func (s *Store) store(stage string, root harness.ConfigRoot, f found) (e Entry, fresh bool, size int64, err error) {
	e = Entry{Harness: root.Harness, Root: root.Label, Path: f.rel}
	dir, err := filepath.EvalSymlinks(root.Dir)
	if err != nil {
		return e, false, 0, err
	}
	src := filepath.Join(dir, filepath.FromSlash(f.rel))
	dst := filepath.Join(stage, root.Harness, root.Label, filepath.FromSlash(f.rel))
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return e, false, 0, err
	}
	if f.info.Mode()&fs.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return e, false, 0, err
		}
		e.Kind, e.Target = KindLink, target
		return e, false, 0, os.Symlink(target, dst)
	}
	e.Kind, e.Mode = KindFile, uint32(f.info.Mode().Perm())
	if f.info.Size() > maxFile {
		e.Skipped = fmt.Sprintf("larger than %d MB", maxFile>>20)
		return e, false, 0, nil
	}
	sum, object, fresh, size, err := s.object(src)
	if err != nil {
		return e, false, 0, err
	}
	e.SHA256, e.Size = sum, size
	if err := os.Link(object, dst); err != nil {
		// a store that cannot hard link keeps a copy instead.
		if err := copyFile(object, dst); err != nil {
			return e, false, 0, err
		}
	}
	return e, fresh, size, nil
}

// object reads src, never through a link, and returns its digest and the
// store's object holding that content, writing it when it is new.
func (s *Store) object(src string) (sum, object string, fresh bool, size int64, err error) {
	in, err := os.OpenFile(src, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return "", "", false, 0, err
	}
	defer in.Close()
	tmp, err := os.CreateTemp(filepath.Join(s.opt.Dir, "objects"), ".new-")
	if err != nil {
		return "", "", false, 0, err
	}
	defer os.Remove(tmp.Name())
	h := sha256.New()
	size, err = io.Copy(io.MultiWriter(tmp, h), io.LimitReader(in, maxFile+1))
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return "", "", false, 0, err
	}
	sum = hex.EncodeToString(h.Sum(nil))
	object = filepath.Join(s.opt.Dir, "objects", sum[:2], sum)
	if _, err := os.Lstat(object); err == nil {
		return sum, object, false, size, nil
	}
	if err := os.MkdirAll(filepath.Dir(object), 0o700); err != nil {
		return "", "", false, 0, err
	}
	if err := os.Chmod(tmp.Name(), 0o400); err != nil {
		return "", "", false, 0, err
	}
	if err := os.Rename(tmp.Name(), object); err != nil {
		return "", "", false, 0, err
	}
	return sum, object, true, size, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o400)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// prune removes the oldest snapshots past the bounds, ordinary and taken at
// a version change counted apart, never the newest, then every object no
// snapshot links to any more.
func (s *Store) prune() error {
	all, err := s.List()
	if err != nil {
		return err
	}
	ordinary, atVersion := 0, 0
	for i, m := range all {
		keep := true
		switch m.Reason {
		case ReasonVersion:
			atVersion++
			keep = atVersion <= s.opt.KeepAtVersion
		default:
			ordinary++
			keep = ordinary <= s.opt.Keep
		}
		if keep || i == 0 {
			continue
		}
		if err := removeAll(filepath.Join(s.opt.Dir, m.Name)); err != nil {
			return err
		}
	}
	return s.collect()
}

// collect removes each object whose only link is its own: no snapshot holds
// it any more.
func (s *Store) collect() error {
	return filepath.WalkDir(filepath.Join(s.opt.Dir, "objects"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat.Nlink <= 1 {
			return os.Remove(path)
		}
		return nil
	})
}

// removeAll removes a snapshot folder, whose files are read-only.
func removeAll(dir string) error {
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			_ = os.Chmod(path, 0o700)
		}
		return nil
	})
	return os.RemoveAll(dir)
}
