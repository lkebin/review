# UI Layer Rewrite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite `internal/ui` to deliver Vim-consistent key bindings, clean layout (file list + single-file diff), fixed line number columns, inline diff highlighting, and optimized Chroma syntax highlighting.

**Architecture:** Preserve `internal/git`, `internal/diff`, `internal/highlight` layers. Rewrite `internal/ui` as 9 focused files (app, keymap, filelist, diffview, lineno, inlinediff, styles, statusbar, viewport). Use `diff.Line`/`diff.LineType` directly—no type duplication. Add `TokenizeFile` batch method to highlight package.

**Tech Stack:** Go 1.24, BubbleTea, lipgloss, Chroma v2, standard library `unicode/utf8` for Myers diff.

---

### Task 1: Scaffold — Remove Old UI, Create Minimal Skeleton

**Files:**
- Delete: `internal/ui/model.go`, `internal/ui/update.go`, `internal/ui/view.go`, `internal/ui/git.go`, `internal/ui/model_test.go`, `internal/ui/view_test.go`
- Create: `internal/ui/app.go`
- Create: `internal/ui/styles.go` (empty Theme placeholder)
- Create: `internal/ui/testmain_test.go`
- Keep: `internal/ui/ui.go` (Options + Run)

- [ ] **Step 1: Delete old UI files**

```bash
cd /Users/kbliu/Workspace/project/vim-code-review
rm internal/ui/model.go internal/ui/update.go internal/ui/view.go internal/ui/git.go
rm internal/ui/model_test.go internal/ui/view_test.go
```

- [ ] **Step 2: Create minimal app.go**

```go
// internal/ui/app.go
package ui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/kbliu/review/internal/diff"
	"github.com/kbliu/review/internal/git"
)

// FileInfo represents a file with diff statistics.
type FileInfo struct {
	Status  string
	Name    string
	Added   int
	Removed int
}

// FocusType indicates which panel has focus.
type FocusType int

const (
	FocusList FocusType = iota
	FocusDiff
)

// Model is the top-level BubbleTea model.
type Model struct {
	opts      Options
	width     int
	height    int
	files     []FileInfo
	err       error
	loading   bool
}

// NewModel creates a new Model.
func NewModel(opts Options) Model {
	return Model{opts: opts, loading: true}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error()
	}
	return "review — press any key to quit"
}

// gitOptions converts ui.Options to git.Options.
func gitOptions(opts Options) git.Options {
	return git.Options{
		Target:       opts.Target,
		Staged:       opts.Staged,
		ContextLines: opts.ContextLines,
	}
}

// loadFileList fetches files and calculates stats.
func loadFileList(opts Options) ([]FileInfo, error) {
	gopts := gitOptions(opts)
	files, err := git.GetFiles(gopts)
	if err != nil {
		return nil, err
	}
	result := make([]FileInfo, len(files))
	for i, f := range files {
		content, err := git.GetDiff(gopts, f.Name)
		if err != nil {
			result[i] = FileInfo{Status: f.Status, Name: f.Name}
			continue
		}
		lines := diff.Parse(content)
		stats := diff.CalculateStats(lines)
		result[i] = FileInfo{
			Status:  f.Status,
			Name:    f.Name,
			Added:   stats.Added,
			Removed: stats.Removed,
		}
	}
	return result, nil
}
```

- [ ] **Step 3: Create testmain_test.go**

```go
// internal/ui/testmain_test.go
package ui

import (
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	os.Exit(m.Run())
}
```

- [ ] **Step 4: Verify build and tests pass**

```bash
cd /Users/kbliu/Workspace/project/vim-code-review
go build ./...
go test ./internal/ui/ -v
```

Expected: build succeeds, no test failures (no tests to run yet, but package compiles).

- [ ] **Step 5: Commit**

```bash
git add -A internal/ui/
git commit -m "scaffold: remove old UI, create minimal skeleton for rewrite"
```

---

### Task 2: styles.go — Theme Struct and Style Definitions

**Files:**
- Create: `internal/ui/styles.go`
- Create: `internal/ui/styles_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/ui/styles_test.go
package ui

import "testing"

func TestDefaultThemeHasAllColors(t *testing.T) {
	th := DefaultTheme()

	checks := []struct {
		name  string
		color string
	}{
		{"AddedBg", th.AddedBg},
		{"RemovedBg", th.RemovedBg},
		{"AddedCursorBg", th.AddedCursorBg},
		{"RemovedCursorBg", th.RemovedCursorBg},
		{"CursorBg", th.CursorBg},
		{"LineNoBg", th.LineNoBg},
		{"LineNoFg", th.LineNoFg},
		{"SepFg", th.SepFg},
		{"HunkFg", th.HunkFg},
		{"NormalFg", th.NormalFg},
		{"StatusBarBg", th.StatusBarBg},
		{"StatusBarFg", th.StatusBarFg},
		{"FileSelectedBg", th.FileSelectedBg},
		{"FileSelectedFg", th.FileSelectedFg},
		{"InlineAddBg", th.InlineAddBg},
		{"InlineDelBg", th.InlineDelBg},
	}
	for _, c := range checks {
		if c.color == "" {
			t.Errorf("DefaultTheme().%s is empty", c.name)
		}
	}
}

func TestStatusColors(t *testing.T) {
	th := DefaultTheme()
	for _, status := range []string{"M", "A", "D", "R", "C"} {
		if th.StatusColor(status) == "" {
			t.Errorf("StatusColor(%q) returned empty", status)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/ -run TestDefaultTheme -v
go test ./internal/ui/ -run TestStatusColors -v
```

Expected: FAIL — `DefaultTheme` and `StatusColor` not defined.

- [ ] **Step 3: Implement styles.go**

```go
// internal/ui/styles.go
package ui

import "github.com/charmbracelet/lipgloss"

// Theme holds all color values as ANSI 256-color strings.
type Theme struct {
	NormalFg   string
	AddedBg    string
	RemovedBg  string
	AddedCursorBg   string
	RemovedCursorBg string
	CursorBg        string
	InlineAddBg     string
	InlineDelBg     string
	LineNoFg        string
	LineNoBg        string
	SepFg           string
	HunkFg          string
	StatusBarFg     string
	StatusBarBg     string
	FileSelectedFg  string
	FileSelectedBg  string
	StatusM string
	StatusA string
	StatusD string
	StatusR string
	StatusC string
}

// DefaultTheme returns the built-in 256-color theme.
func DefaultTheme() Theme {
	return Theme{
		NormalFg:        "252",
		AddedBg:         "22",
		RemovedBg:       "52",
		AddedCursorBg:   "28",
		RemovedCursorBg: "88",
		CursorBg:        "236",
		InlineAddBg:     "28",
		InlineDelBg:     "88",
		LineNoFg:        "240",
		LineNoBg:        "233",
		SepFg:           "238",
		HunkFg:          "140",
		StatusBarFg:     "252",
		StatusBarBg:     "236",
		FileSelectedFg:  "16",
		FileSelectedBg:  "75",
		StatusM:         "178",
		StatusA:         "40",
		StatusD:         "167",
		StatusR:         "133",
		StatusC:         "73",
	}
}

// StatusColor returns the foreground color for a file status badge.
func (th Theme) StatusColor(status string) string {
	switch status {
	case "M":
		return th.StatusM
	case "A":
		return th.StatusA
	case "D":
		return th.StatusD
	case "R":
		return th.StatusR
	case "C":
		return th.StatusC
	default:
		return th.NormalFg
	}
}

// Convenience lipgloss style builders.

func (th Theme) LineNoStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.LineNoFg)).
		Background(lipgloss.Color(th.LineNoBg))
}

func (th Theme) SepStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.SepFg)).
		Background(lipgloss.Color(th.LineNoBg))
}

func (th Theme) HunkStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(th.HunkFg))
}

func (th Theme) StatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.StatusBarFg)).
		Background(lipgloss.Color(th.StatusBarBg))
}

func (th Theme) FileSelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.FileSelectedFg)).
		Background(lipgloss.Color(th.FileSelectedBg)).
		Bold(true)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/ -run "TestDefaultTheme|TestStatusColors" -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/styles.go internal/ui/styles_test.go
git commit -m "feat(ui): add Theme struct and default 256-color scheme"
```

---

### Task 3: keymap.go — Prefix Key State Machine

**Files:**
- Create: `internal/ui/keymap.go`
- Create: `internal/ui/keymap_test.go`

- [ ] **Step 1: Write failing tests for basic key handling**

