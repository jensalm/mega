package pi

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/model"
)

// sink records what a runtime reports, and lets a test wait for it.
type sink struct {
	mu       sync.Mutex
	events   []*model.Event
	requests []map[string]any
	ids      []string
	resolved []string
	// done closes once the runtime exited, with exit its error.
	done    chan struct{}
	exit    error
	changed chan struct{}
}

func newSink() *sink {
	return &sink{done: make(chan struct{}), changed: make(chan struct{}, 1)}
}

func (s *sink) poke() {
	select {
	case s.changed <- struct{}{}:
	default:
	}
}

func (s *sink) Event(ev *model.Event) {
	s.mu.Lock()
	s.events = append(s.events, ev)
	s.mu.Unlock()
	s.poke()
}
func (s *sink) Delta(string) {}
func (s *sink) Request(_ string, payload any) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := "svc-" + string(rune('a'+len(s.ids)))
	s.ids = append(s.ids, id)
	p, _ := payload.(map[string]any)
	s.requests = append(s.requests, p)
	s.poke()
	return id
}
func (s *sink) Resolved(id, outcome string) {
	s.mu.Lock()
	s.resolved = append(s.resolved, id+":"+outcome)
	s.mu.Unlock()
	s.poke()
}
func (s *sink) Background([]model.BackgroundJob) {}
func (s *sink) MCPServers([]model.MCPServer)     {}
func (s *sink) Exited(err error) {
	s.exit = err
	close(s.done)
}

// until waits for cond over what the sink holds.
func (s *sink) until(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.After(15 * time.Second)
	for {
		s.mu.Lock()
		ok := cond()
		s.mu.Unlock()
		if ok {
			return
		}
		select {
		case <-s.changed:
		case <-time.After(50 * time.Millisecond):
		case <-deadline:
			t.Fatalf("timed out waiting for %s", what)
		}
	}
}

// turns counts the turn ends recorded; replies the assistant text.
func (s *sink) turns() []string {
	var out []string
	for _, ev := range s.events {
		if ev.Kind == model.EventTurnEnd {
			out = append(out, ev.Status)
		}
	}
	return out
}

func (s *sink) replies() []string {
	var out []string
	for _, ev := range s.events {
		if ev.Kind == model.EventAssistant {
			out = append(out, ev.Text)
		}
	}
	return out
}

// fake is the committed fake pi, run through node as pi itself is.
func fake(t *testing.T) string {
	t.Helper()
	if _, err := harness.LookPath("node", os.Environ()); err != nil {
		t.Skip("node is not on PATH")
	}
	_, file, _, _ := goruntime.Caller(0)
	return filepath.Join(filepath.Dir(file), "fake", "pi")
}

// place is a scratch checkout with a scratch HOME, where the fake keeps its
// sessions, and node on PATH.
func place(t *testing.T) harness.Place {
	t.Helper()
	home, cwd := t.TempDir(), t.TempDir()
	env := []string{"HOME=" + home, "PATH=" + os.Getenv("PATH")}
	return harness.Place{Cwd: cwd, Env: env, Find: env}
}

func launch(t *testing.T, h *Harness, p harness.Place, nativeID string) (harness.Runtime, *sink) {
	t.Helper()
	s := newSink()
	rt, err := h.Launch(harness.Spec{Cwd: p.Cwd, Env: p.Env, Find: p.Find, NativeID: nativeID, Model: "fake/large", Effort: "low"}, s)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	t.Cleanup(func() {
		_ = rt.Close()
		select {
		case <-s.done:
		case <-time.After(20 * time.Second):
		}
	})
	return rt, s
}

func prompt(t *testing.T, rt harness.Runtime, s *sink, text string, turns int) {
	t.Helper()
	if err := rt.Prompt("uuid", text); err != nil {
		t.Fatalf("prompt %q: %v", text, err)
	}
	s.until(t, "the turn of "+text, func() bool { return len(s.turns()) >= turns })
}

