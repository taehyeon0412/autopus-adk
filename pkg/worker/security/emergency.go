package security

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// EmergencyStop manages graceful-to-forceful subprocess termination.
type EmergencyStop struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	stopped bool
}

// NewEmergencyStop creates an EmergencyStop handler.
func NewEmergencyStop() *EmergencyStop {
	return &EmergencyStop{}
}

// SetProcess registers the active subprocess for emergency termination.
// The subprocess command should be configured with SysProcAttr{Setpgid: true}
// to enable process group termination.
func (e *EmergencyStop) SetProcess(cmd *exec.Cmd) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cmd = cmd
	e.stopped = false
}

// ClearProcess clears the registered subprocess after normal completion.
func (e *EmergencyStop) ClearProcess() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cmd = nil
}

// Stop terminates the subprocess: SIGTERM first, then SIGKILL after 5s.
// reason is logged for audit trail. Thread-safe — only the first call acts.
func (e *EmergencyStop) Stop(reason string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.stopped {
		return nil
	}

	if e.cmd == nil || e.cmd.Process == nil {
		return errors.New("no process registered for emergency stop")
	}

	e.stopped = true
	pid := e.cmd.Process.Pid
	pgid := -pid // Negative PID targets the process group.

	// Send SIGTERM to the process group.
	if err := syscall.Kill(pgid, syscall.SIGTERM); err != nil {
		// Process may have already exited.
		if errors.Is(err, syscall.ESRCH) {
			return nil
		}
		return fmt.Errorf("emergency stop SIGTERM (reason: %s): %w", reason, err)
	}

	// Wait for process exit or force kill after timeout.
	done := make(chan struct{})
	go func() {
		_ = e.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		// Process did not exit in time — escalate to SIGKILL.
		if err := syscall.Kill(pgid, syscall.SIGKILL); err != nil {
			if errors.Is(err, syscall.ESRCH) {
				return nil
			}
			return fmt.Errorf("emergency stop SIGKILL (reason: %s): %w", reason, err)
		}

		// Wait for the killed process to be reaped.
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			_, _ = fmt.Fprintf(os.Stderr, "[EMERGENCY] process %d did not exit after SIGKILL (reason: %s)\n", pid, reason)
		}
		return nil
	}
}
