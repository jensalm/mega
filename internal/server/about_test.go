package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/michaelquigley/mega/internal/harness"
)

// testAbout is the build every test server answers for.
var testAbout = About{
	Version:   "v0.0.x",
	Developer: true,
	Go:        "go1.27.1",
	Target:    "linux/amd64",
	LogFile:   "/home/someone/.local/state/mega/mega.log",
	ClaudeSDK: "0.3.318",
	Interface: "/assets/index-a.js /assets/index-b.css",
	Name:      "made-up-host",
	NameFrom:  "hostname",
}

func TestVersionOnTheWire(t *testing.T) {
	f := start(t)
	status, body := f.do(t, http.MethodGet, "/api/version", "")
	if status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	want := map[string]any{
		"version":      "v0.0.x",
		"developer":    true,
		"go":           "go1.27.1",
		"target":       "linux/amd64",
		"log_file":     "/home/someone/.local/state/mega/mega.log",
		"logs_to_file": false,
		"claude_sdk":   "0.3.318",
		"interface":    "/assets/index-a.js /assets/index-b.css",
		"name":         "made-up-host",
		"name_from":    "hostname",
	}
	for key, value := range want {
		if body[key] != value {
			t.Errorf("%s: %v, want %v", key, body[key], value)
		}
	}
	// the tested ranges, from the one place they are kept, and what the
	// start-up check found (the fake harness reads no version).
	tested, _ := body["tested"].(map[string]any)
	if tested["claude"] != harness.Tested["claude"].String() || tested["codex"] != harness.Tested["codex"].String() {
		t.Errorf("tested = %v", body["tested"])
	}
	deadline := time.Now().Add(5 * time.Second)
	for body["harnesses"] == nil && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
		_, body = f.do(t, http.MethodGet, "/api/version", "")
	}
	if list, _ := body["harnesses"].([]any); len(list) != 1 || list[0].(map[string]any)["name"] != "claude" || list[0].(map[string]any)["tested"] != harness.Tested["claude"].String() {
		t.Errorf("harnesses = %v", body["harnesses"])
	}
	for _, key := range []string{"commit", "built", "branch"} {
		if _, ok := body[key]; ok {
			t.Errorf("%s present on a developer build", key)
		}
	}
}

func TestWorkspaceReadCarriesTheInterfaceBuild(t *testing.T) {
	f := start(t)
	resp, err := http.Get(f.url + "/api/workspace")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get(BuildHeader); got != testAbout.Interface {
		t.Errorf("%s: %q", BuildHeader, got)
	}
}
