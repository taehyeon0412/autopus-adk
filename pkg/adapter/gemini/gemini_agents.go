// Package gemini provides agent content file management for Gemini CLI.
package gemini

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	contentfs "github.com/insajin/autopus-adk/content"
	"github.com/insajin/autopus-adk/pkg/adapter"
)

// renderAgentFiles copies agent content files from embedded FS to
// .gemini/agents/autopus/ and returns file mappings.
func (a *Adapter) renderAgentFiles() ([]adapter.FileMapping, error) {
	targetRelDir := filepath.Join(".gemini", "agents", "autopus")
	absTargetDir := filepath.Join(a.root, targetRelDir)
	if err := os.MkdirAll(absTargetDir, 0755); err != nil {
		return nil, fmt.Errorf("gemini agents 디렉터리 생성 실패: %w", err)
	}

	mappings, err := a.prepareAgentMappings()
	if err != nil {
		return nil, err
	}

	for _, m := range mappings {
		destPath := filepath.Join(a.root, m.TargetPath)
		if err := os.WriteFile(destPath, m.Content, 0644); err != nil {
			return nil, fmt.Errorf("gemini agent 파일 쓰기 실패 %s: %w", destPath, err)
		}
	}

	return mappings, nil
}

// prepareAgentMappings reads agent content files and returns file mappings
// without writing to disk.
func (a *Adapter) prepareAgentMappings() ([]adapter.FileMapping, error) {
	var files []adapter.FileMapping

	entries, err := contentfs.FS.ReadDir("agents")
	if err != nil {
		return nil, fmt.Errorf("에이전트 컨텐츠 디렉터리 읽기 실패: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := fs.ReadFile(contentfs.FS, "agents/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("에이전트 파일 읽기 실패 %s: %w", entry.Name(), err)
		}

		relPath := filepath.Join(".gemini", "agents", "autopus", entry.Name())
		files = append(files, adapter.FileMapping{
			TargetPath:      relPath,
			OverwritePolicy: adapter.OverwriteAlways,
			Checksum:        checksum(string(data)),
			Content:         data,
		})
	}

	return files, nil
}
