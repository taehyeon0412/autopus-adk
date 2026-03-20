// Package claude는 Claude Code 어댑터 테스트이다.
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

func TestClaudeAdapter_Name(t *testing.T) {
	t.Parallel()
	a := claude.New()
	assert.Equal(t, "claude-code", a.Name())
}

func TestClaudeAdapter_CLIBinary(t *testing.T) {
	t.Parallel()
	a := claude.New()
	assert.Equal(t, "claude", a.CLIBinary())
}

func TestClaudeAdapter_Version(t *testing.T) {
	t.Parallel()
	a := claude.New()
	assert.NotEmpty(t, a.Version())
}

func TestClaudeAdapter_SupportsHooks(t *testing.T) {
	t.Parallel()
	a := claude.New()
	assert.True(t, a.SupportsHooks())
}

func TestClaudeAdapter_Detect_NotInstalled(t *testing.T) {
	// t.Setenv는 t.Parallel()과 함께 사용할 수 없음
	t.Setenv("PATH", t.TempDir())
	a := claude.New()
	ok, err := a.Detect(context.Background())
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestClaudeAdapter_Generate_CreatesDirectories(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := claude.NewWithRoot(dir)
	cfg := config.DefaultLiteConfig("test-project")
	cfg.Platforms = []string{"claude-code"}

	files, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, files)

	// 디렉터리 생성 확인
	expectedDirs := []string{
		".claude/rules/autopus",
		".claude/skills/autopus",
		".claude/commands/autopus",
		".claude/agents/autopus",
	}
	for _, d := range expectedDirs {
		info, statErr := os.Stat(filepath.Join(dir, d))
		require.NoError(t, statErr, "디렉터리가 존재해야 함: %s", d)
		assert.True(t, info.IsDir(), "%s는 디렉터리여야 함", d)
	}
}

func TestClaudeAdapter_Generate_ClaudeMD_MarkerSection(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := claude.NewWithRoot(dir)
	cfg := config.DefaultLiteConfig("test-project")

	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	require.NoError(t, err)
	content := string(data)

	// 마커 섹션 확인
	assert.Contains(t, content, "<!-- AUTOPUS:BEGIN -->")
	assert.Contains(t, content, "<!-- AUTOPUS:END -->")
	assert.Contains(t, content, "test-project")
}

func TestClaudeAdapter_Generate_ClaudeMD_PreservesUserContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := claude.NewWithRoot(dir)
	cfg := config.DefaultLiteConfig("test-project")

	// 기존 사용자 컨텐츠가 있는 CLAUDE.md 생성
	userContent := "# My Custom Rules\n\nSome user-defined rules here.\n"
	err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(userContent), 0644)
	require.NoError(t, err)

	_, err = a.Generate(context.Background(), cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	require.NoError(t, err)
	content := string(data)

	// 사용자 컨텐츠가 보존되어야 함
	assert.Contains(t, content, "My Custom Rules")
	assert.Contains(t, content, "Some user-defined rules here.")
	// autopus 섹션도 있어야 함
	assert.Contains(t, content, "<!-- AUTOPUS:BEGIN -->")
	assert.Contains(t, content, "<!-- AUTOPUS:END -->")
}

func TestClaudeAdapter_Update_ChecksumComparison(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := claude.NewWithRoot(dir)
	cfg := config.DefaultLiteConfig("test-project")

	// 초기 생성
	files1, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, files1)

	// 업데이트 (변경 없음)
	files2, err := a.Update(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, files2)

	// 파일 수가 동일해야 함
	assert.Equal(t, len(files1.Files), len(files2.Files))
}

func TestClaudeAdapter_Update_PreservesMarkerContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := claude.NewWithRoot(dir)
	cfg := config.DefaultLiteConfig("test-project")

	// 초기 생성
	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)

	// CLAUDE.md에 사용자 컨텐츠 추가 (마커 외부)
	claudePath := filepath.Join(dir, "CLAUDE.md")
	data, err := os.ReadFile(claudePath)
	require.NoError(t, err)

	// 마커 뒤에 사용자 컨텐츠 추가
	userExtra := "\n\n## User Added Section\n\nThis should be preserved.\n"
	err = os.WriteFile(claudePath, append(data, []byte(userExtra)...), 0644)
	require.NoError(t, err)

	// 업데이트 실행
	_, err = a.Update(context.Background(), cfg)
	require.NoError(t, err)

	// 사용자 컨텐츠가 보존되어야 함
	updated, err := os.ReadFile(claudePath)
	require.NoError(t, err)
	assert.Contains(t, string(updated), "User Added Section")
	assert.Contains(t, string(updated), "This should be preserved.")
}

func TestClaudeAdapter_InstallHooks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := claude.NewWithRoot(dir)

	hooks := []interface{ GetEvent() string }{} // 빈 훅 테스트
	_ = hooks

	// 빈 훅 목록으로 설치
	err := a.InstallHooks(context.Background(), nil)
	require.NoError(t, err)

	// settings.json 생성 확인
	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	_, statErr := os.Stat(settingsPath)
	require.NoError(t, statErr, "settings.json이 생성되어야 함")
}

func TestClaudeAdapter_Validate_MissingFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := claude.NewWithRoot(dir)

	// 파일 없이 검증
	errs, err := a.Validate(context.Background())
	require.NoError(t, err)
	// 파일이 없으므로 검증 오류가 있어야 함
	assert.NotEmpty(t, errs)
}

func TestClaudeAdapter_Validate_AfterGenerate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := claude.NewWithRoot(dir)
	cfg := config.DefaultLiteConfig("test-project")

	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)

	errs, err := a.Validate(context.Background())
	require.NoError(t, err)
	assert.Empty(t, errs, "Generate 후에는 검증 오류가 없어야 함")
}

func TestClaudeAdapter_Clean(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := claude.NewWithRoot(dir)
	cfg := config.DefaultLiteConfig("test-project")

	// 파일 생성 후 정리
	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)

	err = a.Clean(context.Background())
	require.NoError(t, err)

	// autopus 디렉터리가 제거되어야 함
	_, statErr := os.Stat(filepath.Join(dir, ".claude", "rules", "autopus"))
	assert.True(t, os.IsNotExist(statErr), "autopus 규칙 디렉터리가 제거되어야 함")
}
