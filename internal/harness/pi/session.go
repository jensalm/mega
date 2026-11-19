package pi

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/michaelquigley/df/dd"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
)

// maxTitle bounds a tool-use summary line, in bytes.
const maxTitle = 200

// agentDir is where pi keeps its configuration and sessions: the directory
// PI_CODING_AGENT_DIR names, else ~/.pi/agent, under the home env names
// (pi 1.5.2, docs/environment-variables.md).
func agentDir(env []string) string {
	if dir, ok := harness.Lookup(env, "PI_CODING_AGENT_DIR"); ok && dir != "" {
		return dir
	}
	return filepath.Join(harness.Home(env), ".pi", "agent")
}

// sessionsDir is where pi keeps session files: PI_CODING_AGENT_SESSION_DIR,
// else sessions/ in the agent directory. pi's sessionDir setting, which
// moves them too, is not read here.
func sessionsDir(env []string) string {
	if dir, ok := harness.Lookup(env, "PI_CODING_AGENT_SESSION_DIR"); ok && dir != "" {
		return dir
	}
	return filepath.Join(agentDir(env), "sessions")
}

// checkoutDir is the folder pi keeps a checkout's sessions in: the checkout's
// resolved path without its leading slash, with every '/', '\' and ':' a
// dash, between double dashes (pi 1.5.2, docs/session-format.md).
func checkoutDir(env []string, cwd string) string {
	resolved, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		resolved = cwd
	}
	name := strings.NewReplacer("/", "-", `\`, "-", ":", "-").Replace(strings.TrimPrefix(resolved, "/"))
	return filepath.Join(sessionsDir(env), "--"+name+"--")
}

// sessionFile finds the file of the session id: in the checkout's own folder
// first, then in any checkout's, since a session id is unique. pi's own
// --session <id> would search the same way but may ask on stdin before RPC
// starts when the match is in another checkout, so the adapter always passes
// the file.
func sessionFile(env []string, cwd, id string) (string, error) {
	suffix := "_" + id + ".jsonl"
	if path, ok := findIn(checkoutDir(env, cwd), suffix); ok {
		return path, nil
	}
	dirs, err := os.ReadDir(sessionsDir(env))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	for _, d := range dirs {
		if d.IsDir() {
			if path, ok := findIn(filepath.Join(sessionsDir(env), d.Name()), suffix); ok {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("pi has no session '%s' in '%s'", id, sessionsDir(env))
}

func findIn(dir, suffix string) (string, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			return filepath.Join(dir, e.Name()), true
		}
	}
	return "", false
}

// entry is one line of a session file: the header, or an entry of the tree.
type entry struct {
	Type      string
	ID        string
	ParentID  string `dd:"parentId,+nullable"`
	Timestamp string
	Cwd       string
	Message   map[string]any `dd:",+opaque,+nullable"`
}

// readFile reads a session file; a test may stand in for it.
var readFile = os.ReadFile

// readSession reads a session file: its header and its entries, in order.
// a line that is not an entry is passed over.
func readSession(path string) (*entry, []*entry, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, nil, err
	}
	var header *entry
	var entries []*entry
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 64*1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		e := &entry{}
		if err := dd.BindJSON(e, line); err != nil {
			continue
		}
		if e.Type == "session" {
			if header == nil {
				header = e
			}
			continue
		}
		entries = append(entries, e)
	}
	if header == nil {
		return nil, nil, fmt.Errorf("'%s' is not a pi session file", path)
	}
	return header, entries, scanner.Err()
}

// conversation is the session's chain, oldest first: from the file's last
// entry back by parentId to its root. no entry marks the current leaf; pi
// itself reopens a session at its last entry the same way (pi 1.5.2,
// session-manager), so a file two writers branched follows whichever wrote
// last (capability report, 2026-11-16).
func conversation(entries []*entry) []*entry {
	if len(entries) == 0 {
		return nil
	}
	byID := make(map[string]*entry, len(entries))
	for _, e := range entries {
		byID[e.ID] = e
	}
	var chain []*entry
	seen := map[string]bool{}
	for e := entries[len(entries)-1]; e != nil && !seen[e.ID]; e = byID[e.ParentID] {
		seen[e.ID] = true
		chain = append(chain, e)
		if e.ParentID == "" {
			break
		}
	}
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

// message is the part of a pi AgentMessage the adapter reads, live and in a
// session file alike (pi 1.5.2, docs/message-types.md).
type message struct {
	Role         string
	Content      any `dd:",+opaque,+nullable"`
	Timestamp    int64
	StopReason   string `dd:"stopReason"`
	ErrorMessage string `dd:"errorMessage"`
	ToolCallID   string `dd:"toolCallId"`
	ToolName     string `dd:"toolName"`
	IsError      bool   `dd:"isError"`
	Provider     string
	Model        string
	CustomType   string `dd:"customType"`
	ResponseID   string `dd:"responseId"`
	Command      string
	Summary      string
}

func bindMessage(raw map[string]any) (*message, bool) {
	m := &message{}
	if raw == nil || dd.Bind(m, raw) != nil || m.Role == "" {
		return nil, false
	}
	return m, true
}

// mapMessage turns one message into the events to record, under id, its
// session entry's id. the live stream and an import share it, so a message
// seen live and later imported yields events with the same NativeIDs. a
// system message (the prompt and tools pi declares, written before a user
// message) is pi's own, and yields none.
func mapMessage(m *message, raw map[string]any, id string) []*model.Event {
	switch m.Role {
	case "system":
		return nil
	case "user":
		return []*model.Event{{Kind: model.EventUser, Text: contentText(m.Content), NativeID: id, Raw: raw}}
	case "assistant":
		var evs []*model.Event
		for i, block := range blocks(m.Content) {
			bid := fmt.Sprintf("%s/%d", id, i)
			switch block["type"] {
			case "text":
				text, _ := block["text"].(string)
				if strings.TrimSpace(text) == "" {
					continue
				}
				evs = append(evs, &model.Event{Kind: model.EventAssistant, Text: text, NativeID: bid, Raw: block})
			case "toolCall":
				name, _ := block["name"].(string)
				callID, _ := block["id"].(string)
				args, _ := block["arguments"].(map[string]any)
				evs = append(evs, &model.Event{Kind: model.EventToolUse, Tool: name, ToolUseID: callID, Text: toolTitle(name, args), NativeID: bid, Raw: block})
			case "thinking":
				evs = append(evs, &model.Event{Kind: model.EventNative, Text: "assistant/thinking", NativeID: bid, Raw: block})
			default:
				kind, _ := block["type"].(string)
				evs = append(evs, &model.Event{Kind: model.EventNative, Text: "assistant/" + kind, NativeID: bid, Raw: block})
			}
		}
		if m.StopReason == "error" && m.ErrorMessage != "" {
			// pi persists a failed response as an assistant message with its
			// error, which is not speech: kept under the native toggle.
			evs = append(evs, &model.Event{Kind: model.EventNative, Text: "error: " + m.ErrorMessage, NativeID: id + "/error", Raw: raw})
		}
		return evs
	case "toolResult":
		// a result pi marks as an error (a failed command, a call an
		// extension blocked, an abort) is a failed action, as claude's and
		// codex's are.
		ev := &model.Event{Kind: model.EventToolResult, Tool: m.ToolName, ToolUseID: m.ToolCallID, Text: contentText(m.Content), NativeID: id, Raw: raw}
		if m.IsError {
			ev.Status = model.TurnError
		}
		return []*model.Event{ev}
	default:
		// a direct shell command, an extension's context message, a branch
		// or compaction summary: pi's own, under the native toggle.
		text := m.Role
		if m.CustomType != "" {
			text += "/" + m.CustomType
		}
		return []*model.Event{{Kind: model.EventNative, Text: text, NativeID: id, Raw: raw}}
	}
}

func blocks(content any) []map[string]any {
	list, _ := content.([]any)
	var out []map[string]any
	for _, b := range list {
		if block, ok := b.(map[string]any); ok {
			out = append(out, block)
		}
	}
	return out
}

// contentText is a message's text: its content when that is a string, else
// its text blocks joined.
func contentText(content any) string {
	if s, ok := content.(string); ok {
		return s
	}
	var parts []string
	for _, block := range blocks(content) {
		if block["type"] == "text" {
			if text, ok := block["text"].(string); ok {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "\n")
}

// toolTitle is a tool call's summary line: the tool and what it was given, a
// shell command by its command, cut at maxTitle bytes.
func toolTitle(name string, args map[string]any) string {
	detail := ""
	for _, key := range []string{"command", "path", "file_path", "pattern", "query", "url"} {
		if v, ok := args[key].(string); ok && v != "" {
			detail = v
			break
		}
	}
	title := strings.TrimSpace(name + " " + strings.ReplaceAll(detail, "\n", " "))
	if len(title) <= maxTitle {
		return title
	}
	cut := maxTitle
	for cut > 0 && !utf8.RuneStart(title[cut]) {
		cut--
	}
	return title[:cut] + "…"
}

// history is a session's conversation as events, from its file.
func history(path string) ([]*model.Event, error) {
	evs, _, err := historyTurns(path)
	return evs, err
}

// historyTurns reads a session file's conversation, as history does, and
// follows its turns (harness.TurnTail). pi's file carries no turn status,
// but every assistant message carries its stopReason: toolUse leaves the turn
// open, as a tool's result and a prompt with no reply do; stop, length,
// error and aborted end it, as does none at all (the pi-rt1 fixture's 43
// sessions, pi 1.5.2: every assistant message has one).
func historyTurns(path string) ([]*model.Event, *harness.TurnTail, error) {
	_, entries, err := readSession(path)
	if err != nil {
		return nil, nil, err
	}
	evs, turn := turnsOf(entries)
	return evs, turn, nil
}

// turnsOf maps a session's entries, as read, to events and follows their
// turns, as historyTurns does.
func turnsOf(entries []*entry) ([]*model.Event, *harness.TurnTail) {
	turn := &harness.TurnTail{}
	var evs []*model.Event
	for _, e := range conversation(entries) {
		if e.Type != "message" {
			// model and thinking changes, labels, compactions and the like
			// are pi's bookkeeping: the live stream reports none of them.
			continue
		}
		m, ok := bindMessage(e.Message)
		if !ok {
			continue
		}
		mapped := mapMessage(m, e.Message, e.ID)
		switch m.Role {
		case "user":
			turn.Prompt(len(evs))
		case "assistant":
			if m.StopReason == "toolUse" {
				turn.Working()
			} else {
				turn.Ended(len(evs) + len(mapped))
			}
		case "toolResult":
			turn.Working()
		}
		evs = append(evs, mapped...)
	}
	return evs, turn
}

// discover lists the sessions stored for a checkout, newest first, with the
// first prompt as a preview.
func discover(env []string, cwd string, limit int) ([]harness.Session, error) {
	dir := checkoutDir(env, cwd)
	files, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []harness.Session
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
			continue
		}
		path := filepath.Join(dir, f.Name())
		header, entries, err := readSession(path)
		if err != nil {
			continue
		}
		s := harness.Session{NativeID: header.ID}
		s.CreatedAt, _ = time.Parse(time.RFC3339Nano, header.Timestamp)
		if info, err := f.Info(); err == nil {
			s.UpdatedAt = info.ModTime()
		}
		for _, e := range entries {
			if m, ok := bindMessage(e.Message); ok && e.Type == "message" && m.Role == "user" {
				s.Preview = contentText(m.Content)
				break
			}
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
