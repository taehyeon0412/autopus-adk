package content

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadProfilesFromFS_SingleGoProfile tests S1: load a single Go builtin profile.
// R1: Load profiles from embedded FS — parse frontmatter + body into ProfileDefinition.
func TestLoadProfilesFromFS_SingleGoProfile(t *testing.T) {
	t.Parallel()

	goProfileContent := `---
name: go
stack: go
framework: ""
tools:
  - gopls
  - golangci-lint
test_framework: go test
linter: golangci-lint
---
# Go Executor Profile

Use Go idioms and standard library conventions.
`
	fsys := fstest.MapFS{
		"profiles/executor/go.md": &fstest.MapFile{
			Data: []byte(goProfileContent),
		},
	}

	profiles, err := LoadProfilesFromFS(fsys, "profiles/executor")
	require.NoError(t, err)
	require.Len(t, profiles, 1)

	p := profiles[0]
	assert.Equal(t, "go", p.Name)
	assert.Equal(t, "go", p.Stack)
	assert.Equal(t, "go test", p.TestFramework)
	assert.Equal(t, "golangci-lint", p.Linter)
	assert.NotEmpty(t, p.Instructions)
}

// TestLoadProfilesFromFS_AllFiveBuiltins tests S2: load all 5 builtin profiles.
// R1: Load profiles from embedded FS — parse frontmatter + body into ProfileDefinition.
func TestLoadProfilesFromFS_AllFiveBuiltins(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"profiles/executor/go.md": &fstest.MapFile{
			Data: []byte("---\nname: go\nstack: go\n---\nGo profile"),
		},
		"profiles/executor/typescript.md": &fstest.MapFile{
			Data: []byte("---\nname: typescript\nstack: typescript\n---\nTS profile"),
		},
		"profiles/executor/nextjs.md": &fstest.MapFile{
			Data: []byte("---\nname: nextjs\nstack: typescript\nframework: nextjs\nextends: typescript\n---\nNext.js profile"),
		},
		"profiles/executor/python.md": &fstest.MapFile{
			Data: []byte("---\nname: python\nstack: python\n---\nPython profile"),
		},
		"profiles/executor/fastapi.md": &fstest.MapFile{
			Data: []byte("---\nname: fastapi\nstack: python\nframework: fastapi\nextends: python\n---\nFastAPI profile"),
		},
	}

	profiles, err := LoadProfilesFromFS(fsys, "profiles/executor")
	require.NoError(t, err)
	assert.Len(t, profiles, 5)
}

// TestLoadProfilesFromFS_ParsesFrontmatterAndBody tests R1: parse frontmatter + body.
func TestLoadProfilesFromFS_ParsesFrontmatterAndBody(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"profiles/executor/ts.md": &fstest.MapFile{
			Data: []byte("---\nname: typescript\nstack: typescript\ntools:\n  - tsc\n  - eslint\nlinter: eslint\ntest_framework: jest\n---\n# TS Profile\nUse strict TypeScript."),
		},
	}

	profiles, err := LoadProfilesFromFS(fsys, "profiles/executor")
	require.NoError(t, err)
	require.Len(t, profiles, 1)

	p := profiles[0]
	assert.Equal(t, "typescript", p.Name)
	assert.Equal(t, "typescript", p.Stack)
	assert.Contains(t, p.Tools, "tsc")
	assert.Equal(t, "eslint", p.Linter)
	assert.Equal(t, "jest", p.TestFramework)
	assert.Contains(t, p.Instructions, "TS Profile")
}

// TestLoadProfilesFromFS_MissingNameSkipsWithWarning tests R11: missing required field → skip with warning.
// Also covers S12: Missing name field → skip with warning.
func TestLoadProfilesFromFS_MissingNameSkipsWithWarning(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		// Profile without required name field
		"profiles/executor/noname.md": &fstest.MapFile{
			Data: []byte("---\nstack: go\n---\nNo name here"),
		},
		"profiles/executor/valid.md": &fstest.MapFile{
			Data: []byte("---\nname: valid\nstack: go\n---\nValid profile"),
		},
	}

	profiles, warnings, err := LoadProfilesFromFSWithWarnings(fsys, "profiles/executor")
	require.NoError(t, err)
	// valid profile is loaded, noname is skipped
	assert.Len(t, profiles, 1)
	assert.Equal(t, "valid", profiles[0].Name)
	assert.NotEmpty(t, warnings, "missing name must produce a warning")
}

// TestLoadProfilesFromFSWithWarnings_MissingStackSkips verifies missing stack also generates a warning.
func TestLoadProfilesFromFSWithWarnings_MissingStackSkips(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"profiles/executor/nostack.md": &fstest.MapFile{
			Data: []byte("---\nname: nostack\n---\nNo stack here"),
		},
		"profiles/executor/valid.md": &fstest.MapFile{
			Data: []byte("---\nname: valid\nstack: go\n---\nValid profile"),
		},
	}

	profiles, warnings, err := LoadProfilesFromFSWithWarnings(fsys, "profiles/executor")
	require.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.NotEmpty(t, warnings)
}

