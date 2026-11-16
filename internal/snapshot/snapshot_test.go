package snapshot

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/michaelquigley/mega/internal/harness"
	"github.com/michaelquigley/mega/internal/harness/claude"
	"github.com/michaelquigley/mega/internal/harness/codex"
	"github.com/michaelquigley/mega/internal/model"
)

// write puts a file at path, making its directories; every value here is
// made up.
func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// home fabricates a home holding both harnesses' configuration beside what a
// snapshot must never copy: credentials, a login, history, transcripts,
// sessions, databases, caches, and the skills each harness installs itself.
func home(t *testing.T) string {
	t.Helper()
	h := t.TempDir()
	for path, body := range map[string]string{
		".claude.json":                                   `{"numStartups":1}`,
		".claude/settings.json":                          `{"effortLevel":"medium"}`,
		".claude/CLAUDE.md":                              "# mine\n",
		".claude/skills/mine/SKILL.md":                   "a skill\n",
		".claude/skills/synced/bucket/x/SKILL.md":        "synced from the account\n",
		".claude/agents/reviewer.md":                     "an agent\n",
		".claude/plugins/installed_plugins.json":         `{}`,
		".claude/plugins/known_marketplaces.json":        `{}`,
		".claude/plugins/blocklist.json":                 `{}`,
		".claude/plugins/plugin-directory-cache-v2.json": `{}`,
		".claude/.credentials.json":                      `{"token":"made-up"}`,
		".claude/history.jsonl":                          "{}\n",
		".claude/projects/p/s.jsonl":                     "{}\n",
		".claude/sessions/s.json":                        "{}\n",
		".claude/stats-cache.json":                       "{}\n",
		".codex/config.toml":                             "model = \"m\"\n",
		".codex/local.config.toml":                       "model = \"n\"\n",
		".codex/AGENTS.md":                               "agents\n",
		".codex/rules/default.rules":                     "rule\n",
		".codex/skills/mine/SKILL.md":                    "a skill\n",
		".codex/skills/.system/x/SKILL.md":               "bundled\n",
		".codex/version.json":                            `{"latest_version":"0.0.0"}`,
		".codex/.sandbox_migration":                      "1\n",
		".codex/.personality_migration":                  "1\n",
		".codex/auth.json":                               `{"token":"made-up"}`,
		".codex/history.jsonl":                           "{}\n",
		".codex/state_5.sqlite":                          "db",
		".codex/models_cache.json":                       "{}",
		".codex/sessions/2026/s.jsonl":                   "{}\n",
	} {
		write(t, filepath.Join(h, path), body)
	}
	return h
}

func roots(env []string) []harness.ConfigRoot {
	return append(claude.New(claude.Config{}).Configuration(env), codex.New(codex.Config{}).Configuration(env)...)
}

func names(m *Manifest) []string {
	var out []string
	for _, e := range m.Files {
		out = append(out, e.Harness+"/"+e.Root+"/"+e.Path)
	}
	sort.Strings(out)
	return out
}

func newStore(t *testing.T, keep, atVersion int) (*Store, *time.Time) {
	t.Helper()
	now := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	return NewStore(Options{Dir: filepath.Join(t.TempDir(), "data", "snapshots"), Keep: keep, KeepAtVersion: atVersion, Now: func() time.Time { return now }}), &now
}

