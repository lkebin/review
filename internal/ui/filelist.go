// internal/ui/filelist.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// FileList manages the left panel file list.
type FileList struct {
	files  []FileInfo
	cursor int
	offset int
}

// NewFileList creates a file list from the given files.
func NewFileList(files []FileInfo) *FileList {
	return &FileList{files: files}
}

// Cursor returns the current cursor position.
func (fl *FileList) Cursor() int { return fl.cursor }

// SetFiles replaces the file list and resets cursor.
func (fl *FileList) SetFiles(files []FileInfo) {
	fl.files = files
	fl.cursor = 0
	fl.offset = 0
}

// CursorDown moves cursor down.
func (fl *FileList) CursorDown() {
	if fl.cursor < len(fl.files)-1 {
		fl.cursor++
	}
}

// CursorUp moves cursor up.
func (fl *FileList) CursorUp() {
	if fl.cursor > 0 {
		fl.cursor--
	}
}

// PageDown moves cursor down by height items.
func (fl *FileList) PageDown(height int) {
	if len(fl.files) == 0 {
		return
	}
	fl.cursor += height
	if fl.cursor >= len(fl.files) {
		fl.cursor = len(fl.files) - 1
	}
}

// PageUp moves cursor up by height items.
func (fl *FileList) PageUp(height int) {
	fl.cursor -= height
	if fl.cursor < 0 {
		fl.cursor = 0
	}
}

// GotoTop moves cursor to first item.
func (fl *FileList) GotoTop() {
	fl.cursor = 0
	fl.offset = 0
}

// GotoBottom moves cursor to last item.
func (fl *FileList) GotoBottom() {
	if len(fl.files) > 0 {
		fl.cursor = len(fl.files) - 1
	}
}

func (fl *FileList) ensureVisible(height int) {
	if fl.cursor < fl.offset {
		fl.offset = fl.cursor
	}
	if height > 0 && fl.cursor >= fl.offset+height {
		fl.offset = fl.cursor - height + 1
	}
	if fl.offset < 0 {
		fl.offset = 0
	}
}

// SearchNext moves cursor to the next file whose name contains query (case-insensitive, wraps around).
// Returns true if the cursor moved.
func (fl *FileList) SearchNext(query string) bool {
	if query == "" || len(fl.files) == 0 {
		return false
	}
	q := strings.ToLower(query)
	for i := 1; i <= len(fl.files); i++ {
		idx := (fl.cursor + i) % len(fl.files)
		if strings.Contains(strings.ToLower(fl.files[idx].Name), q) {
			fl.cursor = idx
			return true
		}
	}
	return false
}

// SearchPrev moves cursor to the previous file whose name contains query (case-insensitive, wraps around).
// Returns true if the cursor moved.
func (fl *FileList) SearchPrev(query string) bool {
	if query == "" || len(fl.files) == 0 {
		return false
	}
	q := strings.ToLower(query)
	n := len(fl.files)
	for i := 1; i <= n; i++ {
		idx := (fl.cursor - i + n) % n
		if strings.Contains(strings.ToLower(fl.files[idx].Name), q) {
			fl.cursor = idx
			return true
		}
	}
	return false
}

// SelectedFile returns the currently selected file.
func (fl *FileList) SelectedFile() FileInfo {
	if fl.cursor < len(fl.files) {
		return fl.files[fl.cursor]
	}
	return FileInfo{}
}

// CalcWidth calculates the default width based on file names. Capped at 50.
func (fl *FileList) CalcWidth() int {
	w := 20 // minimum
	for _, f := range fl.files {
		// badge(1) + space(1) + filename visual width + padding(2)
		lineW := 1 + 1 + runewidth.StringWidth(f.Name) + 2
		if lineW > w {
			w = lineW
		}
	}
	if w > 50 {
		w = 50
	}
	return w
}

// Render draws the file list to fit the given width and height.
// The rightmost character of every row is always the panel separator │,
// so the dividing line is visible regardless of diff content.
func (fl *FileList) Render(width, height int, theme Theme) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	fl.ensureVisible(height)

	// contentWidth is the area available before the separator.
	contentWidth := width - 1
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SepFg)).Render("│")

	var rows []string
	end := fl.offset + height
	if end > len(fl.files) {
		end = len(fl.files)
	}
	for i := fl.offset; i < end; i++ {
		f := fl.files[i]

		var content string
		if i == fl.cursor {
			content = theme.FileSelectedStyle().Width(contentWidth).Render(
				fmt.Sprintf(" %s %s", f.Status, truncateName(f.Name, contentWidth-3)))
		} else {
			statusColor := theme.StatusColor(f.Status)
			badge := lipgloss.NewStyle().
				Foreground(lipgloss.Color(statusColor)).
				Bold(true).
				Render(f.Status)
			name := truncateName(f.Name, contentWidth-3)
			raw := fmt.Sprintf(" %s %s", badge, name)
			// Pad to contentWidth so the separator always sits at a fixed column.
			if pad := contentWidth - lipgloss.Width(raw); pad > 0 {
				raw += strings.Repeat(" ", pad)
			}
			content = raw
		}

		rows = append(rows, content+sep)
	}

	// Pad remaining height — separator must still appear on empty rows.
	for len(rows) < height {
		rows = append(rows, strings.Repeat(" ", contentWidth)+sep)
	}

	return strings.Join(rows, "\n")
}

// truncateName truncates a path to at most maxWidth visible columns, keeping
// the rightmost portion so the filename is always visible: "…rc/app/alert.c".
// Uses visual column widths so wide (CJK) characters are counted as 2 columns.
func truncateName(name string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if runewidth.StringWidth(name) <= maxWidth {
		return name
	}
	// Walk from the right accumulating visual width until we fill maxWidth-1
	// columns (1 column is reserved for the leading "…").
	target := maxWidth - 1
	runes := []rune(name)
	acc := 0
	start := len(runes)
	for i := len(runes) - 1; i >= 0; i-- {
		w := runewidth.RuneWidth(runes[i])
		if acc+w > target {
			break
		}
		acc += w
		start = i
	}
	return "…" + string(runes[start:])
}
