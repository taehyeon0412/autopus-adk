// Package claude_test는 Claude 어댑터 추가 테스트이다.
package claude_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter/claude"
	"github.com/insajin/autopus-adk/pkg/config"
)

// TestClaudeAdapter_CleanRemovesFiles는 Clean이 파일을 삭제하는지 테스트한다.
func TestClaudeAdapter_CleanRemovesFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	a := claude.NewWithRoot(dir)
	cfg := config.DefaultFullConfig("test-project")

	// 먼저 Generate로 파일 생성
	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)

	// 디렉터리가 생성되었는지 확인
	_, statErr := os.Stat(filepath.Join(dir, ".claude", "rules", "autopus"))
	require.NoError(t, statErr)

	// Clean 실행
	err = a.Clean(context.Background())
	require.NoError(t, err)

	// autopus 디렉터리가 삭제되었는지 확인
	_, statErr = os.Stat(filepath.Join(dir, ".claude", "rules", "autopus"))
	assert.True(t, os.IsNotExist(statErr), "autopus 디렉터리가 삭제되어야 함")
}

// TestClaudeAdapter_Clean_NonExistent는 존재하지 않는 디렉터리에 대한 Clean을 테스트한다.
func TestClaudeAdapter_Clean_NonExistent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	a := claude.NewWithRoot(dir)

	// 파일 없는 상태에서 Clean 실행 (오류 없어야 함)
	err := a.Clean(context.Background())
	assert.NoError(t, err)
}

// TestClaudeAdapter_Validate_NoErrors는 Generate 후 Validate 오류가 없는지 테스트한다.
func TestClaudeAdapter_Validate_NoErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	a := claude.NewWithRoot(dir)
	cfg := config.DefaultFullConfig("test-project")

	// Generate 실행
	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)

	// Validate 실행 - 오류 없어야 함
	errs, err := a.Validate(context.Background())
	require.NoError(t, err)
	// 생성 직후이므로 오류 없어야 함
	for _, e := range errs {
		assert.NotEqual(t, "error", e.Level, "오류 수준 검증 오류: %s", e.Message)
	}
}

// TestClaudeAdapter_Validate_MissingDirectories는 디렉터리 없는 상태에서 Validate를 테스트한다.
func TestClaudeAdapter_Validate_MissingDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	a := claude.NewWithRoot(dir)

	// Generate 없이 Validate 실행
	errs, err := a.Validate(context.Background())
	require.NoError(t, err)
	// 오류가 있어야 함
	assert.NotEmpty(t, errs)
}

// TestClaudeAdapter_Generate_FullMode는 Full 모드에서 Generate를 테스트한다.
func TestClaudeAdapter_Generate_FullMode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	a := claude.NewWithRoot(dir)
	cfg := config.DefaultFullConfig("test-project")

	files, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, files)

	// CLAUDE.md 내용 확인
	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "AUTOPUS:BEGIN")
}

// TestClaudeAdapter_Generate_WithExistingMarker는 기존 마커 섹션 업데이트를 테스트한다.
func TestClaudeAdapter_Generate_WithExistingMarker(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	a := claude.NewWithRoot(dir)
	cfg := config.DefaultFullConfig("project-v1")

	// 첫 번째 Generate
	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)

	// 프로젝트 이름 변경 후 두 번째 Generate
	cfg2 := config.DefaultFullConfig("project-v2")
	_, err = a.Generate(context.Background(), cfg2)
	require.NoError(t, err)

	// 두 번째 설정 내용이 반영되었는지 확인
	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "project-v2")
}
