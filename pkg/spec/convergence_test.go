package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// REQ-007: Circuit Breaker — halt when open+regressed count doesn't decrease

func TestCircuitBreaker_HaltsWhenOpenCountDoesNotDecrease(t *testing.T) {
	t.Parallel()

	// Given: two consecutive rounds with same open+regressed count
	prev := []ReviewFinding{
		{ID: "F-001", Status: FindingStatusOpen, Category: FindingCategoryCorrectness},
		{ID: "F-002", Status: FindingStatusRegressed, Category: FindingCategoryCompleteness},
	}
	curr := []ReviewFinding{
		{ID: "F-001", Status: FindingStatusOpen, Category: FindingCategoryCorrectness},
		{ID: "F-002", Status: FindingStatusRegressed, Category: FindingCategoryCompleteness},
	}

	// When: circuit breaker is evaluated
	tripped := ShouldTripCircuitBreaker(prev, curr)

	// Then: breaker trips
	assert.True(t, tripped, "circuit breaker must trip when open+regressed count does not decrease")
}

func TestCircuitBreaker_ContinuesWhenCountDecreases(t *testing.T) {
	t.Parallel()

	prev := []ReviewFinding{
		{ID: "F-001", Status: FindingStatusOpen, Category: FindingCategoryCorrectness},
		{ID: "F-002", Status: FindingStatusOpen, Category: FindingCategoryFeasibility},
	}
	curr := []ReviewFinding{
		{ID: "F-001", Status: FindingStatusResolved, Category: FindingCategoryCorrectness},
		{ID: "F-002", Status: FindingStatusOpen, Category: FindingCategoryFeasibility},
	}

	tripped := ShouldTripCircuitBreaker(prev, curr)

	assert.False(t, tripped, "circuit breaker must not trip when open+regressed count decreases")
}

func TestCircuitBreaker_EscapeHatchExcluded(t *testing.T) {
	t.Parallel()

	// Given: one critical/security finding marked as escape hatch, count otherwise same
	prev := []ReviewFinding{
		{ID: "F-001", Status: FindingStatusOpen, Category: FindingCategoryCorrectness},
	}
	curr := []ReviewFinding{
		{ID: "F-001", Status: FindingStatusOpen, Category: FindingCategoryCorrectness},
		{ID: "F-002", Status: FindingStatusOpen, Category: FindingCategorySecurity, EscapeHatch: true},
	}

	// Escape hatch findings are excluded from circuit breaker count
	tripped := ShouldTripCircuitBreaker(prev, curr)

	assert.False(t, tripped, "escape hatch findings must not count toward circuit breaker threshold")
}

// REQ-006: Scope Lock — out_of_scope for new non-critical findings in verify mode

func TestScopeFilter_TagsNewNonCriticalAsOutOfScope_InVerifyMode(t *testing.T) {
	t.Parallel()

	prior := []ReviewFinding{
		{ID: "F-001", Status: FindingStatusOpen, Category: FindingCategoryCorrectness},
	}
	incoming := []ReviewFinding{
		{ID: "F-002", Status: FindingStatusOpen, Category: FindingCategoryStyle, Description: "Naming could be better"},
	}

	// When: scope filter applied in verify mode
	filtered := ApplyScopeLock(incoming, prior, ReviewModeVerify)

	// Then: new non-critical finding is tagged out_of_scope
	require.Len(t, filtered, 1)
	assert.Equal(t, FindingStatusOutOfScope, filtered[0].Status)
}

func TestScopeFilter_AllowsSecurityEscapeHatch_InVerifyMode(t *testing.T) {
	t.Parallel()

	prior := []ReviewFinding{}
	incoming := []ReviewFinding{
		{ID: "F-003", Status: FindingStatusOpen, Category: FindingCategorySecurity, Description: "SQL injection risk"},
	}

	filtered := ApplyScopeLock(incoming, prior, ReviewModeVerify)

	require.Len(t, filtered, 1)
	assert.Equal(t, FindingStatusOpen, filtered[0].Status, "security finding must not be locked out")
	assert.True(t, filtered[0].EscapeHatch, "security finding in verify mode must be marked as escape hatch")
}

