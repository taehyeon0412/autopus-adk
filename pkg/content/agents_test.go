// Package content_test는 에이전트 콘텐츠 패키지의 테스트이다.
package content_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/content"
)

func TestLoadAgents(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeAgentFile(t, dir, "planner.md", `---
name: planner
role: 기획 전문 에이전트
model_tier: opus
category: planning
triggers:
  - "plan"
  - "기획"
skills:
  - planning
  - brainstorming
---

# Planner Agent

기획을 담당하는 에이전트입니다.
`)

	agents, err := content.LoadAgents(dir)
	require.NoError(t, err)
	require.Len(t, agents, 1)
	assert.Equal(t, "planner", agents[0].Name)
	assert.Equal(t, "기획 전문 에이전트", agents[0].Role)
	assert.Equal(t, "opus", agents[0].ModelTier)
	assert.Equal(t, "planning", agents[0].Category)
	assert.Equal(t, []string{"planning", "brainstorming"}, agents[0].Skills)
}

func TestLoadAgents_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	agents, err := content.LoadAgents(dir)
	require.NoError(t, err)
	assert.Len(t, agents, 0)
}

func TestConvertAgentToPlatform_Claude(t *testing.T) {
	t.Parallel()

	agent := content.AgentDefinition{
		Name:      "planner",
		Role:      "기획 전문 에이전트",
		ModelTier: "opus",
		Category:  "planning",
		Triggers:  []string{"plan"},
		Skills:    []string{"planning"},
	}

	result, err := content.ConvertAgentToPlatform(agent, "claude")
	require.NoError(t, err)
	// claude: .claude/agents/autopus/<name>.md 포맷
	assert.Contains(t, result, "planner")
	assert.Contains(t, result, "기획 전문 에이전트")
}

func TestConvertAgentToPlatform_Codex(t *testing.T) {
	t.Parallel()

	agent := content.AgentDefinition{
		Name: "planner",
		Role: "기획 전문 에이전트",
	}

	result, err := content.ConvertAgentToPlatform(agent, "codex")
	require.NoError(t, err)
	// codex: AGENTS.md 섹션 포맷
	assert.Contains(t, result, "planner")
}

func TestConvertAgentToPlatform_Gemini(t *testing.T) {
	t.Parallel()

	agent := content.AgentDefinition{
		Name: "planner",
		Role: "기획 전문 에이전트",
	}

	result, err := content.ConvertAgentToPlatform(agent, "gemini")
	require.NoError(t, err)
	// gemini: .gemini/skills/auto-agent-<name>/SKILL.md
	assert.Contains(t, result, "auto-agent-planner")
}

func TestConvertAgentToPlatform_UnknownPlatform(t *testing.T) {
	t.Parallel()

	agent := content.AgentDefinition{Name: "test"}
	_, err := content.ConvertAgentToPlatform(agent, "unknown")
	assert.Error(t, err)
}

// writeAgentFile은 테스트용 에이전트 파일을 생성한다.
func writeAgentFile(t *testing.T, dir, name, body string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644)
	require.NoError(t, err)
}
