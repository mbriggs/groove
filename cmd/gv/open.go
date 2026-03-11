package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mbriggs/groove/internal/config"
	"github.com/mbriggs/groove/internal/git"
	"github.com/mbriggs/groove/internal/hooks"
	"github.com/mbriggs/groove/internal/normalize"
	"github.com/mbriggs/groove/internal/ports"
	"github.com/mbriggs/groove/internal/resolve"
	"github.com/mbriggs/groove/internal/state"
	"github.com/mbriggs/groove/internal/tmux"
	"github.com/mbriggs/groove/internal/ui"
	"github.com/spf13/cobra"
)

func openCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <branch-name>",
		Short: "Create a new worktree from a new branch",
		Args:  cobra.ExactArgs(1),
		RunE:  runOpen,
	}
}

func runOpen(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	branchInput := args[0]
	globalCfg, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	proj, err := resolve.Project(globalCfg)
	if err != nil {
		return err
	}

	defaultBranch, err := git.DefaultBranch(proj.Root, globalCfg.DefaultRemote)
	if err != nil {
		return err
	}

	fullBranch := branchInput
	if globalCfg.BranchPrefix != "" {
		fullBranch = globalCfg.BranchPrefix + "/" + branchInput
	}
	normalized := normalize.Branch(fullBranch)

	// Check if branch already exists
	if git.LocalBranchExists(proj.Root, fullBranch) {
		store, err := state.Load()
		if err != nil {
			return err
		}

		wtID := proj.Name + "-" + normalized
		if wt, ok := store.Worktrees[wtID]; ok {
			attach, err := ui.Confirm(fmt.Sprintf("Branch %s already exists. Attach to existing worktree?", fullBranch))
			if err != nil || !attach {
				return fmt.Errorf("branch %s already exists", fullBranch)
			}

			shell := config.DetectShell(globalCfg)
			envVars := resolve.BuildEnvVars(wt)
			if err := tmux.WriteEnvFile(wt.Path, envVars); err != nil {
				return err
			}
			if !tmux.SessionExists(wt.SessionName(proj.Config.SessionPrefix)) {
				if err := tmux.CreateSession(wt.SessionName(proj.Config.SessionPrefix), wt.Path, shell, envVars); err != nil {
					return err
				}
			}
			fmt.Fprintf(os.Stderr, "switching to session %s\n", wt.SessionName(proj.Config.SessionPrefix))
			return tmux.SwitchToSession(wt.SessionName(proj.Config.SessionPrefix))
		}

		return fmt.Errorf("branch %s already exists but has no groove worktree — resolve manually", fullBranch)
	}

	// Create branch from default branch
	startPoint := globalCfg.DefaultRemote + "/" + defaultBranch
	if err := git.CreateBranch(proj.Root, fullBranch, startPoint); err != nil {
		return err
	}

	if err := setupWorktree(globalCfg, proj, fullBranch, normalized, defaultBranch); err != nil {
		// Clean up the branch we just created
		fmt.Fprintf(os.Stderr, "cleaning up branch %s after failure\n", fullBranch)
		git.DeleteBranch(proj.Root, fullBranch)
		return err
	}
	return nil
}

func setupWorktree(globalCfg *config.Global, proj *resolve.ProjectInfo, branch, normalized, defaultBranch string) error {
	wtRoot, err := globalCfg.ExpandWorktreeRoot()
	if err != nil {
		return err
	}

	wtPath := filepath.Join(wtRoot, proj.Name, normalized)
	wtID := proj.Name + "-" + normalized

	// Track whether we created the worktree so we can clean up on failure
	createdWorktree := false

	// Check if path already exists
	if _, err := os.Stat(wtPath); err == nil {
		fmt.Fprintf(os.Stderr, "worktree path %s already exists\n", wtPath)
		adopt, err := ui.Confirm("Adopt into groove state?")
		if err != nil || !adopt {
			return fmt.Errorf("worktree path already exists: %s", wtPath)
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
			return err
		}
		if err := git.CreateWorktree(proj.Root, wtPath, branch); err != nil {
			return err
		}
		createdWorktree = true
	}

	cleanup := func() {
		if createdWorktree {
			fmt.Fprintf(os.Stderr, "cleaning up worktree at %s after failure\n", wtPath)
			git.RemoveWorktree(proj.Root, wtPath)
		}
	}

	// Claim ports
	store, lockFile, err := state.LoadLocked()
	if err != nil {
		cleanup()
		return err
	}

	claimed := store.ClaimedPorts()
	assignedPorts, err := ports.ClaimPorts(proj.Config.Ports, claimed, func(name string, port int, claimedBy string) bool {
		ok, _ := ui.Confirm(fmt.Sprintf("port %d (%s) is claimed by %s, continue anyway?", port, name, claimedBy))
		return ok
	})
	if err != nil {
		store.SaveAndUnlock(lockFile)
		cleanup()
		return err
	}

	wt := &state.Worktree{
		ID:               wtID,
		Project:          proj.Name,
		ProjectRoot:      proj.Root,
		ProjectRemoteURL: proj.RemoteURL,
		Path:             wtPath,
		Branch:           branch,
		DefaultBranch:    defaultBranch,
		Ports:            assignedPorts,
		CreatedAt:        time.Now(),
	}

	store.Worktrees[wtID] = wt
	if err := store.SaveAndUnlock(lockFile); err != nil {
		cleanup()
		return err
	}

	sessionName := wt.SessionName(proj.Config.SessionPrefix)

	// Write env file and create session
	envVars := resolve.BuildEnvVars(wt)
	if err := tmux.WriteEnvFile(wtPath, envVars); err != nil {
		return err
	}

	shell := config.DetectShell(globalCfg)

	// Create the tmux session (always detached first)
	if !tmux.SessionExists(sessionName) {
		if err := tmux.CreateSession(sessionName, wtPath, shell, envVars); err != nil {
			return err
		}
	}

	// Run open hook if configured
	if proj.Config.Hooks.Open != "" {
		fmt.Fprintln(os.Stderr, "running open hook...")
		if err := hooks.Run(proj.Config.Hooks.Open, wtPath, envVars); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}

	// Switch to the session
	fmt.Fprintf(os.Stderr, "switching to session %s\n", sessionName)
	return tmux.SwitchToSession(sessionName)
}
