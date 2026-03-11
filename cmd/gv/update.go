package main

import (
	"fmt"
	"os"

	"github.com/mbriggs/groove/internal/config"
	"github.com/mbriggs/groove/internal/git"
	"github.com/mbriggs/groove/internal/resolve"
	"github.com/mbriggs/groove/internal/state"
	"github.com/spf13/cobra"
)

func updateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Rebase current worktree onto the default branch",
		RunE:  runUpdate,
	}
}

func runUpdate(cmd *cobra.Command, args []string) error {
	globalCfg, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	store, err := state.Load()
	if err != nil {
		return err
	}

	wt, err := resolve.CurrentWorktree(store)
	if err != nil {
		return err
	}

	remote := globalCfg.DefaultRemote

	fmt.Fprintf(os.Stderr, "fetching %s...\n", remote)
	if err := git.Fetch(wt.Path, remote); err != nil {
		return err
	}

	onto := remote + "/" + wt.DefaultBranch
	fmt.Fprintf(os.Stderr, "rebasing onto %s...\n", onto)
	if err := git.Rebase(wt.Path, onto); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "done")
	return nil
}
