package catalog

import (
	"context"
	"strings"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
	"github.com/pawnkit/pawnmigrate/migrate"
)

// YSIOfficial moves the old YSI fork reference to the maintained repository.
type YSIOfficial struct{}

func (YSIOfficial) Metadata() migrate.Metadata {
	return migrate.Metadata{
		ID: "project.ysi-official", Version: 1,
		Description:   "move the Misiur YSI dependency to the official YSI repository",
		Prerequisites: []string{"the project uses YSI 5.x"},
		Confidence:    migrate.ConfidenceHigh, Safety: migrate.ReviewRequired,
	}
}

func (YSIOfficial) Plan(_ context.Context, file migrate.File) ([]textedit.Edit, error) {
	if !isManifest(file.Path) {
		return nil, nil
	}
	var document map[string]any
	if err := decodeManifest(file.Path, file.Content, &document); err != nil {
		return nil, err
	}
	changed := replaceYSIDependency(document)
	if !changed {
		return nil, nil
	}
	formatted, err := encodeManifest(file.Path, document)
	if err != nil {
		return nil, err
	}
	return []textedit.Edit{{
		Span:    source.Span{File: file.ID, Start: 0, End: source.Offset(len(file.Content))},
		NewText: string(formatted),
	}}, nil
}

func replaceYSIDependency(document map[string]any) bool {
	changed := false
	for _, key := range []string{"dependencies", "dev_dependencies"} {
		values, ok := document[key].([]any)
		if !ok {
			continue
		}
		for i, value := range values {
			text, ok := value.(string)
			if !ok || !isMisiurYSI(text) {
				continue
			}
			values[i] = "pawn-lang/YSI-Includes@5.x"
			changed = true
		}
	}
	pawnkit, ok := document["pawnkit"].(map[string]any)
	if !ok {
		return changed
	}
	overrides, ok := pawnkit["dependencyOverrides"].(map[string]any)
	if !ok {
		return changed
	}
	for key, value := range overrides {
		if !isMisiurYSI(key) {
			continue
		}
		delete(overrides, key)
		overrides["pawn-lang/YSI-Includes"] = "pawn-lang/YSI-Includes@5.x"
		changed = true
		if text, ok := value.(string); ok && strings.HasPrefix(text, "Misiur/YSI-Includes") {
			overrides["pawn-lang/YSI-Includes"] = "pawn-lang/YSI-Includes@5.x"
		}
	}
	return changed
}

func isMisiurYSI(value string) bool {
	return value == "Misiur/YSI-Includes" || strings.HasPrefix(value, "Misiur/YSI-Includes@") ||
		strings.HasPrefix(value, "Misiur/YSI-Includes:") || strings.HasPrefix(value, "Misiur/YSI-Includes#")
}
