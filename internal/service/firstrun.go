package service

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"

	"github.com/michaelquigley/df/dl"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
)

// resolveAtOnce bounds the resolutions the service runs in the background:
// each is a login shell and an activation, and a workspace of many projects
// resolving all at once at start-up would start that many shells together.
const resolveAtOnce = 4

// HarnessStatus is whether a harness can run in the operator's shell
// environment, as a project with no activation gets it: where its programs
// were found, or why it cannot run. it is what the start-up log says and what
// the empty workspace shows; a project's own environment may differ, and its
// page says what is missing there.
type HarnessStatus struct {
	Name  string
	Found bool
	Where string
	// Reason is why it cannot run, without the harness's name, which the
	// page and the log put before it.
	Reason string
}

// resolveAllLocked starts, in the background, a resolution of every project
// whose environment may run and has none in memory: at start-up, so what each
// is missing is known before anyone opens anything, and again after a restart.
// a project whose file is unapproved, changed or malformed is skipped, as
// resolveStaleLocked skips it: nothing runs an activation the operator has
// not approved. callers hold s.mu.
func (s *Service) resolveAllLocked() {
	doc := s.reg.Snapshot()
	states := make(map[string]EnvState, len(doc.Projects))
	for _, project := range doc.Projects {
		states[project.ID] = s.envStateLocked(project)
	}
	s.resolveStaleLocked(doc.Projects, states)
}

// resolveRegisteredLocked starts the background resolution of a project just
// registered. callers hold s.mu.
func (s *Service) resolveRegisteredLocked(project model.Project) {
	s.resolveStaleLocked([]model.Project{project}, map[string]EnvState{project.ID: s.envStateLocked(project)})
}

// checkHarnesses finds, once at start-up and in the background, where each
// harness's programs are in the operator's shell environment: the shell run
// in the home directory with no activation, as a project without one gets it.
// a shell that cannot be resolved leaves mega's own environment to look in.
// each harness gets one log line, found where or missing and why, and the
// environment is kept, so the empty workspace's report follows an install
// without a restart. callers do not hold s.mu.
func (s *Service) checkHarnesses() {
	home, err := os.UserHomeDir()
	var find []string
	where := "the operator's shell environment"
	if err == nil {
		// it takes a slot as a project's resolution does.
		s.resolveSlots <- struct{}{}
		resolution, resolveErr := s.resolver.Resolve(context.Background(), home, "")
		<-s.resolveSlots
		if resolveErr == nil {
			find = resolution.Before
		} else {
			err = resolveErr
		}
	}
	if find == nil {
		dl.Warnf("the shell environment a project gets could not be resolved, so the harnesses are looked for in mega's own: %v", err)
		find, where = os.Environ(), "mega's own environment"
	}
	s.mu.Lock()
	s.harnessFind, s.harnessWhere = find, where
	report := s.harnessStatusLocked()
	s.changed()
	s.mu.Unlock()
	for _, status := range report {
		if status.Found {
			dl.Infof("harness '%s' can run, in %s: %s", status.Name, where, status.Where)
		} else {
			dl.Warnf("harness '%s' cannot run: %s (looked for in %s)", status.Name, status.Reason, where)
		}
	}
}

// harnessStatusLocked checks each harness, in name order, in the environment
// the start-up check found; nil until it has. a check reads names and modes
// only, so it is cheap enough for every workspace read. callers hold s.mu.
func (s *Service) harnessStatusLocked() []HarnessStatus {
	if s.harnessFind == nil {
		return nil
	}
	names := make([]string, 0, len(s.harnesses))
	for name := range s.harnesses {
		names = append(names, name)
	}
	slices.Sort(names)
	out := make([]HarnessStatus, 0, len(names))
	for _, name := range names {
		h := s.harnesses[name]
		status := HarnessStatus{Name: name}
		var missing *harness.MissingError
		switch err := h.Check(s.harnessFind); {
		case err == nil:
			status.Found = true
			if locator, ok := h.(harness.Locator); ok {
				status.Where = locator.Locate(s.harnessFind)
			}
		case errors.As(err, &missing):
			// the check reads a project's PATH; here it is the shell's own,
			// before any project exists.
			status.Reason = strings.ReplaceAll(missing.Reason, "the project's PATH", "your shell's PATH")
		default:
			status.Reason = err.Error()
		}
		out = append(out, status)
	}
	return out
}
