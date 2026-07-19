package migrate

import (
	"context"
	"fmt"
	"sort"
)

type Status string

const (
	StatusPending       Status = "pending"
	StatusNotApplicable Status = "not-applicable"
	StatusBlocked       Status = "blocked"
)

type RuleStatus struct {
	Metadata Metadata `json:"migration"`
	Status   Status   `json:"status"`
	Files    []string `json:"files,omitempty"`
}

func Inspect(ctx context.Context, files []File, rules []Rule, allowUnsafe bool) ([]RuleStatus, error) {
	statuses := make([]RuleStatus, 0, len(rules))
	seen := make(map[string]bool, len(rules))
	for _, rule := range rules {
		metadata := rule.Metadata()
		if err := metadata.Validate(); err != nil {
			return nil, err
		}
		if seen[metadata.ID] {
			return nil, fmt.Errorf("duplicate migration ID %q", metadata.ID)
		}
		seen[metadata.ID] = true
		status := RuleStatus{Metadata: metadata, Status: StatusNotApplicable}
		for _, file := range files {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			edits, err := rule.Plan(ctx, file)
			if err != nil {
				return nil, fmt.Errorf("%s: %s: %w", metadata.ID, file.Path, err)
			}
			if len(edits) > 0 {
				status.Files = append(status.Files, file.Path)
			}
		}
		if len(status.Files) > 0 {
			if metadata.Safety == ReviewRequired && !allowUnsafe {
				status.Status = StatusBlocked
			} else {
				status.Status = StatusPending
			}
		}
		sort.Strings(status.Files)
		statuses = append(statuses, status)
	}
	sort.Slice(statuses, func(i, j int) bool { return statuses[i].Metadata.ID < statuses[j].Metadata.ID })
	return statuses, nil
}
