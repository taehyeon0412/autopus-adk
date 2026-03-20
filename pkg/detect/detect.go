// Package detect는 코딩 CLI 바이너리와 의존성의 설치 여부를 감지한다.
package detect

import (
	"os/exec"
	"strings"
)

// Platform은 감지된 코딩 CLI 정보이다.
type Platform struct {
	Name    string // claude-code, codex, gemini-cli 등
	Binary  string // 실행 파일명
	Version string // 감지된 버전
}

// knownCLIs는 알려진 코딩 CLI 목록이다.
var knownCLIs = []struct {
	name       string
	binary     string
	versionArg string
}{
	{"claude-code", "claude", "--version"},
	{"codex", "codex", "--version"},
	{"gemini-cli", "gemini", "--version"},
	{"opencode", "opencode", "--version"},
	{"cursor", "cursor", "--version"},
}

// DetectPlatforms는 PATH에서 코딩 CLI를 감지한다.
func DetectPlatforms() []Platform {
	var platforms []Platform
	for _, cli := range knownCLIs {
		if v, ok := detectBinary(cli.binary, cli.versionArg); ok {
			platforms = append(platforms, Platform{
				Name:    cli.name,
				Binary:  cli.binary,
				Version: v,
			})
		}
	}
	return platforms
}

// IsInstalled는 특정 바이너리의 설치 여부를 확인한다.
func IsInstalled(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}

func detectBinary(binary, versionArg string) (string, bool) {
	path, err := exec.LookPath(binary)
	if err != nil {
		return "", false
	}
	_ = path
	out, err := exec.Command(binary, versionArg).Output()
	if err != nil {
		return "unknown", true
	}
	return strings.TrimSpace(string(out)), true
}

// Dependency는 외부 도구 의존성이다.
type Dependency struct {
	Name        string
	Binary      string
	InstallCmd  string
	Required    bool // true이면 필수, false이면 권장
	Description string
}

// FullModeDeps는 Full 모드의 의존성 목록이다.
var FullModeDeps = []Dependency{
	{Name: "ast-grep", Binary: "sg", InstallCmd: "npm i -g @ast-grep/cli", Required: false, Description: "Structural code search"},
	{Name: "playwright", Binary: "playwright", InstallCmd: "npm i -g playwright", Required: false, Description: "E2E testing + screenshots"},
	{Name: "agent-browser", Binary: "agent-browser", InstallCmd: "npm i -g @anthropic-ai/agent-browser", Required: false, Description: "Web browsing"},
	{Name: "gh", Binary: "gh", InstallCmd: "", Required: false, Description: "GitHub CLI"},
}

// CheckDependencies는 의존성 상태를 확인한다.
func CheckDependencies(deps []Dependency) []DependencyStatus {
	var statuses []DependencyStatus
	for _, d := range deps {
		statuses = append(statuses, DependencyStatus{
			Dependency: d,
			Installed:  IsInstalled(d.Binary),
		})
	}
	return statuses
}

// DependencyStatus는 의존성 상태이다.
type DependencyStatus struct {
	Dependency
	Installed bool
}
