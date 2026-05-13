package ui

import "github.com/charmbracelet/lipgloss"

// Theme holds all color values as ANSI 256-color strings.
type Theme struct {
	NormalFg        string
	AddedBg         string
	RemovedBg       string
	AddedCursorBg   string
	RemovedCursorBg string
	CursorBg        string
	InlineAddBg     string
	InlineDelBg     string
	LineNoFg        string
	SepFg           string
	HunkFg          string
	StatusBarFg     string
	StatusBarBg     string
	FileSelectedFg  string
	FileSelectedBg  string
	StatusM         string
	StatusA         string
	StatusD         string
	StatusR         string
	StatusC         string
}

// DefaultTheme returns the built-in 256-color theme.
func DefaultTheme() Theme {
	return Theme{
		NormalFg:        "252",
		AddedBg:         "22",
		RemovedBg:       "52",
		AddedCursorBg:   "28",
		RemovedCursorBg: "88",
		CursorBg:        "236",
		InlineAddBg:     "28",
		InlineDelBg:     "88",
		LineNoFg:        "241",
		SepFg:           "237",
		HunkFg:          "140",
		StatusBarFg:     "252",
		StatusBarBg:     "236",
		FileSelectedFg:  "16",
		FileSelectedBg:  "75",
		StatusM:         "178",
		StatusA:         "40",
		StatusD:         "167",
		StatusR:         "133",
		StatusC:         "73",
	}
}

// StatusColor returns the foreground color for a file status badge.
func (th Theme) StatusColor(status string) string {
	switch status {
	case "M":
		return th.StatusM
	case "A":
		return th.StatusA
	case "D":
		return th.StatusD
	case "R":
		return th.StatusR
	case "C":
		return th.StatusC
	default:
		return th.NormalFg
	}
}

// Convenience lipgloss style builders.

// LineNoStyle returns a style for line number text (foreground only; background is set per-line).
func (th Theme) LineNoStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(th.LineNoFg))
}

// SepStyle returns a style for the separator character (foreground only; background is set per-line).
func (th Theme) SepStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(th.SepFg))
}

func (th Theme) HunkStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(th.HunkFg))
}

func (th Theme) StatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.StatusBarFg)).
		Background(lipgloss.Color(th.StatusBarBg))
}

func (th Theme) FileSelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.FileSelectedFg)).
		Background(lipgloss.Color(th.FileSelectedBg)).
		Bold(true)
}
