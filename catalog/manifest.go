// Package catalog contains built-in migrations.
package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
	"github.com/pawnkit/pawnmigrate/migrate"
	"gopkg.in/yaml.v3"
)

type ManifestSchema struct{}

func (ManifestSchema) Metadata() migrate.Metadata {
	return migrate.Metadata{
		ID: "project.manifest-schema-v1", Version: 1,
		Description: "add the PawnKit manifest schema version",
		Confidence:  migrate.ConfidenceHigh, Safety: migrate.Safe,
	}
}

func (ManifestSchema) Plan(_ context.Context, file migrate.File) ([]textedit.Edit, error) {
	if !isManifest(file.Path) {
		return nil, nil
	}
	var document map[string]any
	if err := decodeManifest(file.Path, file.Content, &document); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	pawnkit, ok := document["pawnkit"].(map[string]any)
	if !ok {
		pawnkit = make(map[string]any)
		document["pawnkit"] = pawnkit
	}
	if _, exists := pawnkit["schemaVersion"]; exists {
		return nil, nil
	}
	pawnkit["schemaVersion"] = 1
	formatted, err := encodeManifest(file.Path, document)
	if err != nil {
		return nil, err
	}
	if string(formatted) == file.Content {
		return nil, nil
	}
	return []textedit.Edit{{Span: source.Span{File: file.ID, Start: 0, End: source.Offset(len(file.Content))}, NewText: string(formatted)}}, nil
}

func isManifest(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return name == "pawn.json" || name == "pawn.yaml" || name == "pawn.yml"
}

func decodeManifest(path, content string, document *map[string]any) error {
	if strings.EqualFold(filepath.Ext(path), ".json") {
		decoder := json.NewDecoder(bytes.NewBufferString(content))
		decoder.UseNumber()
		if err := decoder.Decode(document); err != nil {
			return err
		}
		var trailing any
		if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
			if err == nil {
				return errors.New("multiple JSON values")
			}
			return err
		}
		return nil
	}
	return yaml.Unmarshal([]byte(content), document)
}

func encodeManifest(path string, document map[string]any) ([]byte, error) {
	if !strings.EqualFold(filepath.Ext(path), ".json") {
		return yaml.Marshal(document)
	}
	formatted, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(formatted, '\n'), nil
}

func Builtins() []migrate.Rule { return []migrate.Rule{ManifestSchema{}, OpenMPInclude{}} }
