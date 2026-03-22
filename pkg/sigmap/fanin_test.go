package sigmap

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateFanIn_Basic(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Write go.mod.
	require.NoError(t, os.WriteFile(
		dir+"/go.mod",
		[]byte("module example.com/myapp\n\ngo 1.21\n"),
		0o644,
	))

	// Write a file that imports pkg/util and pkg/config.
	require.NoError(t, os.MkdirAll(dir+"/cmd", 0o755))
	require.NoError(t, os.WriteFile(dir+"/cmd/main.go", []byte(`package main

import (
	"example.com/myapp/pkg/util"
	"example.com/myapp/pkg/config"
)

func main() { _ = util.X; _ = config.Y }
`), 0o644))

	sm := &SignatureMap{
		ModulePath: "example.com/myapp",
		Packages: []Package{
			{Path: "pkg/util"},
			{Path: "pkg/config"},
			{Path: "pkg/unused"},
		},
	}

	err := CalculateFanIn(dir, sm)
	require.NoError(t, err)

	assert.Equal(t, 1, sm.Packages[0].FanIn, "pkg/util should have fan-in 1")
	assert.Equal(t, 1, sm.Packages[1].FanIn, "pkg/config should have fan-in 1")
	assert.Equal(t, 0, sm.Packages[2].FanIn, "pkg/unused should have fan-in 0")
}

func TestCalculateFanIn_SkipsTestFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		dir+"/go.mod",
		[]byte("module example.com/app\n\ngo 1.21\n"),
		0o644,
	))
	require.NoError(t, os.MkdirAll(dir+"/pkg/service", 0o755))
	// Test file that imports pkg/util — should be skipped.
	require.NoError(t, os.WriteFile(dir+"/pkg/service/svc_test.go", []byte(`package service

import "example.com/app/pkg/util"

func TestHelper() { _ = util.X }
`), 0o644))

	sm := &SignatureMap{
		ModulePath: "example.com/app",
		Packages:   []Package{{Path: "pkg/util"}},
	}
	require.NoError(t, CalculateFanIn(dir, sm))
	assert.Equal(t, 0, sm.Packages[0].FanIn, "imports in _test.go must not count")
}

func TestCalculateFanIn_SkipsVendor(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		dir+"/go.mod",
		[]byte("module example.com/app\n\ngo 1.21\n"),
		0o644,
	))
	require.NoError(t, os.MkdirAll(dir+"/vendor/some/dep", 0o755))
	require.NoError(t, os.WriteFile(dir+"/vendor/some/dep/dep.go", []byte(`package dep

import "example.com/app/pkg/util"

func Dep() { _ = util.X }
`), 0o644))

	sm := &SignatureMap{
		ModulePath: "example.com/app",
		Packages:   []Package{{Path: "pkg/util"}},
	}
	require.NoError(t, CalculateFanIn(dir, sm))
	assert.Equal(t, 0, sm.Packages[0].FanIn, "vendor imports must not count")
}
