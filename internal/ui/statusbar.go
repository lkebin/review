// internal/ui/statusbar.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
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

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	fillW := width - leftW - rightW
	if fillW < 0 {
		// Truncate right side if too narrow
		available := width - leftW
		if available > 0 && lipgloss.Width(right) > available {
			right = truncateRight(right, available)
			fillW = 0
		} else {
			right = ""
			fillW = width - leftW
			if fillW < 0 {
				left = truncateRight(left, width)
				fillW = 0
			}
		}
	}

	fill := strings.Repeat(" ", fillW)
	return style.Width(width).Render(left + fill + right)
}

// truncateRight truncates s to at most maxW visible columns from the left.
func truncateRight(s string, maxW int) string {
	if lipgloss.Width(s) <= maxW {
		return s
	}
	acc := 0
	for i, r := range s {
		w := runewidth.RuneWidth(r)
		if acc+w > maxW {
			return s[:i]
		}
		acc += w
	}
	return s
}

// RenderCmdLine renders the vim-style command line (bottom row).
// Shows "/query" while a search is active; otherwise blank.
func RenderCmdLine(query string, width int, typing bool) string {
	if width <= 0 {
		return ""
	}
	if !typing && query == "" {
		return lipgloss.NewStyle().Width(width).Render("")
	}
	prompt := lipgloss.NewStyle().Bold(true).Render("/") + query
	return lipgloss.NewStyle().Width(width).Render(prompt)
}
