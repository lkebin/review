package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
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

	formatter := formatters.Get("tokens")
	if formatter == nil {
		return h.noHighlight(content)
	}

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return h.noHighlight(content)
	}

	tokens, err := formatter.Format(iterator)
	if err != nil {
		return h.noHighlight(content)
	}

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
		color := h.getColor(tokenType)

		// Split token by newlines
		parts := strings.Split(token.Value, "\n")
		for i, part := range parts {
			if i > 0 {
				// New line, save current and start new
				lines = append(lines, Line{Tokens: currentLine})
				currentLine = nil
			}
			if part != "" {
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

// getColor returns the ANSI color for a token type
func (h *Highlighter) getColor(tokenType chroma.TokenType) string {
	entry := h.style.Get(tokenType)
	if entry.IsZero() {
		return ""
	}

	var codes []string

	// Foreground color
	if entry.Colour.IsSet() {
		// Convert to ANSI 256 color
		c := entry.Colour
		code := 16 + (36 * (c.Red() * 5 / 255)) + (6 * (c.Green() * 5 / 255)) + (c.Blue() * 5 / 255)
		codes = append(codes, "38;5;"+string(rune('0'+int(code))))
	}

	if len(codes) > 0 {
		return "\x1b[" + strings.Join(codes, ";") + "m"
	}
	return ""
}

// HighlightDiffLine highlights a single diff line (without the +/- prefix)
func (h *Highlighter) HighlightDiffLine(line string, filename string) []Token {
	// Remove diff prefix if present
	content := line
	if len(line) > 0 && (line[0] == '+' || line[0] == '-' || line[0] == ' ') {
		content = line[1:]
	}

	// Keep the prefix
	prefix := line[:len(line)-len(content)]

	tokens := h.Highlight(content, filename)
	if len(tokens) == 0 {
		return []Token{{Text: line, Color: ""}}
	}

	// Prepend the prefix to the first token
	result := []Token{{Text: prefix, Color: ""}}
	if len(tokens[0].Tokens) > 0 {
		result = append(result, tokens[0].Tokens...)
	}

	return result
}
