# Go TUI UI Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix layout corruption, enable syntax highlighting, redesign status bar, add mouse scrolling, and add a proper vim-style diff cursor in the Go `review` TUI.

**Architecture:** All changes are within `internal/ui/`. The width model is simplified (no cap in `getListWidth()`), `loadDiffMsg` carries the filename so `m.currentFile` is always set, and Chroma highlighting is wired up in `renderDiff()`. Mouse support uses `tea.WithMouseCellMotion()` and a new `tea.MouseMsg` branch in `Update()`.

**Tech Stack:** Go 1.24, `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/charmbracelet/bubbles`, `github.com/alecthomas/chroma/v2`

---

## File Map

| File | Role / Changes |
|------|---------------|
| `internal/ui/ui.go` | Add `tea.WithMouseCellMotion()` to program options |
| `internal/ui/model.go` | Initial `listWidth = 32`; remove `WidthInc`/`WidthDec` from keyMap/defaultKeys; add `addedBgStyle`/`removedBgStyle` styles |
| `internal/ui/update.go` | Add `file string` to `loadDiffMsg`; set `m.currentFile` in handler; remove `-/=` key handlers; fix cursor scroll-sync in `handleDiffKeys`; add `tea.MouseMsg` handler |
| `internal/ui/view.go` | Fix `getListWidth()`/`getDiffWidth()`; fix `renderFileList()` to apply `listStyle`; enable syntax highlighting in `renderDiff()`; redesign `renderStatusBar()`; update `wrapHighlightedLine()` to use bg style |
| `internal/ui/model_test.go` | New file — unit tests for model update logic |
| `internal/ui/view_test.go` | New file — unit tests for rendering helpers |

---

## Task 1: Fix `loadDiffMsg` to carry filename

**Files:**
- Modify: `internal/ui/update.go`
- Create: `internal/ui/model_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/ui/model_test.go`:

```go
package ui

import (
	"testing"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel() Model {
	return NewModel(Options{Target: "HEAD"})
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
	if len(updated.diffLines) != 1 {
		t.Errorf("expected 1 diffLine, got %d", len(updated.diffLines))
	}
}
```

- [ ] **Step 2: Run test — expect compile error (loadDiffMsg has no `file` field)**

```bash
cd internal/ui && go test ./... -run TestLoadDiffMsgSetsCurrentFile 2>&1 | head -20
```

Expected: compile error `unknown field 'file' in struct literal`

- [ ] **Step 3: Add `file string` to `loadDiffMsg` and wire it up**

In `internal/ui/update.go`, change:

```go
// loadDiffMsg is sent when diff is loaded
type loadDiffMsg struct {
	lines []DiffLine
	file  string
	err   error
}
```

Change `loadDiff` function to set the field:

```go
func loadDiff(opts Options, file string) tea.Cmd {
	return func() tea.Msg {
		lines, err := getDiff(opts, file)
		return loadDiffMsg{lines: lines, file: file, err: err}
	}
}
```

In the `Update()` `loadDiffMsg` handler, add `m.currentFile = msg.file`:

```go
case loadDiffMsg:
	m.loading = false
	m.currentFile = msg.file
	if msg.err != nil {
		m.err = msg.err
		return m, nil
	}
	m.diffLines = msg.lines
	m.diffCursor = 0
	m.diffViewport.SetContent(m.renderDiff())
	return m, nil
```

- [ ] **Step 4: Run test — expect PASS**

```bash
cd internal/ui && go test ./... -run TestLoadDiffMsgSetsCurrentFile -v
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/ui/update.go internal/ui/model_test.go
git commit -m "fix: loadDiffMsg carries filename, sets m.currentFile

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

## Task 2: Fix width model and layout

**Files:**
- Modify: `internal/ui/view.go` (`getListWidth`, `getDiffWidth`, `renderFileList`)
- Modify: `internal/ui/model.go` (initial `listWidth`)
- Modify: `internal/ui/update.go` (`WindowSizeMsg` handler, remove stale `-2` discrepancy)
- Create: `internal/ui/view_test.go`

- [ ] **Step 1: Write failing test for width math**

Create `internal/ui/view_test.go`:

```go
package ui

import "testing"

