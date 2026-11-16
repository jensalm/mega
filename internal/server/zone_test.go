package server

import (
	"testing"
	"time"
)

// the service names its zone from TZ, in each form TZ takes, and keeps a
// name only where it agrees with the local offset; one it cannot name is
// empty, and the page formats in the browser's zone.
func TestTheServiceNamesItsTimeZone(t *testing.T) {
	in := func(name string) time.Time {
		loc, err := time.LoadLocation(name)
		if err != nil {
			t.Skipf("no zone data for %s: %v", name, err)
		}
		return time.Date(2026, 11, 16, 12, 0, 0, 0, loc)
	}
	tz := func(value string) func(string) (string, bool) {
		return func(key string) (string, bool) {
			if key == "TZ" {
				return value, true
			}
			return "", false
		}
	}
	for _, c := range []struct {
		tz   string
		now  time.Time
		want string
	}{
		{"Asia/Tokyo", in("Asia/Tokyo"), "Asia/Tokyo"},
		{":Asia/Tokyo", in("Asia/Tokyo"), "Asia/Tokyo"},
		{"/usr/share/zoneinfo/Europe/Paris", in("Europe/Paris"), "Europe/Paris"},
		{"", in("UTC"), "UTC"},
		// a name that disagrees with the local offset is not trusted.
		{"Asia/Tokyo", in("UTC"), ""},
		{"Not/AZone", in("UTC"), ""},
		{"/somewhere/else", in("UTC"), ""},
	} {
		if got := timeZone(tz(c.tz), c.now); got != c.want {
			t.Errorf("TZ=%q: %q, want %q", c.tz, got, c.want)
		}
	}
}
