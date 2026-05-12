// internal/ui/app_test.go
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelDefaults(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	if m.focus != FocusList {
		t.Errorf("initial focus = %v, want FocusList", m.focus)
	}
	if m.loading != true {
		t.Error("initial loading should be true")
	}
}

func TestActionDispatchQuit(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = result
	if cmd == nil {
		t.Error("q key should produce quit command")
	}
}

func TestActionDispatchFocusSwitch(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.files = []FileInfo{{Status: "M", Name: "a.go"}}
	m.focus = FocusList

	// Ctrl+W then l → focus to diff
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	m = result.(Model)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = result.(Model)

	if m.focus != FocusDiff {
		t.Errorf("focus = %v, want FocusDiff", m.focus)
	}
}

func TestActionDispatchCursorInFileList(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.files = []FileInfo{
		{Status: "M", Name: "a.go"},
		{Status: "A", Name: "b.go"},
	}
	m.fileList.SetFiles(m.files)
	m.focus = FocusList

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(Model)

	if m.fileList.Cursor() != 1 {
		t.Errorf("fileList cursor = %d, want 1", m.fileList.Cursor())
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(Model)

	if m.width != 120 || m.height != 40 {
		t.Errorf("dimensions = %dx%d, want 120x40", m.width, m.height)
	}
}

func TestHelpToggle(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.files = []FileInfo{{Status: "M", Name: "a.go"}}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = result.(Model)
	if !m.showHelp {
		t.Error("? should toggle help on")
	}

	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = result.(Model)
	if m.showHelp {
		t.Error("? again should toggle help off")
	}
}
