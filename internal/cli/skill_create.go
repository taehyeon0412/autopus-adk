package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

// @AX:NOTE [AUTO] @AX:REASON: skill template format must match .claude/skills/ conventions
const skillTemplate = `---
name: {{.Name}}
description: {{.Description}}
triggers:
{{- range .Triggers}}
  - {{.}}
{{- end}}
category: {{.Category}}
---

# {{.Name}}

{{.Description}}

## 사용법

이 스킬을 사용하려면 다음 트리거 중 하나를 사용하세요:
{{range .Triggers}}- {{.}}
{{end}}
## 구현 지침

TODO: 이 스킬의 구현 지침을 작성하세요.
`

type skillTemplateData struct {
	Name        string
	Description string
	Triggers    []string
	Category    string
}

// skillFrontmatter mirrors the required fields of a skill .md frontmatter block.
type skillFrontmatter struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Triggers    []string `yaml:"triggers"`
}

// newSkillCreateCmd creates the `auto skill create` subcommand.
func newSkillCreateCmd() *cobra.Command {
	var (
		description string
		triggers    string
		category    string
		write       bool
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "새 스킬 파일 생성",
		Long:  "새 스킬 .md 파일을 생성합니다. 기본값은 dry-run(stdout 출력)이며, --write 플래그로 파일을 저장합니다.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			return runSkillCreate(cmd, name, description, triggers, category, write)
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "스킬 설명 (필수)")
	cmd.Flags().StringVar(&triggers, "triggers", "", "트리거 목록 (쉼표 구분)")
	cmd.Flags().StringVar(&category, "category", "general", "스킬 카테고리")
	cmd.Flags().BoolVar(&write, "write", false, "파일로 저장 (.claude/skills/autopus/)")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func runSkillCreate(cmd *cobra.Command, name, description, triggers, category string, write bool) error {
	if strings.ContainsAny(name, "/\\") || filepath.Base(name) != name {
		return fmt.Errorf("invalid skill name %q: must not contain path separators", name)
	}

	triggerList := parseTriggers(triggers, name)

	// Validate: check for trigger conflicts with existing skills
	if write {
		if conflicts := findTriggerConflicts(".claude/skills/autopus", triggerList); len(conflicts) > 0 {
			return fmt.Errorf("trigger conflict: %s already used in %s", conflicts[0].trigger, conflicts[0].file)
		}
	}

	data := skillTemplateData{
		Name:        name,
		Description: description,
		Triggers:    triggerList,
		Category:    category,
	}

	tmpl, err := template.New("skill").Parse(skillTemplate)
	if err != nil {
		return fmt.Errorf("template parse error: %w", err)
	}

	// Render to buffer first so we can validate the frontmatter before writing.
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("template render error: %w", err)
	}

	if err := validateSkillFrontmatter(buf.Bytes()); err != nil {
		return err
	}

	if !write {
		_, err := cmd.OutOrStdout().Write(buf.Bytes())
		return err
	}

	// Write mode: save to .claude/skills/autopus/
	skillsDir := ".claude/skills/autopus"
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	outPath := filepath.Join(skillsDir, name+".md")
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("failed to write skill file: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Skill created: %s\n", outPath)
	return nil
}

// validateSkillFrontmatter extracts and parses the YAML frontmatter block from
// content, verifying that name, description, and triggers fields are present.
func validateSkillFrontmatter(content []byte) error {
	fm, err := extractFrontmatter(content)
	if err != nil {
		return fmt.Errorf("skill frontmatter invalid YAML: %w", err)
	}
	var s skillFrontmatter
	if err := yaml.Unmarshal(fm, &s); err != nil {
		return fmt.Errorf("skill frontmatter parse error: %w", err)
	}
	if s.Name == "" {
		return fmt.Errorf("skill frontmatter missing required field: name")
	}
	if s.Description == "" {
		return fmt.Errorf("skill frontmatter missing required field: description")
	}
	if len(s.Triggers) == 0 {
		return fmt.Errorf("skill frontmatter missing required field: triggers")
	}
	return nil
}

// extractFrontmatter returns the YAML bytes between the first --- delimiters.
func extractFrontmatter(content []byte) ([]byte, error) {
	s := string(content)
	if !strings.HasPrefix(s, "---") {
		return nil, fmt.Errorf("no frontmatter delimiter found")
	}
	rest := s[3:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return nil, fmt.Errorf("frontmatter closing delimiter not found")
	}
	return []byte(strings.TrimSpace(rest[:end])), nil
}

type triggerConflict struct {
	trigger string
	file    string
}

func findTriggerConflicts(skillsDir string, triggers []string) []triggerConflict {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil // directory doesn't exist yet: no conflicts
	}

	var conflicts []triggerConflict
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(skillsDir, entry.Name()))
		if err != nil {
			continue
		}
		content := string(data)
		for _, t := range triggers {
			if strings.Contains(content, "- "+t) {
				conflicts = append(conflicts, triggerConflict{trigger: t, file: entry.Name()})
			}
		}
	}
	return conflicts
}

func parseTriggers(triggers, name string) []string {
	if triggers == "" {
		return []string{name}
	}
	parts := strings.Split(triggers, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
