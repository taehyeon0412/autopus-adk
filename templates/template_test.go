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

func TestClaudeRouterTemplate(t *testing.T) {
	t.Parallel()
	e := tmpl.New()
	cfg := config.DefaultFullConfig("cmd-project")

	tmplPath := filepath.Join(templateRoot(), "claude", "commands", "auto-router.md.tmpl")
	result, err := e.RenderFile(tmplPath, cfg)
	require.NoError(t, err, "라우터 템플릿 렌더링 실패")
	assert.Contains(t, result, "cmd-project", "프로젝트명이 포함되어야 함")
	assert.True(t, len(result) > 100, "템플릿 결과가 너무 짧음")

	// 모든 서브커맨드가 포함되어야 함
	subcommands := []string{"plan", "go", "fix", "map", "review", "secure", "stale", "sync", "why"}
	for _, sub := range subcommands {
		assert.Contains(t, result, sub, "서브커맨드 %q가 포함되어야 함", sub)
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
	root := templateRoot()

	liteCfg := config.DefaultLiteConfig("test")
	fullCfg := config.DefaultFullConfig("test")

	// 라우터 템플릿에서 Full 모드 조건부 블록 확인
	tmplPath := filepath.Join(root, "claude", "commands", "auto-router.md.tmpl")

	liteResult, err := e.RenderFile(tmplPath, liteCfg)
	require.NoError(t, err)

	fullResult, err := e.RenderFile(tmplPath, fullCfg)
	require.NoError(t, err)

	// Full 모드에서는 go/review/secure 서브커맨드의 스킬 참조가 포함됨
	assert.Contains(t, fullResult, "tdd.md")
	// Lite 모드에서는 Full 전용 안내 메시지가 표시됨
	assert.Contains(t, liteResult, "Full mode only")
}
