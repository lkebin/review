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
// Format: " <left> <right> " → digitWidth*2 + 3 (leading space + separator + trailing space).
func LineNoColumnWidth(digitWidth int) int {
	return digitWidth*2 + 3
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
	return " " + l + " " + r + " "
}