// a launch reports its start with the session, the model as provider/id and
// the thinking level it was given; a turn's reply arrives and ends success;
// the prompt is recorded under pi's own identity, so an import of the
// session afterwards brings in nothing the live stream recorded.
func TestALaunchRunsATurnAndAnImportAddsNothing(t *testing.T) {
	h := New(Config{Binary: fake(t)})
	p := place(t)
	rt, s := launch(t, h, p, "")
	s.mu.Lock()
	started := s.events[0]
	s.mu.Unlock()
	if started.Status != model.StatusStarted || started.Facts.Session != rt.NativeID() || started.Facts.Model != "fake/large" || started.Facts.Effort != "low" || started.Facts.PermissionMode != "" {
		t.Fatalf("started = %+v %+v", started, started.Facts)
	}
	prompt(t, rt, s, "echo hello", 1)
	if got := s.replies(); len(got) != 1 || got[0] != "hello" {
		t.Errorf("replies = %v", got)
	}
	if got := s.turns(); got[0] != model.TurnSuccess {
		t.Errorf("turn = %v", got)
	}
	live := map[string]bool{}
	s.mu.Lock()
	for _, ev := range s.events {
		if ev.NativeID != "" {
			live[ev.NativeID] = true
		}
	}
	s.mu.Unlock()
	_ = rt.Close()
	<-s.done
	hist, err := h.History(context.Background(), p, rt.NativeID())
	if err != nil {
		t.Fatal(err)
	}
	if len(hist.Events) != 2 {
		t.Errorf("history = %d events", len(hist.Events))
	}
	for _, ev := range hist.Events {
		if !live[ev.NativeID] {
			t.Errorf("an import would add %s %q again", ev.NativeID, ev.Text)
		}
	}
}

// every dialog is a question offering Dismiss and never a session-wide
// allow; each answer reaches pi as its kind takes it, and Dismiss is pi's
// own cancel.
func TestDialogsAreQuestions(t *testing.T) {
	h := New(Config{Binary: fake(t)})
	rt, s := launch(t, h, place(t), "")
	cases := []struct {
		prompt string
		answer harness.Answer
		reply  string
	}{
		// the gate is answered Block: allowing it is not what this test is
		// about, and a test allows only the approval it is about.
		{"ask", harness.Answer{Answers: map[string]string{"Approve bash command: sleep 1": "Block"}}, "blocked"},
		{"choose", harness.Answer{Decision: harness.DecisionDismiss}, "dismissed"},
		{"confirm", harness.Answer{Answers: map[string]string{"Proceed?": "No"}}, "declined"},
		{"input", harness.Answer{Answers: map[string]string{"value": "Ada"}}, "you typed: Ada"},
		{"editor", harness.Answer{Answers: map[string]string{"value": "one\ntwo"}}, "edited: one\ntwo"},
	}
	for i, c := range cases {
		if err := rt.Prompt("uuid", c.prompt); err != nil {
			t.Fatal(err)
		}
		s.until(t, "the dialog of "+c.prompt, func() bool { return len(s.ids) > i })
		s.mu.Lock()
		payload, id := s.requests[i], s.ids[i]
		s.mu.Unlock()
		if payload["dismiss"] != true || payload["allow_session"] != nil {
			t.Errorf("%s: payload %v", c.prompt, payload)
		}
		if err := rt.Answer(id, c.answer); err != nil {
			t.Fatalf("%s: answer: %v", c.prompt, err)
		}
		s.until(t, "the reply to "+c.prompt, func() bool { return len(s.turns()) > i })
		if got := s.replies(); got[len(got)-1] != c.reply {
			t.Errorf("%s: reply %q, want %q", c.prompt, got[len(got)-1], c.reply)
		}
		if err := rt.Answer(id, c.answer); err != harness.ErrNotPending {
			t.Errorf("%s: a second answer: %v", c.prompt, err)
		}
	}
	s.mu.Lock()
	ask, form := s.requests[0], s.requests[3]
	s.mu.Unlock()
	if q, _ := ask["questions"].([]any); len(q) != 1 || ask["tool"] != "question" {
		t.Errorf("a select: %v", ask)
	}
	if f, _ := form["fields"].([]any); len(f) != 1 || form["tool"] != "form" || form["message"] != "Your name?" {
		t.Errorf("an input: %v", form)
	}
	// a choice pi did not offer is refused, and the dialog stays pending.
	if err := rt.Prompt("uuid", "choose"); err != nil {
		t.Fatal(err)
	}
	s.until(t, "another dialog", func() bool { return len(s.ids) > len(cases) })
	s.mu.Lock()
	id := s.ids[len(cases)]
	s.mu.Unlock()
	if err := rt.Answer(id, harness.Answer{Answers: map[string]string{"Which color do you prefer?": "Green"}}); err == nil {
		t.Error("an answer pi did not offer was sent")
	}
	if err := rt.Answer(id, harness.Answer{Answers: map[string]string{"Which color do you prefer?": "Blue"}}); err != nil {
		t.Errorf("the dialog did not stay pending: %v", err)
	}
}

