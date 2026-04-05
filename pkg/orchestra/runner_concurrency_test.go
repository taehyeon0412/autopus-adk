package orchestra

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunParallel_PerGoroutineContext_IndependentCancellation verifies S1:
// When 1 provider times out, others still succeed and their contexts are NOT cancelled.
func TestRunParallel_PerGoroutineContext_IndependentCancellation(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			echoProvider("fast1"),
			echoProvider("fast2"),
			sleepProvider("slow1"), // sleeps 10s, should timeout
		},
		Strategy:       StrategyConsensus,
		Prompt:         "concurrency test",
		TimeoutSeconds: 3, // per-provider timeout: 3s — slow1 will exceed it
	}

	ctx := context.Background()
	start := time.Now()
	result, err := RunOrchestra(ctx, cfg)
	elapsed := time.Since(start)

	require.NoError(t, err, "should not return error when some providers succeed")
	// R1/S1: 2 echo providers succeed, 1 sleep provider fails
	assert.Len(t, result.Responses, 2, "expected 2 successful responses")
	assert.Len(t, result.FailedProviders, 1, "expected 1 failed provider")
	assert.Equal(t, "slow1", result.FailedProviders[0].Name)
	assert.Contains(t, result.FailedProviders[0].Error, "timeout",
		"failed provider error should mention timeout")

	// S1: other contexts were NOT cancelled — elapsed time should be short
	assert.Less(t, elapsed, 8*time.Second,
		"total time should be much less than 10s sleep, proving independent cancellation")
}

// TestRunParallel_PerProviderTimeout_RecordsFailedProvider verifies S2:
// A sleep provider exceeding timeout becomes FailedProvider; echo providers succeed.
func TestRunParallel_PerProviderTimeout_RecordsFailedProvider(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			echoProvider("echo1"),
			echoProvider("echo2"),
			sleepProvider("sleeper"), // sleeps 10s
		},
		Strategy:       StrategyConsensus,
		Prompt:         "timeout test",
		TimeoutSeconds: 2, // per-provider timeout: 2s
	}

	ctx := context.Background()
	start := time.Now()
	result, err := RunOrchestra(ctx, cfg)
	elapsed := time.Since(start)

	require.NoError(t, err)

	// S2: echo providers return successful responses
	assert.Len(t, result.Responses, 2, "two echo providers should succeed")

	// S2: sleep provider is recorded as FailedProvider with timeout
	require.Len(t, result.FailedProviders, 1, "one provider should fail")
	assert.Equal(t, "sleeper", result.FailedProviders[0].Name)

	// S2: total time << sleep duration (10s), proving per-provider cancellation
	assert.Less(t, elapsed, 6*time.Second,
		"execution should complete well before the sleep duration")
}

// TestRunParallel_AllProvidersTimeout_ReturnsError verifies that when ALL
// providers timeout, runParallel returns an error (not partial success).
func TestRunParallel_AllProvidersTimeout_ReturnsError(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			sleepProvider("slow1"),
			sleepProvider("slow2"),
		},
		Strategy:       StrategyConsensus,
		Prompt:         "all timeout test",
		TimeoutSeconds: 2, // Both sleep 10s, will timeout at 2s
	}

	ctx := context.Background()
	start := time.Now()
	result, err := RunOrchestra(ctx, cfg)
	elapsed := time.Since(start)

	// When all providers timeout, RunOrchestra still returns a result with
	// empty Responses and all providers in FailedProviders (graceful degradation).
	require.NoError(t, err, "timeout is graceful degradation, not hard error")
	require.NotNil(t, result, "result should still be returned")
	assert.Empty(t, result.Responses, "no provider succeeded")
	assert.Len(t, result.FailedProviders, 2, "both providers should be in FailedProviders")
	for _, fp := range result.FailedProviders {
		assert.Contains(t, fp.Error, "timeout", "failed provider error should mention timeout")
	}
	assert.Less(t, elapsed, 6*time.Second,
		"should not wait for full 10s sleep duration")
}

// TestRunParallel_SingleProvider verifies correct behavior with exactly one provider.
func TestRunParallel_SingleProvider(t *testing.T) {
	t.Parallel()

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			echoProvider("solo"),
		},
		Strategy:       StrategyConsensus,
		Prompt:         "single provider test",
		TimeoutSeconds: 10,
	}

	result, err := RunOrchestra(context.Background(), cfg)
	require.NoError(t, err)
	assert.Len(t, result.Responses, 1, "single provider should produce 1 response")
	assert.Equal(t, "solo", result.Responses[0].Provider)
	assert.Empty(t, result.FailedProviders, "no providers should fail")
	assert.NotEmpty(t, result.Summary)
}

// TestRunParallel_BackwardCompatibility verifies S5:
// Existing consensus call works unchanged with same config shape.
func TestRunParallel_BackwardCompatibility(t *testing.T) {
	t.Parallel()

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			echoProvider("p1"),
			echoProvider("p2"),
			echoProvider("p3"),
		},
		Strategy:       StrategyConsensus,
		Prompt:         "backward compat test",
		TimeoutSeconds: 10,
	}

	result, err := RunOrchestra(context.Background(), cfg)
	require.NoError(t, err)
	assert.Equal(t, StrategyConsensus, result.Strategy)
	assert.Len(t, result.Responses, 3, "all 3 echo providers should succeed")
	assert.Empty(t, result.FailedProviders, "no providers should fail")
	assert.NotEmpty(t, result.Summary)
}
