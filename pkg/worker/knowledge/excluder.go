package knowledge

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// alwaysExcluded lists directories and extensions that are always excluded.
var alwaysExcluded = []string{
	".git/",
	"node_modules/",
}

// alwaysExcludedExts lists file extensions that are always excluded.
var alwaysExcludedExts = []string{
	".exe",
	".bin",
}

// Excluder determines whether a file path should be excluded from sync,
// based on .gitignore patterns and built-in exclusion rules.
type Excluder struct {
	patterns []pattern
}

type pattern struct {
	glob     string
	isDir    bool
	negation bool
}

// NewExcluder creates an Excluder by parsing the given .gitignore file.
// If the file does not exist, only built-in exclusions are applied.
func NewExcluder(gitignorePath string) (*Excluder, error) {
	e := &Excluder{}

	f, err := os.Open(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return e, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		p := pattern{}
		if strings.HasPrefix(line, "!") {
			p.negation = true
			line = line[1:]
		}
		if strings.HasSuffix(line, "/") {
			p.isDir = true
			line = strings.TrimSuffix(line, "/")
		}
		p.glob = line
		e.patterns = append(e.patterns, p)
	}

	return e, scanner.Err()
}

// IsExcluded returns true if the given path should be excluded from sync.
func (e *Excluder) IsExcluded(path string) bool {
	// Check built-in directory exclusions.
	normalized := filepath.ToSlash(path)
	for _, dir := range alwaysExcluded {
		if strings.Contains(normalized+"/", "/"+dir) || strings.HasPrefix(normalized, dir) {
			return true
		}
	}

	// Check built-in extension exclusions.
	ext := filepath.Ext(path)
	for _, excluded := range alwaysExcludedExts {
		if strings.EqualFold(ext, excluded) {
			return true
		}
	}

	// Evaluate .gitignore patterns in order; last matching pattern wins.
	excluded := false
	base := filepath.Base(path)
	for _, p := range e.patterns {
		matched, _ := filepath.Match(p.glob, base)
		if !matched {
			matched, _ = filepath.Match(p.glob, normalized)
		}
		if matched {
			excluded = !p.negation
		}
	}

	return excluded
}
