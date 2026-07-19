package catalog

import (
	"context"
	"path/filepath"
	"strings"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
	"github.com/pawnkit/pawnmigrate/migrate"
)

type OpenMPInclude struct{}

func (OpenMPInclude) Metadata() migrate.Metadata {
	return migrate.Metadata{
		ID: "source.openmp-include", Version: 1,
		Description: "replace the legacy a_samp include with open.mp",
		Profiles:    []string{"openmp"}, Confidence: migrate.ConfidenceHigh, Safety: migrate.ReviewRequired,
	}
}

func (OpenMPInclude) Plan(_ context.Context, file migrate.File) ([]textedit.Edit, error) {
	ext := strings.ToLower(filepath.Ext(file.Path))
	if ext != ".pwn" && ext != ".inc" {
		return nil, nil
	}
	parsed := parser.Parse([]byte(file.Content))
	if parsed.HasParseErrors() {
		return nil, nil
	}
	var edits []textedit.Edit
	walk(parsed.Root, func(node *parser.Node) {
		if node.Kind != parser.KindDirectiveInclude && node.Kind != parser.KindDirectiveTryInclude {
			return
		}
		path := node.Field("path")
		if path == nil {
			return
		}
		text := path.Text(parsed.Source)
		replacement, ok := replaceIncludePath(text)
		if !ok {
			return
		}
		edits = append(edits, textedit.Edit{
			Span:    source.Span{File: file.ID, Start: source.Offset(path.Start), End: source.Offset(path.End)},
			NewText: replacement,
		})
	})
	return edits, nil
}

func replaceIncludePath(path string) (string, bool) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "<a_samp>" {
		return strings.Replace(path, "<a_samp>", "<open.mp>", 1), true
	}
	if trimmed == `"a_samp"` || trimmed == `"a_samp.inc"` {
		return strings.Replace(path, trimmed, `"open.mp"`, 1), true
	}
	return "", false
}

func walk(node *parser.Node, visit func(*parser.Node)) {
	if node == nil {
		return
	}
	visit(node)
	for _, child := range node.Children {
		walk(child, visit)
	}
}