func TestWidthMath(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.listWidth = 32

	if got := m.getListWidth(); got != 32 {
		t.Errorf("getListWidth: expected 32, got %d", got)
	}
	if got := m.getDiffWidth(); got != 88 { // 120 - 32
		t.Errorf("getDiffWidth: expected 88, got %d", got)
	}
}
```

- [ ] **Step 2: Run test — expect FAIL (current getListWidth caps at width/2)**

```bash
cd internal/ui && go test ./... -run TestWidthMath -v
```

Expected: FAIL — `getListWidth: expected 32, got 32` (might pass coincidentally for this input; the cap issue manifests when listWidth > width/2; the getDiffWidth test will fail because old formula is `width - getListWidth - 1 = 87`, not `88`)

- [ ] **Step 3: Fix `getListWidth()` and `getDiffWidth()` in `view.go`**

In `internal/ui/view.go`, replace the two helper functions:

```go
func (m Model) getListWidth() int {
	return m.listWidth
}

func (m Model) getDiffWidth() int {
	return m.width - m.listWidth
}
```

- [ ] **Step 4: Fix initial `listWidth` in `model.go`**

In `internal/ui/model.go`, in `NewModel()`, change:

```go
return Model{
	options:      opts,
	focus:        FocusList,
	loading:      true,
	listWidth:    32,        // was 30
	// ... rest unchanged
```

- [ ] **Step 5: Fix `renderFileList()` in `view.go` to apply `listStyle`**

Replace the existing `renderFileList()`:

```go
func (m Model) renderFileList() string {
	return listStyle.Width(m.listWidth - 1).Render(m.fileList.View())
}
```

Note: `listStyle` has a right border (1 char). `Width(m.listWidth - 1)` sets inner content width. Lipgloss adds the 1-char border, giving total rendered width of `m.listWidth`.

- [ ] **Step 6: Fix `fileList` SetWidth calls to use `m.listWidth - 2`**

In `internal/ui/update.go`, in the `WindowSizeMsg` handler, ensure:

```go
case tea.WindowSizeMsg:
	m.width = msg.Width
	m.height = msg.Height
	m.diffViewport.Width = m.getDiffWidth()
	m.diffViewport.Height = m.getDiffHeight()
	m.fileList.SetWidth(m.listWidth - 2) // -1 for border, -1 for list internal padding
	m.fileList.SetHeight(m.getListHeight())
	if len(m.diffLines) > 0 {
		m.diffViewport.SetContent(m.renderDiff())
	}
	return m, nil
```

Also fix the same `SetWidth` calls inside the `-`/`=` key handlers — those will be **removed** in Task 4, so skip for now.

- [ ] **Step 7: Run tests**

```bash
cd internal/ui && go test ./... -v
```

Expected: all pass. Also run the app manually to verify the separator line appears and the layout fills the terminal correctly:

```bash
cd /path/to/repo && go build -o review ./cmd/review && ./review
```

- [ ] **Step 8: Commit**

```bash
git add internal/ui/view.go internal/ui/model.go internal/ui/update.go internal/ui/view_test.go
git commit -m "fix: correct width model, apply listStyle border in renderFileList

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

## Task 3: Enable syntax highlighting with background colors

**Files:**
- Modify: `internal/ui/view.go` (`renderDiff`, `wrapHighlightedLine`, styles)

- [ ] **Step 1: Add `addedBgStyle` and `removedBgStyle` to styles in `view.go`**

In `internal/ui/view.go`, add to the `var ( ... )` styles block:

```go
addedBgStyle = lipgloss.NewStyle().
    Background(lipgloss.Color("22")) // dark green bg

removedBgStyle = lipgloss.NewStyle().
    Background(lipgloss.Color("52")) // dark red bg
```

- [ ] **Step 2: Update `wrapHighlightedLine` signature to use `bgStyle`**

The function currently accepts `baseStyle lipgloss.Style` and uses it as a foreground color style on the prefix (`+`/`-`). Change it to use `bgStyle` for background (applied to prefix and continuation lines):

Replace the function signature and update the prefix rendering:

```go
func wrapHighlightedLine(prefix, highlightedCode string, maxWidth int, bgStyle lipgloss.Style, isCurrentLine bool) []string {
	if maxWidth <= 0 {
		var result string
		if prefix != "" {
			result = bgStyle.Render(prefix+highlightedCode)
		} else {
			result = bgStyle.Render(highlightedCode)
		}
		if isCurrentLine {
			result = currentLineBg.Render(result)
		}
		return []string{result}
	}

	contentWidth := maxWidth
	if prefix != "" {
		contentWidth = maxWidth - 1
	}

	visibleLen := lipgloss.Width(highlightedCode)
	if visibleLen <= contentWidth {
		var result string
		if prefix != "" {
			result = bgStyle.Render(prefix) + highlightedCode
		} else {
			result = highlightedCode
		}
		if isCurrentLine {
			result = currentLineBg.Render(result)
		}
		return []string{result}
	}

	var result []string
	remaining := highlightedCode
	isFirstLine := true

	for lipgloss.Width(remaining) > contentWidth {
		breakPoint := findBreakPoint(remaining, contentWidth)
		line := remaining[:breakPoint]
		remaining = strings.TrimLeft(remaining[breakPoint:], " ")

		var lineStr string
		if isFirstLine && prefix != "" {
			lineStr = bgStyle.Render(prefix) + line
			isFirstLine = false
		} else {
			lineStr = line
		}
		if isCurrentLine {
			lineStr = currentLineBg.Render(lineStr)
		}
		result = append(result, lineStr)
	}

	if len(remaining) > 0 {
		var lineStr string
		if isFirstLine && prefix != "" {
			lineStr = bgStyle.Render(prefix) + remaining
		} else {
			lineStr = remaining
		}
		if isCurrentLine {
			lineStr = currentLineBg.Render(lineStr)
		}
		result = append(result, lineStr)
	}

	return result
}
```

- [ ] **Step 3: Enable highlighting in `renderDiff()`**

In the `default:` branch of `renderDiff()`'s line rendering loop, replace:

```go
// OLD — skip highlighting
highlightedCode := codeContent
```

with:

```go
// Enable token-level syntax highlighting
var highlightedCode string
if m.currentFile != "" && m.highlighter != nil {
    tokens := m.highlighter.HighlightDiffLine(line.Content, m.currentFile)
    highlightedCode = m.renderTokens(tokens)
} else {
    highlightedCode = codeContent
}
```

And replace the `lineStyle` selection block:

```go
// OLD
var lineStyle lipgloss.Style
switch line.Type {
case LineAdded:
    lineStyle = addedStyle
case LineRemoved:
    lineStyle = removedStyle
default:
    lineStyle = lipgloss.NewStyle()
}
wrappedLines := wrapHighlightedLine(prefix, highlightedCode, contentWidth, lineStyle, isCurrentLine)
```

with:

```go
// NEW — use background style
var bgStyle lipgloss.Style
switch line.Type {
case LineAdded:
    bgStyle = addedBgStyle
case LineRemoved:
    bgStyle = removedBgStyle
default:
    bgStyle = lipgloss.NewStyle()
}
wrappedLines := wrapHighlightedLine(prefix, highlightedCode, contentWidth, bgStyle, isCurrentLine)
```

Also remove the unused `visibleLineNo` variable throughout `renderDiff()` (it was incremented but its value was never used for output).

- [ ] **Step 4: Write test verifying highlighting produces output**

Add to `internal/ui/view_test.go`:

```go
func TestRenderDiffHighlighting(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.listWidth = 32
	m.currentFile = "main.go"
	m.diffLines = []DiffLine{
		{Type: LineAdded, OldLineNo: 0, NewLineNo: 1, Content: "+func hello() {}"},
		{Type: LineRemoved, OldLineNo: 1, NewLineNo: 0, Content: "-func world() {}"},
		{Type: LineContext, OldLineNo: 2, NewLineNo: 2, Content: " // unchanged"},
	}
	// Update diffViewport dimensions to prevent zero-width rendering
	m.diffViewport.Width = m.getDiffWidth()
	m.diffViewport.Height = 24

	result := m.renderDiff()

	if result == "" {
		t.Fatal("renderDiff returned empty string")
	}
	// ANSI escape codes should be present (from token coloring or bg styles)
	if !strings.Contains(result, "\x1b[") {
		t.Error("expected ANSI escape codes in renderDiff output, got none")
	}
}
```

Add `"strings"` to the import in `view_test.go`.

- [ ] **Step 5: Run tests**

```bash
cd internal/ui && go test ./... -v
```

Expected: all pass.

- [ ] **Step 6: Manual verification**

```bash
go build -o review ./cmd/review && ./review
```

Navigate to a Go file in the diff. Added lines should have dark green background with token colors. Removed lines should have dark red background with token colors.

- [ ] **Step 7: Commit**

```bash
git add internal/ui/view.go internal/ui/view_test.go
git commit -m "feat: enable syntax highlighting with bg colors in diff view

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

## Task 4: Remove `-/=` keys and fix diff cursor scroll sync

**Files:**
- Modify: `internal/ui/model.go` (keyMap)
- Modify: `internal/ui/update.go` (`handleKey`, `handleDiffKeys`)

- [ ] **Step 1: Remove `WidthInc`/`WidthDec` from keyMap in `model.go`**

In `internal/ui/model.go`:

1. Remove from `keyMap` struct:
```go
// DELETE these two fields:
WidthInc key.Binding
WidthDec key.Binding
```

2. Remove from `FullHelp()`:
```go
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Switch, k.Layout},             // remove k.WidthInc, k.WidthDec
		{k.Edit, k.Refresh, k.Help, k.Quit},
	}
}
```

3. Remove from `defaultKeys`:
```go
// DELETE:
WidthInc: key.NewBinding(
    key.WithKeys("="),
    key.WithHelp("=", "wider list"),
),
WidthDec: key.NewBinding(
    key.WithKeys("-"),
    key.WithHelp("-", "narrower list"),
),
```

- [ ] **Step 2: Remove `-/=` handlers from `handleKey` in `update.go`**

In `internal/ui/update.go`, in `handleKey()`, delete the entire two `if` blocks:

```go
// DELETE this entire block:
keyStr := msg.String()
if keyStr == "-" || keyStr == "_" {
    // ... (entire block)
    return m, nil
}
if keyStr == "=" || keyStr == "+" {
    // ... (entire block)
    return m, nil
}
```

Keep the `keyStr` variable only if it's used elsewhere; otherwise remove it. After cleanup, `handleKey` should look like:

```go
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+w" {
		if m.focus == FocusList {
			m.focus = FocusDiff
		} else {
			m.focus = FocusList
		}
		return m, nil
	}

	if m.focus == FocusList {
		return m.handleListKeys(msg)
	}
	return m.handleDiffKeys(msg)
}
```

- [ ] **Step 3: Fix cursor scroll-sync in `handleDiffKeys`**

Replace the `j/down` and `k/up` cases in `handleDiffKeys` to keep the cursor visible in the viewport:

```go
case "j", "down":
	if m.diffCursor < len(m.diffLines)-1 {
		m.diffCursor++
	}
	// Keep cursor visible: scroll down if cursor is below viewport bottom
	viewBottom := m.diffViewport.YOffset + m.diffViewport.Height - 1
	if m.diffCursor > viewBottom {
		m.diffViewport.ScrollDown(1)
	}
	return m, nil

