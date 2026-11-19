package codex

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/michaelquigley/mega/internal/harness"
)

// a child that starts its own session, as codex-linux-sandbox does, is found
// while its parent lives and reported after the parent exits.
func TestDescendantsOutlivingTheirParentAreReported(t *testing.T) {
	cmd := exec.Command("sh", "-c", "setsid sleep 30 & sleep 0.5")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	c := &conn{cmd: cmd}
	var seen []proc
	deadline := time.Now().Add(2 * time.Second)
	for len(seen) < 2 && time.Now().Before(deadline) {
		found, err := c.descendants()
		if err != nil {
			t.Fatal(err)
		}
		seen = found
		time.Sleep(20 * time.Millisecond)
	}
	if len(seen) < 2 {
		t.Fatalf("descendants = %+v", seen)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	alive, err := survivors(seen, 300*time.Millisecond)
	if err != nil || len(alive) != 1 {
		t.Fatalf("survivors = %v from %+v", alive, seen)
	}
	if err := syscall.Kill(alive[0], syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	if left, err := survivors(seen, 2*time.Second); err != nil || len(left) != 0 {
		t.Errorf("still reported after it was killed: %v %v", left, err)
	}
}

// unreadable points the process listing at a directory that does not exist,
// as on a machine with no /proc.
func unreadable(t *testing.T) {
	t.Helper()
	was := procRoot
	procRoot = filepath.Join(t.TempDir(), "no-proc")
	t.Cleanup(func() { procRoot = was })
}

// a listing that cannot be read is an error, never an empty list: nothing
// found is not nothing there.
func TestAnUnreadableListingIsNotAnEmptyOne(t *testing.T) {
	cmd := exec.Command("sleep", "5")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	unreadable(t)
	c := &conn{cmd: cmd}
	if found, err := c.descendants(); err == nil {
		t.Errorf("descendants = %v with no error", found)
	}
	if alive, err := survivors([]proc{{pid: cmd.Process.Pid, start: "1"}}, 0); err == nil {
		t.Errorf("survivors = %v with no error", alive)
	}
}

// fakeProc writes a process's stat into the listing a test points at, with
// its start time; unreadable leaves it there but unreadable.
func fakeProc(t *testing.T, pid int, ppid int, start string, unreadable bool) {
	t.Helper()
	path := filepath.Join(procRoot, strconv.Itoa(pid), "stat")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	fields := []string{"S", strconv.Itoa(ppid)}
	for len(fields) < 19 {
		fields = append(fields, "0")
	}
	fields = append(fields, start)
	mode := os.FileMode(0o644)
	if unreadable {
		mode = 0o000
	}
	if err := os.WriteFile(path, []byte(fmt.Sprintf("%d (sleep) %s", pid, strings.Join(fields, " "))), mode); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(path); unreadable && err == nil {
		t.Skip("an unreadable stat file stays readable here, as it does for root")
	}
}

// a listing read only in part is not a clean one where what could not be
// read matters: a known process that cannot be read could still be running.
// what could be read still counts, so a confirmed survivor is named whatever
// else was unreadable, and a process that went mid-listing, its entry gone,
// is ordinary.
func TestAListingReadInPartIsNotClean(t *testing.T) {
	was := procRoot
	procRoot = t.TempDir()
	t.Cleanup(func() { procRoot = was })
	if err := os.Mkdir(filepath.Join(procRoot, "456"), 0o755); err != nil {
		t.Fatal(err)
	}
	if stats, unread, err := readStats(); err != nil || len(stats) != 0 || len(unread) != 0 {
		t.Fatalf("a process gone mid-listing: %v %v %v", stats, unread, err)
	}
	fakeProc(t, 200, 1, "999", false)
	fakeProc(t, 300, 1, "888", true)
	// a survivor confirmed, beside an unrelated process that cannot be read.
	alive, err := survivors([]proc{{pid: 200, start: "999"}}, 0)
	if err != nil || !slices.Equal(alive, []int{200}) {
		t.Errorf("a readable survivor beside an unrelated unreadable process: %v %v", alive, err)
	}
	// one of ours that cannot be read: not known to have gone.
	alive, err = survivors([]proc{{pid: 200, start: "999"}, {pid: 300, start: "888"}}, 0)
	if err == nil || !slices.Equal(alive, []int{200}) {
		t.Errorf("a known process that cannot be read: %v %v", alive, err)
	}
}

// exitSink records how the runtime exited.
type exitSink struct {
	recordingSink
	exit chan error
}

func (s *exitSink) Exited(err error) { s.exit <- err }

// a release where the processes cannot be listed says so: not a clean stop,
// and not processes left running.
func TestAReleaseThatCouldNotLookSaysSo(t *testing.T) {
	h := fakeHarness(t, "start")
	sink := &exitSink{exit: make(chan error, 1)}
	rt, err := h.Launch(harness.Spec{Env: os.Environ(), Find: os.Environ(), Cwd: t.TempDir(), NativeID: "t-1", Idle: true}, sink)
	if err != nil {
		t.Fatal(err)
	}
	unreadable(t)
	if err := rt.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case exit := <-sink.exit:
		if !errors.Is(exit, harness.ErrUnchecked) || errors.Is(exit, harness.ErrLeftRunning) {
			t.Errorf("exit = %v", exit)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("no exit")
	}
}
