// internal/ui/viewport.go
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kbliu/review/internal/diff"
	"github.com/kbliu/review/internal/highlight"
)

// ViewLine is a display-ready diff line for the viewport.
type ViewLine struct {
	LeftNo      int
	RightNo     int
	Type        diff.LineType
	Prefix      string            // "+", "-", " ", or "@@..."
	RawContent  string            // plain text, no ANSI
	Tokens      []highlight.Token // syntax highlight tokens
	InlineSpans []InlineSpan      // character-level diff emphasis
}

// Viewport manages scrollable content with cursor tracking.
type Viewport struct {
	width  int
	height int
	offset int // first visible logical line
	cursor int // cursor logical line
	lines  []ViewLine
}

// NewViewport creates a viewport with the given dimensions.
func NewViewport(width, height int) *Viewport {
	return &Viewport{width: width, height: height}
}

// SetLines replaces the content and resets cursor/offset.
func (vp *Viewport) SetLines(lines []ViewLine) {
	vp.lines = lines
	vp.cursor = 0
	vp.offset = 0
}

// Resize updates viewport dimensions.
func (vp *Viewport) Resize(width, height int) {
	vp.width = width
	vp.height = height
	vp.ensureVisible()
}

// Cursor returns the current cursor position.
func (vp *Viewport) Cursor() int { return vp.cursor }

// Offset returns the scroll offset.
func (vp *Viewport) Offset() int { return vp.offset }

// Lines returns the viewport's lines.
func (vp *Viewport) Lines() []ViewLine { return vp.lines }

// LineCount returns the number of logical lines.
func (vp *Viewport) LineCount() int { return len(vp.lines) }

// CursorDown moves the cursor down one line.
func (vp *Viewport) CursorDown() {
	if len(vp.lines) == 0 {
		return
	}
	if vp.cursor < len(vp.lines)-1 {
		vp.cursor++
	}
	vp.ensureVisible()
}

// CursorUp moves the cursor up one line.
func (vp *Viewport) CursorUp() {
	if vp.cursor > 0 {
		vp.cursor--
	}
	vp.ensureVisible()
}

// HalfPageDown scrolls down half a page.
func (vp *Viewport) HalfPageDown() {
	if len(vp.lines) == 0 {
		return
	}
	half := vp.height / 2
	if half < 1 {
		half = 1
	}
	vp.cursor += half
	if vp.cursor >= len(vp.lines) {
		vp.cursor = len(vp.lines) - 1
	}
	vp.offset += half
	vp.clampOffset()
	vp.ensureVisible()
}

// HalfPageUp scrolls up half a page.
func (vp *Viewport) HalfPageUp() {
	half := vp.height / 2
	if half < 1 {
		half = 1
	}
	vp.cursor -= half
	if vp.cursor < 0 {
		vp.cursor = 0
	}
	vp.offset -= half
	if vp.offset < 0 {
		vp.offset = 0
	}
	vp.ensureVisible()
}

// PageDown scrolls down one full page.
func (vp *Viewport) PageDown() {
	if len(vp.lines) == 0 {
		return
	}
	vp.cursor += vp.height
	if vp.cursor >= len(vp.lines) {
		vp.cursor = len(vp.lines) - 1
	}
	vp.offset += vp.height
	vp.clampOffset()
	vp.ensureVisible()
}

// PageUp scrolls up one full page.
func (vp *Viewport) PageUp() {
	vp.cursor -= vp.height
	if vp.cursor < 0 {
		vp.cursor = 0
	}
	vp.offset -= vp.height
	if vp.offset < 0 {
		vp.offset = 0
	}
	vp.ensureVisible()
}

// GotoTop jumps to the first line.
func (vp *Viewport) GotoTop() {
	vp.cursor = 0
	vp.offset = 0
}

// GotoBottom jumps to the last line.
func (vp *Viewport) GotoBottom() {
	if len(vp.lines) == 0 {
		return
	}
	vp.cursor = len(vp.lines) - 1
	vp.ensureVisible()
}

// NextHunk moves cursor to the next hunk header after current position.
func (vp *Viewport) NextHunk() {
	for i := vp.cursor + 1; i < len(vp.lines); i++ {
		if vp.lines[i].Type == diff.LineHunkHeader {
			vp.cursor = i
			vp.ensureVisible()
			return
		}
	}
}

