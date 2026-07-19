package report_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
	"github.com/pawnkit/pawnmigrate/migrate"
	"github.com/pawnkit/pawnmigrate/report"
)

func TestDiffUsesCombinedFileResult(t *testing.T) {
	plan := migrate.Plan{SchemaVersion: 1, Changes: []migrate.Change{
		{Path: "main.pwn", Before: "abc\n", Edits: []textedit.Edit{{Span: source.Span{File: 1, Start: 0, End: 1}, NewText: "A"}}},
		{Path: "main.pwn", Before: "abc\n", Edits: []textedit.Edit{{Span: source.Span{File: 1, Start: 1, End: 2}, NewText: "B"}}},
	}}
	var output bytes.Buffer
	if err := report.Diff(&output, plan); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "+ABc") || strings.Count(output.String(), "--- ") != 1 {
		t.Fatalf("diff = %q", output.String())
	}
}

func TestHumanReportsNoChanges(t *testing.T) {
	var output bytes.Buffer
	if err := report.Human(&output, migrate.Plan{}); err != nil || output.String() != "no migrations available\n" {
		t.Fatalf("output/error = %q %v", output.String(), err)
	}
}
