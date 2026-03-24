package sigmap_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/sigmap"
)

// TestTSExtractor_ExportFunction verifies that an exported TypeScript function
// declaration is detected and returned as a Signature with kind="func".
func TestTSExtractor_ExportFunction(t *testing.T) {
	t.Parallel()

	// Given: TypeScript source with an exported function
	src := `export function foo(): string { return "bar"; }`

	// When: TSExtractor extracts signatures
	ex := sigmap.NewTSExtractor()
	sigs, err := ex.Extract([]byte(src), "foo.ts")
	require.NoError(t, err)

	// Then: the exported function is found
	require.Len(t, sigs, 1)
	assert.Equal(t, "foo", sigs[0].Name)
	assert.Equal(t, "func", sigs[0].Kind)
}

// TestTSExtractor_ExportClass verifies that an exported TypeScript class is
// detected and returned as a Signature with kind="type".
func TestTSExtractor_ExportClass(t *testing.T) {
	t.Parallel()

	// Given: TypeScript source with an exported class
	src := `export class Foo { method(): void {} }`

	// When: TSExtractor extracts signatures
	ex := sigmap.NewTSExtractor()
	sigs, err := ex.Extract([]byte(src), "foo.ts")
	require.NoError(t, err)

	// Then: the exported class is found
	require.Len(t, sigs, 1)
	assert.Equal(t, "Foo", sigs[0].Name)
	assert.Equal(t, "type", sigs[0].Kind)
}

// TestTSExtractor_ExportInterface verifies that an exported TypeScript interface
// is detected and returned as a Signature with kind="interface".
func TestTSExtractor_ExportInterface(t *testing.T) {
	t.Parallel()

	// Given: TypeScript source with an exported interface
	src := `export interface Bar { id: number; name: string; }`

	// When: TSExtractor extracts signatures
	ex := sigmap.NewTSExtractor()
	sigs, err := ex.Extract([]byte(src), "bar.ts")
	require.NoError(t, err)

	// Then: the exported interface is found
	require.Len(t, sigs, 1)
	assert.Equal(t, "Bar", sigs[0].Name)
	assert.Equal(t, "interface", sigs[0].Kind)
}

// TestTSExtractor_ExportConst verifies that an exported TypeScript constant is
// detected and returned as a Signature with kind="const".
func TestTSExtractor_ExportConst(t *testing.T) {
	t.Parallel()

	// Given: TypeScript source with an exported constant
	src := `export const x = 42;`

	// When: TSExtractor extracts signatures
	ex := sigmap.NewTSExtractor()
	sigs, err := ex.Extract([]byte(src), "const.ts")
	require.NoError(t, err)

	// Then: the exported constant is found
	require.Len(t, sigs, 1)
	assert.Equal(t, "x", sigs[0].Name)
	assert.Equal(t, "const", sigs[0].Kind)
}

// TestTSExtractor_DefaultExport verifies that a TypeScript default export
// function is detected and returned with Name="default".
func TestTSExtractor_DefaultExport(t *testing.T) {
	t.Parallel()

	// Given: TypeScript source with a default export function
	src := `export default function handler(): void {}`

	// When: TSExtractor extracts signatures
	ex := sigmap.NewTSExtractor()
	sigs, err := ex.Extract([]byte(src), "handler.ts")
	require.NoError(t, err)

	// Then: the default export is found
	require.Len(t, sigs, 1)
	assert.Equal(t, "default", sigs[0].Name)
	assert.Equal(t, "func", sigs[0].Kind)
}

// TestTSExtractor_ReExport verifies that a TypeScript re-export statement is
// detected and returned as a Signature with kind="reexport".
func TestTSExtractor_ReExport(t *testing.T) {
	t.Parallel()

	// Given: TypeScript source with a re-export
	src := `export { foo } from './bar';`

	// When: TSExtractor extracts signatures
	ex := sigmap.NewTSExtractor()
	sigs, err := ex.Extract([]byte(src), "index.ts")
	require.NoError(t, err)

	// Then: the re-export is found
	require.Len(t, sigs, 1)
	assert.Equal(t, "foo", sigs[0].Name)
	assert.Equal(t, "reexport", sigs[0].Kind)
}

