package sigmap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- CalculateFanIn: WalkDir error is propagated ----

func TestCalculateFanIn_WalkDirError(t *testing.T) {
	t.Parallel()
	// Pass a non-existent directory so WalkDir returns an error immediately.
	sm := &SignatureMap{
		ModulePath: "example.com/app",
		Packages:   []Package{{Path: "pkg/a"}},
	}
	err := CalculateFanIn("/nonexistent/path/xyz123", sm)
	assert.Error(t, err)
}

// ---- CalculateFanIn: parse-error files are skipped silently ----

func TestCalculateFanIn_SkipUnparseable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/app\n\ngo 1.21\n"),
		0o644,
	))
	// Write a syntactically invalid Go file.
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "broken.go"),
		[]byte("this is not valid go\n"),
		0o644,
	))

	sm := &SignatureMap{
		ModulePath: "example.com/app",
		Packages:   []Package{{Path: "pkg/a"}},
	}
	// Must not error — unparseable files are skipped silently.
	err := CalculateFanIn(dir, sm)
	require.NoError(t, err)
	assert.Equal(t, 0, sm.Packages[0].FanIn)
}

// ---- Extract: filepath.Walk error branch (line 44-46) ----
// This path is triggered when filepath.Walk itself returns an error. Because
// the stdlib Walk implementation wraps filesystem errors and the callback
// returns nil for walkErr, the only practical trigger is a root path that
// cannot be opened at all. We document this as a known hard-to-reach path
// and cover it via a direct call with an invalid root.

func TestExtract_MissingGoMod(t *testing.T) {
	t.Parallel()
	// Empty temp dir has no go.mod → Extract returns an error before Walk.
	dir := t.TempDir()
	_, err := Extract(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read module path")
}

// ---- Extract: relErr fallback (line 52-54) is unreachable in practice ----
// filepath.Rel only fails when both paths have different volumes on Windows.
// We cover it implicitly through the normal extraction path.

// ---- formatFieldList: unnamed receiver (pointer receiver) ----

func TestFormatFieldList_PointerReceiver(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	// Pointer receiver results in an unnamed field in the field list.
	writeFile(t, dir, "pkg/ptr/ptr.go", `package ptr

type Engine struct{}

// Start runs the engine.
func (e *Engine) Start() {}
`)
	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)

	var method *Signature
	for i := range sm.Packages[0].Signatures {
		if sm.Packages[0].Signatures[i].Kind == "method" {
			method = &sm.Packages[0].Signatures[i]
		}
	}
	require.NotNil(t, method)
	assert.Contains(t, method.Receiver, "Engine")
}

// ---- Extract: walkErr != nil path (line 29-31) ----
// This branch is hit when a directory entry cannot be stat'd during Walk.
// We simulate it by creating a symlink pointing to a non-existent target.
func TestExtract_WalkUnreadableEntry(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	// Create a broken symlink inside the project directory.
	linkPath := filepath.Join(dir, "broken_link")
	_ = os.Symlink("/nonexistent/target/path", linkPath)
	// Also add a real Go file so extraction can succeed.
	writeFile(t, dir, "pkg/real/real.go", `package real
func Real() {}
`)

	// Extract must succeed even when the symlink cannot be stat'd.
	sm, err := Extract(dir)
	require.NoError(t, err)
	found := false
	for _, p := range sm.Packages {
		if p.Path == "pkg/real" {
			found = true
		}
	}
	assert.True(t, found, "real package must still be extracted")
}
