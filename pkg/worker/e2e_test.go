package worker

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/insajin/autopus-adk/pkg/worker/adapter"
	"github.com/insajin/autopus-adk/pkg/worker/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2EWorkflow exercises the full task lifecycle:
// config → A2A server → task receipt → prompt build → (mock) execution → result.
func TestE2EWorkflow(t *testing.T) {
	t.Parallel()

	// Step 1: Create worker config.
	cfg := LoopConfig{
		BackendURL: "ws://localhost:0",
		WorkerName: "e2e-worker",
		Skills:     []string{"code", "test"},
		Provider:   adapter.NewClaudeAdapter(),
		MCPConfig:  "",
		WorkDir:    t.TempDir(),
		AuthToken:  "test-token",
	}

	// Step 2: Create WorkerLoop (includes A2A server setup).
	wl := NewWorkerLoop(cfg)
	require.NotNil(t, wl, "NewWorkerLoop must succeed")
	require.NotNil(t, wl.server, "A2A server must be initialized")

	// Step 3: Simulate a task payload (like A2A backend would send).
	taskPayload := taskPayloadMessage{
		Description:   "Implement the frobulator service",
		PMNotes:       "Priority: high, deadline: next sprint",
		PolicySummary: "No network access allowed",
		KnowledgeCtx:  "Existing frobulator uses gRPC",
		SpecID:        "SPEC-FROB-001",
		SessionID:     "sess-e2e-001",
	}
	payloadJSON, err := json.Marshal(taskPayload)
	require.NoError(t, err)

	// Step 4: Verify prompt is built correctly (Layer 4 context).
	prompt := wl.builder.Build(TaskPayload{
		TaskID:        "task-e2e-1",
		Description:   taskPayload.Description,
		PMNotes:       taskPayload.PMNotes,
		PolicySummary: taskPayload.PolicySummary,
		KnowledgeCtx:  taskPayload.KnowledgeCtx,
		SpecID:        taskPayload.SpecID,
	})
	assert.Contains(t, prompt, "task-e2e-1")
	assert.Contains(t, prompt, "frobulator service")
	assert.Contains(t, prompt, "PM Notes")
	assert.Contains(t, prompt, "Security Policy")
	assert.Contains(t, prompt, "Knowledge Context")
	assert.Contains(t, prompt, "SPEC-FROB-001")

	// Step 5: Verify task payload can be parsed (as handleTask would).
	var parsed taskPayloadMessage
	require.NoError(t, json.Unmarshal(payloadJSON, &parsed))
	assert.Equal(t, taskPayload.Description, parsed.Description)
	assert.Equal(t, taskPayload.SessionID, parsed.SessionID)

	// Step 6: Verify TaskConfig is built correctly for subprocess.
	taskCfg := adapter.TaskConfig{
		TaskID:    "task-e2e-1",
		SessionID: parsed.SessionID,
		Prompt:    prompt,
		MCPConfig: cfg.MCPConfig,
		WorkDir:   cfg.WorkDir,
	}
	cmd := cfg.Provider.BuildCommand(context.Background(), taskCfg)
	require.NotNil(t, cmd, "BuildCommand must produce a command")
	assert.Equal(t, cfg.WorkDir, cmd.Dir)
}

// TestE2EContextBuilderSections verifies that optional sections are only
// included when their data is non-empty.
func TestE2EContextBuilderSections(t *testing.T) {
	t.Parallel()

	builder := ContextBuilder{}

	// Minimal payload — only required fields.
	minimal := builder.Build(TaskPayload{
		TaskID:      "t-min",
		Description: "do something",
	})
	assert.Contains(t, minimal, "# Task: t-min")
	assert.Contains(t, minimal, "do something")
	assert.NotContains(t, minimal, "PM Notes")
	assert.NotContains(t, minimal, "Security Policy")
	assert.NotContains(t, minimal, "Knowledge Context")
	assert.NotContains(t, minimal, "Reference")

	// Full payload — all sections present.
	full := builder.Build(TaskPayload{
		TaskID:        "t-full",
		Description:   "build it",
		PMNotes:       "urgent",
		PolicySummary: "no net",
		KnowledgeCtx:  "prior art",
		SpecID:        "SPEC-X",
	})
	for _, section := range []string{"PM Notes", "Security Policy", "Knowledge Context", "Reference"} {
		assert.Contains(t, full, section, "full payload should contain %s", section)
	}
}

