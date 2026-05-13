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
		{12, 15, 4, "   12   15 "},
		{0, 17, 4, "        17 "},
		{14, 0, 4, "   14      "},
		{0, 0, 4, "           "}, // hunk header placeholder
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
	// Total column width = digitWidth*2 + 3 for the " LLLL RRRR " format
	if got, want := LineNoColumnWidth(4), 11; got != want {
		t.Errorf("LineNoColumnWidth(4) = %d, want %d", got, want)
	}
	if got, want := LineNoColumnWidth(3), 9; got != want {
		t.Errorf("LineNoColumnWidth(3) = %d, want %d", got, want)
	}
}
