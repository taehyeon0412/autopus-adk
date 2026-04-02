package adapter

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeAdapterName(t *testing.T) {
	a := NewClaudeAdapter()
	assert.Equal(t, "claude", a.Name())
}

func TestClaudeAdapterBuildCommand(t *testing.T) {
	a := NewClaudeAdapter()
	task := TaskConfig{
		TaskID:    "task-123",
		Prompt:    "do something",
		MCPConfig: "/tmp/worker-mcp.json",
		WorkDir:   "/tmp/work",
		EnvVars:   map[string]string{"EXTRA": "val"},
	}

	cmd := a.BuildCommand(context.Background(), task)

	assert.Equal(t, "claude", cmd.Path[len(cmd.Path)-len("claude"):])
	assert.Contains(t, cmd.Args, "--print")
	assert.Contains(t, cmd.Args, "--output-format")
	assert.Contains(t, cmd.Args, "stream-json")
	assert.Contains(t, cmd.Args, "--verbose")
	assert.Contains(t, cmd.Args, "--resume")
	assert.Contains(t, cmd.Args, "worker-sess-task-123")
	assert.Contains(t, cmd.Args, "--mcp-config")
	assert.Contains(t, cmd.Args, "/tmp/worker-mcp.json")
	assert.Equal(t, "/tmp/work", cmd.Dir)

	envContains(t, cmd.Env, "AUTOPUS_TASK_ID=task-123")
	envContains(t, cmd.Env, "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1")
	envContains(t, cmd.Env, "EXTRA=val")
}

func TestClaudeAdapterBuildCommandWithSessionID(t *testing.T) {
	a := NewClaudeAdapter()
	task := TaskConfig{
		TaskID:    "task-456",
		SessionID: "custom-session",
	}

	cmd := a.BuildCommand(context.Background(), task)
	assert.Contains(t, cmd.Args, "custom-session")
}

func TestClaudeAdapterBuildCommandNoMCPConfig(t *testing.T) {
	a := NewClaudeAdapter()
	task := TaskConfig{TaskID: "task-789"}

	cmd := a.BuildCommand(context.Background(), task)
	assert.False(t, slices.Contains(cmd.Args, "--mcp-config"))
}

func TestClaudeAdapterParseEvent(t *testing.T) {
	a := NewClaudeAdapter()

	line := []byte(`{"type":"system.init","mcp_servers":["fs"]}`)
	evt, err := a.ParseEvent(line)
	require.NoError(t, err)
	assert.Equal(t, "system", evt.Type)
	assert.Equal(t, "init", evt.Subtype)
	assert.NotEmpty(t, evt.Data)
}

func TestClaudeAdapterParseEventInvalid(t *testing.T) {
	a := NewClaudeAdapter()
	_, err := a.ParseEvent([]byte("not json"))
	require.Error(t, err)
}

func TestClaudeAdapterExtractResult(t *testing.T) {
	a := NewClaudeAdapter()

	evt := StreamEvent{
		Type:    "result",
		Subtype: "",
		Data:    []byte(`{"cost_usd":0.05,"duration_ms":1200,"session_id":"sess-1","output":"done"}`),
	}

	result := a.ExtractResult(evt)
	assert.InDelta(t, 0.05, result.CostUSD, 0.001)
	assert.Equal(t, int64(1200), result.DurationMS)
	assert.Equal(t, "sess-1", result.SessionID)
	assert.Equal(t, "done", result.Output)
}

func TestClaudeAdapterExtractResultInvalidJSON(t *testing.T) {
	a := NewClaudeAdapter()

	evt := StreamEvent{Data: []byte("bad json")}
	result := a.ExtractResult(evt)
	assert.Equal(t, "bad json", result.Output)
}

// envContains checks that the env slice contains the expected key=value pair.
func envContains(t *testing.T, env []string, expected string) {
	t.Helper()
	if !slices.Contains(env, expected) {
		t.Errorf("env does not contain %q", expected)
	}
}
