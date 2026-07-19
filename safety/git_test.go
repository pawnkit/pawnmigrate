package safety_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pawnkit/pawnmigrate/safety"
)

func TestCheckGit(t *testing.T) {
	dir := t.TempDir()
	if err := safety.CheckGit(dir); !errors.Is(err, safety.ErrNotRepository) {
		t.Fatalf("non-repository error = %v", err)
	}
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.email", "test@example.test")
	runGit(t, dir, "config", "user.name", "Test")
	path := filepath.Join(dir, "pawn.json")
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := safety.CheckGit(dir); !errors.Is(err, safety.ErrDirty) {
		t.Fatalf("dirty error = %v", err)
	}
	runGit(t, dir, "add", "pawn.json")
	runGit(t, dir, "commit", "-qm", "fixture")
	if err := safety.CheckGit(dir); err != nil {
		t.Fatalf("clean error = %v", err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
