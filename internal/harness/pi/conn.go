package pi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
)

// errClosed fails a command whose pi process exited before answering.
var errClosed = errors.New("the pi process exited")

// handler receives what pi sends besides command responses: session events
// and extension UI requests, in order, on the connection's reader goroutine.
type handler interface {
	frame(f map[string]any)
}

// conn is one `pi --mode rpc` child speaking JSONL on stdio: commands in,
// responses, events and extension UI requests out (pi 1.5.2, docs/rpc.md).
// it runs in its own process group, so a terminal interrupt aimed at the
// service does not reach it, and so what it starts can be ended with it.
type conn struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser

	mu      sync.Mutex
	nextID  int64
	calls   map[string]chan reply
	closed  bool
	inClose bool // input closed: nothing more is written

	done chan struct{} // closed once the process has exited and been waited
	err  error         // the exit, set before done closes
}

// reply is a command's response: its data, or why it failed.
type reply struct {
	data map[string]any
	err  error
}

// start runs pi in the place's checkout with the project's environment and
// nothing of mega's own. pi, a #!/usr/bin/env node script, and its node are
// located in place.Find, the environment before the activation, so the
// activation never decides which program mega runs.
func start(binary string, place harness.Place, args []string, h handler) (*conn, error) {
	argv, err := harness.Program(binary, place.Find)
	if err != nil {
		var missing *harness.MissingError
		if errors.As(err, &missing) {
			missing.Harness = model.HarnessPi
		}
		return nil, err
	}
	cmd := exec.Command(argv[0], append(argv[1:], args...)...)
	cmd.Dir = place.Cwd
	// never nil, which exec would fill with mega's own environment.
	cmd.Env = append([]string{}, place.Env...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting pi: %w", err)
	}
	c := &conn{cmd: cmd, stdin: stdin, calls: make(map[string]chan reply), done: make(chan struct{})}
	go logStderr(cmd.Process.Pid, stderr)
	go c.read(stdout, h)
	return c, nil
}

// call sends a command and waits for its response, binding its data into
// out, which may be nil.
func (c *conn) call(ctx context.Context, command map[string]any, out any) error {
	id, ch, err := c.send(command)
	if err != nil {
		return err
	}
	select {
	case r := <-ch:
		if r.err != nil {
			return fmt.Errorf("%s: %w", command["type"], r.err)
		}
		if out == nil || r.data == nil {
			return nil
		}
		return dd.Bind(out, r.data)
	case <-ctx.Done():
		c.forget(id)
		return fmt.Errorf("%s: %w", command["type"], ctx.Err())
	}
}

// send writes a command with a fresh id and returns the channel its
// response arrives on, without waiting for it.
func (c *conn) send(command map[string]any) (string, chan reply, error) {
	c.mu.Lock()
	if c.closed || c.inClose {
		c.mu.Unlock()
		return "", nil, errClosed
	}
	c.nextID++
	id := "m" + strconv.FormatInt(c.nextID, 10)
	ch := make(chan reply, 1)
	c.calls[id] = ch
	c.mu.Unlock()
	frame := map[string]any{"id": id}
	for k, v := range command {
		frame[k] = v
	}
	if err := c.write(frame); err != nil {
		c.forget(id)
		return "", nil, err
	}
	return id, ch, nil
}

func (c *conn) forget(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.calls, id)
}

// write sends one record, terminated by LF. a command's fields vary with
// its type, so it is a map, which encoding/json writes and dd does not.
func (c *conn) write(frame map[string]any) error {
	line, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	line = append(line, '\n')
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || c.inClose {
		return errClosed
	}
	if _, err := c.stdin.Write(line); err != nil {
		return fmt.Errorf("writing to pi: %w", err)
	}
	return nil
}

// closeInput ends pi's input: its request for an orderly shutdown.
func (c *conn) closeInput() error {
	c.mu.Lock()
	if c.inClose {
		c.mu.Unlock()
		return nil
	}
	c.inClose = true
	c.mu.Unlock()
	return c.stdin.Close()
}

// signal sends sig to pi itself, or with group to its whole process group.
func (c *conn) signal(sig syscall.Signal, group bool) {
	pid := c.cmd.Process.Pid
	if group {
		pid = -pid
	}
	_ = syscall.Kill(pid, sig)
}

// read splits pi's stdout on LF only, as pi's framing requires: a JSON string
// may hold U+2028 or U+2029, which a generic line reader would split on.
func (c *conn) read(stdout io.Reader, h handler) {
	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadBytes('\n')
		if trimmed := strings.TrimRight(string(line), "\r\n"); strings.TrimSpace(trimmed) != "" {
			c.dispatch([]byte(trimmed), h)
		}
		if err != nil {
			break
		}
	}
	waitErr := c.cmd.Wait()
	c.mu.Lock()
	c.closed = true
	for id, ch := range c.calls {
		ch <- reply{err: errClosed}
		delete(c.calls, id)
	}
	c.err = waitErr
	c.mu.Unlock()
	close(c.done)
}

func (c *conn) dispatch(line []byte, h handler) {
	// a record is a map whose shape varies with its type; encoding/json
	// gives its numbers as float64, which dd binds from as it does a file's.
	var frame map[string]any
	if err := json.Unmarshal(line, &frame); err != nil {
		dl.Warnf("unreadable record from pi %d: %v", c.cmd.Process.Pid, err)
		return
	}
	if frame["type"] == "response" {
		id, _ := frame["id"].(string)
		c.mu.Lock()
		ch, ok := c.calls[id]
		delete(c.calls, id)
		c.mu.Unlock()
		if !ok {
			return
		}
		if success, _ := frame["success"].(bool); !success {
			message, _ := frame["error"].(string)
			if message == "" {
				message = "pi refused the command"
			}
			ch <- reply{err: errors.New(message)}
			return
		}
		data, _ := frame["data"].(map[string]any)
		ch <- reply{data: data}
		return
	}
	h.frame(frame)
}

func logStderr(pid int, stderr io.Reader) {
	scanner := bufio.NewScanner(stderr)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			dl.Debugf("pi %d: %s", pid, line)
		}
	}
}
