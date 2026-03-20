package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHarnessConfig_Validate_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cfg  HarnessConfig
	}{
		{
			name: "full mode",
			cfg: HarnessConfig{
				Mode:        ModeFull,
				ProjectName: "test-project",
				Platforms:   []string{"claude-code"},
			},
		},
		{
			name: "lite mode",
			cfg: HarnessConfig{
				Mode:        ModeLite,
				ProjectName: "test-project",
				Platforms:   []string{"claude-code", "codex"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.Validate()
			require.NoError(t, err)
		})
	}
}

func TestHarnessConfig_Validate_Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		cfg     HarnessConfig
		wantErr string
	}{
		{
			name:    "invalid mode",
			cfg:     HarnessConfig{Mode: "invalid", ProjectName: "p", Platforms: []string{"claude-code"}},
			wantErr: "invalid mode",
		},
		{
			name:    "empty project name",
			cfg:     HarnessConfig{Mode: ModeFull, ProjectName: "", Platforms: []string{"claude-code"}},
			wantErr: "project_name is required",
		},
		{
			name:    "no platforms",
			cfg:     HarnessConfig{Mode: ModeFull, ProjectName: "p", Platforms: []string{}},
			wantErr: "at least one platform",
		},
		{
			name:    "invalid platform",
			cfg:     HarnessConfig{Mode: ModeFull, ProjectName: "p", Platforms: []string{"invalid"}},
			wantErr: "invalid platform",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestHarnessConfig_ModeHelpers(t *testing.T) {
	t.Parallel()
	full := HarnessConfig{Mode: ModeFull}
	assert.True(t, full.IsFullMode())
	assert.False(t, full.IsLiteMode())

	lite := HarnessConfig{Mode: ModeLite}
	assert.False(t, lite.IsFullMode())
	assert.True(t, lite.IsLiteMode())
}