```go
// internal/ui/keymap_test.go
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(s string) tea.KeyMsg {
	// Handles special keys
	switch s {
	case "ctrl+w":
		return tea.KeyMsg{Type: tea.KeyCtrlW}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+u":
		return tea.KeyMsg{Type: tea.KeyCtrlU}
	case "ctrl+f":
		return tea.KeyMsg{Type: tea.KeyCtrlF}
	case "ctrl+b":
		return tea.KeyMsg{Type: tea.KeyCtrlB}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestKeyMapSimpleKeys(t *testing.T) {
	km := NewKeyMapper()

	cases := []struct {
		key    string
		focus  FocusType
		expect Action
	}{
		{"j", FocusList, ActionCursorDown},
		{"k", FocusList, ActionCursorUp},
		{"j", FocusDiff, ActionCursorDown},
		{"k", FocusDiff, ActionCursorUp},
		{"enter", FocusList, ActionEnter},
		{"G", FocusList, ActionBottom},
		{"G", FocusDiff, ActionBottom},
		{"n", FocusDiff, ActionNextHunk},
		{"N", FocusDiff, ActionPrevHunk},
		{"q", FocusList, ActionQuit},
		{"?", FocusList, ActionHelp},
		{"ctrl+d", FocusDiff, ActionHalfPageDown},
		{"ctrl+u", FocusDiff, ActionHalfPageUp},
		{"ctrl+f", FocusDiff, ActionPageDown},
		{"ctrl+b", FocusDiff, ActionPageUp},
	}

	for _, tc := range cases {
		t.Run(tc.key+"_"+focusName(tc.focus), func(t *testing.T) {
			km.Reset()
			action := km.HandleKey(key(tc.key), tc.focus)
			if action != tc.expect {
				t.Errorf("HandleKey(%q, %v) = %v, want %v", tc.key, tc.focus, action, tc.expect)
			}
		})
	}
}

func focusName(f FocusType) string {
	if f == FocusList {
		return "list"
	}
	return "diff"
}

func TestKeyMapCtrlWPrefix(t *testing.T) {
	km := NewKeyMapper()

	// Ctrl+W should enter prefix state, return ActionNone
	action := km.HandleKey(key("ctrl+w"), FocusDiff)
	if action != ActionNone {
		t.Fatalf("Ctrl+W alone = %v, want ActionNone", action)
	}

	// Follow with 'h' → ActionFocusLeft
	action = km.HandleKey(key("h"), FocusDiff)
	if action != ActionFocusLeft {
		t.Errorf("Ctrl+W h = %v, want ActionFocusLeft", action)
	}

	// Should be back to Normal state
	km.Reset()
	action = km.HandleKey(key("ctrl+w"), FocusList)
	if action != ActionNone {
		t.Fatalf("Ctrl+W alone = %v, want ActionNone", action)
	}
	action = km.HandleKey(key("l"), FocusList)
	if action != ActionFocusRight {
		t.Errorf("Ctrl+W l = %v, want ActionFocusRight", action)
	}
}

func TestKeyMapCtrlWResize(t *testing.T) {
	km := NewKeyMapper()

	km.HandleKey(key("ctrl+w"), FocusList)
	action := km.HandleKey(key(">"), FocusList)
	if action != ActionGrowPanel {
		t.Errorf("Ctrl+W > = %v, want ActionGrowPanel", action)
	}

	km.HandleKey(key("ctrl+w"), FocusList)
	action = km.HandleKey(key("<"), FocusList)
	if action != ActionShrinkPanel {
		t.Errorf("Ctrl+W < = %v, want ActionShrinkPanel", action)
	}
}

func TestKeyMapGGPrefix(t *testing.T) {
	km := NewKeyMapper()

	action := km.HandleKey(key("g"), FocusDiff)
	if action != ActionNone {
		t.Fatalf("single g = %v, want ActionNone", action)
	}

	action = km.HandleKey(key("g"), FocusDiff)
	if action != ActionTop {
		t.Errorf("gg = %v, want ActionTop", action)
	}
}

func TestKeyMapPrefixInvalidFollowUp(t *testing.T) {
	km := NewKeyMapper()

	// Ctrl+W followed by invalid key → ActionNone, back to Normal
	km.HandleKey(key("ctrl+w"), FocusDiff)
	action := km.HandleKey(key("x"), FocusDiff)
	if action != ActionNone {
		t.Errorf("Ctrl+W x = %v, want ActionNone", action)
	}

	// Next key should work normally
	action = km.HandleKey(key("j"), FocusDiff)
	if action != ActionCursorDown {
		t.Errorf("j after invalid prefix = %v, want ActionCursorDown", action)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/ -run TestKeyMap -v
```

Expected: FAIL — `NewKeyMapper`, `Action` types not defined.

- [ ] **Step 3: Implement keymap.go**

```go
// internal/ui/keymap.go
package ui

import tea "github.com/charmbracelet/bubbletea"

// Action represents a semantic UI action.
type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionHelp
	ActionCursorUp
	ActionCursorDown
	ActionEnter
	ActionTop
	ActionBottom
	ActionHalfPageDown
	ActionHalfPageUp
	ActionPageDown
	ActionPageUp
	ActionNextHunk
	ActionPrevHunk
	ActionFocusLeft
	ActionFocusRight
	ActionGrowPanel
	ActionShrinkPanel
)

// keyState represents the state machine state.
type keyState int

const (
	stateNormal keyState = iota
	stateCtrlW
	stateG
)

// KeyMapper translates key events into semantic Actions using a prefix key state machine.
type KeyMapper struct {
	state keyState
}

// NewKeyMapper creates a KeyMapper in Normal state.
func NewKeyMapper() *KeyMapper {
	return &KeyMapper{state: stateNormal}
}

// Reset returns the state machine to Normal.
func (km *KeyMapper) Reset() {
	km.state = stateNormal
}

// HandleKey processes a key event and returns the corresponding Action.
func (km *KeyMapper) HandleKey(msg tea.KeyMsg, focus FocusType) Action {
	switch km.state {
	case stateCtrlW:
		km.state = stateNormal
		return km.handleCtrlW(msg)
	case stateG:
		km.state = stateNormal
		if msg.String() == "g" {
			return ActionTop
		}
		return ActionNone
	default:
		return km.handleNormal(msg, focus)
	}
}

func (km *KeyMapper) handleNormal(msg tea.KeyMsg, focus FocusType) Action {
	s := msg.String()

	// Global keys (work in any focus)
	switch s {
	case "q", "ctrl+c":
		return ActionQuit
	case "?":
		return ActionHelp
	case "ctrl+w":
		km.state = stateCtrlW
		return ActionNone
	case "g":
		km.state = stateG
		return ActionNone
	case "G":
		return ActionBottom
	case "j", "down":
		return ActionCursorDown
	case "k", "up":
		return ActionCursorUp
	}

	// Focus-specific keys
	if focus == FocusList {
		switch s {
		case "enter":
			return ActionEnter
		}
	} else {
		switch s {
		case "ctrl+d":
			return ActionHalfPageDown
		case "ctrl+u":
			return ActionHalfPageUp
		case "ctrl+f":
			return ActionPageDown
		case "ctrl+b":
			return ActionPageUp
		case "n":
			return ActionNextHunk
		case "N":
			return ActionPrevHunk
		}
	}

	return ActionNone
}

func (km *KeyMapper) handleCtrlW(msg tea.KeyMsg) Action {
	switch msg.String() {
	case "h", "left":
		return ActionFocusLeft
	case "l", "right":
		return ActionFocusRight
	case ">":
		return ActionGrowPanel
	case "<":
		return ActionShrinkPanel
	}
	return ActionNone
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/ -run TestKeyMap -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/keymap.go internal/ui/keymap_test.go
git commit -m "feat(ui): add prefix key state machine (Ctrl+W, gg)"
```

---

### Task 4: inlinediff.go — Character-Level Inline Diff

**Files:**
- Create: `internal/ui/inlinediff.go`
- Create: `internal/ui/inlinediff_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/ui/inlinediff_test.go
package ui

import (
	"testing"

	"github.com/kbliu/review/internal/diff"
)

func TestComputeInlineDiffSimple(t *testing.T) {
	// "hello" → "hallo": the 'e' changed to 'a'
	old := "hello"
	new := "hallo"
	oldSpans, newSpans := ComputeInlineDiff(old, new)

	if len(oldSpans) != 1 {
		t.Fatalf("oldSpans count = %d, want 1", len(oldSpans))
	}
	if len(newSpans) != 1 {
		t.Fatalf("newSpans count = %d, want 1", len(newSpans))
	}

	// 'e' is at byte 1
	if oldSpans[0].Start != 1 || oldSpans[0].End != 2 {
		t.Errorf("oldSpan = [%d,%d), want [1,2)", oldSpans[0].Start, oldSpans[0].End)
	}
	if newSpans[0].Start != 1 || newSpans[0].End != 2 {
		t.Errorf("newSpan = [%d,%d), want [1,2)", newSpans[0].Start, newSpans[0].End)
	}
}

func TestComputeInlineDiffPrefixInsertion(t *testing.T) {
	// "println()" → "fmt.Println()"
	old := "println()"
	new := "fmt.Println()"
	oldSpans, newSpans := ComputeInlineDiff(old, new)

	// old: "p" changed (byte 0..1)
	// new: "fmt.P" inserted/changed (byte 0..5)
	if len(oldSpans) != 1 || len(newSpans) != 1 {
		t.Fatalf("spans = old:%d new:%d, want 1 each", len(oldSpans), len(newSpans))
	}
	if oldSpans[0].Start != 0 || oldSpans[0].End != 1 {
		t.Errorf("oldSpan = [%d,%d), want [0,1)", oldSpans[0].Start, oldSpans[0].End)
	}
	if newSpans[0].Start != 0 || newSpans[0].End != 5 {
		t.Errorf("newSpan = [%d,%d), want [0,5)", newSpans[0].Start, newSpans[0].End)
	}
}

func TestComputeInlineDiffIdentical(t *testing.T) {
	oldSpans, newSpans := ComputeInlineDiff("same", "same")
	if len(oldSpans) != 0 || len(newSpans) != 0 {
		t.Errorf("identical strings should have no spans")
	}
}

func TestComputeInlineDiffCompletelyDifferent(t *testing.T) {
	oldSpans, newSpans := ComputeInlineDiff("abc", "xyz")
	if len(oldSpans) != 1 {
		t.Fatalf("oldSpans = %d, want 1", len(oldSpans))
	}
	if oldSpans[0].Start != 0 || oldSpans[0].End != 3 {
		t.Errorf("oldSpan = [%d,%d), want [0,3)", oldSpans[0].Start, oldSpans[0].End)
	}
	if newSpans[0].Start != 0 || newSpans[0].End != 3 {
		t.Errorf("newSpan = [%d,%d), want [0,3)", newSpans[0].Start, newSpans[0].End)
	}
}

func TestPairDiffLines(t *testing.T) {
	lines := []diff.Line{
		{Type: diff.LineContext, Content: " context"},
		{Type: diff.LineRemoved, Content: "-old1"},
		{Type: diff.LineRemoved, Content: "-old2"},
		{Type: diff.LineAdded, Content: "+new1"},
		{Type: diff.LineAdded, Content: "+new2"},
		{Type: diff.LineContext, Content: " context"},
	}

	pairs := PairDiffLines(lines)
	if len(pairs) != 2 {
		t.Fatalf("pair count = %d, want 2", len(pairs))
	}
	if pairs[0].OldIdx != 1 || pairs[0].NewIdx != 3 {
		t.Errorf("pair[0] = {%d,%d}, want {1,3}", pairs[0].OldIdx, pairs[0].NewIdx)
	}
	if pairs[1].OldIdx != 2 || pairs[1].NewIdx != 4 {
		t.Errorf("pair[1] = {%d,%d}, want {2,4}", pairs[1].OldIdx, pairs[1].NewIdx)
	}
}

func TestPairDiffLinesUneven(t *testing.T) {
	// 1 removed, 2 added → only 1 pair
	lines := []diff.Line{
		{Type: diff.LineRemoved, Content: "-old"},
		{Type: diff.LineAdded, Content: "+new1"},
		{Type: diff.LineAdded, Content: "+new2"},
	}
	pairs := PairDiffLines(lines)
	if len(pairs) != 1 {
		t.Fatalf("pair count = %d, want 1", len(pairs))
	}
	if pairs[0].OldIdx != 0 || pairs[0].NewIdx != 1 {
		t.Errorf("pair[0] = {%d,%d}, want {0,1}", pairs[0].OldIdx, pairs[0].NewIdx)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/ -run "TestComputeInlineDiff|TestPairDiffLines" -v
```

