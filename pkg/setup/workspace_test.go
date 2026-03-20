package setup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectWorkspaces_GoWork(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "go.work", `go 1.23

use (
	./api
	./worker
	./shared
)
`)
	writeFile(t, dir, "api/go.mod", "module example/api\n")
	writeFile(t, dir, "worker/go.mod", "module example/worker\n")
	writeFile(t, dir, "shared/go.mod", "module example/shared\n")

	workspaces := DetectWorkspaces(dir)
	require.Len(t, workspaces, 3)

	names := make(map[string]bool)
	for _, ws := range workspaces {
		names[ws.Name] = true
		assert.Equal(t, "go.work", ws.Type)
	}
	assert.True(t, names["api"])
	assert.True(t, names["worker"])
	assert.True(t, names["shared"])
}

func TestDetectWorkspaces_GoWork_SingleLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "go.work", "go 1.23\n\nuse ./mylib\n")
	writeFile(t, dir, "mylib/go.mod", "module example/mylib\n")

	workspaces := DetectWorkspaces(dir)
	require.Len(t, workspaces, 1)
	assert.Equal(t, "mylib", workspaces[0].Name)
}

func TestDetectWorkspaces_NPM(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "package.json", `{
		"workspaces": ["packages/*"]
	}`)
	writeFile(t, dir, "packages/core/package.json", `{"name": "core"}`)
	writeFile(t, dir, "packages/cli/package.json", `{"name": "cli"}`)

	workspaces := DetectWorkspaces(dir)
	require.Len(t, workspaces, 2)

	names := make(map[string]bool)
	for _, ws := range workspaces {
		names[ws.Name] = true
		assert.Equal(t, "npm", ws.Type)
	}
	assert.True(t, names["core"])
	assert.True(t, names["cli"])
}

func TestDetectWorkspaces_Yarn(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "package.json", `{
		"workspaces": {
			"packages": ["packages/*"]
		}
	}`)
	writeFile(t, dir, "yarn.lock", "")
	writeFile(t, dir, "packages/ui/package.json", `{"name": "ui"}`)

	workspaces := DetectWorkspaces(dir)
	require.Len(t, workspaces, 1)
	assert.Equal(t, "ui", workspaces[0].Name)
	assert.Equal(t, "yarn", workspaces[0].Type)
}

func TestDetectWorkspaces_Pnpm(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "package.json", `{"name": "root"}`)
	writeFile(t, dir, "pnpm-workspace.yaml", "packages:\n  - 'apps/*'\n  - 'libs/*'\n")
	writeFile(t, dir, "apps/web/package.json", `{"name": "web"}`)
	writeFile(t, dir, "libs/shared/package.json", `{"name": "shared"}`)

	workspaces := DetectWorkspaces(dir)
	require.Len(t, workspaces, 2)

	names := make(map[string]bool)
	for _, ws := range workspaces {
		names[ws.Name] = true
		assert.Equal(t, "pnpm", ws.Type)
	}
	assert.True(t, names["web"])
	assert.True(t, names["shared"])
}

func TestDetectWorkspaces_CargoWorkspace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "Cargo.toml", `[workspace]
members = [
    "crates/core",
    "crates/cli",
]
`)
	writeFile(t, dir, "crates/core/Cargo.toml", "[package]\nname = \"core\"\n")
	writeFile(t, dir, "crates/cli/Cargo.toml", "[package]\nname = \"cli\"\n")

	workspaces := DetectWorkspaces(dir)
	require.Len(t, workspaces, 2)

	names := make(map[string]bool)
	for _, ws := range workspaces {
		names[ws.Name] = true
		assert.Equal(t, "cargo", ws.Type)
	}
	assert.True(t, names["core"])
	assert.True(t, names["cli"])
}

func TestDetectWorkspaces_CargoWorkspaceGlob(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "Cargo.toml", "[workspace]\nmembers = [\"crates/*\"]\n")
	writeFile(t, dir, "crates/lib-a/Cargo.toml", "[package]\nname = \"lib-a\"\n")
	writeFile(t, dir, "crates/lib-b/Cargo.toml", "[package]\nname = \"lib-b\"\n")

	workspaces := DetectWorkspaces(dir)
	require.Len(t, workspaces, 2)
}

func TestDetectWorkspaces_NoWorkspace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "go.mod", "module example\n\ngo 1.23\n")

	workspaces := DetectWorkspaces(dir)
	assert.Empty(t, workspaces)
}

func TestRender_IndexWithWorkspaces(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{
		Name:      "monorepo",
		Languages: []Language{{Name: "Go"}},
		Workspaces: []Workspace{
			{Name: "api", Path: "api", Type: "go.work"},
			{Name: "worker", Path: "worker", Type: "go.work"},
		},
	}

	ds := Render(info, nil)
	assert.Contains(t, ds.Index, "Workspaces")
	assert.Contains(t, ds.Index, "monorepo")
	assert.Contains(t, ds.Index, "api")
	assert.Contains(t, ds.Index, "worker")
	assert.Contains(t, ds.Index, "go.work")
}