// TestTSExtractor_NoExports_ReturnsEmpty verifies that a TypeScript file with
// no exports returns an empty signature list without error.
func TestTSExtractor_NoExports_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Given: TypeScript source with no exports
	src := `const internal = 42; function helper() { return internal; }`

	// When: TSExtractor extracts signatures
	ex := sigmap.NewTSExtractor()
	sigs, err := ex.Extract([]byte(src), "internal.ts")
	require.NoError(t, err)

	// Then: no signatures are returned
	assert.Empty(t, sigs)
}

// TestTSExtractor_ExtractDir_WithTSFiles verifies that ExtractDir scans a
// directory for .ts/.tsx files and returns aggregated signatures.
func TestTSExtractor_ExtractDir_WithTSFiles(t *testing.T) {
	t.Parallel()

	// Given: a temp directory with two TypeScript files
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "foo.ts"),
		[]byte("export function foo(): void {}"),
		0o644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "bar.tsx"),
		[]byte("export class Bar {}"),
		0o644,
	))
	// Non-.ts file should be ignored
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "ignore.js"),
		[]byte("export function shouldIgnore() {}"),
		0o644,
	))

	// When: ExtractDir is called
	ex := sigmap.NewTSExtractor()
	sm, err := ex.ExtractDir(dir)
	require.NoError(t, err)

	// Then: two packages are returned (one per .ts/.tsx file with exports)
	require.NotNil(t, sm)
	assert.Len(t, sm.Packages, 2)

	// Collect all signature names across packages
	var names []string
	for _, pkg := range sm.Packages {
		for _, sig := range pkg.Signatures {
			names = append(names, sig.Name)
		}
	}
	assert.Contains(t, names, "foo")
	assert.Contains(t, names, "Bar")
}

// TestTSExtractor_ExtractDir_EmptyDir_ReturnsEmptyMap verifies that ExtractDir
// on an empty directory returns an empty SignatureMap without error.
func TestTSExtractor_ExtractDir_EmptyDir_ReturnsEmptyMap(t *testing.T) {
	t.Parallel()

	// Given: an empty directory
	dir := t.TempDir()

	// When: ExtractDir is called
	ex := sigmap.NewTSExtractor()
	sm, err := ex.ExtractDir(dir)
	require.NoError(t, err)

	// Then: an empty map is returned
	require.NotNil(t, sm)
	assert.Empty(t, sm.Packages)
}

// TestTSExtractor_ExtractDir_NonExistentDir_ReturnsError verifies that
// ExtractDir returns an error for a non-existent directory.
func TestTSExtractor_ExtractDir_NonExistentDir_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: a non-existent directory
	dir := filepath.Join(t.TempDir(), "no-such-dir")

	// When: ExtractDir is called
	ex := sigmap.NewTSExtractor()
	_, err := ex.ExtractDir(dir)

	// Then: an error is returned
	require.Error(t, err)
}

// TestTSExtractor_ExtractDir_SubdirIsSkipped verifies that subdirectories
// inside the target directory are skipped (not recursed into).
func TestTSExtractor_ExtractDir_SubdirIsSkipped(t *testing.T) {
	t.Parallel()

	// Given: a directory with a .ts file and a subdirectory
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "subdir"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "main.ts"),
		[]byte("export function main(): void {}"),
		0o644,
	))

	// When: ExtractDir is called
	ex := sigmap.NewTSExtractor()
	sm, err := ex.ExtractDir(dir)
	require.NoError(t, err)

	// Then: only the .ts file is processed (subdir is skipped)
	require.Len(t, sm.Packages, 1)
	assert.Equal(t, "main", sm.Packages[0].Name)
}

// TestTSExtractor_ExtractDir_UnreadableFile_AddsWarning verifies that when a
// .ts file cannot be read, a warning is added to the SignatureMap.
func TestTSExtractor_ExtractDir_UnreadableFile_AddsWarning(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: permission tests are not meaningful")
	}
	t.Parallel()

	// Given: a directory with an unreadable .ts file
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.ts")
	require.NoError(t, os.WriteFile(path, []byte("export const x = 1;"), 0o644))
	require.NoError(t, os.Chmod(path, 0o000))
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	// When: ExtractDir is called
	ex := sigmap.NewTSExtractor()
	sm, err := ex.ExtractDir(dir)
	require.NoError(t, err)

	// Then: a warning is added and no packages are returned
	assert.Empty(t, sm.Packages)
	assert.NotEmpty(t, sm.Warnings)
}
