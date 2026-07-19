// Package migrate plans and applies versioned Pawn migrations.
package migrate

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

var ErrUnknownMigration = errors.New("unknown migration")

type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

type Safety string

const (
	Safe           Safety = "safe"
	ReviewRequired Safety = "review-required"
)

type Metadata struct {
	ID            string     `json:"id"`
	Version       int        `json:"version"`
	Description   string     `json:"description"`
	Prerequisites []string   `json:"prerequisites,omitempty"`
	Profiles      []string   `json:"profiles,omitempty"`
	Confidence    Confidence `json:"confidence"`
	Safety        Safety     `json:"safety"`
}

func (m Metadata) Validate() error {
	if m.ID == "" || m.Version < 1 || m.Description == "" {
		return errors.New("migration metadata requires an ID, version, and description")
	}
	if m.Confidence != ConfidenceHigh && m.Confidence != ConfidenceMedium && m.Confidence != ConfidenceLow {
		return fmt.Errorf("migration %s has invalid confidence %q", m.ID, m.Confidence)
	}
	if m.Safety != Safe && m.Safety != ReviewRequired {
		return fmt.Errorf("migration %s has invalid safety %q", m.ID, m.Safety)
	}
	return nil
}

type File struct {
	Path    string
	ID      source.FileID
	Content string
}

type Change struct {
	Migration Metadata        `json:"migration"`
	Path      string          `json:"path"`
	Edits     []textedit.Edit `json:"-"`
	Before    string          `json:"before"`
	After     string          `json:"after"`
}

type Rule interface {
	Metadata() Metadata
	Plan(context.Context, File) ([]textedit.Edit, error)
}

type Options struct {
	Selected    map[string]bool
	AllowUnsafe bool
}

type Plan struct {
	SchemaVersion int      `json:"schemaVersion"`
	Changes       []Change `json:"changes"`
}

type FileResult struct {
	Path   string `json:"path"`
	Before string `json:"before"`
	After  string `json:"after"`
}

func (p Plan) Preview() ([]FileResult, error) {
	type collected struct {
		before string
		edits  []textedit.Edit
		set    bool
	}
	byPath := make(map[string]collected)
	for _, change := range p.Changes {
		item := byPath[change.Path]
		if !item.set {
			item.before = change.Before
			item.set = true
		} else if item.before != change.Before {
			return nil, fmt.Errorf("inconsistent source snapshots for %s", change.Path)
		}
		item.edits = append(item.edits, change.Edits...)
		byPath[change.Path] = item
	}
	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	results := make([]FileResult, 0, len(paths))
	for _, path := range paths {
		item := byPath[path]
		after, err := textedit.Apply(item.before, item.edits)
		if err != nil {
			return nil, err
		}
		results = append(results, FileResult{Path: path, Before: item.before, After: after})
	}
	return results, nil
}

func Build(ctx context.Context, files []File, rules []Rule, opts Options) (Plan, error) {
	seen := make(map[string]bool, len(rules))
	for _, rule := range rules {
		metadata := rule.Metadata()
		if err := metadata.Validate(); err != nil {
			return Plan{}, err
		}
		if seen[metadata.ID] {
			return Plan{}, fmt.Errorf("duplicate migration ID %q", metadata.ID)
		}
		seen[metadata.ID] = true
	}
	unknown := make([]string, 0)
	for id := range opts.Selected {
		if !seen[id] {
			unknown = append(unknown, id)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return Plan{}, fmt.Errorf("%w: %s", ErrUnknownMigration, strings.Join(unknown, ", "))
	}

	plan := Plan{SchemaVersion: 1}
	for _, rule := range rules {
		metadata := rule.Metadata()
		if len(opts.Selected) > 0 && !opts.Selected[metadata.ID] {
			continue
		}
		if metadata.Safety == ReviewRequired && !opts.AllowUnsafe {
			continue
		}
		for _, file := range files {
			if err := ctx.Err(); err != nil {
				return Plan{}, err
			}
			edits, err := rule.Plan(ctx, file)
			if err != nil {
				return Plan{}, fmt.Errorf("%s: %s: %w", metadata.ID, file.Path, err)
			}
			if len(edits) == 0 {
				continue
			}
			after, err := textedit.Apply(file.Content, edits)
			if err != nil {
				return Plan{}, fmt.Errorf("%s: %s: %w", metadata.ID, file.Path, err)
			}
			plan.Changes = append(plan.Changes, Change{Migration: metadata, Path: file.Path, Edits: edits, Before: file.Content, After: after})
		}
	}
	if err := validateConflicts(plan.Changes); err != nil {
		return Plan{}, err
	}
	sort.SliceStable(plan.Changes, func(i, j int) bool {
		if plan.Changes[i].Path != plan.Changes[j].Path {
			return plan.Changes[i].Path < plan.Changes[j].Path
		}
		return plan.Changes[i].Migration.ID < plan.Changes[j].Migration.ID
	})
	return plan, nil
}

func validateConflicts(changes []Change) error {
	byPath := make(map[string][]textedit.Edit)
	for _, change := range changes {
		byPath[change.Path] = append(byPath[change.Path], change.Edits...)
	}
	for path, edits := range byPath {
		if err := textedit.ValidateEdits(edits); err != nil {
			return fmt.Errorf("conflicting edits for %s: %w", path, err)
		}
	}
	return nil
}
