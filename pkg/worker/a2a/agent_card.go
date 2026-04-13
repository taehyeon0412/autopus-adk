package a2a

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
)

// providerSkills maps known provider names to their skill sets.
var providerSkills = map[string][]string{
	"claude":   {"coding", "analysis", "review"},
	"codex":    {"coding", "generation"},
	"gemini":   {"coding", "analysis", "search"},
	"opencode": {"coding"},
}

// defaultProviderSkills is used for unknown providers.
var defaultProviderSkills = []string{"coding"}

// CardBuilder constructs an AgentCard from worker configuration.
type CardBuilder struct {
	workerName string
	backendURL string
	providers  []string
	version    string
}

// NewCardBuilder creates a CardBuilder with the given worker name and backend URL.
func NewCardBuilder(name, backendURL string) *CardBuilder {
	return &CardBuilder{
		workerName: name,
		backendURL: backendURL,
	}
}

// WithProviders sets the provider list for skill resolution.
func (b *CardBuilder) WithProviders(providers []string) *CardBuilder {
	b.providers = providers
	return b
}

// WithVersion sets the worker version string.
func (b *CardBuilder) WithVersion(version string) *CardBuilder {
	b.version = version
	return b
}

// Build assembles an AgentCard with deduplicated skills from all providers.
func (b *CardBuilder) Build() AgentCard {
	skills := b.resolveSkills()

	description := "Autopus ADK Worker"
	if b.version != "" {
		description = fmt.Sprintf("Autopus ADK Worker v%s", b.version)
	}

	log.Printf("[a2a] built agent card: name=%s providers=%v skills=%v",
		b.workerName, b.providers, skills)

	return AgentCard{
		Name:                b.workerName,
		Description:         description,
		URL:                 b.backendURL,
		Skills:              skills,
		Capabilities:        DefaultCapabilities(),
		SupportedInputModes: []string{"text"},
	}
}

// resolveSkills collects and deduplicates skills from all providers.
func (b *CardBuilder) resolveSkills() []string {
	seen := make(map[string]struct{})
	var skills []string

	for _, provider := range b.providers {
		provSkills, ok := providerSkills[provider]
		if !ok {
			log.Printf("[a2a] unknown provider %q, using default skills", provider)
			provSkills = defaultProviderSkills
		}
		for _, s := range provSkills {
			if _, exists := seen[s]; !exists {
				seen[s] = struct{}{}
				skills = append(skills, s)
			}
		}
	}

	// Stable output for deterministic cards.
	sort.Strings(skills)
	return skills
}

// RegistrationResult holds the parsed response from agent/register.
type RegistrationResult struct {
	Success  bool   `json:"success"`
	WorkerID string `json:"worker_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ParseRegistrationResponse unmarshals a registration response payload.
func ParseRegistrationResponse(data []byte) (*RegistrationResult, error) {
	var result RegistrationResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse registration response: %w", err)
	}
	return &result, nil
}
