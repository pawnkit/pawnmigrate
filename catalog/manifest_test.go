package catalog_test

import (
	"context"
	"strings"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
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

func TestManifestCleanupRemovesDuplicateIncludePaths(t *testing.T) {
	file := migrate.File{Path: "pawn.json", ID: 1, Content: `{
  "pawnkit": {"includePaths": ["include", "vendor", "include", "vendor"]}
}
`}
	edits, err := (catalog.ManifestCleanup{}).Plan(context.Background(), file)
	if err != nil || len(edits) != 1 {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
	if strings.Count(edits[0].NewText, `"include"`) != 1 ||
		strings.Count(edits[0].NewText, `"vendor"`) != 1 {
		t.Fatalf("output = %s", edits[0].NewText)
	}
	file.Content = edits[0].NewText
	edits, err = (catalog.ManifestCleanup{}).Plan(context.Background(), file)
	if err != nil || len(edits) != 0 {
		t.Fatalf("second plan = %#v %v", edits, err)
	}
}

func TestManifestCleanupSupportsYAML(t *testing.T) {
	file := migrate.File{Path: "pawn.yaml", ID: 1, Content: "pawnkit:\n  includePaths:\n    - include\n    - include\n"}
	edits, err := (catalog.ManifestCleanup{}).Plan(context.Background(), file)
	if err != nil || len(edits) != 1 || strings.Count(edits[0].NewText, "- include\n") != 1 {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
}

func TestManifestCleanupLeavesMalformedPathsAlone(t *testing.T) {
	file := migrate.File{Path: "pawn.json", ID: 1, Content: `{"pawnkit":{"includePaths":["include", 1, "include"]}}`}
	edits, err := (catalog.ManifestCleanup{}).Plan(context.Background(), file)
	if err != nil || len(edits) != 0 {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
}

func TestManifestCleanupHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	file := migrate.File{Path: "pawn.json", ID: 1, Content: "{}"}
	if _, err := (catalog.ManifestCleanup{}).Plan(ctx, file); err == nil {
		t.Fatal("canceled plan returned nil error")
	}
}

func TestBuiltinsComposeManifestMigrations(t *testing.T) {
	file := migrate.File{Path: "pawn.json", ID: source.FileID(1), Content: `{
  "pawnkit": {"includePaths": ["include", "include"]}
}
`}
	plan, err := migrate.Build(context.Background(), []migrate.File{file}, catalog.Builtins(), migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	results, err := plan.Preview()
	if err != nil || len(results) != 1 {
		t.Fatalf("results/error = %#v %v", results, err)
	}
	if !strings.Contains(results[0].After, `"schemaVersion": 1`) ||
		strings.Count(results[0].After, `"include"`) != 1 {
		t.Fatalf("after = %s", results[0].After)
	}
}
