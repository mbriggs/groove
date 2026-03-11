package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mbriggs/groove/internal/state"
	"github.com/mbriggs/groove/internal/tmux"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	projectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	branchStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	pathStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	portStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	aliveStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	deadStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

// RenderWorktreeList renders a styled table of worktrees to stdout.
func RenderWorktreeList(worktrees []*state.Worktree) {
	if len(worktrees) == 0 {
		fmt.Fprintln(os.Stderr, "no worktrees found")
		return
	}

	// Compute column widths
	maxProject, maxBranch, maxPath, maxPorts := 7, 6, 4, 5
	for _, wt := range worktrees {
		if len(wt.Project) > maxProject {
			maxProject = len(wt.Project)
		}
		if len(wt.Branch) > maxBranch {
			maxBranch = len(wt.Branch)
		}
		if len(wt.Path) > maxPath {
			maxPath = len(wt.Path)
		}
		ports := formatPorts(wt.Ports)
		if len(ports) > maxPorts {
			maxPorts = len(ports)
		}
	}

	// Header
	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %s",
		maxProject, "PROJECT",
		maxBranch, "BRANCH",
		maxPath, "PATH",
		maxPorts, "PORTS",
		"SESSION",
	)
	fmt.Println(headerStyle.Render(header))

	// Rows
	sessions := tmux.ListSessions()
	sessionSet := make(map[string]bool)
	for _, s := range sessions {
		sessionSet[s] = true
	}

	for _, wt := range worktrees {
		status := deadStyle.Render("dead")
		if sessionSet[wt.SessionName("")] {
			status = aliveStyle.Render("alive")
		}

		ports := formatPorts(wt.Ports)

		row := fmt.Sprintf("%s  %s  %s  %s  %s",
			projectStyle.Render(pad(wt.Project, maxProject)),
			branchStyle.Render(pad(wt.Branch, maxBranch)),
			pathStyle.Render(pad(wt.Path, maxPath)),
			portStyle.Render(pad(ports, maxPorts)),
			status,
		)
		fmt.Println(row)
	}
}

func formatPorts(ports map[string]int) string {
	if len(ports) == 0 {
		return "-"
	}
	var parts []string
	for name, port := range ports {
		parts = append(parts, fmt.Sprintf("%s:%d", name, port))
	}
	return strings.Join(parts, ",")
}

func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
