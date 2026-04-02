//go:build !windows

package security

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmergencyStopNilProcess(t *testing.T) {
	t.Parallel()

	es := NewEmergencyStop()
	err := es.Stop("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no process registered")
}

func TestEmergencyStopSIGTERM(t *testing.T) {
	t.Parallel()

	// Start a process that exits on SIGTERM (default behavior).
	cmd := exec.Command("sleep", "60")
	cmd.SysProcAttr = sysProcAttr()
	require.NoError(t, cmd.Start())

	es := NewEmergencyStop()
	es.SetProcess(cmd)

	err := es.Stop("test termination")
	assert.NoError(t, err)
}

func TestEmergencyStopClearThenStop(t *testing.T) {
	t.Parallel()

	cmd := exec.Command("sleep", "60")
	cmd.SysProcAttr = sysProcAttr()
	require.NoError(t, cmd.Start())

	es := NewEmergencyStop()
	es.SetProcess(cmd)
	es.ClearProcess()

	err := es.Stop("after clear")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no process registered")

	// Clean up the orphaned process.
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
}

func TestEmergencyStopConcurrent(t *testing.T) {
	t.Parallel()

	cmd := exec.Command("sleep", "60")
	cmd.SysProcAttr = sysProcAttr()
	require.NoError(t, cmd.Start())

	es := NewEmergencyStop()
	es.SetProcess(cmd)

	var wg sync.WaitGroup
	errs := make([]error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errs[idx] = es.Stop("concurrent test")
		}(i)
	}
	wg.Wait()

	// All calls should succeed (first acts, rest return nil due to stopped flag).
	for i, err := range errs {
		assert.NoError(t, err, "goroutine %d should not error", i)
	}
}

func TestEmergencyStopSIGKILLEscalation(t *testing.T) {
	t.Parallel()

	// Start a process that traps SIGTERM (ignores it).
	// Write a marker file when trap is installed so we know it's ready.
	marker := filepath.Join(t.TempDir(), "ready")
	script := fmt.Sprintf("trap '' TERM; touch %s; sleep 60", marker)
	cmd := exec.Command("bash", "-c", script)
	cmd.SysProcAttr = sysProcAttr()
	require.NoError(t, cmd.Start())

	// Wait for the trap to be installed.
	for i := 0; i < 50; i++ {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	es := NewEmergencyStop()
	es.SetProcess(cmd)

	// Stop should escalate to SIGKILL after 5s timeout.
	err := es.Stop("escalation test")
	assert.NoError(t, err)
}

func TestEmergencyStopAlreadyStopped(t *testing.T) {
	t.Parallel()

	cmd := exec.Command("sleep", "60")
	cmd.SysProcAttr = sysProcAttr()
	require.NoError(t, cmd.Start())

	es := NewEmergencyStop()
	es.SetProcess(cmd)

	// First stop.
	err := es.Stop("first")
	assert.NoError(t, err)

	// Second stop should return nil (already stopped).
	err = es.Stop("second")
	assert.NoError(t, err)
}

func TestEmergencyStopProcessAlreadyExited(t *testing.T) {
	t.Parallel()

	// Start a process that exits immediately.
	cmd := exec.Command("true")
	cmd.SysProcAttr = sysProcAttr()
	require.NoError(t, cmd.Start())
	// Wait for it to exit.
	_ = cmd.Wait()

	es := NewEmergencyStop()
	es.SetProcess(cmd)

	// Stop should handle ESRCH gracefully.
	err := es.Stop("already exited")
	assert.NoError(t, err)
}

func TestEmergencyStopSetProcessReplaces(t *testing.T) {
	t.Parallel()

	cmd1 := exec.Command("sleep", "60")
	cmd1.SysProcAttr = sysProcAttr()
	require.NoError(t, cmd1.Start())

	cmd2 := exec.Command("sleep", "60")
	cmd2.SysProcAttr = sysProcAttr()
	require.NoError(t, cmd2.Start())

	es := NewEmergencyStop()
	es.SetProcess(cmd1)
	es.SetProcess(cmd2) // Replaces cmd1.

	err := es.Stop("replace test")
	assert.NoError(t, err)

	// Clean up cmd1 which was replaced and not stopped.
	_ = cmd1.Process.Kill()
	_ = cmd1.Wait()
}
