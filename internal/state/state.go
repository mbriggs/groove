package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// Worktree represents a single managed worktree.
type Worktree struct {
	ID               string         `json:"id"`
	Project          string         `json:"project"`
	ProjectRoot      string         `json:"project_root"`
	ProjectRemoteURL string         `json:"project_remote_url"`
	Path             string         `json:"path"`
	Session          string         `json:"session"`
	Branch           string         `json:"branch"`
	DefaultBranch    string         `json:"default_branch"`
	Ports            map[string]int `json:"ports"`
	CreatedAt        time.Time      `json:"created_at"`
}

// Store holds all groove state.
type Store struct {
	Worktrees map[string]*Worktree `json:"worktrees"`
}

func stateFilePath() (string, error) {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataDir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataDir, "groove", "state.json"), nil
}

// Load reads the state file, creating it if needed.
func Load() (*Store, error) {
	path, err := stateFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{Worktrees: make(map[string]*Worktree)}, nil
		}
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}
	if store.Worktrees == nil {
		store.Worktrees = make(map[string]*Worktree)
	}
	return &store, nil
}

// ClaimedPorts returns all ports currently claimed across all worktrees.
func (s *Store) ClaimedPorts() map[int]string {
	ports := make(map[int]string)
	for _, wt := range s.Worktrees {
		for name, port := range wt.Ports {
			ports[port] = wt.ID + ":" + name
		}
	}
	return ports
}

// WorktreesForProject returns all worktrees belonging to a project.
func (s *Store) WorktreesForProject(project string) []*Worktree {
	var result []*Worktree
	for _, wt := range s.Worktrees {
		if wt.Project == project {
			result = append(result, wt)
		}
	}
	return result
}

// LoadLocked reads state with a file lock held, suitable for read-modify-write.
func LoadLocked() (*Store, *os.File, error) {
	path, err := stateFilePath()
	if err != nil {
		return nil, nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, nil, fmt.Errorf("creating state directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("opening state file: %w", err)
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, nil, fmt.Errorf("locking state file: %w", err)
	}

	// Read from the locked fd, not a separate os.ReadFile call
	info, err := f.Stat()
	if err != nil {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
		return nil, nil, fmt.Errorf("stat state file: %w", err)
	}
	data := make([]byte, info.Size())
	if info.Size() > 0 {
		if _, err := f.ReadAt(data, 0); err != nil {
			syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
			f.Close()
			return nil, nil, fmt.Errorf("reading state file: %w", err)
		}
	}

	store := &Store{Worktrees: make(map[string]*Worktree)}
	if len(data) > 0 {
		if err := json.Unmarshal(data, store); err != nil {
			syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
			f.Close()
			return nil, nil, fmt.Errorf("parsing state file: %w", err)
		}
		if store.Worktrees == nil {
			store.Worktrees = make(map[string]*Worktree)
		}
	}

	return store, f, nil
}

// SaveAndUnlock writes state and releases the file lock.
func (s *Store) SaveAndUnlock(f *os.File) error {
	defer func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	return nil
}
