// Package safety checks workspace mutation prerequisites.
package safety

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var (
	ErrNotRepository = errors.New("workspace is not a Git repository")
	ErrDirty         = errors.New("git worktree has uncommitted changes")
)

func CheckGit(root string) error {
	inside, err := exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree").CombinedOutput() //nolint:gosec // git and its arguments are fixed; root is data.
	if err != nil || strings.TrimSpace(string(inside)) != "true" {
		return ErrNotRepository
	}
	output, err := exec.Command("git", "-C", root, "status", "--porcelain=v1", "--untracked-files=all").CombinedOutput() //nolint:gosec // git and its arguments are fixed; root is data.
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if len(strings.TrimSpace(string(output))) > 0 {
		return ErrDirty
	}
	return nil
}
