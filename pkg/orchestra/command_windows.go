//go:build windows

package orchestra

import (
	"os"
	"os/exec"
)

func configureCommand(_ *exec.Cmd) {}

func terminateCommand(cmd *exec.Cmd, _ string) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	proc, err := os.FindProcess(cmd.Process.Pid)
	if err != nil {
		return nil
	}
	return proc.Kill()
}
