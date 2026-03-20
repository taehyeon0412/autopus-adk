// Package templates는 템플릿 렌더링 통합 테스트이다.
package templates_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	tmpl "github.com/insajin/autopus-adk/pkg/template"
)

// 템플릿 루트 디렉터리 — 테스트 파일이 templates/ 디렉터리에 있으므로 현재 디렉터리 사용
func templateRoot() string {
	// 테스트 실행 위치 기준으로 templates/ 디렉터리 찾기
	dir, _ := os.Getwd()
	return dir
}

func TestSharedWorkflowTemplate_Lite(t *testing.T) {
	t.Parallel()
	e := tmpl.New()
	cfg := config.DefaultLiteConfig("my-project")

	tmplPath := filepath.Join(templateRoot(), "shared", "workflow.md.tmpl")
	result, err := e.RenderFile(tmplPath, cfg)
	require.NoError(t, err)

	assert.Contains(t, result, "my-project")
	assert.Contains(t, result, "lite")
	assert.Contains(t, result, "/plan")
	assert.Contains(t, result, "/go")
}

func TestSharedWorkflowTemplate_Full(t *testing.T) {
	t.Parallel()
	e := tmpl.New()
	cfg := config.DefaultFullConfig("full-project")

	tmplPath := filepath.Join(templateRoot(), "shared", "workflow.md.tmpl")
	result, err := e.RenderFile(tmplPath, cfg)
	require.NoError(t, err)

	assert.Contains(t, result, "full-project")
	assert.Contains(t, result, "full")
	assert.Contains(t, result, "Full 모드 기능")
}

func TestSharedAutopusYamlTemplate(t *testing.T) {
	t.Parallel()
	e := tmpl.New()
	cfg := config.DefaultLiteConfig("yaml-project")

	tmplPath := filepath.Join(templateRoot(), "shared", "autopus.yaml.tmpl")
	result, err := e.RenderFile(tmplPath, cfg)
	require.NoError(t, err)

	assert.Contains(t, result, "yaml-project")
	assert.Contains(t, result, "mode: lite")
	assert.Contains(t, result, "claude-code")
}

func TestClaudeCommandTemplates(t *testing.T) {
	t.Parallel()
	e := tmpl.New()
	cfg := config.DefaultFullConfig("cmd-project")

	commands := []string{
		"plan", "go", "sync", "fix", "why",
		"map", "secure", "review", "stale", "auto",
	}

	for _, cmd := range commands {
		cmd := cmd
		t.Run(cmd, func(t *testing.T) {
			t.Parallel()
			tmplPath := filepath.Join(templateRoot(), "claude", "commands", cmd+".md.tmpl")
			result, err := e.RenderFile(tmplPath, cfg)
			require.NoError(t, err, "커맨드 템플릿 렌더링 실패: %s", cmd)
			assert.Contains(t, result, "cmd-project", "프로젝트명이 포함되어야 함: %s", cmd)
			assert.True(t, len(result) > 100, "템플릿 결과가 너무 짧음: %s", cmd)
		})
	}
}

func TestCodexSkillTemplates(t *testing.T) {
	t.Parallel()
	e := tmpl.New()
	cfg := config.DefaultLiteConfig("codex-project")

	skills := []string{
		"auto-plan", "auto-go", "auto-fix", "auto-review", "auto-sync",
	}

	for _, skill := range skills {
		skill := skill
		t.Run(skill, func(t *testing.T) {
			t.Parallel()
			tmplPath := filepath.Join(templateRoot(), "codex", "skills", skill+".md.tmpl")
			result, err := e.RenderFile(tmplPath, cfg)
			require.NoError(t, err, "코덱스 스킬 템플릿 렌더링 실패: %s", skill)
			assert.Contains(t, result, "codex-project")
		})
	}
}

func TestGeminiSkillTemplates_HasFrontmatter(t *testing.T) {
	t.Parallel()
	e := tmpl.New()
	cfg := config.DefaultLiteConfig("gemini-project")

	skills := []string{
		"auto-plan", "auto-go", "auto-fix", "auto-review", "auto-sync",
	}

	for _, skill := range skills {
		skill := skill
		t.Run(skill, func(t *testing.T) {
			t.Parallel()
			tmplPath := filepath.Join(templateRoot(), "gemini", "skills", skill, "SKILL.md.tmpl")
			result, err := e.RenderFile(tmplPath, cfg)
			require.NoError(t, err, "제미니 스킬 템플릿 렌더링 실패: %s", skill)

			// YAML frontmatter 확인
			assert.True(t, strings.HasPrefix(result, "---"), "YAML frontmatter로 시작해야 함: %s", skill)
			assert.Contains(t, result, "name: "+skill)
			assert.Contains(t, result, "gemini-project")
		})
	}
}

func TestTemplates_FullModeConditionals(t *testing.T) {
	t.Parallel()
	e := tmpl.New()
	liteRoot := templateRoot()

	liteCfg := config.DefaultLiteConfig("test")
	fullCfg := config.DefaultFullConfig("test")

	// plan 커맨드에서 Full 모드 조건부 블록 확인
	tmplPath := filepath.Join(liteRoot, "claude", "commands", "plan.md.tmpl")

	liteResult, err := e.RenderFile(tmplPath, liteCfg)
	require.NoError(t, err)

	fullResult, err := e.RenderFile(tmplPath, fullCfg)
	require.NoError(t, err)

	// Full 모드에만 있어야 하는 내용
	assert.Contains(t, fullResult, "Full 모드 추가 기능")
	assert.NotContains(t, liteResult, "Full 모드 추가 기능")
}
