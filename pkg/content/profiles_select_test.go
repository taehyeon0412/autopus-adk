package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResolveExtendsProfile tests R4: extends merging and S6: nextjs extends typescript.
func TestResolveExtendsProfile(t *testing.T) {
	t.Parallel()

	base := ProfileDefinition{
		Name:          "typescript",
		Stack:         "typescript",
		Tools:         []string{"tsc", "eslint"},
		TestFramework: "jest",
		Linter:        "eslint",
		Instructions:  "Use TypeScript strict mode.",
	}

	child := ProfileDefinition{
		Name:         "nextjs",
		Stack:        "typescript",
		Framework:    "nextjs",
		Extends:      "typescript",
		Tools:        []string{"next"},
		Instructions: "Use Next.js conventions.",
	}

	profiles := map[string]ProfileDefinition{
		"typescript": base,
		"nextjs":     child,
	}

	resolved, err := ResolveExtendsProfile(child, profiles)
	require.NoError(t, err)
	// Instructions from parent should be prepended
	assert.Contains(t, resolved.Instructions, "Use TypeScript strict mode.")
	assert.Contains(t, resolved.Instructions, "Use Next.js conventions.")
	// Child's own fields take precedence
	assert.Equal(t, "nextjs", resolved.Name)
	assert.Equal(t, "nextjs", resolved.Framework)
}

// TestResolveExtendsProfile_NoExtends verifies profiles without extends are returned as-is.
func TestResolveExtendsProfile_NoExtends(t *testing.T) {
	t.Parallel()

	p := ProfileDefinition{
		Name:         "go",
		Stack:        "go",
		Instructions: "Go instructions.",
	}

	resolved, err := ResolveExtendsProfile(p, map[string]ProfileDefinition{})
	require.NoError(t, err)
	assert.Equal(t, p.Instructions, resolved.Instructions)
}

// TestResolveExtendsProfile_BaseNotFound tests error when extends references a missing profile.
func TestResolveExtendsProfile_BaseNotFound(t *testing.T) {
	t.Parallel()

	child := ProfileDefinition{
		Name:         "nextjs",
		Stack:        "typescript",
		Extends:      "typescript",
		Instructions: "Next.js instructions.",
	}

	// Empty map — base profile "typescript" is missing
	_, err := ResolveExtendsProfile(child, map[string]ProfileDefinition{})
	assert.Error(t, err, "missing base profile must return an error")
	assert.Contains(t, err.Error(), "typescript")
}

// TestSelectProfile_PriorityFrameworkOverLanguage tests R3: framework > language > none.
func TestSelectProfile_PriorityFrameworkOverLanguage(t *testing.T) {
	t.Parallel()

	profiles := []ProfileDefinition{
		{Name: "go", Stack: "go"},
		{Name: "nextjs", Stack: "typescript", Framework: "nextjs"},
		{Name: "typescript", Stack: "typescript"},
	}

	tests := []struct {
		name      string
		stack     string
		framework string
		wantName  string
	}{
		{"framework match wins over language", "typescript", "nextjs", "nextjs"},
		{"language match when no framework", "typescript", "", "typescript"},
		{"go language match", "go", "", "go"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := SelectProfile(profiles, tc.stack, tc.framework)
			require.True(t, ok)
			assert.Equal(t, tc.wantName, got.Name)
		})
	}
}

// TestSelectProfile_GracefulFallbackNoMatch tests R6, S9: no profile → graceful fallback.
func TestSelectProfile_GracefulFallbackNoMatch(t *testing.T) {
	t.Parallel()

	profiles := []ProfileDefinition{
		{Name: "go", Stack: "go"},
	}

	_, ok := SelectProfile(profiles, "rust", "")
	assert.False(t, ok, "no match must return ok=false without panic")
}

// TestSelectProfile_EmptyProfiles tests S9 edge case: empty profile list.
func TestSelectProfile_EmptyProfiles(t *testing.T) {
	t.Parallel()

	_, ok := SelectProfile([]ProfileDefinition{}, "go", "")
	assert.False(t, ok)
}

// TestSelectProfile_BothEmpty verifies false when both stack and framework are empty.
func TestSelectProfile_BothEmpty(t *testing.T) {
	t.Parallel()

	profiles := []ProfileDefinition{
		{Name: "go", Stack: "go"},
	}

	_, ok := SelectProfile(profiles, "", "")
	assert.False(t, ok, "empty stack and framework must return false")
}

