package resolve

import (
	"fmt"
	"os"
	"strings"

	"github.com/mbriggs/groove/internal/config"
	"github.com/mbriggs/groove/internal/git"
	"github.com/mbriggs/groove/internal/state"
)

// ProjectInfo holds resolved project information.
type ProjectInfo struct {
	Root      string
	Name      string
	RemoteURL string
	Config    *config.Project
}

// Project resolves the current project from env or cwd.
func Project(globalCfg *config.Global) (*ProjectInfo, error) {
	// Check env var first
	if root := os.Getenv("GROOVE_PROJECT_ROOT"); root != "" {
		return projectFromRoot(root, globalCfg)
	}

	// Walk up looking for .groove.yml
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	root, err := config.DetectProjectRoot(cwd)
	if err != nil {
		return nil, err
	}

	return projectFromRoot(root, globalCfg)
}

func projectFromRoot(root string, globalCfg *config.Global) (*ProjectInfo, error) {
	bare, err := git.IsBareRepo(root)
	if err == nil && bare {
		return nil, fmt.Errorf("bare git repository detected at %s — groove requires a regular repository", root)
	}

	remote := globalCfg.DefaultRemote
	remoteURL, err := git.RemoteURL(root, remote)
	if err != nil {
		return nil, err
	}

	name := git.ProjectNameFromRemote(remoteURL)

	cfg, err := config.LoadProject(root)
	if err != nil {
		return nil, err
	}

	return &ProjectInfo{
		Root:      root,
		Name:      name,
		RemoteURL: remoteURL,
		Config:    cfg,
	}, nil
}

// CurrentWorktree finds the current worktree from env or cwd.
func CurrentWorktree(store *state.Store) (*state.Worktree, error) {
	// Check env var first
	if id := os.Getenv("GROOVE_WORKTREE_ID"); id != "" {
		wt, ok := store.Worktrees[id]
		if !ok {
			return nil, fmt.Errorf("worktree %s from GROOVE_WORKTREE_ID not found in state", id)
		}
		return wt, nil
	}

	// Try to match cwd to a worktree path
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for _, wt := range store.Worktrees {
		if strings.HasPrefix(cwd, wt.Path) {
			return wt, nil
		}
	}

	return nil, fmt.Errorf("not inside a groove worktree (set GROOVE_WORKTREE_ID or cd into a worktree)")
}

// BuildEnvVars creates the env vars map for a worktree.
func BuildEnvVars(wt *state.Worktree) map[string]string {
	env := map[string]string{
		"GROOVE_PROJECT":       wt.Project,
		"GROOVE_WORKTREE_ID":   wt.ID,
		"GROOVE_WORKTREE_PATH": wt.Path,
		"GROOVE_PROJECT_ROOT":  wt.ProjectRoot,
		"GROOVE_BRANCH":        wt.Branch,
		"GROOVE_DEFAULT_BRANCH": wt.DefaultBranch,
	}

	for name, port := range wt.Ports {
		key := "GROOVE_PORT_" + strings.ToUpper(name)
		env[key] = fmt.Sprintf("%d", port)
	}

	return env
}
