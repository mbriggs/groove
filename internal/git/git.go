package git

import (
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

// DefaultBranch detects the default branch from the remote.
func DefaultBranch(repoDir, remote string) (string, error) {
	cmd := exec.Command("git", "remote", "show", remote)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("detecting default branch: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "HEAD branch:") {
			branch := strings.TrimSpace(strings.TrimPrefix(line, "HEAD branch:"))
			if branch == "" || branch == "(unknown)" {
				return "", fmt.Errorf("could not determine default branch from remote %s", remote)
			}
			return branch, nil
		}
	}

	return "", fmt.Errorf("could not determine default branch from remote %s", remote)
}

// CreateBranch creates a new branch from the default branch.
func CreateBranch(repoDir, branch, startPoint string) error {
	cmd := exec.Command("git", "branch", branch, startPoint)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating branch %s: %s", branch, strings.TrimSpace(string(out)))
	}
	return nil
}

// CreateWorktree adds a new git worktree.
func CreateWorktree(repoDir, path, branch string) error {
	cmd := exec.Command("git", "worktree", "add", path, branch)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating worktree: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// RemoveWorktree removes a git worktree.
func RemoveWorktree(repoDir, path string) error {
	cmd := exec.Command("git", "worktree", "remove", path, "--force")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("removing worktree: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// DeleteBranch deletes a local branch.
func DeleteBranch(repoDir, branch string) error {
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("deleting branch: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// Fetch fetches from the remote.
func Fetch(repoDir, remote string) error {
	cmd := exec.Command("git", "fetch", remote)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("fetching from %s: %s", remote, strings.TrimSpace(string(out)))
	}
	return nil
}

// RemoteBranchExists checks if a branch exists on the remote.
func RemoteBranchExists(repoDir, remote, branch string) (bool, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", remote, branch)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("checking remote branch: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// CreateTrackingBranch creates a local branch tracking a remote branch.
func CreateTrackingBranch(repoDir, localBranch, remote, remoteBranch string) error {
	ref := remote + "/" + remoteBranch
	cmd := exec.Command("git", "branch", "--track", localBranch, ref)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating tracking branch: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// Rebase rebases the current branch onto a ref.
func Rebase(repoDir, onto string) error {
	cmd := exec.Command("git", "rebase", onto)
	cmd.Dir = repoDir
	cmd.Stdout = nil
	cmd.Stderr = nil
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("rebasing: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// PruneWorktrees runs git worktree prune.
func PruneWorktrees(repoDir string) error {
	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pruning worktrees: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// RemoteURL returns the URL of the given remote.
func RemoteURL(repoDir, remote string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remote)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("remote %q not found — add it with: git remote add %s <url>", remote, remote)
	}
	return strings.TrimSpace(string(out)), nil
}

// ProjectNameFromRemote extracts a project name from a git remote URL.
func ProjectNameFromRemote(remoteURL string) string {
	// Handle SSH URLs: git@github.com:user/myapp.git
	if strings.Contains(remoteURL, ":") && strings.HasPrefix(remoteURL, "git@") {
		parts := strings.SplitN(remoteURL, ":", 2)
		if len(parts) == 2 {
			name := filepath.Base(parts[1])
			return strings.TrimSuffix(name, ".git")
		}
	}

	// Handle HTTPS URLs
	if u, err := url.Parse(remoteURL); err == nil {
		name := filepath.Base(u.Path)
		return strings.TrimSuffix(name, ".git")
	}

	// Fallback
	name := filepath.Base(remoteURL)
	return strings.TrimSuffix(name, ".git")
}

// IsBareRepo checks if the repo at dir is bare.
func IsBareRepo(dir string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-bare-repository")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "true", nil
}

// CheckoutBranch checks out a branch.
func CheckoutBranch(repoDir, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("checking out %s: %s", branch, strings.TrimSpace(string(out)))
	}
	return nil
}
