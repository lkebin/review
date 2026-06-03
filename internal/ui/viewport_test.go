// internal/ui/viewport_test.go
package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/lkebin/review/internal/diff"
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

// TestViewportRenderRowWidth verifies that every rendered row is exactly vp.width
// visible columns wide — no off-by-one that would cause terminal line-wrap.
func TestViewportRenderRowWidth(t *testing.T) {
	const vpWidth = 40
	vp := NewViewport(vpWidth, 5)
	vp.SetLines([]ViewLine{
		{LeftNo: 1, RightNo: 1, Type: diff.LineContext, Prefix: " ", RawContent: "short"},
		{LeftNo: 0, RightNo: 2, Type: diff.LineAdded, Prefix: "+", RawContent: "added line"},
		{LeftNo: 2, RightNo: 0, Type: diff.LineRemoved, Prefix: "-", RawContent: "removed line"},
		{Type: diff.LineHunkHeader, Prefix: "@@ -1,3 +1,4 @@ func foo()"},
	})

	th := DefaultTheme()
	output := vp.Render(th, 2)
	for i, row := range strings.Split(output, "\n") {
		if w := lipgloss.Width(row); w != vpWidth {
			t.Errorf("row[%d] width=%d, want %d: %q", i, w, vpWidth, row)
		}
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

// cursorVisible returns true if the cursor line is within the display window.
func cursorVisible(vp *Viewport) bool {
	if len(vp.lines) == 0 {
		return true
	}
	total := 0
	for i := vp.offset; i <= vp.cursor; i++ {
		total += vp.displayRowsFor(i)
	}
	return total <= vp.height
}

// TestDisplayRowsForCJK verifies that displayRowsFor counts visual column widths
// (CJK chars = 2 cols each) so scroll tracking matches the actual rendered output.
func TestDisplayRowsForCJK(t *testing.T) {
	// Width=20, lineNoWidth=7 → contentW=13 (rowCap=13)
	// "你好世界" = 4 CJK chars × 2 cols = 8 cols of content.
	// prefix(1) + 8 = 9 visual cols → fits in 1 row.
	vp := NewViewport(20, 10)
	vp.lineNoWidth = 7
	vp.SetLines([]ViewLine{{
		LeftNo: 1, RightNo: 1,
		Type:       diff.LineContext,
		Prefix:     " ",
		RawContent: "你好世界",
	}})
	rows := vp.displayRowsFor(0)
	if rows != 1 {
		t.Errorf("displayRowsFor CJK 4-char line = %d rows, want 1", rows)
	}

	// "你好世界你好世界" = 8 CJK chars × 2 cols = 16 cols of content.
	// prefix(1) + 16 = 17 visual cols → needs ceil(17/13) = 2 rows.
	vp.SetLines([]ViewLine{{
		LeftNo: 1, RightNo: 1,
		Type:       diff.LineContext,
		Prefix:     " ",
		RawContent: "你好世界你好世界",
	}})
	rows = vp.displayRowsFor(0)
	if rows != 2 {
		t.Errorf("displayRowsFor CJK 8-char line = %d rows, want 2", rows)
	}
}

// expandGapHelper runs ExpandGap on the viewport at the given cursor index
// with the provided gap content lines. Returns the number of lines inserted.
func expandGapHelper(vp *Viewport, cursorIdx int, gapLines []string) int {
	return vp.ExpandGap(cursorIdx, gapLines, nil, "")
}

func TestExpandGapFirstHunk(t *testing.T) {
	// Simulate a file starting at line 6 (hunk @@ -6,9 +6,9 @@)
	lines := []ViewLine{
		{Type: diff.LineHunkHeader, Prefix: "@@ -6,9 +6,9 @@", RightNo: 0},
		{Type: diff.LineContext, LeftNo: 6, RightNo: 6, Prefix: " ", RawContent: "line 6"},
		{Type: diff.LineContext, LeftNo: 7, RightNo: 7, Prefix: " ", RawContent: "line 7"},
	}
	vp := NewViewport(80, 24)
	vp.SetLines(lines)

	// Gap: no prev line, next line RightNo=6 → gap is lines 1-5
	inserted := expandGapHelper(vp, 0, []string{"line1", "line2", "line3", "line4", "line5"})
	if inserted != 5 {
		t.Fatalf("ExpandGap inserted %d lines, want 5", inserted)
	}
	if len(vp.lines) != 8 {
		t.Fatalf("total lines = %d, want 8", len(vp.lines))
	}
	// First 5 lines should be expanded, line 6 is the hunk header
	for i := 0; i < 5; i++ {
		if !vp.lines[i].Expanded {
			t.Errorf("lines[%d].Expanded = false, want true", i)
		}
		if vp.lines[i].RightNo != i+1 {
			t.Errorf("lines[%d].RightNo = %d, want %d", i, i+1, i+1)
		}
	}
	// Hunk header should be at index 5
	if vp.lines[5].Type != diff.LineHunkHeader {
		t.Errorf("lines[5] should be hunk header, got type %v", vp.lines[5].Type)
	}
	// Cursor should be past the hunk header (skipping hidden header)
	if vp.Cursor() != 6 {
		t.Errorf("cursor = %d, want 6", vp.Cursor())
	}
}

func TestExpandGapMiddleHunk(t *testing.T) {
	lines := []ViewLine{
		{Type: diff.LineContext, LeftNo: 17, RightNo: 17, Prefix: " ", RawContent: "line 17"},
		{Type: diff.LineHunkHeader, Prefix: "@@ -20,5 +20,5 @@", RightNo: 0},
		{Type: diff.LineContext, LeftNo: 20, RightNo: 20, Prefix: " ", RawContent: "line 20"},
	}
	vp := NewViewport(80, 24)
	vp.SetLines(lines)

	// Gap: prev line RightNo=17, next line RightNo=20 → gap is lines 18-19
	inserted := expandGapHelper(vp, 1, []string{"line18", "line19"})
	if inserted != 2 {
		t.Fatalf("ExpandGap inserted %d lines, want 2", inserted)
	}
	if len(vp.lines) != 5 {
		t.Fatalf("total lines = %d, want 5", len(vp.lines))
	}
	// lines[0]=line17, lines[1]=line18(exp), lines[2]=line19(exp), lines[3]=hunk, lines[4]=line20
	if !vp.lines[1].Expanded || !vp.lines[2].Expanded {
		t.Error("gap lines should be marked Expanded")
	}
	// Cursor should be past the hunk header (skipping hidden header)
	if vp.Cursor() != 4 {
		t.Errorf("cursor = %d, want 4", vp.Cursor())
	}
}

func TestExpandGapLastHunk(t *testing.T) {
	// Last hunk has no next line. Use hunk header's +start to determine gap.
	// @@ -90,3 +80,3 @@ → new_start=80
	vp := NewViewport(80, 24)
	vp.SetLines([]ViewLine{
		{Type: diff.LineContext, LeftNo: 75, RightNo: 75, Prefix: " ", RawContent: "line 75"},
		{Type: diff.LineHunkHeader, Prefix: "@@ -90,3 +80,3 @@ func end()", RightNo: 0},
	})
	// Gap: prev line RightNo=75, hunk new_start=80 → gap is lines 76-79
	inserted := expandGapHelper(vp, 1, []string{"line76", "line77", "line78", "line79"})
	if inserted != 4 {
		t.Fatalf("ExpandGap inserted %d lines, want 4", inserted)
	}
	for i := 1; i <= 4; i++ {
		if !vp.lines[i].Expanded {
			t.Errorf("gap line %d should be Expanded", i)
		}
	}
}

func TestExpandGapAlreadyExpanded(t *testing.T) {
	lines := []ViewLine{
		{Type: diff.LineContext, LeftNo: 17, RightNo: 17, Prefix: " ", RawContent: "line 17"},
		{Type: diff.LineHunkHeader, Prefix: "@@ -20,5 +20,5 @@", RightNo: 0},
		{Type: diff.LineContext, LeftNo: 20, RightNo: 20, Prefix: " ", RawContent: "line 20"},
	}
	vp := NewViewport(80, 24)
	vp.SetLines(lines)
	// First expand
	vp.ExpandGap(1, []string{"line18", "line19"}, nil, "")
	// Second expand on same hunk should be no-op
	inserted := vp.ExpandGap(vp.Cursor(), []string{"line18", "line19"}, nil, "")
	if inserted != 0 {
		t.Errorf("re-expand inserted %d lines, want 0", inserted)
	}
}

func TestCollapseGap(t *testing.T) {
	lines := []ViewLine{
		{Type: diff.LineContext, LeftNo: 17, RightNo: 17, Prefix: " ", RawContent: "line 17"},
		{Type: diff.LineHunkHeader, Prefix: "@@ -20,5 +20,5 @@", RightNo: 0},
		{Type: diff.LineContext, LeftNo: 20, RightNo: 20, Prefix: " ", RawContent: "line 20"},
	}
	vp := NewViewport(80, 24)
	vp.SetLines(lines)
	inserted := vp.ExpandGap(1, []string{"line18", "line19"}, nil, "")
	// After expand: [0]line17, [1]line18(exp), [2]line19(exp), [3]h1(exp+hidden), [4]line20
	// cursor is at 4 (past hidden header). Collapse from hunk header at index 3.
	hunkIdx := 1 + inserted // hunk header is at original position + inserted count
	removed := vp.CollapseGap(hunkIdx)
	if removed != 2 {
		t.Fatalf("CollapseGap removed %d lines, want 2", removed)
	}
	if len(vp.lines) != 3 {
		t.Fatalf("total lines = %d, want 3", len(vp.lines))
	}
	if vp.lines[0].RawContent != "line 17" || vp.lines[1].Type != diff.LineHunkHeader {
		t.Error("collapse should restore original lines")
	}
	if vp.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 (original hunk header position)", vp.Cursor())
	}
}

func TestCollapseNoExpand(t *testing.T) {
	lines := []ViewLine{
		{Type: diff.LineHunkHeader, Prefix: "@@ -6,9 +6,9 @@", RightNo: 0},
		{Type: diff.LineContext, LeftNo: 6, RightNo: 6, Prefix: " ", RawContent: "line 6"},
	}
	vp := NewViewport(80, 24)
	vp.SetLines(lines)

	removed := vp.CollapseGap(0)
	if removed != 0 {
		t.Errorf("CollapseGap on unexpanded gap removed %d lines, want 0", removed)
	}
}

func TestExpandGapCursorNotOnHunk(t *testing.T) {
	lines := []ViewLine{
		{Type: diff.LineContext, LeftNo: 1, RightNo: 1, Prefix: " ", RawContent: "line 1"},
	}
	vp := NewViewport(80, 24)
	vp.SetLines(lines)

	inserted := expandGapHelper(vp, 0, []string{"line2"})
	if inserted != 0 {
		t.Errorf("ExpandGap on non-hunk inserted %d lines, want 0", inserted)
	}
}

func TestExpandMultipleGaps(t *testing.T) {
	// Two hunks: expand middle gap then collapse first gap
	lines := []ViewLine{
		{Type: diff.LineHunkHeader, Prefix: "@@ -6,9 +6,9 @@", RightNo: 0},
		{Type: diff.LineContext, LeftNo: 6, RightNo: 6, Prefix: " ", RawContent: "line 6"},
		{Type: diff.LineHunkHeader, Prefix: "@@ -10,3 +10,3 @@", RightNo: 0},
		{Type: diff.LineContext, LeftNo: 10, RightNo: 10, Prefix: " ", RawContent: "line 10"},
	}
	vp := NewViewport(80, 24)
	vp.SetLines(lines)

	// Expand first hunk (lines 1-5)
	vp.ExpandGap(0, []string{"l1", "l2", "l3", "l4", "l5"}, nil, "")
	// After expand: [0-4]exp, [5]h1, [6]line6, [7]h2, [8]line10
	// Second hunk header is at index 7
	inserted2 := vp.ExpandGap(7, []string{"l7", "l8", "l9"}, nil, "")
	if inserted2 != 3 {
		t.Fatalf("second expand inserted %d lines, want 3", inserted2)
	}
	if len(vp.lines) != 12 {
		t.Fatalf("total lines = %d, want 12", len(vp.lines))
	}
	// Collapse first hunk (at index 5)
	vp.cursor = 5
	removed1 := vp.CollapseGap(5)
	if removed1 != 5 {
		t.Errorf("collapse first hunk removed %d lines, want 5", removed1)
	}
	// Verify second hunk's expanded lines still there
	if len(vp.lines) != 7 {
		t.Fatalf("after collapse total lines = %d, want 7", len(vp.lines))
	}
}

// TestViewportCursorVisibleWrapped verifies that the cursor stays on-screen
// when scrolling through wrapped lines.
func TestViewportCursorVisibleWrapped(t *testing.T) {
	// Width=30, lineNoWidth=7 (digitWidth=2 → 2*2+3=7)
	// contentWidth = 30-7 = 23, rowCap = 23
	// Lines with 25 content chars: runeLen=26 > 23 → 2 display rows each
	// height=10 → 5 such lines fit per screen
	vp := NewViewport(30, 10)
	vp.lineNoWidth = 7

	lines := make([]ViewLine, 20)
	for i := range lines {
		lines[i] = ViewLine{
			LeftNo: i + 1, RightNo: i + 1,
			Type:       diff.LineContext,
			Prefix:     " ",
			RawContent: strings.Repeat("x", 25),
		}
	}
	vp.SetLines(lines)

	for i := 0; i < 19; i++ {
		vp.CursorDown()
		if !cursorVisible(vp) {
			t.Errorf("CursorDown #%d: cursor %d not visible (offset=%d, height=%d)",
				i+1, vp.cursor, vp.offset, vp.height)
		}
	}
}

// TestViewportCursorVisibleHalfPage verifies cursor stays visible across half-page jumps
// when lines wrap.
func TestViewportCursorVisibleHalfPage(t *testing.T) {
	vp := NewViewport(30, 10)
	vp.lineNoWidth = 7

	lines := make([]ViewLine, 30)
	for i := range lines {
		lines[i] = ViewLine{
			LeftNo: i + 1, RightNo: i + 1,
			Type:       diff.LineContext,
			Prefix:     " ",
			RawContent: strings.Repeat("x", 25),
		}
	}
	vp.SetLines(lines)

	for i := 0; i < 10; i++ {
		vp.HalfPageDown()
		if !cursorVisible(vp) {
			t.Errorf("HalfPageDown #%d: cursor %d not visible (offset=%d, height=%d)",
				i+1, vp.cursor, vp.offset, vp.height)
		}
	}
}

// TestViewportGotoBottomVisible verifies cursor is visible after GotoBottom with wrapped lines.
func TestViewportGotoBottomVisible(t *testing.T) {
	vp := NewViewport(30, 10)
	vp.lineNoWidth = 7

	lines := make([]ViewLine, 20)
	for i := range lines {
		lines[i] = ViewLine{
			LeftNo: i + 1, RightNo: i + 1,
			Type:       diff.LineContext,
			Prefix:     " ",
			RawContent: strings.Repeat("x", 25),
		}
	}
	vp.SetLines(lines)
	vp.GotoBottom()

	if !cursorVisible(vp) {
		t.Errorf("GotoBottom: cursor %d not visible (offset=%d, height=%d)",
			vp.cursor, vp.offset, vp.height)
	}
}
