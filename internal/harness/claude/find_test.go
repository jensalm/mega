package claude

import (
	"context"
	stdErrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelquigley/mega/internal/harness"
)

// the test binary stands in for a native claude code build: run with
// fakeVersionEnv set, it prints that as --version does and exits. claude
// code must be a native build, never a script, so a script would not do.
const fakeVersionEnv = "MEGA_TEST_FAKE_CLAUDE_VERSION"

func TestMain(m *testing.M) {
	if v := os.Getenv(fakeVersionEnv); v != "" {
		fmt.Printf("%s (Claude Code)\n", v)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// nativeBuild links the test binary in as a native build at path.
func nativeBuild(t *testing.T, path string) {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(self, path); err != nil {
		t.Fatal(err)
	}
}

// findEnv is an environment with home as HOME, dir alone on PATH, and the
// fake's version.
func findEnv(home, dir, version string) []string {
	return []string{"HOME=" + home, "PATH=" + dir, fakeVersionEnv + "=" + version}
}

func shimHarness(t *testing.T, executable string) *Harness {
	t.Helper()
	shim := t.TempDir()
	if err := os.WriteFile(filepath.Join(shim, shimScript), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(shim, "node_modules", "@anthropic-ai", "claude-agent-sdk"), 0o700); err != nil {
		t.Fatal(err)
	}
	return New(Config{Node: "/bin/true", Shim: shim, Executable: executable, Where: "in '/c/mega.yaml', the config file mega read"})
}

// the default, the bare name claude, is found on the PATH mega uses, as the
// installer's link; its version is what it prints.
func TestTheBareNameIsFoundOnThePath(t *testing.T) {
	home, bin := t.TempDir(), t.TempDir()
	nativeBuild(t, filepath.Join(bin, "claude"))
	// a newer build in the versions folder is not taken: the link wins.
	nativeBuild(t, filepath.Join(home, ".local", "share", "claude", "versions", "2.1.999"))
	h := shimHarness(t, "claude")
	find := findEnv(home, bin, "2.1.320")
	path, fallback, err := h.locate(find)
	if err != nil || fallback || path != filepath.Join(bin, "claude") {
		t.Fatalf("located %q (fallback %v), %v", path, fallback, err)
	}
	if v, err := h.Version(context.Background(), find); err != nil || v != "2.1.320" {
		t.Errorf("version %q, %v", v, err)
	}
	if !strings.Contains(h.Locate(find), filepath.Join(bin, "claude")) {
		t.Errorf("Locate = %q", h.Locate(find))
	}
}

// with no claude on the PATH, the newest native build in the installer's
// versions folder is used; scripts and names that are not versions are not
// builds.
func TestNoClaudeOnThePathFallsBackToTheNewestBuild(t *testing.T) {
	home := t.TempDir()
	versions := filepath.Join(home, ".local", "share", "claude", "versions")
	for _, v := range []string{"2.1.318", "2.1.325", "2.1.322"} {
		nativeBuild(t, filepath.Join(versions, v))
	}
	if err := os.WriteFile(filepath.Join(versions, "2.1.999"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(versions, "latest"), nil, 0o755); err != nil {
		t.Fatal(err)
	}
	h := shimHarness(t, "claude")
	find := findEnv(home, t.TempDir(), "2.1.325")
	path, fallback, err := h.locate(find)
	if err != nil || !fallback || path != filepath.Join(versions, "2.1.325") {
		t.Fatalf("located %q (fallback %v), %v", path, fallback, err)
	}
	if err := h.Check(find); err != nil {
		t.Errorf("check: %v", err)
	}
	if loc := h.Locate(find); !strings.Contains(loc, "the newest build installed, with no 'claude' on the PATH") {
		t.Errorf("Locate = %q", loc)
	}
	// with none installed either, it is missing, naming where it looked.
	empty := t.TempDir()
	err = shimHarness(t, "claude").Check(findEnv(empty, t.TempDir(), ""))
	if err == nil || !strings.Contains(err.Error(), "'claude' was not found on the project's PATH, and no build is installed in '"+filepath.Join(empty, ".local", "share", "claude", "versions")+"'") || !strings.Contains(err.Error(), "install Claude Code") {
		t.Errorf("none installed: %v", err)
	}
}

// a configured versioned path keeps working while the build exists; once it
// is pruned, the refusal names the builds installed beside it and how to
// follow the installed one, and never says to install claude code; with none
// beside it, it does.
func TestAPrunedVersionedPathNamesTheBuildsBesideIt(t *testing.T) {
	versions := filepath.Join(t.TempDir(), "versions")
	nativeBuild(t, filepath.Join(versions, "2.1.318"))
	nativeBuild(t, filepath.Join(versions, "2.1.322"))
	nativeBuild(t, filepath.Join(versions, "2.1.325"))
	pinned := filepath.Join(versions, "2.1.318")
	h := shimHarness(t, pinned)
	if err := h.Check(os.Environ()); err != nil {
		t.Fatalf("while it exists: %v", err)
	}
	if err := os.Remove(pinned); err != nil {
		t.Fatal(err)
	}
	err := h.Check(os.Environ())
	var missing *harness.MissingError
	if !stdErrors.As(err, &missing) || missing.Executable != pinned {
		t.Fatalf("pruned: %v", err)
	}
	want := fmt.Sprintf("'%s' is not installed; the builds installed beside it are 2.1.325, 2.1.322. to follow the installed Claude Code, set claude.executable to claude in '/c/mega.yaml', the config file mega read, or remove it from there; or name one of those builds", pinned)
	if missing.Reason != want {
		t.Errorf("got  %q\nwant %q", missing.Reason, want)
	}
	if strings.Contains(err.Error(), "install Claude Code") {
		t.Errorf("says to install what is installed: %v", err)
	}
	// with nothing beside it, installing is the advice.
	alone := filepath.Join(t.TempDir(), "versions", "2.1.318")
	if err := shimHarness(t, alone).Check(os.Environ()); err == nil || !strings.Contains(err.Error(), "does not exist; install Claude Code") {
		t.Errorf("nothing beside it: %v", err)
	}
}
