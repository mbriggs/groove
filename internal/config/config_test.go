package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectRoot(t *testing.T) {
	// Create a temp dir tree: root/.groove.yml, root/sub/deep/
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".groove.yml"), []byte("ports: {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "sub", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	// Should find root from deep subdirectory
	found, err := DetectProjectRoot(sub)
	if err != nil {
		t.Fatalf("DetectProjectRoot() error: %v", err)
	}
	if found != root {
		t.Errorf("DetectProjectRoot() = %q, want %q", found, root)
	}

	// Should find root from root itself
	found, err = DetectProjectRoot(root)
	if err != nil {
		t.Fatalf("DetectProjectRoot() error: %v", err)
	}
	if found != root {
		t.Errorf("DetectProjectRoot() = %q, want %q", found, root)
	}
}

func TestDetectProjectRootNotFound(t *testing.T) {
	tmp := t.TempDir()
	_, err := DetectProjectRoot(tmp)
	if err == nil {
		t.Fatal("expected error when no .groove.yml exists")
	}
}

func TestLoadProjectParsesConfig(t *testing.T) {
	tmp := t.TempDir()
	yml := `ports:
  web: ~
  db: 5432
hooks:
  open: ./open.sh
  archive: ./archive.sh
`
	if err := os.WriteFile(filepath.Join(tmp, ".groove.yml"), []byte(yml), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadProject(tmp)
	if err != nil {
		t.Fatalf("LoadProject() error: %v", err)
	}

	if len(cfg.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(cfg.Ports))
	}
	if cfg.Ports["web"] != nil {
		t.Error("web port should be nil (free)")
	}
	if cfg.Ports["db"] == nil || *cfg.Ports["db"] != 5432 {
		t.Error("db port should be 5432")
	}
	if cfg.Hooks.Open != "./open.sh" {
		t.Errorf("hooks.open = %q, want %q", cfg.Hooks.Open, "./open.sh")
	}
	if cfg.Hooks.Archive != "./archive.sh" {
		t.Errorf("hooks.archive = %q, want %q", cfg.Hooks.Archive, "./archive.sh")
	}
}

func TestLoadGlobalDefaults(t *testing.T) {
	// Point XDG_CONFIG_HOME to a temp dir with no config file
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error: %v", err)
	}

	if cfg.DefaultRemote != "origin" {
		t.Errorf("DefaultRemote = %q, want %q", cfg.DefaultRemote, "origin")
	}
	if cfg.WorktreeRoot != "" {
		t.Errorf("WorktreeRoot = %q, want empty", cfg.WorktreeRoot)
	}
}

func TestDetectShell(t *testing.T) {
	// Explicit config takes precedence
	cfg := &Global{Shell: "fish"}
	if got := DetectShell(cfg); got != "fish" {
		t.Errorf("DetectShell() = %q, want %q", got, "fish")
	}

	// Falls back to $SHELL
	t.Setenv("SHELL", "/bin/bash")
	cfg = &Global{}
	if got := DetectShell(cfg); got != "bash" {
		t.Errorf("DetectShell() = %q, want %q", got, "bash")
	}

	// Falls back to zsh
	t.Setenv("SHELL", "")
	if got := DetectShell(cfg); got != "zsh" {
		t.Errorf("DetectShell() = %q, want %q", got, "zsh")
	}
}

func TestExpandWorktreeRoot(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}

	cfg := &Global{WorktreeRoot: "~/groove"}
	got, err := cfg.ExpandWorktreeRoot()
	if err != nil {
		t.Fatalf("ExpandWorktreeRoot() error: %v", err)
	}
	want := filepath.Join(home, "groove")
	if got != want {
		t.Errorf("ExpandWorktreeRoot() = %q, want %q", got, want)
	}
}