Expected: FAIL — types not defined.

- [ ] **Step 3: Implement inlinediff.go**

```go
// internal/ui/inlinediff.go
package ui

import (
	"github.com/kbliu/review/internal/diff"
)

// InlineSpan marks a byte range in a line's content that was changed.
type InlineSpan struct {
	Start int // byte offset (inclusive)
	End   int // byte offset (exclusive)
}

// LinePair associates a removed line with an added line by index.
type LinePair struct {
	OldIdx int
	NewIdx int
}

// ComputeInlineDiff finds the changed character spans between two strings.
// It uses a common-prefix/suffix approach: finds the matching head and tail,
// and marks the middle region as changed.
func ComputeInlineDiff(oldStr, newStr string) ([]InlineSpan, []InlineSpan) {
	oldRunes := []rune(oldStr)
	newRunes := []rune(newStr)

	// Find common prefix length (in runes)
	prefixLen := 0
	for prefixLen < len(oldRunes) && prefixLen < len(newRunes) &&
		oldRunes[prefixLen] == newRunes[prefixLen] {
		prefixLen++
	}

	// Find common suffix length (in runes), not overlapping with prefix
	suffixLen := 0
	for suffixLen < len(oldRunes)-prefixLen && suffixLen < len(newRunes)-prefixLen &&
		oldRunes[len(oldRunes)-1-suffixLen] == newRunes[len(newRunes)-1-suffixLen] {
		suffixLen++
	}

	// Convert rune offsets back to byte offsets
	oldPrefixBytes := len(string(oldRunes[:prefixLen]))
	oldSuffixBytes := len(string(oldRunes[len(oldRunes)-suffixLen:]))
	newPrefixBytes := len(string(newRunes[:prefixLen]))
	newSuffixBytes := len(string(newRunes[len(newRunes)-suffixLen:]))

	oldStart := oldPrefixBytes
	oldEnd := len(oldStr) - oldSuffixBytes
	newStart := newPrefixBytes
	newEnd := len(newStr) - newSuffixBytes

	if oldStart >= oldEnd && newStart >= newEnd {
		return nil, nil // no difference
	}

	var oldSpans, newSpans []InlineSpan
	if oldStart < oldEnd {
		oldSpans = []InlineSpan{{Start: oldStart, End: oldEnd}}
	}
	if newStart < newEnd {
		newSpans = []InlineSpan{{Start: newStart, End: newEnd}}
	}
	return oldSpans, newSpans
}

// PairDiffLines pairs consecutive removed lines with consecutive added lines.
// Within a block of [-lines][+lines], they are paired 1:1 in order.
// Excess lines on either side are left unpaired.
func PairDiffLines(lines []diff.Line) []LinePair {
	var pairs []LinePair
	i := 0
	for i < len(lines) {
		// Find a block of removed lines
		remStart := i
		for i < len(lines) && lines[i].Type == diff.LineRemoved {
			i++
		}
		remEnd := i

		// Find immediately following block of added lines
		addStart := i
		for i < len(lines) && lines[i].Type == diff.LineAdded {
			i++
		}
		addEnd := i

		// Pair them 1:1
		remCount := remEnd - remStart
		addCount := addEnd - addStart
		pairCount := remCount
		if addCount < pairCount {
			pairCount = addCount
		}
		for j := 0; j < pairCount; j++ {
			pairs = append(pairs, LinePair{
				OldIdx: remStart + j,
				NewIdx: addStart + j,
			})
		}

		// If no removed/added block was found, advance past this line
		if remStart == remEnd && addStart == addEnd {
			i++
		}
	}
	return pairs
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/ -run "TestComputeInlineDiff|TestPairDiffLines" -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/inlinediff.go internal/ui/inlinediff_test.go
git commit -m "feat(ui): add character-level inline diff with line pairing"
```

---

### Task 5: highlight — Add TokenizeFile Batch Method

**Files:**
- Modify: `internal/highlight/chroma.go`
- Create: `internal/highlight/chroma_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/highlight/chroma_test.go
package highlight

import "testing"

func TestTokenizeFileReturnsPerLineTokens(t *testing.T) {
	h := New("github")
	lines := []string{
		"package main",
		"",
		"func main() {",
		"    println(\"hello\")",
		"}",
	}

	result := h.TokenizeFile("main.go", lines)

	if len(result) != len(lines) {
		t.Fatalf("TokenizeFile returned %d lines, want %d", len(result), len(lines))
	}

	// First line should have tokens for "package" and "main"
	if len(result[0]) == 0 {
		t.Error("first line has no tokens")
	}

	// Concatenating all token text for a line should reconstruct the original
	for i, lineTokens := range result {
		var reconstructed string
		for _, tok := range lineTokens {
			reconstructed += tok.Text
		}
		if reconstructed != lines[i] {
			t.Errorf("line %d: reconstructed=%q, want=%q", i, reconstructed, lines[i])
		}
	}
}

func TestTokenizeFileUnknownLanguage(t *testing.T) {
	h := New("github")
	lines := []string{"some content", "more content"}

	result := h.TokenizeFile("unknown.xyz", lines)

	if len(result) != 2 {
		t.Fatalf("result lines = %d, want 2", len(result))
	}
	// Should still return tokens (even if generic)
	for i, lineTokens := range result {
		var text string
		for _, tok := range lineTokens {
			text += tok.Text
		}
		if text != lines[i] {
			t.Errorf("line %d: text=%q, want=%q", i, text, lines[i])
		}
	}
}

func TestTokenizeFileEmpty(t *testing.T) {
	h := New("github")
	result := h.TokenizeFile("main.go", nil)
	if len(result) != 0 {
		t.Errorf("nil input should return empty, got %d", len(result))
	}

	result = h.TokenizeFile("main.go", []string{})
	if len(result) != 0 {
		t.Errorf("empty input should return empty, got %d", len(result))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/highlight/ -run TestTokenizeFile -v
```

Expected: FAIL — `TokenizeFile` method not defined.

- [ ] **Step 3: Implement TokenizeFile in chroma.go**

Add this method to `internal/highlight/chroma.go`:

```go
// TokenizeFile tokenizes multiple lines of code in a single Chroma call.
// filename is used for lexer detection. lines are plain code without diff prefixes.
// Returns per-line token slices. Concatenating tokens per line reconstructs the input.
func (h *SimpleHighlighter) TokenizeFile(filename string, lines []string) [][]Token {
	if len(lines) == 0 {
		return nil
	}

	lexer := lexers.Match(filename)
	if lexer == nil {
		// No lexer found — return each line as a single generic token
		result := make([][]Token, len(lines))
		for i, line := range lines {
			result[i] = []Token{{Text: line, TokenType: ""}}
		}
		return result
	}
	lexer = chroma.Coalesce(lexer)

	// Concatenate all lines with newlines
	full := strings.Join(lines, "\n")

	iterator, err := lexer.Tokenise(nil, full)
	if err != nil {
		result := make([][]Token, len(lines))
		for i, line := range lines {
			result[i] = []Token{{Text: line, TokenType: ""}}
		}
		return result
	}

	// Split tokens by line
	result := make([][]Token, len(lines))
	lineIdx := 0

	for _, tok := range iterator.Tokens() {
		text := tok.Value
		tokType := tok.Type.String()

		// A token may contain newlines; split it across lines
		for text != "" {
			if lineIdx >= len(result) {
				break
			}
			nlPos := strings.Index(text, "\n")
			if nlPos == -1 {
				// No newline — entire token belongs to current line
				result[lineIdx] = append(result[lineIdx], Token{Text: text, TokenType: tokType})
				break
			}
			// Part before newline goes to current line
			if nlPos > 0 {
				result[lineIdx] = append(result[lineIdx], Token{Text: text[:nlPos], TokenType: tokType})
			}
			// Move to next line
			lineIdx++
			text = text[nlPos+1:]
		}
	}

	// Ensure all lines have at least an empty token slice
	for i := range result {
		if result[i] == nil {
			result[i] = []Token{}
		}
	}

	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/highlight/ -run TestTokenizeFile -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/highlight/chroma.go internal/highlight/chroma_test.go
git commit -m "feat(highlight): add TokenizeFile for batch tokenization"
```

---

