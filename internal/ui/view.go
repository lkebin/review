package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kbliu/review/internal/highlight"
)

// Styles
var (
	diffStyle = lipgloss.NewStyle()

	addedBgStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("22")) // dark green bg

	removedBgStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("52")) // dark red bg

	currentLineBg = lipgloss.NewStyle().
			Background(lipgloss.Color("238")) // subtle highlight for cursor line

	hunkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Gray

	lineNoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Gray

	helpPopupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	helpTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("208"))

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	helpDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	topBrandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Background(lipgloss.Color("233")).
			Bold(true)

	topPathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("233"))

	topMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Background(lipgloss.Color("236")).
			Bold(true)

	bottomBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("23")).
			Background(lipgloss.Color("51"))

	railContainerStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("235")).
				Foreground(lipgloss.Color("250"))

	railIdentityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true)

	railMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	railSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("87"))

	railSectionActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("16")).
				Background(lipgloss.Color("46")).
				Bold(true)

	railSecondaryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))

	railFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	railSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("16")).
				Background(lipgloss.Color("51")).
				Bold(true)

	workspaceHeaderStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("233")).
				Foreground(lipgloss.Color("252")).
				Padding(1, 1, 0, 1)

	workspaceLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("87"))

	workspaceTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Bold(true)

	workspaceAddStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true)

	workspaceRemoveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("203")).
				Bold(true)
)

// View renders the UI
func (m Model) View() string {
	if m.loading && len(m.files) == 0 {
		return "Loading..."
	}

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\nPress q to quit."
	}

	if len(m.files) == 0 {
		return "No changes found.\n\nPress q to quit."
	}

	if m.showHelp {
		return m.renderHelpOverlay()
	}

	return m.renderShell()
}

func (m Model) renderShell() string {
	var body string
	if m.layout == LayoutHorizontal {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderNavRail(), m.renderWorkspace())
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left, m.renderNavRail(), m.renderWorkspace())
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderTopHeader(),
		body,
		m.renderBottomBar(),
	)
}

func (m Model) renderTopHeader() string {
	path := "NO_FILE_SELECTED"
	if m.currentFile != "" {
		path = strings.ToUpper(m.currentFile)
	}

	left := topBrandStyle.Render(" CYBERNETIC_MANUSCRIPT ")
	right := topMetaStyle.Render(" ROOT_USER ")

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	centerW := m.width - leftW - rightW
	if centerW < 0 {
		centerW = 0
	}

	displayPath := truncateText(path, max(centerW-2, 0))
	center := fitWidth(" "+displayPath+" ", centerW)

	return left + topPathStyle.Render(center) + right
}

func (m Model) renderBottomBar() string {
	branch := m.options.Target
	if branch == "" {
		branch = "HEAD"
	}
	left := bottomBarStyle.Render(" BRANCH: " + branch + " ")
	center := bottomBarStyle.Render(fmt.Sprintf(" FILES: %d ", len(m.files)))
	right := bottomBarStyle.Render(" [Q]UIT ")

	leftW := lipgloss.Width(left)
	centerW := lipgloss.Width(center)
	rightW := lipgloss.Width(right)
	fillW := m.width - leftW - centerW - rightW
	if fillW < 0 {
		fillW = 0
	}

	fill := bottomBarStyle.Width(fillW).Render("")
	return left + center + fill + right
}

func (m Model) renderNavRail() string {
	width := m.listWidth
	height := m.getContentHeight()
	if width <= 0 || height <= 0 {
		return ""
	}

	var fileLines []string
	for i, f := range m.files {
		row := fmt.Sprintf("[%s] %s", f.Status, strings.ToUpper(f.Name))
		if i == m.cursor {
			row = railSelectedStyle.Render(" " + row + " ")
		} else {
			row = railFileStyle.Render(" " + row)
		}
		fileLines = append(fileLines, row)
	}

	content := strings.Join([]string{
		railIdentityStyle.Render(" TERMINAL_UI "),
		railMetaStyle.Render(" V1.0.0-ALPHA "),
		"",
		railSectionStyle.Render(" PROJECT_TREE "),
		railSectionActiveStyle.Render(" FILES "),
		strings.Join(fileLines, "\n"),
		"",
		railSecondaryStyle.Render(" HISTORY "),
		railSecondaryStyle.Render(" CONFIG "),
	}, "\n")

	return railContainerStyle.Width(width).Height(height).Render(content)
}

func (m Model) renderWorkspace() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderWorkspaceHeader(),
		m.renderDiffView(),
	)
}

