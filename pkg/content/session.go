package content

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// maxContextChars는 ContextSummary 최대 문자 수이다 (2000 토큰 기준 ~8000자).
const maxContextChars = 8000

// SessionState는 세션 연속성 상태이다.
type SessionState struct {
	// WorkflowPhase는 현재 워크플로우 단계이다.
	WorkflowPhase string `yaml:"workflow_phase"`
	// CompletedTasks는 완료된 태스크 목록이다.
	CompletedTasks []string `yaml:"completed_tasks,omitempty"`
	// PendingDecisions는 미결 의사결정 목록이다.
	PendingDecisions []string `yaml:"pending_decisions,omitempty"`
	// ContextSummary는 세션 컨텍스트 요약이다 (최대 2000 토큰).
	ContextSummary string `yaml:"context_summary,omitempty"`
}

// SaveState는 세션 상태를 파일에 저장한다.
// ContextSummary는 최대 2000 토큰 제한을 적용한다.
func SaveState(path string, state *SessionState) error {
	if state == nil {
		return fmt.Errorf("저장할 상태가 nil입니다")
	}

	// ContextSummary 토큰 제한 적용
	if len(state.ContextSummary) > maxContextChars {
		state.ContextSummary = state.ContextSummary[:maxContextChars]
	}

	// YAML 직렬화
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("상태 직렬화 실패: %w", err)
	}

	// 마크다운 형식으로 래핑
	content := buildSessionFile(state, string(data))

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("상태 파일 저장 실패 %s: %w", path, err)
	}
	return nil
}

// LoadState는 파일에서 세션 상태를 로드한다.
func LoadState(path string) (*SessionState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("상태 파일 읽기 실패 %s: %w", path, err)
	}

	content := string(data)
	yamlContent := extractYAMLBlock(content)
	if yamlContent == "" {
		yamlContent = content
	}

	var state SessionState
	if err := yaml.Unmarshal([]byte(yamlContent), &state); err != nil {
		return nil, fmt.Errorf("상태 파싱 실패: %w", err)
	}

	return &state, nil
}

// buildSessionFile은 .auto-continue.md 파일 내용을 생성한다.
func buildSessionFile(state *SessionState, yamlData string) string {
	var sb strings.Builder
	sb.WriteString("# Auto Continue\n\n")
	sb.WriteString("<!-- Autopus-ADK 세션 연속성 파일 — 직접 편집하지 마세요 -->\n\n")
	sb.WriteString("```yaml\n")
	sb.WriteString(yamlData)
	sb.WriteString("```\n\n")

	if state.ContextSummary != "" {
		sb.WriteString("## 컨텍스트 요약\n\n")
		sb.WriteString(state.ContextSummary)
		sb.WriteString("\n")
	}
	return sb.String()
}

// extractYAMLBlock은 마크다운에서 YAML 코드 블록을 추출한다.
func extractYAMLBlock(content string) string {
	start := strings.Index(content, "```yaml\n")
	if start < 0 {
		return ""
	}
	start += len("```yaml\n")
	end := strings.Index(content[start:], "```")
	if end < 0 {
		return ""
	}
	return content[start : start+end]
}
