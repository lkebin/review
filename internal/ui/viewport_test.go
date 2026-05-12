// internal/ui/viewport_test.go
package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kbliu/review/internal/diff"
)

func makeTestLines(n int) []ViewLine {
	lines := make([]ViewLine, n)
	for i := range lines {
		lines[i] = ViewLine{
			LeftNo:     i + 1,
			RightNo:    i + 1,
			Type:       diff.LineContext,
			Prefix:     " ",
			RawContent: fmt.Sprintf("line %d content", i+1),
		}
	}
	return lines
}

func TestViewportCursorDown(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(30))

	for i := 0; i < 5; i++ {
		vp.CursorDown()
	}
	if vp.Cursor() != 5 {
		t.Errorf("cursor = %d, want 5", vp.Cursor())
	}
	if vp.Offset() != 0 {
		t.Errorf("offset = %d, want 0", vp.Offset())
	}
}

func TestViewportCursorDownScrolls(t *testing.T) {
	vp := NewViewport(80, 5)
	vp.SetLines(makeTestLines(20))

	for i := 0; i < 4; i++ {
		vp.CursorDown()
	}
	if vp.Offset() != 0 {
		t.Errorf("offset = %d, want 0 (cursor still visible)", vp.Offset())
	}

	vp.CursorDown()
	if vp.Cursor() != 5 {
		t.Errorf("cursor = %d, want 5", vp.Cursor())
	}
	if vp.Offset() < 1 {
		t.Errorf("offset = %d, should have scrolled", vp.Offset())
	}
}

func TestViewportCursorUp(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(20))
	vp.CursorDown()
	vp.CursorDown()
	vp.CursorUp()
	if vp.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", vp.Cursor())
	}
}

func TestViewportCursorUpAtTop(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(20))
	vp.CursorUp()
	if vp.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", vp.Cursor())
	}
}

func TestViewportCursorDownAtBottom(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(5))
	for i := 0; i < 10; i++ {
		vp.CursorDown()
	}
	if vp.Cursor() != 4 {
		t.Errorf("cursor = %d, want 4 (last line)", vp.Cursor())
	}
}

func TestViewportHalfPageDown(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(30))

	vp.HalfPageDown()
	if vp.Cursor() != 5 {
		t.Errorf("cursor = %d, want 5 (half page = 5)", vp.Cursor())
	}
}

func TestViewportGotoTop(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(30))
	for i := 0; i < 15; i++ {
		vp.CursorDown()
	}
	vp.GotoTop()
	if vp.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", vp.Cursor())
	}
	if vp.Offset() != 0 {
		t.Errorf("offset = %d, want 0", vp.Offset())
	}
}

func TestViewportGotoBottom(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(30))
	vp.GotoBottom()
	if vp.Cursor() != 29 {
		t.Errorf("cursor = %d, want 29", vp.Cursor())
	}
}

func TestViewportNextHunk(t *testing.T) {
	lines := []ViewLine{
		{Type: diff.LineHunkHeader, Prefix: "@@"},
		{Type: diff.LineContext, Prefix: " "},
		{Type: diff.LineAdded, Prefix: "+"},
		{Type: diff.LineHunkHeader, Prefix: "@@"},
		{Type: diff.LineRemoved, Prefix: "-"},
	}
	vp := NewViewport(80, 10)
	vp.SetLines(lines)

	vp.NextHunk()
	if vp.Cursor() != 3 {
		t.Errorf("cursor = %d, want 3 (second hunk header)", vp.Cursor())
	}
	vp.NextHunk()
	if vp.Cursor() != 3 {
		t.Errorf("cursor = %d, want 3 (no more hunks)", vp.Cursor())
	}
}

func TestViewportPrevHunk(t *testing.T) {
	lines := []ViewLine{
		{Type: diff.LineHunkHeader, Prefix: "@@"},
		{Type: diff.LineContext, Prefix: " "},
		{Type: diff.LineHunkHeader, Prefix: "@@"},
		{Type: diff.LineRemoved, Prefix: "-"},
	}
	vp := NewViewport(80, 10)
	vp.SetLines(lines)
	vp.CursorDown()
	vp.CursorDown()
	vp.CursorDown()

	vp.PrevHunk()
	if vp.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2 (second hunk)", vp.Cursor())
	}
	vp.PrevHunk()
	if vp.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 (first hunk)", vp.Cursor())
	}
}

func TestViewportEmptyLines(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(nil)
	vp.CursorDown()
	vp.CursorUp()
	vp.HalfPageDown()
	vp.GotoBottom()
	if vp.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", vp.Cursor())
	}
}

func TestViewportRenderBasic(t *testing.T) {
	vp := NewViewport(40, 3)
	vp.SetLines([]ViewLine{
		{LeftNo: 1, RightNo: 1, Type: diff.LineContext, Prefix: " ", RawContent: "hello"},
		{LeftNo: 0, RightNo: 2, Type: diff.LineAdded, Prefix: "+", RawContent: "world"},
		{LeftNo: 2, RightNo: 0, Type: diff.LineRemoved, Prefix: "-", RawContent: "old"},
	})

	th := DefaultTheme()
	output := vp.Render(th, 2)
	if output == "" {
		t.Fatal("Render returned empty")
	}
	lines := strings.Split(output, "\n")
	if len(lines) != 3 {
		t.Errorf("rendered %d lines, want 3", len(lines))
	}
}

func TestViewportRenderWrapping(t *testing.T) {
	vp := NewViewport(20, 5)
	vp.SetLines([]ViewLine{
		{LeftNo: 1, RightNo: 1, Type: diff.LineContext, Prefix: " ",
			RawContent: "this is a very long line that should wrap"},
	})

	th := DefaultTheme()
	output := vp.Render(th, 2)
	lines := strings.Split(output, "\n")
	if len(lines) != 5 {
		t.Errorf("rendered %d display lines, want 5 (wrapping + padding)", len(lines))
	}
}

func TestViewportRenderEmpty(t *testing.T) {
	vp := NewViewport(40, 5)
	th := DefaultTheme()
	output := vp.Render(th, 4)
	if output != "" {
		t.Errorf("empty viewport should render empty, got %q", output)
	}
}
