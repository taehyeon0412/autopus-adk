package orchestra

import (
	"context"
	"fmt"
	"time"
)

// collectRoundHookResults collects hook-based results for a specific round.
// @AX:NOTE: [AUTO] magic constant 60s default timeout — per-provider wait; overridden by cfg.TimeoutSeconds
func collectRoundHookResults(ctx context.Context, cfg OrchestraConfig, session *HookSession, round int) []ProviderResponse {
	timeout := 60 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}

	var responses []ProviderResponse
	for _, p := range cfg.Providers {
		// Respect context cancellation between provider iterations.
		if ctx.Err() != nil {
			break
		}
		start := time.Now()
		err := session.WaitForDoneRound(timeout, p.Name, round)
		if err != nil {
			responses = append(responses, ProviderResponse{
				Provider: p.Name,
				Duration: time.Since(start),
				TimedOut: true,
			})
			continue
		}
		result, readErr := session.ReadResultRound(p.Name, round)
		output := ""
		if readErr == nil && result != nil {
			output = result.Output
		}
		responses = append(responses, ProviderResponse{
			Provider: p.Name,
			Output:   output,
			Duration: time.Since(start),
		})
	}
	return responses
}

// runJudgeRound executes the judge verdict after all debate rounds.
func runJudgeRound(ctx context.Context, cfg OrchestraConfig, panes []paneInfo, hookSession *HookSession, responses []ProviderResponse, lastRound int) *ProviderResponse {
	judgment := buildJudgmentPrompt(cfg.Prompt, responses)
	judgeCfg := findOrBuildJudgeConfig(cfg)

	// Try to find an existing pane for the judge (if judge is a participant).
	for _, pi := range panes {
		if pi.provider.Name == cfg.JudgeProvider && !pi.skipWait {
			patterns := DefaultCompletionPatterns()
			pollUntilPrompt(ctx, cfg.Terminal, pi.paneID, patterns, 10*time.Second)
			_ = cfg.Terminal.SendCommand(ctx, pi.paneID, judgment)
			time.Sleep(500 * time.Millisecond)
			_ = cfg.Terminal.SendCommand(ctx, pi.paneID, "\n")

			if cfg.HookMode && hookSession != nil {
				judgeRound := lastRound + 1
				resps := collectRoundHookResults(ctx, cfg, hookSession, judgeRound)
				for _, r := range resps {
					if r.Provider == cfg.JudgeProvider {
						r.Provider = cfg.JudgeProvider + " (judge)"
						return &r
					}
				}
			}
			return nil
		}
	}

	// Judge is not a participant — run as process.
	resp, err := runProvider(ctx, judgeCfg, judgment)
	if err != nil {
		return nil
	}
	resp.Provider = cfg.JudgeProvider + " (judge)"
	return resp
}

// consensusReached checks if all responses are substantially similar.
// Uses line-based consensus with 66% threshold.
func consensusReached(responses []ProviderResponse) bool {
	if len(responses) < 2 {
		return false
	}
	// @AX:NOTE: [AUTO] magic constant 0.66 — consensus threshold (66%); lowering increases false-positive consensus
	_, summary := MergeConsensus(responses, 0.66)
	n := countNonEmpty(responses)
	return summary == fmt.Sprintf("합의율: %d/%d (100%%)", n, n)
}

// countNonEmpty counts responses with non-empty output.
func countNonEmpty(responses []ProviderResponse) int {
	n := 0
	for _, r := range responses {
		if r.Output != "" {
			n++
		}
	}
	return n
}

// perRoundTimeout calculates the timeout for each debate round.
func perRoundTimeout(totalSeconds, rounds int) time.Duration {
	if totalSeconds <= 0 {
		totalSeconds = 120
	}
	if rounds <= 0 {
		rounds = 1
	}
	return time.Duration(totalSeconds/rounds) * time.Second
}

// buildDebateResult constructs the final OrchestraResult from debate rounds.
func buildDebateResult(cfg OrchestraConfig, responses []ProviderResponse, roundHistory [][]ProviderResponse, start time.Time) *OrchestraResult {
	merged, summary := mergeByStrategy(cfg.Strategy, responses, cfg)
	if merged == "" {
		merged = fmt.Sprintf("[interactive debate] %d rounds completed", len(roundHistory))
	}
	return &OrchestraResult{
		Strategy:     cfg.Strategy,
		Responses:    responses,
		Merged:       merged,
		Duration:     time.Since(start),
		Summary:      summary,
		RoundHistory: roundHistory,
	}
}

// mergeByStrategyWithRoundHistory creates an OrchestraResult from round history.
func mergeByStrategyWithRoundHistory(rounds [][]ProviderResponse, cfg OrchestraConfig) *OrchestraResult {
	var finalResponses []ProviderResponse
	if len(rounds) > 0 {
		finalResponses = rounds[len(rounds)-1]
	}
	merged, summary := mergeByStrategy(cfg.Strategy, finalResponses, cfg)
	return &OrchestraResult{
		Strategy:     cfg.Strategy,
		Responses:    finalResponses,
		Merged:       merged,
		Summary:      summary,
		RoundHistory: rounds,
	}
}
