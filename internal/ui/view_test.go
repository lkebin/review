package ui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestMain(m *testing.M) {
	// Force ANSI color output so tests can assert on escape codes.
	lipgloss.SetColorProfile(termenv.TrueColor)
	os.Exit(m.Run())
}

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
	m.diffViewport.Width = m.getDiffWidth()
	m.diffViewport.Height = 24

	result := m.renderDiff()

	if result == "" {
		t.Fatal("renderDiff returned empty string")
	}
	// ANSI escape codes must be present (from bg styles on added/removed lines)
	if !strings.Contains(result, "\x1b[") {
		t.Error("expected ANSI escape codes in renderDiff output, got none")
	}
}

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

func TestContentHeightReservesHeaderAndFooter(t *testing.T) {
	m := newTestModel()
	m.height = 40

	if got := m.getContentHeight(); got != 38 {
		t.Fatalf("getContentHeight = %d, want 38", got)
	}
}

func TestRenderTopHeaderWidth(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.currentFile = "src/core/renderer.rs"

	header := m.renderTopHeader()

	if got := lipgloss.Width(header); got != m.width {
		t.Fatalf("top header width = %d, want %d", got, m.width)
	}
	if !strings.Contains(header, "CYBERNETIC_MANUSCRIPT") {
		t.Fatalf("top header missing brand: %q", header)
	}
	if !strings.Contains(header, "SRC/CORE/RENDERER.RS") {
		t.Fatalf("top header missing current path: %q", header)
	}
}

func TestRenderTopHeaderNarrow(t *testing.T) {
	longPath := "src/components/extremely/long/path/that/should/be/truncated/in/the/header/view.go"

	for _, w := range []int{0, 5, 10, 20} {
		m := newTestModel()
		m.width = w
		m.currentFile = longPath
		_ = m.renderTopHeader()
	}

	for _, w := range []int{80, 120} {
		m := newTestModel()
		m.width = w
		m.currentFile = longPath
		header := m.renderTopHeader()
		if got := lipgloss.Width(header); got != w {
			t.Errorf("width=%d: top header width = %d, want %d", w, got, w)
		}
		if strings.Contains(header, longPath) {
			t.Errorf("width=%d: top header should uppercase/truncate rendered path", w)
		}
		if !strings.Contains(header, "SRC/COMPONENTS") {
			t.Errorf("width=%d: top header missing uppercase path prefix: %q", w, header)
		}
	}

	m := newTestModel()
	m.width = 80
	header := m.renderTopHeader()
	if !strings.Contains(header, "NO_FILE_SELECTED") {
		t.Fatalf("top header missing fallback path: %q", header)
	}
}

func TestRenderBottomBarWidth(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.files = []FileInfo{{Name: "src/main.rs"}}
	m.cursor = 0

	bar := m.renderBottomBar()

	if got := lipgloss.Width(bar); got != m.width {
		t.Fatalf("bottom bar width = %d, want %d", got, m.width)
	}
	if !strings.Contains(bar, "FILES: 1") {
		t.Fatalf("bottom bar missing file count: %q", bar)
	}
	if !strings.Contains(bar, "[Q]UIT") {
		t.Fatalf("bottom bar missing quit hint: %q", bar)
	}
}

func TestRenderBottomBarNarrow(t *testing.T) {
	for _, w := range []int{0, 5, 10, 20} {
		m := newTestModel()
		m.width = w
		m.files = []FileInfo{{Name: "a.go"}}
		_ = m.renderBottomBar()
	}

	for _, w := range []int{80, 120} {
		m := newTestModel()
		m.width = w
		m.files = []FileInfo{{Name: "a.go"}}
		bar := m.renderBottomBar()
		if got := lipgloss.Width(bar); got != w {
			t.Errorf("width=%d: bottom bar width = %d, want %d", w, got, w)
		}
	}
}

func TestRenderNavRailShowsSectionsAndOmitsStats(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.files = []FileInfo{
		{Name: "src/main.rs", Status: "M", Added: 12, Removed: 3},
		{Name: "tests/ui_test.py", Status: "A", Added: 8, Removed: 0},
	}

	rail := m.renderNavRail()

	for _, want := range []string{"TERMINAL_UI", "PROJECT_TREE", "FILES", "HISTORY", "CONFIG"} {
		if !strings.Contains(rail, want) {
			t.Fatalf("rail missing %q: %q", want, rail)
		}
	}
	for _, want := range []string{"SRC/MAIN.RS", "TESTS/UI_TEST.PY"} {
		if !strings.Contains(rail, want) {
			t.Fatalf("rail missing uppercase filename %q: %q", want, rail)
		}
	}
	for _, unwanted := range []string{"(+12/-3)", "+12", "-3", "+8"} {
		if strings.Contains(rail, unwanted) {
			t.Fatalf("rail should not show diff stat fragment %q: %q", unwanted, rail)
		}
	}
}

