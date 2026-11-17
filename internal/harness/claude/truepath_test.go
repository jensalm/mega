package claude

import (
	"os/exec"
	"testing"
)

// truePath is where this system keeps the program true. it lives in /bin on
// some systems and /usr/bin on others, so the tests that need a program that
// exists and exits at once find it rather than name a place.
func truePath(t testing.TB) string {
	t.Helper()
	path, err := exec.LookPath("true")
	if err != nil {
		t.Fatalf("this test needs the program true on the PATH: %v", err)
	}
	return path
}
