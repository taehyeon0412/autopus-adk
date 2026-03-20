// Package claude_test는 Claude 어댑터 훅 관련 추가 테스트이다.
package claude_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/adapter/claude"
)

// TestClaudeAdapter_InstallHooks_Empty는 훅이 없는 경우 InstallHooks를 테스트한다.
func TestClaudeAdapter_InstallHooks_Empty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	a := claude.NewWithRoot(dir)

	err := a.InstallHooks(context.Background(), nil)
	assert.NoError(t, err)

	// settings.json이 생성되어야 함
	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	_, statErr := os.Stat(settingsPath)
	assert.NoError(t, statErr)
}

// TestClaudeAdapter_InstallHooks_WithHooks는 훅 설정을 포함한 InstallHooks를 테스트한다.
func TestClaudeAdapter_InstallHooks_WithHooks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	a := claude.NewWithRoot(dir)

	hooks := []adapter.HookConfig{
		{Event: "PostCommit", Command: "echo done", Timeout: 30},
		{Event: "PrePush", Command: "go test ./...", Timeout: 60},
	}

	err := a.InstallHooks(context.Background(), hooks)
	require.NoError(t, err)

	// settings.json 내용 확인
	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	data, readErr := os.ReadFile(settingsPath)
	require.NoError(t, readErr)

	var settings map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &settings))

	// hooks 필드 확인
	hooksVal, ok := settings["hooks"]
	assert.True(t, ok, "hooks 필드가 있어야 함")
	hooksSlice, ok := hooksVal.([]interface{})
	assert.True(t, ok)
	assert.Len(t, hooksSlice, 2)
}

// TestClaudeAdapter_InstallHooks_MergesExisting는 기존 settings.json과 병합을 테스트한다.
func TestClaudeAdapter_InstallHooks_MergesExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// 기존 settings.json 생성
	settingsDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))
	existing := map[string]interface{}{
		"theme": "dark",
	}
	data, _ := json.Marshal(existing)
	require.NoError(t, os.WriteFile(filepath.Join(settingsDir, "settings.json"), data, 0644))

	a := claude.NewWithRoot(dir)
	hooks := []adapter.HookConfig{
		{Event: "PostCommit", Command: "echo hi", Timeout: 10},
	}

	err := a.InstallHooks(context.Background(), hooks)
	require.NoError(t, err)

	// 결과 확인: 기존 필드와 새 hooks 모두 있어야 함
	updated, readErr := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	require.NoError(t, readErr)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(updated, &result))

	// hooks 추가됨
	_, hasHooks := result["hooks"]
	assert.True(t, hasHooks)
}

// TestClaudeAdapter_InstallHooks_InvalidJSON은 잘못된 JSON settings.json이 있을 때를 테스트한다.
func TestClaudeAdapter_InstallHooks_InvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// 잘못된 JSON 파일 생성
	settingsDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte("{invalid json}"), 0644))

	a := claude.NewWithRoot(dir)
	hooks := []adapter.HookConfig{
		{Event: "PostCommit", Command: "echo test", Timeout: 10},
	}

	// 잘못된 JSON이어도 오류 없이 처리되어야 함 (기본 맵으로 초기화)
	err := a.InstallHooks(context.Background(), hooks)
	assert.NoError(t, err)
}

// TestClaudeAdapter_Clean_RemovesMarker는 Clean이 CLAUDE.md 마커 섹션을 제거하는지 테스트한다.
func TestClaudeAdapter_Clean_RemovesMarker(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// CLAUDE.md에 마커 섹션 포함 콘텐츠 생성
	claudePath := filepath.Join(dir, "CLAUDE.md")
	content := "# 내 프로젝트\n\n<!-- AUTOPUS:BEGIN -->\n자동 생성 섹션\n<!-- AUTOPUS:END -->\n\n## 사용자 섹션\n"
	require.NoError(t, os.WriteFile(claudePath, []byte(content), 0644))

	a := claude.NewWithRoot(dir)
	err := a.Clean(context.Background())
	require.NoError(t, err)

	// 마커 섹션이 제거되고 사용자 섹션은 보존되어야 함
	data, readErr := os.ReadFile(claudePath)
	require.NoError(t, readErr)
	assert.NotContains(t, string(data), "AUTOPUS:BEGIN")
	assert.Contains(t, string(data), "사용자 섹션")
}
