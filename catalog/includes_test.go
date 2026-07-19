package catalog_test

import (
	"context"
	"testing"

	"github.com/pawnkit/pawnmigrate/catalog"
	"github.com/pawnkit/pawnmigrate/migrate"
)

func TestOpenMPIncludeUsesDirectiveNodes(t *testing.T) {
	file := migrate.File{Path: "main.pwn", ID: 1, Content: "#include <a_samp>\n#include <other>\n"}
	edits, err := (catalog.OpenMPInclude{}).Plan(context.Background(), file)
	if err != nil || len(edits) != 1 || edits[0].NewText != "<open.mp>" {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
}

func TestOpenMPIncludeRequiresReview(t *testing.T) {
	if got := (catalog.OpenMPInclude{}).Metadata().Safety; got != migrate.ReviewRequired {
		t.Fatalf("safety = %q, want review-required", got)
	}
}

func TestOpenMPIncludeLeavesCommentsAndMacrosAlone(t *testing.T) {
	for _, source := range []string{
		"// #include <a_samp>\n",
		"#define LEGACY <a_samp>\n#include LEGACY\n",
		"#include <a_samp\n",
	} {
		file := migrate.File{Path: "main.pwn", ID: 1, Content: source}
		edits, err := (catalog.OpenMPInclude{}).Plan(context.Background(), file)
		if err != nil || len(edits) != 0 {
			t.Fatalf("source %q: edits/error = %#v %v", source, edits, err)
		}
	}
}
