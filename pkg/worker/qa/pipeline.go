package qa

import (
	"context"
	"fmt"
	"time"

	"github.com/insajin/autopus-adk/pkg/worker/security"
)

// PipelineResult holds the aggregated results of all pipeline stages.
type PipelineResult struct {
	Stages   []*StageResult `json:"stages"`
	Status   string         `json:"status"` // "pass" or "fail"
	Duration time.Duration  `json:"duration"`
}

// Pipeline orchestrates sequential execution of QA stages.
type Pipeline struct {
	stages  []Stage
	workDir string
	scanner *security.SecretScanner
}

// NewPipeline creates a pipeline that runs stages in order within workDir.
// Output from all stages is masked using the default SecretScanner.
func NewPipeline(workDir string, stages []Stage) *Pipeline {
	return &Pipeline{
		stages:  stages,
		workDir: workDir,
		scanner: security.NewSecretScanner(),
	}
}

// Run executes each stage sequentially. If a non-cleanup stage fails,
// remaining non-cleanup stages are skipped. Cleanup stages always run.
func (p *Pipeline) Run(ctx context.Context) (*PipelineResult, error) {
	start := time.Now()
	result := &PipelineResult{Status: "pass"}
	var firstErr error
	failed := false

	// Separate cleanup stages from regular stages.
	var regular []Stage
	var cleanup []Stage
	for _, s := range p.stages {
		if _, ok := s.(*CleanupStage); ok {
			cleanup = append(cleanup, s)
		} else {
			regular = append(regular, s)
		}
	}

	// Run regular stages; stop on first failure.
	for _, stage := range regular {
		if failed {
			result.Stages = append(result.Stages, &StageResult{
				Name:   stage.Name(),
				Status: "skip",
				Output: "skipped due to prior failure",
			})
			continue
		}

		sr, err := stage.Run(ctx, p.workDir)
		if sr != nil {
			sr.Output = p.scanner.Scan(sr.Output)
		}
		if err != nil {
			failed = true
			result.Status = "fail"
			if firstErr == nil {
				firstErr = fmt.Errorf("stage %q failed: %w", stage.Name(), err)
			}
		}
		if sr != nil {
			result.Stages = append(result.Stages, sr)
		}
	}

	// Cleanup stages always run regardless of failure.
	for _, stage := range cleanup {
		sr, err := stage.Run(ctx, p.workDir)
		if sr != nil {
			sr.Output = p.scanner.Scan(sr.Output)
			result.Stages = append(result.Stages, sr)
		}
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("cleanup stage %q failed: %w", stage.Name(), err)
		}
	}

	result.Duration = time.Since(start)
	return result, firstErr
}
