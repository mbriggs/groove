package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// CheckInstalled verifies that tmux is in PATH.
func CheckInstalled() error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("groove requires tmux")
	}
	return nil
}

// WriteEnvFile writes the .groove-env file to the worktree path.
// This is sourced by the initial shell and can be sourced by the user's
// shell rc for new panes/windows.
func WriteEnvFile(worktreePath string, envVars map[string]string) error {
	envFile := filepath.Join(worktreePath, ".groove-env")

	keys := make([]string, 0, len(envVars))
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lines []string
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("export %s=%q", k, envVars[k]))
	}
	lines = append(lines, fmt.Sprintf("export GROOVE_ENV_FILE=%q", envFile))

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(envFile, []byte(content), 0o644)
}

// CreateSession creates a new detached tmux session with env vars and cwd set.
func CreateSession(sessionName, worktreePath, shell string, envVars map[string]string) error {
	envFile := filepath.Join(worktreePath, ".groove-env")
	shellCmd := fmt.Sprintf("source %q && exec %s", envFile, shell)

	args := []string{
		"new-session",
		"-d",
		"-s", sessionName,
		"-c", worktreePath,
		shell, "-c", shellCmd,
	}

	// Set env vars on the session so new windows/panes inherit them
	cmd := exec.Command("tmux", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating session %s: %s", sessionName, strings.TrimSpace(string(out)))
	}

	// Set environment variables on the session for new panes
	for k, v := range envVars {
		setEnv := exec.Command("tmux", "set-environment", "-t", sessionName, k, v)
		setEnv.CombinedOutput()
	}
	// Also set GROOVE_ENV_FILE
	setEnv := exec.Command("tmux", "set-environment", "-t", sessionName, "GROOVE_ENV_FILE", envFile)
	setEnv.CombinedOutput()

	return nil
}

// SwitchToSession switches to a session, working from inside or outside tmux.
func SwitchToSession(sessionName string) error {
	if insideTmux() {
		cmd := exec.Command("tmux", "switch-client", "-t", sessionName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	// Outside tmux — attach
	cmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DeleteSession kills a tmux session.
func DeleteSession(sessionName string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("deleting session %s: %s", sessionName, strings.TrimSpace(string(out)))
	}
	return nil
}

// SessionExists checks if a tmux session exists.
func SessionExists(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	return cmd.Run() == nil
}

// ListSessions returns all active tmux session names.
func ListSessions() []string {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var sessions []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			sessions = append(sessions, line)
		}
	}
	return sessions
}

func insideTmux() bool {
	return os.Getenv("TMUX") != ""
}
