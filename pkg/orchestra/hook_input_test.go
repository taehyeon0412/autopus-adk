package orchestra

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHookInput_JSONMarshal verifies that HookInput struct serializes
// to the expected JSON format with provider, round, and prompt fields.
func TestHookInput_JSONMarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    HookInput
		expected map[string]interface{}
	}{
		{
			"basic input",
			HookInput{Provider: "claude", Round: 1, Prompt: "hello world"},
			map[string]interface{}{
				"provider": "claude",
				"round":    float64(1),
				"prompt":   "hello world",
			},
		},
		{
			"empty prompt",
			HookInput{Provider: "gemini", Round: 0, Prompt: ""},
			map[string]interface{}{
				"provider": "gemini",
				"round":    float64(0),
				"prompt":   "",
			},
		},
		{
			"multiline prompt",
			HookInput{Provider: "opencode", Round: 5, Prompt: "line1\nline2\nline3"},
			map[string]interface{}{
				"provider": "opencode",
				"round":    float64(5),
				"prompt":   "line1\nline2\nline3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// When: marshal HookInput to JSON
			data, err := json.Marshal(tt.input)
			require.NoError(t, err)

			// Then: JSON contains expected fields
			var got map[string]interface{}
			require.NoError(t, json.Unmarshal(data, &got))
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestAtomicWriteJSON_CreatesFile verifies that atomicWriteJSON creates
// a file with correct JSON content and 0o600 permissions.
func TestAtomicWriteJSON_CreatesFile(t *testing.T) {
	t.Parallel()

	// Given: a temporary directory and a HookInput
	dir := t.TempDir()
	target := dir + "/input.json"
	input := HookInput{Provider: "claude", Round: 1, Prompt: "test prompt"}

	// When: atomicWriteJSON writes the file
	err := atomicWriteJSON(target, input)
	require.NoError(t, err)

	// Then: file exists with correct content
	data, err := os.ReadFile(target)
	require.NoError(t, err)

	var got HookInput
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, input.Provider, got.Provider)
	assert.Equal(t, input.Round, got.Round)
	assert.Equal(t, input.Prompt, got.Prompt)

	// Then: file has 0o600 permissions
	info, err := os.Stat(target)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

// TestAtomicWriteJSON_AtomicWrite verifies that atomicWriteJSON uses
// tmp+rename pattern so the file appears atomically (no partial writes).
func TestAtomicWriteJSON_AtomicWrite(t *testing.T) {
	t.Parallel()

	// Given: a temporary directory
	dir := t.TempDir()
	target := dir + "/atomic-test.json"
	input := HookInput{Provider: "claude", Round: 2, Prompt: "atomic test"}

	// When: atomicWriteJSON completes
	err := atomicWriteJSON(target, input)
	require.NoError(t, err)

	// Then: no temporary files remain in the directory
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "only the target file should remain, no tmp files")
	assert.Equal(t, "atomic-test.json", entries[0].Name())
}

// TestAtomicWriteJSON_InvalidPath verifies that atomicWriteJSON returns
// an error when the target path does not exist.
func TestAtomicWriteJSON_InvalidPath(t *testing.T) {
	t.Parallel()

	// Given: an invalid directory path
	target := "/nonexistent/path/that/does/not/exist/input.json"
	input := HookInput{Provider: "claude", Round: 1, Prompt: "fail"}

	// When: atomicWriteJSON tries to write
	err := atomicWriteJSON(target, input)

	// Then: error is returned
	assert.Error(t, err, "atomicWriteJSON must return error for invalid path")
}
