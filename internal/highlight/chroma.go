package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// SimpleHighlighter provides efficient syntax highlighting with caching
type SimpleHighlighter struct {
	style *chroma.Style
	cache map[string][]Token // Cache highlighted lines
}

// Token represents a highlighted segment
type Token struct {
	Text      string
	TokenType string
}

// New creates a new simple highlighter
func New(styleName string) *SimpleHighlighter {
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	return &SimpleHighlighter{
		style: style,
		cache: make(map[string][]Token),
	}
}

// GetColor returns a color code for a token type
func (h *SimpleHighlighter) GetColor(tokenType string) string {
	// Simple color mapping
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
	default:
		return ""
	}
}

// HighlightDiffLine highlights a diff line efficiently
func (h *SimpleHighlighter) HighlightDiffLine(line string, filename string) []Token {
	if len(line) == 0 {
		return []Token{{Text: line, TokenType: ""}}
	}

	// Check cache first
	if cached, ok := h.cache[line]; ok {
		result := make([]Token, len(cached))
		copy(result, cached)
		return result
	}

	// Extract prefix and content
	prefix := line[:1]
	content := line[1:]

	// Get lexer for file type
	lexer := lexers.Match(filename)
	if lexer == nil {
		// No lexer, return simple tokenization
		result := []Token{{Text: prefix, TokenType: ""}, {Text: content, TokenType: ""}}
		h.cache[line] = result
		return result
	}
	lexer = chroma.Coalesce(lexer)

	// Tokenize only the content (not the prefix)
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		result := []Token{{Text: prefix, TokenType: ""}, {Text: content, TokenType: ""}}
		h.cache[line] = result
		return result
	}

	// Convert tokens
	var tokens []Token
	tokens = append(tokens, Token{Text: prefix, TokenType: ""})

	for _, tok := range iterator.Tokens() {
		tokens = append(tokens, Token{
			Text:      tok.Value,
			TokenType: tok.Type.String(),
		})
	}

	// Cache the result
	h.cache[line] = tokens

	return tokens
}

// ClearCache clears the highlight cache (call when switching files)
func (h *SimpleHighlighter) ClearCache() {
	h.cache = make(map[string][]Token)
}

// TokenizeFile tokenizes multiple lines of code in a single Chroma call.
// filename is used for lexer detection. lines are plain code without diff prefixes.
// Returns per-line token slices. Concatenating tokens per line reconstructs the input.
func (h *SimpleHighlighter) TokenizeFile(filename string, lines []string) [][]Token {
	if len(lines) == 0 {
		return nil
	}

	lexer := lexers.Match(filename)
	if lexer == nil {
		result := make([][]Token, len(lines))
		for i, line := range lines {
			result[i] = []Token{{Text: line, TokenType: ""}}
		}
		return result
	}
	lexer = chroma.Coalesce(lexer)

	full := strings.Join(lines, "\n") + "\n"

	iterator, err := lexer.Tokenise(nil, full)
	if err != nil {
		result := make([][]Token, len(lines))
		for i, line := range lines {
			result[i] = []Token{{Text: line, TokenType: ""}}
		}
		return result
	}

	result := make([][]Token, len(lines))
	lineIdx := 0

outer:
	for _, tok := range iterator.Tokens() {
		text := tok.Value
		tokType := tok.Type.String()

		for text != "" {
			if lineIdx >= len(result) {
				break outer
			}
			nlPos := strings.Index(text, "\n")
			if nlPos == -1 {
				result[lineIdx] = append(result[lineIdx], Token{Text: text, TokenType: tokType})
				break
			}
			if nlPos > 0 {
				result[lineIdx] = append(result[lineIdx], Token{Text: text[:nlPos], TokenType: tokType})
			}
			lineIdx++
			text = text[nlPos+1:]
		}
	}

	for i := range result {
		if result[i] == nil {
			result[i] = []Token{}
		}
	}

	return result
}
