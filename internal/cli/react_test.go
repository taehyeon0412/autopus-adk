package cli

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunReactCheck_NoRemoteQuiet(t *testing.T) {
	t.Cleanup(func() {
		reactLookPath = execLookPath
		reactOutput = execOutput
	})
	reactLookPath = func(file string) (string, error) { return "/usr/bin/gh", nil }
	reactOutput = func(name string, args ...string) ([]byte, error) {
		if name == "git" {
			return []byte(""), nil
		}
		return nil, errors.New("unexpected command")
	}

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := runReactCheck(cmd, nil, true)
	require.NoError(t, err)
	assert.Empty(t, out.String())
}

func TestRunReactCheck_NoRemoteVerbose(t *testing.T) {
	t.Cleanup(func() {
		reactLookPath = execLookPath
		reactOutput = execOutput
	})
	reactLookPath = func(file string) (string, error) { return "/usr/bin/gh", nil }
	reactOutput = func(name string, args ...string) ([]byte, error) {
		if name == "git" {
			return []byte(""), nil
		}
		return nil, errors.New("unexpected command")
	}

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := runReactCheck(cmd, nil, false)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "No git remote configured")
}

var (
	execLookPath = reactLookPath
	execOutput   = reactOutput
)
