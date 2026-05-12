package ui

import (
	"errors"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kbliu/review/internal/highlight"
)

// newTestModel returns a zero-state Model suitable for unit testing Update handlers.
// It does not load files or start any commands.
func newTestModel() Model {
	m := NewModel(Options{Target: "HEAD"})
	m.highlighter = highlight.New("github")
	return m
}

func TestLoadDiffMsgSetsCurrentFile(t *testing.T) {
	m := newTestModel()
	m.files = []FileInfo{{Name: "foo/bar.go", Status: "M"}}

	msg := loadDiffMsg{
		file:  "foo/bar.go",
		lines: []DiffLine{{Type: LineContext, Content: " hello"}},
	}

	result, _ := m.Update(msg)
	updated := result.(Model)

	if updated.currentFile != "foo/bar.go" {
		t.Errorf("expected currentFile=%q, got %q", "foo/bar.go", updated.currentFile)
	}
	// diffLines is set synchronously in loadDiffMsg handler; viewport rendering is deferred
	if len(updated.diffLines) != 1 {
		t.Errorf("expected 1 diffLine, got %d", len(updated.diffLines))
	}
}

func TestLoadDiffMsgSetsCurrentFileOnError(t *testing.T) {
	m := newTestModel()
	m.currentFile = "old/file.go" // simulate a previously loaded file

	msg := loadDiffMsg{
		file: "foo/bar.go",
		err:  errors.New("git error"),
	}

	result, _ := m.Update(msg)
	updated := result.(Model)

	// currentFile must be updated even on error — this is the core bug fix
	if updated.currentFile != "foo/bar.go" {
		t.Errorf("expected currentFile=%q even on error, got %q", "foo/bar.go", updated.currentFile)
	}
	if len(updated.diffLines) != 0 {
		t.Errorf("expected no diffLines on error, got %d", len(updated.diffLines))
	}
}

func TestDiffCursorScrollSync(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.listWidth = 32
	// Set up enough diff lines to require scrolling
	m.diffLines = make([]DiffLine, 30)
	for i := range m.diffLines {
		m.diffLines[i] = DiffLine{Type: LineContext, Content: fmt.Sprintf(" line %d", i)}
	}
	m.diffViewport.Height = 10
	m.diffCursor = 0

	jKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	kKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}

	// Pressing j 15 times should scroll viewport to keep cursor visible
	for i := 0; i < 15; i++ {
		result, _ := m.handleDiffKeys(jKey)
		m = result.(Model)
	}
	if m.diffCursor != 15 {
		t.Errorf("diffCursor = %d, want 15", m.diffCursor)
	}
	wantOffset := m.diffCursor - m.diffViewport.Height + 1
	if m.diffViewport.YOffset != wantOffset {
		t.Errorf("viewport not scrolled: YOffset=%d, want %d (cursor=%d, height=%d)",
			m.diffViewport.YOffset, wantOffset, m.diffCursor, m.diffViewport.Height)
	}

	// Pressing k 5 times should scroll viewport back
	for i := 0; i < 5; i++ {
		result, _ := m.handleDiffKeys(kKey)
		m = result.(Model)
	}
	if m.diffCursor != 10 {
		t.Errorf("diffCursor = %d, want 10", m.diffCursor)
	}
	if m.diffViewport.YOffset > m.diffCursor {
		t.Errorf("viewport not scrolled back: YOffset=%d, cursor=%d",
			m.diffViewport.YOffset, m.diffCursor)
	}
}

func TestMouseWheelListPanel(t *testing.T) {
m := newTestModel()
m.width = 120
m.listWidth = 32
m.layout = LayoutHorizontal
m.files = []FileInfo{{Name: "a.go"}, {Name: "b.go"}, {Name: "c.go"}}
m.cursor = 1

// WheelUp in list panel (X=10 < listWidth=32) moves cursor up
result, _ := m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp, X: 10})
m = result.(Model)
if m.cursor != 0 {
t.Errorf("WheelUp: cursor = %d, want 0", m.cursor)
}

// WheelUp at top does not go negative
result, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp, X: 10})
m = result.(Model)
if m.cursor != 0 {
t.Errorf("WheelUp at top: cursor = %d, want 0", m.cursor)
}

// WheelDown in list panel moves cursor down
result, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, X: 10})
m = result.(Model)
if m.cursor != 1 {
t.Errorf("WheelDown: cursor = %d, want 1", m.cursor)
}
}

func TestMouseWheelDiffPanel(t *testing.T) {
m := newTestModel()
m.width = 120
m.listWidth = 32
m.layout = LayoutHorizontal
m.diffLines = make([]DiffLine, 20)
for i := range m.diffLines {
m.diffLines[i] = DiffLine{Type: LineContext, Content: fmt.Sprintf(" line %d", i)}
}
m.diffViewport.Height = 10
m.diffCursor = 0

// WheelDown in diff panel (X=50 >= listWidth=32) moves diffCursor down
result, _ := m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, X: 50})
m = result.(Model)
if m.diffCursor != 1 {
t.Errorf("WheelDown: diffCursor = %d, want 1", m.diffCursor)
}

// Scroll past viewport height triggers YOffset update
for i := 0; i < 10; i++ {
result, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, X: 50})
m = result.(Model)
}
if m.diffCursor != 11 {
t.Errorf("after 11 downs: diffCursor = %d, want 11", m.diffCursor)
}
wantOffset := m.diffCursor - m.diffViewport.Height + 1
if m.diffViewport.YOffset != wantOffset {
t.Errorf("YOffset = %d, want %d", m.diffViewport.YOffset, wantOffset)
}

// WheelUp scrolls back
result, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp, X: 50})
m = result.(Model)
if m.diffCursor != 10 {
t.Errorf("WheelUp: diffCursor = %d, want 10", m.diffCursor)
}
}
