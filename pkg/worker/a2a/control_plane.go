package a2a

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	PolicySigningSecretEnv              = "AUTOPUS_A2A_POLICY_SIGNING_SECRET"
	CapabilityServerModelV1             = "server_model_v1"
	CapabilityPipelinePhasesV1          = "pipeline_phases_v1"
	CapabilityPipelineInstructionsV1    = "pipeline_instructions_v1"
	CapabilityPipelinePromptTemplatesV1 = "pipeline_prompt_templates_v1"
	CapabilityIterationBudgetV1         = "iteration_budget_v1"
	CapabilitySignedPolicyV1            = "signed_policy_v1"
	CapabilitySignedControlPlaneV1      = "signed_control_plane_v1"
)

var defaultCapabilities = []string{
	CapabilityServerModelV1,
	CapabilityPipelinePhasesV1,
	CapabilityPipelineInstructionsV1,
	CapabilityPipelinePromptTemplatesV1,
	CapabilityIterationBudgetV1,
	CapabilitySignedPolicyV1,
	CapabilitySignedControlPlaneV1,
}

// DefaultCapabilities returns the worker control-plane capabilities advertised at registration.
func DefaultCapabilities() []string {
	return append([]string(nil), defaultCapabilities...)
}

// SignedControlPlaneEnforced returns true when the worker is running in a mode
// where server-issued control-plane metadata must be trusted over local fallback.
func SignedControlPlaneEnforced() bool {
	return signingSecret() != ""
}

func signingSecret() string {
	return strings.TrimSpace(os.Getenv(PolicySigningSecretEnv))
}

func validateSecurityPolicySignature(taskID string, policy SecurityPolicy, signature string) error {
	secret := signingSecret()
	if secret == "" {
		return nil
	}
	if strings.TrimSpace(signature) == "" {
		return fmt.Errorf("missing policy signature")
	}
	return verifySecurityPolicySignature(taskID, policy, signature, secret)
}

func validateControlPlaneSignature(taskID, model string, pipelinePhases []string, pipelineInstructions map[string]string, pipelinePromptTemplates map[string]string, iterationBudget *IterationBudget, capabilities []string, signature string) error {
	secret := signingSecret()
	if secret == "" {
		return nil
	}
	if !hasControlPlaneMetadata(model, pipelinePhases, pipelineInstructions, pipelinePromptTemplates, iterationBudget) && len(capabilities) == 0 && strings.TrimSpace(signature) == "" {
		return nil
	}
	if len(capabilities) == 0 {
		return fmt.Errorf("missing control plane capabilities")
	}
	if strings.TrimSpace(signature) == "" {
		return fmt.Errorf("missing control plane signature")
	}
	return verifyControlPlaneSignature(taskID, model, pipelinePhases, pipelineInstructions, pipelinePromptTemplates, iterationBudget, capabilities, signature, secret)
}

func signSecurityPolicy(taskID string, policy any, secret string) (string, error) {
	payload, err := canonicalSecurityPolicyPayload(taskID, policy)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write(payload); err != nil {
		return "", fmt.Errorf("sign policy payload: %w", err)
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func verifySecurityPolicySignature(taskID string, policy any, signature, secret string) error {
	expected, err := signSecurityPolicy(taskID, policy, secret)
	if err != nil {
		return err
	}
	if !hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature))) {
		return fmt.Errorf("policy signature mismatch")
	}
	return nil
}

func signControlPlane(taskID, model string, pipelinePhases []string, pipelineInstructions map[string]string, pipelinePromptTemplates map[string]string, iterationBudget *IterationBudget, capabilities []string, secret string) (string, error) {
	payload, err := canonicalControlPlanePayload(taskID, model, pipelinePhases, pipelineInstructions, pipelinePromptTemplates, iterationBudget, capabilities)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write(payload); err != nil {
		return "", fmt.Errorf("sign control plane payload: %w", err)
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func verifyControlPlaneSignature(taskID, model string, pipelinePhases []string, pipelineInstructions map[string]string, pipelinePromptTemplates map[string]string, iterationBudget *IterationBudget, capabilities []string, signature, secret string) error {
	expected, err := signControlPlane(taskID, model, pipelinePhases, pipelineInstructions, pipelinePromptTemplates, iterationBudget, capabilities, secret)
	if err != nil {
		return err
	}
	if !hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature))) {
		return fmt.Errorf("control plane signature mismatch")
	}
	return nil
}

func canonicalSecurityPolicyPayload(taskID string, policy any) ([]byte, error) {
	payload := struct {
		TaskID string `json:"task_id"`
		Policy any    `json:"policy"`
	}{
		TaskID: taskID,
		Policy: policy,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal canonical policy payload: %w", err)
	}
	return data, nil
}