### Task 6: lineno.go — Line Number Formatting

**Files:**
- Create: `internal/ui/lineno.go`
- Create: `internal/ui/lineno_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/ui/lineno_test.go
package ui

import "testing"

func TestCalcLineNoWidth(t *testing.T) {
	cases := []struct {
		maxLineNo int
		want      int
	}{
		{0, 1},
		{9, 1},
		{10, 2},
		{99, 2},
		{100, 3},
		{999, 3},
		{1000, 4},
		{9999, 4},
		{10000, 5},
		{99999, 5},
	}
	for _, tc := range cases {
		got := CalcLineNoWidth(tc.maxLineNo)
		if got != tc.want {
			t.Errorf("CalcLineNoWidth(%d) = %d, want %d", tc.maxLineNo, got, tc.want)
		}
	}
}

func TestFormatLineNo(t *testing.T) {
	cases := []struct {
		left, right int
		width       int
		want        string
	}{
		{12, 15, 4, "  12   15 "},
		{0, 17, 4, "       17 "},
		{14, 0, 4, "  14      "},
		{0, 0, 4, "          "}, // hunk header placeholder
	}
	for _, tc := range cases {
		got := FormatLineNo(tc.left, tc.right, tc.width)
		if got != tc.want {
			t.Errorf("FormatLineNo(%d, %d, %d) = %q, want %q",
				tc.left, tc.right, tc.width, got, tc.want)
		}
	}
}

func TestLineNoColumnWidth(t *testing.T) {
	// Total column width = digitWidth*2 + 2 (spaces) for the "LLLL RRRR " format
	if got, want := LineNoColumnWidth(4), 10; got != want {
		t.Errorf("LineNoColumnWidth(4) = %d, want %d", got, want)
	}
	if got, want := LineNoColumnWidth(3), 8; got != want {
		t.Errorf("LineNoColumnWidth(3) = %d, want %d", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/ -run "TestCalcLineNoWidth|TestFormatLineNo|TestLineNoColumnWidth" -v
```

Expected: FAIL

- [ ] **Step 3: Implement lineno.go**

```go
// internal/ui/lineno.go
package ui

import (
	"fmt"
	"strings"
)

// CalcLineNoWidth returns the number of digits needed to display maxLineNo.
func CalcLineNoWidth(maxLineNo int) int {
	if maxLineNo <= 0 {
		return 1
	}
	w := 0
	for n := maxLineNo; n > 0; n /= 10 {
		w++
	}
	return w
}

// LineNoColumnWidth returns the total character width of the line number column.
// Format: "<left> <right> " → digitWidth*2 + 2 (one space separator + one trailing space).
func LineNoColumnWidth(digitWidth int) int {
	return digitWidth*2 + 2
}

// FormatLineNo formats a line number pair into a fixed-width string.
// A zero value means "no number" (blank). digitWidth is the width per number.
func FormatLineNo(left, right, digitWidth int) string {
	fmtStr := fmt.Sprintf("%%%dd", digitWidth)
	blank := strings.Repeat(" ", digitWidth)

	var l, r string
	if left > 0 {
		l = fmt.Sprintf(fmtStr, left)
	} else {
		l = blank
	}
	if right > 0 {
		r = fmt.Sprintf(fmtStr, right)
	} else {
		r = blank
	}
	return l + " " + r + " "
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/ -run "TestCalcLineNoWidth|TestFormatLineNo|TestLineNoColumnWidth" -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/lineno.go internal/ui/lineno_test.go
git commit -m "feat(ui): add line number formatting utilities"
```

---

### Task 7: viewport.go — Custom Viewport with Cursor and Scroll

**Files:**
- Create: `internal/ui/viewport.go`
- Create: `internal/ui/viewport_test.go`

- [ ] **Step 1: Write failing tests for data model and scrolling**

```go
// internal/ui/viewport_test.go
package ui

import (
	"testing"

	"github.com/kbliu/review/internal/diff"
)

func makeTestLines(n int) []ViewLine {
	lines := make([]ViewLine, n)
	for i := range lines {
		lines[i] = ViewLine{
			LeftNo:     i + 1,
			RightNo:    i + 1,
			Type:       diff.LineContext,
			Prefix:     " ",
			RawContent: fmt.Sprintf("line %d content", i+1),
		}
	}
	return lines
}

func TestViewportCursorDown(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(30))

	// Move cursor down 5 times
	for i := 0; i < 5; i++ {
		vp.CursorDown()
	}
	if vp.Cursor() != 5 {
		t.Errorf("cursor = %d, want 5", vp.Cursor())
	}
	// Should not scroll yet (cursor within viewport)
	if vp.Offset() != 0 {
		t.Errorf("offset = %d, want 0", vp.Offset())
	}
}

func TestViewportCursorDownScrolls(t *testing.T) {
	vp := NewViewport(80, 5)
	vp.SetLines(makeTestLines(20))

	// Move cursor to line 5 (index 4) — last visible
	for i := 0; i < 4; i++ {
		vp.CursorDown()
	}
	if vp.Offset() != 0 {
		t.Errorf("offset = %d, want 0 (cursor still visible)", vp.Offset())
	}

	// One more → cursor line 5, should scroll
	vp.CursorDown()
	if vp.Cursor() != 5 {
		t.Errorf("cursor = %d, want 5", vp.Cursor())
	}
	if vp.Offset() < 1 {
		t.Errorf("offset = %d, should have scrolled", vp.Offset())
	}
}

func TestViewportCursorUp(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(20))
	vp.CursorDown()
	vp.CursorDown()
	vp.CursorUp()
	if vp.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", vp.Cursor())
	}
}

func TestViewportCursorUpAtTop(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(20))
	vp.CursorUp()
	if vp.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", vp.Cursor())
	}
}

func TestViewportCursorDownAtBottom(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(5))
	for i := 0; i < 10; i++ {
		vp.CursorDown()
	}
	if vp.Cursor() != 4 {
		t.Errorf("cursor = %d, want 4 (last line)", vp.Cursor())
	}
}

func TestViewportHalfPageDown(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(30))

	vp.HalfPageDown()
	if vp.Cursor() != 5 {
		t.Errorf("cursor = %d, want 5 (half page = 5)", vp.Cursor())
	}
}

func TestViewportGotoTop(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(30))
	for i := 0; i < 15; i++ {
		vp.CursorDown()
	}
	vp.GotoTop()
	if vp.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", vp.Cursor())
	}
	if vp.Offset() != 0 {
		t.Errorf("offset = %d, want 0", vp.Offset())
	}
}

func TestViewportGotoBottom(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(makeTestLines(30))
	vp.GotoBottom()
	if vp.Cursor() != 29 {
		t.Errorf("cursor = %d, want 29", vp.Cursor())
	}
}

func TestViewportNextHunk(t *testing.T) {
	lines := []ViewLine{
		{Type: diff.LineHunkHeader, Prefix: "@@"},
		{Type: diff.LineContext, Prefix: " "},
		{Type: diff.LineAdded, Prefix: "+"},
		{Type: diff.LineHunkHeader, Prefix: "@@"},
		{Type: diff.LineRemoved, Prefix: "-"},
	}
	vp := NewViewport(80, 10)
	vp.SetLines(lines)

	vp.NextHunk()
	if vp.Cursor() != 3 {
		t.Errorf("cursor = %d, want 3 (second hunk header)", vp.Cursor())
	}
	// No more hunks, should stay
	vp.NextHunk()
	if vp.Cursor() != 3 {
		t.Errorf("cursor = %d, want 3 (no more hunks)", vp.Cursor())
	}
}

func TestViewportPrevHunk(t *testing.T) {
	lines := []ViewLine{
		{Type: diff.LineHunkHeader, Prefix: "@@"},
		{Type: diff.LineContext, Prefix: " "},
		{Type: diff.LineHunkHeader, Prefix: "@@"},
		{Type: diff.LineRemoved, Prefix: "-"},
	}
	vp := NewViewport(80, 10)
	vp.SetLines(lines)
	// Start at line 3
	vp.CursorDown()
	vp.CursorDown()
	vp.CursorDown()

	vp.PrevHunk()
	if vp.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2 (second hunk)", vp.Cursor())
	}
	vp.PrevHunk()
	if vp.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 (first hunk)", vp.Cursor())
	}
}

func TestViewportEmptyLines(t *testing.T) {
	vp := NewViewport(80, 10)
	vp.SetLines(nil)
	vp.CursorDown()
	vp.CursorUp()
	vp.HalfPageDown()
	vp.GotoBottom()
	// Should not panic
	if vp.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", vp.Cursor())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/ -run TestViewport -v
```

Expected: FAIL — `Viewport`, `ViewLine`, `NewViewport` not defined.

- [ ] **Step 3: Implement viewport.go — data model and scrolling**