// a dialog with a timeout is settled by pi on its own, sending nothing, and
// its record is settled the same way.
func TestADialogWithATimeoutExpires(t *testing.T) {
	h := New(Config{Binary: fake(t)})
	rt, s := launch(t, h, place(t), "")
	prompt(t, rt, s, "timed", 1)
	s.until(t, "the expiry", func() bool { return slices.Contains(s.resolved, "svc-a:"+model.OutcomeExpired) })
	if got := s.replies(); got[0] != "timed out" {
		t.Errorf("reply = %v", got)
	}
}

// an interrupt with a dialog pending says what it waits on, and ends the turn
// only once the operator answers or dismisses the dialog; the command it
// gated never runs. mega never answers it to let the interrupt through.
func TestAnInterruptWaitsOnADialog(t *testing.T) {
	h := New(Config{Binary: fake(t)})
	rt, s := launch(t, h, place(t), "")
	if err := rt.Prompt("uuid", "ask"); err != nil {
		t.Fatal(err)
	}
	s.until(t, "the dialog", func() bool { return len(s.ids) == 1 })
	if err := rt.Interrupt(); err != nil {
		t.Fatal(err)
	}
	s.until(t, "the waiting note", func() bool {
		for _, ev := range s.events {
			if ev.Status == model.StatusWaiting && strings.Contains(ev.Text, "Approve bash command: sleep 1") {
				return true
			}
		}
		return false
	})
	time.Sleep(300 * time.Millisecond)
	s.mu.Lock()
	ended := len(s.turns())
	s.mu.Unlock()
	if ended != 0 {
		t.Fatal("the interrupt went through with the dialog unanswered")
	}
	if err := rt.Answer("svc-a", harness.Answer{Decision: harness.DecisionDismiss}); err != nil {
		t.Fatal(err)
	}
	s.until(t, "the interrupted turn", func() bool { return len(s.turns()) == 1 })
	if got := s.turns(); got[0] != model.TurnInterrupted {
		t.Errorf("turn = %v", got)
	}
	for _, ev := range s.events {
		if ev.Kind == model.EventToolResult && ev.Text == "slept" {
			t.Error("the gated command ran")
		}
	}
}

// a release with a dialog pending ends pi's process: its input closed, then
// signals, since the dialog holds pi open; the dialog is recorded as never
// answered, the command it gated never runs, and the session resumes.
func TestAReleaseWithADialogPendingEndsPi(t *testing.T) {
	h := New(Config{Binary: fake(t)})
	p := place(t)
	s := newSink()
	rt, err := h.Launch(harness.Spec{Cwd: p.Cwd, Env: p.Env, Find: p.Find, Model: "fake/large", Effort: "low"}, s)
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.Prompt("uuid", "ask"); err != nil {
		t.Fatal(err)
	}
	s.until(t, "the dialog", func() bool { return len(s.ids) == 1 })
	begun := time.Now()
	if err := rt.Interrupt(); err != nil {
		t.Fatal(err)
	}
	if err := rt.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-s.done:
	case <-time.After(20 * time.Second):
		t.Fatal("pi was not ended")
	}
	exit := s.exit
	if took := time.Since(begun); took > 12*time.Second {
		t.Errorf("ending pi took %s", took)
	}
	if exit == nil || !strings.Contains(exit.Error(), "mega ended its process") {
		t.Errorf("exit = %v", exit)
	}
	if !slices.Contains(s.resolved, "svc-a:"+model.OutcomeUnanswered) {
		t.Errorf("resolved = %v", s.resolved)
	}
	for _, ev := range s.events {
		if ev.Kind == model.EventToolResult && ev.Text == "slept" {
			t.Error("the gated command ran")
		}
	}
	// the session resumes, by its exact id.
	again, s2 := launch(t, h, p, rt.NativeID())
	if again.NativeID() != rt.NativeID() {
		t.Fatalf("resumed %s, not %s", again.NativeID(), rt.NativeID())
	}
	prompt(t, again, s2, "echo again", 1)
	if got := s2.replies(); got[0] != "again" {
		t.Errorf("reply = %v", got)
	}
}

