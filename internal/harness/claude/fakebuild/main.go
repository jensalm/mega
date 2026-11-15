// fakebuild stands in for a native claude code build in the browser checks:
// built into a scratch home's ~/.local/share/claude/versions/<version>, it
// answers --version with the version its file is named for, as a native
// build does, and refuses everything else. it is a compiled program because
// mega refuses a script as claude code; it never runs a conversation, which
// the checks give to the fake shim.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Printf("%s (Claude Code)\n", filepath.Base(os.Args[0]))
		return
	}
	fmt.Fprintln(os.Stderr, "a stand-in for claude code: it answers --version only")
	os.Exit(2)
}
