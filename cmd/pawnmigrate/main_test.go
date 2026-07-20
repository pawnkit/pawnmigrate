package main

import "testing"

func TestRunRejectsInvalidOptionsBeforeLoadingProject(t *testing.T) {
	t.Parallel()

	tests := [][]string{
		{"--output", "xml"},
		{"--status", "--output", "diff"},
		{"--apply", "--status"},
		{"unexpected"},
	}
	for _, args := range tests {
		if code := run(args); code != 2 {
			t.Errorf("run(%q) = %d, want 2", args, code)
		}
	}
}

func TestRunVersion(t *testing.T) {
	t.Parallel()

	if code := run([]string{"--version"}); code != 0 {
		t.Fatalf("run(--version) = %d, want 0", code)
	}
}
