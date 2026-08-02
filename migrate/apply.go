package migrate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/pawnkit/pawnkit-core/textedit"
)

var ErrStalePlan = errors.New("migration plan is stale")

type Writer interface {
	ReadFile(string) ([]byte, error)
	WriteAtomic(string, []byte) error
}

type OSWriter struct{}

type fileState struct{ path, before, after string }

type ApplyOptions struct {
	Format func(path string, content []byte) ([]byte, error)
}

func (OSWriter) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path) //nolint:gosec // The caller supplies migration targets.
}

func (OSWriter) WriteAtomic(path string, content []byte) error {
	dir := filepath.Dir(path)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, ".pawnmigrate-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if _, err := temp.Write(content); err != nil {
		return errors.Join(err, temp.Close())
	}
	if err := temp.Chmod(info.Mode().Perm()); err != nil {
		return errors.Join(err, temp.Close())
	}
	if err := temp.Sync(); err != nil {
		return errors.Join(err, temp.Close())
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return atomicReplace(tempPath, path)
}

func Apply(writer Writer, plan Plan) error {
	return ApplyWithOptions(writer, plan, ApplyOptions{})
}

func ApplyWithOptions(writer Writer, plan Plan, opts ApplyOptions) error {
	byPath := make(map[string][]Change)
	for _, change := range plan.Changes {
		byPath[change.Path] = append(byPath[change.Path], change)
	}
	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	stagedFiles := make([]fileState, 0, len(byPath))
	for _, path := range paths {
		changes := byPath[path]
		sort.SliceStable(changes, func(i, j int) bool { return changes[i].order < changes[j].order })
		content, err := writer.ReadFile(path)
		if err != nil {
			return err
		}
		before := string(content)
		after, err := applyChanges(path, before, changes)
		if err != nil {
			return err
		}
		if opts.Format != nil {
			formatted, err := opts.Format(path, []byte(after))
			if err != nil {
				return fmt.Errorf("format %s: %w", path, err)
			}
			after = string(formatted)
		}
		stagedFiles = append(stagedFiles, fileState{path: path, before: before, after: after})
	}
	written := make([]fileState, 0, len(stagedFiles))
	for _, file := range stagedFiles {
		if err := writer.WriteAtomic(file.path, []byte(file.after)); err != nil {
			rollbackErr := rollback(writer, written)
			return errors.Join(fmt.Errorf("write %s: %w", file.path, err), rollbackErr)
		}
		written = append(written, file)
	}
	return nil
}

func applyChanges(path, before string, changes []Change) (string, error) {
	if len(changes) == 0 {
		return before, nil
	}
	allSameBefore := true
	for _, change := range changes[1:] {
		if change.Before != changes[0].Before {
			allSameBefore = false
			break
		}
	}
	if allSameBefore {
		if before != changes[0].Before {
			return "", fmt.Errorf("%w: %s changed after planning", ErrStalePlan, path)
		}
		var edits []textedit.Edit
		for _, change := range changes {
			edits = append(edits, change.Edits...)
		}
		after, err := textedit.Apply(before, edits)
		if err != nil {
			return "", err
		}
		return after, nil
	}

	current := before
	for _, change := range changes {
		if current != change.Before {
			return "", fmt.Errorf("%w: %s changed after planning", ErrStalePlan, path)
		}
		var err error
		current, err = textedit.Apply(current, change.Edits)
		if err != nil {
			return "", err
		}
	}
	return current, nil
}

func rollback(writer Writer, files []fileState) error {
	var result error
	for _, file := range slices.Backward(files) {
		if err := writer.WriteAtomic(file.path, []byte(file.before)); err != nil {
			result = errors.Join(result, fmt.Errorf("rollback %s: %w", file.path, err))
		}
	}
	return result
}
