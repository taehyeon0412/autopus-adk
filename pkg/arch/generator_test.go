package arch_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/arch"
)

func TestGenerate_BasicOutput(t *testing.T) {
	t.Parallel()

	archMap := &arch.ArchitectureMap{
		Domains: []arch.Domain{
			{Name: "service", Path: "pkg/service", Description: "서비스 레이어", Packages: []string{"user", "order"}},
			{Name: "model", Path: "pkg/model", Description: "모델 레이어", Packages: []string{"user"}},
		},
		Layers: []arch.Layer{
			{Name: "cmd", Level: 3, AllowedDeps: []string{"pkg", "internal"}},
			{Name: "pkg", Level: 2, AllowedDeps: []string{"pkg"}},
			{Name: "internal", Level: 1, AllowedDeps: []string{"pkg"}},
		},
		Dependencies: []arch.Dependency{
			{From: "cmd/app", To: "pkg/service", Type: "import"},
			{From: "pkg/service", To: "pkg/model", Type: "import"},
		},
		Violations: []arch.Violation{},
	}

	result, err := arch.Generate(archMap)
	require.NoError(t, err)

	// 필수 섹션 확인
	assert.Contains(t, result, "# Architecture")
	assert.Contains(t, result, "## Domains")
	assert.Contains(t, result, "## Layers")
	assert.Contains(t, result, "## Dependencies")
	assert.Contains(t, result, "service")
	assert.Contains(t, result, "model")
	assert.Contains(t, result, "cmd")
	assert.Contains(t, result, "pkg")
}

func TestGenerate_IncludesDiagram(t *testing.T) {
	t.Parallel()

	archMap := &arch.ArchitectureMap{
		Domains: []arch.Domain{
			{Name: "api", Path: "pkg/api", Description: "API 레이어"},
		},
		Layers: []arch.Layer{
			{Name: "pkg", Level: 2},
		},
		Dependencies: []arch.Dependency{
			{From: "cmd", To: "pkg/api", Type: "import"},
		},
	}

	result, err := arch.Generate(archMap)
	require.NoError(t, err)

	// 의존성 다이어그램 포함 확인
	assert.True(t, strings.Contains(result, "-->") || strings.Contains(result, "->"))
}

func TestGenerate_EmptyMap(t *testing.T) {
	t.Parallel()

	archMap := &arch.ArchitectureMap{}

	result, err := arch.Generate(archMap)
	require.NoError(t, err)
	assert.Contains(t, result, "# Architecture")
}

func TestGenerate_WithViolations(t *testing.T) {
	t.Parallel()

	archMap := &arch.ArchitectureMap{
		Violations: []arch.Violation{
			{
				Rule:        "layer-boundary",
				From:        "pkg/service",
				To:          "internal/repo",
				Message:     "pkg는 internal에 의존할 수 없습니다",
				Remediation: "internal 패키지를 pkg로 이동하거나 인터페이스를 정의하세요",
			},
		},
	}

	result, err := arch.Generate(archMap)
	require.NoError(t, err)
	assert.Contains(t, result, "Violation")
}
