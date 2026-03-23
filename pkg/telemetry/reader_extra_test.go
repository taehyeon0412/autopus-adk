package telemetry_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/insajin/autopus-adk/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadEvents_MalformedJSON_ReturnsError verifies that a line containing
// invalid JSON causes LoadEvents to return a non-nil error.
func TestLoadEvents_MalformedJSON_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.jsonl")
	require.NoError(t, os.WriteFile(path, []byte("this is not json\n"), 0o644))

	_, err := telemetry.LoadEvents(path)
	assert.Error(t, err, "malformed JSON line must return an error")
}

// TestLoadEvents_EmptyFile_ReturnsEmpty verifies that an empty JSONL file
// produces an empty slice without an error.
func TestLoadEvents_EmptyFile_ReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(""), 0o644))

	events, err := telemetry.LoadEvents(path)
	require.NoError(t, err)
	assert.Empty(t, events)
}

// TestLoadAllEvents_SkipsNonJsonlFiles verifies that plain files and
// sub-directories inside the telemetry dir are ignored gracefully.
func TestLoadAllEvents_SkipsNonJsonlFiles(t *testing.T) {
	baseDir := t.TempDir()
	telDir := filepath.Join(baseDir, ".autopus", "telemetry")
	require.NoError(t, os.MkdirAll(telDir, 0o755))

	// Write a valid JSONL file.
	now := time.Now().UTC()
	event := telemetry.Event{
		Type:      telemetry.EventTypeAgentRun,
		Timestamp: now,
		Data:      json.RawMessage(`{}`),
	}
	line, err := json.Marshal(event)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(
		filepath.Join(telDir, "valid.jsonl"),
		append(line, '\n'),
		0o644,
	))

	// Also write a non-JSONL file that should be ignored.
	require.NoError(t, os.WriteFile(
		filepath.Join(telDir, "readme.txt"),
		[]byte("not a jsonl file"),
		0o644,
	))

	// Create a sub-directory that should be skipped.
	require.NoError(t, os.MkdirAll(filepath.Join(telDir, "subdir"), 0o755))

	events, err := telemetry.LoadAllEvents(baseDir)
	require.NoError(t, err)
	assert.Len(t, events, 1, "only the .jsonl file should be loaded")
}

// TestLoadAllEvents_MalformedJSONL_ReturnsError verifies that a malformed
// line in any JSONL file causes LoadAllEvents to propagate the error.
func TestLoadAllEvents_MalformedJSONL_ReturnsError(t *testing.T) {
	baseDir := t.TempDir()
	telDir := filepath.Join(baseDir, ".autopus", "telemetry")
	require.NoError(t, os.MkdirAll(telDir, 0o755))

	require.NoError(t, os.WriteFile(
		filepath.Join(telDir, "bad.jsonl"),
		[]byte("{{invalid json}}\n"),
		0o644,
	))

	_, err := telemetry.LoadAllEvents(baseDir)
	assert.Error(t, err)
}

// TestLatestPipelineRun_MalformedData_ReturnsError verifies that an event
// whose Data field is not a valid PipelineRun returns an error.
func TestLatestPipelineRun_MalformedData_ReturnsError(t *testing.T) {
	baseDir := t.TempDir()

	// Write a pipeline_end event with corrupt data.
	badEvent := telemetry.Event{
		Type:      telemetry.EventTypePipelineEnd,
		Timestamp: time.Now().UTC(),
		Data:      json.RawMessage(`"not an object"`),
	}
	writeTelemetryJSONL(t, baseDir, "bad_pipeline.jsonl", []telemetry.Event{badEvent})

	_, err := telemetry.LatestPipelineRun(baseDir)
	assert.Error(t, err, "malformed pipeline_end data must return an error")
}

// TestPipelineRunsBySpecID_MalformedData_ReturnsError verifies that corrupt
// pipeline_end data causes PipelineRunsBySpecID to return an error.
func TestPipelineRunsBySpecID_MalformedData_ReturnsError(t *testing.T) {
	baseDir := t.TempDir()

	badEvent := telemetry.Event{
		Type:      telemetry.EventTypePipelineEnd,
		Timestamp: time.Now().UTC(),
		Data:      json.RawMessage(`"not an object"`),
	}
	writeTelemetryJSONL(t, baseDir, "bad_pipeline.jsonl", []telemetry.Event{badEvent})

	_, err := telemetry.PipelineRunsBySpecID(baseDir, "SPEC-ANY")
	assert.Error(t, err, "malformed pipeline_end data must return an error")
}
