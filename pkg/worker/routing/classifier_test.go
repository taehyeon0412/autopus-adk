package routing

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestClassify(t *testing.T) {
	t.Parallel()

	defaultThresholds := ClassifierThresholds{
		SimpleMaxChars:  200,
		ComplexMinChars: 1000,
	}
	c := NewClassifier(defaultThresholds)

	tests := []struct {
		name           string
		message        string
		wantComplexity Complexity
		wantCodeBlocks bool
	}{
		{
			name:           "S3: short simple Korean message",
			message:        "현재 상태 확인해줘",
			wantComplexity: ComplexitySimple,
			wantCodeBlocks: false,
		},
		{
			name:           "S4: long message with code blocks and complex keywords",
			message:        strings.Repeat("x", 1200) + "\n```go\nfunc main() {}\n```\n리팩토링과 아키텍처 설계가 필요합니다",
			wantComplexity: ComplexityComplex,
			wantCodeBlocks: true,
		},
		{
			name:           "S11: medium length with medium keywords",
			message:        strings.Repeat("a", 500) + " 에러 처리 수정이 필요하고 변경 사항을 추가해주세요",
			wantComplexity: ComplexityMedium,
			wantCodeBlocks: false,
		},
		{
			name:           "edge: empty string",
			message:        "",
			wantComplexity: ComplexitySimple,
			wantCodeBlocks: false,
		},
		{
			name:           "edge: only code blocks no keywords",
			message:        strings.Repeat("z", 500) + "\n```\nsome code\n```\n",
			wantComplexity: ComplexityComplex, // code blocks (+1) push score above 0
			wantCodeBlocks: true,
		},
		{
			name:           "English simple keywords",
			message:        "check status list",
			wantComplexity: ComplexitySimple,
			wantCodeBlocks: false,
		},
		{
			name:           "Korean boundary: 70 runes (210 bytes) should be simple",
			message:        strings.Repeat("가", 70), // 70 runes, 210 bytes
			wantComplexity: ComplexitySimple,
			wantCodeBlocks: false,
		},
		{
			name:           "English complex keywords",
			message:        strings.Repeat("w", 1100) + " refactor the architecture and analyze the design",
			wantComplexity: ComplexityComplex,
			wantCodeBlocks: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			level, signals := c.Classify(tt.message)
			assert.Equal(t, tt.wantComplexity, level)
			assert.Equal(t, tt.wantCodeBlocks, signals.HasCodeBlocks)
			assert.Equal(t, utf8.RuneCountInString(tt.message), signals.CharCount)
		})
	}
}
