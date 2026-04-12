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
)

// View renders the UI
func (m Model) View() string {
	if m.loading {
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

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
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

	status := fmt.Sprintf("%s > %s | Files: %d | %d/%d | Layout: %s | Focus: %s | </> resize | [l]ayout | [q]uit",
		"current", target, len(m.files), m.cursor+1, len(m.files), layout, focus)

	return statusBarStyle.Width(m.width).Render(status)
}

func (m Model) renderDiff() string {
	var lines []string

	// Get current filename for syntax highlighting
	filename := ""
	if m.cursor < len(m.files) {
		filename = m.files[m.cursor].Name
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

		// Render content with syntax highlighting
		var content string
		switch line.Type {
		case LineHunkHeader:
			content = hunkStyle.Render(line.Content)
		default:
			// Use syntax highlighting for code lines
			tokens := m.highlighter.HighlightDiffLine(line.Content, filename)
			var highlighted strings.Builder
			for _, token := range tokens {
				if token.Color != "" {
					highlighted.WriteString(token.Color)
					highlighted.WriteString(token.Text)
					highlighted.WriteString("\x1b[0m") // Reset
				} else {
					// Apply diff type color
					text := token.Text
					switch line.Type {
					case LineAdded:
						text = addedStyle.Render(text)
					case LineRemoved:
						text = removedStyle.Render(text)
					}
					highlighted.WriteString(text)
				}
			}
			content = highlighted.String()
		}

		lines = append(lines, lineNo+content)
	}

	return strings.Join(lines, "\n")
}