// TestE2ESecurityPolicyIntegration verifies that security policies work
// end-to-end with the policy cache and command validation.
func TestE2ESecurityPolicyIntegration(t *testing.T) {
	t.Parallel()

	policy := security.SecurityPolicy{
		AllowedCommands: []string{"claude ", "codex "},
		AllowedDirs:     []string{t.TempDir()},
		TimeoutSec:      300,
	}

	// Policy cache round-trip.
	cache := security.NewPolicyCache()
	require.NoError(t, cache.Write("task-e2e-sec", policy))

	loaded, err := cache.Read("task-e2e-sec")
	require.NoError(t, err)
	assert.Equal(t, policy.AllowedCommands, loaded.AllowedCommands)
	assert.Equal(t, policy.TimeoutSec, loaded.TimeoutSec)

	// Command validation with loaded policy.
	ok, _ := loaded.ValidateCommand("claude --print", loaded.AllowedDirs[0])
	assert.True(t, ok, "claude command should be allowed in allowed dir")

	ok, reason := loaded.ValidateCommand("rm -rf /", loaded.AllowedDirs[0])
	assert.False(t, ok, "dangerous command should be denied: %s", reason)
}

// TestE2EArtifactConversion verifies adapter→A2A artifact conversion.
func TestE2EArtifactConversion(t *testing.T) {
	t.Parallel()

	src := []adapter.Artifact{
		{Name: "result.md", MimeType: "text/markdown", Data: "# Done"},
		{Name: "diff.patch", MimeType: "text/plain", Data: "+new line"},
	}

	got := convertArtifacts(src)
	require.Len(t, got, 2)
	assert.Equal(t, "result.md", got[0].Name)
	assert.Equal(t, "text/markdown", got[0].MimeType)
	assert.Equal(t, "# Done", got[0].Data)

	// Nil input returns nil.
	assert.Nil(t, convertArtifacts(nil))
}

// TestE2EStreamParsing verifies the stream parser handles mock events.
func TestE2EStreamParsing(t *testing.T) {
	t.Parallel()

	// Simulate stream-json lines as a subprocess would emit.
	// Fields are at the top level of the JSON object (matching Claude stream-json format).
	lines := []string{
		`{"type":"system.init"}`,
		`{"type":"system.task_started"}`,
		`{"type":"result","cost_usd":0.05,"duration_ms":1200,"session_id":"s1","output":"done"}`,
	}
	input := strings.Join(lines, "\n")

	wl := &WorkerLoop{
		config: LoopConfig{Provider: adapter.NewClaudeAdapter()},
	}

	result, err := wl.parseStream(strings.NewReader(input), "t-stream")
	require.NoError(t, err)
	assert.Equal(t, "s1", result.SessionID)
	assert.NotEmpty(t, result.Output)
}

// TestE2EPipelineAggregation verifies multi-phase result aggregation.
func TestE2EPipelineAggregation(t *testing.T) {
	t.Parallel()

	pe := NewPipelineExecutor(adapter.NewClaudeAdapter(), "", t.TempDir())
	phases := []PhaseResult{
		{Phase: PhasePlanner, Output: "plan", CostUSD: 0.01, DurationMS: 100},
		{Phase: PhaseExecutor, Output: "code", CostUSD: 0.10, DurationMS: 2000},
		{Phase: PhaseTester, Output: "tests pass", CostUSD: 0.03, DurationMS: 500},
		{Phase: PhaseReviewer, Output: "lgtm", CostUSD: 0.01, DurationMS: 200},
	}

	result := pe.aggregateResults(phases, 0.15, 2800)
	assert.InDelta(t, 0.15, result.CostUSD, 0.001)
	assert.Equal(t, int64(2800), result.DurationMS)
	for _, keyword := range []string{"plan", "code", "tests pass", "lgtm"} {
		assert.Contains(t, result.Output, keyword)
	}
}
