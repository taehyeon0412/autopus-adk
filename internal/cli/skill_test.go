// Package cli_test는 skill 명령어 테스트이다.
package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/internal/cli"
)

func TestSkillListCmd(t *testing.T) {
	t.Parallel()

	// 임시 스킬 디렉토리 설정
	dir := t.TempDir()
	writeTestSkill(t, dir, "tdd.md", `---
name: tdd
description: TDD 스킬
category: methodology
triggers:
  - tdd
---
body`)
	writeTestSkill(t, dir, "planning.md", `---
name: planning
description: 기획 스킬
category: workflow
triggers:
  - plan
---
body`)

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"skill", "list", "--skills-dir", dir})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "tdd")
	assert.Contains(t, output, "planning")
}

func TestSkillInfoCmd(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeTestSkill(t, dir, "tdd.md", `---
name: tdd
description: TDD 스킬
category: methodology
triggers:
  - tdd
  - test
---

# TDD Skill

RED-GREEN-REFACTOR 사이클을 적용합니다.`)

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"skill", "info", "tdd", "--skills-dir", dir})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "tdd")
	assert.Contains(t, output, "TDD 스킬")
	assert.Contains(t, output, "methodology")
}

func TestSkillInfoCmd_NotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"skill", "info", "nonexistent", "--skills-dir", dir})

	var buf bytes.Buffer
	cmd.SetErr(&buf)
	err := cmd.Execute()
	assert.Error(t, err)
}

// writeTestSkill은 테스트용 스킬 파일을 생성한다.
func writeTestSkill(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	require.NoError(t, err)
}
