package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadSaveRoundTrip(t *testing.T) {
	// Use a temp dir for XDG_DATA_HOME
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)

	wt := &Worktree{
		ID:               "myapp-my-feature",
		Project:          "myapp",
		ProjectRoot:      "/home/user/code/myapp",
		ProjectRemoteURL: "git@github.com:user/myapp.git",
		Path:             "/home/user/groove/myapp/my-feature",
		Branch:           "mbriggs/my-feature",
		DefaultBranch:    "main",
		Ports:            map[string]int{"web": 3001, "db": 5432},
		CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Save
	store := &Store{Worktrees: map[string]*Worktree{wt.ID: wt}}
	s, f, err := LoadLocked()
	if err != nil {
		t.Fatalf("LoadLocked() error: %v", err)
	}
	s.Worktrees = store.Worktrees
	if err := s.SaveAndUnlock(f); err != nil {
		t.Fatalf("SaveAndUnlock() error: %v", err)
	}

	// Reload and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	got, ok := loaded.Worktrees["myapp-my-feature"]
	if !ok {
		t.Fatal("worktree not found after round-trip")
	}
	if got.Project != "myapp" {
		t.Errorf("Project = %q, want %q", got.Project, "myapp")
	}
	if got.Branch != "mbriggs/my-feature" {
		t.Errorf("Branch = %q, want %q", got.Branch, "mbriggs/my-feature")
	}
	if got.Ports["web"] != 3001 {
		t.Errorf("Ports[web] = %d, want 3001", got.Ports["web"])
	}
	if got.Ports["db"] != 5432 {
		t.Errorf("Ports[db] = %d, want 5432", got.Ports["db"])
	}
}

func TestLoadEmptyState(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)

	store, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(store.Worktrees) != 0 {
		t.Fatalf("expected empty worktrees, got %d", len(store.Worktrees))
	}
}

func TestClaimedPorts(t *testing.T) {
	store := &Store{
		Worktrees: map[string]*Worktree{
			"a": {ID: "a", Ports: map[string]int{"web": 3001, "db": 5432}},
			"b": {ID: "b", Ports: map[string]int{"web": 3002}},
		},
	}

	claimed := store.ClaimedPorts()
	if len(claimed) != 3 {
		t.Fatalf("ClaimedPorts() returned %d entries, want 3", len(claimed))
	}
	if claimed[3001] != "a:web" {
		t.Errorf("claimed[3001] = %q, want %q", claimed[3001], "a:web")
	}
	if claimed[5432] != "a:db" {
		t.Errorf("claimed[5432] = %q, want %q", claimed[5432], "a:db")
	}
	if claimed[3002] != "b:web" {
		t.Errorf("claimed[3002] = %q, want %q", claimed[3002], "b:web")
	}
}

func TestWorktreesForProject(t *testing.T) {
	store := &Store{
		Worktrees: map[string]*Worktree{
			"a-feat1": {ID: "a-feat1", Project: "a"},
			"a-feat2": {ID: "a-feat2", Project: "a"},
			"b-feat1": {ID: "b-feat1", Project: "b"},
		},
	}

	wts := store.WorktreesForProject("a")
	if len(wts) != 2 {
		t.Fatalf("WorktreesForProject(a) returned %d, want 2", len(wts))
	}

	wts = store.WorktreesForProject("b")
	if len(wts) != 1 {
		t.Fatalf("WorktreesForProject(b) returned %d, want 1", len(wts))
	}

	wts = store.WorktreesForProject("c")
	if len(wts) != 0 {
		t.Fatalf("WorktreesForProject(c) returned %d, want 0", len(wts))
	}
}

func TestStateFileCreatedInCorrectLocation(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)

	store, f, err := LoadLocked()
	if err != nil {
		t.Fatalf("LoadLocked() error: %v", err)
	}
	store.Worktrees["test"] = &Worktree{ID: "test"}
	if err := store.SaveAndUnlock(f); err != nil {
		t.Fatalf("SaveAndUnlock() error: %v", err)
	}

	expected := filepath.Join(tmp, "groove", "state.json")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Fatalf("state file not created at %s", expected)
	}
}
