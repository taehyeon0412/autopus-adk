package a2a

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// mustMarshal is a test-only helper that panics on marshal failure.
func mustMarshal(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("mustMarshal: %v", err))
	}
	return data
}

func mustSignPolicy(t *testing.T, taskID string, policy SecurityPolicy, secret string) string {
	t.Helper()
	signature, err := signSecurityPolicy(taskID, policy, secret)
	require.NoError(t, err)
	return signature
}

func mustSignControlPlane(t *testing.T, taskID, model string, pipelinePhases []string, pipelineInstructions map[string]string, pipelinePromptTemplates map[string]string, iterationBudget *IterationBudget, capabilities []string, secret string) string {
	t.Helper()
	signature, err := signControlPlane(taskID, model, pipelinePhases, pipelineInstructions, pipelinePromptTemplates, iterationBudget, capabilities, secret)
	require.NoError(t, err)
	return signature
}
