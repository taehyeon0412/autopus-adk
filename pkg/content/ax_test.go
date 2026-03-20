// Package content_test는 MX 어노테이션 패키지의 테스트이다.
package content_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insajin/autopus-adk/pkg/content"
)

func TestGenerateAXInstruction(t *testing.T) {
	t.Parallel()

	result := content.GenerateAXInstruction()
	assert.NotEmpty(t, result)

	// 4가지 태그 타입 모두 포함
	tagTypes := []string{"@AX:NOTE", "@AX:WARN", "@AX:ANCHOR", "@AX:TODO"}
	for _, tag := range tagTypes {
		assert.Contains(t, result, tag)
	}
}

func TestGenerateAXInstruction_LifecycleRules(t *testing.T) {
	t.Parallel()

	result := content.GenerateAXInstruction()
	// 라이프사이클 규칙 포함
	assert.Contains(t, result, "ANCHOR")
	assert.Contains(t, result, "fan_in")
}

func TestGenerateAXInstruction_NotEmpty(t *testing.T) {
	t.Parallel()

	result := content.GenerateAXInstruction()
	// 최소 500자 이상의 의미있는 지침
	assert.Greater(t, len(result), 500)
}
