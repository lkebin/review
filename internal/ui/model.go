package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kbliu/review/internal/highlight"
)

// FileInfo represents a file in the review
type FileInfo struct {
	Status  string // A, M, D, R, C
	Name    string
	Added   int
	Removed int
}

// FocusType indicates which panel has focus
type FocusType int

const (
	FocusList FocusType = iota
	FocusDiff
)

// LayoutType indicates the current layout
type LayoutType int

const (
	LayoutHorizontal LayoutType = iota
	LayoutVertical
)

// Model represents the application state
type Model struct {
	// Options
	options Options

	// Layout
	layout LayoutType

	// Dimensions
	width     int
	height    int
	listWidth int // Adjustable file list width

	// File list
	files  []FileInfo
	cursor int

	// Diff content
	currentFile string
	diffLines   []DiffLine
	diffCursor  int

	// Focus
	focus FocusType

	// Viewport for diff scrolling
	diffViewport viewport.Model

	// Syntax highlighter
	highlighter *highlight.SimpleHighlighter

	// Status
	err     error
	loading bool

	// Help popup
	showHelp bool
}

// LineType represents the type of a diff line
type LineType int

const (
	LineContext LineType = iota
	LineAdded
	LineRemoved
	LineHunkHeader
)

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type      LineType
	OldLineNo int // 0 means not present in old file
	NewLineNo int // 0 means not present in new file
	Content   string
}

// NewModel creates a new model with the given options
func NewModel(opts Options) Model {
	return Model{
		options:     opts,
		layout:      LayoutHorizontal,
		focus:       FocusList,
		loading:     true,
		listWidth:   32,
		highlighter: highlight.New("github"),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return loadFiles(m.options)
}
