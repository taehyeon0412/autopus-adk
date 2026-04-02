package worker

import (
	"strings"
	"testing"

	"github.com/insajin/autopus-adk/pkg/worker/adapter"
)

func TestPipelineExecutor_AggregateResults(t *testing.T) {
	pe := NewPipelineExecutor(adapter.NewClaudeAdapter(), "", "/tmp")

	results := []PhaseResult{
		{Phase: PhasePlanner, Output: "plan output", CostUSD: 0.01, DurationMS: 100},
		{Phase: PhaseExecutor, Output: "exec output", CostUSD: 0.05, DurationMS: 500},
		{Phase: PhaseTester, Output: "test output", CostUSD: 0.02, DurationMS: 200},
		{Phase: PhaseReviewer, Output: "review output", CostUSD: 0.01, DurationMS: 100},
	}

	tr := pe.aggregateResults(results, 0.09, 900)

	if tr.CostUSD != 0.09 {
		t.Errorf("CostUSD = %f, want 0.09", tr.CostUSD)
	}
	if tr.DurationMS != 900 {
		t.Errorf("DurationMS = %d, want 900", tr.DurationMS)
	}

	for _, phase := range []string{"planner", "executor", "tester", "reviewer"} {
		if !strings.Contains(tr.Output, phase) {
			t.Errorf("output missing phase %q", phase)
		}
	}
}

func TestPipelineExecutor_PhasePrompts(t *testing.T) {
	pe := NewPipelineExecutor(adapter.NewClaudeAdapter(), "", "/tmp")

	tests := []struct {
		name   string
		fn     func(string) string
		expect string
	}{
		{"planner", pe.plannerPrompt, "PLANNER"},
		{"executor", pe.executorPrompt, "EXECUTOR"},
		{"tester", pe.testerPrompt, "TESTER"},
		{"reviewer", pe.reviewerPrompt, "REVIEWER"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn("test input")
			if !strings.Contains(result, tt.expect) {
				t.Errorf("prompt missing %q role", tt.expect)
			}
			if !strings.Contains(result, "test input") {
				t.Error("prompt missing input content")
			}
		})
	}
}

func TestIsContextOverflow(t *testing.T) {
	tests := []struct {
		name string
		evt  adapter.StreamEvent
		want bool
	}{
		{"context window error", adapter.StreamEvent{Type: "error", Data: []byte(`{"error":"context window exceeded"}`)}, true},
		{"token limit error", adapter.StreamEvent{Type: "error", Data: []byte(`{"error":"Token limit reached"}`)}, true},
		{"other error", adapter.StreamEvent{Type: "error", Data: []byte(`{"error":"network timeout"}`)}, false},
		{"non-error event", adapter.StreamEvent{Type: "result", Data: []byte(`{"output":"ok"}`)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContextOverflow(tt.evt); got != tt.want {
				t.Errorf("IsContextOverflow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPipelineExecutor(t *testing.T) {
	pe := NewPipelineExecutor(adapter.NewClaudeAdapter(), "/tmp/mcp.json", "/work")
	if pe == nil {
		t.Fatal("expected non-nil PipelineExecutor")
	}
	if pe.mcpConfig != "/tmp/mcp.json" {
		t.Errorf("mcpConfig = %q, want %q", pe.mcpConfig, "/tmp/mcp.json")
	}
	if pe.workDir != "/work" {
		t.Errorf("workDir = %q, want %q", pe.workDir, "/work")
	}
}
