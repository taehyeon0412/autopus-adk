package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectFramework_NextJS tests S3: detect Next.js framework (next.config.*).
// R2: /auto setup detects project stack and maps to profile.
func TestDetectFramework_NextJS(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "next.config.ts"), []byte("// next config"), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "nextjs", fw.Name)
	assert.Equal(t, "typescript", fw.Stack)
}

// TestDetectFramework_NextJSConfigJS tests S3 variant: next.config.js also triggers Next.js detection.
func TestDetectFramework_NextJSConfigJS(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "next.config.js"), []byte("module.exports = {}"), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "nextjs", fw.Name)
}

// TestDetectFramework_FastAPI tests S4: detect FastAPI (fastapi in pyproject.toml).
// R2: /auto setup detects project stack and maps to profile.
func TestDetectFramework_FastAPI(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pyprojectContent := `[tool.poetry]
name = "my-api"
[tool.poetry.dependencies]
python = "^3.11"
fastapi = "^0.100"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(pyprojectContent), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "fastapi", fw.Name)
	assert.Equal(t, "python", fw.Stack)
}

// TestDetectFramework_NoFramework tests S5: no framework detected → empty/nil result.
// R6: Graceful fallback when no profile matches.
func TestDetectFramework_NoFramework(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Empty dir — no framework signals present

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	assert.Nil(t, fw, "no framework signals must return nil, not panic")
}

// TestDetectFramework_EmptyDir tests edge case: completely empty directory.
func TestDetectFramework_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	assert.Nil(t, fw)
}

// TestDetectFramework_GoProject tests: Go project with go.mod → go stack, no framework.
// R3: Profile assignment priority: framework > language > none.
func TestDetectFramework_GoProject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	// Go has no framework in SPEC — expect nil framework, stack detection handled elsewhere
	// OR a "go" stack detection. Either nil or stack="go" with no framework name is acceptable.
	if fw != nil {
		assert.Equal(t, "go", fw.Stack)
		assert.Empty(t, fw.Name, "pure go project must not have a framework name")
	}
}

// TestFramework_FieldsComplete verifies the Framework struct has the expected fields.
// This validates the struct definition exists as specified in SPEC-PROF-001.
func TestFramework_FieldsComplete(t *testing.T) {
	t.Parallel()

	fw := Framework{
		Name:   "nextjs",
		Stack:  "typescript",
		Signal: "next.config.ts",
	}

	assert.Equal(t, "nextjs", fw.Name)
	assert.Equal(t, "typescript", fw.Stack)
	assert.Equal(t, "next.config.ts", fw.Signal)
}

// TestDetectFramework_NonexistentDir tests graceful handling of missing directory.
func TestDetectFramework_NonexistentDir(t *testing.T) {
	t.Parallel()

	_, err := DetectFramework("/nonexistent/path/xyz/abc")
	// Either returns error or nil,nil — must not panic
	assert.NotPanics(t, func() {
		_, _ = DetectFramework("/nonexistent/path/xyz/abc")
	})
	_ = err
}

// TestDetectStackFromDir tests stack detection (go/typescript/python) from directory signals.
// R2: /auto setup detects project stack and maps to profile.
func TestDetectStackFromDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fileToAdd string
		content   string
		wantStack string
	}{
		{
			name:      "go.mod → go stack",
			fileToAdd: "go.mod",
			content:   "module example.com/test\n\ngo 1.21\n",
			wantStack: "go",
		},
		{
			name:      "package.json → typescript or javascript stack",
			fileToAdd: "package.json",
			content:   `{"name":"test"}`,
			wantStack: "typescript",
		},
		{
			name:      "pyproject.toml → python stack",
			fileToAdd: "pyproject.toml",
			content:   "[tool.poetry]\nname = \"test\"\n",
			wantStack: "python",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, tc.fileToAdd), []byte(tc.content), 0644))

			stack, err := DetectStack(dir)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStack, stack)
		})
	}
}

// TestDetectStack_EmptyDir tests R6: graceful fallback when nothing detected.
func TestDetectStack_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	stack, err := DetectStack(dir)
	require.NoError(t, err)
	assert.Empty(t, stack, "empty dir must return empty stack string")
}

// TestDetectStack_PythonFromPyproject tests pyproject.toml → python stack.
func TestDetectStack_PythonFromPyproject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[tool.poetry]\nname = \"test\"\n"), 0644))

	stack, err := DetectStack(dir)
	require.NoError(t, err)
	assert.Equal(t, "python", stack)
}

// TestDetectFramework_DjangoFromPyproject tests Django detection from pyproject.toml.
func TestDetectFramework_DjangoFromPyproject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := "[tool.poetry.dependencies]\ndjango = \"^4.0\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(content), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "django", fw.Name)
	assert.Equal(t, "python", fw.Stack)
}

// TestDetectFramework_FlaskFromRequirements tests Flask detection from requirements.txt.
func TestDetectFramework_FlaskFromRequirements(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask==2.3.0\ngunicorn\n"), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "flask", fw.Name)
	assert.Equal(t, "python", fw.Stack)
}

// TestDetectFramework_EchoFromGoMod tests Echo framework detection from go.mod.
func TestDetectFramework_EchoFromGoMod(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	goMod := "module example.com/test\n\ngo 1.21\n\nrequire github.com/labstack/echo/v4 v4.11.0\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "echo", fw.Name)
	assert.Equal(t, "go", fw.Stack)
}

// TestDetectFramework_ChiFromGoMod tests Chi router detection from go.mod.
func TestDetectFramework_ChiFromGoMod(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	goMod := "module example.com/test\n\ngo 1.21\n\nrequire github.com/go-chi/chi/v5 v5.0.0\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "chi", fw.Name)
	assert.Equal(t, "go", fw.Stack)
}

// TestDetectFramework_SignalFieldIsSet verifies the Signal field is populated in results.
func TestDetectFramework_SignalFieldIsSet(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "next.config.ts"), []byte(""), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.NotEmpty(t, fw.Signal, "Framework.Signal must not be empty")
}