func TestRenderNavRailSkipsInvalidDimensions(t *testing.T) {
	m := newTestModel()
	m.files = []FileInfo{{Name: "src/main.rs", Status: "M"}}

	cases := []struct {
		name      string
		listWidth int
		height    int
	}{
		{name: "zero width", listWidth: 0, height: 10},
		{name: "negative width", listWidth: -1, height: 10},
		{name: "zero body height", listWidth: 32, height: 2},
		{name: "negative body height", listWidth: 32, height: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m.listWidth = tc.listWidth
			m.height = tc.height

			if got := m.renderNavRail(); got != "" {
				t.Fatalf("renderNavRail() = %q, want empty string", got)
			}
		})
	}
}

func TestRenderWorkspaceHeaderShowsFileAndStats(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.listWidth = 28
	m.files = []FileInfo{{Name: "src/main.rs", Status: "M", Added: 1240, Removed: 302}}
	m.currentFile = "src/main.rs"

	header := m.renderWorkspaceHeader()

	for _, want := range []string{"EDITING_FILE:", "SRC/MAIN.RS", "+1,240", "-302"} {
		if !strings.Contains(header, want) {
			t.Fatalf("workspace header missing %q: %q", want, header)
		}
	}
}

func TestFormatStatCount(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1240, "1,240"},
		{999999, "999,999"},
		{1000000, "1,000,000"},
	}

	for _, tc := range cases {
		if got := formatStatCount(tc.n); got != tc.want {
			t.Errorf("formatStatCount(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}

func TestRenderWorkspaceHeaderFallback(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.listWidth = 28

	header := m.renderWorkspaceHeader()
	if !strings.Contains(header, "NO_FILE_SELECTED") {
		t.Fatalf("expected NO_FILE_SELECTED fallback: %q", header)
	}
}

func TestRenderWorkspaceHeaderNarrowWidth(t *testing.T) {
	for _, width := range []int{0, 5, 10, 17, 20} {
		m := newTestModel()
		m.width = width
		m.listWidth = 0
		m.files = []FileInfo{{Name: "src/main.rs", Status: "M", Added: 1240, Removed: 302}}
		m.currentFile = "src/main.rs"
		_ = m.renderWorkspaceHeader()
	}
}

func TestRenderDiffViewHeightAccountsForWorkspaceHeader(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.listWidth = 28
	m.diffViewport.SetContent("diff")

	view := m.renderDiffView()

	if got, want := lipgloss.Height(view), m.getContentHeight()-m.getWorkspaceHeaderHeight(); got != want {
		t.Fatalf("renderDiffView height = %d, want %d", got, want)
	}
}

func TestWindowSizeMsgAccountsForWorkspaceHeaderHeight(t *testing.T) {
	m := newTestModel()
	m.listWidth = 28

	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	updated := result.(Model)

	if got, want := updated.diffViewport.Width, updated.getDiffWidth(); got != want {
		t.Fatalf("diffViewport.Width = %d, want %d", got, want)
	}
	if got, want := updated.diffViewport.Height, updated.getContentHeight()-updated.getWorkspaceHeaderHeight(); got != want {
		t.Fatalf("diffViewport.Height = %d, want %d", got, want)
	}
}

func TestLayoutToggleAccountsForWorkspaceHeaderHeight(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.height = 40
	m.listWidth = 28

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}})
	updated := result.(Model)

	if updated.layout != LayoutVertical {
		t.Fatalf("layout = %v, want %v", updated.layout, LayoutVertical)
	}
	if got, want := updated.diffViewport.Width, updated.getDiffWidth(); got != want {
		t.Fatalf("diffViewport.Width = %d, want %d", got, want)
	}
	if got, want := updated.diffViewport.Height, updated.getContentHeight()-updated.getWorkspaceHeaderHeight(); got != want {
		t.Fatalf("diffViewport.Height = %d, want %d", got, want)
	}
}
