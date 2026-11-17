//go:build darwin

package main

import "golang.org/x/sys/unix"

// getTermios is the ioctl request that reads a terminal's attributes.
const getTermios = unix.TIOCGETA
