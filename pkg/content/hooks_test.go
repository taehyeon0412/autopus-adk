// Package content_test는 훅 설정 생성 패키지의 테스트이다.
package content_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/content"
)

func TestGenerateHookConfigs_WithHooks(t *testing.T) {
	t.Parallel()

	cfg := config.HooksConf{
		PreCommitArch:  true,
		PreCommitLore:  true,
		ReactCIFailure: false,
		ReactReview:    false,
	}

	hooks, gitHooks, err := content.GenerateHookConfigs(cfg, "claude", true)
	require.NoError(t, err)
	// CLI 훅 지원 시 HookConfig 반환
	assert.NotEmpty(t, hooks)
	assert.Empty(t, gitHooks)
	// PreCommitArch, PreCommitLore 각각 포함
	events := make([]string, 0, len(hooks))
	for _, h := range hooks {
		events = append(events, h.Event)
	}
	assert.Contains(t, events, "PreToolUse")
}

func TestGenerateHookConfigs_WithoutHooks(t *testing.T) {
	t.Parallel()

	cfg := config.HooksConf{
		PreCommitArch: true,
		PreCommitLore: true,
	}

	hooks, gitHooks, err := content.GenerateHookConfigs(cfg, "codex", false)
	require.NoError(t, err)
	// CLI 훅 미지원 시 git 훅 스크립트 반환
	assert.Empty(t, hooks)
	assert.NotEmpty(t, gitHooks)
	// .git/hooks/pre-commit 포함
	var paths []string
	for _, g := range gitHooks {
		paths = append(paths, g.Path)
	}
	assert.Contains(t, paths, ".git/hooks/pre-commit")
}

func TestGenerateHookConfigs_AllDisabled(t *testing.T) {
	t.Parallel()

	cfg := config.HooksConf{
		PreCommitArch:  false,
		PreCommitLore:  false,
		ReactCIFailure: false,
		ReactReview:    false,
	}

	hooks, gitHooks, err := content.GenerateHookConfigs(cfg, "claude", true)
	require.NoError(t, err)
	assert.Empty(t, hooks)
	assert.Empty(t, gitHooks)
}

func TestGitHookScript_Content(t *testing.T) {
	t.Parallel()

	cfg := config.HooksConf{
		PreCommitArch: true,
	}

	_, gitHooks, err := content.GenerateHookConfigs(cfg, "gemini", false)
	require.NoError(t, err)
	require.NotEmpty(t, gitHooks)

	// 스크립트에 auto check --arch --quiet 포함
	assert.Contains(t, gitHooks[0].Content, "auto check --arch --quiet")
}
