package catalog

import (
	"context"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
	"github.com/pawnkit/pawnmigrate/migrate"
)

// ManifestCleanup removes exact duplicate PawnKit include roots.
type ManifestCleanup struct{}

func (ManifestCleanup) Metadata() migrate.Metadata {
	return migrate.Metadata{
		ID: "project.manifest-cleanup-v1", Version: 1,
		Description: "remove duplicate PawnKit include paths",
		Confidence:  migrate.ConfidenceHigh, Safety: migrate.Safe,
	}
}

func (ManifestCleanup) Plan(ctx context.Context, file migrate.File) ([]textedit.Edit, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !isManifest(file.Path) {
		return nil, nil
	}
	var document map[string]any
	if err := decodeManifest(file.Path, file.Content, &document); err != nil {
		return nil, err
	}
	pawnkit, ok := document["pawnkit"].(map[string]any)
	if !ok {
		return nil, nil
	}
	paths, ok := pawnkit["includePaths"].([]any)
	if !ok || len(paths) < 2 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(paths))
	unique := make([]any, 0, len(paths))
	for _, value := range paths {
		path, ok := value.(string)
		if !ok {
			return nil, nil
		}
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		unique = append(unique, value)
	}
	if len(unique) == len(paths) {
		return nil, nil
	}
	pawnkit["includePaths"] = unique
	formatted, err := encodeManifest(file.Path, document)
	if err != nil {
		return nil, err
	}
	return []textedit.Edit{{
		Span:    source.Span{File: file.ID, Start: 0, End: source.Offset(len(file.Content))},
		NewText: string(formatted),
	}}, nil
}
