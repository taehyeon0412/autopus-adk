// Package content_test는 인텐트 게이트 패키지의 테스트이다.
package content_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insajin/autopus-adk/pkg/content"
)

func TestDefaultRules(t *testing.T) {
	t.Parallel()

	rules := content.DefaultRules()
	assert.NotEmpty(t, rules)

	// 각 규칙은 Pattern, Priority를 가져야 함
	for _, r := range rules {
		assert.NotEmpty(t, r.Pattern)
		assert.Greater(t, r.Priority, 0)
	}
}

func TestGenerateIntentGateInstruction(t *testing.T) {
	t.Parallel()

	rules := []content.IntentRule{
		{Pattern: "plan.*feature", TargetSkill: "planning", Priority: 10},
		{Pattern: "debug.*error", TargetAgent: "debugger", Priority: 20},
		{Pattern: "test.*", TargetSkill: "tdd", Priority: 15},
	}

	result := content.GenerateIntentGateInstruction(rules)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "planning")
	assert.Contains(t, result, "debugger")
	assert.Contains(t, result, "tdd")
}

func TestIntentRule_TargetSkillOrAgent(t *testing.T) {
	t.Parallel()

	// 스킬 또는 에이전트 중 하나만 설정 가능
	skillRule := content.IntentRule{
		Pattern:     "plan",
		TargetSkill: "planning",
		Priority:    10,
	}
	assert.NotEmpty(t, skillRule.TargetSkill)
	assert.Empty(t, skillRule.TargetAgent)

	agentRule := content.IntentRule{
		Pattern:     "debug",
		TargetAgent: "debugger",
		Priority:    20,
	}
	assert.Empty(t, agentRule.TargetSkill)
	assert.NotEmpty(t, agentRule.TargetAgent)
}
