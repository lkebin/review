// internal/ui/viewport.go
package ui

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/kbliu/review/internal/diff"
	"github.com/kbliu/review/internal/highlight"
	"github.com/mattn/go-runewidth"
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
	width       int
	height      int
	offset      int // first visible logical line
	cursor      int // cursor logical line
	lines       []ViewLine
	lineNoWidth int // width of the line-number column; kept in sync by DiffView
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

// SearchNext moves cursor to the next line whose content contains query (case-insensitive, wraps around).
// Returns true if the cursor moved.
func (vp *Viewport) SearchNext(query string) bool {
	if query == "" || len(vp.lines) == 0 {
		return false
	}
	q := strings.ToLower(query)
	for i := 1; i <= len(vp.lines); i++ {
		idx := (vp.cursor + i) % len(vp.lines)
		line := vp.lines[idx]
		if line.Type == diff.LineHunkHeader {
			continue
		}
		if strings.Contains(strings.ToLower(line.RawContent), q) {
			vp.cursor = idx
			vp.ensureVisible()
			return true
		}
	}
	return false
}

// SearchPrev moves cursor to the previous line whose content contains query (case-insensitive, wraps around).
// Returns true if the cursor moved.
func (vp *Viewport) SearchPrev(query string) bool {
	if query == "" || len(vp.lines) == 0 {
		return false
	}
	q := strings.ToLower(query)
	n := len(vp.lines)
	for i := 1; i <= n; i++ {
		idx := (vp.cursor - i + n) % n
		line := vp.lines[idx]
		if line.Type == diff.LineHunkHeader {
			continue
		}
		if strings.Contains(strings.ToLower(line.RawContent), q) {
			vp.cursor = idx
			vp.ensureVisible()
			return true
		}
	}
	return false
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

// displayRowsFor returns how many terminal rows logical line idx occupies.
// It counts rune length (prefix + content) and divides by the row capacity,
// mirroring the wrapping logic in Render without needing ANSI rendering.
func (vp *Viewport) displayRowsFor(idx int) int {
	if idx < 0 || idx >= len(vp.lines) {
		return 1
	}
	line := vp.lines[idx]
	if line.Type == diff.LineHunkHeader {
		return 1 // always truncated to width, never wraps
	}
	contentW := vp.width - vp.lineNoWidth
	if contentW < 1 {
		contentW = 1
	}
	rowCap := contentW + 1 // each display row: 1 prefix + contentW content chars
	runeLen := 1 + len([]rune(line.RawContent))
	rows := (runeLen + rowCap - 1) / rowCap
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (vp *Viewport) ensureVisible() {
	if len(vp.lines) == 0 || vp.height == 0 {
		return
	}
	// Cursor above the window → scroll up so cursor is at the top.
	if vp.cursor < vp.offset {
		vp.offset = vp.cursor
		if vp.offset < 0 {
			vp.offset = 0
		}
		return
	}
	// Count display rows from offset through cursor (inclusive).
	// Advance offset until cursor's rows fit within the viewport height.
	total := 0
	for i := vp.offset; i <= vp.cursor; i++ {
		total += vp.displayRowsFor(i)
	}
	for total > vp.height && vp.offset < vp.cursor {
		total -= vp.displayRowsFor(vp.offset)
		vp.offset++
	}
	// Do NOT call clampOffset here — its logical-line maxOffset would roll back
	// the display-row-aware offset we just computed.
}

// clampOffset prevents the scroll position from going past the last content.
// It uses display-row counts so that wrapped lines are accounted for correctly.
func (vp *Viewport) clampOffset() {
	if vp.offset < 0 {
		vp.offset = 0
	}
	if len(vp.lines) == 0 {
		vp.offset = 0
		return
	}
	// Walk backward from the last line, accumulating display rows until we
	// reach vp.height. That line index is the maximum valid offset.
	total := 0
	maxOffset := 0
	for i := len(vp.lines) - 1; i >= 0; i-- {
		total += vp.displayRowsFor(i)
		if total >= vp.height {
			maxOffset = i
			break
		}
	}
	if vp.offset > maxOffset {
		vp.offset = maxOffset
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
				// Pad each chunk to fill the full content width so the line
				// background color extends to the right edge of the panel.
				if wrW := lipgloss.Width(wr); wrW < contentWidth+1 {
					padSt := lipgloss.NewStyle()
					if bgColor != "" {
						padSt = padSt.Background(lipgloss.Color(bgColor))
					}
					wr += padSt.Render(strings.Repeat(" ", contentWidth+1-wrW))
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

// findANSISafeBreak returns a byte index in s where the visible column count
// reaches maxWidth, walking by rune (never inside a multi-byte UTF-8 sequence)
// and accounting for wide characters. ANSI SGR escapes (\x1b...m) are skipped.
func findANSISafeBreak(s string, maxWidth int) int {
	visible := 0
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			j := i + 1
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j < len(s) {
				j++ // consume the terminating 'm'
			}
			i = j
			continue
		}
		r, sz := utf8.DecodeRuneInString(s[i:])
		w := runewidth.RuneWidth(r)
		if w == 0 {
			// Zero-width (combining marks, ZWJ, …): attaches to previous rune.
			i += sz
			continue
		}
		if visible+w > maxWidth {
			// Including this rune would exceed maxWidth; break before it.
			return i
		}
		visible += w
		i += sz
		if visible >= maxWidth {
			return i
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
