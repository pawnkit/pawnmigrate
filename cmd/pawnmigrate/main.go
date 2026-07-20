package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pawnkit/pawn-api/pawnapi"
	"github.com/pawnkit/pawn-project/fsx"
	"github.com/pawnkit/pawn-project/workspace"
	"github.com/pawnkit/pawnfmt"
	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnmigrate/catalog"
	"github.com/pawnkit/pawnmigrate/migrate"
	"github.com/pawnkit/pawnmigrate/report"
	"github.com/pawnkit/pawnmigrate/safety"
)

const outputJSON = "json"

var version = "dev"

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	flags := flag.NewFlagSet("pawnmigrate", flag.ContinueOnError)
	apply := flags.Bool("apply", false, "apply the migration plan")
	allowUnsafe := flags.Bool("allow-unsafe", false, "include review-required migrations")
	selected := flags.String("only", "", "comma-separated migration IDs")
	project := flags.String("project", ".", "project directory or file")
	output := flags.String("output", "human", "output format: human, json, or diff")
	status := flags.Bool("status", false, "report migration applicability")
	allowDirty := flags.Bool("allow-dirty", false, "apply outside a clean Git worktree")
	showVersion := flags.Bool("version", false, "print the version")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if code, handled := earlyExit(flags.NArg(), *showVersion); handled {
		return code
	}
	if *apply && *status {
		fmt.Fprintln(os.Stderr, "pawnmigrate: --apply and --status cannot be combined")
		return 2
	}
	if !validOutput(*output, *status) {
		fmt.Fprintf(os.Stderr, "pawnmigrate: invalid output format %q\n", *output)
		return 2
	}
	root, err := workspace.FindRoot(fsx.OS{}, absolute(*project))
	if err != nil {
		fmt.Fprintln(os.Stderr, "pawnmigrate:", err)
		return 2
	}
	content, err := os.ReadFile(root.ManifestPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pawnmigrate:", err)
		return 2
	}
	registry := source.NewRegistry()
	files := []migrate.File{{Path: root.ManifestPath, ID: registry.Intern(source.FileURI(root.ManifestPath)), Content: string(content)}}
	sources, err := sourceFiles(root.Dir, registry)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pawnmigrate:", err)
		return 2
	}
	files = append(files, sources...)
	rules := catalog.Builtins()
	apiIndex, err := pawnapi.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pawnmigrate:", err)
		return 3
	}
	rules = append(rules, catalog.DeprecatedCalls{Index: apiIndex})
	if *status {
		statuses, err := migrate.Inspect(context.Background(), files, rules, *allowUnsafe)
		if err != nil {
			fmt.Fprintln(os.Stderr, "pawnmigrate:", err)
			return 2
		}
		if *output == outputJSON {
			err = report.StatusJSON(os.Stdout, statuses)
		} else {
			err = report.StatusHuman(os.Stdout, statuses)
		}
		if err != nil {
			return 3
		}
		return 0
	}
	plan, err := migrate.Build(context.Background(), files, rules, migrate.Options{Selected: selection(*selected), AllowUnsafe: *allowUnsafe})
	if err != nil {
		fmt.Fprintln(os.Stderr, "pawnmigrate:", err)
		return 2
	}
	if *apply {
		if !*allowDirty {
			if err := safety.CheckGit(root.Dir); err != nil {
				fmt.Fprintln(os.Stderr, "pawnmigrate:", err, "(commit changes or pass --allow-dirty)")
				return 2
			}
		}
		if err := migrate.ApplyWithOptions(migrate.OSWriter{}, plan, migrate.ApplyOptions{Format: formatPawn}); err != nil {
			fmt.Fprintln(os.Stderr, "pawnmigrate:", err)
			return 2
		}
	}
	if err := writeReport(*output, plan); err != nil {
		fmt.Fprintln(os.Stderr, "pawnmigrate:", err)
		return 3
	}
	return 0
}

func earlyExit(positional int, showVersion bool) (int, bool) {
	if positional != 0 {
		_, _ = fmt.Fprintln(os.Stderr, "pawnmigrate: unexpected positional arguments")
		return 2, true
	}
	if showVersion {
		if _, err := fmt.Fprintln(os.Stdout, version); err != nil {
			return 3, true
		}
		return 0, true
	}
	return 0, false
}

func writeReport(format string, plan migrate.Plan) error {
	switch format {
	case "human":
		return report.Human(os.Stdout, plan)
	case outputJSON:
		return report.JSON(os.Stdout, plan)
	case "diff":
		return report.Diff(os.Stdout, plan)
	default:
		return fmt.Errorf("unknown output format %q", format)
	}
}

func formatPawn(path string, content []byte) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".pwn" && ext != ".inc" {
		return content, nil
	}
	return pawnfmt.Format(content, pawnfmt.Options{TabSize: 4})
}

func validOutput(output string, status bool) bool {
	if status {
		return output == "human" || output == "json"
	}
	return output == "human" || output == "json" || output == "diff"
}

func sourceFiles(root string, registry *source.Registry) ([]migrate.File, error) {
	files := make([]migrate.File, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == ".pawn") {
			return filepath.SkipDir
		}
		ext := strings.ToLower(filepath.Ext(path))
		if entry.IsDir() || (ext != ".pwn" && ext != ".inc") {
			return nil
		}
		if len(files) >= 10_000 {
			return errors.New("project contains more than 10000 Pawn files")
		}
		content, err := os.ReadFile(path) //nolint:gosec // WalkDir supplies paths below the selected project root.
		if err != nil {
			return err
		}
		files = append(files, migrate.File{Path: path, ID: registry.Intern(source.FileURI(path)), Content: string(content)})
		return nil
	})
	return files, err
}

func absolute(path string) string {
	result, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return result
}

func selection(value string) map[string]bool {
	if value == "" {
		return nil
	}
	result := make(map[string]bool)
	for id := range strings.SplitSeq(value, ",") {
		result[strings.TrimSpace(id)] = true
	}
	return result
}
