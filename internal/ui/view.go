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
		Background(lipgloss.Color("235")).
		Padding(1, 2)

	helpTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("208"))

	helpKeyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))
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
		return m.overlayHelp(mainView, helpView)
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
			var highlighted []string
			for _, token := range tokens {
				text := token.Text
				color := m.highlighter.GetColor(token.TokenType)

				// Apply syntax color or diff type color
				if color != "" {
					style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
					// Also apply diff background for +/- lines
					switch line.Type {
					case LineAdded:
						style = style.Foreground(lipgloss.Color("82")) // Green for additions
					case LineRemoved:
						style = style.Foreground(lipgloss.Color("196")) // Red for deletions
					}
					highlighted = append(highlighted, style.Render(text))
				} else {
					// Just apply diff type color
					switch line.Type {
					case LineAdded:
						text = addedStyle.Render(text)
					case LineRemoved:
						text = removedStyle.Render(text)
					}
					highlighted = append(highlighted, text)
				}
			}
			content = strings.Join(highlighted, "")
		}

		lines = append(lines, lineNo+content)
	}

	return strings.Join(lines, "\n")
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
		{"Ctrl+_ / Ctrl+=", "Resize file list width"},
		{"g / G", "Go to top/bottom of diff"},
		{"Ctrl+D / Ctrl+U", "Half page down/up"},
		{"?", "Toggle this help"},
		{"q", "Quit"},
	}

	var lines []string
	for _, s := range shortcuts {
		key := helpKeyStyle.Render(fmt.Sprintf("%-18s", s.key))
		lines = append(lines, key+" "+s.desc)
	}

	helpText += strings.Join(lines, "\n")
	helpText += "\n\nPress any key to close..."

	return helpPopupStyle.Render(helpText)
}

func (m Model) overlayHelp(mainView, helpView string) string {
	// Center the help popup
	helpWidth := lipgloss.Width(helpView)
	helpHeight := lipgloss.Height(helpView)

	// Calculate position to center
	x := (m.width - helpWidth) / 2
	if x < 0 {
		x = 0
	}
	y := (m.height - helpHeight) / 2
	if y < 0 {
		y = 0
	}

	// Split main view into lines
	mainLines := strings.Split(mainView, "\n")

	// Create a new view with help overlaid
	var result []string
	for i, line := range mainLines {
		if i >= y && i < y+helpHeight {
			// This line should have help content
			helpLine := getLine(helpView, i-y)
			if x < len(line) {
				// Replace portion of line with help
				before := ""
				if x > 0 {
					before = line[:x]
				}
				after := ""
				if x+len(helpLine) < len(line) {
					after = line[x+len(helpLine):]
				}
				result = append(result, before+helpLine+after)
			} else {
				result = append(result, line)
			}
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func getLine(s string, n int) string {
	lines := strings.Split(s, "\n")
	if n < len(lines) {
		return lines[n]
	}
	return ""
}
