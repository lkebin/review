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

	// Tab → toggle focus to diff
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
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

func TestCursorPositionerSearchOpen(t *testing.T) {
	cp := &cursorPositioner{}
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.files = []FileInfo{{Status: "M", Name: "a.go"}}
	m.cursorPos = cp

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if cp.col.Load() == 0 {
		t.Error("opening search should set a non-zero cursor column")
	}
}

func TestCursorPositionerSearchClose(t *testing.T) {
	cp := &cursorPositioner{}
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.searchMode = true
	m.searchQuery = "foo"
	m.cursorPos = cp
	cp.setCol(5)

	// Enter should clear the column
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cp.col.Load() != 0 {
		t.Error("enter in search mode should clear cursor column")
	}

	// Reset, then Esc should also clear
	cp.setCol(5)
	m2 := NewModel(Options{Target: "HEAD"})
	m2.loading = false
	m2.searchMode = true
	m2.searchQuery = "foo"
	m2.cursorPos = cp
	m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cp.col.Load() != 0 {
		t.Error("esc in search mode should clear cursor column")
	}
}

func TestCursorPositionerTypingUpdatesCol(t *testing.T) {
	cp := &cursorPositioner{}
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.searchMode = true
	m.searchQuery = ""
	m.cursorPos = cp

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = result.(Model)
	colAfterA := cp.col.Load()
	if colAfterA == 0 {
		t.Error("typing in search mode should set cursor column")
	}

	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	m = result.(Model)
	colAfterAB := cp.col.Load()
	if colAfterAB <= colAfterA {
		t.Errorf("typing more chars should increase cursor column: got %d after 'a', %d after 'ab'", colAfterA, colAfterAB)
	}

	// Backspace should decrease column
	m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	colAfterBackspace := cp.col.Load()
	if colAfterBackspace >= colAfterAB {
		t.Errorf("backspace should decrease cursor column: got %d after 'ab', %d after backspace", colAfterAB, colAfterBackspace)
	}
}

func TestSearchSpaceInput(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.searchMode = true
	m.searchQuery = "hello"

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = result.(Model)
	if m.searchQuery != "hello " {
		t.Errorf("space key: searchQuery = %q, want %q", m.searchQuery, "hello ")
	}
}
