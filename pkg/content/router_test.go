// Package content_test는 라우터 콘텐츠 패키지의 테스트이다.
package content_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/content"
)

func TestGenerateRoutingInstruction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      config.RouterConf
		contains []string
	}{
		{
			name: "기본 카테고리 라우팅",
			cfg: config.RouterConf{
				Strategy:   "category",
				Tiers:      map[string]string{"fast": "gemini-flash", "smart": "claude-sonnet", "ultra": "claude-opus"},
				Categories: map[string]string{"visual": "fast", "deep": "ultra", "quick": "fast"},
				IntentGate: true,
			},
			contains: []string{"visual", "deep", "quick"},
		},
		{
			name: "인텐트 게이트 포함",
			cfg: config.RouterConf{
				Strategy:   "category",
				Tiers:      map[string]string{"smart": "claude-sonnet"},
				Categories: map[string]string{"writing": "smart"},
				IntentGate: true,
			},
			contains: []string{"writing", "Intent"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := content.GenerateRoutingInstruction(tt.cfg)
			assert.NotEmpty(t, result)
			for _, s := range tt.contains {
				assert.Contains(t, result, s)
			}
		})
	}
}

func TestGenerateRoutingInstruction_AllCategories(t *testing.T) {
	t.Parallel()

	cfg := config.RouterConf{
		Strategy: "category",
		Tiers: map[string]string{
			"fast":  "gemini-flash",
			"smart": "claude-sonnet",
			"ultra": "claude-opus",
		},
		Categories: map[string]string{
			"visual":     "fast",
			"deep":       "ultra",
			"quick":      "fast",
			"ultrabrain": "ultra",
			"writing":    "smart",
			"git":        "fast",
			"adaptive":   "smart",
		},
		IntentGate: true,
	}

	result := content.GenerateRoutingInstruction(cfg)
	// 7개 카테고리 모두 포함
	categories := []string{"visual", "deep", "quick", "ultrabrain", "writing", "git", "adaptive"}
	for _, cat := range categories {
		assert.Contains(t, result, cat)
	}
}
