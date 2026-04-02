package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderApprovalDialog_ContainsAction(t *testing.T) {
	t.Parallel()

	req := ApprovalRequest{
		Action:    "delete-all-files",
		RiskLevel: "high",
	}
	out := renderApprovalDialog(req, 80)
	assert.Contains(t, out, "delete-all-files")
}

func TestRenderApprovalDialog_ContainsRiskBadge(t *testing.T) {
	t.Parallel()

	req := ApprovalRequest{
		Action:    "some-action",
		RiskLevel: "critical",
	}
	out := renderApprovalDialog(req, 80)
	assert.Contains(t, out, "CRITICAL")
}

func TestRenderApprovalDialog_ContainsKeys(t *testing.T) {
	t.Parallel()

	req := ApprovalRequest{
		Action:    "test-action",
		RiskLevel: "low",
	}
	out := renderApprovalDialog(req, 80)
	assert.Contains(t, out, "pprove")
	assert.Contains(t, out, "eny")
}

func TestRenderRiskBadge_AllLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level    string
		expected string
	}{
		{"low", "LOW"},
		{"medium", "MEDIUM"},
		{"high", "HIGH"},
		{"critical", "CRITICAL"},
		{"something-else", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			badge := renderRiskBadge(tt.level)
			assert.Contains(t, badge, tt.expected)
		})
	}
}

func TestWrapText_ShortText(t *testing.T) {
	t.Parallel()

	lines := wrapText("hello world", 80)
	assert.Len(t, lines, 1)
	assert.Equal(t, "hello world", lines[0])
}

func TestWrapText_LongText(t *testing.T) {
	t.Parallel()

	// Each word is 5 chars, with space separator = 6 per word.
	// Width 15 should fit ~2 words per line.
	lines := wrapText("alpha bravo charlie delta echo", 15)
	assert.Greater(t, len(lines), 1, "long text should wrap into multiple lines")
}

func TestWrapText_EmptyText(t *testing.T) {
	t.Parallel()

	lines := wrapText("", 80)
	assert.Nil(t, lines)
}

func TestWrapText_MinWidth(t *testing.T) {
	t.Parallel()

	// Width < 10 should be clamped to 10.
	lines := wrapText("short text here for testing", 3)
	assert.NotNil(t, lines)
	// Verify wrapping happened with effective width of 10.
	for _, line := range lines {
		assert.LessOrEqual(t, len(line), 15, "lines should respect min width wrapping")
	}
}
