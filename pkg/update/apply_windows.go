//go:build windows

package update

import (
	"os/exec"
	"syscall"
)

const (
	createNoWindow         = 0x08000000
	createNewProcessGroup  = 0x00000200
	createBreakawayFromJob = 0x01000000
)

func startDetached(name string, args ...string) error {
	// No console. Breakaway keeps the helper alive after this process quits.
	err := startHidden(name, args, createNoWindow|createNewProcessGroup|createBreakawayFromJob)
	if err != nil {
		return startHidden(name, args, createNoWindow|createNewProcessGroup)
	}
	return nil
}

func startHidden(name string, args []string, flags uint32) error {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: flags,
	}
	return cmd.Start()
}
