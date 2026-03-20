package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// AgentDefinition은 에이전트 정의이다.
type AgentDefinition struct {
	// Name은 에이전트 이름이다.
	Name string `yaml:"name"`
	// Role은 에이전트 역할 설명이다.
	Role string `yaml:"role"`
	// ModelTier는 에이전트가 사용하는 모델 티어이다 (opus, sonnet, haiku).
	ModelTier string `yaml:"model_tier"`
	// Category는 에이전트 카테고리이다.
	Category string `yaml:"category"`
	// Triggers는 에이전트 활성화 트리거이다.
	Triggers []string `yaml:"triggers"`
	// Skills는 에이전트가 사용하는 스킬 목록이다.
	Skills []string `yaml:"skills"`
	// Instructions는 에이전트 역할 지침이다.
	Instructions string `yaml:"-"`
}

// agentFrontmatter는 마크다운 프론트매터 파싱용 내부 구조체이다.
type agentFrontmatter struct {
	Name      string   `yaml:"name"`
	Role      string   `yaml:"role"`
	ModelTier string   `yaml:"model_tier"`
	Category  string   `yaml:"category"`
	Triggers  []string `yaml:"triggers"`
	Skills    []string `yaml:"skills"`
}

// LoadAgents는 디렉토리에서 에이전트 정의 파일을 로드한다.
func LoadAgents(dir string) ([]AgentDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("에이전트 디렉토리 읽기 실패 %s: %w", dir, err)
	}

	var agents []AgentDefinition
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		agent, err := parseAgentFile(path)
		if err != nil {
			return nil, fmt.Errorf("에이전트 파일 파싱 실패 %s: %w", path, err)
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// ConvertAgentToPlatform은 에이전트를 플랫폼 형식으로 변환한다.
// claude: .claude/agents/autopus/<name>.md
// codex: AGENTS.md 섹션
// gemini: .gemini/skills/auto-agent-<name>/SKILL.md
func ConvertAgentToPlatform(agent AgentDefinition, platform string) (string, error) {
	switch platform {
	case "claude", "claude-code":
		return convertAgentClaude(agent), nil
	case "codex":
		return convertAgentCodex(agent), nil
	case "gemini", "gemini-cli":
		return convertAgentGemini(agent), nil
	default:
		return "", fmt.Errorf("지원하지 않는 플랫폼: %q", platform)
	}
}

// convertAgentClaude는 Claude agents 형식으로 변환한다.
func convertAgentClaude(agent AgentDefinition) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", agent.Name))
	sb.WriteString(fmt.Sprintf("description: %s\n", agent.Role))
	if agent.ModelTier != "" {
		sb.WriteString(fmt.Sprintf("model_tier: %s\n", agent.ModelTier))
	}
	if len(agent.Triggers) > 0 {
		sb.WriteString("triggers:\n")
		for _, t := range agent.Triggers {
			sb.WriteString(fmt.Sprintf("  - %s\n", t))
		}
	}
	if len(agent.Skills) > 0 {
		sb.WriteString("skills:\n")
		for _, s := range agent.Skills {
			sb.WriteString(fmt.Sprintf("  - %s\n", s))
		}
	}
	sb.WriteString("---\n\n")
	if agent.Instructions != "" {
		sb.WriteString(agent.Instructions)
	} else {
		sb.WriteString(fmt.Sprintf("# %s\n\n%s\n", agent.Name, agent.Role))
	}
	return sb.String()
}

// convertAgentCodex는 AGENTS.md 섹션 형식으로 변환한다.
func convertAgentCodex(agent AgentDefinition) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Agent: %s\n\n", agent.Name))
	sb.WriteString(fmt.Sprintf("**Role:** %s\n\n", agent.Role))
	if agent.ModelTier != "" {
		sb.WriteString(fmt.Sprintf("**Model Tier:** %s\n\n", agent.ModelTier))
	}
	if len(agent.Skills) > 0 {
		sb.WriteString("**Skills:** ")
		sb.WriteString(strings.Join(agent.Skills, ", "))
		sb.WriteString("\n\n")
	}
	if agent.Instructions != "" {
		sb.WriteString(agent.Instructions)
	}
	return sb.String()
}

// convertAgentGemini는 Gemini SKILL.md 형식으로 변환한다.
func convertAgentGemini(agent AgentDefinition) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: auto-agent-%s\n", agent.Name))
	sb.WriteString(fmt.Sprintf("description: %s\n", agent.Role))
	if len(agent.Triggers) > 0 {
		sb.WriteString("triggers:\n")
		for _, t := range agent.Triggers {
			sb.WriteString(fmt.Sprintf("  - %s\n", t))
		}
	}
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# auto-agent-%s\n\n", agent.Name))
	if agent.Instructions != "" {
		sb.WriteString(agent.Instructions)
	} else {
		sb.WriteString(agent.Role + "\n")
	}
	return sb.String()
}

// parseAgentFile은 마크다운 에이전트 파일을 파싱한다.
func parseAgentFile(path string) (AgentDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AgentDefinition{}, fmt.Errorf("파일 읽기 실패: %w", err)
	}

	content := string(data)
	fm, body, err := splitFrontmatter(content)
	if err != nil {
		return AgentDefinition{}, fmt.Errorf("프론트매터 파싱 실패: %w", err)
	}

	var frontmatter agentFrontmatter
	if err := yaml.Unmarshal([]byte(fm), &frontmatter); err != nil {
		return AgentDefinition{}, fmt.Errorf("YAML 파싱 실패: %w", err)
	}

	agent := AgentDefinition{
		Name:         frontmatter.Name,
		Role:         frontmatter.Role,
		ModelTier:    frontmatter.ModelTier,
		Category:     frontmatter.Category,
		Triggers:     frontmatter.Triggers,
		Skills:       frontmatter.Skills,
		Instructions: strings.TrimSpace(body),
	}

	if agent.Name == "" {
		base := filepath.Base(path)
		agent.Name = strings.TrimSuffix(base, ".md")
	}

	return agent, nil
}
