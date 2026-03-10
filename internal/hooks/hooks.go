package hooks

import (
	"fmt"
	"os"
	"os/exec"
)

// Run executes a hook script in the given directory with the provided env vars.
func Run(script, dir string, env map[string]string) error {
	if script == "" {
		return nil
	}

	cmd := exec.Command("sh", "-c", script)
	cmd.Dir = dir
	cmd.Stdout = os.Stderr // hooks output goes to stderr
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook %q failed: %w", script, err)
	}
	return nil
}