case "k", "up":
	if m.diffCursor > 0 {
		m.diffCursor--
	}
	// Keep cursor visible: scroll up if cursor is above viewport top
	if m.diffCursor < m.diffViewport.YOffset {
		m.diffViewport.ScrollUp(1)
	}
	return m, nil
```

Also update `ctrl+d/u` and `ctrl+f/b` to keep `diffCursor` clamped:

```go
case "ctrl+d":
	half := m.diffViewport.Height / 2
	m.diffViewport.ScrollDown(half)
	m.diffCursor = min(m.diffCursor+half, len(m.diffLines)-1)
	return m, nil

case "ctrl+u":
	half := m.diffViewport.Height / 2
	m.diffViewport.ScrollUp(half)
	m.diffCursor = max(m.diffCursor-half, 0)
	return m, nil

case "ctrl+f":
	m.diffViewport.ScrollDown(m.diffViewport.Height)
	m.diffCursor = min(m.diffCursor+m.diffViewport.Height, len(m.diffLines)-1)
	return m, nil

case "ctrl+b":
	m.diffViewport.ScrollUp(m.diffViewport.Height)
	m.diffCursor = max(m.diffCursor-m.diffViewport.Height, 0)
	return m, nil
```

Go 1.21+ has `min` and `max` as builtins. Since this project uses Go 1.24, **remove** the existing custom `min` function at the bottom of `update.go` (it shadows the builtin) and do NOT define a custom `max`. The builtin versions work for `int` arguments directly.

- [ ] **Step 4: Add scroll-sync test**

Add to `internal/ui/model_test.go`:

```go
func TestDiffCursorScrollSync(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.listWidth = 32
	m.diffViewport.Width = 88
	m.diffViewport.Height = 5
	// Create 10 diff lines
	for i := 0; i < 10; i++ {
		m.diffLines = append(m.diffLines, DiffLine{
			Type: LineContext, NewLineNo: i + 1, Content: " line",
		})
	}
	m.focus = FocusDiff

	// Move cursor down 6 times — should scroll viewport
	for i := 0; i < 6; i++ {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = result.(Model)
	}

	if m.diffCursor != 6 {
		t.Errorf("expected diffCursor=6, got %d", m.diffCursor)
	}
	// Viewport should have scrolled (YOffset > 0)
	if m.diffViewport.YOffset == 0 {
		t.Error("expected viewport to have scrolled down, YOffset is still 0")
	}
}
```

- [ ] **Step 5: Run tests**

```bash
cd internal/ui && go test ./... -v
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add internal/ui/model.go internal/ui/update.go internal/ui/model_test.go
git commit -m "refactor: remove -/= keys, fix diff cursor scroll sync

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