// TestLoadProfilesFromFS_InvalidDir tests error from nonexistent embedded directory.
func TestLoadProfilesFromFS_InvalidDir(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{}
	_, err := LoadProfilesFromFS(fsys, "nonexistent/dir")
	assert.Error(t, err)
}

// TestLoadProfilesFromFS_SkipsSubdirectories verifies directories are not processed as profiles.
func TestLoadProfilesFromFS_SkipsSubdirectories(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"profiles/executor/go.md":         &fstest.MapFile{Data: []byte("---\nname: go\nstack: go\n---\nbody")},
		"profiles/executor/subdir/sub.md": &fstest.MapFile{Data: []byte("---\nname: sub\nstack: go\n---\nbody")},
	}

	profiles, err := LoadProfilesFromFS(fsys, "profiles/executor")
	require.NoError(t, err)
	// Only root-level .md should be loaded; subdirectory entries skipped
	assert.Len(t, profiles, 1)
	assert.Equal(t, "go", profiles[0].Name)
}

// TestLoadProfilesFromFSWithWarnings_InvalidYAML tests YAML parse error in frontmatter.
func TestLoadProfilesFromFSWithWarnings_InvalidYAML(t *testing.T) {
	t.Parallel()

	// Unclosed bracket in YAML causes unmarshal failure
	fsys := fstest.MapFS{
		"profiles/executor/bad.md": &fstest.MapFile{
			Data: []byte("---\nname: [unclosed\nstack: go\n---\nbody"),
		},
	}

	_, _, err := LoadProfilesFromFSWithWarnings(fsys, "profiles/executor")
	assert.Error(t, err, "invalid YAML frontmatter must return an error")
}

// TestProfileDefinition_SourceField verifies the Source field is set correctly from LoadProfilesFromFS.
func TestProfileDefinition_SourceField(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"profiles/executor/go.md": &fstest.MapFile{
			Data: []byte("---\nname: go\nstack: go\n---\nGo profile"),
		},
	}

	profiles, err := LoadProfilesFromFS(fsys, "profiles/executor")
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	assert.Equal(t, "builtin", profiles[0].Source)
}

// TestLoadCustomProfileDir tests R8: custom profile directory loading.
func TestLoadCustomProfileDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	customContent := "---\nname: my-go\nstack: go\nlinter: custom-linter\n---\nCustom Go profile"

	require.NoError(t, os.WriteFile(dir+"/my-go.md", []byte(customContent), 0644))

	profiles, err := LoadProfilesFromDir(dir)
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	assert.Equal(t, "my-go", profiles[0].Name)
	assert.Equal(t, "custom-linter", profiles[0].Linter)
}

// TestLoadCustomProfileDir_SkipsNonMarkdown verifies non-.md files are ignored.
func TestLoadCustomProfileDir_SkipsNonMarkdown(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(dir+"/profile.md", []byte("---\nname: valid\nstack: go\n---\nbody"), 0644))
	require.NoError(t, os.WriteFile(dir+"/readme.txt", []byte("ignored"), 0644))
	require.NoError(t, os.WriteFile(dir+"/config.yaml", []byte("ignored: true"), 0644))

	profiles, err := LoadProfilesFromDir(dir)
	require.NoError(t, err)
	assert.Len(t, profiles, 1, "only .md files must be loaded")
}

// TestLoadCustomProfileDir_NonexistentDir tests error handling for missing directory.
func TestLoadCustomProfileDir_NonexistentDir(t *testing.T) {
	t.Parallel()

	_, err := LoadProfilesFromDir("/nonexistent/path/xyz/abc")
	assert.Error(t, err, "missing directory must return an error")
}

// TestLoadProfilesFromDir_MultipleProfiles verifies multiple profiles are loaded from a directory.
func TestLoadProfilesFromDir_MultipleProfiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files := map[string]string{
		"go.md":     "---\nname: go\nstack: go\n---\nGo profile",
		"python.md": "---\nname: python\nstack: python\n---\nPython profile",
		"ts.md":     "---\nname: typescript\nstack: typescript\n---\nTS profile",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(dir+"/"+name, []byte(content), 0644))
	}

	profiles, err := LoadProfilesFromDir(dir)
	require.NoError(t, err)
	assert.Len(t, profiles, 3)
}

// TestLoadProfilesFromDir_SourceFieldIsCustom verifies Source="custom" for dir-loaded profiles.
func TestLoadProfilesFromDir_SourceFieldIsCustom(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(dir+"/go.md", []byte("---\nname: go\nstack: go\n---\nbody"), 0644))

	profiles, err := LoadProfilesFromDir(dir)
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	assert.Equal(t, "custom", profiles[0].Source)
}

// TestLoadProfilesFromDir_InvalidYAML tests YAML parse error returns error from LoadProfilesFromDir.
func TestLoadProfilesFromDir_InvalidYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(dir+"/bad.md", []byte("---\nname: [unclosed\nstack: go\n---\nbody"), 0644))

	_, err := LoadProfilesFromDir(dir)
	assert.Error(t, err, "invalid YAML frontmatter must return an error")
}
