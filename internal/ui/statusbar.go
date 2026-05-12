// internal/ui/statusbar.go
package ui

import (
	"fmt"
	"strings"
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
