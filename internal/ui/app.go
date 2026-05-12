// internal/ui/app.go
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kbliu/review/internal/diff"
	"github.com/kbliu/review/internal/git"
)

// FileInfo represents a file with diff statistics.
type FileInfo struct {
	Status  string
	Name    string
	Added   int
	Removed int
}

// FocusType indicates which panel has focus.
type FocusType int

const (
	FocusList FocusType = iota
	FocusDiff
)

// Model is the top-level BubbleTea model.
type Model struct {
	opts    Options
	width   int
	height  int
	files   []FileInfo
	err     error
	loading bool
}

// NewModel creates a new Model.
func NewModel(opts Options) Model {
	return Model{opts: opts, loading: true}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error()
	}
	return "review — press any key to quit"
}

// gitOptions converts ui.Options to git.Options.
func gitOptions(opts Options) git.Options {
	return git.Options{
		Target:       opts.Target,
		Staged:       opts.Staged,
		ContextLines: opts.ContextLines,
	}
}

// loadFileList fetches files and calculates stats.
func loadFileList(opts Options) ([]FileInfo, error) {
	gopts := gitOptions(opts)
	files, err := git.GetFiles(gopts)
	if err != nil {
		return nil, err
	}
	result := make([]FileInfo, len(files))
	for i, f := range files {
		content, err := git.GetDiff(gopts, f.Name)
		if err != nil {
			result[i] = FileInfo{Status: f.Status, Name: f.Name}
			continue
		}
		lines := diff.Parse(content)
		stats := diff.CalculateStats(lines)
		result[i] = FileInfo{
			Status:  f.Status,
			Name:    f.Name,
			Added:   stats.Added,
			Removed: stats.Removed,
		}
	}
	return result, nil
}
