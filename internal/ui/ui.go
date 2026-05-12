package ui

import (
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
	m := NewModel(opts)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
