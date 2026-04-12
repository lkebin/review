package ui

import (
	"github.com/charmbracelet/bubbletea"
)

// loadFilesMsg is sent when files are loaded
type loadFilesMsg struct {
	files []FileInfo
	err   error
}

// loadDiffMsg is sent when diff is loaded
type loadDiffMsg struct {
	lines []DiffLine
	err   error
}

// loadFiles is a command to load the file list
func loadFiles(opts Options) tea.Cmd {
	return func() tea.Msg {
		files, err := getFiles(opts)
		return loadFilesMsg{files: files, err: err}
	}
}

// loadDiff is a command to load diff for a file
func loadDiff(opts Options, file string) tea.Cmd {
	return func() tea.Msg {
		lines, err := getDiff(opts, file)
		return loadDiffMsg{lines: lines, err: err}
	}
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.diffViewport.Width = m.getDiffWidth()
		m.diffViewport.Height = m.getContentHeight()
		return m, nil

	case loadFilesMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.files = msg.files
		if len(m.files) > 0 {
			return m, loadDiff(m.options, m.files[0].Name)
		}
		return m, nil

	case loadDiffMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.diffLines = msg.lines
		m.diffViewport.SetContent(m.renderDiff())
		return m, nil
	}

	// Handle viewport updates
	if m.focus == FocusDiff {
		var cmd tea.Cmd
		m.diffViewport, cmd = m.diffViewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "tab":
		if m.focus == FocusList {
			m.focus = FocusDiff
		} else {
			m.focus = FocusList
		}
		return m, nil

	case "h", "left":
		m.focus = FocusList
		return m, nil

	case "l", "right":
		if m.focus == FocusDiff {
			// Toggle layout
			if m.layout == LayoutHorizontal {
				m.layout = LayoutVertical
			} else {
				m.layout = LayoutHorizontal
			}
			// Resize viewport
			m.diffViewport.Width = m.getDiffWidth()
			m.diffViewport.Height = m.getContentHeight()
		} else {
			m.focus = FocusDiff
		}
		return m, nil

	case "j", "down":
		if m.focus == FocusList {
			if m.cursor < len(m.files)-1 {
				m.cursor++
				return m, loadDiff(m.options, m.files[m.cursor].Name)
			}
		}
		return m, nil

	case "k", "up":
		if m.focus == FocusList {
			if m.cursor > 0 {
				m.cursor--
				return m, loadDiff(m.options, m.files[m.cursor].Name)
			}
		}
		return m, nil

	case "enter":
		if m.focus == FocusList && m.cursor < len(m.files) {
			m.focus = FocusDiff
			return m, loadDiff(m.options, m.files[m.cursor].Name)
		}
		return m, nil
	}

	return m, nil
}

// Helper methods for dimensions
func (m Model) getListWidth() int {
	if m.layout == LayoutHorizontal {
		return min(30, m.width/3)
	}
	return m.width
}

func (m Model) getDiffWidth() int {
	if m.layout == LayoutHorizontal {
		return m.width - m.getListWidth() - 1
	}
	return m.width
}

func (m Model) getContentHeight() int {
	// Reserve 1 line for status bar
	return m.height - 1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
