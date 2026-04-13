package lore_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/lore"
)

func TestBuildCommit_WithTrailers(t *testing.T) {
	t.Parallel()

	entry := &lore.LoreEntry{
		Constraint:    "stateless 세션만",
		Confidence:    "high",
		ScopeRisk:     "module",
		Reversibility: "moderate",
	}

	result, err := lore.BuildCommit(entry, "feat: 인증 구현")
	require.NoError(t, err)

	// 제목이 첫 줄
	lines := strings.Split(result, "\n")
	assert.Equal(t, "feat: 인증 구현", lines[0])

	// 트레일러 포함 확인
	assert.Contains(t, result, "Constraint: stateless 세션만")
	assert.Contains(t, result, "Confidence: high")
	assert.Contains(t, result, "Scope-risk: module")
	assert.Contains(t, result, "Reversibility: moderate")
	assert.Contains(t, result, "🐙 Autopus <noreply@autopus.co>")
}

func TestBuildCommit_EmptyEntry(t *testing.T) {
	t.Parallel()

	entry := &lore.LoreEntry{}
	result, err := lore.BuildCommit(entry, "chore: 설정 업데이트")
	require.NoError(t, err)

	lines := strings.Split(result, "\n")
	assert.Equal(t, "chore: 설정 업데이트", lines[0])
	// 트레일러가 없어도 Lore sign-off는 유지되어야 한다.
	assert.NotContains(t, result, "Constraint:")
	assert.Contains(t, result, "🐙 Autopus <noreply@autopus.co>")
}

func TestBuildCommit_EmptyMessage(t *testing.T) {
	t.Parallel()

	entry := &lore.LoreEntry{Confidence: "low"}
	_, err := lore.BuildCommit(entry, "")
	assert.Error(t, err)
}

func TestBuildCommit_MultilineMessage(t *testing.T) {
	t.Parallel()

	entry := &lore.LoreEntry{
		Directive: "항상 검증",
	}
	result, err := lore.BuildCommit(entry, "feat: 유효성 검사\n\n입력 유효성을 검사합니다.")
	require.NoError(t, err)

	assert.Contains(t, result, "feat: 유효성 검사")
	assert.Contains(t, result, "Directive: 항상 검증")
	assert.Contains(t, result, "🐙 Autopus <noreply@autopus.co>")
}