## Task 5: Redesign status bar

**Files:**
- Modify: `internal/ui/view.go` (`renderStatusBar`)

- [ ] **Step 1: Write test for status bar width**

Add to `internal/ui/view_test.go`:

```go
func TestRenderStatusBarWidth(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.listWidth = 32
	m.files = []FileInfo{{Name: "internal/ui/view.go", Status: "M", Added: 24, Removed: 8}}
	m.cursor = 0
	m.currentFile = "internal/ui/view.go"
	m.focus = FocusDiff
	m.diffCursor = 5

	bar := m.renderStatusBar()
	barWidth := lipgloss.Width(bar)

	if barWidth != 100 {
		t.Errorf("expected status bar width=100, got %d", barWidth)
	}
}
```

Add `"github.com/charmbracelet/lipgloss"` to the imports in `view_test.go`.

- [ ] **Step 2: Implement the new `renderStatusBar()`**

Replace the existing `renderStatusBar()` in `internal/ui/view.go`:

```go
func (m Model) renderStatusBar() string {
	target := m.options.Target
	if target == "" {
		target = "HEAD"
	}

	// Left brand section (accent color)
	brandStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Padding(0, 1)
	brand := brandStyle.Render("review")

	// Right section: file position + focus + line + hints
	focusStr := "List"
	if m.focus == FocusDiff {
		focusStr = "Diff"
	}
	lineInfo := ""
	if m.focus == FocusDiff && len(m.diffLines) > 0 {
		lineInfo = fmt.Sprintf(" ln %d", m.diffCursor+1)
	}
	rightContent := fmt.Sprintf("%d/%d  [%s]%s  ? q", m.cursor+1, len(m.files), focusStr, lineInfo)
	right := statusBarStyle.Padding(0, 1).Render(rightContent)

	// Middle section: target ref + current file + stats (fills remaining width)
	stats := ""
	for _, f := range m.files {
		if f.Name == m.currentFile && (f.Added > 0 || f.Removed > 0) {
			stats = fmt.Sprintf("  +%d -%d", f.Added, f.Removed)
			break
		}
	}
	fileDisplay := m.currentFile
	if fileDisplay == "" {
		fileDisplay = "—"
	}
	middleContent := fmt.Sprintf("  %s  │  %s%s  ", target, fileDisplay, stats)
	middleWidth := m.width - lipgloss.Width(brand) - lipgloss.Width(right)
	if middleWidth < 0 {
		middleWidth = 0
	}
	middle := statusBarStyle.Width(middleWidth).Render(middleContent)

	return brand + middle + right
}
```