func formatStatCount(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	lead := len(s) % 3
	if lead == 0 {
		lead = 3
	}
	b.WriteString(s[:lead])
	for i := lead; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

func (m Model) renderWorkspaceHeader() string {
	title := "NO_FILE_SELECTED"
	added := 0
	removed := 0
	if m.currentFile != "" {
		title = strings.ToUpper(m.currentFile)
	}
	for _, f := range m.files {
		if f.Name == m.currentFile {
			added = f.Added
			removed = f.Removed
			break
		}
	}

	left := workspaceLabelStyle.Render(" EDITING_FILE: ") + "\n" +
		workspaceTitleStyle.Render(" "+title)
	right := workspaceAddStyle.Render("+"+formatStatCount(added)) + " " +
		workspaceRemoveStyle.Render("-"+formatStatCount(removed))

	diffWidth := m.getDiffWidth()
	rightWidth := 24
	if diffWidth < rightWidth {
		rightWidth = diffWidth
	}
	leftWidth := diffWidth - rightWidth
	if leftWidth < 0 {
		leftWidth = 0
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		workspaceHeaderStyle.Width(leftWidth).Render(left),
		workspaceHeaderStyle.Align(lipgloss.Right).Width(rightWidth).Render(right),
	)
}

func (m Model) renderDiffView() string {
	width := m.getDiffWidth()
	height := m.getContentHeight() - m.getWorkspaceHeaderHeight()
	if height < 0 {
		height = 0
	}
	return diffStyle.Width(width).Height(height).Render(m.diffViewport.View())
}

func (m Model) renderDiff() string {
	var lines []string

	// Calculate available width for content (excluding line numbers)
	lineNoWidth := 10 // "XXXX XXXX " format
	contentWidth := m.getDiffWidth() - lineNoWidth - 2
	if contentWidth < 20 {
		contentWidth = 20
	}

	for i, line := range m.diffLines {
		oldNo := ""
		newNo := ""

		if line.OldLineNo > 0 {
			oldNo = fmt.Sprintf("%4d", line.OldLineNo)
		} else {
			oldNo = "    "
		}

		if line.NewLineNo > 0 {
			newNo = fmt.Sprintf("%4d", line.NewLineNo)
		} else {
			newNo = "    "
		}

		lineNo := lineNoStyle.Render(fmt.Sprintf("%s %s ", oldNo, newNo))

		// Render content with syntax highlighting and word wrap
		var content string
		switch line.Type {
		case LineHunkHeader:
			content = hunkStyle.Render(line.Content)
			lines = append(lines, lineNo+content)
		default:
			// Extract prefix character (+/-/ ) and code content
			prefix := ""
			codeContent := line.Content
			if len(line.Content) > 0 {
				prefix = line.Content[:1]
				codeContent = line.Content[1:]
			}

			// Enable token-level syntax highlighting
			var highlightedCode string
			if m.currentFile != "" && m.highlighter != nil {
				tokens := m.highlighter.HighlightDiffLine(line.Content, m.currentFile)
				highlightedCode = m.renderTokens(tokens)
			} else {
				highlightedCode = codeContent
			}

			var bgStyle lipgloss.Style
			switch line.Type {
			case LineAdded:
				bgStyle = addedBgStyle
			case LineRemoved:
				bgStyle = removedBgStyle
			default:
				bgStyle = lipgloss.NewStyle()
			}
			isCurrentLine := i == m.diffCursor
			wrappedLines := wrapHighlightedLine(prefix, highlightedCode, contentWidth, bgStyle, isCurrentLine)
			for wIdx, wrapped := range wrappedLines {
				if wIdx == 0 {
					lines = append(lines, lineNo+wrapped)
				} else {
					// Continuation line with padding instead of line numbers
					continuationPrefix := lineNoStyle.Render("       │ ")
					lines = append(lines, continuationPrefix+wrapped)
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// renderTokens converts syntax-highlighted tokens to a colored string.
// It skips tokens[0] (the diff prefix character) and returns only the code content.
func (m Model) renderTokens(tokens []highlight.Token) string {
	var sb strings.Builder
	for i, tok := range tokens {
		if i == 0 {
			continue // skip the diff prefix (+/-/ )
		}
		color := m.highlighter.GetColor(tok.TokenType)
		if color != "" {
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(tok.Text))
		} else {
			sb.WriteString(tok.Text)
		}
	}
	return sb.String()
}

// findBreakPoint returns the byte index in s where visible character count reaches maxWidth.
// It handles ANSI escape sequences correctly by not counting them toward width.
func findBreakPoint(s string, maxWidth int) int {
	visibleWidth := 0
	inEscape := false
	lastSpaceIdx := -1
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
		if s[i] == ' ' {
			lastSpaceIdx = i
		}
		visibleWidth++
		if visibleWidth >= maxWidth {
			if lastSpaceIdx > 0 {
				return lastSpaceIdx
			}
			return i + 1
		}
	}
	return len(s)
}

// wrapHighlightedLine wraps syntax-highlighted content, applying bgStyle to the prefix character.
func wrapHighlightedLine(prefix, highlightedCode string, maxWidth int, bgStyle lipgloss.Style, isCurrentLine bool) []string {
	if maxWidth <= 0 {
		var lineStr string
		if prefix != "" {
			if isCurrentLine {
				lineStr = bgStyle.Render(prefix) + currentLineBg.Render(highlightedCode)
			} else {
				lineStr = bgStyle.Render(prefix) + highlightedCode
			}
		} else {
			if isCurrentLine {
				lineStr = currentLineBg.Render(highlightedCode)
			} else {
				lineStr = highlightedCode
			}
		}
		return []string{lineStr}
	}

	contentWidth := maxWidth
	if prefix != "" {
		contentWidth = maxWidth - 1
	}

	visibleLen := lipgloss.Width(highlightedCode)
	if visibleLen <= contentWidth {
		var lineStr string
		if prefix != "" {
			if isCurrentLine {
				lineStr = bgStyle.Render(prefix) + currentLineBg.Render(highlightedCode)
			} else {
				lineStr = bgStyle.Render(prefix) + highlightedCode
			}
		} else {
			if isCurrentLine {
				lineStr = currentLineBg.Render(highlightedCode)
			} else {
				lineStr = highlightedCode
			}
		}
		return []string{lineStr}
	}

	var result []string
	remaining := highlightedCode
	isFirstLine := true

	for lipgloss.Width(remaining) > contentWidth {
		breakPoint := findBreakPoint(remaining, contentWidth)
		line := remaining[:breakPoint]
		remaining = strings.TrimLeft(remaining[breakPoint:], " ")

		var lineStr string
		if isFirstLine && prefix != "" {
			if isCurrentLine {
				lineStr = bgStyle.Render(prefix) + currentLineBg.Render(line)
			} else {
				lineStr = bgStyle.Render(prefix) + line
			}
			isFirstLine = false
		} else {
			if isCurrentLine {
				lineStr = currentLineBg.Render(line)
			} else {
				lineStr = line
			}
		}
		result = append(result, lineStr)
	}

	if len(remaining) > 0 {
		var lineStr string
		if isFirstLine && prefix != "" {
			if isCurrentLine {
				lineStr = bgStyle.Render(prefix) + currentLineBg.Render(remaining)
			} else {
				lineStr = bgStyle.Render(prefix) + remaining
			}
		} else {
			if isCurrentLine {
				lineStr = currentLineBg.Render(remaining)
			} else {
				lineStr = remaining
			}
		}
		result = append(result, lineStr)
	}

	return result
}

func (m Model) renderHelp() string {
	helpText := helpTitleStyle.Render("Keyboard Shortcuts") + "\n\n"

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"j/k or ↑/↓", "Navigate files (list) / Scroll (diff)"},
		{"Enter", "View diff for selected file"},
		{"Ctrl+W", "Switch focus between panels"},
		{"L (shift+l)", "Toggle horizontal/vertical layout"},
		{"g / G", "Go to top/bottom of diff"},
		{"Ctrl+D / Ctrl+U", "Half page down/up"},
		{"Ctrl+F / Ctrl+B", "Page forward/backward (vim style)"},
		{"e", "Open file in $EDITOR"},
		{"r", "Refresh (reload files)"},
		{"?", "Toggle this help"},
		{"q", "Quit"},
	}

	var lines []string
	for _, s := range shortcuts {
		key := helpKeyStyle.Render(fmt.Sprintf("%-20s", s.key))
		lines = append(lines, key+" "+s.desc)
	}

	helpText += strings.Join(lines, "\n")
	helpText += "\n\n" + helpDimStyle.Render("Press any key to close...")

	return helpPopupStyle.Render(helpText)
}

func (m Model) renderHelpOverlay() string {
	// Render main view first
	mainView := m.renderShell()

	// Dim the main view
	dimmedView := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(mainView)

	// Place help on top
	helpContent := m.renderHelp()
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		helpContent,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("0")),
	) + dimmedView
}

func fitWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	s = truncateText(s, width)
	if pad := width - lipgloss.Width(s); pad > 0 {
		return s + strings.Repeat(" ", pad)
	}
	return s
}

func truncateText(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxWidth {
		return s
	}

	var b strings.Builder
	width := 0
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if width+rw > maxWidth {
			break
		}
		b.WriteRune(r)
		width += rw
	}
	return b.String()
}
