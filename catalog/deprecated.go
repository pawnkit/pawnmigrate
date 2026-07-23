package catalog

import (
	"context"
	"path/filepath"
	"strings"

	analysis "github.com/pawnkit/pawn-analysis"
	"github.com/pawnkit/pawn-api/pawnapi"
	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
	"github.com/pawnkit/pawnmigrate/migrate"
)

type DeprecatedCalls struct{ Index *pawnapi.Index }

func (DeprecatedCalls) Metadata() migrate.Metadata {
	return migrate.Metadata{
		ID: "api.deprecated-calls", Version: 1,
		Description:   "replace deprecated API calls with documented replacements",
		Prerequisites: []string{"resolved call targets"}, Confidence: migrate.ConfidenceMedium, Safety: migrate.ReviewRequired,
	}
}

func (r DeprecatedCalls) Plan(ctx context.Context, file migrate.File) ([]textedit.Edit, error) {
	if r.Index == nil || !isPawnSource(file.Path) {
		return nil, nil
	}
	parsed := parser.Parse([]byte(file.Content))
	if parsed.HasParseErrors() {
		return nil, nil
	}
	semantic, err := analysis.AnalyzeContext(ctx, []byte(file.Content), analysis.Options{URI: source.FileURI(file.Path)})
	if err != nil {
		return nil, err
	}
	var edits []textedit.Edit
	walk(parsed.Root, func(node *parser.Node) {
		if node.Kind != parser.KindCallExpression {
			return
		}
		callee := node.Field("function")
		if callee == nil || callee.Kind != parser.KindIdentifier {
			return
		}
		name := callee.Text(parsed.Source)
		if !unresolvedCall(semantic, name, callee.Start, callee.End) {
			return
		}
		entries := r.Index.ByName(name)
		if len(entries) != 1 || entries[0].Name != name || entries[0].Deprecated == nil {
			return
		}
		deprecated := entries[0]
		replacement, ok := r.Index.ByID(deprecated.Deprecated.Replacement)
		if !ok || replacement.Name == "" || replacement.Name == name ||
			!reviewedAPIEntry(deprecated) || !reviewedAPIEntry(replacement) {
			return
		}
		edits = append(edits, textedit.Edit{
			Span:    source.Span{File: file.ID, Start: source.Offset(callee.Start), End: source.Offset(callee.End)},
			NewText: replacement.Name,
		})
	})
	return edits, nil
}

func reviewedAPIEntry(entry pawnapi.Entry) bool {
	return entry.Confidence == pawnapi.ConfidenceHigh && entry.ReviewStatus == pawnapi.ReviewReviewed
}

func unresolvedCall(result *analysis.Result, name string, start, end int) bool {
	if result == nil || result.Symbols == nil {
		return false
	}
	for _, reference := range result.Symbols.References {
		if reference.IsCall && reference.Name == name && int(reference.Span.Start) == start && int(reference.Span.End) == end {
			return reference.Resolved == 0
		}
	}
	return false
}

func isPawnSource(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".pwn" || ext == ".inc"
}
