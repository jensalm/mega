// fakebuild stands in for a native claude code build in the browser checks:
// built into a scratch home's ~/.local/share/claude/versions/<version>, it
// answers --version with the version its file is named for, as a native
// build does, and refuses everything else. the name is its own file's,
// links followed, so one run through a link such as ~/.local/bin/claude
// says the build the link points at, as the updater's link does. it is a
// compiled program because mega refuses a script as claude code; it never
// runs a conversation, which the checks give to the fake shim.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		self, err := os.Executable()
		if err != nil {
			self = os.Args[0]
		}
		if resolved, err := filepath.EvalSymlinks(self); err == nil {
			self = resolved
		}
		fmt.Printf("%s (Claude Code)\n", filepath.Base(self))
		return
	}
	fmt.Fprintln(os.Stderr, "a stand-in for claude code: it answers --version only")
	os.Exit(2)
}
