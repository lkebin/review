package diff

import (
	"regexp"
	"strconv"
	"strings"
)

// LineType represents the type of a diff line
type LineType int

const (
	LineContext LineType = iota
	LineAdded
	LineRemoved
	LineHunkHeader
)

// Line represents a single line in a diff
type Line struct {
	Type      LineType
	OldLineNo int // 0 means not present in old file
	NewLineNo int // 0 means not present in new file
	Content   string
}

// Parse parses git diff output into Lines
func Parse(output string) []Line {
	var lines []Line
	oldLine := 0
	newLine := 0
	inHunk := false

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}

		// Check for hunk header: @@ -old_start,old_count +new_start,new_count @@
		if strings.HasPrefix(line, "@@") {
			inHunk = true
			lines = append(lines, Line{
				Type:    LineHunkHeader,
				Content: line,
			})

			// Parse line numbers from hunk header
			// Format: @@ -start,count +start,count @@
			oldLine, newLine = parseHunkHeader(line)
			continue
		}

		if !inHunk {
			// Skip header lines (---, +++, diff --git, etc.)
			continue
		}

		if len(line) == 0 {
			continue
		}

		// Handle both space-prefixed context lines and other lines
		// Context lines start with ' ' but may have leading whitespace
		firstChar := line[0]
		if firstChar != '+' && firstChar != '-' && firstChar != ' ' {
			// Skip lines that don't start with +, -, or space
			continue
		}

		switch firstChar {
		case '+':
			lines = append(lines, Line{
				Type:      LineAdded,
				NewLineNo: newLine,
				Content:   line,
			})
			newLine++

		case '-':
			lines = append(lines, Line{
				Type:      LineRemoved,
				OldLineNo: oldLine,
				Content:   line,
			})
			oldLine++

		case ' ':
			lines = append(lines, Line{
				Type:      LineContext,
				OldLineNo: oldLine,
				NewLineNo: newLine,
				Content:   line,
			})
			oldLine++
			newLine++
		}
	}

	return lines
}

var hunkHeaderRe = regexp.MustCompile(`@@ -(\d+),\d+ \+(\d+),\d+ @@`)

func parseHunkHeader(line string) (oldStart, newStart int) {
	matches := hunkHeaderRe.FindStringSubmatch(line)
	if len(matches) >= 3 {
		oldStart, _ = strconv.Atoi(matches[1])
		newStart, _ = strconv.Atoi(matches[2])
	}
	return
}

// Stats represents diff statistics
type Stats struct {
	Added   int
	Removed int
}

// CalculateStats calculates the diff statistics from lines
func CalculateStats(lines []Line) Stats {
	var stats Stats
	for _, line := range lines {
		switch line.Type {
		case LineAdded:
			stats.Added++
		case LineRemoved:
			stats.Removed++
		}
	}
	return stats
}
