package migrate_test

import (
	"context"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
	"github.com/pawnkit/pawnmigrate/migrate"
)

func TestInspectReportsPendingBlockedAndNotApplicable(t *testing.T) {
	file := migrate.File{Path: "main.pwn", ID: 1, Content: "abc"}
	statuses, err := migrate.Inspect(context.Background(), []migrate.File{file}, []migrate.Rule{
		rule{metadata: metadata("pending", migrate.Safe), edits: []textedit.Edit{{Span: source.Span{File: 1, Start: 0, End: 1}, NewText: "A"}}},
		rule{metadata: metadata("blocked", migrate.ReviewRequired), edits: []textedit.Edit{{Span: source.Span{File: 1, Start: 1, End: 2}, NewText: "B"}}},
		rule{metadata: metadata("none", migrate.Safe)},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]migrate.Status{"pending": migrate.StatusPending, "blocked": migrate.StatusBlocked, "none": migrate.StatusNotApplicable}
	for _, status := range statuses {
		if status.Status != want[status.Metadata.ID] {
			t.Fatalf("%s status = %s", status.Metadata.ID, status.Status)
		}
	}
}