// TestApplyCustomOverride_ShallowMerge tests R9, S11: override shallow merge (frontmatter fields only).
func TestApplyCustomOverride_ShallowMerge(t *testing.T) {
	t.Parallel()

	base := ProfileDefinition{
		Name:          "go",
		Stack:         "go",
		Tools:         []string{"gopls"},
		Linter:        "golangci-lint",
		TestFramework: "go test",
		Instructions:  "Base instructions.",
	}

	override := ProfileDefinition{
		Name:   "go",
		Stack:  "go",
		Linter: "staticcheck",
	}

	result := ApplyCustomOverride(base, override)
	// Overridden field
	assert.Equal(t, "staticcheck", result.Linter)
	// Non-overridden fields preserved from base
	assert.Equal(t, "go test", result.TestFramework)
	assert.Equal(t, []string{"gopls"}, result.Tools)
}

// TestApplyCustomOverride_CustomOverridesBuiltin tests S10: custom overrides builtin.
func TestApplyCustomOverride_CustomOverridesBuiltin(t *testing.T) {
	t.Parallel()

	builtin := ProfileDefinition{
		Name:         "python",
		Stack:        "python",
		Linter:       "pylint",
		Instructions: "Use Python 3.10+.",
	}

	custom := ProfileDefinition{
		Name:   "python",
		Stack:  "python",
		Linter: "ruff",
		Source: "custom",
	}

	result := ApplyCustomOverride(builtin, custom)
	assert.Equal(t, "ruff", result.Linter, "custom linter must override builtin")
}

// TestApplyCustomOverride_AllFields verifies override applies all overridable fields.
func TestApplyCustomOverride_AllFields(t *testing.T) {
	t.Parallel()

	base := ProfileDefinition{
		Name:          "base",
		Stack:         "go",
		Tools:         []string{"oldtool"},
		Linter:        "old-linter",
		TestFramework: "old-test",
		Instructions:  "Base instructions.",
		Source:        "builtin",
	}

	override := ProfileDefinition{
		Name:          "base",
		Stack:         "go",
		Tools:         []string{"newtool"},
		Linter:        "new-linter",
		TestFramework: "new-test",
		Source:        "custom",
	}

	result := ApplyCustomOverride(base, override)
	assert.Equal(t, []string{"newtool"}, result.Tools)
	assert.Equal(t, "new-linter", result.Linter)
	assert.Equal(t, "new-test", result.TestFramework)
	assert.Equal(t, "custom", result.Source)
	// Instructions never overridden
	assert.Equal(t, "Base instructions.", result.Instructions)
}

// TestProfilesConfDefaultOverride tests R7: autopus.yaml profiles.executor.default override.
func TestProfilesConfDefaultOverride(t *testing.T) {
	t.Parallel()

	profiles := []ProfileDefinition{
		{Name: "go", Stack: "go"},
		{Name: "typescript", Stack: "typescript"},
	}

	// When no stack detected, default profile name should be selected
	got, ok := SelectProfileWithConf(profiles, "", "", "go")
	require.True(t, ok, "default from conf must resolve to a valid profile")
	assert.Equal(t, "go", got.Name)
}

// TestSelectProfileWithConf_StackMatchSkipsDefault verifies stack match takes priority over default.
func TestSelectProfileWithConf_StackMatchSkipsDefault(t *testing.T) {
	t.Parallel()

	profiles := []ProfileDefinition{
		{Name: "go", Stack: "go"},
		{Name: "python", Stack: "python"},
	}

	// stack="python" must match python profile, not fall back to default "go"
	got, ok := SelectProfileWithConf(profiles, "python", "", "go")
	require.True(t, ok)
	assert.Equal(t, "python", got.Name)
}

// TestSelectProfileWithConf_NoMatchNoDefault verifies false when no stack match and no default.
func TestSelectProfileWithConf_NoMatchNoDefault(t *testing.T) {
	t.Parallel()

	profiles := []ProfileDefinition{
		{Name: "go", Stack: "go"},
	}

	_, ok := SelectProfileWithConf(profiles, "rust", "", "")
	assert.False(t, ok, "no match and no default must return false")
}

// TestSelectProfileWithConf_DefaultNotInList verifies false when default name is not in profile list.
func TestSelectProfileWithConf_DefaultNotInList(t *testing.T) {
	t.Parallel()

	profiles := []ProfileDefinition{
		{Name: "go", Stack: "go"},
	}

	_, ok := SelectProfileWithConf(profiles, "", "", "nonexistent")
	assert.False(t, ok, "default name not in profile list must return false")
}