```go
// internal/ui/viewport.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kbliu/review/internal/diff"
	"github.com/kbliu/review/internal/highlight"
)

// ViewLine is a display-ready diff line for the viewport.
type ViewLine struct {
	LeftNo      int
	RightNo     int
	Type        diff.LineType
	Prefix      string           // "+", "-", " ", or "@@..."
	RawContent  string           // plain text, no ANSI
	Tokens      []highlight.Token // syntax highlight tokens
	InlineSpans []InlineSpan     // character-level diff emphasis
}

// Viewport manages scrollable content with cursor tracking.
type Viewport struct {
	width  int
	height int
	offset int // first visible logical line
	cursor int // cursor logical line
	lines  []ViewLine
}

// NewViewport creates a viewport with the given dimensions.
func NewViewport(width, height int) *Viewport {
	return &Viewport{width: width, height: height}
}

// SetLines replaces the content and resets cursor/offset.
func (vp *Viewport) SetLines(lines []ViewLine) {
	vp.lines = lines
	vp.cursor = 0
	vp.offset = 0
}

// Resize updates viewport dimensions.
func (vp *Viewport) Resize(width, height int) {
	vp.width = width
	vp.height = height
	vp.ensureVisible()
}

// Cursor returns the current cursor position.
func (vp *Viewport) Cursor() int { return vp.cursor }

// Offset returns the scroll offset.
func (vp *Viewport) Offset() int { return vp.offset }

// Lines returns the viewport's lines.
func (vp *Viewport) Lines() []ViewLine { return vp.lines }

// LineCount returns the number of logical lines.
func (vp *Viewport) LineCount() int { return len(vp.lines) }

// CursorDown moves the cursor down one line.
func (vp *Viewport) CursorDown() {
	if len(vp.lines) == 0 {
		return
	}
	if vp.cursor < len(vp.lines)-1 {
		vp.cursor++
	}
	vp.ensureVisible()
}

// CursorUp moves the cursor up one line.
func (vp *Viewport) CursorUp() {
	if vp.cursor > 0 {
		vp.cursor--
	}
	vp.ensureVisible()
}

// HalfPageDown scrolls down half a page.
func (vp *Viewport) HalfPageDown() {
	if len(vp.lines) == 0 {
		return
	}
	half := vp.height / 2
	if half < 1 {
		half = 1
	}
	vp.cursor += half
	if vp.cursor >= len(vp.lines) {
		vp.cursor = len(vp.lines) - 1
	}
	vp.offset += half
	vp.clampOffset()
	vp.ensureVisible()
}

// HalfPageUp scrolls up half a page.
func (vp *Viewport) HalfPageUp() {
	half := vp.height / 2
	if half < 1 {
		half = 1
	}
	vp.cursor -= half
	if vp.cursor < 0 {
		vp.cursor = 0
	}
	vp.offset -= half
	if vp.offset < 0 {
		vp.offset = 0
	}
	vp.ensureVisible()
}

// PageDown scrolls down one full page.
func (vp *Viewport) PageDown() {
	if len(vp.lines) == 0 {
		return
	}
	vp.cursor += vp.height
	if vp.cursor >= len(vp.lines) {
		vp.cursor = len(vp.lines) - 1
	}
	vp.offset += vp.height
	vp.clampOffset()
	vp.ensureVisible()
}

// PageUp scrolls up one full page.
func (vp *Viewport) PageUp() {
	vp.cursor -= vp.height
	if vp.cursor < 0 {
		vp.cursor = 0
	}
	vp.offset -= vp.height
	if vp.offset < 0 {
		vp.offset = 0
	}
	vp.ensureVisible()
}

// GotoTop jumps to the first line.
func (vp *Viewport) GotoTop() {
	vp.cursor = 0
	vp.offset = 0
}

// GotoBottom jumps to the last line.
func (vp *Viewport) GotoBottom() {
	if len(vp.lines) == 0 {
		return
	}
	vp.cursor = len(vp.lines) - 1
	vp.ensureVisible()
}

// NextHunk moves cursor to the next hunk header after current position.
func (vp *Viewport) NextHunk() {
	for i := vp.cursor + 1; i < len(vp.lines); i++ {
		if vp.lines[i].Type == diff.LineHunkHeader {
			vp.cursor = i
			vp.ensureVisible()
			return
		}
	}
}

// PrevHunk moves cursor to the previous hunk header before current position.
func (vp *Viewport) PrevHunk() {
	for i := vp.cursor - 1; i >= 0; i-- {
		if vp.lines[i].Type == diff.LineHunkHeader {
			vp.cursor = i
			vp.ensureVisible()
			return
		}
	}
}

func (vp *Viewport) ensureVisible() {
	if vp.cursor < vp.offset {
		vp.offset = vp.cursor
	}
	if vp.height > 0 && vp.cursor >= vp.offset+vp.height {
		vp.offset = vp.cursor - vp.height + 1
	}
	vp.clampOffset()
}

func (vp *Viewport) clampOffset() {
	if len(vp.lines) == 0 {
		vp.offset = 0
		return
	}
	maxOffset := len(vp.lines) - vp.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if vp.offset > maxOffset {
		vp.offset = maxOffset
	}
	if vp.offset < 0 {
		vp.offset = 0
	}
}

// Render returns the visible portion of the viewport as a styled string.
// Handles content wrapping: if a line exceeds contentWidth, it produces
// continuation rows with blank line numbers and a "│" marker.
func (vp *Viewport) Render(theme Theme, digitWidth int) string {
	if len(vp.lines) == 0 || vp.height <= 0 || vp.width <= 0 {
		return ""
	}

	lineNoWidth := LineNoColumnWidth(digitWidth)
	sepWidth := 1 // "│"
	contentWidth := vp.width - lineNoWidth - sepWidth
	if contentWidth < 1 {
		contentWidth = 1
	}

	lineNoStyle := theme.LineNoStyle()
	sepStyle := theme.SepStyle()

	var rows []string
	displayLines := 0

	end := len(vp.lines)

	for i := vp.offset; i < end && displayLines < vp.height; i++ {
		line := vp.lines[i]
		isCursor := i == vp.cursor

		// Format line numbers — hunk headers show "··"
		var lineNoPart string
		if line.Type == diff.LineHunkHeader {
			dots := strings.Repeat("·", digitWidth)
			lineNoPart = lineNoStyle.Render(dots + " " + dots + " ")
		} else {
			lineNo := FormatLineNo(line.LeftNo, line.RightNo, digitWidth)
			lineNoPart = lineNoStyle.Render(lineNo)
		}
		sepPart := sepStyle.Render("│")

		// Determine content background
		bgColor := vp.lineBackground(line.Type, isCursor, theme)

		// Render the full content line
		fullContent := vp.renderContent(line, contentWidth, bgColor, theme)
		visibleLen := lipgloss.Width(fullContent)

		if visibleLen <= contentWidth+1 { // +1 for prefix char
			rows = append(rows, lineNoPart+sepPart+fullContent)
			displayLines++
		} else {
			// Content wrapping: split into multiple display rows
			// First row gets real line numbers, continuation rows get blank + "│"
			wrappedRows := wrapRenderedLine(fullContent, contentWidth+1)
			blankLineNo := lineNoStyle.Render(strings.Repeat(" ", lineNoWidth))
			contSep := sepStyle.Render("│")

			for wi, wr := range wrappedRows {
				if displayLines >= vp.height {
					break
				}
				if wi == 0 {
					rows = append(rows, lineNoPart+sepPart+wr)
				} else {
					rows = append(rows, blankLineNo+contSep+wr)
				}
				displayLines++
			}
		}
	}

	// Pad remaining height with empty rows
	for displayLines < vp.height {
		emptyLineNo := lineNoStyle.Render(strings.Repeat(" ", lineNoWidth))
		emptySep := sepStyle.Render("│")
		emptyContent := strings.Repeat(" ", contentWidth)
		rows = append(rows, emptyLineNo+emptySep+emptyContent)
		displayLines++
	}

	return strings.Join(rows, "\n")
}

// wrapRenderedLine splits a rendered (ANSI-containing) line into chunks of
// maxVisibleWidth visible characters. Uses ANSI-aware width calculation.
func wrapRenderedLine(rendered string, maxVisibleWidth int) []string {
	if maxVisibleWidth <= 0 {
		return []string{rendered}
	}
	total := lipgloss.Width(rendered)
	if total <= maxVisibleWidth {
		return []string{rendered}
	}

	var result []string
	remaining := rendered
	for lipgloss.Width(remaining) > maxVisibleWidth {
		bp := findANSISafeBreak(remaining, maxVisibleWidth)
		result = append(result, remaining[:bp])
		remaining = remaining[bp:]
	}
	if len(remaining) > 0 {
		result = append(result, remaining)
	}
	return result
}

// findANSISafeBreak returns a byte index in s where visible character count
// reaches maxWidth, correctly skipping ANSI escape sequences.
func findANSISafeBreak(s string, maxWidth int) int {
	visible := 0
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		visible++
		if visible >= maxWidth {
			return i + 1
		}
	}
	return len(s)
}

func (vp *Viewport) lineBackground(lt diff.LineType, isCursor bool, th Theme) string {
	if isCursor {
		switch lt {
		case diff.LineAdded:
			return th.AddedCursorBg
		case diff.LineRemoved:
			return th.RemovedCursorBg
		default:
			return th.CursorBg
		}
	}
	switch lt {
	case diff.LineAdded:
		return th.AddedBg
	case diff.LineRemoved:
		return th.RemovedBg
	default:
		return ""
	}
}

func (vp *Viewport) renderContent(line ViewLine, width int, bgColor string, th Theme) string {
	if line.Type == diff.LineHunkHeader {
		hunkText := line.Prefix
		if len(hunkText) > width {
			hunkText = hunkText[:width]
		}
		padded := hunkText + strings.Repeat(" ", max(0, width-lipgloss.Width(hunkText)))
		return th.HunkStyle().Render(padded)
	}

	// Build the content from prefix + raw content
	prefix := line.Prefix
	raw := line.RawContent

	// Build a styled string segment by segment
	var result strings.Builder

	// Render prefix character with line background
	bgStyle := lipgloss.NewStyle()
	if bgColor != "" {
		bgStyle = bgStyle.Background(lipgloss.Color(bgColor))
	}
	result.WriteString(bgStyle.Render(prefix))

	// Build inline span lookup for fast checking
	inlineSet := make(map[int]bool)
	for _, span := range line.InlineSpans {
		for b := span.Start; b < span.End && b < len(raw); b++ {
			inlineSet[b] = true
		}
	}

	// Determine emphasis background for inline diff
	var emphBg string
	switch line.Type {
	case diff.LineAdded:
		emphBg = th.InlineAddBg
	case diff.LineRemoved:
		emphBg = th.InlineDelBg
	}

	// Render content using tokens if available, otherwise raw
	if len(line.Tokens) > 0 {
		byteOffset := 0
		for _, tok := range line.Tokens {
			for _, r := range tok.Text {
				rs := string(r)
				rLen := len(rs)
				style := lipgloss.NewStyle()
				// Foreground from syntax highlighting
				color := getTokenColor(tok.TokenType)
				if color != "" {
					style = style.Foreground(lipgloss.Color(color))
				}
				// Background: inline emphasis or line background
				if inlineSet[byteOffset] && emphBg != "" {
					style = style.Background(lipgloss.Color(emphBg))
				} else if bgColor != "" {
					style = style.Background(lipgloss.Color(bgColor))
				}
				result.WriteString(style.Render(rs))
				byteOffset += rLen
			}
		}
	} else {
		for i, r := range raw {
			rs := string(r)
			style := lipgloss.NewStyle()
			if bgColor != "" {
				style = style.Background(lipgloss.Color(bgColor))
			}
			if inlineSet[i] && emphBg != "" {
				style = style.Background(lipgloss.Color(emphBg))
			}
			result.WriteString(style.Render(rs))
		}
	}

	// Pad to fill content width
	rendered := result.String()
	visibleWidth := lipgloss.Width(rendered)
	if visibleWidth < width+1 { // +1 for prefix
		padLen := width + 1 - visibleWidth
		padStyle := lipgloss.NewStyle()
		if bgColor != "" {
			padStyle = padStyle.Background(lipgloss.Color(bgColor))
		}
		rendered += padStyle.Render(strings.Repeat(" ", padLen))
	}

	return rendered
}

func getTokenColor(tokenType string) string {
	switch {
	case strings.Contains(tokenType, "Keyword"):
		return "204"
	case strings.Contains(tokenType, "String"):
		return "192"
	case strings.Contains(tokenType, "Comment"):
		return "243"
	case strings.Contains(tokenType, "Number"):
		return "180"
	case strings.Contains(tokenType, "Function"):
		return "117"
	default:
		return ""
	}
}

```

