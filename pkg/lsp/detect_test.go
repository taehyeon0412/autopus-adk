package lsp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/lsp"
)

func TestDetectServer_GoProject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.23\n"), 0o644))

	serverCmd, args, err := lsp.DetectServer(dir)
	require.NoError(t, err)

	assert.Equal(t, "gopls", serverCmd)
	assert.NotNil(t, args)
}

func TestDetectServer_TSProject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0o644))

	serverCmd, args, err := lsp.DetectServer(dir)
	require.NoError(t, err)

	assert.Equal(t, "typescript-language-server", serverCmd)
	assert.Contains(t, args, "--stdio")
}

func TestDetectServer_PythonProject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "setup.py"), []byte("from setuptools import setup\nsetup(name='test')\n"), 0o644))

	serverCmd, args, err := lsp.DetectServer(dir)
	require.NoError(t, err)

	assert.Equal(t, "pyright", serverCmd)
	assert.Contains(t, args, "--stdio")
}

func TestDetectServer_UnknownProject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// 인식할 수 없는 프로젝트
	_, _, err := lsp.DetectServer(dir)
	assert.Error(t, err)
}

func TestDetectServer_NonExistentDir(t *testing.T) {
	t.Parallel()

	_, _, err := lsp.DetectServer("/nonexistent/path")
	assert.Error(t, err)
}
