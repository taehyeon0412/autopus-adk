package e2e

import (
	"path/filepath"
	"testing"
)

func TestParseBuildLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		line     string
		expected []BuildEntry
	}{
		{
			name: "multi-build line with labels",
			line: "go build ./cmd/auto/ (ADK), go build ./cmd/server/ (Backend), npm run build (Frontend)",
			expected: []BuildEntry{
				{Command: "go build ./cmd/auto/", Label: "ADK", SubmodulePath: "autopus-adk"},
				{Command: "go build ./cmd/server/", Label: "Backend", SubmodulePath: "Autopus"},
				{Command: "npm run build", Label: "Frontend", SubmodulePath: "Autopus/frontend"},
			},
		},
		{
			name: "single build command without label",
			line: "go build -o auto ./cmd/auto",
			expected: []BuildEntry{
				{Command: "go build -o auto ./cmd/auto", Label: "", SubmodulePath: ""},
			},
		},
		{
			name: "empty build line",
			line: "",
			expected: nil,
		},
		{
			name: "whitespace only",
			line: "   ",
			expected: nil,
		},
		{
			name: "extra whitespace around entries",
			line: "  go build ./cmd/auto/ (ADK) ,  go build ./cmd/server/ (Backend)  ",
			expected: []BuildEntry{
				{Command: "go build ./cmd/auto/", Label: "ADK", SubmodulePath: "autopus-adk"},
				{Command: "go build ./cmd/server/", Label: "Backend", SubmodulePath: "Autopus"},
			},
		},
		{
			name: "single entry with label",
			line: "go build ./cmd/bridge/ (Bridge)",
			expected: []BuildEntry{
				{Command: "go build ./cmd/bridge/", Label: "Bridge", SubmodulePath: "autopus-bridge"},
			},
		},
		{
			name: "unknown label",
			line: "cargo build (Rust)",
			expected: []BuildEntry{
				{Command: "cargo build", Label: "Rust", SubmodulePath: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseBuildLine(tt.line)

			if len(got) != len(tt.expected) {
				t.Fatalf("ParseBuildLine(%q): got %d entries, want %d", tt.line, len(got), len(tt.expected))
			}
			for i, entry := range got {
				exp := tt.expected[i]
				if entry.Command != exp.Command {
					t.Errorf("entry[%d].Command = %q, want %q", i, entry.Command, exp.Command)
				}
				if entry.Label != exp.Label {
					t.Errorf("entry[%d].Label = %q, want %q", i, entry.Label, exp.Label)
				}
				if entry.SubmodulePath != exp.SubmodulePath {
					t.Errorf("entry[%d].SubmodulePath = %q, want %q", i, entry.SubmodulePath, exp.SubmodulePath)
				}
			}
		})
	}
}

func TestResolveBuildDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		projectDir string
		entry      BuildEntry
		expected   string
	}{
		{
			name:       "ADK label resolves to submodule",
			projectDir: "/home/user/project",
			entry:      BuildEntry{Label: "ADK", SubmodulePath: "autopus-adk"},
			expected:   filepath.Join("/home/user/project", "autopus-adk"),
		},
		{
			name:       "unknown label falls back to projectDir",
			projectDir: "/home/user/project",
			entry:      BuildEntry{Label: "Unknown", SubmodulePath: ""},
			expected:   "/home/user/project",
		},
		{
			name:       "empty label falls back to projectDir",
			projectDir: "/home/user/project",
			entry:      BuildEntry{Label: "", SubmodulePath: ""},
			expected:   "/home/user/project",
		},
		{
			name:       "Frontend label resolves to nested path",
			projectDir: "/home/user/project",
			entry:      BuildEntry{Label: "Frontend", SubmodulePath: "Autopus/frontend"},
			expected:   filepath.Join("/home/user/project", "Autopus/frontend"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveBuildDir(tt.projectDir, tt.entry)
			if got != tt.expected {
				t.Errorf("ResolveBuildDir(%q, %+v) = %q, want %q", tt.projectDir, tt.entry, got, tt.expected)
			}
		})
	}
}

func TestMatchBuild(t *testing.T) {
	t.Parallel()

	builds := []BuildEntry{
		{Command: "go build ./cmd/auto/", Label: "ADK", SubmodulePath: "autopus-adk"},
		{Command: "go build ./cmd/server/", Label: "Backend", SubmodulePath: "Autopus"},
		{Command: "npm run build", Label: "Frontend", SubmodulePath: "Autopus/frontend"},
	}

	tests := []struct {
		name     string
		scenario Scenario
		builds   []BuildEntry
		wantNil  bool
		wantCmd  string
	}{
		{
			name:     "ADK CLI Scenarios matches ADK build",
			scenario: Scenario{Section: "ADK CLI Scenarios"},
			builds:   builds,
			wantCmd:  "go build ./cmd/auto/",
		},
		{
			name:     "Backend API Scenarios matches Backend build",
			scenario: Scenario{Section: "Backend API Scenarios"},
			builds:   builds,
			wantCmd:  "go build ./cmd/server/",
		},
		{
			name:     "Frontend Scenarios matches Frontend build",
			scenario: Scenario{Section: "Frontend Scenarios"},
			builds:   builds,
			wantCmd:  "npm run build",
		},
		{
			name:     "no matching section returns nil",
			scenario: Scenario{Section: "Unknown Scenarios"},
			builds:   builds,
			wantNil:  true,
		},
		{
			name:     "empty builds returns nil",
			scenario: Scenario{Section: "ADK CLI Scenarios"},
			builds:   nil,
			wantNil:  true,
		},
		{
			name:     "empty section returns nil",
			scenario: Scenario{Section: ""},
			builds:   builds,
			wantNil:  true,
		},
		{
			name: "single build without label always matches",
			scenario: Scenario{Section: "Any Section"},
			builds: []BuildEntry{
				{Command: "go build -o auto ./cmd/auto", Label: "", SubmodulePath: ""},
			},
			wantCmd: "go build -o auto ./cmd/auto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MatchBuild(tt.scenario, tt.builds)
			if tt.wantNil {
				if got != nil {
					t.Errorf("MatchBuild() = %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("MatchBuild() = nil, want non-nil")
			}
			if got.Command != tt.wantCmd {
				t.Errorf("MatchBuild().Command = %q, want %q", got.Command, tt.wantCmd)
			}
		})
	}
}
