package resolve

import (
	"testing"

	"github.com/mbriggs/groove/internal/state"
)

func TestBuildEnvVars(t *testing.T) {
	wt := &state.Worktree{
		ID:            "myapp-my-feature",
		Project:       "myapp",
		ProjectRoot:   "/home/user/code/myapp",
		Path:          "/home/user/groove/myapp/my-feature",
		Branch:        "mbriggs/my-feature",
		DefaultBranch: "main",
		Ports:         map[string]int{"web": 3001, "db": 5432},
	}

	env := BuildEnvVars(wt)

	expected := map[string]string{
		"GROOVE_PROJECT":        "myapp",
		"GROOVE_WORKTREE_ID":    "myapp-my-feature",
		"GROOVE_WORKTREE_PATH":  "/home/user/groove/myapp/my-feature",
		"GROOVE_PROJECT_ROOT":   "/home/user/code/myapp",
		"GROOVE_BRANCH":         "mbriggs/my-feature",
		"GROOVE_DEFAULT_BRANCH": "main",
		"GROOVE_PORT_WEB":       "3001",
		"GROOVE_PORT_DB":        "5432",
	}

	for key, want := range expected {
		got, ok := env[key]
		if !ok {
			t.Errorf("missing key %s", key)
			continue
		}
		if got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}

	// Verify no extra keys
	if len(env) != len(expected) {
		t.Errorf("got %d env vars, want %d", len(env), len(expected))
		for k, v := range env {
			if _, ok := expected[k]; !ok {
				t.Errorf("unexpected key %s = %q", k, v)
			}
		}
	}
}

func TestBuildEnvVarsNoPorts(t *testing.T) {
	wt := &state.Worktree{
		ID:            "myapp-my-feature",
		Project:       "myapp",
		ProjectRoot:   "/home/user/code/myapp",
		Path:          "/home/user/groove/myapp/my-feature",
		Branch:        "my-feature",
		DefaultBranch: "main",
		Ports:         map[string]int{},
	}

	env := BuildEnvVars(wt)
	if len(env) != 6 {
		t.Errorf("got %d env vars, want 6 (no port vars)", len(env))
	}
}

func TestBuildEnvVarsPortKeysUppercased(t *testing.T) {
	wt := &state.Worktree{
		ID:            "test",
		Project:       "test",
		ProjectRoot:   "/tmp",
		Path:          "/tmp",
		Branch:        "main",
		DefaultBranch: "main",
		Ports:         map[string]int{"my-port": 3000},
	}

	env := BuildEnvVars(wt)
	if _, ok := env["GROOVE_PORT_MY-PORT"]; !ok {
		t.Error("expected GROOVE_PORT_MY-PORT to be set")
		for k := range env {
			t.Logf("  got key: %s", k)
		}
	}
}
