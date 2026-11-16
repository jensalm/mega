package pi

import (
	"os"
	"path/filepath"

	"github.com/michaelquigley/df/dd"
)

// protected is what pi loads from a checkout only once the checkout is
// trusted: its .pi settings, MCP servers, extensions, skills, prompts,
// themes and system prompt files, and .agents/skills, in the checkout or an
// ancestor (pi 1.5.2, docs/security.md).
var protected = []string{
	".pi/settings.json", ".pi/mcp.json", ".pi/extensions", ".pi/skills", ".pi/prompts",
	".pi/themes", ".pi/SYSTEM.md", ".pi/APPEND_SYSTEM.md", ".agents/skills",
}

// trustNotice says, for the start of a runtime, that pi runs this checkout
// without its project resources because the operator has not trusted it in
// pi's own terminal: empty when the checkout has none, when a saved decision
// trusts it, or when pi's defaultProjectTrust is "always". it reads pi's
// trust file and settings, and writes neither: trusting a checkout is the
// operator's, in pi's own terminal, and mega never passes --approve.
func trustNotice(env []string, cwd string) string {
	dir, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		dir = cwd
	}
	if !hasProtected(dir) {
		return ""
	}
	settings := &struct {
		DefaultProjectTrust string `dd:"defaultProjectTrust"`
	}{}
	_ = dd.BindJSONFile(settings, filepath.Join(agentDir(env), "settings.json"))
	if settings.DefaultProjectTrust == "always" {
		return ""
	}
	decisions := map[string]any{}
	if data, err := os.ReadFile(filepath.Join(agentDir(env), "trust.json")); err == nil {
		if decoded, err := dd.DecodeStrictJSON(data); err == nil {
			decisions = decoded
		}
	}
	// the nearest saved decision, for the checkout or an ancestor.
	for d := dir; ; d = filepath.Dir(d) {
		if v, ok := decisions[d]; ok {
			if trusted, _ := v.(bool); trusted {
				return ""
			}
			break
		}
		if d == filepath.Dir(d) {
			break
		}
	}
	return "pi has not been trusted in this checkout, so its project extensions, settings and MCP servers did not load. pi asks only in its own terminal: run pi in the checkout once to decide; mega never trusts a checkout for you"
}

func hasProtected(dir string) bool {
	for d := dir; ; d = filepath.Dir(d) {
		for _, p := range protected {
			if _, err := os.Stat(filepath.Join(d, p)); err == nil {
				return true
			}
		}
		if d == filepath.Dir(d) {
			return false
		}
	}
}
