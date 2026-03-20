package arch

import (
	"fmt"
	"strings"
)

// Lint는 ArchitectureMap에 대해 LintRule 목록을 검사하고 위반 목록을 반환한다.
func Lint(archMap *ArchitectureMap, rules []LintRule) []Violation {
	var violations []Violation

	for _, dep := range archMap.Dependencies {
		for _, rule := range rules {
			if rule.Allowed {
				// 허용 규칙은 위반 검사 불필요
				continue
			}
			fromLayer := extractLayer(dep.From)
			toLayer := extractLayer(dep.To)

			if fromLayer == rule.FromLayer && toLayer == rule.ToLayer {
				violations = append(violations, Violation{
					Rule:        rule.Name,
					From:        dep.From,
					To:          dep.To,
					Message:     fmt.Sprintf("%s 레이어에서 %s 레이어로의 의존은 규칙 '%s'를 위반합니다", fromLayer, toLayer, rule.Name),
					Remediation: rule.Remediation,
				})
			}
		}
	}

	return violations
}

// extractLayer는 패키지 경로에서 최상위 레이어명을 추출한다.
// 예: "pkg/service" -> "pkg", "internal/repo" -> "internal"
func extractLayer(pkg string) string {
	parts := strings.SplitN(pkg, "/", 2)
	return parts[0]
}