// PrevHunk moves cursor to the previous hunk header before current position.
func (vp *Viewport) PrevHunk() {
	for i := vp.cursor - 1; i >= 0; i-- {
		if vp.lines[i].Type == diff.LineHunkHeader {
			vp.cursor = i
			vp.ensureVisible()
			return
		}
	}
}

func (vp *Viewport) ensureVisible() {
	if vp.cursor < vp.offset {
		vp.offset = vp.cursor
	}
	if vp.height > 0 && vp.cursor >= vp.offset+vp.height {
		vp.offset = vp.cursor - vp.height + 1
	}
	vp.clampOffset()
}

func (vp *Viewport) clampOffset() {
	if len(vp.lines) == 0 {
		vp.offset = 0
		return
	}
	maxOffset := len(vp.lines) - vp.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if vp.offset > maxOffset {
		vp.offset = maxOffset
	}
	if vp.offset < 0 {
		vp.offset = 0
	}
}

// Render returns the visible portion of the viewport as a styled string.
// Layout per row: │ [line numbers] [content]
// Handles content wrapping: long lines produce continuation rows with blank left columns.
func (vp *Viewport) Render(theme Theme, digitWidth int) string {
	if len(vp.lines) == 0 || vp.height <= 0 || vp.width <= 0 {
		return ""
	}

	lineNoWidth := LineNoColumnWidth(digitWidth)
	contentWidth := vp.width - lineNoWidth
	if contentWidth < 1 {
		contentWidth = 1
	}

	var rows []string
	displayLines := 0

	end := len(vp.lines)

	for i := vp.offset; i < end && displayLines < vp.height; i++ {
		line := vp.lines[i]
		isCursor := i == vp.cursor

		bgColor := vp.lineBackground(line.Type, isCursor, theme)

		// Line number style carries the line background color.
		lineNoSt := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.LineNoFg))
		if bgColor != "" {
			lineNoSt = lineNoSt.Background(lipgloss.Color(bgColor))
		}

		// Format line numbers — hunk headers show "··"
		var lineNoPart string
		if line.Type == diff.LineHunkHeader {
			dots := strings.Repeat("·", digitWidth)
			lineNoPart = lineNoSt.Render(" " + dots + " " + dots + " ")
		} else {
			lineNo := FormatLineNo(line.LeftNo, line.RightNo, digitWidth)
			lineNoPart = lineNoSt.Render(lineNo)
		}

		// Render the full content line
		fullContent := vp.renderContent(line, contentWidth, bgColor, theme)
		visibleLen := lipgloss.Width(fullContent)

		// Blank line-number column for continuation rows: carries bgColor so
		// removed/added lines keep their background in that area.
		blankLineNoPart := lineNoSt.Render(strings.Repeat(" ", lineNoWidth))

		if visibleLen <= contentWidth+1 { // +1 for prefix char
			// Append reset so unclosed ANSI sequences don't bleed into the next row.
			rows = append(rows, lineNoPart+fullContent+"\x1b[0m")
			displayLines++
		} else {
			// Content wrapping — continuation rows use blank line-number columns.
			wrappedRows := wrapRenderedLine(fullContent, contentWidth+1)

			for wi, wr := range wrappedRows {
				if displayLines >= vp.height {
					break
				}
				if wi == 0 {
					rows = append(rows, lineNoPart+wr+"\x1b[0m")
				} else {
					rows = append(rows, blankLineNoPart+wr+"\x1b[0m")
				}
				displayLines++
			}
		}
	}

	// Pad remaining height with blank rows.
	for displayLines < vp.height {
		emptyContent := strings.Repeat(" ", lineNoWidth+contentWidth)
		rows = append(rows, emptyContent+"\x1b[0m")
		displayLines++
	}

	return strings.Join(rows, "\n")
}

