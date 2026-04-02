package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		policy  SecurityPolicy
		command string
		workDir string
		wantOK  bool
		wantMsg string
	}{
		{
			name: "allowed by prefix match",
			policy: SecurityPolicy{
				AllowedCommands: []string{"go ", "git "},
			},
			command: "go test ./...",
			wantOK:  true,
		},
		{
			name: "denied by regex pattern",
			policy: SecurityPolicy{
				AllowedCommands: []string{"sh "},
				DeniedPatterns:  []string{`rm\s+-rf`},
			},
			command: "sh -c rm -rf /",
			wantOK:  false,
			wantMsg: "command matches denied pattern",
		},
		{
			name: "command not in allowed list",
			policy: SecurityPolicy{
				AllowedCommands: []string{"go ", "git "},
			},
			command: "curl http://example.com",
			wantOK:  false,
			wantMsg: "command not in allowed list",
		},
		{
			name:    "empty AllowedCommands denies all",
			policy:  SecurityPolicy{},
			command: "echo hello",
			wantOK:  false,
			wantMsg: "no allowed commands configured (fail-closed)",
		},
		{
			name: "workDir not in AllowedDirs",
			policy: SecurityPolicy{
				AllowedCommands: []string{"go "},
				AllowedDirs:     []string{"/home/user/project"},
			},
			command: "go build",
			workDir: "/tmp/evil",
			wantOK:  false,
			wantMsg: "working directory not in allowed list",
		},
		{
			name: "workDir in AllowedDirs",
			policy: SecurityPolicy{
				AllowedCommands: []string{"go "},
				AllowedDirs:     []string{"/home/user/project"},
			},
			command: "go build",
			workDir: "/home/user/project/cmd",
			wantOK:  true,
		},
		{
			name: "invalid regex in DeniedPatterns",
			policy: SecurityPolicy{
				AllowedCommands: []string{"echo "},
				DeniedPatterns:  []string{"[invalid"},
			},
			command: "echo hello",
			wantOK:  false,
			wantMsg: "invalid denied pattern",
		},
		{
			name: "empty workDir skips dir check",
			policy: SecurityPolicy{
				AllowedCommands: []string{"go "},
				AllowedDirs:     []string{"/home/user/project"},
			},
			command: "go test",
			workDir: "",
			wantOK:  true,
		},
		{
			name: "denied pattern takes precedence over allowed",
			policy: SecurityPolicy{
				AllowedCommands: []string{"git "},
				DeniedPatterns:  []string{`push.*--force`},
			},
			command: "git push --force",
			wantOK:  false,
			wantMsg: "command matches denied pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ok, msg := tt.policy.ValidateCommand(tt.command, tt.workDir)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantMsg != "" {
				assert.Contains(t, msg, tt.wantMsg)
			}
		})
	}
}