func TestScopeFilter_DiscoverModeAllowsAllFindings(t *testing.T) {
	t.Parallel()

	prior := []ReviewFinding{}
	incoming := []ReviewFinding{
		{ID: "F-004", Status: FindingStatusOpen, Category: FindingCategoryStyle},
		{ID: "F-005", Status: FindingStatusOpen, Category: FindingCategoryCompleteness},
	}

	filtered := ApplyScopeLock(incoming, prior, ReviewModeDiscover)

	require.Len(t, filtered, 2)
	for _, f := range filtered {
		assert.NotEqual(t, FindingStatusOutOfScope, f.Status)
	}
}

// REQ-011: Multi-provider supermajority merge (2/3 threshold)

func TestSupermajorityMerge_ReachesThreshold(t *testing.T) {
	t.Parallel()

	findings := []ReviewFinding{
		{ID: "F-001", Provider: "claude", Category: FindingCategoryCorrectness, Description: "Missing validation", ScopeRef: "REQ-001"},
		{ID: "F-002", Provider: "gemini", Category: FindingCategoryCorrectness, Description: "Missing validation", ScopeRef: "REQ-001"},
		{ID: "F-003", Provider: "openai", Category: FindingCategoryStyle, Description: "Naming issue", ScopeRef: "types.go:10"},
	}

	// When: 3 providers, 2 agree on same finding (2/3 supermajority)
	merged := MergeSupermajority(findings, 3, 0.67)

	// Then: merged finding present
	var correctnessFindings []ReviewFinding
	for _, f := range merged {
		if f.Category == FindingCategoryCorrectness {
			correctnessFindings = append(correctnessFindings, f)
		}
	}
	assert.NotEmpty(t, correctnessFindings, "supermajority finding must survive merge")
}

func TestSupermajorityMerge_DropsBelowThreshold(t *testing.T) {
	t.Parallel()

	findings := []ReviewFinding{
		{ID: "F-001", Provider: "claude", Category: FindingCategoryStyle, Description: "Naming issue", ScopeRef: "types.go:10"},
	}

	// Only 1 out of 3 providers agrees → below 2/3
	merged := MergeSupermajority(findings, 3, 0.67)

	var styleFindings []ReviewFinding
	for _, f := range merged {
		if f.Category == FindingCategoryStyle {
			styleFindings = append(styleFindings, f)
		}
	}
	assert.Empty(t, styleFindings, "finding below supermajority threshold must be dropped")
}

func TestSupermajorityMerge_CriticalAlwaysKept(t *testing.T) {
	t.Parallel()

	// Critical findings bypass the supermajority filter
	findings := []ReviewFinding{
		{ID: "F-001", Provider: "claude", Category: FindingCategorySecurity, Description: "Critical auth bypass", ScopeRef: "auth.go:42"},
	}

	merged := MergeSupermajority(findings, 3, 0.67)

	var securityFindings []ReviewFinding
	for _, f := range merged {
		if f.Category == FindingCategorySecurity {
			securityFindings = append(securityFindings, f)
		}
	}
	assert.NotEmpty(t, securityFindings, "critical/security findings must always survive merge regardless of supermajority")
}

// REQ-012: ScopeRef normalization

func TestNormalizeScopeRef_FilePathWithLineNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "absolute path normalized to relative",
			input:    "/Users/user/project/pkg/spec/types.go:42",
			expected: "pkg/spec/types.go:42",
		},
		{
			name:     "relative path unchanged",
			input:    "pkg/spec/types.go:42",
			expected: "pkg/spec/types.go:42",
		},
		{
			name:     "requirement ref unchanged",
			input:    "REQ-001",
			expected: "REQ-001",
		},
		{
			name:     "empty string stays empty",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NormalizeScopeRef(tt.input, "/Users/user/project")
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestNormalizeScopeRef_LineNumberPreserved(t *testing.T) {
	t.Parallel()

	got := NormalizeScopeRef("internal/cli/run.go:100", "/project")
	assert.Contains(t, got, ":100", "line number must be preserved after normalization")
}
