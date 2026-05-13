// internal/ui/diffview_test.go
package ui

import (
	"testing"

	"github.com/kbliu/review/internal/diff"
	"github.com/kbliu/review/internal/highlight"
)

func TestBuildViewLinesBasic(t *testing.T) {
	diffLines := []diff.Line{
		{Type: diff.LineHunkHeader, Content: "@@ -1,3 +1,4 @@"},
		{Type: diff.LineContext, OldLineNo: 1, NewLineNo: 1, Content: " hello"},
		{Type: diff.LineRemoved, OldLineNo: 2, Content: "-old"},
		{Type: diff.LineAdded, NewLineNo: 2, Content: "+new"},
		{Type: diff.LineContext, OldLineNo: 3, NewLineNo: 3, Content: " end"},
	}

	hl := highlight.New("github")
	viewLines := BuildViewLines(diffLines, "main.go", hl)

	if len(viewLines) != 5 {
		t.Fatalf("viewLines count = %d, want 5", len(viewLines))
	}

	if viewLines[0].Type != diff.LineHunkHeader {
		t.Errorf("line 0 type = %v, want HunkHeader", viewLines[0].Type)
	}
	if viewLines[0].Prefix != "@@ -1,3 +1,4 @@" {
		t.Errorf("hunk prefix = %q", viewLines[0].Prefix)
	}

	if viewLines[1].LeftNo != 1 || viewLines[1].RightNo != 1 {
		t.Errorf("context line nos = %d,%d, want 1,1", viewLines[1].LeftNo, viewLines[1].RightNo)
	}
	if viewLines[1].RawContent != "hello" {
		t.Errorf("context content = %q, want 'hello'", viewLines[1].RawContent)
	}

	if viewLines[2].LeftNo != 2 || viewLines[2].RightNo != 0 {
		t.Errorf("removed line nos = %d,%d, want 2,0", viewLines[2].LeftNo, viewLines[2].RightNo)
	}
	if viewLines[2].Prefix != "-" {
		t.Errorf("removed prefix = %q, want '-'", viewLines[2].Prefix)
	}

	if viewLines[3].LeftNo != 0 || viewLines[3].RightNo != 2 {
		t.Errorf("added line nos = %d,%d, want 0,2", viewLines[3].LeftNo, viewLines[3].RightNo)
	}
}

func TestBuildViewLinesWithInlineDiff(t *testing.T) {
	diffLines := []diff.Line{
		{Type: diff.LineRemoved, OldLineNo: 1, Content: "-println(\"hello\")"},
		{Type: diff.LineAdded, NewLineNo: 1, Content: "+fmt.Println(\"hello\")"},
	}

	hl := highlight.New("github")
	viewLines := BuildViewLines(diffLines, "main.go", hl)

	if len(viewLines[0].InlineSpans) == 0 {
		t.Error("removed line should have inline spans")
	}
	if len(viewLines[1].InlineSpans) == 0 {
		t.Error("added line should have inline spans")
	}
}

func TestBuildViewLinesSyntaxTokens(t *testing.T) {
	diffLines := []diff.Line{
		{Type: diff.LineContext, OldLineNo: 1, NewLineNo: 1, Content: " func main() {}"},
	}

	hl := highlight.New("github")
	viewLines := BuildViewLines(diffLines, "main.go", hl)

	if len(viewLines[0].Tokens) == 0 {
		t.Error("context line should have syntax tokens")
	}
}

func TestCalcMaxLineNo(t *testing.T) {
	lines := []diff.Line{
		{OldLineNo: 10, NewLineNo: 15},
		{OldLineNo: 0, NewLineNo: 200},
		{OldLineNo: 50, NewLineNo: 0},
	}
	got := calcMaxLineNo(lines)
	if got != 200 {
		t.Errorf("CalcMaxLineNo = %d, want 200", got)
	}
}
