// internal/ui/filelist_test.go
package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
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

// TestTruncateNameCJK verifies that truncateName uses visual column widths so
// that wide (CJK) characters are counted as 2 columns, not 1.
func TestTruncateNameCJK(t *testing.T) {
	// "测试文件.dart": 测(2)+试(2)+文(2)+件(2)+.(1)+d(1)+a(1)+r(1)+t(1) = 13 cols, 9 runes
	name := "测试文件.dart"

	// maxWidth=8: should truncate — 13 cols > 8
	got := truncateName(name, 8)
	if got == name {
		t.Error("truncateName did not truncate a CJK name that exceeds maxWidth in columns")
	}

	// maxWidth=13: exact fit — should not truncate
	got = truncateName(name, 13)
	if got != name {
		t.Errorf("truncateName truncated a name that fits exactly: got %q", got)
	}

	// Result must not exceed maxWidth visual columns
	for _, maxW := range []int{4, 6, 8, 10, 13, 20} {
		result := truncateName(name, maxW)
		w := lipgloss.Width(result)
		if w > maxW {
			t.Errorf("truncateName(name, %d) = %q, visual width %d exceeds max", maxW, result, w)
		}
	}
}

// TestFileListRenderCJKRowWidth verifies that file list rows containing CJK
// filenames are exactly `width` visible columns wide (no overflow).
func TestFileListRenderCJKRowWidth(t *testing.T) {
	const listWidth = 20
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "测试文件.dart"},
		{Status: "A", Name: "short.go"},
	})
	th := DefaultTheme()
	output := fl.Render(listWidth, 5, th)
	for i, row := range strings.Split(output, "\n") {
		w := lipgloss.Width(row)
		if w != listWidth {
			t.Errorf("row[%d] width=%d, want %d: %q", i, w, listWidth, row)
		}
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
