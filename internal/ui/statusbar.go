// internal/ui/statusbar.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderStatusBar renders a two-part status bar:
// Left: branch + file count | Right: current file + stats
func RenderStatusBar(branch string, fileCount int, currentFile string, added, removed, width int, theme Theme) string {
	if width <= 0 {
		return ""
	}

	style := theme.StatusBarStyle()

	left := fmt.Sprintf(" %s | %d files", branch, fileCount)
	right := ""
	if currentFile != "" {
		right = fmt.Sprintf(" %s +%d -%d ", currentFile, added, removed)
	}

	leftW := len(left)
	rightW := len(right)
	fillW := width - leftW - rightW
	if fillW < 0 {
		// Truncate right side if too narrow
		available := width - leftW
		if available > 0 && len(right) > available {
			right = right[:available]
			fillW = 0
		} else {
			right = ""
			fillW = width - leftW
			if fillW < 0 {
				left = left[:width]
				fillW = 0
			}
		}
	}

	fill := strings.Repeat(" ", fillW)
	return style.Width(width).Render(left + fill + right)
}

// RenderSearchBar renders the search status bar.
// When typing is true the input cursor ▌ is shown; when false the query is
// displayed as a read-only indicator (confirmed search still active).
// Left: /query[▌]   Right: [files] or [diff]
func RenderSearchBar(query string, focus FocusType, width int, theme Theme, typing bool) string {
	if width <= 0 {
		return ""
	}
	panel := "[diff]"
	if focus == FocusList {
		panel = "[files]"
	}
	cursor := ""
	if typing {
		cursor = "▌"
	}
	prompt := lipgloss.NewStyle().Bold(true).Render("/") + query + cursor
	right := " " + panel + " "
	gap := width - lipgloss.Width(prompt) - len(right)
	if gap < 0 {
		gap = 0
	}
	bar := prompt + strings.Repeat(" ", gap) + right
	return theme.StatusBarStyle().Width(width).Render(bar)
}
