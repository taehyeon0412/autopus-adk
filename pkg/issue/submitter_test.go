package issue_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/issue"
)

// mockRunner is a test double for CommandRunner.
type mockRunner struct {
	lookPathFn func(name string) (string, error)
	runFn      func(name string, args ...string) ([]byte, error)
}

func (m *mockRunner) LookPath(name string) (string, error) {
	if m.lookPathFn != nil {
		return m.lookPathFn(name)
	}
	return "/usr/bin/" + name, nil
}

func (m *mockRunner) Run(name string, args ...string) ([]byte, error) {
	if m.runFn != nil {
		return m.runFn(name, args...)
	}
	return nil, nil
}

func TestCheckGH_NotInstalled(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		lookPathFn: func(name string) (string, error) {
			return "", errors.New("not found")
		},
	}
	submitter := issue.NewSubmitter(runner)
	err := submitter.CheckGH()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gh CLI")
}

func TestCheckGH_NotAuthenticated(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		lookPathFn: func(name string) (string, error) {
			return "/usr/bin/gh", nil
		},
		runFn: func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("not authenticated")
		},
	}
	submitter := issue.NewSubmitter(runner)
	err := submitter.CheckGH()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authenticated")
}

func TestCheckGH_OK(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		lookPathFn: func(name string) (string, error) {
			return "/usr/bin/gh", nil
		},
		runFn: func(name string, args ...string) ([]byte, error) {
			return []byte("Logged in to github.com"), nil
		},
	}
	submitter := issue.NewSubmitter(runner)
	err := submitter.CheckGH()
	require.NoError(t, err)
}

func TestComputeHash(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{}
	submitter := issue.NewSubmitter(runner)

	h1 := submitter.ComputeHash("error msg", "auto plan")
	h2 := submitter.ComputeHash("error msg", "auto plan")
	h3 := submitter.ComputeHash("different", "auto plan")

	assert.Equal(t, h1, h2, "same inputs must yield same hash")
	assert.NotEqual(t, h1, h3, "different inputs must yield different hash")
	assert.NotEmpty(t, h1)
}

func TestFindDuplicate_Found(t *testing.T) {
	t.Parallel()

	type ghIssue struct {
		Number int    `json:"number"`
		URL    string `json:"url"`
	}
	payload, _ := json.Marshal([]ghIssue{{Number: 42, URL: "https://github.com/org/repo/issues/42"}})

	runner := &mockRunner{
		runFn: func(name string, args ...string) ([]byte, error) {
			return payload, nil
		},
	}
	submitter := issue.NewSubmitter(runner)

	url, err := submitter.FindDuplicate("org/repo", "deadbeef")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/org/repo/issues/42", url)
}

func TestFindDuplicate_NotFound(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		runFn: func(name string, args ...string) ([]byte, error) {
			return []byte("[]"), nil
		},
	}
	submitter := issue.NewSubmitter(runner)

	url, err := submitter.FindDuplicate("org/repo", "deadbeef")
	require.NoError(t, err)
	assert.Empty(t, url)
}

func TestFindDuplicate_Error(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		runFn: func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("network error")
		},
	}
	submitter := issue.NewSubmitter(runner)

	_, err := submitter.FindDuplicate("org/repo", "hash")
	require.Error(t, err)
}

func TestFindDuplicate_InvalidJSON(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		runFn: func(name string, args ...string) ([]byte, error) {
			return []byte("not-json"), nil
		},
	}
	submitter := issue.NewSubmitter(runner)

	_, err := submitter.FindDuplicate("org/repo", "hash")
	require.Error(t, err)
}

func TestCreateIssue_Success(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		runFn: func(name string, args ...string) ([]byte, error) {
			return []byte("https://github.com/org/repo/issues/99"), nil
		},
	}
	submitter := issue.NewSubmitter(runner)

	result, err := submitter.CreateIssue("org/repo", "Test Issue", "body content", []string{"auto-report"})
	require.NoError(t, err)
	assert.Equal(t, "created", result.Action)
	assert.Equal(t, "https://github.com/org/repo/issues/99", result.IssueURL)
	assert.Equal(t, 99, result.IssueNumber)
}

func TestCreateIssue_Error(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{
		runFn: func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("gh: repository not found")
		},
	}
	submitter := issue.NewSubmitter(runner)

	_, err := submitter.CreateIssue("org/repo", "title", "body", nil)
	require.Error(t, err)
}

func TestNewSubmitter_NilRunner(t *testing.T) {
	t.Parallel()

	submitter := issue.NewSubmitter(nil)
	require.NotNil(t, submitter)
}
