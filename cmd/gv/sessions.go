package main

import (
	"fmt"
	"os"

	"github.com/mbriggs/groove/internal/config"
	"github.com/mbriggs/groove/internal/resolve"
	"github.com/mbriggs/groove/internal/state"
	"github.com/mbriggs/groove/internal/tmux"
	"github.com/spf13/cobra"
)

func sessionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage tmux sessions for worktrees",
	}
	cmd.AddCommand(sessionsUpCmd(), sessionsDownCmd())
	return cmd
}

func sessionsUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Create missing tmux sessions",
		RunE:  runSessionsUp,
	}
	cmd.Flags().Bool("all", false, "all projects, not just current")
	return cmd
}

func sessionsDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Tear down tmux sessions (keeps worktrees and state)",
		RunE:  runSessionsDown,
	}
	cmd.Flags().Bool("all", false, "all projects, not just current")
	return cmd
}

func runSessionsUp(cmd *cobra.Command, args []string) error {
	if err := tmux.CheckInstalled(); err != nil {
		return err
	}

	allProjects, _ := cmd.Flags().GetBool("all")

	globalCfg, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	store, err := state.Load()
	if err != nil {
		return err
	}

	var worktrees []*state.Worktree
	if allProjects {
		for _, wt := range store.Worktrees {
			worktrees = append(worktrees, wt)
		}
	} else {
		proj, err := resolve.Project(globalCfg)
		if err != nil {
			return err
		}
		worktrees = store.WorktreesForProject(proj.Name)
	}

	shell := config.DetectShell(globalCfg)
	var created int

	for _, wt := range worktrees {
		if tmux.SessionExists(wt.Session) {
			continue
		}

		envVars := resolve.BuildEnvVars(wt)
		if err := tmux.WriteEnvFile(wt.Path, envVars); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not write env file for %s: %v\n", wt.ID, err)
			continue
		}

		if err := tmux.CreateSession(wt.Session, wt.Path, shell, envVars); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not create session for %s: %v\n", wt.ID, err)
			continue
		}

		fmt.Fprintf(os.Stderr, "created session %s\n", wt.Session)
		created++
	}

	if created == 0 {
		fmt.Fprintln(os.Stderr, "all sessions already exist")
	} else {
		fmt.Fprintf(os.Stderr, "created %d sessions\n", created)
	}

	return nil
}

func runSessionsDown(cmd *cobra.Command, args []string) error {
	allProjects, _ := cmd.Flags().GetBool("all")

	globalCfg, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	store, err := state.Load()
	if err != nil {
		return err
	}

	var worktrees []*state.Worktree
	if allProjects {
		for _, wt := range store.Worktrees {
			worktrees = append(worktrees, wt)
		}
	} else {
		proj, err := resolve.Project(globalCfg)
		if err != nil {
			return err
		}
		worktrees = store.WorktreesForProject(proj.Name)
	}

	var killed int
	for _, wt := range worktrees {
		if !tmux.SessionExists(wt.Session) {
			continue
		}
		if err := tmux.DeleteSession(wt.Session); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			continue
		}
		fmt.Fprintf(os.Stderr, "killed session %s\n", wt.Session)
		killed++
	}

	if killed == 0 {
		fmt.Fprintln(os.Stderr, "no sessions to kill")
	} else {
		fmt.Fprintf(os.Stderr, "killed %d sessions\n", killed)
	}

	return nil
}
