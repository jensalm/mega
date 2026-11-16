package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/michaelquigley/push/build"
	"github.com/spf13/cobra"

	"github.com/michaelquigley/mega/internal/config"
	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/harness/claude"
	"github.com/michaelquigley/mega/internal/harness/codex"
	"github.com/michaelquigley/mega/internal/model"
	"github.com/michaelquigley/mega/internal/server"
	"github.com/michaelquigley/mega/ui"
)

func init() {
	rootCmd.AddCommand(versionCmd())
}

// versionCmd is push's version command, followed by where mega logs when it
// serves with no terminal, so the operator can find it.
func versionCmd() *cobra.Command {
	cmd := build.NewVersionCmd("mega")
	run := cmd.Run
	cmd.Run = func(c *cobra.Command, args []string) {
		run(c, args)
		if path, err := logPath(); err == nil {
			fmt.Printf("log file, when serving with no terminal: %s\n", path)
		}
		printHarnesses()
	}
	return cmd
}

// printHarnesses says, for each harness, the range mega was tested with and
// what is found in mega's own environment, by the config mega would read: a
// project's own environment may find another. the service's start-up check
// says the same for the operator's shell, in its log and on the first page.
func printHarnesses() {
	cfg, err := config.Load(configPath, dataPath, listenAddr)
	harnesses := []harness.Harness{}
	if err == nil {
		harnesses = append(harnesses,
			claude.New(claude.Config{Node: cfg.Claude.Node, Shim: cfg.Claude.Shim, Executable: cfg.Claude.Executable, Where: cfg.Where(), FileRead: cfg.FileRead}),
			codex.New(codex.Config{Binary: cfg.Codex.Binary, Where: cfg.Where()}))
	}
	fmt.Println("harnesses, as tested and as found in mega's own environment:")
	for _, name := range []string{model.HarnessClaude, model.HarnessCodex} {
		line := fmt.Sprintf("  %s: tested %s", name, harness.Tested[name])
		for _, h := range harnesses {
			if h.Name() != name {
				continue
			}
			line += "; " + found(h)
		}
		fmt.Println(line)
	}
	if err != nil {
		fmt.Printf("  (the config could not be read, so nothing was looked for: %v)\n", err)
	}
}

// found is what one harness is in mega's own environment: its version and
// where, with how it stands against the tested range, or why it cannot run.
func found(h harness.Harness) string {
	find := os.Environ()
	if err := h.Check(find); err != nil {
		return "not found: " + err.Error()
	}
	where := ""
	if l, ok := h.(harness.Locator); ok {
		where = l.Locate(find)
	}
	version := ""
	if v, ok := h.(harness.Versioner); ok {
		version, _ = v.Version(context.Background(), find)
	}
	standing, note := harness.Stand(h.Name(), version)
	switch standing {
	case harness.StandingTested:
		return fmt.Sprintf("found '%s', %s", version, where)
	case harness.StandingNewer:
		return fmt.Sprintf("found '%s', newer, %s", version, where)
	}
	return fmt.Sprintf("found %s; %s", where, note)
}

// about is what the service answers on /api/version: what 'mega version'
// prints, from the same build information, with where it logs, the Claude
// Agent SDK beside the shim, and the identity of the interface it serves.
func about(claudeHarness *claude.Harness) server.About {
	a := server.About{
		Version:    build.Version,
		Developer:  build.Version == "" || build.Hash == "",
		Go:         runtime.Version(),
		Target:     runtime.GOOS + "/" + runtime.GOARCH,
		LogsToFile: logsToFile,
		Interface:  ui.Build(),
	}
	if a.Developer {
		a.Version = build.DevVersion
	} else {
		a.Commit, a.Built, a.Branch = build.Hash, build.Date, build.Branch
	}
	if build.CGO == "1" {
		a.Go += " (cgo)"
	}
	if path, err := logPath(); err == nil {
		a.LogFile = path
	} else {
		a.LogFile = fmt.Sprintf("unknown: %v", err)
	}
	if version, err := claudeHarness.SDKVersion(); err == nil {
		a.ClaudeSDK = version
	} else {
		a.ClaudeSDK = fmt.Sprintf("unknown: %v", err)
	}
	return a
}
