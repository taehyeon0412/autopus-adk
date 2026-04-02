package qa

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStage is a test double for Stage.
type mockStage struct {
	name   string
	result *StageResult
	err    error
}

func (m *mockStage) Name() string { return m.name }
func (m *mockStage) Run(_ context.Context, _ string) (*StageResult, error) {
	return m.result, m.err
}

func TestPipeline_AllStagesPass(t *testing.T) {
	t.Parallel()

	stages := []Stage{
		&mockStage{name: "build", result: &StageResult{Name: "build", Status: "pass", Output: "ok"}},
		&mockStage{name: "test", result: &StageResult{Name: "test", Status: "pass", Output: "ok"}},
	}

	p := NewPipeline(t.TempDir(), stages)
	result, err := p.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
	assert.Len(t, result.Stages, 2)
	assert.Equal(t, "pass", result.Stages[0].Status)
	assert.Equal(t, "pass", result.Stages[1].Status)
}

func TestPipeline_EarlyFailureSkipsRemaining(t *testing.T) {
	t.Parallel()

	stages := []Stage{
		&mockStage{name: "build", result: &StageResult{Name: "build", Status: "fail", Output: "error"}, err: assert.AnError},
		&mockStage{name: "test", result: &StageResult{Name: "test", Status: "pass", Output: "ok"}},
	}

	p := NewPipeline(t.TempDir(), stages)
	result, err := p.Run(context.Background())

	require.Error(t, err)
	assert.Equal(t, "fail", result.Status)
	assert.Len(t, result.Stages, 2)
	assert.Equal(t, "fail", result.Stages[0].Status)
	assert.Equal(t, "skip", result.Stages[1].Status)
}

func TestPipeline_CleanupAlwaysRuns(t *testing.T) {
	t.Parallel()

	cleanupRan := false
	stages := []Stage{
		&mockStage{name: "build", result: &StageResult{Name: "build", Status: "fail", Output: "err"}, err: assert.AnError},
		&CleanupStage{Commands: []string{"go version"}},
	}

	// Override cleanup to track execution via a wrapper.
	// Use the actual CleanupStage but verify it ran via pipeline result.
	p := NewPipeline(t.TempDir(), stages)
	result, _ := p.Run(context.Background())

	// Cleanup stage should appear in results even though build failed.
	for _, sr := range result.Stages {
		if sr.Name == "cleanup" {
			cleanupRan = true
		}
	}
	assert.True(t, cleanupRan, "cleanup stage should always run")
}

func TestPipeline_SecretMasking(t *testing.T) {
	t.Parallel()

	stages := []Stage{
		&mockStage{
			name: "leak",
			result: &StageResult{
				Name:   "leak",
				Status: "pass",
				Output: "token: sk-abcdefghijklmnopqrstuvwxyz1234567890",
			},
		},
	}

	p := NewPipeline(t.TempDir(), stages)
	result, err := p.Run(context.Background())

	require.NoError(t, err)
	assert.NotContains(t, result.Stages[0].Output, "sk-abcdefghijklmnopqrstuvwxyz1234567890")
	assert.Contains(t, result.Stages[0].Output, "***REDACTED***")
}

func TestPipeline_EmptyStages(t *testing.T) {
	t.Parallel()

	p := NewPipeline(t.TempDir(), nil)
	result, err := p.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
	assert.Empty(t, result.Stages)
}
