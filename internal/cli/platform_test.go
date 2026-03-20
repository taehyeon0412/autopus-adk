// Package cli는 platform 커맨드 테스트이다.
package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlatformListCmd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// init 후 platform list
	initCmd := newTestRootCmd()
	initCmd.SetArgs([]string{"init", "--lite", "--dir", dir, "--project", "test-proj", "--platforms", "claude-code"})
	require.NoError(t, initCmd.Execute())

	var out bytes.Buffer
	listCmd := newTestRootCmd()
	listCmd.SetOut(&out)
	listCmd.SetArgs([]string{"platform", "list", "--dir", dir})
	require.NoError(t, listCmd.Execute())

	output := out.String()
	assert.Contains(t, output, "claude-code")
}

func TestPlatformAddCmd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// init
	initCmd := newTestRootCmd()
	initCmd.SetArgs([]string{"init", "--lite", "--dir", dir, "--project", "test-proj", "--platforms", "claude-code"})
	require.NoError(t, initCmd.Execute())

	// codex 추가
	addCmd := newTestRootCmd()
	addCmd.SetArgs([]string{"platform", "add", "codex", "--dir", dir})
	require.NoError(t, addCmd.Execute())

	// autopus.yaml에 codex가 추가되었는지 확인
	data, err := os.ReadFile(filepath.Join(dir, "autopus.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "codex")
}

func TestPlatformRemoveCmd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// init with two platforms
	initCmd := newTestRootCmd()
	initCmd.SetArgs([]string{"init", "--lite", "--dir", dir, "--project", "test-proj",
		"--platforms", "claude-code,codex"})
	require.NoError(t, initCmd.Execute())

	// codex 제거
	removeCmd := newTestRootCmd()
	removeCmd.SetArgs([]string{"platform", "remove", "codex", "--dir", dir})
	require.NoError(t, removeCmd.Execute())

	// autopus.yaml에서 codex가 제거되었는지 확인
	data, err := os.ReadFile(filepath.Join(dir, "autopus.yaml"))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "codex")
}

func TestPlatformAddCmd_InvalidPlatform(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	initCmd := newTestRootCmd()
	initCmd.SetArgs([]string{"init", "--lite", "--dir", dir, "--project", "test-proj", "--platforms", "claude-code"})
	require.NoError(t, initCmd.Execute())

	// 잘못된 플랫폼 추가 시도
	addCmd := newTestRootCmd()
	addCmd.SetArgs([]string{"platform", "add", "invalid-platform", "--dir", dir})
	err := addCmd.Execute()
	assert.Error(t, err, "잘못된 플랫폼은 에러를 반환해야 함")
}