// the include lists copy each harness's configuration and nothing else: no
// credential or login beside them, no history, transcript, session,
// database or cache, and not the skills a harness installs itself.
func TestTheIncludeListCopiesConfigurationOnly(t *testing.T) {
	h := home(t)
	s, _ := newStore(t, 5, 5)
	res, err := s.Take(roots([]string{"HOME=" + h}), nil, false)
	if err != nil || res.Manifest == nil {
		t.Fatalf("take: %+v, %v", res, err)
	}
	want := []string{
		"claude/config/CLAUDE.md",
		"claude/config/agents/reviewer.md",
		"claude/config/plugins/blocklist.json",
		"claude/config/plugins/installed_plugins.json",
		"claude/config/plugins/known_marketplaces.json",
		"claude/config/settings.json",
		"claude/config/skills/mine/SKILL.md",
		"claude/global/.claude.json",
		"codex/home/.personality_migration",
		"codex/home/.sandbox_migration",
		"codex/home/AGENTS.md",
		"codex/home/config.toml",
		"codex/home/local.config.toml",
		"codex/home/rules/default.rules",
		"codex/home/skills/mine/SKILL.md",
		"codex/home/version.json",
	}
	if got := names(res.Manifest); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("copied:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	// nothing that was not listed reached the store, under any name.
	err = filepath.WalkDir(s.Dir(), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		body, _ := os.ReadFile(path)
		if strings.Contains(string(body), "made-up") || strings.Contains(string(body), "synced from") || strings.Contains(string(body), "bundled") {
			t.Errorf("%s holds what was never to be copied", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// the store is the operator's alone, and its copies read-only.
	for _, dir := range []string{s.Dir(), filepath.Join(s.Dir(), res.Manifest.Name)} {
		if info, err := os.Stat(dir); err != nil || info.Mode().Perm() != 0o700 {
			t.Errorf("%s: %v %v", dir, info.Mode(), err)
		}
	}
	copied := s.Path(res.Manifest.Name, res.Manifest.Files[0])
	if info, err := os.Stat(copied); err != nil || info.Mode().Perm() != 0o400 {
		t.Errorf("%s: %v %v", copied, info.Mode(), err)
	}
	// the live configuration was only read.
	if body, _ := os.ReadFile(filepath.Join(h, ".claude", "settings.json")); string(body) != `{"effortLevel":"medium"}` {
		t.Errorf("the live settings changed: %s", body)
	}
}

// CLAUDE_CONFIG_DIR and CODEX_HOME, from the operator's shell environment,
// move where the configuration is read; claude keeps .claude.json in its
// config directory then.
func TestTheConfigDirectoriesFollowTheEnvironment(t *testing.T) {
	h := home(t)
	custom := t.TempDir()
	write(t, filepath.Join(custom, "claude", "settings.json"), "{}")
	write(t, filepath.Join(custom, "claude", ".claude.json"), "{}")
	write(t, filepath.Join(custom, "codex", "config.toml"), "")
	env := []string{"HOME=" + h, "CLAUDE_CONFIG_DIR=" + filepath.Join(custom, "claude"), "CODEX_HOME=" + filepath.Join(custom, "codex")}
	s, _ := newStore(t, 5, 5)
	res, err := s.Take(roots(env), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	want := "claude/config/settings.json\nclaude/global/.claude.json\ncodex/home/config.toml"
	if got := strings.Join(names(res.Manifest), "\n"); got != want {
		t.Errorf("copied:\n%s\nwant:\n%s", got, want)
	}
	for _, r := range res.Manifest.Roots {
		if !strings.HasPrefix(r.Dir, custom) {
			t.Errorf("root %s/%s read from %s", r.Harness, r.Label, r.Dir)
		}
	}
}

// a link is copied as a link, with its target, and never followed: not a
// link to a file outside the configuration directory, nor one to a
// directory, whose files stay out.
func TestALinkIsCopiedAsALinkAndNeverFollowed(t *testing.T) {
	h := home(t)
	outside := t.TempDir()
	write(t, filepath.Join(outside, "secret.md"), "made-up secret outside")
	write(t, filepath.Join(outside, "tree", "inner.md"), "made-up secret in a tree")
	if err := os.Symlink(filepath.Join(outside, "secret.md"), filepath.Join(h, ".claude", "agents", "linked.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "tree"), filepath.Join(h, ".claude", "skills", "linked-tree")); err != nil {
		t.Fatal(err)
	}
	s, _ := newStore(t, 5, 5)
	res, err := s.Take(roots([]string{"HOME=" + h}), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	links := map[string]string{}
	for _, e := range res.Manifest.Files {
		if strings.Contains(e.Path, "inner.md") {
			t.Errorf("followed a linked tree: %s", e.Path)
		}
		if e.Kind == KindLink {
			links[e.Path] = e.Target
		}
	}
	if links["agents/linked.md"] != filepath.Join(outside, "secret.md") || links["skills/linked-tree"] != filepath.Join(outside, "tree") {
		t.Errorf("links = %v", links)
	}
	copied := filepath.Join(s.Dir(), res.Manifest.Name, "claude", "config", "agents", "linked.md")
	if info, err := os.Lstat(copied); err != nil || info.Mode()&fs.ModeSymlink == 0 {
		t.Errorf("the copy is not a link: %v %v", info, err)
	}
	_ = filepath.WalkDir(s.Dir(), func(path string, d fs.DirEntry, err error) error {
		if err == nil && d.Type().IsRegular() {
			if body, _ := os.ReadFile(path); strings.Contains(string(body), "made-up secret") {
				t.Errorf("%s holds a linked file's content", path)
			}
		}
		return nil
	})
}

// a start with nothing changed stores nothing; a change stores a snapshot
// whose unchanged files are the earlier snapshot's, by hard link, so only
// the changed file takes space.
func TestNothingChangedStoresNothingAndUnchangedFilesAreShared(t *testing.T) {
	h := home(t)
	env := []string{"HOME=" + h}
	s, now := newStore(t, 5, 5)
	first, err := s.Take(roots(env), nil, false)
	if err != nil || first.Manifest == nil {
		t.Fatal(err)
	}
	*now = now.Add(time.Hour)
	again, err := s.Take(roots(env), nil, false)
	if err != nil || again.Manifest != nil {
		t.Fatalf("an unchanged configuration was stored again: %+v, %v", again, err)
	}
	*now = now.Add(time.Hour)
	write(t, filepath.Join(h, ".claude.json"), `{"numStartups":2}`)
	second, err := s.Take(roots(env), nil, false)
	if err != nil || second.Manifest == nil {
		t.Fatal(err)
	}
	if second.NewFiles != 1 || second.NewBytes != int64(len(`{"numStartups":2}`)) {
		t.Errorf("stored %d new files, %d bytes; want the one changed file", second.NewFiles, second.NewBytes)
	}
	a, _ := os.Stat(filepath.Join(s.Dir(), first.Manifest.Name, "claude", "config", "settings.json"))
	b, _ := os.Stat(filepath.Join(s.Dir(), second.Manifest.Name, "claude", "config", "settings.json"))
	if a == nil || b == nil || !os.SameFile(a, b) {
		t.Errorf("an unchanged file is not shared between snapshots")
	}
	if list, _ := s.List(); len(list) != 2 || list[0].Name != second.Manifest.Name {
		t.Errorf("snapshots = %d, newest %v", len(list), list)
	}
}

// the store keeps a bounded number of ordinary snapshots, and of those
// taken at a version change, apart, so a version change outlives ordinary
// starts; an object no snapshot holds any more is removed.
func TestTheBoundOnHowManyAreKept(t *testing.T) {
	h := home(t)
	env := []string{"HOME=" + h}
	s, now := newStore(t, 2, 1)
	take := func(body string, versions map[string]string) *Manifest {
		t.Helper()
		*now = now.Add(time.Hour)
		write(t, filepath.Join(h, ".claude.json"), body)
		res, err := s.Take(roots(env), versions, false)
		if err != nil || res.Manifest == nil {
			t.Fatalf("take %s: %+v %v", body, res, err)
		}
		return res.Manifest
	}
	take(`{"n":1}`, map[string]string{model.HarnessClaude: "2.1.325"})
	atVersion := take(`{"n":2}`, map[string]string{model.HarnessClaude: "2.1.326"})
	if atVersion.Reason != ReasonVersion {
		t.Fatalf("a new version was taken as %q", atVersion.Reason)
	}
	for i := 3; i <= 6; i++ {
		take(`{"n":`+string(rune('0'+i))+`}`, map[string]string{model.HarnessClaude: "2.1.326"})
	}
	list, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	var reasons []string
	for _, m := range list {
		reasons = append(reasons, m.Reason)
	}
	if strings.Join(reasons, ",") != "start-up,start-up,version" || list[2].Name != atVersion.Name {
		t.Errorf("kept %v, want two ordinary and the version change", reasons)
	}
	// objects: one per distinct content the three kept snapshots hold.
	distinct := map[string]bool{}
	for _, m := range list {
		for _, e := range m.Files {
			distinct[e.SHA256] = true
		}
	}
	objects := 0
	_ = filepath.WalkDir(filepath.Join(s.Dir(), "objects"), func(_ string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			objects++
		}
		return nil
	})
	if objects != len(distinct) {
		t.Errorf("%d objects for %d distinct contents", objects, len(distinct))
	}
}

// a version no snapshot has recorded forces a snapshot, even with nothing
// changed; a version already recorded, or a harness seen for the first
// time, forces none, and a check for a new version alone takes no ordinary
// snapshot.
func TestANewVersionForcesASnapshot(t *testing.T) {
	h := home(t)
	env := []string{"HOME=" + h}
	s, now := newStore(t, 5, 5)
	if res, err := s.Take(roots(env), map[string]string{model.HarnessClaude: "2.1.325"}, false); err != nil || res.Manifest.Reason != ReasonStartUp {
		t.Fatalf("first: %+v %v", res, err)
	}
	*now = now.Add(time.Hour)
	if res, err := s.Take(roots(env), map[string]string{model.HarnessClaude: "2.1.325", model.HarnessCodex: "0.166.0"}, true); err != nil || res.Manifest != nil {
		t.Errorf("a harness seen for the first time forced a snapshot: %+v %v", res, err)
	}
	write(t, filepath.Join(h, ".claude.json"), `{"numStartups":9}`)
	if res, err := s.Take(roots(env), map[string]string{model.HarnessClaude: "2.1.325"}, true); err != nil || res.Manifest != nil {
		t.Errorf("a version check took an ordinary snapshot: %+v %v", res, err)
	}
	*now = now.Add(time.Hour)
	res, err := s.Take(roots(env), map[string]string{model.HarnessClaude: "2.1.326"}, true)
	if err != nil || res.Manifest == nil || res.Manifest.Reason != ReasonVersion || res.Manifest.Versions[model.HarnessClaude] != "2.1.326" {
		t.Fatalf("a new version: %+v %v", res, err)
	}
	*now = now.Add(time.Hour)
	if res, err := s.Take(roots(env), map[string]string{model.HarnessClaude: "2.1.326"}, true); err != nil || res.Manifest != nil {
		t.Errorf("a version already recorded forced another: %+v %v", res, err)
	}
	// and with nothing changed at all, a new version is still taken.
	*now = now.Add(time.Hour)
	if res, err := s.Take(roots(env), map[string]string{model.HarnessClaude: "2.1.327"}, false); err != nil || res.Manifest == nil || res.NewFiles != 0 {
		t.Errorf("a new version with nothing changed: %+v %v", res, err)
	}
}

// Diff names what differs between a snapshot and the live configuration,
// and the command that would put each snapshot copy back; it writes nothing.
func TestDiffNamesTheDifferencesAndTheCommands(t *testing.T) {
	h := home(t)
	env := []string{"HOME=" + h}
	s, _ := newStore(t, 5, 5)
	res, err := s.Take(roots(env), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	m := res.Manifest
	write(t, filepath.Join(h, ".claude", "settings.json"), `{"effortLevel":"high"}`)
	if err := os.Remove(filepath.Join(h, ".codex", "rules", "default.rules")); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(h, ".codex", "prompts", "new.md"), "new\n")
	diff, err := s.Diff(m)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Difference{}
	for _, d := range diff {
		got[d.Entry.Harness+"/"+d.Entry.Root+"/"+d.Entry.Path] = d
	}
	if len(got) != 3 {
		t.Errorf("differences: %+v", diff)
	}
	settings := got["claude/config/settings.json"]
	wantCmd := "install -D -m 0644 '" + filepath.Join(s.Dir(), m.Name, "claude", "config", "settings.json") + "' '" + filepath.Join(h, ".claude", "settings.json") + "'"
	if settings.Change != ChangeChanged || settings.Restore != wantCmd {
		t.Errorf("settings: %+v\nwant %s", settings, wantCmd)
	}
	if rules := got["codex/home/rules/default.rules"]; rules.Change != ChangeGone || !strings.HasPrefix(rules.Restore, "install -D -m 0644 ") || rules.Live != filepath.Join(h, ".codex", "rules", "default.rules") {
		t.Errorf("rules: %+v", rules)
	}
	if added := got["codex/home/prompts/new.md"]; added.Change != ChangeAdded || added.Restore != "" {
		t.Errorf("added: %+v", added)
	}
	if body, _ := os.ReadFile(filepath.Join(h, ".claude", "settings.json")); string(body) != `{"effortLevel":"high"}` {
		t.Errorf("diff wrote into the configuration: %s", body)
	}
}

// a home with no configuration in it has nothing to keep: no snapshot is
// taken, and nothing is stored.
func TestAnEmptyConfigurationStoresNothing(t *testing.T) {
	s, _ := newStore(t, 5, 5)
	res, err := s.Take(roots([]string{"HOME=" + t.TempDir()}), map[string]string{model.HarnessClaude: "2.1.325"}, false)
	if err != nil || res.Manifest != nil {
		t.Fatalf("an empty configuration: %+v %v", res, err)
	}
	if list, _ := s.List(); len(list) != 0 {
		t.Errorf("snapshots = %d", len(list))
	}
}
