package orchestra

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// failProvider returns a ProviderConfig that exits with code 1 (/usr/bin/false on macOS/Linux).
// It passes detect.IsInstalled checks since the binary exists.
func failProvider(name string) ProviderConfig {
	if runtime.GOOS == "windows" {
		// On Windows, use a command that exits non-zero
		return ProviderConfig{Name: name, Binary: "cmd", Args: []string{"/c", "exit 1"}, PromptViaArgs: true}
	}
	return ProviderConfig{Name: name, Binary: "/usr/bin/false", Args: []string{}, PromptViaArgs: true}
}

func TestRunParallel_PartialFailure(t *testing.T) {
	// Not parallel: modifies provider execution results based on order
	if runtime.GOOS == "windows" {
		t.Skip("Windows에서 건너뜁니다")
	}

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			echoProvider("good1"),
			failProvider("bad1"),
			echoProvider("good2"),
		},
		Strategy:       StrategyConsensus,
		Prompt:         "partial failure test",
		TimeoutSeconds: 10,
	}

	responses, failed, err := runParallel(context.Background(), cfg)

	require.NoError(t, err)
	assert.Len(t, responses, 2, "should have 2 successful responses")
	assert.Len(t, failed, 1, "should have 1 failed provider")
	assert.Equal(t, "bad1", failed[0].Name)
	assert.NotEmpty(t, failed[0].Error)
}

func TestRunParallel_AllFail(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows에서 건너뜁니다")
	}

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			failProvider("fail1"),
			failProvider("fail2"),
		},
		Strategy:       StrategyConsensus,
		Prompt:         "all fail test",
		TimeoutSeconds: 10,
	}

	responses, failed, err := runParallel(context.Background(), cfg)

	assert.Error(t, err, "should return error when all fail")
	assert.Nil(t, responses)
	assert.Len(t, failed, 2)
}

func TestRunOrchestra_GracefulDegradation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows에서 건너뜁니다")
	}

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			echoProvider("ok1"),
			failProvider("broken"),
		},
		Strategy:       StrategyConsensus,
		Prompt:         "graceful degradation test",
		TimeoutSeconds: 10,
	}

	result, err := RunOrchestra(context.Background(), cfg)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.FailedProviders, 1)
	assert.Equal(t, "broken", result.FailedProviders[0].Name)
	assert.Contains(t, result.Summary, "실패")
	assert.Contains(t, result.Summary, "broken")
}
