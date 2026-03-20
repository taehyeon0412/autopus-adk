package setup

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insajin/autopus-adk/pkg/arch"
	"github.com/insajin/autopus-adk/pkg/lore"
)

func TestRender_IndexLineLimit(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{
		Name: "test-project",
		Languages: []Language{
			{Name: "Go", Version: "1.23"},
		},
	}

	ds := Render(info, nil)
	lines := strings.Split(ds.Index, "\n")
	assert.LessOrEqual(t, len(lines), maxIndexLines, "index.md should not exceed %d lines", maxIndexLines)
}

func TestRender_IndexContainsProjectInfo(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{
		Name: "my-project",
		Languages: []Language{
			{Name: "Go", Version: "1.23"},
			{Name: "TypeScript", Version: "5.0"},
		},
		EntryPoints: []EntryPoint{
			{Path: "cmd/app/main.go", Description: "CLI entry point"},
		},
	}

	ds := Render(info, nil)
	assert.Contains(t, ds.Index, "my-project")
	assert.Contains(t, ds.Index, "Go")
	assert.Contains(t, ds.Index, "TypeScript")
	assert.Contains(t, ds.Index, "cmd/app/main.go")
}

func TestRender_CommandsNoBuildFiles(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{Name: "empty"}

	ds := Render(info, nil)
	assert.Contains(t, ds.Commands, "No build files detected")
}

func TestRender_CommandsWithMakefile(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{
		Name: "test",
		BuildFiles: []BuildFile{
			{
				Path: "Makefile",
				Type: "makefile",
				Commands: map[string]string{
					"build": "make build",
					"test":  "make test",
				},
			},
		},
	}

	ds := Render(info, nil)
	assert.Contains(t, ds.Commands, "make build")
	assert.Contains(t, ds.Commands, "make test")
}

func TestRender_AllDocsUnderLineLimit(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{
		Name: "test",
		Languages: []Language{
			{Name: "Go", Version: "1.23"},
		},
		BuildFiles: []BuildFile{
			{Path: "go.mod", Type: "go.mod", Commands: map[string]string{"build": "go build ./..."}},
		},
		TestConfig: TestConfiguration{
			Framework: "go test",
			Command:   "go test ./...",
		},
	}

	ds := Render(info, nil)

	docs := map[string]string{
		"index":        ds.Index,
		"commands":     ds.Commands,
		"structure":    ds.Structure,
		"conventions":  ds.Conventions,
		"boundaries":   ds.Boundaries,
		"architecture": ds.Architecture,
		"testing":      ds.Testing,
	}

	for name, content := range docs {
		lines := strings.Split(content, "\n")
		limit := maxDocLines
		if name == "index" {
			limit = maxIndexLines
		}
		assert.LessOrEqual(t, len(lines), limit, "%s exceeds line limit", name)
	}
}

func TestRender_WithArchMap(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{Name: "test"}
	opts := &RenderOptions{
		ArchMap: &arch.ArchitectureMap{
			Layers: []arch.Layer{
				{Name: "cmd", Level: 3, AllowedDeps: []string{"pkg", "internal"}},
				{Name: "pkg", Level: 2, AllowedDeps: []string{"pkg"}},
			},
			Domains: []arch.Domain{
				{Name: "api", Path: "pkg/api", Description: "API handlers"},
			},
		},
	}

	ds := Render(info, opts)
	assert.Contains(t, ds.Architecture, "cmd")
	assert.Contains(t, ds.Architecture, "pkg")
	assert.Contains(t, ds.Architecture, "api")
}

func TestRender_WithLoreItems(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{Name: "test"}
	opts := &RenderOptions{
		LoreItems: []lore.LoreEntry{
			{
				CommitMsg:  "Use Cobra for CLI framework",
				Constraint: "Must support subcommands",
				Rejected:   "urfave/cli - less flexible",
			},
		},
	}

	ds := Render(info, opts)
	assert.Contains(t, ds.Architecture, "Use Cobra")
	assert.Contains(t, ds.Architecture, "Must support subcommands")
}

func TestRender_BoundariesStructure(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{
		Name:      "test",
		Languages: []Language{{Name: "Go"}},
	}

	ds := Render(info, nil)
	assert.Contains(t, ds.Boundaries, "Always Do")
	assert.Contains(t, ds.Boundaries, "Ask First")
	assert.Contains(t, ds.Boundaries, "Never Do")
}

func TestRender_TestingWithFramework(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{
		Name:      "test",
		Languages: []Language{{Name: "Go"}},
		TestConfig: TestConfiguration{
			Framework:  "go test",
			Command:    "go test -race ./...",
			Dirs:       []string{"pkg/setup", "internal/cli"},
			CoverageOn: true,
		},
	}

	ds := Render(info, nil)
	assert.Contains(t, ds.Testing, "go test")
	assert.Contains(t, ds.Testing, "Coverage")
}

func TestTruncateLines(t *testing.T) {
	t.Parallel()

	lines := strings.Repeat("line\n", 300)
	result := truncateLines(lines, 200)
	assert.LessOrEqual(t, strings.Count(result, "\n"), 200)
	assert.Contains(t, result, "Truncated")
}
