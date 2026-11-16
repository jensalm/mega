package service

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// OutsidePath is a path the @ search found under a directory the operator
// named beyond the checkout: Shown, as the list shows it, with ~ for the home
// directory and a slash after a directory, and Full, the absolute path.
type OutsidePath struct {
	Shown string
	Full  string
}

// outsideSearch lists the directories the operator named for the @ search
// (config path_search.directories), each briefly held as a checkout's
// listing is.
type outsideSearch struct {
	dirs []string

	mu      sync.Mutex
	at      time.Time
	listing []outsideEntry
}

// outsideEntry is one entry of a named directory: its absolute path, a
// slash after a directory, and when it last changed.
type outsideEntry struct {
	full     string
	modified time.Time
}

// searches reports whether an @ word starting with q searches the named
// directories rather than the checkout: it starts with ~/ or /, and at least
// one directory is named. with none, the search is the checkout's, as it
// always was.
func (o *outsideSearch) searches(q string) bool {
	return len(o.dirs) > 0 && (q == "~" || strings.HasPrefix(q, "~/") || strings.HasPrefix(q, "/"))
}

// find lists the entries of the named directories matching q, files newest
// first and then directories, at most limit. q is matched as it is typed: a query starting with ~
// against each path shown with ~, any other against the absolute path,
// fragment by fragment as the checkout's search matches. only what is under
// a named directory is ever listed: the query filters the listing and is
// never read as a path.
func (o *outsideSearch) find(q string, now time.Time, limit int) []OutsidePath {
	home, _ := os.UserHomeDir()
	entries := o.entries(now)
	fragments := strings.FieldsFunc(strings.ToLower(q), func(r rune) bool { return r == '/' || r == ' ' })
	tilde := strings.HasPrefix(q, "~")
	var matched []outsideEntry
	for _, e := range entries {
		against := e.full
		if tilde {
			against = shownPath(home, e.full)
		}
		if ok, _ := matchPath(strings.ToLower(against), fragments); ok {
			matched = append(matched, e)
		}
	}
	// files first, newest first, then directories: adding a file changes its
	// directory's modification time, so a directory would otherwise stand
	// above the file just made, the one nearly always wanted.
	slices.SortStableFunc(matched, func(a, b outsideEntry) int {
		if ad, bd := strings.HasSuffix(a.full, "/"), strings.HasSuffix(b.full, "/"); ad != bd {
			if ad {
				return 1
			}
			return -1
		}
		if c := b.modified.Compare(a.modified); c != 0 {
			return c
		}
		return strings.Compare(a.full, b.full)
	})
	out := make([]OutsidePath, 0, min(limit, len(matched)))
	for _, e := range matched[:min(limit, len(matched))] {
		out = append(out, OutsidePath{Shown: shownPath(home, e.full), Full: e.full})
	}
	return out
}

// entries is the named directories' listing, read again once it is older
// than pathsFresh.
func (o *outsideSearch) entries(now time.Time) []outsideEntry {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.listing == nil || now.Sub(o.at) > pathsFresh {
		o.listing, o.at = listOutside(o.dirs), now
	}
	return o.listing
}

// listOutside walks each named directory: the directory itself, its files
// and directories, skipping directories whose names start with a dot,
// following no symbolic link, and leaving out a link that resolves outside
// the directory it is in, or nowhere. names and times only; it reads no
// file's contents. the walk stops at pathsListed entries in all.
func listOutside(dirs []string) []outsideEntry {
	var out []outsideEntry
	for _, dir := range dirs {
		dir = filepath.Clean(dir)
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		root := resolvedRoot(dir)
		out = append(out, outsideEntry{full: dir + "/", modified: info.ModTime()})
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if d != nil && d.IsDir() && path != dir {
					return fs.SkipDir
				}
				return nil
			}
			if path == dir {
				return nil
			}
			if len(out) >= pathsListed {
				return fs.SkipAll
			}
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}
			rel, relErr := filepath.Rel(dir, path)
			if relErr != nil || !offered(dir, root, filepath.ToSlash(rel)) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			full := path
			if d.IsDir() {
				full += "/"
			}
			out = append(out, outsideEntry{full: full, modified: info.ModTime()})
			return nil
		})
	}
	return out
}

// shownPath is an absolute path as the list shows it: with ~ for the home
// directory when it is under it.
func shownPath(home, full string) string {
	if home != "" && (full == home+"/" || strings.HasPrefix(full, home+"/")) {
		return "~" + strings.TrimPrefix(full, home)
	}
	return full
}

// OutsidePaths searches the directories the operator named beyond the
// checkout, for an @ word starting with ~/ or /, newest first, at most
// PathsLimit. searched is false for any other query, or with no directory
// named, and the checkout's search (Paths) answers it. the directories are
// the operator's own, so they are the same in every project; the project
// must still exist.
func (s *Service) OutsidePaths(projectID, q string) (found []OutsidePath, searched bool, err error) {
	s.mu.Lock()
	_, err = s.projectLocked(projectID)
	s.mu.Unlock()
	if err != nil {
		return nil, false, err
	}
	if !s.outside.searches(q) {
		return nil, false, nil
	}
	return s.outside.find(q, time.Now(), PathsLimit), true, nil
}
