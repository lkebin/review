package diff

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	input := "@@ -10,7 +10,8 @@ func main() {\n" +
		" // Setup\n" +
		"-fmt.Println(\"old\")\n" +
		"+fmt.Println(\"new\")\n" +
		"+fmt.Println(\"added\")\n" +
		" // Run\n" +
		"}\n"

	lines := Parse(input)

	if len(lines) == 0 {
		t.Fatal("Expected lines, got none")
	}

	// Check hunk header
	if lines[0].Type != LineHunkHeader {
		t.Errorf("Expected first line to be hunk header, got %d", lines[0].Type)
	}

	// Find removed line
	foundRemoved := false
	for _, line := range lines {
		if line.Type == LineRemoved {
			foundRemoved = true
			if line.OldLineNo != 11 {
				t.Errorf("Expected removed line OldLineNo=11, got %d", line.OldLineNo)
			}
			if line.NewLineNo != 0 {
				t.Errorf("Expected removed line NewLineNo=0, got %d", line.NewLineNo)
			}
			break
		}
	}
	if !foundRemoved {
		t.Error("Expected to find removed line")
	}

	// Find added lines
	addedCount := 0
	for _, line := range lines {
		if line.Type == LineAdded {
			addedCount++
		}
	}
	if addedCount != 2 {
		t.Errorf("Expected 2 added lines, got %d", addedCount)
	}

	// Check context lines have both line numbers
	for _, line := range lines {
		if line.Type == LineContext {
			if line.OldLineNo == 0 || line.NewLineNo == 0 {
				t.Errorf("Context line should have both line numbers, got old=%d new=%d",
					line.OldLineNo, line.NewLineNo)
			}
		}
	}
}

func TestParseMultipleHunks(t *testing.T) {
	input := "@@ -1,3 +1,4 @@\n" +
		" line1\n" +
		"+added1\n" +
		" line2\n" +
		" line3\n" +
		"@@ -10,3 +11,4 @@\n" +
		" line10\n" +
		"+added2\n" +
		" line11\n" +
		" line12\n"

	lines := Parse(input)

	hunkCount := 0
	for _, line := range lines {
		if line.Type == LineHunkHeader {
			hunkCount++
		}
	}

	if hunkCount != 2 {
		t.Errorf("Expected 2 hunks, got %d", hunkCount)
	}
}

func TestParseEmptyDiff(t *testing.T) {
	lines := Parse("")
	if len(lines) != 0 {
		t.Errorf("Expected 0 lines for empty input, got %d", len(lines))
	}
}

// TestParseCRLF verifies that Windows-style CRLF line endings do not leave
// a trailing \r in the parsed content (which would corrupt terminal rendering).
func TestParseCRLF(t *testing.T) {
	input := "@@ -1,3 +1,3 @@\r\n" +
		" context line\r\n" +
		"-old line\r\n" +
		"+new line\r\n"

	lines := Parse(input)

	for _, l := range lines {
		if strings.HasSuffix(l.Content, "\r") {
			t.Errorf("line content still contains \\r: %q (type=%v)", l.Content, l.Type)
		}
	}
}
