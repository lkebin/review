// internal/ui/wrap_test.go
package ui

import (
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// TestFindANSISafeBreakRespectsRuneBoundaries verifies the function never
// returns a byte index that falls inside a multi-byte UTF-8 character.
// Splitting at such an index produces orphan bytes that render as ◊
// replacement characters in the terminal.
func TestFindANSISafeBreakRespectsRuneBoundaries(t *testing.T) {
	// `█` is U+2588 → 3 bytes (0xE2 0x96 0x88), 1 column wide.
	input := "abc█def"

	for maxW := 1; maxW <= utf8.RuneCountInString(input); maxW++ {
		bp := findANSISafeBreak(input, maxW)
		if bp < 0 || bp > len(input) {
			t.Fatalf("maxW=%d: break point %d out of range [0,%d]", maxW, bp, len(input))
		}
		first := input[:bp]
		second := input[bp:]
		if !utf8.ValidString(first) {
			t.Errorf("maxW=%d: first chunk %q has invalid UTF-8 (split a rune)", maxW, first)
		}
		if !utf8.ValidString(second) {
			t.Errorf("maxW=%d: second chunk %q has invalid UTF-8 (split a rune)", maxW, second)
		}
	}
}

// TestFindANSISafeBreakCountsColumnsNotBytes verifies that the break point
// is computed by visible column count, not raw byte count. With a wide CJK
// character (2 columns), the function must not advance more than maxWidth
// columns in the returned prefix.
func TestFindANSISafeBreakCountsColumnsNotBytes(t *testing.T) {
	// "你" is 3 bytes wide and 2 columns wide.
	// "a你b" → 5 bytes, 4 columns.
	input := "a你b"
	cases := []struct {
		maxW    int
		wantCol int // max allowed visible columns in the first chunk
	}{
		{1, 1}, // only "a" fits (can't fit 你 which would push to 3 cols)
		{2, 1}, // still only "a"; "a你" would be 3 cols > 2
		{3, 3}, // "a你" fits exactly
		{4, 4}, // whole string fits
	}
	for _, c := range cases {
		bp := findANSISafeBreak(input, c.maxW)
		got := lipgloss.Width(input[:bp])
		if got > c.wantCol {
			t.Errorf("maxW=%d: first chunk has %d columns, want ≤ %d (chunk=%q)",
				c.maxW, got, c.wantCol, input[:bp])
		}
	}
}

// TestFindANSISafeBreakCJKAtBoundary is a regression test for the bug where
// breaking before a wide CJK character put its opening ANSI escape in the
// first chunk and the raw character bytes in the second, causing the character
// to render with the terminal default (black) background instead of the styled one.
func TestFindANSISafeBreakCJKAtBoundary(t *testing.T) {
	st := lipgloss.NewStyle().Background(lipgloss.Color("22"))
	// 'a' is 1 column, '。' is 2 columns.  At maxWidth=2 the break falls
	// between them: 'a' goes in chunk 1, '。' must start chunk 2 WITH its ANSI prefix.
	s := st.Render("a") + st.Render("。")

	bp := findANSISafeBreak(s, 2)
	second := s[bp:]

	if len(second) == 0 {
		t.Fatal("second chunk is empty; expected it to contain '。'")
	}
	if second[0] != '\x1b' {
		t.Errorf("second chunk starts with raw byte %#x (%q), want ANSI escape \\x1b; "+
			"the opening color code was left in the first chunk so the character "+
			"renders with the terminal default (black) background",
			second[0], second[:min(20, len(second))])
	}
}

// TestWrapRenderedLineMultiByteContent verifies the full wrapping pipeline
// produces only valid UTF-8 chunks when content includes multi-byte chars.
// Reproduces the reported bug where wrapping at certain widths produced ◊
// replacement characters from orphan UTF-8 bytes.
func TestWrapRenderedLineMultiByteContent(t *testing.T) {
	// Mimic renderContent's per-rune styling, the real source of the bug.
	st := lipgloss.NewStyle().Background(lipgloss.Color("52"))
	var rendered string
	for _, r := range "// fake █ — neither state" {
		rendered += st.Render(string(r))
	}
	for maxW := 5; maxW <= 25; maxW++ {
		chunks := wrapRenderedLine(rendered, maxW)
		for i, c := range chunks {
			if !utf8.ValidString(c) {
				t.Errorf("maxW=%d chunk[%d]=%q has invalid UTF-8", maxW, i, c)
			}
		}
	}
}
