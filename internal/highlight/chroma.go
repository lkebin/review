package highlight

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Token represents a highlighted token with color
type Token struct {
	Text  string
	Color string // ANSI color code
}

// Line represents a highlighted line
type Line struct {
	Tokens []Token
}

// Highlighter handles syntax highlighting
type Highlighter struct {
	style *chroma.Style
}

// New creates a new highlighter with the given style
func New(styleName string) *Highlighter {
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	return &Highlighter{style: style}
}

// Highlight highlights the given code content
func (h *Highlighter) Highlight(content string, filename string) []Line {
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Analyse(content)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Tokenise the content
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return h.noHighlight(content)
	}

	// Get all tokens
	tokens := iterator.Tokens()

	return h.tokensToLines(tokens)
}

// noHighlight returns lines without highlighting
func (h *Highlighter) noHighlight(content string) []Line {
	lines := strings.Split(content, "\n")
	result := make([]Line, len(lines))
	for i, line := range lines {
		result[i] = Line{
			Tokens: []Token{{Text: line, Color: ""}},
		}
	}
	return result
}

// tokensToLines converts chroma tokens to our Line format
func (h *Highlighter) tokensToLines(tokens []chroma.Token) []Line {
	var lines []Line
	var currentLine []Token

	for _, token := range tokens {
		tokenType := token.Type
		color := h.getANSIColor(tokenType)

		// Split token by newlines
		parts := strings.Split(token.Value, "\n")
		for i, part := range parts {
			if i > 0 && len(currentLine) > 0 {
				// New line, save current and start new
				lines = append(lines, Line{Tokens: currentLine})
				currentLine = nil
			}
			if part != "" || len(currentLine) > 0 {
				currentLine = append(currentLine, Token{
					Text:  part,
					Color: color,
				})
			}
		}
	}

	// Don't forget the last line
	if len(currentLine) > 0 {
		lines = append(lines, Line{Tokens: currentLine})
	}

	return lines
}

// getANSIColor returns the ANSI color code for a token type
func (h *Highlighter) getANSIColor(tokenType chroma.TokenType) string {
	entry := h.style.Get(tokenType)
	if entry.IsZero() {
		return ""
	}

	// Convert color to ANSI 256 color code
	if entry.Colour.IsSet() {
		c := entry.Colour
		// Calculate 256-color code
		r := c.Red() * 5 / 255
		g := c.Green() * 5 / 255
		b := c.Blue() * 5 / 255
		code := 16 + 36*r + 6*g + b
		return fmt.Sprintf("\x1b[38;5;%dm", code)
	}
	return ""
}

// HighlightDiffLine highlights a single diff line
func (h *Highlighter) HighlightDiffLine(line string, filename string) []Token {
	// For diff lines, we need to:
	// 1. Keep the +/- prefix
	// 2. Highlight the rest of the content

	if len(line) == 0 {
		return []Token{{Text: line, Color: ""}}
	}

	// Get the prefix (+, -, or space)
	prefix := line[:1]
	content := line[1:]

	// Highlight the content
	lines := h.Highlight(content, filename)
	if len(lines) == 0 {
		return []Token{{Text: line, Color: ""}}
	}

	// Prepend the prefix as first token
	result := []Token{{Text: prefix, Color: ""}}
	result = append(result, lines[0].Tokens...)

	return result
}
