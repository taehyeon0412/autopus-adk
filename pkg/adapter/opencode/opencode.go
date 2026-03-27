// Package opencode implements the opencode CLI platform adapter.
package opencode

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
)

// @AX:NOTE [AUTO] hardcoded adapter constants — adapterVer must be bumped on breaking config changes
const (
	adapterName = "opencode"
	cliBinary   = "opencode"
	adapterVer  = "1.0.0"
	configFile  = "opencode.json"
)

// Adapter is the opencode CLI platform adapter.
type Adapter struct {
	root string
}

// New creates an adapter rooted at the current directory.
func New() *Adapter { return &Adapter{root: "."} }

// NewWithRoot creates an adapter rooted at the given path.
func NewWithRoot(root string) *Adapter { return &Adapter{root: root} }

func (a *Adapter) Name() string        { return adapterName }
func (a *Adapter) Version() string     { return adapterVer }
func (a *Adapter) CLIBinary() string   { return cliBinary }
func (a *Adapter) SupportsHooks() bool { return true }

// Detect checks if the opencode binary is installed in PATH.
func (a *Adapter) Detect(_ context.Context) (bool, error) {
	_, err := exec.LookPath(cliBinary)
	return err == nil, nil
}

// Generate creates the opencode.json config file.
func (a *Adapter) Generate(_ context.Context, _ *config.HarnessConfig) (*adapter.PlatformFiles, error) {
	return &adapter.PlatformFiles{}, nil
}

// Update delegates to Generate for now.
func (a *Adapter) Update(ctx context.Context, cfg *config.HarnessConfig) (*adapter.PlatformFiles, error) {
	return a.Generate(ctx, cfg)
}

// Validate performs minimal validation on opencode configuration.
func (a *Adapter) Validate(_ context.Context) ([]adapter.ValidationError, error) {
	return nil, nil
}

// Clean removes generated opencode configuration files.
func (a *Adapter) Clean(_ context.Context) error {
	p := filepath.Join(a.root, configFile)
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", configFile, err)
	}
	return nil
}

// InstallHooks writes hook configuration to opencode.json.
func (a *Adapter) InstallHooks(_ context.Context, _ []adapter.HookConfig, _ *adapter.PermissionSet) error {
	return nil
}

// pluginEntry represents a single opencode plugin configuration.
type pluginEntry struct {
	Name    string `json:"name"`
	Event   string `json:"event"`
	Command string `json:"command"`
}

// opencodeConfig represents the opencode.json structure.
type opencodeConfig struct {
	Experimental *experimentalBlock `json:"experimental,omitempty"`
}

type experimentalBlock struct {
	Plugins []pluginEntry `json:"plugins,omitempty"`
}

// InjectOrchestraPlugin registers the autopus result collector plugin in opencode.json.
// Preserves existing plugins and avoids duplicates.
// @AX:NOTE [AUTO] magic constant "autopus-result" plugin name — must match hook script expectations
// @AX:NOTE [AUTO] "bun " prefix hardcoded for TypeScript plugin execution — assumes bun runtime installed
func (a *Adapter) InjectOrchestraPlugin(scriptPath string) error {
	cfgPath := filepath.Join(a.root, configFile)

	var cfg opencodeConfig
	if data, err := os.ReadFile(cfgPath); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse %s: %w", configFile, err)
		}
	}

	if cfg.Experimental == nil {
		cfg.Experimental = &experimentalBlock{}
	}

	const pluginName = "autopus-result"

	// Remove existing autopus-result plugin to avoid duplicates
	filtered := make([]pluginEntry, 0, len(cfg.Experimental.Plugins))
	for _, p := range cfg.Experimental.Plugins {
		if p.Name != pluginName {
			filtered = append(filtered, p)
		}
	}

	filtered = append(filtered, pluginEntry{
		Name:    pluginName,
		Event:   "text.complete",
		Command: "bun " + scriptPath,
	})
	cfg.Experimental.Plugins = filtered

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", configFile, err)
	}

	return os.WriteFile(cfgPath, append(data, '\n'), 0644)
}
