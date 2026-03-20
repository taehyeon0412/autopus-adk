package content

import (
	"fmt"
	"strings"

	"github.com/insajin/autopus-adk/pkg/config"
)

// 라우팅 카테고리 목록 (7개)
var routingCategories = []string{
	"visual",     // 시각적 작업 (UI, 다이어그램)
	"deep",       // 심층 분석 (아키텍처, 리팩토링)
	"quick",      // 빠른 작업 (자동완성, 간단한 수정)
	"ultrabrain", // 초고성능 추론 (복잡한 알고리즘)
	"writing",    // 문서 작성 (README, 커밋 메시지)
	"git",        // Git 작업 (커밋, PR)
	"adaptive",   // 컨텍스트 기반 적응 라우팅
}

// GenerateRoutingInstruction은 RouterConf에서 라우팅 지침 텍스트를 생성한다.
func GenerateRoutingInstruction(cfg config.RouterConf) string {
	var sb strings.Builder

	sb.WriteString("# Model Routing Instructions\n\n")
	sb.WriteString(fmt.Sprintf("Strategy: **%s**\n\n", cfg.Strategy))

	// 티어 정의
	if len(cfg.Tiers) > 0 {
		sb.WriteString("## Model Tiers\n\n")
		for tier, model := range cfg.Tiers {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", tier, model))
		}
		sb.WriteString("\n")
	}

	// 카테고리 라우팅
	if len(cfg.Categories) > 0 {
		sb.WriteString("## Category Routing\n\n")
		sb.WriteString("작업 유형에 따라 최적의 모델을 자동으로 선택합니다:\n\n")

		// 정렬된 순서로 출력 (7개 카테고리 우선)
		for _, cat := range routingCategories {
			tier, ok := cfg.Categories[cat]
			if !ok {
				continue
			}
			model := tierToModel(cfg.Tiers, tier)
			sb.WriteString(fmt.Sprintf("- **%s** → `%s` (%s)\n", cat, tier, model))
		}

		// 추가 카테고리
		for cat, tier := range cfg.Categories {
			if !isKnownCategory(cat) {
				model := tierToModel(cfg.Tiers, tier)
				sb.WriteString(fmt.Sprintf("- **%s** → `%s` (%s)\n", cat, tier, model))
			}
		}
		sb.WriteString("\n")
	}

	// 인텐트 게이트
	if cfg.IntentGate {
		sb.WriteString("## Intent Gate\n\n")
		sb.WriteString("사용자 요청을 분석하여 작업 유형을 자동 감지하고 적절한 모델로 라우팅합니다.\n\n")
		sb.WriteString("Intent 감지 패턴:\n")
		sb.WriteString("- 이미지/UI 관련 요청 → visual\n")
		sb.WriteString("- 아키텍처/설계 요청 → deep\n")
		sb.WriteString("- 빠른 수정 요청 → quick\n")
		sb.WriteString("- 복잡한 알고리즘 → ultrabrain\n")
		sb.WriteString("- 문서 작성 요청 → writing\n")
		sb.WriteString("- Git 관련 요청 → git\n")
		sb.WriteString("- 기타 → adaptive\n")
	}

	return sb.String()
}

// tierToModel은 티어 이름으로 모델명을 반환한다.
func tierToModel(tiers map[string]string, tier string) string {
	if model, ok := tiers[tier]; ok {
		return model
	}
	return tier
}

// isKnownCategory는 알려진 카테고리인지 확인한다.
func isKnownCategory(cat string) bool {
	for _, known := range routingCategories {
		if known == cat {
			return true
		}
	}
	return false
}
