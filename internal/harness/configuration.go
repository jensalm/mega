package harness

import (
	"os"
	"strings"
)

// ConfigRoot is one directory where a harness keeps the operator's
// configuration, and which of it a configuration snapshot copies. Include
// names what is copied, relative to Dir: a file ("settings.json"), a whole
// tree ("skills/"), or a glob of the directory's own entries ("*.toml").
// anything Include does not name is never copied, so a file a new harness
// version adds stays out until someone decides about it. Exclude names
// trees under an included one that are not copied either.
//
// an include list names configuration only: never a credential or token
// file, a transcript, a session store, a history, a database or a cache.
type ConfigRoot struct {
	// Harness is the harness's name, and Label the folder this root takes
	// in a snapshot ("config", "home").
	Harness string
	Label   string
	Dir     string
	Include []string
	Exclude []string
}

// Configurer is a harness that can say where the operator's configuration
// for it is, in env (the operator's shell environment, as the start-up check
// resolves it), and which of it a configuration snapshot copies.
type Configurer interface {
	Configuration(env []string) []ConfigRoot
}

// Lookup reads a variable from an environment list.
func Lookup(env []string, name string) (string, bool) {
	for _, kv := range env {
		if n, v, ok := strings.Cut(kv, "="); ok && n == name {
			return v, true
		}
	}
	return "", false
}

// Home is the home directory env names, else mega's own.
func Home(env []string) string {
	if home, ok := Lookup(env, "HOME"); ok && home != "" {
		return home
	}
	home, _ := os.UserHomeDir()
	return home
}
