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

func TestRenderCmdLineNoCursor(t *testing.T) {
	// Neither typing nor confirmed state should show the fake block cursor.
	bar := RenderCmdLine("foo", 80, true)
	if strings.Contains(bar, "▌") {
		t.Error("typing=true: cmd line must not contain fake cursor ▌")
	}

	bar2 := RenderCmdLine("foo", 80, false)
	if strings.Contains(bar2, "▌") {
		t.Error("typing=false: cmd line must not contain ▌")
	}
}

func TestRenderCmdLineWidth(t *testing.T) {
	for _, w := range []int{40, 80, 120} {
		bar := RenderCmdLine("foo", w, true)
		got := lipgloss.Width(bar)
		if got != w {
			t.Errorf("width %d: cmd line visual width = %d, want %d", w, got, w)
		}
	}
}

func TestRenderCmdLineBlankWhenIdle(t *testing.T) {
	bar := RenderCmdLine("", 80, false)
	if strings.Contains(bar, "/") {
		t.Error("idle cmd line must not contain /")
	}
}

func TestRenderCmdLineShowsQuery(t *testing.T) {
	bar := RenderCmdLine("hello", 80, false)
	if !strings.Contains(bar, "/") || !strings.Contains(bar, "hello") {
		t.Error("confirmed search must show /query")
	}
}
