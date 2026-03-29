package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insajin/autopus-adk/pkg/config"
)

// --- SPEC-ORCHCFG-001 Phase 1.5: Threshold Test Scaffolds ---

// R1: CommandEntry must have ConsensusThreshold field
func TestCommandEntry_ConsensusThresholdField(t *testing.T) {
	t.Parallel()

	entry := config.CommandEntry{
		Strategy:           "consensus",
		Providers:          []string{"claude", "gemini"},
		ConsensusThreshold: 0.8,
	}
	assert.Equal(t, 0.8, entry.ConsensusThreshold)
}

// R6 + R2: resolveThreshold 4-level fallback (table-driven)
func TestResolveThreshold_Fallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		conf          *config.OrchestraConf
		commandName   string
		flagThreshold float64
		expected      float64
	}{
		{
			name: "flag overrides all",
			conf: &config.OrchestraConf{
				ConsensusThreshold: 0.5,
				Commands: map[string]config.CommandEntry{
					"review": {ConsensusThreshold: 0.7},
				},
			},
			commandName:   "review",
			flagThreshold: 0.9,
			expected:      0.9,
		},
		{
			name: "command config when no flag",
			conf: &config.OrchestraConf{
				ConsensusThreshold: 0.5,
				Commands: map[string]config.CommandEntry{
					"review": {ConsensusThreshold: 0.7},
				},
			},
			commandName:   "review",
			flagThreshold: 0,
			expected:      0.7,
		},
		{
			name: "global config when no flag and no command",
			conf: &config.OrchestraConf{
				ConsensusThreshold: 0.5,
				Commands:           map[string]config.CommandEntry{},
			},
			commandName:   "review",
			flagThreshold: 0,
			expected:      0.5,
		},
		{
			name: "default 0.66 when nothing set",
			conf: &config.OrchestraConf{
				Commands: map[string]config.CommandEntry{},
			},
			commandName:   "review",
			flagThreshold: 0,
			expected:      0.66,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := resolveThreshold(tt.conf, tt.commandName, tt.flagThreshold)
			assert.InDelta(t, tt.expected, got, 0.001)
		})
	}
}

// R5: range validation (table-driven)
func TestResolveThreshold_RangeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		threshold float64
		wantErr   bool
	}{
		{"negative value", -0.1, true},
		{"above 1.0", 1.1, true},
		{"zero is valid", 0.0, false},
		{"mid-range valid", 0.5, false},
		{"exactly 1.0", 1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateThreshold(tt.threshold)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
