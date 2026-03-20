package arch_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/arch"
)

func TestAnalyze_GoProject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Go 프로젝트 구조 생성
	dirs := []string{
		"cmd/app",
		"pkg/service",
		"pkg/model",
		"internal/repo",
	}
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, d), 0o755))
	}

	// go.mod 생성
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/myapp\n\ngo 1.23\n"), 0o644))

	// Go 소스 파일 생성 (import 포함)
	cmdMain := `package main

import (
	"example.com/myapp/pkg/service"
	"example.com/myapp/internal/repo"
)

func main() {
	_ = service.New()
	_ = repo.New()
}
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "cmd/app/main.go"), []byte(cmdMain), 0o644))

	pkgService := `package service

import "example.com/myapp/pkg/model"

func New() *Service { return &Service{} }
type Service struct{}
func (s *Service) Do(m *model.M) {}
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg/service/service.go"), []byte(pkgService), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg/model/model.go"), []byte("package model\n\ntype M struct{}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "internal/repo/repo.go"), []byte("package repo\n\nfunc New() *Repo { return &Repo{} }\ntype Repo struct{}\n"), 0o644))

	archMap, err := arch.Analyze(dir)
	require.NoError(t, err)
	require.NotNil(t, archMap)

	// 레이어 감지 확인
	assert.NotEmpty(t, archMap.Layers)

	// 도메인 감지 확인
	assert.NotEmpty(t, archMap.Domains)

	// 의존성 감지 확인
	assert.NotEmpty(t, archMap.Dependencies)
}

func TestAnalyze_TSProject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	dirs := []string{
		"src/components",
		"src/services",
		"src/models",
	}
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, d), 0o755))
	}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"myapp","version":"1.0.0"}`), 0o644))

	tsFile := `import { UserService } from './services/UserService';
import { UserModel } from './models/User';

export class App {
	private service = new UserService();
}
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "src/App.ts"), []byte(tsFile), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "src/services/UserService.ts"), []byte("export class UserService {}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "src/models/User.ts"), []byte("export class UserModel {}\n"), 0o644))

	archMap, err := arch.Analyze(dir)
	require.NoError(t, err)
	require.NotNil(t, archMap)

	assert.NotEmpty(t, archMap.Domains)
}

func TestAnalyze_PythonProject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	dirs := []string{
		"myapp/services",
		"myapp/models",
		"tests",
	}
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, d), 0o755))
	}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "setup.py"), []byte("from setuptools import setup\nsetup(name='myapp')\n"), 0o644))

	pyFile := `from myapp.services import UserService
from myapp.models import User

def main():
    svc = UserService()
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "myapp/__init__.py"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "myapp/main.py"), []byte(pyFile), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "myapp/services/__init__.py"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "myapp/services/user.py"), []byte("class UserService:\n    pass\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "myapp/models/__init__.py"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "myapp/models/user.py"), []byte("class User:\n    pass\n"), 0o644))

	archMap, err := arch.Analyze(dir)
	require.NoError(t, err)
	require.NotNil(t, archMap)

	assert.NotEmpty(t, archMap.Domains)
}

func TestAnalyze_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	archMap, err := arch.Analyze(dir)
	require.NoError(t, err)
	require.NotNil(t, archMap)
	// 빈 디렉터리는 빈 맵 반환
	assert.Empty(t, archMap.Dependencies)
}

func TestAnalyze_NonExistentDir(t *testing.T) {
	t.Parallel()

	_, err := arch.Analyze("/nonexistent/path/xyz")
	assert.Error(t, err)
}
