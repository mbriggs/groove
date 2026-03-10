package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Project represents a .groove.yml file at the repo root.
type Project struct {
	Ports map[string]*int `yaml:"ports"`
	Hooks struct {
		Open    string `yaml:"open"`
		Archive string `yaml:"archive"`
	} `yaml:"hooks"`
}

// Global represents ~/.config/groove/config.yml.
type Global struct {
	WorktreeRoot  string `yaml:"worktree_root"`
	DefaultRemote string `yaml:"default_remote"`
	Shell         string `yaml:"shell"`
	BranchPrefix  string `yaml:"branch_prefix"`
}

// LoadProject reads the .groove.yml file from the given directory.
func LoadProject(dir string) (*Project, error) {
	path := filepath.Join(dir, ".groove.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading project config: %w", err)
	}
	var cfg Project
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing .groove.yml: %w", err)
	}
	if cfg.Ports == nil {
		cfg.Ports = make(map[string]*int)
	}
	return &cfg, nil
}

// LoadGlobal reads the global config file.
func LoadGlobal() (*Global, error) {
	cfg := &Global{
		DefaultRemote: "origin",
	}

	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return cfg, nil
		}
		configDir = filepath.Join(home, ".config")
	}

	path := filepath.Join(configDir, "groove", "config.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading global config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing global config: %w", err)
	}

	if cfg.DefaultRemote == "" {
		cfg.DefaultRemote = "origin"
	}

	return cfg, nil
}

// ExpandWorktreeRoot expands ~ in the worktree root path.
func (g *Global) ExpandWorktreeRoot() (string, error) {
	root := g.WorktreeRoot
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, "groove")
	}

	if len(root) > 0 && root[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, root[1:])
	}

	return root, nil
}

// DetectProjectRoot walks up from dir looking for .groove.yml.
func DetectProjectRoot(dir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".groove.yml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .groove.yml found in current directory or any parent")
		}
		dir = parent
	}
}

// DetectShell returns the user's shell for layout generation.
func DetectShell(globalCfg *Global) string {
	if globalCfg != nil && globalCfg.Shell != "" {
		return globalCfg.Shell
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		return filepath.Base(shell)
	}
	return "zsh"
}
