package migrate_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
	"github.com/pawnkit/pawnmigrate/migrate"
)

type memoryWriter struct {
	files  map[string][]byte
	failAt string
}

func (m *memoryWriter) ReadFile(path string) ([]byte, error) {
	return append([]byte(nil), m.files[path]...), nil
}

func (m *memoryWriter) WriteAtomic(path string, content []byte) error {
	if path == m.failAt {
		m.failAt = ""
		return errors.New("injected failure")
	}
	m.files[path] = append([]byte(nil), content...)
	return nil
}

func TestApplyRollsBackWrittenFiles(t *testing.T) {
	writer := &memoryWriter{files: map[string][]byte{"a": []byte("a"), "b": []byte("b")}, failAt: "b"}
	plan := migrate.Plan{Changes: []migrate.Change{
		{Path: "a", Before: "a", Edits: []textedit.Edit{{Span: source.Span{File: 1, Start: 0, End: 1}, NewText: "A"}}},
		{Path: "b", Before: "b", Edits: []textedit.Edit{{Span: source.Span{File: 2, Start: 0, End: 1}, NewText: "B"}}},
	}}
	if err := migrate.Apply(writer, plan); err == nil {
		t.Fatal("expected failure")
	}
	if string(writer.files["a"]) != "a" || string(writer.files["b"]) != "b" {
		t.Fatalf("files were not rolled back: %#v", writer.files)
	}
}

func TestApplyFormatsBeforeWriting(t *testing.T) {
	writer := &memoryWriter{files: map[string][]byte{"a.pwn": []byte("a")}}
	plan := migrate.Plan{Changes: []migrate.Change{{Path: "a.pwn", Before: "a", Edits: []textedit.Edit{{Span: source.Span{File: 1, Start: 0, End: 1}, NewText: "A"}}}}}
	err := migrate.ApplyWithOptions(writer, plan, migrate.ApplyOptions{Format: func(_ string, content []byte) ([]byte, error) {
		return append(content, '\n'), nil
	}})
	if err != nil || string(writer.files["a.pwn"]) != "A\n" {
		t.Fatalf("content/error = %q %v", writer.files["a.pwn"], err)
	}
}

func TestApplyRejectsStalePlanBeforeWriting(t *testing.T) {
	writer := &memoryWriter{files: map[string][]byte{"a.pwn": []byte("changed"), "b.pwn": []byte("b")}}
	plan := migrate.Plan{Changes: []migrate.Change{
		{Path: "a.pwn", Before: "a", Edits: []textedit.Edit{{Span: source.Span{File: 1, Start: 0, End: 1}, NewText: "A"}}},
		{Path: "b.pwn", Before: "b", Edits: []textedit.Edit{{Span: source.Span{File: 2, Start: 0, End: 1}, NewText: "B"}}},
	}}
	err := migrate.Apply(writer, plan)
	if !errors.Is(err, migrate.ErrStalePlan) {
		t.Fatalf("error = %v", err)
	}
	if string(writer.files["b.pwn"]) != "b" {
		t.Fatal("another file was written before stale-plan rejection")
	}
}

func TestOSWriterReplacesFileAndPreservesMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.pwn")
	if err := os.WriteFile(path, []byte("before"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := (migrate.OSWriter{}).WriteAtomic(path, []byte("after")); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "after" {
		t.Fatalf("content/mode = %q %o", content, info.Mode().Perm())
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o640 {
		t.Fatalf("mode = %o, want 640", info.Mode().Perm())
	}
}
