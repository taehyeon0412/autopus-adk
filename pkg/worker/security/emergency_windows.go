//go:build windows

package security

import (
	"fmt"
	"os"
	"syscall"
)

// sendSignal terminates the process on Windows using os.Process.Kill.
// Windows does not support Unix signals or process groups.
func sendSignal(pgid int, _ syscall.Signal, reason string) error {
	// pgid is negative on Unix; use absolute value for the PID.
	pid := pgid
	if pid < 0 {
		pid = -pid
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil // process not found — already exited
	}

	if err := proc.Kill(); err != nil {
		return fmt.Errorf("emergency stop kill (reason: %s): %w", reason, err)
	}
	return nil
}
