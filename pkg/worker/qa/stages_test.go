package qa

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cmd     string
		wantErr bool
	}{
		{"go build allowed", "go build ./...", false},
		{"go test allowed", "go test -race ./...", false},
		{"npm test allowed", "npm test", false},
		{"make allowed", "make build", false},
		{"bare make allowed", "make", false},
		{"docker build allowed", "docker build .", false},
		{"rm denied", "rm -rf /", true},
		{"curl denied", "curl http://evil.com", true},
		{"bash denied", "bash -c 'echo pwned'", true},
		{"empty string denied", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateCommand(tt.cmd)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.cmd != "" {
					assert.Contains(t, err.Error(), "not in allowlist")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildStage_AllowedCommand(t *testing.T) {
	t.Parallel()

	s := &BuildStage{Command: "go version"}
	result, err := s.Run(context.Background(), t.TempDir())

	require.NoError(t, err)
	assert.Equal(t, "build", result.Name)
	assert.Equal(t, "pass", result.Status)
	assert.Contains(t, result.Output, "go version")
}

func TestBuildStage_DeniedCommand(t *testing.T) {
	t.Parallel()

	s := &BuildStage{Command: "rm -rf /"}
	result, err := s.Run(context.Background(), t.TempDir())

	require.Error(t, err)
	assert.Equal(t, "fail", result.Status)
	assert.Contains(t, result.Output, "not in allowlist")
}

func TestTestStage_AllowedCommand(t *testing.T) {
	t.Parallel()

	s := &TestStage{Command: "go version"}
	result, err := s.Run(context.Background(), t.TempDir())

	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, "pass", result.Status)
}

func TestTestStage_DeniedCommand(t *testing.T) {
	t.Parallel()

	s := &TestStage{Command: "curl http://evil.com"}
	result, err := s.Run(context.Background(), t.TempDir())

	require.Error(t, err)
	assert.Equal(t, "fail", result.Status)
}

func TestServiceHealthStage_HealthyServer(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := &ServiceHealthStage{
		URL:      srv.URL,
		Interval: 50 * time.Millisecond,
		Timeout:  2 * time.Second,
	}
	result, err := s.Run(context.Background(), "")

	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
	assert.Contains(t, result.Output, "health check passed")
}

func TestServiceHealthStage_Timeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	s := &ServiceHealthStage{
		URL:      srv.URL,
		Interval: 50 * time.Millisecond,
		Timeout:  200 * time.Millisecond,
	}
	result, err := s.Run(context.Background(), "")

	require.Error(t, err)
	assert.Equal(t, "fail", result.Status)
	assert.Contains(t, result.Output, "timed out")
}

func TestServiceHealthStage_ContextCancelled(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	s := &ServiceHealthStage{
		URL:      srv.URL,
		Interval: 50 * time.Millisecond,
		Timeout:  10 * time.Second,
	}

	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	result, err := s.Run(ctx, "")
	require.Error(t, err)
	assert.Equal(t, "fail", result.Status)
	assert.Contains(t, result.Output, "context cancelled")
}

func TestCleanupStage_ValidCommands(t *testing.T) {
	t.Parallel()

	s := &CleanupStage{Commands: []string{"go version"}}
	result, err := s.Run(context.Background(), t.TempDir())

	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
	assert.Contains(t, result.Output, "[go version] ok")
}

func TestCleanupStage_BlockedCommand(t *testing.T) {
	t.Parallel()

	s := &CleanupStage{Commands: []string{"rm -rf /"}}
	result, err := s.Run(context.Background(), t.TempDir())

	require.Error(t, err)
	assert.Equal(t, "fail", result.Status)
	assert.Contains(t, result.Output, "blocked")
}
