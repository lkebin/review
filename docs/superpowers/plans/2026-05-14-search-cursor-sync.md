# Search Mode Cursor Sync Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show the real terminal cursor (instead of a fake `▌` character) when the user is typing in search mode, so cursor shape and behavior match the terminal's own configuration.

**Architecture:** Two targeted changes — `RenderSearchBar` in `statusbar.go` embeds ANSI save/restore sequences (`\033[s` / `\033[u`) around the search content so the terminal cursor parks at the end of the query; `app.go` emits `tea.ShowCursor`/`tea.HideCursor` commands on the three mode-transition events.

**Tech Stack:** Go, charmbracelet/bubbletea v1.3.10, charmbracelet/lipgloss

---

### Task 1: Replace fake cursor in `RenderSearchBar` with ANSI save/restore

**Files:**
- Modify: `internal/ui/statusbar.go`
- Test: `internal/ui/statusbar_test.go`

- [ ] **Step 1: Write the failing tests**

Add to `internal/ui/statusbar_test.go`:

```go
func TestRenderSearchBarTypingCursor(t *testing.T) {
	th := DefaultTheme()

	// typing=true: no ▌, but ANSI save/restore present
	bar := RenderSearchBar("foo", FocusList, 80, th, true)
	if strings.Contains(bar, "▌") {
		t.Error("typing=true: bar must not contain fake cursor ▌")
	}
	if !strings.Contains(bar, "\033[s") {
		t.Error("typing=true: bar must contain ANSI save cursor \\033[s")
	}
	if !strings.Contains(bar, "\033[u") {
		t.Error("typing=true: bar must contain ANSI restore cursor \\033[u")
	}

	// typing=false: no cursor character or ANSI sequences
	bar2 := RenderSearchBar("foo", FocusList, 80, th, false)
	if strings.Contains(bar2, "▌") {
		t.Error("typing=false: bar must not contain ▌")
	}
	if strings.Contains(bar2, "\033[s") {
		t.Error("typing=false: bar must not contain ANSI save cursor")
	}
}

func TestRenderSearchBarWidthWithCursor(t *testing.T) {
	th := DefaultTheme()
	for _, w := range []int{40, 80, 120} {
		bar := RenderSearchBar("foo", FocusList, w, th, true)
		got := lipgloss.Width(bar)
		if got != w {
			t.Errorf("width %d: search bar visual width = %d, want %d", w, got, w)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```
go test ./internal/ui/ -run TestRenderSearchBarTypingCursor -v
go test ./internal/ui/ -run TestRenderSearchBarWidthWithCursor -v
```

Expected: `FAIL` — `TestRenderSearchBarTypingCursor` fails because bar still contains `▌` and lacks `\033[s`.

- [ ] **Step 3: Implement the change in `RenderSearchBar`**

Replace the cursor/prompt/gap block in `internal/ui/statusbar.go` (lines 61–72):

```go
// Before (lines 61-72):
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
```

```go
// After:
	prompt := lipgloss.NewStyle().Bold(true).Render("/") + query
	right := " " + panel + " "
	gap := width - lipgloss.Width(prompt) - len(right)
	if gap < 0 {
		gap = 0
	}
	var bar string
	if typing {
		// \033[s saves the terminal cursor position (end of query).
		// \033[u restores it after rendering the rest of the line, so the
		// real terminal cursor parks here instead of at the line end.
		bar = prompt + "\033[s" + strings.Repeat(" ", gap) + right + "\033[u"
	} else {
		bar = prompt + strings.Repeat(" ", gap) + right
	}
	return theme.StatusBarStyle().Width(width).Render(bar)
```

- [ ] **Step 4: Run tests to verify they pass**

```
go test ./internal/ui/ -run TestRenderSearchBar -v
```

Expected: all `TestRenderSearchBar*` tests `PASS`.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/statusbar.go internal/ui/statusbar_test.go
git commit -m "feat(ui): use ANSI save/restore cursor in search bar instead of fake ▌"
```

---

### Task 2: Emit ShowCursor/HideCursor on search mode transitions

**Files:**
- Modify: `internal/ui/app.go`
- Test: `internal/ui/app_test.go`

- [ ] **Step 1: Write the failing tests**

Add to `internal/ui/app_test.go`:

```go
func TestSearchOpenShowsCursor(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.files = []FileInfo{{Status: "M", Name: "a.go"}}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if cmd == nil {
		t.Error("/ key should produce ShowCursor command")
	}
}

func TestSearchEnterHidesCursor(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.searchMode = true
	m.searchQuery = "" // empty query so doSearch returns nil cmd

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("Enter in search mode should produce HideCursor command")
	}
}

func TestSearchEscHidesCursor(t *testing.T) {
	m := NewModel(Options{Target: "HEAD"})
	m.loading = false
	m.searchMode = true
	m.searchQuery = "foo"

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(Model)
	if m.searchMode {
		t.Error("Esc should exit search mode")
	}
	if cmd == nil {
		t.Error("Esc in search mode should produce HideCursor command")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```
go test ./internal/ui/ -run "TestSearchOpenShowsCursor|TestSearchEnterHidesCursor|TestSearchEscHidesCursor" -v
```

Expected: `FAIL` — all three tests fail because the commands are currently `nil`.

- [ ] **Step 3: Implement the three mode-transition changes in `app.go`**

**Change 1** — `ActionSearchOpen` (line 242–245):

```go
// Before:
	case ActionSearchOpen:
		m.searchMode = true
		m.searchQuery = ""
		return m, nil
```

```go
// After:
	case ActionSearchOpen:
		m.searchMode = true
		m.searchQuery = ""
		return m, tea.ShowCursor
```

**Change 2** — `handleSearchKey` enter case (lines 187–189):

```go
// Before:
	case "enter":
		m.searchMode = false
		return m.doSearch(true) // jump to first match
```

```go
// After:
	case "enter":
		m.searchMode = false
		model, cmd := m.doSearch(true)
		return model, tea.Batch(tea.HideCursor, cmd)
```

**Change 3** — `handleSearchKey` esc case (lines 190–193):

```go
// Before:
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		return m, nil
```

```go
// After:
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		return m, tea.HideCursor
```

- [ ] **Step 4: Run all tests**

```
go test ./internal/ui/ -v
```

Expected: all tests `PASS`, including the three new cursor tests and all pre-existing tests.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat(ui): show/hide real terminal cursor on search mode entry/exit"
```
