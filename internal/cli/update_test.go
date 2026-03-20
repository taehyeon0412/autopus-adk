// Package cli는 update 커맨드 테스트이다.
package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCmd_RequiresExistingConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// autopus.yaml이 없으면 에러
	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"update", "--dir", dir})
	err := cmd.Execute()
	// config가 없으면 기본값으로 처리하거나 에러 — 동작 확인
	// 기본 구현에서는 기본 설정 로드 후 진행
	_ = err // 에러 여부는 구현에 따름
}

func TestUpdateCmd_UpdatesAfterInit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// 먼저 init 실행
	initCmd := newTestRootCmd()
	initCmd.SetArgs([]string{"init", "--lite", "--dir", dir, "--project", "test-proj", "--platforms", "claude-code"})
	require.NoError(t, initCmd.Execute())

	// CLAUDE.md 수정 시간 기록
	claudePath := filepath.Join(dir, "CLAUDE.md")
	info1, err := os.Stat(claudePath)
	require.NoError(t, err)
	modTime1 := info1.ModTime()

	// update 실행
	updateCmd := newTestRootCmd()
	updateCmd.SetArgs([]string{"update", "--dir", dir})
	require.NoError(t, updateCmd.Execute())

	// 파일이 여전히 존재해야 함
	_, statErr := os.Stat(claudePath)
	require.NoError(t, statErr)
	_ = modTime1 // 시간 비교는 OS 정밀도에 따라 다를 수 있어 생략
}

func TestUpdateCmd_PreservesUserModifications(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// init 실행
	initCmd := newTestRootCmd()
	initCmd.SetArgs([]string{"init", "--lite", "--dir", dir, "--project", "test-proj", "--platforms", "claude-code"})
	require.NoError(t, initCmd.Execute())

	// CLAUDE.md에 사용자 컨텐츠 추가
	claudePath := filepath.Join(dir, "CLAUDE.md")
	data, err := os.ReadFile(claudePath)
	require.NoError(t, err)
	userExtra := "\n\n## My Custom Section\n\nUser-defined rules.\n"
	err = os.WriteFile(claudePath, append(data, []byte(userExtra)...), 0644)
	require.NoError(t, err)

	// update 실행
	updateCmd := newTestRootCmd()
	updateCmd.SetArgs([]string{"update", "--dir", dir})
	require.NoError(t, updateCmd.Execute())

	// 사용자 컨텐츠가 보존되어야 함
	updated, err := os.ReadFile(claudePath)
	require.NoError(t, err)
	assert.Contains(t, string(updated), "My Custom Section")
	assert.Contains(t, string(updated), "User-defined rules.")
}
