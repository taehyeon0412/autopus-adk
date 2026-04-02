package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTree_Nil(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", RenderTree(nil))
}

func TestRenderTree_SingleNode(t *testing.T) {
	t.Parallel()

	root := &SubagentNode{Name: "planner", Status: "completed"}
	result := RenderTree(root)

	assert.Contains(t, result, "planner")
}

func TestRenderTree_WithChildren(t *testing.T) {
	t.Parallel()

	root := &SubagentNode{
		Name:   "orchestrator",
		Status: "running",
		Children: []*SubagentNode{
			{Name: "executor-1", Status: "completed"},
			{Name: "executor-2", Status: "running"},
			{Name: "executor-3", Status: "failed"},
		},
	}

	result := RenderTree(root)
	assert.Contains(t, result, "orchestrator")
	assert.Contains(t, result, "executor-1")
	assert.Contains(t, result, "executor-2")
	assert.Contains(t, result, "executor-3")
	// Box-drawing: first two children use ├──, last uses └──
	assert.Contains(t, result, "├──")
	assert.Contains(t, result, "└──")
}

func TestRenderTree_NestedChildren(t *testing.T) {
	t.Parallel()

	root := &SubagentNode{
		Name:   "root",
		Status: "running",
		Children: []*SubagentNode{
			{
				Name:   "child-a",
				Status: "running",
				Children: []*SubagentNode{
					{Name: "grandchild", Status: "completed"},
				},
			},
		},
	}

	result := RenderTree(root)
	assert.Contains(t, result, "root")
	assert.Contains(t, result, "child-a")
	assert.Contains(t, result, "grandchild")
}

func TestParseStreamEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		eventType  string
		data       string
		wantStatus string
	}{
		{"start", "agent_start", "planner", "running"},
		{"complete", "agent_complete", "executor", "completed"},
		{"fail", "agent_fail", "tester", "failed"},
		{"unknown", "other", "misc", "running"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			node := ParseStreamEvent(tt.eventType, tt.data)
			assert.Equal(t, tt.data, node.Name)
			assert.Equal(t, tt.wantStatus, node.Status)
		})
	}
}
