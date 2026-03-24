package pipeline_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/pipeline"
)

// TestCheckpoint_Save_CreatesYAMLFile verifies that Save writes a valid YAML
// file containing Phase, TaskStatus, and GitCommitHash fields.
func TestCheckpoint_Save_CreatesYAMLFile(t *testing.T) {
	t.Parallel()

	// Given: a Checkpoint with known data and a temp directory
	dir := t.TempDir()
	cp := pipeline.Checkpoint{
		Phase:         "phase2",
		GitCommitHash: "deadbeef1234",
		TaskStatus: map[string]pipeline.CheckpointStatus{
			"task-a": pipeline.CheckpointStatusDone,
		},
	}

	// When: Save is called
	err := cp.Save(dir)
	require.NoError(t, err)

	// Then: a YAML file exists with the expected content
	path := filepath.Join(dir, ".autopus-checkpoint.yaml")
	data, readErr := os.ReadFile(path)
	require.NoError(t, readErr)

	content := string(data)
	assert.Contains(t, content, "phase2")
	assert.Contains(t, content, "deadbeef1234")
	assert.Contains(t, content, "task-a")
	assert.Contains(t, content, "done")
}

// TestCheckpoint_Load_ReadsExistingCheckpoint verifies that Load correctly
// parses a valid YAML checkpoint file.
func TestCheckpoint_Load_ReadsExistingCheckpoint(t *testing.T) {
	t.Parallel()

	// Given: a YAML checkpoint file on disk
	dir := t.TempDir()
	yamlContent := `phase: phase1
git_commit_hash: abc123
task_status:
  task-x: done
  task-y: in_progress
`
	path := filepath.Join(dir, ".autopus-checkpoint.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yamlContent), 0o644))

	// When: Load is called
	cp, err := pipeline.Load(dir)
	require.NoError(t, err)

	// Then: the checkpoint fields match the YAML content
	assert.Equal(t, "phase1", cp.Phase)
	assert.Equal(t, "abc123", cp.GitCommitHash)
	assert.Equal(t, pipeline.CheckpointStatusDone, cp.TaskStatus["task-x"])
	assert.Equal(t, pipeline.CheckpointStatusInProgress, cp.TaskStatus["task-y"])
}

// TestCheckpoint_Load_StaleCheckpoint_WarnsOnHashMismatch verifies that Load
// detects when the saved git commit hash differs from the current HEAD and
// returns a stale warning.
func TestCheckpoint_Load_StaleCheckpoint_WarnsOnHashMismatch(t *testing.T) {
	t.Parallel()

	// Given: a checkpoint saved with a different git hash than the current HEAD
	dir := t.TempDir()
	yamlContent := `phase: phase2
git_commit_hash: staleoldhash999
task_status:
  task-z: done
`
	path := filepath.Join(dir, ".autopus-checkpoint.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yamlContent), 0o644))

	// When: Load is called with a mismatched current hash
	cp, err := pipeline.LoadWithHash(dir, "newcurrenthash111")
	require.NoError(t, err)

	// Then: the checkpoint is returned with Stale=true
	assert.True(t, cp.Stale)
	assert.Equal(t, "staleoldhash999", cp.GitCommitHash)
}

// TestCheckpoint_Load_FileNotFound_ReturnsError verifies that Load returns an
// error when no checkpoint file exists in the given directory.
func TestCheckpoint_Load_FileNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: a directory with no checkpoint file
	dir := t.TempDir()

	// When: Load is called
	_, err := pipeline.Load(dir)

	// Then: an error is returned
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checkpoint")
}

// TestCheckpoint_Save_NonExistentDir_ReturnsError verifies that Save returns
// an error when the target directory does not exist.
func TestCheckpoint_Save_NonExistentDir_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: a non-existent directory path
	dir := filepath.Join(t.TempDir(), "nonexistent", "deep", "path")
	cp := pipeline.Checkpoint{Phase: "phase1"}

	// When: Save is called
	err := cp.Save(dir)

	// Then: an error is returned
	require.Error(t, err)
}

// TestCheckpoint_LoadWithHash_MatchingHash_NotStale verifies that LoadWithHash
// sets Stale=false when the saved hash matches the current hash.
func TestCheckpoint_LoadWithHash_MatchingHash_NotStale(t *testing.T) {
	t.Parallel()

	// Given: a checkpoint file with a known hash
	dir := t.TempDir()
	yamlContent := "phase: phase1\ngit_commit_hash: abc123\ntask_status: {}\n"
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ".autopus-checkpoint.yaml"),
		[]byte(yamlContent), 0o644),
	)

	// When: LoadWithHash is called with the same hash
	cp, err := pipeline.LoadWithHash(dir, "abc123")
	require.NoError(t, err)

	// Then: Stale is false
	assert.False(t, cp.Stale)
}

// TestCheckpoint_LoadWithHash_MissingFile_ReturnsError verifies that
// LoadWithHash propagates errors from Load when the file is missing.
func TestCheckpoint_LoadWithHash_MissingFile_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: an empty directory (no checkpoint file)
	dir := t.TempDir()

	// When: LoadWithHash is called
	_, err := pipeline.LoadWithHash(dir, "anyhash")

	// Then: an error is returned
	require.Error(t, err)
}

// TestCheckpoint_Load_InvalidYAML_ReturnsError verifies that Load returns an
// error when the checkpoint file contains invalid YAML.
func TestCheckpoint_Load_InvalidYAML_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: a checkpoint file with malformed YAML
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ".autopus-checkpoint.yaml"),
		[]byte("phase: [broken: yaml: content"),
		0o644,
	))

	// When: Load is called
	_, err := pipeline.Load(dir)

	// Then: an error is returned
	require.Error(t, err)
}

// TestCheckpoint_Save_EmptyTaskStatus verifies that Save works correctly when
// TaskStatus is an empty (non-nil) map.
func TestCheckpoint_Save_EmptyTaskStatus(t *testing.T) {
	t.Parallel()

	// Given: a Checkpoint with an empty TaskStatus map
	dir := t.TempDir()
	cp := pipeline.Checkpoint{
		Phase:         "phase3",
		GitCommitHash: "emptymap123",
		TaskStatus:    map[string]pipeline.CheckpointStatus{},
	}

	// When: Save is called
	err := cp.Save(dir)
	require.NoError(t, err)

	// Then: the file exists
	_, statErr := os.Stat(filepath.Join(dir, ".autopus-checkpoint.yaml"))
	assert.NoError(t, statErr)
}
