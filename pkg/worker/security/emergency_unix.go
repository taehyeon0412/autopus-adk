//go:build !windows

package security

import (
	"errors"
	"fmt"
	"syscall"
)

// sendSignal sends a signal to the process group identified by pgid.
func sendSignal(pgid int, sig syscall.Signal, reason string) error {
	if err := syscall.Kill(pgid, sig); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return nil
		}
		return fmt.Errorf("emergency stop %s (reason: %s): %w", sig, reason, err)
	}
	return nil
}
