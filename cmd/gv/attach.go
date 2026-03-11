package main

import (
	"fmt"
	"os"

	"github.com/mbriggs/groove/internal/config"
	"github.com/mbriggs/groove/internal/resolve"
	"github.com/mbriggs/groove/internal/state"
	"github.com/mbriggs/groove/internal/tmux"
	"github.com/mbriggs/groove/internal/ui"
	"github.com/spf13/cobra"
)

func attachCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "attach",
		Short: "Fuzzy-find and reattach to an existing worktree",
		RunE:  runAttach,
	}
}

func runAttach(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

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
	if len(worktrees) == 0 {
		return fmt.Errorf("no worktrees found for project %s", proj.Name)
	}

	items := make([]ui.PickerItem, len(worktrees))
	wtMap := make(map[string]*state.Worktree)
	for i, wt := range worktrees {
		items[i] = ui.PickerItem{
			Name: wt.Branch,
			Desc: wt.Path,
		}
		wtMap[wt.Branch] = wt
	}

	choice, err := ui.RunPicker("Select worktree", items)
	if err != nil {
		return err
	}

	wt := wtMap[choice]
	shell := config.DetectShell(globalCfg)

	// Ensure env file and session exist
	envVars := resolve.BuildEnvVars(wt)
	if err := tmux.WriteEnvFile(wt.Path, envVars); err != nil {
		return err
	}

	if !tmux.SessionExists(wt.Session) {
		if err := tmux.CreateSession(wt.Session, wt.Path, shell, envVars); err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stderr, "switching to session %s\n", wt.Session)
	return tmux.SwitchToSession(wt.Session)
}
