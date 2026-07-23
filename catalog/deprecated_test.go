package catalog_test

import (
	"context"
	"testing"

	"github.com/pawnkit/pawn-api/pawnapi"
	"github.com/pawnkit/pawnmigrate/catalog"
	"github.com/pawnkit/pawnmigrate/migrate"
)

func TestDeprecatedCallsUsesAPIReplacements(t *testing.T) {
	old := apiEntry("OldCall")
	old.Deprecated = &pawnapi.Deprecation{Since: "1.0.0", Replacement: "native:NewCall"}
	index, err := pawnapi.NewIndex([]pawnapi.Entry{old, apiEntry("NewCall")})
	if err != nil {
		t.Fatal(err)
	}
	file := migrate.File{Path: "main.pwn", ID: 1, Content: "main() { OldCall(1); }\n"}
	edits, err := (catalog.DeprecatedCalls{Index: index}).Plan(context.Background(), file)
	if err != nil || len(edits) != 1 || edits[0].NewText != "NewCall" {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
}

func TestDeprecatedCallsRejectsAmbiguousNames(t *testing.T) {
	old := apiEntry("OldCall")
	old.Deprecated = &pawnapi.Deprecation{Since: "1.0.0", Replacement: "native:NewCall"}
	duplicate := apiEntry("OldCall")
	duplicate.ID = "function:OldCall"
	duplicate.Kind = pawnapi.KindFunction
	index, err := pawnapi.NewIndex([]pawnapi.Entry{old, duplicate, apiEntry("NewCall")})
	if err != nil {
		t.Fatal(err)
	}
	file := migrate.File{Path: "main.pwn", ID: 1, Content: "main() { OldCall(); }\n"}
	edits, err := (catalog.DeprecatedCalls{Index: index}).Plan(context.Background(), file)
	if err != nil || len(edits) != 0 {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
}

func TestDeprecatedCallsLeavesLocalFunctionAlone(t *testing.T) {
	old := apiEntry("OldCall")
	old.Deprecated = &pawnapi.Deprecation{Since: "1.0.0", Replacement: "native:NewCall"}
	index, err := pawnapi.NewIndex([]pawnapi.Entry{old, apiEntry("NewCall")})
	if err != nil {
		t.Fatal(err)
	}
	file := migrate.File{Path: "main.pwn", ID: 1, Content: "OldCall() {}\nmain() { OldCall(); }\n"}
	edits, err := (catalog.DeprecatedCalls{Index: index}).Plan(context.Background(), file)
	if err != nil || len(edits) != 0 {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
}

func TestDeprecatedCallsRequiresReviewedMetadata(t *testing.T) {
	old := apiEntry("OldCall")
	old.Deprecated = &pawnapi.Deprecation{Since: "1.0.0", Replacement: "native:NewCall"}
	old.ReviewStatus = pawnapi.ReviewGenerated
	index, err := pawnapi.NewIndex([]pawnapi.Entry{old, apiEntry("NewCall")})
	if err != nil {
		t.Fatal(err)
	}
	file := migrate.File{Path: "main.pwn", ID: 1, Content: "main() { OldCall(); }\n"}
	edits, err := (catalog.DeprecatedCalls{Index: index}).Plan(context.Background(), file)
	if err != nil || len(edits) != 0 {
		t.Fatalf("edits/error = %#v %v", edits, err)
	}
}

func apiEntry(name string) pawnapi.Entry {
	return pawnapi.Entry{
		ID: "native:" + name, Kind: pawnapi.KindNative, Name: name,
		Signature: &pawnapi.Signature{}, Availability: []pawnapi.Availability{{Profile: "openmp"}},
		Source:       pawnapi.Source{Path: "test.inc", Repository: "https://example.test/repo", Commit: "0123456789abcdef0123456789abcdef01234567", License: "MIT"},
		Confidence:   pawnapi.ConfidenceHigh,
		ReviewStatus: pawnapi.ReviewReviewed,
	}
}
