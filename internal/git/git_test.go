package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFileContent(t *testing.T) {
	// Create a temp file in a temp dir
	dir := t.TempDir()
	fpath := filepath.Join(dir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(fpath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to the temp dir so the relative path resolves
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })

	lines, err := GetFileContent("test.txt", false)
	if err != nil {
		t.Fatalf("GetFileContent() error: %v", err)
	}
	if len(lines) != 5 {
		t.Fatalf("got %d lines, want 5", len(lines))
	}
	if lines[0] != "line1" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "line1")
	}
	if lines[4] != "line5" {
		t.Errorf("lines[4] = %q, want %q", lines[4], "line5")
	}
}

func TestGetFileContentNonexistent(t *testing.T) {
	_, err := GetFileContent("nonexistent.txt", false)
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestGetFileContentEmptyFile(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(fpath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })

	lines, err := GetFileContent("empty.txt", false)
	if err != nil {
		t.Fatalf("GetFileContent() error: %v", err)
	}
	if lines != nil {
		t.Errorf("expected nil for empty file, got %v", lines)
	}
}

func TestGetFileContentSingleLineNoNewline(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "single.txt")
	if err := os.WriteFile(fpath, []byte("just this"), 0644); err != nil {
		t.Fatal(err)
	}
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })

	lines, err := GetFileContent("single.txt", false)
	if err != nil {
		t.Fatalf("GetFileContent() error: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	if lines[0] != "just this" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "just this")
	}
}

func TestGetFileContentTrailingNewlines(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "trailing.txt")
	if err := os.WriteFile(fpath, []byte("a\n\n\n"), 0644); err != nil {
		t.Fatal(err)
	}
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })

	lines, err := GetFileContent("trailing.txt", false)
	if err != nil {
		t.Fatalf("GetFileContent() error: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3 ('a', '', '')", len(lines))
	}
}
