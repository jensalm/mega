package codex

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/michaelquigley/df/dd"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
)

// manifestName is the file a whole codex install keeps beside its programs:
// the release's codex-package archive, and npm's platform package under
// vendor/<target>/, lay it out the same way (capability report,
// 2026-11-14).
const manifestName = "codex-package.json"

// codeModeHost is the helper every package seen keeps beside its entrypoint,
// though the manifest does not name it.
const codeModeHost = "codex-code-mode-host"

// packageManifest is the part of codex-package.json the helper check reads:
// the programs it names, relative to the manifest's directory.
type packageManifest struct {
	LayoutVersion int    `dd:"layoutVersion"`
	Version       string `dd:"version"`
	Entrypoint    string `dd:"entrypoint"`
	ExecServer    string `dd:"execServer"`
}

// helpers checks that the codex mega runs, at path, was installed whole: its
// package's manifest beside it, every program the manifest names, and the
// code-mode host beside the entrypoint. a lone binary, as the release's lone
// codex archive unpacks, runs a turn but cannot run a shell command (codex
// 0.169.0, 2026-11-14), so it is reported as missing its helpers.
func (h *Harness) helpers(path string) error {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = path
	}
	manifest, dir, found := findManifest(resolved)
	if !found {
		return &harness.MissingError{Harness: model.HarnessCodex, Executable: path, Reason: fmt.Sprintf(
			"'codex' at '%s' is a lone binary; install Codex with npm or from the release's codex-package archive, whose helper programs it needs", path)}
	}
	programs := []string{manifest.Entrypoint}
	if manifest.ExecServer != "" {
		programs = append(programs, manifest.ExecServer)
	}
	programs = append(programs, filepath.Join(filepath.Dir(manifest.Entrypoint), codeModeHost))
	for _, program := range programs {
		full := filepath.Join(dir, program)
		if info, err := os.Stat(full); err != nil || info.IsDir() || info.Mode()&0o111 == 0 {
			return &harness.MissingError{Harness: model.HarnessCodex, Executable: full, Reason: fmt.Sprintf(
				"'%s', a helper program the codex at '%s' needs, is missing; reinstall Codex whole, with npm or from the release's codex-package archive", full, path)}
		}
	}
	return nil
}

// findManifest finds the package manifest for the codex at path:
//   - beside a native binary, in its own directory or its parent (the
//     package's bin/), when the manifest's entrypoint is that binary;
//   - for npm's codex.js, in the platform package npm installs beside it,
//     node_modules/@openai/codex-<platform>/vendor/<target>/, nested or
//     hoisted.
func findManifest(path string) (packageManifest, string, bool) {
	for _, dir := range []string{filepath.Dir(path), filepath.Dir(filepath.Dir(path))} {
		m, ok := readManifest(dir)
		if ok && sameFile(filepath.Join(dir, m.Entrypoint), path) {
			return m, dir, true
		}
	}
	if strings.HasSuffix(path, ".js") {
		pkg := filepath.Dir(filepath.Dir(path))
		for _, pattern := range []string{
			filepath.Join(pkg, "node_modules", "@openai", "codex-*", "vendor", "*", manifestName),
			filepath.Join(filepath.Dir(pkg), "codex-*", "vendor", "*", manifestName),
		} {
			matches, _ := filepath.Glob(pattern)
			for _, match := range matches {
				if m, ok := readManifest(filepath.Dir(match)); ok {
					return m, filepath.Dir(match), true
				}
			}
		}
	}
	return packageManifest{}, "", false
}

func readManifest(dir string) (packageManifest, bool) {
	m := packageManifest{}
	if err := dd.BindJSONFile(&m, filepath.Join(dir, manifestName)); err != nil || m.Entrypoint == "" {
		return packageManifest{}, false
	}
	return m, true
}

func sameFile(a, b string) bool {
	ia, err := os.Stat(a)
	if err != nil {
		return false
	}
	ib, err := os.Stat(b)
	return err == nil && os.SameFile(ia, ib)
}

// resolve is where the configured codex is: a path as it is, a bare name on
// find's PATH.
func (h *Harness) resolve(find []string) (string, error) {
	if strings.Contains(h.cfg.Binary, "/") {
		return h.cfg.Binary, nil
	}
	return harness.LookPath(h.cfg.Binary, find)
}

// Version is the version the codex mega would run prints for --version,
// through the same interpreters mega runs it with.
func (h *Harness) Version(ctx context.Context, find []string) (string, error) {
	argv, err := harness.Program(h.cfg.Binary, find)
	if err != nil {
		return "", err
	}
	return harness.ReadVersion(ctx, append(argv, "--version"), find, "codex-cli")
}
