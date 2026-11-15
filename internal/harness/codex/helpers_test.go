package codex

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"

	"github.com/michaelquigley/mega/internal/harness"
)

// program writes an executable file at path, making its directories.
func program(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

// codexPackage lays out a codex package in dir, as the release's
// codex-package archive and npm's vendor/<target>/ do: the manifest, the
// entrypoint and the code-mode host, and the exec server when execServer.
func codexPackage(t *testing.T, dir string, execServer bool) string {
	t.Helper()
	manifest := `{"layoutVersion":1,"version":"0.169.0","target":"x86_64-unknown-linux-musl","variant":"codex","entrypoint":"bin/codex","resourcesDir":"codex-resources","pathDir":"codex-path"`
	if execServer {
		manifest += `,"execServer":"bin/codex-exec-server"`
		program(t, filepath.Join(dir, "bin", "codex-exec-server"), "#!/bin/sh\n")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, manifestName), []byte(manifest+"}"), 0o644); err != nil {
		t.Fatal(err)
	}
	program(t, filepath.Join(dir, "bin", "codex"), "#!/bin/sh\necho codex-cli 0.169.0\n")
	program(t, filepath.Join(dir, "bin", codeModeHost), "#!/bin/sh\n")
	return filepath.Join(dir, "bin", "codex")
}

func missingFrom(t *testing.T, binary string) *harness.MissingError {
	t.Helper()
	err := New(Config{Binary: binary}).Check(os.Environ())
	var missing *harness.MissingError
	if !errors.As(err, &missing) {
		t.Fatalf("%s: %v", binary, err)
	}
	return missing
}

// a codex installed whole passes: the release's package as it unpacks, with
// or without an exec server (0.163.0 had none).
func TestAWholePackagePasses(t *testing.T) {
	for _, execServer := range []bool{true, false} {
		bin := codexPackage(t, t.TempDir(), execServer)
		if err := New(Config{Binary: bin}).Check(os.Environ()); err != nil {
			t.Errorf("exec server %v: %v", execServer, err)
		}
	}
}

// a lone binary, as the release's lone codex archive unpacks, runs a turn
// but cannot run a shell command (0.169.0, 2026-11-14): it is reported as
// missing its helpers, with what to do.
func TestALoneBinaryIsMissingItsHelpers(t *testing.T) {
	lone := filepath.Join(t.TempDir(), "codex-x86_64-unknown-linux-musl")
	program(t, lone, "#!/bin/sh\necho codex-cli 0.169.0\n")
	m := missingFrom(t, lone)
	want := "'codex' at '" + lone + "' is a lone binary; install Codex with npm or from the release's codex-package archive, whose helper programs it needs"
	if m.Reason != want || m.Executable != lone {
		t.Errorf("got  %q\nwant %q", m.Reason, want)
	}
}

// a manifest naming a program that is not there, and a missing code-mode
// host, each name the missing program.
func TestAMissingHelperIsNamed(t *testing.T) {
	dir := t.TempDir()
	bin := codexPackage(t, dir, true)
	if err := os.Remove(filepath.Join(dir, "bin", "codex-exec-server")); err != nil {
		t.Fatal(err)
	}
	m := missingFrom(t, bin)
	if m.Executable != filepath.Join(dir, "bin", "codex-exec-server") || !strings.Contains(m.Reason, "a helper program the codex at '"+bin+"' needs, is missing; reinstall Codex whole") {
		t.Errorf("exec server: %+v", m)
	}
	dir = t.TempDir()
	bin = codexPackage(t, dir, false)
	if err := os.Remove(filepath.Join(dir, "bin", codeModeHost)); err != nil {
		t.Fatal(err)
	}
	if m := missingFrom(t, bin); m.Executable != filepath.Join(dir, "bin", codeModeHost) {
		t.Errorf("code-mode host: %+v", m)
	}
}

// npm's codex.js finds its platform package's manifest, whether npm nested
// it under the codex package or hoisted it beside it.
func TestNpmsLayoutPasses(t *testing.T) {
	if _, err := harness.LookPath("node", os.Environ()); err != nil {
		t.Skip("node is not on PATH")
	}
	for _, hoisted := range []bool{false, true} {
		prefix := t.TempDir()
		pkg := filepath.Join(prefix, "lib", "node_modules", "@openai", "codex")
		program(t, filepath.Join(pkg, "bin", "codex.js"), "#!/usr/bin/env node\n")
		platform := filepath.Join(pkg, "node_modules", "@openai", "codex-linux-x64")
		if hoisted {
			platform = filepath.Join(prefix, "lib", "node_modules", "@openai", "codex-linux-x64")
		}
		codexPackage(t, filepath.Join(platform, "vendor", "x86_64-unknown-linux-musl"), false)
		link := filepath.Join(prefix, "bin", "codex")
		if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(pkg, "bin", "codex.js"), link); err != nil {
			t.Fatal(err)
		}
		if err := New(Config{Binary: link}).Check(os.Environ()); err != nil {
			t.Errorf("hoisted %v: %v", hoisted, err)
		}
	}
	// an npm codex.js whose platform package is not there is a lone script.
	prefix := t.TempDir()
	js := filepath.Join(prefix, "lib", "node_modules", "@openai", "codex", "bin", "codex.js")
	program(t, js, "#!/usr/bin/env node\n")
	if m := missingFrom(t, js); !strings.Contains(m.Reason, "is a lone binary") {
		t.Errorf("no platform package: %+v", m)
	}
}

// the fake codex the browser checks run is a whole package too, and says
// its version as codex-cli does.
func TestTheFakeIsWholeAndSaysItsVersion(t *testing.T) {
	_, file, _, _ := goruntime.Caller(0)
	fake := filepath.Join(filepath.Dir(file), "fake", "codex")
	h := New(Config{Binary: fake})
	if err := h.Check(os.Environ()); err != nil {
		t.Fatalf("check: %v", err)
	}
	if v, err := h.Version(context.Background(), os.Environ()); err != nil || v != "0.166.0" {
		t.Errorf("version %q, %v", v, err)
	}
}
