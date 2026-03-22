package sigmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- firstSentence ----

func TestFirstSentence_Simple(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "NewServer creates a new server.", firstSentence("NewServer creates a new server. Other details follow."))
}

func TestFirstSentence_NoTerminator(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "No period here", firstSentence("No period here"))
}

func TestFirstSentence_Empty(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", firstSentence(""))
}

func TestFirstSentence_PeriodAtEnd(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Single sentence.", firstSentence("Single sentence."))
}

// ---- isExported ----

func TestIsExported_Exported(t *testing.T) {
	t.Parallel()
	assert.True(t, isExported("MyFunc"))
	assert.True(t, isExported("Server"))
}

func TestIsExported_Unexported(t *testing.T) {
	t.Parallel()
	assert.False(t, isExported("myFunc"))
	assert.False(t, isExported("_internal"))
	assert.False(t, isExported(""))
}

// ---- extractModulePath ----

func TestExtractModulePath_Valid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module github.com/acme/myapp\n\ngo 1.21\n")
	mod, err := extractModulePath(dir)
	require.NoError(t, err)
	assert.Equal(t, "github.com/acme/myapp", mod)
}

func TestExtractModulePath_Missing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := extractModulePath(dir)
	assert.Error(t, err)
}

func TestExtractModulePath_NoModuleLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "go 1.21\n")
	_, err := extractModulePath(dir)
	assert.Error(t, err)
}

// ---- formatFieldList edge cases ----

func TestFormatFieldList_Nil(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "()", formatFieldList(nil))
}

func TestFormatFieldList_NoNames(t *testing.T) {
	t.Parallel()
	// A function returning a single unnamed type, e.g. "func() error"
	dir := makeProject(t)
	writeFile(t, dir, "pkg/ret/ret.go", `package ret

func ReturnError() error { return nil }
func ReturnMultiple() (int, error) { return 0, nil }
`)
	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)

	names := make(map[string]string)
	for _, s := range sm.Packages[0].Signatures {
		names[s.Name] = s.Returns
	}
	assert.Contains(t, names["ReturnError"], "error")
	assert.Contains(t, names["ReturnMultiple"], "int")
}

// ---- formatTypeParams edge cases ----

func TestFormatTypeParams_Nil(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", formatTypeParams(nil))
}
