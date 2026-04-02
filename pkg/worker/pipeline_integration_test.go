package worker

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/insajin/autopus-adk/pkg/worker/adapter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pipelineMockAdapter implements ProviderAdapter using a shell script.
type pipelineMockAdapter struct {
	script string
}

func (m *pipelineMockAdapter) Name() string { return "pipeline-mock" }

func (m *pipelineMockAdapter) BuildCommand(ctx context.Context, task adapter.TaskConfig) *exec.Cmd {
	return exec.CommandContext(ctx, "sh", "-c", m.script)
}

func (m *pipelineMockAdapter) ParseEvent(line []byte) (adapter.StreamEvent, error) {
	var raw struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &raw); err != nil {
		return adapter.StreamEvent{}, err
	}
	return adapter.StreamEvent{
		Type: raw.Type,
		Data: json.RawMessage(append([]byte(nil), line...)),
	}, nil
}

func (m *pipelineMockAdapter) ExtractResult(event adapter.StreamEvent) adapter.TaskResult {
	var data struct {
		Output     string  `json:"output"`
		CostUSD    float64 `json:"cost_usd"`
		DurationMS int64   `json:"duration_ms"`
		SessionID  string  `json:"session_id"`
	}
	json.Unmarshal(event.Data, &data)
	return adapter.TaskResult{
		Output:     data.Output,
		CostUSD:    data.CostUSD,
		DurationMS: data.DurationMS,
		SessionID:  data.SessionID,
	}
}

func TestPipelineExecute_HappyPath(t *testing.T) {
	script := `cat /dev/stdin > /dev/null; echo '{"type":"result","output":"phase ok","cost_usd":0.01,"duration_ms":100}'`
	mock := &pipelineMockAdapter{script: script}

	pe := NewPipelineExecutor(mock, "", t.TempDir())
	result, err := pe.Execute(context.Background(), "pipe-1", "initial prompt")

	require.NoError(t, err)
	assert.InDelta(t, 0.04, result.CostUSD, 0.001)
	assert.Equal(t, int64(400), result.DurationMS)
	for _, phase := range []string{"planner", "executor", "tester", "reviewer"} {
		assert.True(t, strings.Contains(result.Output, phase),
			"output should contain phase %q", phase)
	}
}

func TestPipelineExecute_PhaseFailure(t *testing.T) {
	script := `cat /dev/stdin > /dev/null; exit 1`
	mock := &pipelineMockAdapter{script: script}

	pe := NewPipelineExecutor(mock, "", t.TempDir())
	_, err := pe.Execute(context.Background(), "pipe-fail", "prompt")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "phase planner")
}

func TestPipelineExecute_ContextCancel(t *testing.T) {
	script := `cat /dev/stdin > /dev/null; sleep 30`
	mock := &pipelineMockAdapter{script: script}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pe := NewPipelineExecutor(mock, "", t.TempDir())
	_, err := pe.Execute(ctx, "pipe-cancel", "prompt")

	require.Error(t, err)
}

func TestPipelineRunPhase_NoResultEvent(t *testing.T) {
	script := `cat /dev/stdin > /dev/null; echo '{"type":"system.init"}'`
	mock := &pipelineMockAdapter{script: script}

	pe := &PipelineExecutor{provider: mock, workDir: t.TempDir()}
	_, err := pe.runPhase(context.Background(), "task-x", PhasePlanner, "prompt")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no result event")
}

func TestPipelineRunPhase_FailWithOutput(t *testing.T) {
	// Exits non-zero but has a result event.
	script := `cat /dev/stdin > /dev/null; echo '{"type":"result","output":"partial"}'; exit 1`
	mock := &pipelineMockAdapter{script: script}

	pe := &PipelineExecutor{provider: mock, workDir: t.TempDir()}
	result, err := pe.runPhase(context.Background(), "task-y", PhaseExecutor, "prompt")

	require.NoError(t, err)
	assert.Equal(t, "partial", result.Output)
	assert.Equal(t, PhaseExecutor, result.Phase)
}
