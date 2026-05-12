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

	if len(result[0]) == 0 {
		t.Error("first line has no tokens")
	}

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
