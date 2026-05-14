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

func TestRenderSearchBarTypingCursor(t *testing.T) {
	th := DefaultTheme()

	// typing=true: ▌ cursor shown
	bar := RenderSearchBar("foo", FocusList, 80, th, true)
	if !strings.Contains(bar, "▌") {
		t.Error("typing=true: bar must contain cursor ▌")
	}

	// typing=false: no cursor
	bar2 := RenderSearchBar("foo", FocusList, 80, th, false)
	if strings.Contains(bar2, "▌") {
		t.Error("typing=false: bar must not contain ▌")
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
