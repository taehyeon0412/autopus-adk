package telemetry_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/insajin/autopus-adk/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRecorder_EmptySpecID_UsesUnknown verifies that an empty specID is
// normalized to "unknown" to avoid creating a file with a bare date prefix.
func TestNewRecorder_EmptySpecID_UsesUnknown(t *testing.T) {
	dir := t.TempDir()
	r, err := telemetry.NewRecorder(dir, "")
	require.NoError(t, err)
	require.NotNil(t, r)

	telDir := filepath.Join(dir, ".autopus", "telemetry")
	entries, err := os.ReadDir(telDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Contains(t, entries[0].Name(), "unknown")
}

// TestNewRecorder_DotSpecID_UsesUnknown verifies that "." is normalized to "unknown".
func TestNewRecorder_DotSpecID_UsesUnknown(t *testing.T) {
	dir := t.TempDir()
	r, err := telemetry.NewRecorder(dir, ".")
	require.NoError(t, err)
	require.NotNil(t, r)

	telDir := filepath.Join(dir, ".autopus", "telemetry")
	entries, err := os.ReadDir(telDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Contains(t, entries[0].Name(), "unknown")
}

// TestRecorder_DoubleFinalize_IsIdempotent ensures a second Finalize call does
// not panic and returns consistent data (file is nil after the first close).
func TestRecorder_DoubleFinalize_IsIdempotent(t *testing.T) {
	dir := t.TempDir()
	r, err := telemetry.NewRecorder(dir, "SPEC-DOUBLE-001")
	require.NoError(t, err)

	r.StartPipeline("SPEC-DOUBLE-001", "balanced")
	run1 := r.Finalize(telemetry.StatusPass)

	// Second Finalize must not panic; result should still be valid.
	assert.NotPanics(t, func() {
		run2 := r.Finalize(telemetry.StatusPass)
		assert.Equal(t, run1.SpecID, run2.SpecID)
	})
}

// TestRecorder_RecordAgentWithoutStartPhase_DoesNotPanic ensures RecordAgent is
// safe when called before any StartPhase (currentPhase == nil).
func TestRecorder_RecordAgentWithoutStartPhase_DoesNotPanic(t *testing.T) {
	dir := t.TempDir()
	r, err := telemetry.NewRecorder(dir, "SPEC-NOPHA-001")
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		r.RecordAgent(telemetry.AgentRun{AgentName: "executor", Status: telemetry.StatusPass})
	})

	// Agent must NOT appear in the final pipeline (no phase was opened).
	run := r.Finalize(telemetry.StatusPass)
	assert.Empty(t, run.Phases)
}

// TestRecorder_ConcurrentWrites_DoesNotRace exercises concurrent RecordAgent
// calls to verify that the mutex prevents data races (run with -race).
func TestRecorder_ConcurrentWrites_DoesNotRace(t *testing.T) {
	dir := t.TempDir()
	r, err := telemetry.NewRecorder(dir, "SPEC-CONC-001")
	require.NoError(t, err)

	r.StartPipeline("SPEC-CONC-001", "ultra")
	r.StartPhase("RED")

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			r.RecordAgent(telemetry.AgentRun{
				AgentName: "tester",
				StartTime: time.Now(),
				Status:    telemetry.StatusPass,
			})
		}()
	}
	wg.Wait()

	r.EndPhase(telemetry.StatusPass)
	run := r.Finalize(telemetry.StatusPass)

	require.Len(t, run.Phases, 1)
	assert.Len(t, run.Phases[0].Agents, goroutines)
}

// TestRecorder_CleanExpired_DirNotExist_ReturnsError verifies that
// CleanExpired surfaces a meaningful error when the telemetry dir is absent.
func TestRecorder_CleanExpired_DirNotExist_ReturnsError(t *testing.T) {
	// Use a baseDir where the telemetry sub-directory does NOT exist.
	baseDir := t.TempDir()
	r, err := telemetry.NewRecorder(baseDir, "SPEC-CLEAN-002")
	require.NoError(t, err)

	// Remove the whole .autopus directory that NewRecorder just created.
	require.NoError(t, os.RemoveAll(filepath.Join(baseDir, ".autopus")))

	err = r.CleanExpired(7)
	assert.Error(t, err, "CleanExpired should fail when the telemetry dir does not exist")
}
