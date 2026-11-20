//go:build e2e

package pi

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/michaelquigley/df/dd"
)

// sessionHeader is the first line of a session file, as pi 1.5.2 writes it.
type sessionHeader struct {
	Type      string
	Version   int
	ID        string
	Timestamp string
	Cwd       string
}

// PlantSession copies the recorded session file at recorded into pi's folder
// for the checkout cwd, in the sessions root env names, as a new session: a
// fresh id, cwd as its checkout and the time now. it returns the new id, by
// which discovery finds it. pi's session format and layout stay here, in the
// adapter; the caller owns the root, and removes it.
func PlantSession(env []string, cwd, recorded string) (string, error) {
	data, err := os.ReadFile(recorded)
	if err != nil {
		return "", err
	}
	first, rest, _ := bytes.Cut(data, []byte("\n"))
	header := &sessionHeader{}
	if err := dd.BindJSON(header, first); err != nil || header.Type != "session" {
		return "", fmt.Errorf("'%s' does not start with a session header", recorded)
	}
	resolved, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	header.ID, header.Cwd, header.Timestamp = newUUID(), resolved, now.Format(time.RFC3339Nano)
	line, err := dd.UnbindJSONL(header)
	if err != nil {
		return "", err
	}
	dir := checkoutDir(env, cwd)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	// pi names a session's file for its start time and its id.
	name := strings.NewReplacer(":", "-", ".", "-").Replace(now.Format("2006-01-02T15:04:05.000Z")) + "_" + header.ID + ".jsonl"
	if err := os.WriteFile(filepath.Join(dir, name), append(line, rest...), 0o600); err != nil {
		return "", err
	}
	return header.ID, nil
}
