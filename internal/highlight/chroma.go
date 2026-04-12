package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// SimpleHighlighter provides basic syntax highlighting
type SimpleHighlighter struct {
	style *chroma.Style
}

// Token represents a highlighted segment
type Token struct {
	Text      string
	TokenType string // e.g., "keyword", "string", "comment"
}

// New creates a new simple highlighter
func New(styleName string) *SimpleHighlighter {
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	return &SimpleHighlighter{style: style}
}

// HighlightLine highlights a single line of code
func (h *SimpleHighlighter) HighlightLine(content string, filename string) []Token {
	// Empty content
	if len(content) == 0 {
		return []Token{{Text: content, TokenType: ""}}
	}

	// Get lexer
	lexer := lexers.Match(filename)
	if lexer == nil {
		return []Token{{Text: content, TokenType: ""}}
	}
	lexer = chroma.Coalesce(lexer)

	// Tokenise just this line
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return []Token{{Text: content, TokenType: ""}}
	}

	// Convert tokens
	var tokens []Token
	for _, tok := range iterator.Tokens() {
		tokens = append(tokens, Token{
			Text:      tok.Value,
			TokenType: tok.Type.String(),
		})
	}

	return tokens
}

// HighlightDiffLine highlights a diff line (preserves +/- prefix)
func (h *SimpleHighlighter) HighlightDiffLine(line string, filename string) []Token {
	if len(line) == 0 {
		return []Token{{Text: line, TokenType: ""}}
	}

	// Extract prefix and content
	prefix := line[:1]
	content := line[1:]

	// Highlight content only
	tokens := h.HighlightLine(content, filename)

	// Prepend prefix
	result := []Token{{Text: prefix, TokenType: ""}}
	result = append(result, tokens...)

	return result
}

// GetColor returns the color for a token type
func (h *SimpleHighlighter) GetColor(tokenType string) string {
	// Map common token types to lipgloss colors
	switch {
	case strings.Contains(tokenType, "Keyword"):
		return "204" // Pink
	case strings.Contains(tokenType, "String"):
		return "192" // Light green
	case strings.Contains(tokenType, "Comment"):
		return "243" // Gray
	case strings.Contains(tokenType, "Number"):
		return "180" // Orange
	case strings.Contains(tokenType, "Function"):
		return "117" // Light blue
	case strings.Contains(tokenType, "Operator"):
		return "186" // Yellow
	default:
		return "" // Default
	}
}