- [ ] **Step 3: Run tests**

```bash
cd internal/ui && go test ./... -v
```

Expected: all pass including `TestRenderStatusBarWidth`.

- [ ] **Step 4: Manual verification**

```bash
go build -o review ./cmd/review && ./review
```

Status bar should show: purple `review` brand on left, target+file+stats in center, `N/M [List/Diff] ln X ? q` on right, filling the full terminal width.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/view.go internal/ui/view_test.go
git commit -m "feat: redesign status bar with brand, file info, and focus indicator

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

## Task 6: Add mouse scrolling support

**Files:**
- Modify: `internal/ui/ui.go`
- Modify: `internal/ui/update.go`

- [ ] **Step 1: Enable mouse in `ui.go`**

In `internal/ui/ui.go`, change:

```go
p := tea.NewProgram(m, tea.WithAltScreen())
```

to:

```go
p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
```

- [ ] **Step 2: Add `tea.MouseMsg` handler in `update.go`**

In `Update()`, add a new case after `tea.WindowSizeMsg` (but before the viewport update block at the bottom):

```go
case tea.MouseMsg:
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if msg.X < m.listWidth {
			// Scroll file list up
			if m.cursor > 0 {
				m.fileList.CursorUp()
				m.cursor = m.fileList.Index()
				return m, loadDiff(m.options, m.files[m.cursor].Name)
			}
		} else {
			// Scroll diff viewport up
			m.diffViewport.ScrollUp(3)
			m.diffCursor = max(m.diffCursor-3, 0)
		}
		return m, nil

	case tea.MouseButtonWheelDown:
		if msg.X < m.listWidth {
			// Scroll file list down
			if m.cursor < len(m.files)-1 {
				m.fileList.CursorDown()
				m.cursor = m.fileList.Index()
				return m, loadDiff(m.options, m.files[m.cursor].Name)
			}
		} else {
			// Scroll diff viewport down
			m.diffViewport.ScrollDown(3)
			m.diffCursor = min(m.diffCursor+3, len(m.diffLines)-1)
		}
		return m, nil
	}
	return m, nil
```

