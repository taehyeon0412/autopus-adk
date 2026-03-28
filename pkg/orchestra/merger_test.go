package orchestra

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func makeResp(provider, output string) ProviderResponse {
	return ProviderResponse{
		Provider: provider,
		Output:   output,
		Duration: 50 * time.Millisecond,
	}
}

func TestNormalizeLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello world  ", "hello world"},
		{"Hello, World!", "hello world"},
		{"Go is fast.", "go is fast"},
		{"", ""},
		{"   ", ""},
		{"UPPER CASE TEXT", "upper case text"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			result := normalizeLine(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSplitLines(t *testing.T) {
	t.Parallel()
	text := "line1\nline2\n\nline3\n  \nline4"
	lines := splitLines(text)
	assert.Equal(t, []string{"line1", "line2", "line3", "line4"}, lines)
}

func TestSplitLines_Empty(t *testing.T) {
	t.Parallel()
	assert.Empty(t, splitLines(""))
	assert.Empty(t, splitLines("\n\n\n"))
}

func TestMergeConsensus_AllAgree(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		makeResp("p1", "go is great\nfast language"),
		makeResp("p2", "go is great\nfast language"),
		makeResp("p3", "go is great\nfast language"),
	}

	merged, summary := MergeConsensus(responses, 0.66)
	assert.Contains(t, merged, "✓")
	assert.Contains(t, summary, "합의율")
	// 100% 합의이므로 합의 섹션에 항목이 있어야 한다
	assert.Contains(t, merged, "합의된 내용")
}

func TestMergeConsensus_NoneAgree(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		makeResp("p1", "unique line from p1"),
		makeResp("p2", "unique line from p2"),
		makeResp("p3", "unique line from p3"),
	}

	merged, summary := MergeConsensus(responses, 0.66)
	// 이견 항목이 있어야 한다
	assert.Contains(t, merged, "이견")
	assert.Contains(t, summary, "합의율")
	_ = summary
}

func TestMergeConsensus_Empty(t *testing.T) {
	t.Parallel()
	merged, summary := MergeConsensus(nil, 0.66)
	assert.Empty(t, merged)
	assert.Equal(t, "응답 없음", summary)
}

func TestMergeConsensus_PartialAgreement(t *testing.T) {
	t.Parallel()
	// p1, p2는 동의하지만 p3는 다른 내용
	responses := []ProviderResponse{
		makeResp("p1", "shared content\np1 only"),
		makeResp("p2", "shared content\np2 only"),
		makeResp("p3", "shared content\np3 only"),
	}

	merged, _ := MergeConsensus(responses, 0.66)
	// "shared content"는 3/3 = 100% 합의
	assert.Contains(t, merged, "shared content")
}

func TestFormatPipeline(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		makeResp("designer", "design output"),
		makeResp("coder", "code output"),
	}

	result := FormatPipeline(responses)
	assert.Contains(t, result, "Stage 1")
	assert.Contains(t, result, "designer")
	assert.Contains(t, result, "design output")
	assert.Contains(t, result, "Stage 2")
	assert.Contains(t, result, "coder")
}

func TestFormatPipeline_Empty(t *testing.T) {
	t.Parallel()
	result := FormatPipeline(nil)
	assert.Empty(t, result)
}

func TestFormatDebate(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		makeResp("claude", "use REST\nJSON format"),
		makeResp("codex", "use GraphQL\nJSON format"),
	}

	result := FormatDebate(responses)
	assert.Contains(t, result, "claude의 의견")
	assert.Contains(t, result, "codex의 의견")
	assert.Contains(t, result, "use REST")
	assert.Contains(t, result, "use GraphQL")
	assert.Contains(t, result, "주요 차이점")
}

func TestFormatDebate_Empty(t *testing.T) {
	t.Parallel()
	result := FormatDebate(nil)
	assert.Empty(t, result)
}

func TestFormatDebate_SingleResponse(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		makeResp("solo", "solo response"),
	}
	result := FormatDebate(responses)
	assert.Contains(t, result, "solo의 의견")
	// 단일 응답은 차이점 섹션이 없다
	assert.False(t, strings.Contains(result, "주요 차이점"))
}

func TestMax1(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 1, max1(0))
	assert.Equal(t, 1, max1(-5))
	assert.Equal(t, 5, max1(5))
}

func TestFindDifferences(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		makeResp("p1", "common line\np1 unique"),
		makeResp("p2", "common line\np2 unique"),
	}
	diffs := findDifferences(responses)
	assert.NotEmpty(t, diffs)
	// p2 unique는 p1에 없으므로 차이점으로 나타나야 한다
	found := false
	for _, d := range diffs {
		if strings.Contains(d, "p2") {
			found = true
		}
	}
	assert.True(t, found)
}

func TestFindDifferences_SingleResponse(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{makeResp("p1", "content")}
	diffs := findDifferences(responses)
	assert.Nil(t, diffs)
}

// --- SPEC-ORCH-013 R3: Diff Noise Refinement ---

// TestFormatDebate_CleanOutput_MCPNoiseRemoved verifies MCP noise is removed
// from debate diff output before comparing.
// S5: Diff section must not contain MCP noise fragments.
func TestFormatDebate_CleanOutput_MCPNoiseRemoved(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		makeResp("claude", "use REST API\nMCP issues detected. Run /mcp list for status.JSON format"),
		makeResp("gemini", "use GraphQL\nJSON format"),
	}
	result := FormatDebate(responses)
	assert.NotContains(t, result, "MCP issues detected",
		"FormatDebate must clean MCP noise from responses before comparing")
	assert.NotContains(t, result, "/mcp list",
		"FormatDebate must clean MCP noise fragments from diff output")
}

// TestFindDifferences_CleanOutput_MCPNoiseExcluded verifies findDifferences
// uses cleanScreenOutput on responses so MCP noise doesn't appear as diffs.
// S5: MCP noise lines should not be reported as differences.
func TestFindDifferences_CleanOutput_MCPNoiseExcluded(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		makeResp("p1", "shared line\nMCP issues detected. Run /mcp list for status."),
		makeResp("p2", "shared line"),
	}
	diffs := findDifferences(responses)
	for _, d := range diffs {
		assert.NotContains(t, d, "MCP issues detected",
			"findDifferences must not report MCP noise as a difference")
	}
}

// TestFormatDebate_CleanOutput_ANSIEscapesRemoved verifies ANSI escape sequences
// are stripped from debate output before comparing.
// S6: Diff section must not contain ANSI escape codes.
func TestFormatDebate_CleanOutput_ANSIEscapesRemoved(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		makeResp("claude", "\x1b[31muse REST\x1b[0m\nJSON format"),
		makeResp("gemini", "use GraphQL\nJSON format"),
	}
	result := FormatDebate(responses)
	assert.NotContains(t, result, "\x1b[",
		"FormatDebate must strip ANSI escape sequences before comparing")
}

// TestFindDifferences_CleanOutput_ANSIStripped verifies ANSI codes don't cause
// false differences between otherwise identical content.
// S6: ANSI-only differences should not appear in diff output.
func TestFindDifferences_CleanOutput_ANSIStripped(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		makeResp("p1", "\x1b[1mshared line\x1b[0m"),
		makeResp("p2", "shared line"),
	}
	diffs := findDifferences(responses)
	assert.Empty(t, diffs,
		"ANSI-only differences must not appear as diffs after cleaning")
}