// discovery lists the checkout's sessions with their first prompt; a resume
// of a session pi does not hold is refused before anything starts.
// a listing read only in part is not a clean one: a process that cannot be
// read could be in pi's group. what could be read still counts, so a
// confirmed survivor is named beside it, and a process that went
// mid-listing, its entry gone, is ordinary.
func TestAListingReadInPartIsNotClean(t *testing.T) {
	was := procRoot
	procRoot = t.TempDir()
	t.Cleanup(func() { procRoot = was })
	if err := os.Mkdir(filepath.Join(procRoot, "456"), 0o755); err != nil {
		t.Fatal(err)
	}
	if found, unread, err := inGroup(123); err != nil || len(found) != 0 || len(unread) != 0 {
		t.Fatalf("a process gone mid-listing: %v %v %v", found, unread, err)
	}
	write := func(pid int, mode os.FileMode) {
		path := filepath.Join(procRoot, strconv.Itoa(pid), "stat")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(strconv.Itoa(pid)+" (sleep) S 1 123"), mode); err != nil {
			t.Fatal(err)
		}
	}
	write(200, 0o644)
	write(300, 0o000)
	if _, err := os.ReadFile(filepath.Join(procRoot, "300", "stat")); err == nil {
		t.Skip("an unreadable stat file stays readable here, as it does for root")
	}
	alive, err := groupSurvivors(123, 0)
	if err == nil || !slices.Equal(alive, []int{200}) {
		t.Errorf("a survivor in the group beside a process that cannot be read: %v %v", alive, err)
	}
}

// a release where the processes cannot be listed, as with no /proc, says so:
// not a clean stop, and not processes left running.
func TestAReleaseThatCouldNotLookSaysSo(t *testing.T) {
	h := New(Config{Binary: fake(t)})
	p := place(t)
	s := newSink()
	rt, err := h.Launch(harness.Spec{Cwd: p.Cwd, Env: p.Env, Find: p.Find, Model: "fake/large", Effort: "low"}, s)
	if err != nil {
		t.Fatal(err)
	}
	was := procRoot
	procRoot = filepath.Join(t.TempDir(), "no-proc")
	t.Cleanup(func() { procRoot = was })
	if found, _, err := inGroup(1); err == nil {
		t.Errorf("an unreadable listing gave %v with no error", found)
	}
	if err := rt.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-s.done:
	case <-time.After(20 * time.Second):
		t.Fatal("pi did not exit")
	}
	if !errors.Is(s.exit, harness.ErrUnchecked) || errors.Is(s.exit, harness.ErrLeftRunning) {
		t.Errorf("exit = %v", s.exit)
	}
}

// once a stop is fenced, an answer to a dialog that would let work go on is
// refused, even one accepted before the stop, and never reaches pi; a
// dismissal still does, which is what lets an interrupt pi holds go through.
func TestNoAnswerLetsWorkGoOnAfterTheStopIsFenced(t *testing.T) {
	h := New(Config{Binary: fake(t)})
	p := place(t)
	s := newSink()
	rt, err := h.Launch(harness.Spec{Cwd: p.Cwd, Env: p.Env, Find: p.Find, Model: "fake/large", Effort: "low"}, s)
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.Prompt("uuid", "ask"); err != nil {
		t.Fatal(err)
	}
	s.until(t, "the dialog", func() bool { return len(s.ids) == 1 })
	rt.Stopping()
	if err := rt.Answer("svc-a", harness.Answer{Answers: map[string]string{"Approve bash command: sleep 1": "Allow"}}); !errors.Is(err, harness.ErrNotPending) {
		t.Errorf("a chosen answer after the fence: %v", err)
	}
	if err := rt.Answer("svc-a", harness.Answer{Decision: harness.DecisionDismiss}); err != nil {
		t.Errorf("a dismissal after the fence: %v", err)
	}
	s.until(t, "the turn", func() bool { return len(s.turns()) == 1 })
	for _, ev := range s.events {
		if ev.Kind == model.EventToolResult && ev.Text == "slept" {
			t.Error("the gated command ran")
		}
	}
	if err := rt.Close(); err != nil {
		t.Fatal(err)
	}
	<-s.done
}

