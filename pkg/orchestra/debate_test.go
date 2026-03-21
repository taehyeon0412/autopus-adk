package orchestra

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildRebuttalPrompt(t *testing.T) {
	t.Parallel()
	others := []ProviderResponse{
		{Provider: "alice", Output: "I think X"},
		{Provider: "bob", Output: "I think Y"},
	}
	prompt := buildRebuttalPrompt("original topic", others)
	assert.Contains(t, prompt, "original topic")
	assert.Contains(t, prompt, "alice")
	assert.Contains(t, prompt, "I think X")
	assert.Contains(t, prompt, "bob")
	assert.Contains(t, prompt, "rebuttal")
}

func TestBuildJudgmentPrompt(t *testing.T) {
	t.Parallel()
	args := []ProviderResponse{
		{Provider: "alice", Output: "Argument A"},
		{Provider: "bob", Output: "Argument B"},
	}
	prompt := buildJudgmentPrompt("test topic", args)
	assert.Contains(t, prompt, "test topic")
	assert.Contains(t, prompt, "alice")
	assert.Contains(t, prompt, "Argument A")
	assert.Contains(t, prompt, "judge")
}

func TestFindOrBuildJudgeConfig_Found(t *testing.T) {
	t.Parallel()
	cfg := OrchestraConfig{
		JudgeProvider: "myJudge",
		Providers: []ProviderConfig{
			{Name: "p1", Binary: "cat"},
			{Name: "myJudge", Binary: "/usr/bin/judge"},
		},
	}
	jc := findOrBuildJudgeConfig(cfg)
	assert.Equal(t, "/usr/bin/judge", jc.Binary)
}

func TestFindOrBuildJudgeConfig_NotFound(t *testing.T) {
	t.Parallel()
	cfg := OrchestraConfig{
		JudgeProvider: "judge",
		Providers: []ProviderConfig{
			{Name: "p1", Binary: "cat"},
		},
	}
	jc := findOrBuildJudgeConfig(cfg)
	assert.Equal(t, "judge", jc.Name)
	assert.Equal(t, "judge", jc.Binary)
}

func TestBuildDebateMerged_NoJudge(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		{Provider: "alice", Output: "Alice output"},
		{Provider: "bob", Output: "Bob output"},
	}
	cfg := OrchestraConfig{JudgeProvider: ""}
	merged, summary := buildDebateMerged(responses, cfg)
	assert.NotEmpty(t, merged)
	assert.Contains(t, summary, "판정")
	assert.Contains(t, summary, "없음")
}

func TestBuildDebateMerged_WithJudge(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		{Provider: "alice", Output: "Alice output"},
		{Provider: "bob", Output: "Bob output"},
		{Provider: "judge (judge)", Output: "Final verdict: Alice wins"},
	}
	cfg := OrchestraConfig{JudgeProvider: "judge"}
	merged, summary := buildDebateMerged(responses, cfg)
	assert.NotEmpty(t, merged)
	assert.Contains(t, summary, "judge")
	// Judge verdict should be reflected
	assert.True(t, strings.Contains(summary, "verdict") || strings.Contains(summary, "판정"))
}

func TestBuildDebateMerged_Empty(t *testing.T) {
	t.Parallel()
	_, summary := buildDebateMerged(nil, OrchestraConfig{})
	assert.Contains(t, summary, "결과 없음")
}
