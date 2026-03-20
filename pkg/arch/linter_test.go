package arch_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insajin/autopus-adk/pkg/arch"
)

func TestLint_NoViolations(t *testing.T) {
	t.Parallel()

	archMap := &arch.ArchitectureMap{
		Dependencies: []arch.Dependency{
			{From: "cmd", To: "pkg", Type: "import"},
			{From: "pkg", To: "pkg", Type: "import"},
		},
	}

	rules := []arch.LintRule{
		{
			Name:        "no-pkg-to-internal",
			FromLayer:   "pkg",
			ToLayer:     "internal",
			Allowed:     false,
			Remediation: "pkg 레이어는 internal에 의존할 수 없습니다",
		},
	}

	violations := arch.Lint(archMap, rules)
	assert.Empty(t, violations)
}

func TestLint_WithViolation(t *testing.T) {
	t.Parallel()

	archMap := &arch.ArchitectureMap{
		Dependencies: []arch.Dependency{
			{From: "pkg/service", To: "internal/repo", Type: "import"},
		},
	}

	rules := []arch.LintRule{
		{
			Name:        "no-pkg-to-internal",
			FromLayer:   "pkg",
			ToLayer:     "internal",
			Allowed:     false,
			Remediation: "pkg 레이어는 internal에 의존할 수 없습니다. 인터페이스를 정의하세요.",
		},
	}

	violations := arch.Lint(archMap, rules)
	assert.Len(t, violations, 1)
	assert.Equal(t, "no-pkg-to-internal", violations[0].Rule)
	assert.Equal(t, "pkg/service", violations[0].From)
	assert.Equal(t, "internal/repo", violations[0].To)
	assert.NotEmpty(t, violations[0].Remediation)
}

func TestLint_AllowedDependency(t *testing.T) {
	t.Parallel()

	archMap := &arch.ArchitectureMap{
		Dependencies: []arch.Dependency{
			{From: "cmd/app", To: "pkg/service", Type: "import"},
		},
	}

	rules := []arch.LintRule{
		{
			Name:      "cmd-can-use-pkg",
			FromLayer: "cmd",
			ToLayer:   "pkg",
			Allowed:   true,
		},
	}

	violations := arch.Lint(archMap, rules)
	assert.Empty(t, violations)
}

func TestLint_MultipleViolations(t *testing.T) {
	t.Parallel()

	archMap := &arch.ArchitectureMap{
		Dependencies: []arch.Dependency{
			{From: "pkg/a", To: "internal/b", Type: "import"},
			{From: "pkg/c", To: "internal/d", Type: "import"},
		},
	}

	rules := []arch.LintRule{
		{
			Name:        "no-pkg-to-internal",
			FromLayer:   "pkg",
			ToLayer:     "internal",
			Allowed:     false,
			Remediation: "pkg는 internal에 의존할 수 없습니다",
		},
	}

	violations := arch.Lint(archMap, rules)
	assert.Len(t, violations, 2)
	for _, v := range violations {
		assert.NotEmpty(t, v.Remediation)
	}
}

func TestLint_EmptyRules(t *testing.T) {
	t.Parallel()

	archMap := &arch.ArchitectureMap{
		Dependencies: []arch.Dependency{
			{From: "pkg/a", To: "internal/b", Type: "import"},
		},
	}

	violations := arch.Lint(archMap, nil)
	assert.Empty(t, violations)
}
