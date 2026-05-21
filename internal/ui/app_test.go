// internal/ui/app_test.go
package ui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lkebin/review/internal/diff"
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

// TestStaleDiffResponseIgnored verifies that a loadDiffMsg for a file that is
// no longer selected does not overwrite the current diff view content.
func TestStaleDiffResponseIgnored(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.files = []FileInfo{
		{Status: "M", Name: "file1.go"},
		{Status: "A", Name: "file2.go"},
	}
	m.fileList.SetFiles(m.files)

	// Navigate to file2 (cursor = 1)
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(Model)

	// A stale diff for file1 arrives (was in-flight from initial load)
	staleLines := []diff.Line{
		{Type: diff.LineAdded, NewLineNo: 1, Content: "+stale file1 content"},
	}
	result, _ = m.Update(loadDiffMsg{file: "file1.go", lines: staleLines})
	m = result.(Model)

	// currentFile must not be set to the stale file
	if m.currentFile == "file1.go" {
		t.Error("stale diff response overwrote currentFile; should have been ignored")
	}
	// Diff view must not contain stale content
	if m.diffView.Viewport().LineCount() > 0 {
		line := m.diffView.Viewport().Lines()[0]
		if line.RawContent == "stale file1 content" {
			t.Error("stale diff response populated the diff view; should have been ignored")
		}
	}
}

// TestDiffErrorCleared verifies that a successful diff load for the current
// file clears a previously-set error so the UI recovers.
func TestDiffErrorCleared(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.files = []FileInfo{{Status: "M", Name: "file1.go"}}
	m.fileList.SetFiles(m.files)

	// Inject a diff error for the current file
	result, _ := m.Update(loadDiffMsg{file: "file1.go", err: fmt.Errorf("git error")})
	m = result.(Model)
	if m.err == nil {
		t.Fatal("expected error to be set after error message")
	}

	// A successful diff for the same file arrives (e.g. after a retry or the next nav)
	goodLines := []diff.Line{
		{Type: diff.LineContext, OldLineNo: 1, NewLineNo: 1, Content: " ok"},
	}
	result, _ = m.Update(loadDiffMsg{file: "file1.go", lines: goodLines})
	m = result.(Model)
	if m.err != nil {
		t.Errorf("error should be cleared after successful diff load, got: %v", m.err)
	}
}
