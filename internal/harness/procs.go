package harness

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ProcRoot is where processes are listed, /proc; a test points it at a
// directory that cannot be read, as on a machine with none. there is a
// listing only where /proc is: on macOS there is none yet.
var ProcRoot = "/proc"

// Gone is whether reading a process's entry failed because the process went
// while it was being listed, which is ordinary and leaves nothing unknown.
func Gone(err error) bool {
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ESRCH)
}

// GroupSurvivors waits up to grace for every process in a process group to go
// and returns the pids of those confirmed running, with an error when it
// could not tell: the processes could not be listed, or some could not be
// read, and any of those might be in the group. an adapter that runs its
// harness in a group of its own uses it once the harness has exited.
func GroupSurvivors(pgid int, grace time.Duration) ([]int, error) {
	deadline := time.Now().Add(grace)
	for {
		alive, unread, err := inGroup(pgid)
		if err != nil {
			return nil, err
		}
		if len(alive) == 0 && len(unread) == 0 {
			return nil, nil
		}
		if time.Now().After(deadline) {
			if len(unread) > 0 {
				return alive, fmt.Errorf("listing processes: %d could not be read, and any might be in the group: pids %v", len(unread), unread)
			}
			return alive, nil
		}
		time.Sleep(250 * time.Millisecond)
	}
}

// Leftovers is what an exit adds for the processes a runtime started: those
// confirmed still running (ErrLeftRunning), and that the check could not be
// made, where it could not (ErrUnchecked). nil when the check found nothing.
func Leftovers(pids []int, scan error) error {
	var err error
	if len(pids) > 0 {
		err = errors.Join(err, fmt.Errorf("%w: pids %v", ErrLeftRunning, pids))
	}
	if scan != nil {
		err = errors.Join(err, fmt.Errorf("%w: %v", ErrUnchecked, scan))
	}
	return err
}

// inGroup lists the live processes in a process group and names those that
// are there but could not be read; or it says the processes could not be
// listed at all.
func inGroup(pgid int) ([]int, []int, error) {
	entries, err := os.ReadDir(ProcRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("listing processes: %w", err)
	}
	var out, unread []int
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		data, err := os.ReadFile(filepath.Join(ProcRoot, e.Name(), "stat"))
		if Gone(err) {
			continue
		}
		if err != nil {
			// a process that is there but cannot be read could be in the
			// group: the listing is incomplete there, not clean.
			unread = append(unread, pid)
			continue
		}
		closing := strings.LastIndexByte(string(data), ')')
		if closing < 0 {
			continue
		}
		// after the command: state, ppid, pgrp.
		fields := strings.Fields(string(data[closing+1:]))
		if len(fields) < 3 || fields[0] == "Z" {
			continue
		}
		if g, err := strconv.Atoi(fields[2]); err == nil && g == pgid {
			out = append(out, pid)
		}
	}
	return out, unread, nil
}
