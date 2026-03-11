package main

import (
	"fmt"
	"os"

	"github.com/mbriggs/groove/internal/config"
	"github.com/mbriggs/groove/internal/git"
	"github.com/mbriggs/groove/internal/hooks"
	"github.com/mbriggs/groove/internal/resolve"
	"github.com/mbriggs/groove/internal/state"
	"github.com/mbriggs/groove/internal/tmux"
	"github.com/mbriggs/groove/internal/ui"
	"github.com/spf13/cobra"
)

func archiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive",
		Short: "Archive the current worktree",
		RunE:  runArchive,
	}
}

func runArchive(cmd *cobra.Command, args []string) error {
	store, err := state.Load()
	if err != nil {
		return err
	}

	wt, err := resolve.CurrentWorktree(store)
	if err != nil {
		return err
	}

	wtID := wt.ID

	// Load project config for hooks
	projCfg, _ := config.LoadProject(wt.ProjectRoot)

	// Run archive hook
	if projCfg != nil && projCfg.Hooks.Archive != "" {
		fmt.Fprintln(os.Stderr, "running archive hook...")
		envVars := resolve.BuildEnvVars(wt)
		if err := hooks.Run(projCfg.Hooks.Archive, wt.Path, envVars); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}

	// Prompt to delete local branch
	deleteBranch, _ := ui.Confirm(fmt.Sprintf("Delete local branch %s?", wt.Branch))
	if deleteBranch {
		if err := git.CheckoutBranch(wt.ProjectRoot, wt.DefaultBranch); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not checkout %s: %v\n", wt.DefaultBranch, err)
		} else {
			if err := git.DeleteBranch(wt.ProjectRoot, wt.Branch); err != nil {
				fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			}
		}
	}

	// Remove worktree from disk first, before updating state
	if err := git.RemoveWorktree(wt.ProjectRoot, wt.Path); err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	// Kill tmux session
	if tmux.SessionExists(wt.Session) {
		if err := tmux.DeleteSession(wt.Session); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}

	// Now take the lock and remove from state
	store, lockFile, err := state.LoadLocked()
	if err != nil {
		return err
	}
	if _, ok := store.Worktrees[wtID]; !ok {
		store.SaveAndUnlock(lockFile)
		return fmt.Errorf("worktree %s was already removed", wtID)
	}
	delete(store.Worktrees, wtID)
	if err := store.SaveAndUnlock(lockFile); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "archived worktree %s\n", wtID)
	return nil
}

// archiveWorktree performs the archive flow for a single worktree (used by prune).
func archiveWorktree(wt *state.Worktree, store *state.Store) {
	projCfg, _ := config.LoadProject(wt.ProjectRoot)

	if projCfg != nil && projCfg.Hooks.Archive != "" {
		envVars := resolve.BuildEnvVars(wt)
		if err := hooks.Run(projCfg.Hooks.Archive, wt.Path, envVars); err != nil {
			fmt.Fprintf(os.Stderr, "warning: hook failed for %s: %v\n", wt.ID, err)
		}
	}

	// Remove worktree from disk first
	if err := git.RemoveWorktree(wt.ProjectRoot, wt.Path); err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	if tmux.SessionExists(wt.Session) {
		if err := tmux.DeleteSession(wt.Session); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}

	// Then remove from state
	delete(store.Worktrees, wt.ID)

	fmt.Fprintf(os.Stderr, "archived %s\n", wt.ID)
}
