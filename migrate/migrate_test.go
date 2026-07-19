package migrate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
	"github.com/pawnkit/pawnmigrate/migrate"
)

type rule struct {
	metadata migrate.Metadata
	edits    []textedit.Edit
}

func (r rule) Metadata() migrate.Metadata                                  { return r.metadata }
func (r rule) Plan(context.Context, migrate.File) ([]textedit.Edit, error) { return r.edits, nil }

func TestBuildFiltersUnsafeRules(t *testing.T) {
	fileID := source.FileID(1)
	file := migrate.File{Path: "main.pwn", ID: fileID, Content: "abc"}
	rules := []migrate.Rule{
		rule{metadata: metadata("safe", migrate.Safe), edits: []textedit.Edit{{Span: source.Span{File: fileID, Start: 0, End: 1}, NewText: "A"}}},
		rule{metadata: metadata("review", migrate.ReviewRequired), edits: []textedit.Edit{{Span: source.Span{File: fileID, Start: 1, End: 2}, NewText: "B"}}},
	}
	plan, err := migrate.Build(context.Background(), []migrate.File{file}, rules, migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Changes) != 1 || plan.Changes[0].After != "Abc" {
		t.Fatalf("plan = %#v", plan)
	}
}

func TestBuildRejectsConflictsAcrossRules(t *testing.T) {
	fileID := source.FileID(1)
	file := migrate.File{Path: "main.pwn", ID: fileID, Content: "abc"}
	rules := []migrate.Rule{
		rule{metadata: metadata("one", migrate.Safe), edits: []textedit.Edit{{Span: source.Span{File: fileID, Start: 0, End: 2}, NewText: "x"}}},
		rule{metadata: metadata("two", migrate.Safe), edits: []textedit.Edit{{Span: source.Span{File: fileID, Start: 1, End: 3}, NewText: "y"}}},
	}
	if _, err := migrate.Build(context.Background(), []migrate.File{file}, rules, migrate.Options{}); !errors.Is(err, textedit.ErrOverlappingEdits) {
		t.Fatalf("error = %v", err)
	}
}

func TestBuildRejectsUnknownSelection(t *testing.T) {
	_, err := migrate.Build(context.Background(), nil, []migrate.Rule{
		rule{metadata: metadata("known", migrate.Safe)},
	}, migrate.Options{Selected: map[string]bool{"missing": true}})
	if !errors.Is(err, migrate.ErrUnknownMigration) {
		t.Fatalf("error = %v, want ErrUnknownMigration", err)
	}
}

func TestPreviewRejectsInconsistentEmptySnapshot(t *testing.T) {
	plan := migrate.Plan{Changes: []migrate.Change{
		{Path: "main.pwn", Before: ""},
		{Path: "main.pwn", Before: "changed"},
	}}
	if _, err := plan.Preview(); err == nil {
		t.Fatal("inconsistent snapshots accepted")
	}
}

func metadata(id string, safety migrate.Safety) migrate.Metadata {
	return migrate.Metadata{ID: id, Version: 1, Description: id, Confidence: migrate.ConfidenceHigh, Safety: safety}
}
