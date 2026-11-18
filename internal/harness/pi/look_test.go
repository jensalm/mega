package pi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// a look at a recorded pi session, from its file where pi keeps it: the
// first prompt, the last, and the last reply, read with nothing written.
func TestLookReadsARecordedSession(t *testing.T) {
	p := place(t)
	data, err := os.ReadFile(filepath.Join("testdata", "p1-and-p4.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	header := struct{ ID string }{}
	if err := json.Unmarshal([]byte(strings.SplitN(string(data), "\n", 2)[0]), &header); err != nil || header.ID == "" {
		t.Fatalf("header: %v", err)
	}
	dir := checkoutDir(p.Env, p.Cwd)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "2026-11-16T00-00-00-000Z_"+header.ID+".jsonl")
	if err := os.WriteFile(file, data, 0o600); err != nil {
		t.Fatal(err)
	}
	before, _ := os.Stat(file)
	time.Sleep(10 * time.Millisecond)
	look, err := New(Config{}).Look(t.Context(), p, header.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(look.FirstPrompt, "ALPHA-1a2b") || !strings.Contains(look.LastPrompt, "ZETA-1b2c") || look.LastReply != "ZETA-1b2c" || look.Prompts != 7 || len(look.Cut) != 0 {
		t.Errorf("look: %+v", look)
	}
	after, _ := os.Stat(file)
	entries, _ := os.ReadDir(dir)
	if !after.ModTime().Equal(before.ModTime()) || after.Size() != before.Size() || len(entries) != 1 {
		t.Error("the look wrote to pi's store")
	}
}