func TestDiscoverAndAMissingSession(t *testing.T) {
	h := New(Config{Binary: fake(t)})
	p := place(t)
	rt, s := launch(t, h, p, "")
	prompt(t, rt, s, "echo first", 1)
	found, err := h.Discover(context.Background(), p)
	if err != nil || len(found) != 1 || found[0].NativeID != rt.NativeID() || found[0].Preview != "echo first" {
		t.Fatalf("discover = %+v, %v", found, err)
	}
	_, err = h.Launch(harness.Spec{Cwd: p.Cwd, Env: p.Env, Find: p.Find, NativeID: "00000000-0000-4000-8000-000000000000"}, newSink())
	if err == nil || !strings.Contains(err.Error(), "pi has no session") {
		t.Errorf("a missing session: %v", err)
	}
}

// the model list names each model provider/id with the thinking levels it
// takes, none for a model without reasoning; the settings are what pi itself
// starts with, from its settings.json, clamped to the model.
func TestModelsAndSettings(t *testing.T) {
	h := New(Config{Binary: fake(t)})
	p := place(t)
	catalog, err := h.Models(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Version != "1.5.2" || len(catalog.Models) != 2 {
		t.Fatalf("catalog = %+v", catalog)
	}
	large, small := catalog.Models[0], catalog.Models[1]
	if large.ID != "fake/large" || !large.IsDefault || large.DefaultEffort != "medium" || strings.Join(large.Efforts, ",") != "off,minimal,low,medium,high" {
		t.Errorf("large = %+v", large)
	}
	if small.ID != "fake/small" || small.IsDefault || len(small.Efforts) != 0 {
		t.Errorf("small = %+v", small)
	}
	if got, err := h.Settings(context.Background(), p); err != nil || got.Model != "fake/large" || got.Effort != "medium" {
		t.Errorf("settings = %+v, %v", got, err)
	}
	home, _ := harness.Lookup(p.Env, "HOME")
	write(t, filepath.Join(home, ".pi", "agent", "settings.json"), `{"defaultProvider":"fake","defaultModel":"small","defaultThinkingLevel":"high"}`)
	if got, err := h.Settings(context.Background(), p); err != nil || got.Model != "fake/small" || got.Effort != "" {
		t.Errorf("settings naming a model without reasoning = %+v, %v", got, err)
	}
	if v, err := h.Version(context.Background(), p.Find); err != nil || v != "1.5.2" {
		t.Errorf("version = %q, %v", v, err)
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// a checkout with project resources pi has not been trusted in says so at
// start; one trusted, one without such resources, or a pi that trusts every
// project, says nothing. mega reads pi's trust and writes nothing.
func TestTheTrustNotice(t *testing.T) {
	home, cwd := t.TempDir(), t.TempDir()
	env := []string{"HOME=" + home}
	if got := trustNotice(env, cwd); got != "" {
		t.Errorf("no project resources: %q", got)
	}
	write(t, filepath.Join(cwd, ".pi", "extensions", "gate.ts"), "")
	if got := trustNotice(env, cwd); !strings.Contains(got, "has not been trusted") || !strings.Contains(got, "mega never trusts a checkout for you") {
		t.Errorf("untrusted: %q", got)
	}
	resolved, _ := filepath.EvalSymlinks(cwd)
	trust := filepath.Join(home, ".pi", "agent", "trust.json")
	write(t, trust, `{"`+resolved+`": false}`)
	if got := trustNotice(env, cwd); got == "" {
		t.Error("a saved decision not to trust it reads as trusted")
	}
	write(t, trust, `{"`+filepath.Dir(resolved)+`": true}`)
	if got := trustNotice(env, cwd); got != "" {
		t.Errorf("trusted through an ancestor: %q", got)
	}
	before, _ := os.ReadFile(trust)
	write(t, trust, `{}`)
	write(t, filepath.Join(home, ".pi", "agent", "settings.json"), `{"defaultProjectTrust":"always"}`)
	if got := trustNotice(env, cwd); got != "" {
		t.Errorf("defaultProjectTrust always: %q", got)
	}
	_ = before
}

// the include list is pi's configuration, never its credentials, trust,
// caches, packages or sessions; PI_CODING_AGENT_DIR moves it.
func TestTheConfigurationIncludeList(t *testing.T) {
	h := New(Config{})
	roots := h.Configuration([]string{"HOME=/h"})
	if len(roots) != 1 || roots[0].Dir != "/h/.pi/agent" || roots[0].Harness != model.HarnessPi {
		t.Fatalf("roots = %+v", roots)
	}
	for _, never := range []string{"auth.json", "mcp-auth.json", "trust.json", "models-store.json", "mcp-cache.json", "sessions/", "npm/", "bin/"} {
		if slices.Contains(roots[0].Include, never) {
			t.Errorf("%s is on the include list", never)
		}
	}
	for _, want := range []string{"settings.json", "models.json", "mcp.json", "extensions/"} {
		if !slices.Contains(roots[0].Include, want) {
			t.Errorf("%s is not on the include list", want)
		}
	}
	if got := h.Configuration([]string{"HOME=/h", "PI_CODING_AGENT_DIR=/elsewhere"}); got[0].Dir != "/elsewhere" {
		t.Errorf("PI_CODING_AGENT_DIR: %s", got[0].Dir)
	}
}

// recorded pi 1.5.2 sessions from the pi-rt1 fixture (2026-11-16, system
// prompts and usage dropped): the probe's P1 to P4, and P7, a file two
// processes wrote at once. the conversation is the chain from the last
// entry back, as pi reopens it; system messages yield nothing, and every
// event carries the identity the live stream gives it.
// a session read back whose last turn has no end says so, as the session of
// a pi that was lost mid-turn does: cut from the recorded p1 and p4 session
// at a tool's result, after a call pi made with stopReason toolUse. the turn
// before it, ended with stop, is not open; the whole session is not.
func TestHistorySaysWhenTheLastTurnHasNoEnd(t *testing.T) {
	open, err := historyOf(filepath.Join("testdata", "open-turn.jsonl"), "")
	if err != nil {
		t.Fatal(err)
	}
	var unfinished, finished []string
	for _, ev := range open.Events {
		switch {
		case strings.HasPrefix(ev.NativeID, "26df6a94") || ev.NativeID == "e3e76658" || ev.NativeID == "1795b865":
			unfinished = append(unfinished, ev.NativeID)
		case strings.HasPrefix(ev.NativeID, "97782afa") || ev.NativeID == "d8a83bc4":
			finished = append(finished, ev.NativeID)
		}
	}
	if len(unfinished) == 0 || len(finished) == 0 {
		t.Fatalf("events %+v", open.Events)
	}
	for _, id := range unfinished {
		if !open.Open[id] {
			t.Errorf("%s of the unfinished turn is not open: %v", id, open.Open)
		}
	}
	for _, id := range finished {
		if open.Open[id] {
			t.Errorf("%s of the turn that ended is open", id)
		}
	}
	closed, err := historyOf(filepath.Join("testdata", "p1-and-p4.jsonl"), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(closed.Open) != 0 {
		t.Errorf("a session that ends with stop has open events: %v", closed.Open)
	}
}

// History reads only the session it was asked for: a file whose name ends
// with the id but whose header names another session is refused, before
// anything is imported.
func TestHistoryChecksTheSessionsOwnID(t *testing.T) {
	dir := t.TempDir()
	recorded, err := os.ReadFile(filepath.Join("testdata", "p1-and-p4.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	checkout := t.TempDir()
	env := append(os.Environ(), "PI_CODING_AGENT_SESSION_DIR="+dir)
	if err := os.MkdirAll(checkoutDir(env, checkout), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(checkoutDir(env, checkout), "2026-11-19T00-00-00-000Z_not-this-one.jsonl"), recorded, 0o600); err != nil {
		t.Fatal(err)
	}
	h := New(Config{Binary: "pi"})
	if history, err := h.History(t.Context(), harness.Place{Cwd: checkout, Env: env, Find: env}, "not-this-one"); err == nil {
		t.Errorf("a file holding another session was imported: %d events", len(history.Events))
	}
}

func TestRecordedSessions(t *testing.T) {
	evs, err := history(filepath.Join("testdata", "p1-and-p4.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var users, tools []string
	last := ""
	for _, ev := range evs {
		if ev.NativeID == "" || strings.HasPrefix(ev.NativeID, "system") {
			t.Errorf("event %s %q: identity %q", ev.Kind, ev.Text, ev.NativeID)
		}
		switch ev.Kind {
		case model.EventUser:
			users = append(users, ev.Text)
		case model.EventToolUse:
			tools = append(tools, ev.Text)
		case model.EventAssistant:
			last = ev.Text
		}
	}
	if len(users) != 7 || !strings.Contains(users[0], "ALPHA-1a2b") || !strings.Contains(users[6], "ZETA-1b2c") {
		t.Errorf("prompts = %d: %q", len(users), users)
	}
	if strings.Join(tools, " | ") != "bash cat fixture.txt | ask_user | bash sleep 3 | bash sleep 2" {
		t.Errorf("tool calls = %q", tools)
	}
	if last != "ZETA-1b2c" {
		t.Errorf("the last reply = %q", last)
	}
	branched, err := history(filepath.Join("testdata", "p7-two-writers.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var replies []string
	for _, ev := range branched {
		if ev.Kind == model.EventAssistant {
			replies = append(replies, ev.Text)
		}
	}
	if len(replies) != 3 || replies[0] != "SHARED-A-1" || replies[1] != "SHARED-B-1" || !strings.HasSuffix(replies[2], "SHARED-B-2") {
		t.Errorf("the branched file reads %q; want the last writer's chain", replies)
	}
}

// an extension that passes its dialog the abort signal lets an interrupt
// cancel the dialog: the turn ends interrupted, and the dialog is settled as
// aborted, with nothing sent to pi; a later answer is refused.
func TestAnInterruptCancelsADialogThatTakesTheSignal(t *testing.T) {
	h := New(Config{Binary: fake(t)})
	rt, s := launch(t, h, place(t), "")
	if err := rt.Prompt("uuid", "signal"); err != nil {
		t.Fatal(err)
	}
	s.until(t, "the dialog", func() bool { return len(s.ids) == 1 })
	if err := rt.Interrupt(); err != nil {
		t.Fatal(err)
	}
	s.until(t, "the interrupted turn", func() bool { return len(s.turns()) == 1 })
	if got := s.turns(); got[0] != model.TurnInterrupted {
		t.Errorf("turn = %v", got)
	}
	s.until(t, "the dialog settled", func() bool { return slices.Contains(s.resolved, "svc-a:"+model.OutcomeAborted) })
	if err := rt.Answer("svc-a", harness.Answer{Answers: map[string]string{"Cancelled by an abort": "Go"}}); err != harness.ErrNotPending {
		t.Errorf("a late answer: %v", err)
	}
}

// two tool results of one response can share a millisecond: on import each
// keeps its own identity, its entry's.
func TestToolResultsInOneMillisecondKeepTheirIdentities(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.jsonl")
	write(t, path, `{"type":"session","version":3,"id":"s","cwd":"/c"}
{"type":"message","id":"u1","parentId":null,"message":{"role":"user","content":"go","timestamp":1794841243300}}
{"type":"message","id":"a1","parentId":"u1","message":{"role":"assistant","responseId":"resp-1","content":[{"type":"toolCall","id":"call_a","name":"read","arguments":{"path":"x"}},{"type":"toolCall","id":"call_b","name":"read","arguments":{"path":"y"}}],"timestamp":1794841243345}}
{"type":"message","id":"r1","parentId":"a1","message":{"role":"toolResult","toolCallId":"call_a","toolName":"read","content":[{"type":"text","text":"x"}],"timestamp":1794841243346}}
{"type":"message","id":"r2","parentId":"r1","message":{"role":"toolResult","toolCallId":"call_b","toolName":"read","content":[{"type":"text","text":"y"}],"timestamp":1794841243346}}
`)
	evs, err := history(path)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, ev := range evs {
		if seen[ev.NativeID] {
			t.Errorf("two events share %q", ev.NativeID)
		}
		seen[ev.NativeID] = true
	}
	if !seen["r1"] || !seen["r2"] || !seen["a1/0"] || !seen["a1/1"] {
		t.Errorf("identities = %v", seen)
	}
}

// two messages of one role can share a millisecond: each is identified by
// its own session entry, live and on import, so neither is dropped.
func TestTwoMessagesInOneMillisecondAreBothKept(t *testing.T) {
	h := New(Config{Binary: fake(t)})
	p := place(t)
	rt, s := launch(t, h, p, "")
	prompt(t, rt, s, "twins", 1)
	if got := s.replies(); strings.Join(got, ",") != "first,second" {
		t.Fatalf("replies = %v", got)
	}
	live := map[string]bool{}
	s.mu.Lock()
	for _, ev := range s.events {
		if ev.NativeID != "" {
			if live[ev.NativeID] {
				t.Errorf("two events share %q", ev.NativeID)
			}
			live[ev.NativeID] = true
		}
	}
	s.mu.Unlock()
	_ = rt.Close()
	<-s.done
	hist, err := h.History(context.Background(), p, rt.NativeID())
	if err != nil {
		t.Fatal(err)
	}
	for _, ev := range hist.Events {
		if !live[ev.NativeID] {
			t.Errorf("an import would add %s %q again", ev.NativeID, ev.Text)
		}
	}
}

// once pi can no longer be asked (its input closed for a release), a message
// it reported is identified by its entry in the session file, the id an
// import reads, never by an id of the adapter's own.
func TestAnEntryIDComesFromTheFileWhenPiCannotBeAsked(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.jsonl")
	write(t, path, `{"type":"session","version":3,"id":"s","cwd":"/c"}
{"type":"message","id":"u1","parentId":null,"message":{"role":"user","content":"go","timestamp":1794841243300}}
{"type":"message","id":"a1","parentId":"u1","message":{"role":"assistant","content":[{"type":"text","text":"one"}],"timestamp":1794841243345}}
{"type":"message","id":"a2","parentId":"a1","message":{"role":"assistant","content":[{"type":"text","text":"two"}],"timestamp":1794841243345}}
`)
	r := newRuntime(newSink())
	r.conn = &conn{inClose: true, calls: map[string]chan reply{}, done: make(chan struct{})}
	r.sessionFile, r.cursor = path, "u1"
	first := &message{Role: "assistant", Timestamp: 1794841243345}
	if id := r.entryID(first); id != "a1" {
		t.Errorf("the first reply is %q", id)
	}
	if id := r.entryID(first); id != "a2" {
		t.Errorf("the second, in the same millisecond, is %q", id)
	}
}

// a dialog's timeout runs from when pi sent it: one that waited past its
// deadline before it was handled is settled as expired at once, and an
// answer to it is refused, since pi no longer reads one.
func TestADialogsDeadlineRunsFromItsArrival(t *testing.T) {
	s := newSink()
	r := newRuntime(s)
	r.conn = &conn{inClose: true, calls: map[string]chan reply{}, done: make(chan struct{})}
	r.uiRequest(map[string]any{"type": "extension_ui_request", "id": "d1", "method": "select", "title": "Now?", "options": []any{"Now"}, "timeout": float64(1000)}, time.Now().Add(-2*time.Second))
	s.until(t, "the expiry", func() bool { return slices.Contains(s.resolved, "svc-a:"+model.OutcomeExpired) })
	if err := r.Answer("svc-a", harness.Answer{Answers: map[string]string{"Now?": "Now"}}); err != harness.ErrNotPending {
		t.Errorf("an answer past the deadline: %v", err)
	}
}

// a message whose session entry cannot be found, from pi or its file, waits
// with what follows it, and is looked for again as the run settles; still
// unfound, it is shown and counted as output with no identity, and the turn
// ends in error saying an import may bring it in again.
func TestAMessageWithNoEntryIsShownAndTheTurnSaysSo(t *testing.T) {
	s := newSink()
	r := newRuntime(s)
	r.conn = &conn{inClose: true, calls: map[string]chan reply{}, done: make(chan struct{})}
	r.sessionFile = filepath.Join(t.TempDir(), "missing.jsonl")
	r.handle(map[string]any{"type": "message_end", "message": map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "text", "text": "hi"}}, "timestamp": float64(1794841243345), "stopReason": "stop"}}, time.Now())
	if len(s.events) != 0 {
		t.Fatalf("recorded before the run settled: %+v", s.events)
	}
	r.handle(map[string]any{"type": "agent_settled", "aborted": false}, time.Now())
	if len(s.events) != 2 || s.events[0].Kind != model.EventAssistant || s.events[0].Text != "hi" || s.events[0].NativeID != "" {
		t.Fatalf("events = %+v", s.events)
	}
	if end := s.events[1]; end.Kind != model.EventTurnEnd || end.Status != model.TurnError || !strings.Contains(end.Text, "may bring them in again") {
		t.Errorf("turn end = %+v", end)
	}
}