// wrapRenderedLine splits a rendered (ANSI-containing) line into chunks of
// maxVisibleWidth visible characters.
func wrapRenderedLine(rendered string, maxVisibleWidth int) []string {
	if maxVisibleWidth <= 0 {
		return []string{rendered}
	}
	total := lipgloss.Width(rendered)
	if total <= maxVisibleWidth {
		return []string{rendered}
	}

	var result []string
	remaining := rendered
	for lipgloss.Width(remaining) > maxVisibleWidth {
		bp := findANSISafeBreak(remaining, maxVisibleWidth)
		result = append(result, remaining[:bp])
		remaining = remaining[bp:]
	}
	if len(remaining) > 0 {
		result = append(result, remaining)
	}
	return result
}

// findANSISafeBreak returns a byte index in s where visible character count
// reaches maxWidth, correctly skipping ANSI escape sequences.
func findANSISafeBreak(s string, maxWidth int) int {
	visible := 0
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		visible++
		if visible >= maxWidth {
			return i + 1
		}
	}
	return len(s)
}

func (vp *Viewport) lineBackground(lt diff.LineType, isCursor bool, th Theme) string {
	if isCursor {
		switch lt {
		case diff.LineAdded:
			return th.AddedCursorBg
		case diff.LineRemoved:
			return th.RemovedCursorBg
		default:
			return th.CursorBg
		}
	}
	switch lt {
	case diff.LineAdded:
		return th.AddedBg
	case diff.LineRemoved:
		return th.RemovedBg
	default:
		return ""
	}
}

func (vp *Viewport) renderContent(line ViewLine, width int, bgColor string, th Theme) string {
	if line.Type == diff.LineHunkHeader {
		hunkText := line.Prefix
		if len(hunkText) > width {
			hunkText = hunkText[:width]
		}
		padded := hunkText + strings.Repeat(" ", max(0, width-lipgloss.Width(hunkText)))
		return th.HunkStyle().Render(padded)
	}

	prefix := line.Prefix
	raw := line.RawContent

	var result strings.Builder

	bgStyle := lipgloss.NewStyle()
	if bgColor != "" {
		bgStyle = bgStyle.Background(lipgloss.Color(bgColor))
	}
	result.WriteString(bgStyle.Render(prefix))

	// Build inline span lookup
	inlineSet := make(map[int]bool)
	for _, span := range line.InlineSpans {
		for b := span.Start; b < span.End && b < len(raw); b++ {
			inlineSet[b] = true
		}
	}

	var emphBg string
	switch line.Type {
	case diff.LineAdded:
		emphBg = th.InlineAddBg
	case diff.LineRemoved:
		emphBg = th.InlineDelBg
	}

	if len(line.Tokens) > 0 {
		byteOffset := 0
		for _, tok := range line.Tokens {
			for _, r := range tok.Text {
				rs := string(r)
				rLen := len(rs)
				style := lipgloss.NewStyle()
				color := getTokenColor(tok.TokenType)
				if color != "" {
					style = style.Foreground(lipgloss.Color(color))
				}
				if inlineSet[byteOffset] && emphBg != "" {
					style = style.Background(lipgloss.Color(emphBg))
				} else if bgColor != "" {
					style = style.Background(lipgloss.Color(bgColor))
				}
				result.WriteString(style.Render(rs))
				byteOffset += rLen
			}
		}
	} else {
		for i, r := range raw {
			rs := string(r)
			style := lipgloss.NewStyle()
			if bgColor != "" {
				style = style.Background(lipgloss.Color(bgColor))
			}
			if inlineSet[i] && emphBg != "" {
				style = style.Background(lipgloss.Color(emphBg))
			}
			result.WriteString(style.Render(rs))
		}
	}

	// Pad to fill content width
	rendered := result.String()
	visibleWidth := lipgloss.Width(rendered)
	if visibleWidth < width+1 { // +1 for prefix
		padLen := width + 1 - visibleWidth
		padStyle := lipgloss.NewStyle()
		if bgColor != "" {
			padStyle = padStyle.Background(lipgloss.Color(bgColor))
		}
		rendered += padStyle.Render(strings.Repeat(" ", padLen))
	}

	return rendered
}

func getTokenColor(tokenType string) string {
	switch {
	case strings.Contains(tokenType, "Keyword"):
		return "204"
	case strings.Contains(tokenType, "String"):
		return "192"
	case strings.Contains(tokenType, "Comment"):
		return "243"
	case strings.Contains(tokenType, "Number"):
		return "180"
	case strings.Contains(tokenType, "Function"):
		return "117"
	default:
		return ""
	}
}
