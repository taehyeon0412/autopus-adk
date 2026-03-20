package content

import (
	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
)

// GitHookScript는 Git 훅 스크립트이다.
type GitHookScript struct {
	// Path는 훅 파일 경로이다.
	Path string
	// Content는 스크립트 내용이다.
	Content string
}

// GenerateHookConfigs는 설정에 따라 훅 설정을 생성한다.
// supportsHooks가 true이면 CLI 훅 설정을 반환하고,
// false이면 .git/hooks/ 스크립트를 반환한다.
func GenerateHookConfigs(cfg config.HooksConf, platform string, supportsHooks bool) ([]adapter.HookConfig, []GitHookScript, error) {
	if supportsHooks {
		return generateCLIHooks(cfg, platform), nil, nil
	}
	return nil, generateGitHooks(cfg), nil
}

// generateCLIHooks는 CLI 훅 설정을 생성한다.
func generateCLIHooks(cfg config.HooksConf, _ string) []adapter.HookConfig {
	var hooks []adapter.HookConfig

	if cfg.PreCommitArch {
		hooks = append(hooks, adapter.HookConfig{
			Event:   "PreToolUse",
			Command: "auto check --arch --quiet",
			Timeout: 30,
		})
	}

	if cfg.PreCommitLore {
		hooks = append(hooks, adapter.HookConfig{
			Event:   "PreToolUse",
			Command: "auto check --lore --quiet",
			Timeout: 30,
		})
	}

	if cfg.ReactCIFailure {
		hooks = append(hooks, adapter.HookConfig{
			Event:   "PostToolUse",
			Command: "auto react --ci-failure --quiet",
			Timeout: 60,
		})
	}

	if cfg.ReactReview {
		hooks = append(hooks, adapter.HookConfig{
			Event:   "PostToolUse",
			Command: "auto react --review --quiet",
			Timeout: 60,
		})
	}

	return hooks
}

// generateGitHooks는 .git/hooks/ 스크립트를 생성한다.
func generateGitHooks(cfg config.HooksConf) []GitHookScript {
	// arch 또는 lore 활성화된 경우에만 pre-commit 생성
	if !cfg.PreCommitArch && !cfg.PreCommitLore {
		return nil
	}

	script := buildPreCommitScript(cfg)
	return []GitHookScript{
		{
			Path:    ".git/hooks/pre-commit",
			Content: script,
		},
	}
}

// buildPreCommitScript는 pre-commit 스크립트를 생성한다.
func buildPreCommitScript(cfg config.HooksConf) string {
	s := "#!/bin/sh\n# Autopus-ADK pre-commit hook (자동 생성)\nset -e\n\n"

	if cfg.PreCommitArch {
		s += "# 아키텍처 규칙 검사\nauto check --arch --quiet\n\n"
	}

	if cfg.PreCommitLore {
		s += "# Lore 커밋 메시지 검사\nauto check --lore --quiet\n\n"
	}

	s += "exit 0\n"
	return s
}
