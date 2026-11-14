package harness

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindExecutable(t *testing.T) {
	dir := t.TempDir()
	tool := filepath.Join(dir, "tool")
	if err := os.WriteFile(tool, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	plain := filepath.Join(dir, "plain")
	if err := os.WriteFile(plain, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	for _, ok := range []string{"tool", tool} {
		if err := FindExecutable("h", ok, os.Environ(), ""); err != nil {
			t.Errorf("%s: %v", ok, err)
		}
	}
	for name, reason := range map[string]string{
		"absent":                   "'absent' was not found on the project's PATH",
		filepath.Join(dir, "gone"): "does not exist",
		plain:                      "is not an executable file",
		dir:                        "is not an executable file",
		"":                         "no executable is configured",
	} {
		err := FindExecutable("h", name, os.Environ(), "")
		var missing *MissingError
		if !errors.As(err, &missing) || !errors.Is(err, ErrExecutableMissing) || !strings.Contains(err.Error(), reason) || missing.Executable != name {
			t.Errorf("%q: %v", name, err)
		}
	}
}

// a configured path that cannot run ends its reason with what to do; a bare
// name does not, since the project's PATH is where to look.
func TestFindExecutableAdvisesOnAConfiguredPath(t *testing.T) {
	dir := t.TempDir()
	advice := "install it, or set x.binary in '/c/mega.yaml'"
	gone := filepath.Join(dir, "gone")
	err := FindExecutable("h", gone, os.Environ(), advice)
	if err == nil || !strings.Contains(err.Error(), "'"+gone+"' does not exist; "+advice) {
		t.Errorf("path: %v", err)
	}
	t.Setenv("PATH", dir)
	if err := FindExecutable("h", "absent", os.Environ(), advice); err == nil || strings.Contains(err.Error(), advice) {
		t.Errorf("bare name: %v", err)
	}
}
