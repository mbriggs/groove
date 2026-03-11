package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mbriggs/groove/internal/config"
	"github.com/mbriggs/groove/internal/git"
	"github.com/mbriggs/groove/internal/resolve"
	"github.com/mbriggs/groove/internal/state"
	"github.com/mbriggs/groove/internal/ui"
	"github.com/spf13/cobra"
)

func pruneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prune",
		Short: "Clean up worktrees whose branches have been merged/deleted on remote",
		RunE:  runPrune,
	}
}

func runPrune(cmd *cobra.Command, args []string) error {
	globalCfg, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	proj, err := resolve.Project(globalCfg)
	if err != nil {
		return err
	}

	remote := globalCfg.DefaultRemote

	fmt.Fprintf(os.Stderr, "fetching from %s...\n", remote)
	if err := git.Fetch(proj.Root, remote); err != nil {
		return err
	}

	store, err := state.Load()
	if err != nil {
		return err
	}

	worktrees := store.WorktreesForProject(proj.Name)
	if len(worktrees) == 0 {
		fmt.Fprintln(os.Stderr, "no worktrees to prune")
		return nil
	}

	// Check which branches still exist on remote
	var candidates []*state.Worktree
	for _, wt := range worktrees {
		// Strip the remote prefix if present for the check
		branchName := wt.Branch
		if strings.HasPrefix(branchName, remote+"/") {
			branchName = strings.TrimPrefix(branchName, remote+"/")
		}

		exists, err := git.RemoteBranchExists(proj.Root, remote, branchName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not check branch %s: %v\n", branchName, err)
			continue
		}
		if !exists {
			candidates = append(candidates, wt)
		}
	}

	if len(candidates) == 0 {
		fmt.Fprintln(os.Stderr, "no worktrees to prune — all branches still exist on remote")
		return nil
	}

	// Show candidates for confirmation
	names := make([]string, len(candidates))
	wtMap := make(map[string]*state.Worktree)
	for i, wt := range candidates {
		names[i] = fmt.Sprintf("%s (%s)", wt.Branch, wt.ID)
		wtMap[names[i]] = wt
	}

	selected, err := ui.RunPruneSelector(names)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		fmt.Fprintln(os.Stderr, "nothing selected, aborting")
		return nil
	}

	// Re-load with lock for modifications
	store, lockFile, err := state.LoadLocked()
	if err != nil {
		return err
	}

	for _, name := range selected {
		wt := wtMap[name]
		archiveWorktree(wt, store)
	}

	return store.SaveAndUnlock(lockFile)
}
