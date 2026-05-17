// internal/ui/diffview.go
package ui

import (
	"github.com/lkebin/review/internal/diff"
	"github.com/lkebin/review/internal/highlight"
)

// DiffView manages the right panel diff display.
type DiffView struct {
	viewport   *Viewport
	theme      Theme
	digitWidth int
}

// NewDiffView creates a new diff view.
func NewDiffView(width, height int, theme Theme) *DiffView {
	return &DiffView{
		viewport: NewViewport(width, height),
		theme:    theme,
	}
}

// Viewport returns the underlying viewport.
func (dv *DiffView) Viewport() *Viewport { return dv.viewport }

// DigitWidth returns the current line number digit width.
func (dv *DiffView) DigitWidth() int { return dv.digitWidth }

// LoadFile parses diff lines, applies syntax highlighting and inline diff,
// and sets the viewport content.
func (dv *DiffView) LoadFile(diffLines []diff.Line, filename string, hl *highlight.SimpleHighlighter) {
	dv.digitWidth = CalcLineNoWidth(calcMaxLineNo(diffLines))
	dv.viewport.lineNoWidth = LineNoColumnWidth(dv.digitWidth)
	viewLines := BuildViewLines(diffLines, filename, hl)
	dv.viewport.SetLines(viewLines)
}

// Render returns the rendered diff view.
func (dv *DiffView) Render() string {
	return dv.viewport.Render(dv.theme, dv.digitWidth)
}

// Resize updates dimensions.
func (dv *DiffView) Resize(width, height int) {
	dv.viewport.Resize(width, height)
}

func calcMaxLineNo(lines []diff.Line) int {
	maxNo := 0
	for _, l := range lines {
		if l.OldLineNo > maxNo {
			maxNo = l.OldLineNo
		}
		if l.NewLineNo > maxNo {
			maxNo = l.NewLineNo
		}
	}
	return maxNo
}

// BuildViewLines converts parsed diff lines into display-ready ViewLines.
// It applies syntax highlighting (via batch TokenizeFile) and inline diff.
func BuildViewLines(lines []diff.Line, filename string, hl *highlight.SimpleHighlighter) []ViewLine {
	if len(lines) == 0 {
		return nil
	}

	// 1. Extract code content for syntax highlighting (strip prefix)
	codeLines := make([]string, len(lines))
	for i, l := range lines {
		if l.Type == diff.LineHunkHeader {
			codeLines[i] = ""
		} else if len(l.Content) > 1 {
			codeLines[i] = l.Content[1:] // strip +/-/space prefix
		}
	}

	// 2. Batch tokenize
	var tokensByLine [][]highlight.Token
	if hl != nil {
		tokensByLine = hl.TokenizeFile(filename, codeLines)
	}

	// 3. Compute inline diff pairs
	pairs := PairDiffLines(lines)
	inlineMap := make(map[int][]InlineSpan)
	for _, pair := range pairs {
		oldContent := codeLines[pair.OldIdx]
		newContent := codeLines[pair.NewIdx]
		oldSpans, newSpans := ComputeInlineDiff(oldContent, newContent)
		if len(oldSpans) > 0 {
			inlineMap[pair.OldIdx] = oldSpans
		}
		if len(newSpans) > 0 {
			inlineMap[pair.NewIdx] = newSpans
		}
	}

	// 4. Build ViewLines
	viewLines := make([]ViewLine, len(lines))
	for i, l := range lines {
		vl := ViewLine{
			LeftNo:  l.OldLineNo,
			RightNo: l.NewLineNo,
			Type:    l.Type,
		}

		if l.Type == diff.LineHunkHeader {
			vl.Prefix = l.Content
			vl.RawContent = ""
		} else {
			if len(l.Content) > 0 {
				vl.Prefix = l.Content[:1]
				vl.RawContent = codeLines[i]
			}
		}

		if tokensByLine != nil && i < len(tokensByLine) {
			vl.Tokens = tokensByLine[i]
		}

		if spans, ok := inlineMap[i]; ok {
			vl.InlineSpans = spans
		}

		viewLines[i] = vl
	}

	return viewLines
}
