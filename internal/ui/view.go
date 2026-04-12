package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	listStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color("240"))

	diffStyle = lipgloss.NewStyle()

	statusBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("250")).
		Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("238")).
		Bold(true)

	addedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")) // Green

	removedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")) // Red

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
)

// View renders the UI
func (m Model) View() string {
	if m.loading && len(m.files) == 0 {
		return "Loading..."
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	if len(m.files) == 0 {
		return "No changes found.\n\nPress q to quit."
	}

	var content string

	if m.layout == LayoutHorizontal {
		listView := m.renderFileList()
		diffView := m.renderDiffView()
		content = lipgloss.JoinHorizontal(lipgloss.Top, listView, diffView)
	} else {
		listView := m.renderFileList()
		diffView := m.renderDiffView()
		content = lipgloss.JoinVertical(lipgloss.Left, listView, diffView)
	}

	statusBar := m.renderStatusBar()
	mainView := lipgloss.JoinVertical(lipgloss.Left, content, statusBar)

	// Overlay help if shown
	if m.showHelp {
		helpView := m.renderHelp()
		return m.overlayCenter(mainView, helpView)
	}

	return mainView
}

func (m Model) renderFileList() string {
	var lines []string

	for i, f := range m.files {
		line := fmt.Sprintf("%s %s", f.Status, f.Name)
		if i == m.cursor {
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	width := m.getListWidth()
	height := m.getContentHeight()

	return listStyle.Width(width).Height(height).Render(content)
}

func (m Model) renderDiffView() string {
	width := m.getDiffWidth()
	height := m.getContentHeight()
	return diffStyle.Width(width).Height(height).Render(m.diffViewport.View())
}

func (m Model) renderStatusBar() string {
	target := m.options.Target
	if target == "" {
		target = "HEAD"
	}

	layout := "H"
	if m.layout == LayoutVertical {
		layout = "V"
	}

	focus := "List"
	if m.focus == FocusDiff {
		focus = "Diff"
	}

	status := fmt.Sprintf("%s > %s | Files: %d | %d/%d | Layout: %s | Focus: %s | [?]help | [q]uit",
		"current", target, len(m.files), m.cursor+1, len(m.files), layout, focus)

	return statusBarStyle.Width(m.width).Render(status)
}

func (m Model) renderDiff() string {
	var lines []string

	// Calculate available width for content (excluding line numbers)
	lineNoWidth := 10 // "XXXX XXXX " format
	contentWidth := m.getDiffWidth() - lineNoWidth - 2
	if contentWidth < 20 {
		contentWidth = 20
	}

	for _, line := range m.diffLines {
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
			// Simple diff highlighting (no syntax highlighting for performance)
			content = line.Content
			switch line.Type {
			case LineAdded:
				content = addedStyle.Render(content)
			case LineRemoved:
				content = removedStyle.Render(content)
			}

			// Handle word wrap for long lines
			wrappedLines := wrapLine(content, contentWidth)
			for i, wrapped := range wrappedLines {
				if i == 0 {
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

// wrapLine wraps a line to fit within maxWidth
func wrapLine(line string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{line}
	}

	var result []string
	remaining := line

	for len(remaining) > maxWidth {
		// Find a good break point
		breakPoint := maxWidth
		for breakPoint > 0 && remaining[breakPoint] != ' ' {
			breakPoint--
		}
		if breakPoint == 0 {
			// No space found, hard break
			breakPoint = maxWidth
		}

		result = append(result, remaining[:breakPoint])
		remaining = remaining[breakPoint:]
		// Trim leading space from remaining
		remaining = strings.TrimLeft(remaining, " ")
	}

	if len(remaining) > 0 {
		result = append(result, remaining)
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
		{"Ctrl+_ / Ctrl+=", "Resize focused panel (shrink/grow)"},
		{"L (shift+l)", "Toggle horizontal/vertical layout"},
		{"g / G", "Go to top/bottom of diff"},
		{"Ctrl+D / Ctrl+U", "Half page down/up"},
		{"Ctrl+F / Ctrl+B", "Page forward/backward (vim style)"},
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

func (m Model) overlayCenter(background, foreground string) string {
	// Simple centered overlay using lipgloss.Place
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		foreground,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("0")),
	)
}

func getLine(s string, n int) string {
	lines := strings.Split(s, "\n")
	if n < len(lines) {
		return lines[n]
	}
	return ""
}
