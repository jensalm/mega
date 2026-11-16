package server

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// timeZone names the time zone the service runs in, as the IANA database
// names it, such as America/New_York, so the page formats every time in the
// operator's zone even where the browser reports another: a hardened browser
// may tell a page it is in UTC. it is TZ when set, else the zone the
// /etc/localtime link names, else /etc/timezone; a name is kept only where
// it loads and agrees with the local offset now. empty when none can be
// named, and the page then formats in the browser's own zone.
func timeZone(getenv func(string) (string, bool), now time.Time) string {
	var candidates []string
	if tz, ok := getenv("TZ"); ok {
		tz = strings.TrimPrefix(tz, ":")
		if tz == "" {
			// TZ set and empty is UTC, as Go and the C library read it.
			tz = "UTC"
		}
		candidates = append(candidates, zoneName(tz))
	} else {
		if target, err := filepath.EvalSymlinks("/etc/localtime"); err == nil {
			candidates = append(candidates, zoneName(target))
		}
		if data, err := os.ReadFile("/etc/timezone"); err == nil {
			candidates = append(candidates, strings.TrimSpace(string(data)))
		}
	}
	_, local := now.Zone()
	for _, name := range candidates {
		if name == "" || name == "Local" {
			continue
		}
		loc, err := time.LoadLocation(name)
		if err != nil {
			continue
		}
		if _, offset := now.In(loc).Zone(); offset == local {
			return name
		}
	}
	return ""
}

// zoneName is the zone a TZ value or a zoneinfo path names: the part after
// zoneinfo/ for a path, the value itself otherwise.
func zoneName(value string) string {
	if i := strings.LastIndex(value, "zoneinfo/"); i >= 0 {
		return value[i+len("zoneinfo/"):]
	}
	if strings.HasPrefix(value, "/") {
		return ""
	}
	return value
}
