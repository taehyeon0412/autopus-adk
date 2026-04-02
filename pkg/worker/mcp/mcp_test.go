package mcp

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockReporter captures progress reports for testing.
type mockReporter struct {
	mu      sync.Mutex
	reports []ProgressReport
	err     error
}

func (m *mockReporter) ReportProgress(report ProgressReport) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.reports = append(m.reports, report)
	return nil
}

func TestProgressReport_MarshalJSON(t *testing.T) {
	t.Parallel()

	report := ProgressReport{
		TaskID: "task-100",
		Phase:  "execution",
		Status: "in_progress",
		Details: map[string]any{
			"step":  3,
			"total": 4,
		},
	}

	data, err := json.Marshal(report)
	require.NoError(t, err)

	var decoded ProgressReport
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "task-100", decoded.TaskID)
	assert.Equal(t, "execution", decoded.Phase)
	assert.Equal(t, "in_progress", decoded.Status)
	assert.Equal(t, float64(3), decoded.Details["step"])
	assert.Equal(t, float64(4), decoded.Details["total"])
}

func TestProgressReport_MinimalFields(t *testing.T) {
	t.Parallel()

	report := ProgressReport{
		TaskID: "task-min",
		Phase:  "done",
		Status: "completed",
	}

	data, err := json.Marshal(report)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "task-min", decoded["task_id"])
	assert.Equal(t, "done", decoded["phase"])
	assert.Equal(t, "completed", decoded["status"])
	// Optional field should be omitted.
	_, hasDetails := decoded["details"]
	assert.False(t, hasDetails, "details should be omitted when nil")
}

func TestMockReporter_ReportProgress(t *testing.T) {
	t.Parallel()

	reporter := &mockReporter{}
	reports := []ProgressReport{
		{TaskID: "t1", Phase: "init", Status: "started"},
		{TaskID: "t1", Phase: "execution", Status: "in_progress"},
		{TaskID: "t1", Phase: "done", Status: "completed"},
	}

	for _, r := range reports {
		require.NoError(t, reporter.ReportProgress(r))
	}

	assert.Len(t, reporter.reports, 3)
	assert.Equal(t, "init", reporter.reports[0].Phase)
	assert.Equal(t, "completed", reporter.reports[2].Status)
}

func TestToolCall_MarshalRoundTrip(t *testing.T) {
	t.Parallel()

	call := ToolCall{
		Name:   "report_progress",
		Params: json.RawMessage(`{"task_id":"t1","progress":0.5}`),
	}

	data, err := json.Marshal(call)
	require.NoError(t, err)

	var decoded ToolCall
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "report_progress", decoded.Name)
	assert.JSONEq(t, `{"task_id":"t1","progress":0.5}`, string(decoded.Params))
}

func TestToolResult_Success(t *testing.T) {
	t.Parallel()

	result := ToolResult{
		Success: true,
		Data:    json.RawMessage(`{"reported":true}`),
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded ToolResult
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.True(t, decoded.Success)
	assert.Empty(t, decoded.Error)
}

func TestToolResult_Error(t *testing.T) {
	t.Parallel()

	result := ToolResult{
		Success: false,
		Error:   "task not found",
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded ToolResult
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.False(t, decoded.Success)
	assert.Equal(t, "task not found", decoded.Error)
}
