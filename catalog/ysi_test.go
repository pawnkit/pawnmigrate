package catalog_test

import (
	"context"
	"strings"
	"testing"

	"github.com/pawnkit/pawnmigrate/catalog"
	"github.com/pawnkit/pawnmigrate/migrate"
)

func TestYSIOfficialReplacesDependencyAndOverride(t *testing.T) {
	file := migrate.File{Path: "pawn.json", ID: 1, Content: `{
  "dependencies": ["Misiur/YSI-Includes", "owner/other"],
  "pawnkit": {"dependencyOverrides": {"Misiur/YSI-Includes": "Misiur/YSI-Includes@5.x"}}
}
`}
	edits, err := (catalog.YSIOfficial{}).Plan(context.Background(), file)
	if err != nil || len(edits) != 1 {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
	if strings.Contains(edits[0].NewText, "Misiur/YSI-Includes") ||
		!strings.Contains(edits[0].NewText, "pawn-lang/YSI-Includes@5.x") {
		t.Fatalf("output = %s", edits[0].NewText)
	}
}

func TestYSIOfficialLeavesOtherDependenciesAlone(t *testing.T) {
	file := migrate.File{Path: "pawn.json", ID: 1, Content: `{"dependencies":["owner/other"]}`}
	edits, err := (catalog.YSIOfficial{}).Plan(context.Background(), file)
	if err != nil || len(edits) != 0 {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
}

func TestYSIOfficialSupportsYAML(t *testing.T) {
	file := migrate.File{Path: "pawn.yaml", ID: 1, Content: "dependencies:\n  - Misiur/YSI-Includes\n"}
	edits, err := (catalog.YSIOfficial{}).Plan(context.Background(), file)
	if err != nil || len(edits) != 1 || !strings.Contains(edits[0].NewText, "pawn-lang/YSI-Includes@5.x") {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
}