Note: Go 1.21+ provides built-in `max`/`min`, do not define local versions.
Note: add `"fmt"` to the imports in `viewport_test.go`.

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/ -run TestViewport -v
```

Expected: all PASS

- [ ] **Step 5: Write additional rendering test**

```go
// Add to internal/ui/viewport_test.go

func TestViewportRenderBasic(t *testing.T) {
	vp := NewViewport(40, 3)
	vp.SetLines([]ViewLine{
		{LeftNo: 1, RightNo: 1, Type: diff.LineContext, Prefix: " ", RawContent: "hello"},
		{LeftNo: 0, RightNo: 2, Type: diff.LineAdded, Prefix: "+", RawContent: "world"},
		{LeftNo: 2, RightNo: 0, Type: diff.LineRemoved, Prefix: "-", RawContent: "old"},
	})

	th := DefaultTheme()
	output := vp.Render(th, 2)
	if output == "" {
		t.Fatal("Render returned empty")
	}
	// Should contain line numbers and content
	lines := strings.Split(output, "\n")
	if len(lines) != 3 {
		t.Errorf("rendered %d lines, want 3", len(lines))
	}
}

func TestViewportRenderWrapping(t *testing.T) {
	// A 30-char wide viewport with long content should produce wrapped rows
	vp := NewViewport(20, 5)
	vp.SetLines([]ViewLine{
		{LeftNo: 1, RightNo: 1, Type: diff.LineContext, Prefix: " ",
			RawContent: "this is a very long line that should wrap"},
	})

	th := DefaultTheme()
	output := vp.Render(th, 2)
	lines := strings.Split(output, "\n")
	// Should have more than 1 display row due to wrapping + padding rows
	if len(lines) != 5 {
		t.Errorf("rendered %d display lines, want 5 (wrapping + padding)", len(lines))
	}
}

func TestViewportRenderEmpty(t *testing.T) {
	vp := NewViewport(40, 5)
	th := DefaultTheme()
	output := vp.Render(th, 4)
	if output != "" {
		t.Errorf("empty viewport should render empty, got %q", output)
	}
}
```

- [ ] **Step 6: Run all viewport tests**

```bash
go test ./internal/ui/ -run TestViewport -v
```

Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/ui/viewport.go internal/ui/viewport_test.go
git commit -m "feat(ui): add custom viewport with cursor tracking and rendering"
```

---

### Task 8: filelist.go — File List Component

**Files:**
- Create: `internal/ui/filelist.go`
- Create: `internal/ui/filelist_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/ui/filelist_test.go
package ui

import (
	"strings"
	"testing"
)

func TestFileListCursorDown(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "a.go"},
		{Status: "A", Name: "b.go"},
		{Status: "D", Name: "c.go"},
	})
	fl.CursorDown()
	if fl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", fl.Cursor())
	}
}

func TestFileListCursorDownAtBottom(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "a.go"},
		{Status: "A", Name: "b.go"},
	})
	fl.CursorDown()
	fl.CursorDown()
	fl.CursorDown()
	if fl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 (clamped)", fl.Cursor())
	}
}

func TestFileListCursorUpAtTop(t *testing.T) {
	fl := NewFileList([]FileInfo{{Status: "M", Name: "a.go"}})
	fl.CursorUp()
	if fl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", fl.Cursor())
	}
}

func TestFileListGotoTop(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "a.go"},
		{Status: "A", Name: "b.go"},
		{Status: "D", Name: "c.go"},
	})
	fl.CursorDown()
	fl.CursorDown()
	fl.GotoTop()
	if fl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", fl.Cursor())
	}
}

func TestFileListGotoBottom(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "a.go"},
		{Status: "A", Name: "b.go"},
	})
	fl.GotoBottom()
	if fl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", fl.Cursor())
	}
}

func TestFileListSelectedFile(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "a.go"},
		{Status: "A", Name: "b.go"},
	})
	fl.CursorDown()
	f := fl.SelectedFile()
	if f.Name != "b.go" {
		t.Errorf("selected = %q, want b.go", f.Name)
	}
}

func TestFileListRender(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "src/main.go"},
		{Status: "A", Name: "README.md"},
	})
	th := DefaultTheme()
	output := fl.Render(30, 10, th)

	if !strings.Contains(output, "src/main.go") {
		t.Error("render missing file name src/main.go")
	}
	if !strings.Contains(output, "README.md") {
		t.Error("render missing file name README.md")
	}
}

func TestFileListCalcWidth(t *testing.T) {
	fl := NewFileList([]FileInfo{
		{Status: "M", Name: "short.go"},
		{Status: "A", Name: "very/long/path/to/file.go"},
	})
	w := fl.CalcWidth()
	// Should be longest path + status badge + padding
	if w < len("very/long/path/to/file.go")+4 {
		t.Errorf("width = %d, too narrow", w)
	}
	if w > 50 {
		t.Errorf("width = %d, should cap at 50", w)
	}
}

func TestFileListEmpty(t *testing.T) {
	fl := NewFileList(nil)
	if fl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", fl.Cursor())
	}
	f := fl.SelectedFile()
	if f.Name != "" {
		t.Errorf("selected = %q, want empty", f.Name)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/ -run TestFileList -v
```

Expected: FAIL

- [ ] **Step 3: Implement filelist.go**

```go
// internal/ui/filelist.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FileList manages the left panel file list.
type FileList struct {
	files  []FileInfo
	cursor int
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

// GotoTop moves cursor to first item.
func (fl *FileList) GotoTop() {
	fl.cursor = 0
}

// GotoBottom moves cursor to last item.
func (fl *FileList) GotoBottom() {
	if len(fl.files) > 0 {
		fl.cursor = len(fl.files) - 1
	}
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
		// "M " + filename + padding
		lineW := len(f.Status) + 1 + len(f.Name) + 2
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
func (fl *FileList) Render(width, height int, theme Theme) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	var rows []string
	for i, f := range fl.files {
		if i >= height {
			break
		}
		statusColor := theme.StatusColor(f.Status)
		badge := lipgloss.NewStyle().
			Foreground(lipgloss.Color(statusColor)).
			Bold(true).
			Render(f.Status)

		name := f.Name
		maxNameW := width - 4 // badge + spaces
		if maxNameW > 0 && len(name) > maxNameW {
			name = name[:maxNameW]
		}

		line := fmt.Sprintf(" %s %s", badge, name)

		if i == fl.cursor {
			line = theme.FileSelectedStyle().Width(width).Render(
				fmt.Sprintf(" %s %s", f.Status, name))
		}

		rows = append(rows, line)
	}

	// Pad remaining height
	for len(rows) < height {
		rows = append(rows, strings.Repeat(" ", width))
	}

	return strings.Join(rows, "\n")
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/ -run TestFileList -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/filelist.go internal/ui/filelist_test.go
git commit -m "feat(ui): add file list component with cursor navigation"
```

---

### Task 9: statusbar.go — Status Bar Rendering

**Files:**
- Create: `internal/ui/statusbar.go`
- Create: `internal/ui/statusbar_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/ui/statusbar_test.go
package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderStatusBar(t *testing.T) {
	th := DefaultTheme()
	bar := RenderStatusBar("main", 5, "src/utils.go", 15, 3, 80, th)

	if !strings.Contains(bar, "main") {
		t.Error("status bar missing branch name")
	}
	if !strings.Contains(bar, "5") {
		t.Error("status bar missing file count")
	}
	if !strings.Contains(bar, "src/utils.go") {
		t.Error("status bar missing file name")
	}
}

func TestRenderStatusBarWidth(t *testing.T) {
	th := DefaultTheme()
	for _, w := range []int{40, 80, 120} {
		bar := RenderStatusBar("main", 3, "a.go", 10, 2, w, th)
		got := lipgloss.Width(bar)
		if got != w {
			t.Errorf("width %d: bar width = %d, want %d", w, got, w)
		}
	}
}

func TestRenderStatusBarNarrow(t *testing.T) {
	th := DefaultTheme()
	// Should not panic at narrow widths
	for _, w := range []int{0, 5, 10} {
		_ = RenderStatusBar("main", 3, "a.go", 10, 2, w, th)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/ -run TestRenderStatusBar -v
```

