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
