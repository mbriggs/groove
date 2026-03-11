package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gv",
		Short: "groove — git worktree manager with tmux sessions and port management",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(
		openCmd(),
		checkoutCmd(),
		attachCmd(),
		archiveCmd(),
		pruneCmd(),
		gcCmd(),
		sessionsCmd(),
		updateCmd(),
		listCmd(),
		statusCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
