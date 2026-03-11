package main

import (
	"fmt"
	"os"

	"github.com/mbriggs/groove/internal/git"
	"github.com/mbriggs/groove/internal/state"
	"github.com/mbriggs/groove/internal/tmux"
	"github.com/spf13/cobra"
)

func gcCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gc",
		Short: "Reconcile state against reality",
		RunE:  runGC,
	}
}

func runGC(cmd *cobra.Command, args []string) error {
	store, lockFile, err := state.LoadLocked()
	if err != nil {
		return err
	}

	projectRoots := make(map[string]bool)
	var removed int

	for id, wt := range store.Worktrees {
		projectRoots[wt.ProjectRoot] = true

		if _, err := os.Stat(wt.Path); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "removing stale worktree %s (path missing)\n", id)
			delete(store.Worktrees, id)
			removed++
			continue
		}

		if !tmux.SessionExists(wt.Session) {
			fmt.Fprintf(os.Stderr, "note: session %s is dead for worktree %s\n", wt.Session, id)
		}
	}

	if err := store.SaveAndUnlock(lockFile); err != nil {
		return err
	}

	for root := range projectRoots {
		if err := git.PruneWorktrees(root); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}

	if removed > 0 {
		fmt.Fprintf(os.Stderr, "removed %d stale entries\n", removed)
	} else {
		fmt.Fprintln(os.Stderr, "state is clean")
	}

	return nil
}
