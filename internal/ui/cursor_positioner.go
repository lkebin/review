package ui

import (
	"fmt"
	"os"
	"regexp"
	"sync/atomic"
)

// cursorPositioner wraps os.Stdout, implements term.File (so bubbletea treats
// it as a real TTY), and steers the terminal cursor to the correct column after
// every frame bubbletea renders.
//
// In altscreen mode bubbletea's renderer unconditionally emits \033[{row};0H
// at the end of each frame, parking the cursor at column 1. When a target
// column is set (search mode) we detect that sequence and append
// \033[{col}G (CHA – Cursor Horizontal Absolute) so the real terminal cursor
// lands at the end of the search query instead.
type cursorPositioner struct {
	inner *os.File
	col   atomic.Int32 // 0 = pass-through; >0 = 1-indexed target column
}

func (w *cursorPositioner) setCol(col int32) {
	if w != nil {
		w.col.Store(col)
	}
}
func (w *cursorPositioner) clearCol() {
	if w != nil {
		w.col.Store(0)
	}
}

// Fd, Read, Close delegate to the underlying file so bubbletea's TTY detection
// (term.IsTerminal, term.GetSize) works correctly.
func (w *cursorPositioner) Fd() uintptr        { return w.inner.Fd() }
func (w *cursorPositioner) Read(p []byte) (int, error) { return w.inner.Read(p) }
func (w *cursorPositioner) Close() error               { return w.inner.Close() }

// bubbletea altscreen flush ends every frame with \033[{row};H or \033[H
// (charmbracelet/x/ansi omits the column param when col<=0, so it's ";" not ";0").
var endCursorRe = regexp.MustCompile(`\x1b\[(\d+;)?H$`)

func (w *cursorPositioner) Write(b []byte) (int, error) {
	col := w.col.Load()
	if col <= 0 || !endCursorRe.Match(b) {
		return w.inner.Write(b)
	}
	// Append CHA to override the column-0 park: \033[{col}G
	suffix := fmt.Sprintf("\x1b[%dG", col)
	combined := make([]byte, len(b)+len(suffix))
	copy(combined, b)
	copy(combined[len(b):], suffix)
	n, err := w.inner.Write(combined)
	if err != nil {
		return min(n, len(b)), err
	}
	return len(b), nil
}
