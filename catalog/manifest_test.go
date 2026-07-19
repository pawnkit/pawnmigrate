package catalog_test

import (
	"context"
	"strings"
	"testing"

	"github.com/pawnkit/pawnmigrate/catalog"
	"github.com/pawnkit/pawnmigrate/migrate"
)

func TestManifestSchemaIsIdempotent(t *testing.T) {
	rule := catalog.ManifestSchema{}
	file := migrate.File{Path: "/project/pawn.json", ID: 1, Content: "{\"entry\":\"main.pwn\"}\n"}
	edits, err := rule.Plan(context.Background(), file)
	if err != nil || len(edits) != 1 {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
	if !strings.Contains(edits[0].NewText, `"schemaVersion": 1`) {
		t.Fatalf("output = %s", edits[0].NewText)
	}
	file.Content = edits[0].NewText
	edits, err = rule.Plan(context.Background(), file)
	if err != nil || len(edits) != 0 {
		t.Fatalf("second plan = %#v %v", edits, err)
	}
}

func TestManifestSchemaSupportsYAML(t *testing.T) {
	rule := catalog.ManifestSchema{}
	file := migrate.File{Path: "/project/pawn.yaml", ID: 1, Content: "entry: main.pwn\n"}
	edits, err := rule.Plan(context.Background(), file)
	if err != nil || len(edits) != 1 {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
	if !strings.Contains(edits[0].NewText, "schemaVersion: 1") {
		t.Fatalf("output = %s", edits[0].NewText)
	}
	file.Content = edits[0].NewText
	edits, err = rule.Plan(context.Background(), file)
	if err != nil || len(edits) != 0 {
		t.Fatalf("second plan = %#v %v", edits, err)
	}
}

func TestManifestSchemaRejectsTrailingJSON(t *testing.T) {
	file := migrate.File{Path: "/project/pawn.json", ID: 1, Content: "{} {}"}
	if _, err := (catalog.ManifestSchema{}).Plan(context.Background(), file); err == nil {
		t.Fatal("trailing JSON accepted")
	}
}
