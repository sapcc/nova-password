//go:build windows
// +build windows

package main

import (
	"syscall"

	"golang.org/x/term"
)

func readPassword(fd syscall.Handle) ([]byte, error) {
	return term.ReadPassword(int(fd))
}
