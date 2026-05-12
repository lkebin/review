// internal/ui/filelist_test.go
package ui

import (
	"strings"
	"testing"
)

func TestFileListCursorDown(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "a.go"},
		{Status: "A", Name: "b.go"},
		{Status: "D", Name: "c.go"},
	})
	fl.CursorDown()
	if fl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", fl.Cursor())
	}
}

func TestFileListCursorDownAtBottom(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "a.go"},
		{Status: "A", Name: "b.go"},
	})
	fl.CursorDown()
	fl.CursorDown()
	fl.CursorDown()
	if fl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 (clamped)", fl.Cursor())
	}
}

func TestFileListCursorUpAtTop(t *testing.T) {
	fl := NewFileList([]FileInfo{{Status: "M", Name: "a.go"}})
	fl.CursorUp()
	if fl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", fl.Cursor())
	}
}

func TestFileListGotoTop(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "a.go"},
		{Status: "A", Name: "b.go"},
		{Status: "D", Name: "c.go"},
	})
	fl.CursorDown()
	fl.CursorDown()
	fl.GotoTop()
	if fl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", fl.Cursor())
	}
}

func TestFileListGotoBottom(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "a.go"},
		{Status: "A", Name: "b.go"},
	})
	fl.GotoBottom()
	if fl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", fl.Cursor())
	}
}

func TestFileListSelectedFile(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "a.go"},
		{Status: "A", Name: "b.go"},
	})
	fl.CursorDown()
	f := fl.SelectedFile()
	if f.Name != "b.go" {
		t.Errorf("selected = %q, want b.go", f.Name)
	}
}

func TestFileListRender(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "src/main.go"},
		{Status: "A", Name: "README.md"},
	})
	th := DefaultTheme()
	output := fl.Render(30, 10, th)

	if !strings.Contains(output, "src/main.go") {
		t.Error("render missing file name src/main.go")
	}
	if !strings.Contains(output, "README.md") {
		t.Error("render missing file name README.md")
	}
}

func TestFileListCalcWidth(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "short.go"},
		{Status: "A", Name: "very/long/path/to/file.go"},
	})
	w := fl.CalcWidth()
	if w < len("very/long/path/to/file.go")+4 {
		t.Errorf("width = %d, too narrow", w)
	}
	if w > 50 {
		t.Errorf("width = %d, should cap at 50", w)
	}
}

func TestFileListEmpty(t *testing.T) {
	fl := NewFileList(nil)
	if fl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", fl.Cursor())
	}
	f := fl.SelectedFile()
	if f.Name != "" {
		t.Errorf("selected = %q, want empty", f.Name)
	}
}