func canonicalControlPlanePayload(taskID, model string, pipelinePhases []string, pipelineInstructions map[string]string, pipelinePromptTemplates map[string]string, iterationBudget *IterationBudget, capabilities []string) ([]byte, error) {
	payload := struct {
		TaskID                  string            `json:"task_id"`
		Model                   string            `json:"model,omitempty"`
		PipelinePhases          []string          `json:"pipeline_phases,omitempty"`
		PipelineInstructions    map[string]string `json:"pipeline_instructions,omitempty"`
		PipelinePromptTemplates map[string]string `json:"pipeline_prompt_templates,omitempty"`
		IterationBudget         *IterationBudget  `json:"iteration_budget,omitempty"`
		Capabilities            []string          `json:"capabilities,omitempty"`
	}{
		TaskID:                  taskID,
		Model:                   strings.TrimSpace(model),
		PipelinePhases:          append([]string(nil), pipelinePhases...),
		PipelineInstructions:    cloneStringMap(pipelineInstructions),
		PipelinePromptTemplates: cloneStringMap(pipelinePromptTemplates),
		IterationBudget:         cloneIterationBudget(iterationBudget),
		Capabilities:            append([]string(nil), capabilities...),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal canonical control plane payload: %w", err)
	}
	return data, nil
}

func hasControlPlaneMetadata(model string, pipelinePhases []string, pipelineInstructions map[string]string, pipelinePromptTemplates map[string]string, iterationBudget *IterationBudget) bool {
	return strings.TrimSpace(model) != "" || len(pipelinePhases) > 0 || len(pipelineInstructions) > 0 || len(pipelinePromptTemplates) > 0 || hasIterationBudget(iterationBudget)
}

func applyControlPlaneCapabilities(model string, pipelinePhases []string, pipelineInstructions map[string]string, pipelinePromptTemplates map[string]string, iterationBudget *IterationBudget, capabilities []string) (string, []string, map[string]string, map[string]string, *IterationBudget) {
	if len(capabilities) == 0 {
		return strings.TrimSpace(model), append([]string(nil), pipelinePhases...), cloneStringMap(pipelineInstructions), cloneStringMap(pipelinePromptTemplates), cloneIterationBudget(iterationBudget)
	}

	var filteredModel string
	var filteredPhases []string
	var filteredInstructions map[string]string
	var filteredPromptTemplates map[string]string
	var filteredIterationBudget *IterationBudget

	if hasCapability(capabilities, CapabilityServerModelV1) {
		filteredModel = strings.TrimSpace(model)
	}
	if hasCapability(capabilities, CapabilityPipelinePhasesV1) {
		filteredPhases = append([]string(nil), pipelinePhases...)
	}
	if hasCapability(capabilities, CapabilityPipelineInstructionsV1) {
		filteredInstructions = cloneStringMap(pipelineInstructions)
	}
	if hasCapability(capabilities, CapabilityPipelinePromptTemplatesV1) {
		filteredPromptTemplates = cloneStringMap(pipelinePromptTemplates)
	}
	if hasCapability(capabilities, CapabilityIterationBudgetV1) {
		filteredIterationBudget = cloneIterationBudget(iterationBudget)
	}
	return filteredModel, filteredPhases, filteredInstructions, filteredPromptTemplates, filteredIterationBudget
}

func hasCapability(capabilities []string, target string) bool {
	for _, capability := range capabilities {
		if capability == target {
			return true
		}
	}
	return false
}

func hasIterationBudget(iterationBudget *IterationBudget) bool {
	return iterationBudget != nil && iterationBudget.Limit > 0
}

func cloneIterationBudget(iterationBudget *IterationBudget) *IterationBudget {
	if iterationBudget == nil {
		return nil
	}
	cloned := *iterationBudget
	return &cloned
}

func policySignaturePath(policyPath string) string {
	return policyPath + ".sig"
}

func writePolicySignature(policyPath, signature string) error {
	if strings.TrimSpace(signature) == "" {
		return nil
	}
	path := policySignaturePath(policyPath)
	if err := os.WriteFile(path, []byte(signature+"\n"), 0o600); err != nil {
		return fmt.Errorf("write policy signature: %w", err)
	}
	return nil
}

func readPolicySignature(policyPath string) (string, error) {
	data, err := os.ReadFile(policySignaturePath(policyPath))
	if err != nil {
		return "", fmt.Errorf("read policy signature: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func taskIDFromPolicyPath(policyPath string) (string, error) {
	base := filepath.Base(policyPath)
	const prefix = "autopus-policy-"
	const suffix = ".json"
	if !strings.HasPrefix(base, prefix) || !strings.HasSuffix(base, suffix) {
		return "", fmt.Errorf("unexpected policy filename: %s", base)
	}
	return strings.TrimSuffix(strings.TrimPrefix(base, prefix), suffix), nil
}

// VerifyCachedPolicyFile verifies the sidecar signature for a cached policy file when
// signature validation is enabled via AUTOPUS_A2A_POLICY_SIGNING_SECRET.
func VerifyCachedPolicyFile(policyPath string, policy any) error {
	secret := signingSecret()
	if secret == "" {
		return nil
	}
	taskID, err := taskIDFromPolicyPath(policyPath)
	if err != nil {
		return err
	}
	signature, err := readPolicySignature(policyPath)
	if err != nil {
		return err
	}
	return verifySecurityPolicySignature(taskID, policy, signature, secret)
}
