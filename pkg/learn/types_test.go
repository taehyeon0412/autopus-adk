package learn

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEntryType_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    EntryType
		expected string
	}{
		{"gate_fail type", EntryTypeGateFail, "gate_fail"},
		{"coverage_gap type", EntryTypeCoverageGap, "coverage_gap"},
		{"review_issue type", EntryTypeReviewIssue, "review_issue"},
		{"executor_error type", EntryTypeExecutorError, "executor_error"},
		{"fix_pattern type", EntryTypeFixPattern, "fix_pattern"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.value))
		})
	}
}

func TestSeverity_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    Severity
		expected string
	}{
		{"low severity", SeverityLow, "low"},
		{"medium severity", SeverityMedium, "medium"},
		{"high severity", SeverityHigh, "high"},
		{"critical severity", SeverityCritical, "critical"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.value))
		})
	}
}

func TestLearningEntry_RequiredFields(t *testing.T) {
	t.Parallel()

	// Given: a fully populated LearningEntry
	entry := LearningEntry{
		ID:         "L-001",
		Timestamp:  time.Now(),
		Type:       EntryTypeGateFail,
		Phase:      "test",
		Files:      []string{"pkg/learn/store.go"},
		Packages:   []string{"learn"},
		Pattern:    "coverage below threshold",
		Resolution: "added missing tests",
		Severity:   SeverityMedium,
		ReuseCount: 0,
	}

	// Then: all required fields are set
	assert.Equal(t, "L-001", entry.ID)
	assert.Equal(t, EntryTypeGateFail, entry.Type)
	assert.Equal(t, "test", entry.Phase)
	assert.NotEmpty(t, entry.Files)
	assert.NotEmpty(t, entry.Packages)
	assert.NotEmpty(t, entry.Pattern)
	assert.NotEmpty(t, entry.Resolution)
	assert.Equal(t, SeverityMedium, entry.Severity)
	assert.Equal(t, 0, entry.ReuseCount)
	assert.False(t, entry.Timestamp.IsZero())
}

func TestLearningEntry_IDFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{"valid L-001", "L-001", true},
		{"valid L-100", "L-100", true},
		{"invalid no prefix", "001", false},
		{"invalid wrong prefix", "X-001", false},
		{"empty id", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsValidEntryID(tt.id)
			assert.Equal(t, tt.valid, got)
		})
	}
}

func TestLearningEntry_ZeroValue(t *testing.T) {
	t.Parallel()

	// Given: zero-value entry
	var entry LearningEntry

	// Then: fields should be zero values
	assert.Empty(t, entry.ID)
	assert.True(t, entry.Timestamp.IsZero())
	assert.Empty(t, string(entry.Type))
	assert.Empty(t, entry.Phase)
	assert.Nil(t, entry.Files)
	assert.Nil(t, entry.Packages)
	assert.Empty(t, entry.Pattern)
	assert.Empty(t, entry.Resolution)
	assert.Empty(t, string(entry.Severity))
	assert.Equal(t, 0, entry.ReuseCount)
}
