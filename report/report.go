// Package report renders migration plans.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pawnkit/pawnmigrate/migrate"
)

func JSON(w io.Writer, plan migrate.Plan) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(plan)
}

func StatusJSON(w io.Writer, statuses []migrate.RuleStatus) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(struct {
		SchemaVersion int                  `json:"schemaVersion"`
		Migrations    []migrate.RuleStatus `json:"migrations"`
	}{SchemaVersion: 1, Migrations: statuses})
}

func StatusHuman(w io.Writer, statuses []migrate.RuleStatus) error {
	for _, status := range statuses {
		if _, err := fmt.Fprintf(w, "%s v%d: %s (%s, %s)\n", status.Metadata.ID, status.Metadata.Version, status.Status, status.Metadata.Safety, status.Metadata.Confidence); err != nil {
			return err
		}
	}
	return nil
}

func Human(w io.Writer, plan migrate.Plan) error {
	results, err := plan.Preview()
	if err != nil {
		return err
	}
	if len(results) == 0 {
		_, err = fmt.Fprintln(w, "no migrations available")
		return err
	}
	for _, result := range results {
		count := 0
		for _, change := range plan.Changes {
			if change.Path == result.Path {
				count++
			}
		}
		if _, err := fmt.Fprintf(w, "%s: %d migration(s)\n", result.Path, count); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(w, "%d file(s) would change\n", len(results))
	return err
}

func Diff(w io.Writer, plan migrate.Plan) error {
	results, err := plan.Preview()
	if err != nil {
		return err
	}
	for _, result := range results {
		if _, err := fmt.Fprintf(w, "--- a/%s\n+++ b/%s\n", result.Path, result.Path); err != nil {
			return err
		}
		before := splitLines(result.Before)
		after := splitLines(result.After)
		if _, err := fmt.Fprintf(w, "@@ -1,%d +1,%d @@\n", len(before), len(after)); err != nil {
			return err
		}
		for _, line := range before {
			if _, err := fmt.Fprintln(w, "-"+line); err != nil {
				return err
			}
		}
		for _, line := range after {
			if _, err := fmt.Fprintln(w, "+"+line); err != nil {
				return err
			}
		}
	}
	return nil
}

func splitLines(content string) []string {
	content = strings.TrimSuffix(content, "\n")
	if content == "" {
		return nil
	}
	return strings.Split(content, "\n")
}
