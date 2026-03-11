package main

import (
	"github.com/mbriggs/groove/internal/config"
	"github.com/mbriggs/groove/internal/resolve"
	"github.com/mbriggs/groove/internal/state"
	"github.com/mbriggs/groove/internal/ui"
	"github.com/spf13/cobra"
)

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "List worktrees for the current project",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	globalCfg, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	proj, err := resolve.Project(globalCfg)
	if err != nil {
		return err
	}

	store, err := state.Load()
	if err != nil {
		return err
	}

	worktrees := store.WorktreesForProject(proj.Name)
	ui.RenderWorktreeList(worktrees)
	return nil
}
