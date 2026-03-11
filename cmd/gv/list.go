package main

import (
	"github.com/mbriggs/groove/internal/state"
	"github.com/mbriggs/groove/internal/ui"
	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all worktrees across all projects",
		Aliases: []string{"ls"},
		RunE:  runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	store, err := state.Load()
	if err != nil {
		return err
	}

	var all []*state.Worktree
	for _, wt := range store.Worktrees {
		all = append(all, wt)
	}

	ui.RenderWorktreeList(all)
	return nil
}
