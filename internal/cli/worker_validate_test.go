package cli

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/worker/a2a"
	"github.com/insajin/autopus-adk/pkg/worker/security"
)

func TestRunWorkerValidate_VerifiesPolicySignature(t *testing.T) {
	t.Setenv(a2a.PolicySigningSecretEnv, "secret")

	dir := t.TempDir()
	policyPath := filepath.Join(dir, "autopus-policy-task-1.json")
	policy := security.SecurityPolicy{
		AllowFS:         true,
		AllowedCommands: []string{"go test"},
		TimeoutSec:      30,
	}

	data, err := json.Marshal(policy)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(policyPath, data, 0o600))

	signature, err := a2aSignPolicyForTest("task-1", policy, "secret")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(policyPath+".sig", []byte(signature+"\n"), 0o600))

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	err = runWorkerValidate(cmd, policyPath, "go test ./...", "")
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "PASS")
}

func a2aSignPolicyForTest(taskID string, policy security.SecurityPolicy, secret string) (string, error) {
	payload := struct {
		TaskID string                  `json:"task_id"`
		Policy security.SecurityPolicy `json:"policy"`
	}{
		TaskID: taskID,
		Policy: policy,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}
