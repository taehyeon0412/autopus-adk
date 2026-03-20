// Package content_test는 워크플로우 패키지의 테스트이다.
package content_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/content"
)

func TestParseWorkflow(t *testing.T) {
	t.Parallel()

	workflowContent := `# WORKFLOW

## Policies
- 테스트 없이 코드 작성 금지
- 리뷰 없이 병합 금지

## Phases

### Phase 1: Planning
기능 기획 단계

### Phase 2: Implementation
구현 단계
`
	doc, err := content.ParseWorkflow(workflowContent)
	require.NoError(t, err)
	assert.NotNil(t, doc)
	assert.NotEmpty(t, doc.Policies)
	assert.NotEmpty(t, doc.Phases)
}

func TestGenerateWorkflow(t *testing.T) {
	t.Parallel()

	cfg := &config.HarnessConfig{
		Mode:        config.ModeFull,
		ProjectName: "test-project",
		Platforms:   []string{"claude-code"},
		Methodology: config.MethodologyConf{
			Mode:       "tdd",
			Enforce:    true,
			ReviewGate: true,
		},
		Hooks: config.HooksConf{
			PreCommitLore: true,
			PreCommitArch: true,
		},
	}

	result, err := content.GenerateWorkflow(cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "test-project")
	assert.Contains(t, result, "tdd")
}

func TestParseWorkflow_InvalidContent(t *testing.T) {
	t.Parallel()

	// 빈 컨텐츠는 에러 없이 빈 doc 반환
	doc, err := content.ParseWorkflow("")
	require.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestWorkflowDoc_Phases(t *testing.T) {
	t.Parallel()

	workflowContent := `# WORKFLOW

## Phases

### Phase 1: Planning
기획

### Phase 2: Implementation
구현

### Phase 3: Review
리뷰
`
	doc, err := content.ParseWorkflow(workflowContent)
	require.NoError(t, err)
	assert.Len(t, doc.Phases, 3)
}
