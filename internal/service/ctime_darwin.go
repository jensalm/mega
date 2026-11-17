//go:build darwin

package service

import "syscall"

// changeTime is the inode change time of st, as seconds and nanoseconds.
func changeTime(st *syscall.Stat_t) (int64, int64) {
	return int64(st.Ctimespec.Sec), int64(st.Ctimespec.Nsec)
}
