package pipeline_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/pipeline"
)

// TestCheckpointStatus_String verifies that CheckpointStatus converts to its
// canonical string representation for each defined status value.
func TestCheckpointStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   pipeline.CheckpointStatus
		expected string
	}{
		{
			name:     "pending status",
			status:   pipeline.CheckpointStatusPending,
			expected: "pending",
		},
		{
			name:     "in_progress status",
			status:   pipeline.CheckpointStatusInProgress,
			expected: "in_progress",
		},
		{
			name:     "done status",
			status:   pipeline.CheckpointStatusDone,
			expected: "done",
		},
		{
			name:     "failed status",
			status:   pipeline.CheckpointStatusFailed,
			expected: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Given: a CheckpointStatus value
			// When: String() is called
			// Then: the expected string representation is returned
			got := tt.status.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestCheckpointYAML_RoundTrip verifies that a Checkpoint can be marshalled to
// YAML and unmarshalled back without data loss.
func TestCheckpointYAML_RoundTrip(t *testing.T) {
	t.Parallel()

	// Given: a Checkpoint with known fields
	original := pipeline.Checkpoint{
		Phase:         "phase2",
		GitCommitHash: "abc123def456",
		TaskStatus: map[string]pipeline.CheckpointStatus{
			"task-1": pipeline.CheckpointStatusDone,
			"task-2": pipeline.CheckpointStatusInProgress,
		},
	}

	// When: marshalled to YAML and back
	data, err := original.MarshalYAML()
	require.NoError(t, err)

	var restored pipeline.Checkpoint
	err = restored.UnmarshalYAML(data)
	require.NoError(t, err)

	// Then: all fields are preserved
	assert.Equal(t, original.Phase, restored.Phase)
	assert.Equal(t, original.GitCommitHash, restored.GitCommitHash)
	assert.Equal(t, original.TaskStatus, restored.TaskStatus)
}

// TestCheckpointYAML_RoundTrip_NilTaskStatus verifies that a Checkpoint with
// nil TaskStatus survives a YAML round-trip without panicking.
func TestCheckpointYAML_RoundTrip_NilTaskStatus(t *testing.T) {
	t.Parallel()

	// Given: a Checkpoint with nil TaskStatus
	original := pipeline.Checkpoint{
		Phase:         "phase1",
		GitCommitHash: "nil-task-hash",
		TaskStatus:    nil,
	}

	// When: marshalled to YAML and back
	data, err := original.MarshalYAML()
	require.NoError(t, err)

	var restored pipeline.Checkpoint
	err = restored.UnmarshalYAML(data)
	require.NoError(t, err)

	// Then: no panic and Phase/Hash are preserved
	assert.Equal(t, original.Phase, restored.Phase)
	assert.Equal(t, original.GitCommitHash, restored.GitCommitHash)
}

// TestCheckpointUnmarshalYAML_InvalidYAML verifies that UnmarshalYAML returns
// an error for malformed YAML input.
func TestCheckpointUnmarshalYAML_InvalidYAML(t *testing.T) {
	t.Parallel()

	// Given: malformed YAML bytes
	bad := []byte("phase: [invalid: yaml: here")

	// When: UnmarshalYAML is called
	var cp pipeline.Checkpoint
	err := cp.UnmarshalYAML(bad)

	// Then: an error is returned
	require.Error(t, err)
}