Expected: FAIL

- [ ] **Step 3: Implement statusbar.go**

```go
// internal/ui/statusbar.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/ -run TestRenderStatusBar -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/statusbar.go internal/ui/statusbar_test.go
git commit -m "feat(ui): add status bar rendering"
```

---

### Task 10: diffview.go — Diff View Data Pipeline

**Files:**
- Create: `internal/ui/diffview.go`
- Create: `internal/ui/diffview_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/ui/diffview_test.go
package ui

import (
	"testing"

	"github.com/kbliu/review/internal/diff"
	"github.com/kbliu/review/internal/highlight"
)

func TestBuildViewLinesBasic(t *testing.T) {
	diffLines := []diff.Line{
		{Type: diff.LineHunkHeader, Content: "@@ -1,3 +1,4 @@"},
		{Type: diff.LineContext, OldLineNo: 1, NewLineNo: 1, Content: " hello"},
		{Type: diff.LineRemoved, OldLineNo: 2, Content: "-old"},
		{Type: diff.LineAdded, NewLineNo: 2, Content: "+new"},
		{Type: diff.LineContext, OldLineNo: 3, NewLineNo: 3, Content: " end"},
	}

	hl := highlight.New("github")
	viewLines := BuildViewLines(diffLines, "main.go", hl)

	if len(viewLines) != 5 {
		t.Fatalf("viewLines count = %d, want 5", len(viewLines))
	}

	// Check hunk header
	if viewLines[0].Type != diff.LineHunkHeader {
		t.Errorf("line 0 type = %v, want HunkHeader", viewLines[0].Type)
	}
	if viewLines[0].Prefix != "@@ -1,3 +1,4 @@" {
		t.Errorf("hunk prefix = %q", viewLines[0].Prefix)
	}

	// Check context line
	if viewLines[1].LeftNo != 1 || viewLines[1].RightNo != 1 {
		t.Errorf("context line nos = %d,%d, want 1,1", viewLines[1].LeftNo, viewLines[1].RightNo)
	}
	if viewLines[1].RawContent != "hello" {
		t.Errorf("context content = %q, want 'hello'", viewLines[1].RawContent)
	}

	// Check removed line
	if viewLines[2].LeftNo != 2 || viewLines[2].RightNo != 0 {
		t.Errorf("removed line nos = %d,%d, want 2,0", viewLines[2].LeftNo, viewLines[2].RightNo)
	}
	if viewLines[2].Prefix != "-" {
		t.Errorf("removed prefix = %q, want '-'", viewLines[2].Prefix)
	}

	// Check added line
	if viewLines[3].LeftNo != 0 || viewLines[3].RightNo != 2 {
		t.Errorf("added line nos = %d,%d, want 0,2", viewLines[3].LeftNo, viewLines[3].RightNo)
	}
}

func TestBuildViewLinesWithInlineDiff(t *testing.T) {
	diffLines := []diff.Line{
		{Type: diff.LineRemoved, OldLineNo: 1, Content: "-println(\"hello\")"},
		{Type: diff.LineAdded, NewLineNo: 1, Content: "+fmt.Println(\"hello\")"},
	}

	hl := highlight.New("github")
	viewLines := BuildViewLines(diffLines, "main.go", hl)

	// Both lines should have inline spans
	if len(viewLines[0].InlineSpans) == 0 {
		t.Error("removed line should have inline spans")
	}
	if len(viewLines[1].InlineSpans) == 0 {
		t.Error("added line should have inline spans")
	}
}

func TestBuildViewLinesSyntaxTokens(t *testing.T) {
	diffLines := []diff.Line{
		{Type: diff.LineContext, OldLineNo: 1, NewLineNo: 1, Content: " func main() {}"},
	}

	hl := highlight.New("github")
	viewLines := BuildViewLines(diffLines, "main.go", hl)

	if len(viewLines[0].Tokens) == 0 {
		t.Error("context line should have syntax tokens")
	}
}

func TestCalcMaxLineNo(t *testing.T) {
	lines := []diff.Line{
		{OldLineNo: 10, NewLineNo: 15},
		{OldLineNo: 0, NewLineNo: 200},
		{OldLineNo: 50, NewLineNo: 0},
	}
	got := CalcMaxLineNo(lines)
	if got != 200 {
		t.Errorf("CalcMaxLineNo = %d, want 200", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/ -run "TestBuildViewLines|TestCalcMaxLineNo" -v
```

Expected: FAIL

- [ ] **Step 3: Implement diffview.go**

```go
// internal/ui/diffview.go
package ui

import (
	"github.com/kbliu/review/internal/diff"
	"github.com/kbliu/review/internal/highlight"
)

// DiffView manages the right panel diff display.
type DiffView struct {
	viewport   *Viewport
	theme      Theme
	digitWidth int
}

// NewDiffView creates a new diff view.
func NewDiffView(width, height int, theme Theme) *DiffView {
	return &DiffView{
		viewport: NewViewport(width, height),
		theme:    theme,
	}
}

// Viewport returns the underlying viewport.
func (dv *DiffView) Viewport() *Viewport { return dv.viewport }

// DigitWidth returns the current line number digit width.
func (dv *DiffView) DigitWidth() int { return dv.digitWidth }

// LoadFile parses diff lines, applies syntax highlighting and inline diff,
// and sets the viewport content.
func (dv *DiffView) LoadFile(diffLines []diff.Line, filename string, hl *highlight.SimpleHighlighter) {
	dv.digitWidth = CalcLineNoWidth(CalcMaxLineNo(diffLines))
	viewLines := BuildViewLines(diffLines, filename, hl)
	dv.viewport.SetLines(viewLines)
}

// Render returns the rendered diff view.
func (dv *DiffView) Render() string {
	return dv.viewport.Render(dv.theme, dv.digitWidth)
}

// Resize updates dimensions.
func (dv *DiffView) Resize(width, height int) {
	dv.viewport.Resize(width, height)
}

// CalcMaxLineNo finds the highest line number across all diff lines.
func CalcMaxLineNo(lines []diff.Line) int {
	maxNo := 0
	for _, l := range lines {
		if l.OldLineNo > maxNo {
			maxNo = l.OldLineNo
		}
		if l.NewLineNo > maxNo {
			maxNo = l.NewLineNo
		}
	}
	return maxNo
}

// BuildViewLines converts parsed diff lines into display-ready ViewLines.
// It applies syntax highlighting (via batch TokenizeFile) and inline diff.
func BuildViewLines(lines []diff.Line, filename string, hl *highlight.SimpleHighlighter) []ViewLine {
	if len(lines) == 0 {
		return nil
	}

	// 1. Extract code content for syntax highlighting (strip prefix)
	codeLines := make([]string, len(lines))
	for i, l := range lines {
		if l.Type == diff.LineHunkHeader {
			codeLines[i] = ""
		} else if len(l.Content) > 1 {
			codeLines[i] = l.Content[1:] // strip +/-/space prefix
		}
	}

	// 2. Batch tokenize
	var tokensByLine [][]highlight.Token
	if hl != nil {
		tokensByLine = hl.TokenizeFile(filename, codeLines)
	}

	// 3. Compute inline diff pairs
	pairs := PairDiffLines(lines)
	// Build a map from line index → inline spans
	inlineMap := make(map[int][]InlineSpan)
	for _, pair := range pairs {
		oldContent := codeLines[pair.OldIdx]
		newContent := codeLines[pair.NewIdx]
		oldSpans, newSpans := ComputeInlineDiff(oldContent, newContent)
		if len(oldSpans) > 0 {
			inlineMap[pair.OldIdx] = oldSpans
		}
		if len(newSpans) > 0 {
			inlineMap[pair.NewIdx] = newSpans
		}
	}

	// 4. Build ViewLines
	viewLines := make([]ViewLine, len(lines))
	for i, l := range lines {
		vl := ViewLine{
			LeftNo:  l.OldLineNo,
			RightNo: l.NewLineNo,
			Type:    l.Type,
		}

		if l.Type == diff.LineHunkHeader {
			vl.Prefix = l.Content
			vl.RawContent = ""
		} else {
			if len(l.Content) > 0 {
				vl.Prefix = l.Content[:1]
				vl.RawContent = codeLines[i]
			}
		}

		if tokensByLine != nil && i < len(tokensByLine) {
			vl.Tokens = tokensByLine[i]
		}

		if spans, ok := inlineMap[i]; ok {
			vl.InlineSpans = spans
		}

		viewLines[i] = vl
	}

	return viewLines
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/ -run "TestBuildViewLines|TestCalcMaxLineNo" -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/diffview.go internal/ui/diffview_test.go
git commit -m "feat(ui): add diff view with highlight + inline diff pipeline"
```

---

### Task 11: app.go — Full Model Wiring Everything Together

**Files:**
- Modify: `internal/ui/app.go`
- Create: `internal/ui/app_test.go`

- [ ] **Step 1: Write failing tests**

```go
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
	// cmd should be tea.Quit
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/ -run "TestNewModel|TestActionDispatch|TestWindowSizeMsg|TestHelpToggle" -v
```

Expected: FAIL — the stub app.go doesn't have fileList, diffView, keyMapper, etc.

- [ ] **Step 3: Rewrite app.go with full implementation**

Replace `internal/ui/app.go` with:

