package content

import (
	"fmt"
	"strings"

	"github.com/insajin/autopus-adk/pkg/config"
)

// WorkflowDoc는 파싱된 워크플로우 문서이다.
type WorkflowDoc struct {
	// Policies는 워크플로우 정책 목록이다.
	Policies []Policy
	// Phases는 워크플로우 단계 목록이다.
	Phases []Phase
}

// Policy는 워크플로우 정책이다.
type Policy struct {
	// Description은 정책 설명이다.
	Description string
}

// Phase는 워크플로우 단계이다.
type Phase struct {
	// Name은 단계 이름이다.
	Name string
	// Description은 단계 설명이다.
	Description string
}

// ParseWorkflow는 워크플로우 마크다운을 파싱한다.
func ParseWorkflow(content string) (*WorkflowDoc, error) {
	doc := &WorkflowDoc{}

	if content == "" {
		return doc, nil
	}

	lines := strings.Split(content, "\n")
	var currentSection string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 섹션 헤더 감지
		if strings.HasPrefix(trimmed, "## ") {
			section := strings.TrimPrefix(trimmed, "## ")
			currentSection = strings.ToLower(strings.TrimSpace(section))
			continue
		}

		// Phase 헤더 감지 (### Phase N: Name)
		if strings.HasPrefix(trimmed, "### Phase ") {
			phase := parsePhaseHeader(trimmed, i, lines)
			doc.Phases = append(doc.Phases, phase)
			continue
		}

		// Policies 섹션 내 항목 파싱
		if currentSection == "policies" && strings.HasPrefix(trimmed, "- ") {
			policy := Policy{Description: strings.TrimPrefix(trimmed, "- ")}
			doc.Policies = append(doc.Policies, policy)
		}
	}

	return doc, nil
}

// parsePhaseHeader는 Phase 헤더에서 Phase 정보를 추출한다.
func parsePhaseHeader(header string, lineIdx int, lines []string) Phase {
	// "### Phase N: Name" 형식에서 이름 추출
	name := header
	if idx := strings.Index(header, ": "); idx >= 0 {
		name = header[idx+2:]
	}

	// 다음 줄에서 설명 추출
	var desc string
	if lineIdx+1 < len(lines) {
		next := strings.TrimSpace(lines[lineIdx+1])
		if next != "" && !strings.HasPrefix(next, "#") {
			desc = next
		}
	}

	return Phase{Name: strings.TrimSpace(name), Description: desc}
}

// GenerateWorkflow는 HarnessConfig에서 WORKFLOW.md를 생성한다.
func GenerateWorkflow(cfg *config.HarnessConfig) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("설정이 nil입니다")
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s Workflow\n\n", cfg.ProjectName))
	sb.WriteString("<!-- Autopus-ADK 워크플로우 문서 -->\n\n")

	// 정책 섹션
	sb.WriteString("## Policies\n\n")
	sb.WriteString("- 모든 코드 변경은 리뷰를 거쳐야 합니다\n")

	if cfg.Methodology.Enforce {
		sb.WriteString(fmt.Sprintf("- %s 방법론을 강제합니다\n", cfg.Methodology.Mode))
	}
	if cfg.Hooks.PreCommitLore {
		sb.WriteString("- Lore 커밋 메시지 형식을 준수해야 합니다\n")
	}
	if cfg.Hooks.PreCommitArch {
		sb.WriteString("- 아키텍처 규칙을 준수해야 합니다\n")
	}
	if cfg.Methodology.ReviewGate {
		sb.WriteString("- 각 단계 완료 후 리뷰 게이트를 통과해야 합니다\n")
	}
	sb.WriteString("\n")

	// 방법론별 단계 정의
	sb.WriteString("## Phases\n\n")
	switch strings.ToLower(cfg.Methodology.Mode) {
	case "tdd":
		sb.WriteString("### Phase 1: Red\n테스트 먼저 작성합니다.\n\n")
		sb.WriteString("### Phase 2: Green\n최소 구현으로 테스트를 통과시킵니다.\n\n")
		sb.WriteString("### Phase 3: Refactor\n코드 품질을 개선합니다.\n\n")
	case "ddd":
		sb.WriteString("### Phase 1: Analyze\n기존 동작을 분석합니다.\n\n")
		sb.WriteString("### Phase 2: Preserve\n기존 동작을 보존합니다.\n\n")
		sb.WriteString("### Phase 3: Improve\n점진적으로 개선합니다.\n\n")
	default:
		sb.WriteString("### Phase 1: Planning\n기능을 기획합니다.\n\n")
		sb.WriteString("### Phase 2: Implementation\n구현합니다.\n\n")
		sb.WriteString("### Phase 3: Review\n결과를 검토합니다.\n\n")
	}

	// 플랫폼 정보
	if len(cfg.Platforms) > 0 {
		sb.WriteString("## Supported Platforms\n\n")
		for _, p := range cfg.Platforms {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
