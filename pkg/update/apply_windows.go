//go:build windows

package update

import (
	"os/exec"
	"syscall"
)

const (
	createNewConsole       = 0x00000010
	createNewProcessGroup  = 0x00000200
	createBreakawayFromJob = 0x01000000
)

func startDetached(name string, args ...string) error {
	// A new console stays on screen. Breakaway keeps it alive after this process quits.
	err := startConsole(name, args, createNewConsole|createNewProcessGroup|createBreakawayFromJob)
	if err != nil {
		return startConsole(name, args, createNewConsole|createNewProcessGroup)
	}
	return nil
}

func startConsole(name string, args []string, flags uint32) error {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: flags}
	return cmd.Start()
}