```go
// internal/ui/app.go
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kbliu/review/internal/diff"
	"github.com/kbliu/review/internal/git"
	"github.com/kbliu/review/internal/highlight"
)

// FileInfo represents a file with diff statistics.
type FileInfo struct {
	Status  string
	Name    string
	Added   int
	Removed int
}

// FocusType indicates which panel has focus.
type FocusType int

const (
	FocusList FocusType = iota
	FocusDiff
)

// Model is the top-level BubbleTea model.
type Model struct {
	opts   Options
	theme  Theme
	keys   *KeyMapper
	width  int
	height int

	// Components
	fileList *FileList
	diffView *DiffView

	// State
	files       []FileInfo
	focus       FocusType
	listWidth   int
	showHelp    bool
	loading     bool
	err         error
	currentFile string
	highlighter *highlight.SimpleHighlighter
}

// NewModel creates a new Model.
func NewModel(opts Options) Model {
	theme := DefaultTheme()
	return Model{
		opts:        opts,
		theme:       theme,
		keys:        NewKeyMapper(),
		fileList:    NewFileList(nil),
		diffView:    NewDiffView(80, 24, theme),
		focus:       FocusList,
		loading:     true,
		listWidth:   30,
		highlighter: highlight.New("github"),
	}
}

// --- BubbleTea messages ---

type loadFilesMsg struct {
	files []FileInfo
	err   error
}

type loadDiffMsg struct {
	file  string
	lines []diff.Line
	err   error
}

// --- Init ---

func (m Model) Init() tea.Cmd {
	return loadFilesCmd(m.opts)
}

func loadFilesCmd(opts Options) tea.Cmd {
	return func() tea.Msg {
		files, err := loadFileList(opts)
		return loadFilesMsg{files: files, err: err}
	}
}

func loadDiffCmd(opts Options, file string) tea.Cmd {
	return func() tea.Msg {
		gopts := gitOptions(opts)
		content, err := git.GetDiff(gopts, file)
		if err != nil {
			return loadDiffMsg{file: file, err: err}
		}
		lines := diff.Parse(content)
		return loadDiffMsg{file: file, lines: lines}
	}
}

// --- Update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeComponents()
		return m, nil

	case loadFilesMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.files = msg.files
		m.fileList.SetFiles(m.files)
		m.listWidth = m.fileList.CalcWidth()
		m.resizeComponents()
		if len(m.files) > 0 {
			return m, loadDiffCmd(m.opts, m.files[0].Name)
		}
		return m, nil

	case loadDiffMsg:
		m.currentFile = msg.file
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.diffView.LoadFile(msg.lines, msg.file, m.highlighter)
		return m, nil

	case tea.KeyMsg:
		// If help is shown, any key closes it
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		return m.handleAction(m.keys.HandleKey(msg, m.focus))
	}
	return m, nil
}

func (m Model) handleAction(action Action) (tea.Model, tea.Cmd) {
	switch action {
	case ActionNone:
		return m, nil
	case ActionQuit:
		return m, tea.Quit
	case ActionHelp:
		m.showHelp = true
		return m, nil

	// Focus
	case ActionFocusLeft:
		m.focus = FocusList
		return m, nil
	case ActionFocusRight:
		m.focus = FocusDiff
		return m, nil
	case ActionEnter:
		if m.focus == FocusList {
			m.focus = FocusDiff
		}
		return m, nil

	// Panel resize
	case ActionGrowPanel:
		if m.listWidth < 60 {
			m.listWidth += 2
			m.resizeComponents()
		}
		return m, nil
	case ActionShrinkPanel:
		if m.listWidth > 10 {
			m.listWidth -= 2
			m.resizeComponents()
		}
		return m, nil

	// Cursor movement
	case ActionCursorDown:
		if m.focus == FocusList {
			prevCursor := m.fileList.Cursor()
			m.fileList.CursorDown()
			if m.fileList.Cursor() != prevCursor {
				return m, loadDiffCmd(m.opts, m.fileList.SelectedFile().Name)
			}
		} else {
			m.diffView.Viewport().CursorDown()
		}
		return m, nil

	case ActionCursorUp:
		if m.focus == FocusList {
			prevCursor := m.fileList.Cursor()
			m.fileList.CursorUp()
			if m.fileList.Cursor() != prevCursor {
				return m, loadDiffCmd(m.opts, m.fileList.SelectedFile().Name)
			}
		} else {
			m.diffView.Viewport().CursorUp()
		}
		return m, nil

	case ActionTop:
		if m.focus == FocusList {
			m.fileList.GotoTop()
			if len(m.files) > 0 {
				return m, loadDiffCmd(m.opts, m.fileList.SelectedFile().Name)
			}
		} else {
			m.diffView.Viewport().GotoTop()
		}
		return m, nil

	case ActionBottom:
		if m.focus == FocusList {
			m.fileList.GotoBottom()
			if len(m.files) > 0 {
				return m, loadDiffCmd(m.opts, m.fileList.SelectedFile().Name)
			}
		} else {
			m.diffView.Viewport().GotoBottom()
		}
		return m, nil

	case ActionHalfPageDown:
		m.diffView.Viewport().HalfPageDown()
		return m, nil
	case ActionHalfPageUp:
		m.diffView.Viewport().HalfPageUp()
		return m, nil
	case ActionPageDown:
		m.diffView.Viewport().PageDown()
		return m, nil
	case ActionPageUp:
		m.diffView.Viewport().PageUp()
		return m, nil
	case ActionNextHunk:
		m.diffView.Viewport().NextHunk()
		return m, nil
	case ActionPrevHunk:
		m.diffView.Viewport().PrevHunk()
		return m, nil
	}

	return m, nil
}

func (m *Model) resizeComponents() {
	contentHeight := m.height - 1 // 1 for status bar
	if contentHeight < 0 {
		contentHeight = 0
	}
	diffWidth := m.width - m.listWidth
	if diffWidth < 1 {
		diffWidth = 1
	}
	m.diffView.Resize(diffWidth, contentHeight)
}

// --- View ---

func (m Model) View() string {
	if m.loading && len(m.files) == 0 {
		return "Loading..."
	}
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit."
	}
	if len(m.files) == 0 {
		return "No changes found.\n\nPress q to quit."
	}
	if m.showHelp {
		return m.renderHelp()
	}

	contentHeight := m.height - 1
	if contentHeight < 0 {
		contentHeight = 0
	}

	// File list
	listView := m.fileList.Render(m.listWidth, contentHeight, m.theme)

	// Diff view
	diffContent := m.diffView.Render()

	// Compose horizontally
	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, diffContent)

	// Status bar
	selected := m.fileList.SelectedFile()
	bar := RenderStatusBar(
		m.opts.Target, len(m.files),
		m.currentFile, selected.Added, selected.Removed,
		m.width, m.theme,
	)

	return body + "\n" + bar
}

func (m Model) renderHelp() string {
	help := `Keyboard Shortcuts

  j/k          Navigate files / Move cursor in diff
  Enter        Switch focus to diff view
  Ctrl+W h/l   Switch focus between panels
  Ctrl+W >/<   Adjust panel width
  gg / G       Go to top / bottom
  Ctrl+D/U     Half page down / up
  Ctrl+F/B     Page down / up
  n / N        Next / previous hunk
  ?            Toggle this help
  q            Quit

Press any key to close...`

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		style.Render(help))
}

// --- Data bridge ---

func gitOptions(opts Options) git.Options {
	return git.Options{
		Target:       opts.Target,
		Staged:       opts.Staged,
		ContextLines: opts.ContextLines,
	}
}

func loadFileList(opts Options) ([]FileInfo, error) {
	gopts := gitOptions(opts)
	files, err := git.GetFiles(gopts)
	if err != nil {
		return nil, err
	}
	result := make([]FileInfo, len(files))
	for i, f := range files {
		content, err := git.GetDiff(gopts, f.Name)
		if err != nil {
			result[i] = FileInfo{Status: f.Status, Name: f.Name}
			continue
		}
		lines := diff.Parse(content)
		stats := diff.CalculateStats(lines)
		result[i] = FileInfo{
			Status:  f.Status,
			Name:    f.Name,
			Added:   stats.Added,
			Removed: stats.Removed,
		}
	}
	return result, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/ -run "TestNewModel|TestActionDispatch|TestWindowSizeMsg|TestHelpToggle" -v
```

Expected: all PASS

- [ ] **Step 5: Run full test suite**

```bash
go test ./... -v
```

Expected: all packages PASS

- [ ] **Step 6: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat(ui): wire up full app model with keymap, file list, and diff view"
```

---

### Task 12: Integration Verification

**Files:**
- No new files. Build and test everything.

- [ ] **Step 1: Full build**

```bash
cd /Users/kbliu/Workspace/project/vim-code-review
go build ./...
```

Expected: clean build, no errors.

- [ ] **Step 2: Full test suite**

```bash
go test ./... -v -count=1
```

Expected: all tests pass across all packages.

- [ ] **Step 3: Run the binary manually**

```bash
cd /Users/kbliu/Workspace/project/vim-code-review
go run ./cmd/review/ HEAD
```

Expected: TUI launches with file list on left, diff on right. Verify:
- `j`/`k` moves through file list and loads diffs
- `Enter` switches focus to diff
- `Ctrl+W h` returns focus to file list
- `gg` and `G` work in both panels
- `Ctrl+D`/`Ctrl+U` scroll in diff
- `n`/`N` jump between hunks
- `?` shows help overlay
- `q` quits
- Line numbers are right-aligned, dynamic width
- Added lines have green background, removed lines have red background
- Inline diff shows darker emphasis on changed characters
- No all-caps text, no brand labels

- [ ] **Step 4: Fix any issues found during manual testing**

Iterate on any visual or behavioral issues. Common fixes:
- Padding/width calculations off by one
- Colors not rendering (check terminal supports 256-color)
- Viewport height miscalculated

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "feat(ui): complete UI layer rewrite with Vim keybindings and clean layout"
```
