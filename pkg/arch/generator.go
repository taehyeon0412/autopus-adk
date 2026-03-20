package arch

import (
	"fmt"
	"strings"
)

// Generate는 ArchitectureMap을 ARCHITECTURE.md 마크다운으로 변환한다.
func Generate(archMap *ArchitectureMap) (string, error) {
	var sb strings.Builder

	sb.WriteString("# Architecture\n\n")
	sb.WriteString("이 문서는 자동 생성된 프로젝트 아키텍처 맵입니다.\n\n")

	// Domains 섹션
	sb.WriteString("## Domains\n\n")
	if len(archMap.Domains) == 0 {
		sb.WriteString("도메인이 없습니다.\n\n")
	} else {
		for _, d := range archMap.Domains {
			sb.WriteString(fmt.Sprintf("### %s\n\n", d.Name))
			sb.WriteString(fmt.Sprintf("- **경로**: `%s`\n", d.Path))
			sb.WriteString(fmt.Sprintf("- **설명**: %s\n", d.Description))
			if len(d.Packages) > 0 {
				sb.WriteString(fmt.Sprintf("- **패키지**: %s\n", strings.Join(d.Packages, ", ")))
			}
			sb.WriteString("\n")
		}
	}

	// Layers 섹션
	sb.WriteString("## Layers\n\n")
	if len(archMap.Layers) == 0 {
		sb.WriteString("레이어가 없습니다.\n\n")
	} else {
		sb.WriteString("| 레이어 | 레벨 | 허용 의존 |\n")
		sb.WriteString("|--------|------|----------|\n")
		for _, l := range archMap.Layers {
			allowedDeps := strings.Join(l.AllowedDeps, ", ")
			if allowedDeps == "" {
				allowedDeps = "-"
			}
			sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", l.Name, l.Level, allowedDeps))
		}
		sb.WriteString("\n")
	}

	// Dependencies 섹션
	sb.WriteString("## Dependencies\n\n")
	if len(archMap.Dependencies) == 0 {
		sb.WriteString("의존성이 없습니다.\n\n")
	} else {
		sb.WriteString("```\n")
		seen := map[string]bool{}
		for _, d := range archMap.Dependencies {
			key := d.From + " --> " + d.To
			if seen[key] {
				continue
			}
			seen[key] = true
			sb.WriteString(fmt.Sprintf("%s --> %s\n", d.From, d.To))
		}
		sb.WriteString("```\n\n")

		// 의존성 테이블
		sb.WriteString("| 출처 | 대상 | 유형 |\n")
		sb.WriteString("|------|------|------|\n")
		seenTable := map[string]bool{}
		for _, d := range archMap.Dependencies {
			key := d.From + "|" + d.To
			if seenTable[key] {
				continue
			}
			seenTable[key] = true
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", d.From, d.To, d.Type))
		}
		sb.WriteString("\n")
	}

	// Violations 섹션
	if len(archMap.Violations) > 0 {
		sb.WriteString("## Violations\n\n")
		sb.WriteString("> 아키텍처 규칙 위반 목록입니다.\n\n")
		for _, v := range archMap.Violations {
			sb.WriteString(fmt.Sprintf("### %s\n\n", v.Rule))
			sb.WriteString(fmt.Sprintf("- **From**: `%s`\n", v.From))
			sb.WriteString(fmt.Sprintf("- **To**: `%s`\n", v.To))
			sb.WriteString(fmt.Sprintf("- **메시지**: %s\n", v.Message))
			sb.WriteString(fmt.Sprintf("- **수정 방법**: %s\n", v.Remediation))
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}
