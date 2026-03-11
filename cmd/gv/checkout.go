package main

import (
	"fmt"
	"os"

	"github.com/mbriggs/groove/internal/config"
	"github.com/mbriggs/groove/internal/git"
	"github.com/mbriggs/groove/internal/normalize"
	"github.com/mbriggs/groove/internal/resolve"
	"github.com/mbriggs/groove/internal/tmux"
	"github.com/spf13/cobra"
)

func checkoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "checkout <remote-branch>",
		Short: "Create a worktree from an existing remote branch",
		Args:  cobra.ExactArgs(1),
		RunE:  runCheckout,
	}
}

func runCheckout(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	remoteBranch := args[0]
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

	exists, err := git.RemoteBranchExists(proj.Root, remote, remoteBranch)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("branch %s not found on %s", remoteBranch, remote)
	}

	defaultBranch, err := git.DefaultBranch(proj.Root, remote)
	if err != nil {
		return err
	}

	normalized := normalize.Branch(remoteBranch)

	// Create local tracking branch
	if err := git.CreateTrackingBranch(proj.Root, remoteBranch, remote, remoteBranch); err != nil {
		return fmt.Errorf("creating tracking branch: %w", err)
	}

	if err := setupWorktree(globalCfg, proj, remoteBranch, normalized, defaultBranch); err != nil {
		// Clean up the tracking branch we just created
		fmt.Fprintf(os.Stderr, "cleaning up branch %s after failure\n", remoteBranch)
		git.DeleteBranch(proj.Root, remoteBranch)
		return err
	}
	return nil
}
