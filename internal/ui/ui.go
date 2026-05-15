package ui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Options contains the command line options
type Options struct {
	Target       string
	Staged       bool
	ContextLines int
}

// Run starts the TUI application
func Run(opts Options) error {
	cp := &cursorPositioner{inner: os.Stdout}
	m := NewModel(opts)
	m.cursorPos = cp
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithOutput(cp))
	_, err := p.Run()
	return err
}
