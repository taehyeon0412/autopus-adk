package adapter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeminiAdapterName(t *testing.T) {
	a := NewGeminiAdapter()
	assert.Equal(t, "gemini", a.Name())
}

func TestGeminiAdapterBuildCommand(t *testing.T) {
	a := NewGeminiAdapter()
	task := TaskConfig{
		TaskID:  "task-g1",
		Prompt:  "analyze code",
		WorkDir: "/tmp/gemini-work",
		EnvVars: map[string]string{"KEY": "val"},
	}

	cmd := a.BuildCommand(context.Background(), task)

	assert.Contains(t, cmd.Args, "--output-format")
	assert.Contains(t, cmd.Args, "stream-json")
	assert.Contains(t, cmd.Args, "--resume")
	assert.Contains(t, cmd.Args, "worker-sess-task-g1")
	assert.Contains(t, cmd.Args, "-p")
	assert.Contains(t, cmd.Args, "analyze code")
	assert.Equal(t, "/tmp/gemini-work", cmd.Dir)

	envContains(t, cmd.Env, "AUTOPUS_TASK_ID=task-g1")
	envContains(t, cmd.Env, "KEY=val")
}

func TestGeminiAdapterBuildCommandWithSession(t *testing.T) {
	a := NewGeminiAdapter()
	task := TaskConfig{
		TaskID:    "task-g2",
		SessionID: "gem-sess",
	}

	cmd := a.BuildCommand(context.Background(), task)
	assert.Contains(t, cmd.Args, "gem-sess")
}

func TestGeminiAdapterParseEvent(t *testing.T) {
	a := NewGeminiAdapter()

	line := []byte(`{"type":"result","output":"ok"}`)
	evt, err := a.ParseEvent(line)
	require.NoError(t, err)
	assert.Equal(t, "result", evt.Type)
}

func TestGeminiAdapterParseEventInvalid(t *testing.T) {
	a := NewGeminiAdapter()
	_, err := a.ParseEvent([]byte("bad"))
	require.Error(t, err)
}

func TestGeminiAdapterExtractResult(t *testing.T) {
	a := NewGeminiAdapter()

	evt := StreamEvent{
		Type: "result",
		Data: []byte(`{"type":"result","output":"finished","cost_usd":0.03,"duration_ms":900,"session_id":"gs1"}`),
	}

	result := a.ExtractResult(evt)
	assert.InDelta(t, 0.03, result.CostUSD, 0.001)
	assert.Equal(t, int64(900), result.DurationMS)
	assert.Equal(t, "gs1", result.SessionID)
	assert.Equal(t, "finished", result.Output)
}

func TestGeminiAdapterExtractResultInvalidJSON(t *testing.T) {
	a := NewGeminiAdapter()

	evt := StreamEvent{Data: []byte("nope")}
	result := a.ExtractResult(evt)
	assert.Equal(t, "nope", result.Output)
}
