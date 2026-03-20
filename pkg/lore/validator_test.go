package lore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insajin/autopus-adk/pkg/lore"
)

func TestValidate_ValidMessage(t *testing.T) {
	t.Parallel()

	config := lore.LoreConfig{
		RequiredTrailers: []string{"Constraint", "Confidence"},
	}

	commitMsg := "feat: 인증\n\nConstraint: stateless\nConfidence: high\n"
	errs := lore.Validate(commitMsg, config)
	assert.Empty(t, errs)
}

func TestValidate_MissingRequiredTrailer(t *testing.T) {
	t.Parallel()

	config := lore.LoreConfig{
		RequiredTrailers: []string{"Constraint", "Confidence"},
	}

	commitMsg := "feat: 인증\n\nConstraint: stateless\n"
	errs := lore.Validate(commitMsg, config)
	assert.Len(t, errs, 1)
	assert.Equal(t, "Confidence", errs[0].Field)
}

func TestValidate_InvalidConfidence(t *testing.T) {
	t.Parallel()

	config := lore.LoreConfig{}
	commitMsg := "feat: 인증\n\nConfidence: very-high\n"
	errs := lore.Validate(commitMsg, config)
	assert.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.Field == "Confidence" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestValidate_InvalidScopeRisk(t *testing.T) {
	t.Parallel()

	config := lore.LoreConfig{}
	commitMsg := "feat: 테스트\n\nScope-risk: global\n"
	errs := lore.Validate(commitMsg, config)
	assert.NotEmpty(t, errs)
}

func TestValidate_InvalidReversibility(t *testing.T) {
	t.Parallel()

	config := lore.LoreConfig{}
	commitMsg := "feat: 테스트\n\nReversibility: impossible\n"
	errs := lore.Validate(commitMsg, config)
	assert.NotEmpty(t, errs)
}

func TestValidate_EmptyConfig(t *testing.T) {
	t.Parallel()

	config := lore.LoreConfig{}
	commitMsg := "feat: 간단한 변경\n"
	errs := lore.Validate(commitMsg, config)
	assert.Empty(t, errs)
}

func TestValidate_AllRequiredMissing(t *testing.T) {
	t.Parallel()

	config := lore.LoreConfig{
		RequiredTrailers: []string{"Constraint", "Confidence", "Directive"},
	}

	commitMsg := "feat: 변경\n"
	errs := lore.Validate(commitMsg, config)
	assert.Len(t, errs, 3)
}
