package lore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/lore"
)

func TestParseTrailers_FullEntry(t *testing.T) {
	t.Parallel()

	commitMsg := `feat: 사용자 인증 구현

JWT 기반 인증을 구현했습니다.

Constraint: stateless 세션만 허용
Rejected: 세션 기반 인증 (서버 상태 유지 불가)
Confidence: high
Scope-risk: module
Reversibility: moderate
Directive: 항상 HTTPS 사용
Tested: JWT 토큰 생성, 검증
Not-tested: 토큰 만료 엣지 케이스
Related: SPEC-AUTH-001
`

	entry, err := lore.ParseTrailers(commitMsg)
	require.NoError(t, err)
	require.NotNil(t, entry)

	assert.Equal(t, "stateless 세션만 허용", entry.Constraint)
	assert.Equal(t, "세션 기반 인증 (서버 상태 유지 불가)", entry.Rejected)
	assert.Equal(t, "high", entry.Confidence)
	assert.Equal(t, "module", entry.ScopeRisk)
	assert.Equal(t, "moderate", entry.Reversibility)
	assert.Equal(t, "항상 HTTPS 사용", entry.Directive)
	assert.Equal(t, "JWT 토큰 생성, 검증", entry.Tested)
	assert.Equal(t, "토큰 만료 엣지 케이스", entry.NotTested)
	assert.Equal(t, "SPEC-AUTH-001", entry.Related)
}

func TestParseTrailers_PartialEntry(t *testing.T) {
	t.Parallel()

	commitMsg := `fix: 버그 수정

Constraint: 기존 API 호환성 유지
Confidence: medium
`

	entry, err := lore.ParseTrailers(commitMsg)
	require.NoError(t, err)
	require.NotNil(t, entry)

	assert.Equal(t, "기존 API 호환성 유지", entry.Constraint)
	assert.Equal(t, "medium", entry.Confidence)
	assert.Empty(t, entry.Rejected)
	assert.Empty(t, entry.Directive)
}

func TestParseTrailers_NoTrailers(t *testing.T) {
	t.Parallel()

	commitMsg := "feat: 새 기능 추가\n\n간단한 기능입니다.\n"

	entry, err := lore.ParseTrailers(commitMsg)
	require.NoError(t, err)
	require.NotNil(t, entry)

	assert.Empty(t, entry.Constraint)
	assert.Empty(t, entry.Confidence)
}

func TestParseTrailers_EmptyMessage(t *testing.T) {
	t.Parallel()

	entry, err := lore.ParseTrailers("")
	require.NoError(t, err)
	require.NotNil(t, entry)
}

func TestFormatTrailers_FullEntry(t *testing.T) {
	t.Parallel()

	entry := &lore.LoreEntry{
		Constraint:    "stateless 세션만 허용",
		Rejected:      "세션 기반 인증",
		Confidence:    "high",
		ScopeRisk:     "module",
		Reversibility: "moderate",
		Directive:     "항상 HTTPS 사용",
		Tested:        "JWT 생성, 검증",
		NotTested:     "만료 엣지 케이스",
		Related:       "SPEC-AUTH-001",
	}

	result := lore.FormatTrailers(entry)

	assert.Contains(t, result, "Constraint: stateless 세션만 허용")
	assert.Contains(t, result, "Rejected: 세션 기반 인증")
	assert.Contains(t, result, "Confidence: high")
	assert.Contains(t, result, "Scope-risk: module")
	assert.Contains(t, result, "Reversibility: moderate")
	assert.Contains(t, result, "Directive: 항상 HTTPS 사용")
	assert.Contains(t, result, "Tested: JWT 생성, 검증")
	assert.Contains(t, result, "Not-tested: 만료 엣지 케이스")
	assert.Contains(t, result, "Related: SPEC-AUTH-001")
}

func TestFormatTrailers_EmptyEntry(t *testing.T) {
	t.Parallel()

	entry := &lore.LoreEntry{}
	result := lore.FormatTrailers(entry)
	assert.Empty(t, result)
}

func TestFormatTrailers_PartialEntry(t *testing.T) {
	t.Parallel()

	entry := &lore.LoreEntry{
		Confidence: "low",
		Directive:  "주의",
	}

	result := lore.FormatTrailers(entry)
	assert.Contains(t, result, "Confidence: low")
	assert.Contains(t, result, "Directive: 주의")
	assert.NotContains(t, result, "Constraint:")
}

func TestRoundTrip_ParseFormat(t *testing.T) {
	t.Parallel()

	original := &lore.LoreEntry{
		Constraint:    "테스트 제약",
		Confidence:    "high",
		ScopeRisk:     "local",
		Reversibility: "trivial",
	}

	formatted := lore.FormatTrailers(original)
	commitMsg := "feat: 테스트\n\n" + formatted

	parsed, err := lore.ParseTrailers(commitMsg)
	require.NoError(t, err)

	assert.Equal(t, original.Constraint, parsed.Constraint)
	assert.Equal(t, original.Confidence, parsed.Confidence)
	assert.Equal(t, original.ScopeRisk, parsed.ScopeRisk)
	assert.Equal(t, original.Reversibility, parsed.Reversibility)
}
