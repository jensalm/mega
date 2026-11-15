package server

import (
	"net/http"

	"github.com/michaelquigley/mega/internal/harness"
)

// BuildHeader carries the identity of the interface this service serves on
// every workspace read, so an open page can compare it with its own and say
// when a newer mega is running.
const BuildHeader = "Mega-Interface-Build"

// About is what the service says of its own build: what 'mega version'
// prints, where it logs, the Claude Agent SDK the shim runs, and the
// interface it serves. it is fixed for the service's life and is not
// workspace state.
type About struct {
	// Version is the stamped version, or the developer base version.
	Version string
	// Developer is an unstamped developer build.
	Developer bool
	// Commit, Built and Branch are stamped at build time; absent otherwise.
	Commit string `dd:",+omitempty"`
	Built  string `dd:",+omitempty"`
	Branch string `dd:",+omitempty"`
	// Go is the toolchain it was built with, and Target its OS and
	// architecture.
	Go     string
	Target string
	// LogFile is where mega logs when it serves with no terminal; LogsToFile
	// is whether this one does.
	LogFile    string
	LogsToFile bool
	// ClaudeSDK is the Agent SDK version installed beside the shim, or why it
	// could not be read.
	ClaudeSDK string
	// Interface is the identity of the embedded interface (ui.Build).
	Interface string
}

// versionView is About, with each harness's tested range and the version
// found, as the start-up check of the harnesses read it.
type versionView struct {
	About
	// Tested is each harness's tested range, by name (harness.Tested).
	Tested map[string]string
	// Harnesses is what the start-up check found; absent until it has run.
	Harnesses []*harnessStatusView `dd:",+omitempty"`
}

// getVersion answers from the build information the server was given, with
// the tested ranges and the versions the service has found.
func (s *Server) getVersion(w http.ResponseWriter, r *http.Request) {
	view := &versionView{About: s.about, Tested: map[string]string{}}
	for name, r := range harness.Tested {
		view.Tested[name] = r.String()
	}
	for _, h := range s.svc.Harnesses() {
		view.Harnesses = append(view.Harnesses, newHarnessStatusView(h))
	}
	writeJSON(w, r, http.StatusOK, view)
}