Note: `tea.MouseButtonWheelUp` and `tea.MouseButtonWheelDown` are the constants from `github.com/charmbracelet/bubbletea`. If the build fails with undefined constants, check the bubbletea version — in v1.x they may be `tea.MouseWheelUp`/`tea.MouseWheelDown` as `Action` values. In that case use:

```go
case tea.MouseMsg:
	if msg.Action == tea.MouseActionPress {
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			// ... as above
		case tea.MouseButtonWheelDown:
			// ... as above
		}
	}
	return m, nil
```

- [ ] **Step 3: Build and verify compilation**

```bash
cd /path/to/repo && go build ./...
```

Expected: builds with no errors. If there are undefined mouse constants, check bubbletea's actual API with `go doc github.com/charmbracelet/bubbletea MouseMsg` and adjust accordingly.

- [ ] **Step 4: Run all tests**

```bash
cd internal/ui && go test ./... -v
```

Expected: all pass.

- [ ] **Step 5: Manual verification**

```bash
go build -o review ./cmd/review && ./review
```

Try scrolling with the mouse wheel in the file list area and in the diff area. Both should scroll their respective panels.

- [ ] **Step 6: Commit**

```bash
git add internal/ui/ui.go internal/ui/update.go
git commit -m "feat: add mouse wheel scrolling for file list and diff panel

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

## Task 7: Final integration check

**Files:**
- Review: all modified files

- [ ] **Step 1: Run full test suite**

```bash
go test ./... -v
```

Expected: all tests pass.

- [ ] **Step 2: Build and smoke test**

```bash
go build -o review ./cmd/review
./review HEAD~1
```

Check:
- [ ] Layout: list panel and diff panel separated by a visible border line
- [ ] Diff: added lines have dark green background + syntax colors; removed lines have dark red background
- [ ] Status bar: purple `review` brand | file path and stats | file count + `[List/Diff]` + line number
- [ ] `j/k` in diff mode: cursor moves with `>>` highlight, viewport scrolls to keep cursor visible
- [ ] `ctrl+d/u`, `g/G` work correctly
- [ ] Mouse wheel in list area scrolls through files (and loads diff)
- [ ] Mouse wheel in diff area scrolls the diff viewport
- [ ] `?` shows help overlay
- [ ] `q` quits

- [ ] **Step 3: Final commit (tag if desired)**

```bash
git add -A
git commit -m "chore: final integration — ui refactor complete

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```
