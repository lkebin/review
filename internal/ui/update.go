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
	// Handle help popup first
	if m.showHelp {
		switch msgType := msg.(type) {
		case tea.KeyMsg:
			// Any key closes help
			_ = msgType
			m.showHelp = false
			return m, nil
		case tea.MouseMsg:
			m.showHelp = false
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.diffViewport.Width = m.getDiffWidth()
		m.diffViewport.Height = m.getContentHeight()
		// Re-render diff when window size changes (for word wrap)
		if len(m.diffLines) > 0 {
			m.diffViewport.SetContent(m.renderDiff())
		}
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
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.diffLines = msg.lines
		// Defer diff rendering to next frame to improve perceived performance
		return m, func() tea.Msg {
			return diffContentReadyMsg{lines: msg.lines}
		}
	case diffContentReadyMsg:
		m.diffViewport.SetContent(m.renderDiff())
		m.diffViewport.GotoTop()
		return m, nil
	}

	// Handle viewport updates (mouse scrolling etc)
	if m.focus == FocusDiff {
		var cmd tea.Cmd
		m.diffViewport, cmd = m.diffViewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

type diffContentReadyMsg struct {
	lines []DiffLine
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Help popup toggle
	if msg.String() == "?" {
		m.showHelp = !m.showHelp
		return m, nil
	}

	// Quit
	if msg.String() == "q" || msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// Ctrl+W - toggle focus
	if msg.String() == "ctrl+w" {
		if m.focus == FocusList {
			m.focus = FocusDiff
		} else {
			m.focus = FocusList
		}
		return m, nil
	}

	// H/h - decrease/increase file list width (like vim)
	switch msg.String() {
	case "h":
		if m.layout == LayoutHorizontal && m.listWidth > 10 {
			m.listWidth -= 5
			m.diffViewport.Width = m.getDiffWidth()
		}
		return m, nil
	case "H":
		if m.layout == LayoutHorizontal && m.listWidth < m.width/2 {
			m.listWidth += 5
			m.diffViewport.Width = m.getDiffWidth()
		}
		return m, nil
	}

	// Shift+L - toggle layout
	if msg.String() == "L" {
		if m.layout == LayoutHorizontal {
			m.layout = LayoutVertical
		} else {
			m.layout = LayoutHorizontal
		}
		// Resize viewport
		m.diffViewport.Width = m.getDiffWidth()
		m.diffViewport.Height = m.getContentHeight()
		return m, nil
	}

	// Window-specific key handling
	if m.focus == FocusList {
		return m.handleListKeys(msg)
	}
	return m.handleDiffKeys(msg)
}

func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(m.files)-1 {
			m.cursor++
			m.loading = true
			return m, loadDiff(m.options, m.files[m.cursor].Name)
		}
		return m, nil

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.loading = true
			return m, loadDiff(m.options, m.files[m.cursor].Name)
		}
		return m, nil

	case "enter":
		m.focus = FocusDiff
		return m, nil
	}

	return m, nil
}

func (m Model) handleDiffKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.diffViewport.ScrollDown(1)
		return m, nil

	case "k", "up":
		m.diffViewport.ScrollUp(1)
		return m, nil

	case "ctrl+d":
		m.diffViewport.ScrollDown(m.diffViewport.Height / 2)
		return m, nil

	case "ctrl+u":
		m.diffViewport.ScrollUp(m.diffViewport.Height / 2)
		return m, nil

	case "ctrl+f":
		// Page down (forward)
		m.diffViewport.ScrollDown(m.diffViewport.Height)
		return m, nil

	case "ctrl+b":
		// Page up (backward)
		m.diffViewport.ScrollUp(m.diffViewport.Height)
		return m, nil

	case "g":
		m.diffViewport.GotoTop()
		return m, nil

	case "G":
		m.diffViewport.GotoBottom()
		return m, nil
	}

	return m, nil
}

// Helper methods for dimensions
func (m Model) getListWidth() int {
	if m.layout == LayoutHorizontal {
		return min(m.listWidth, m.width/2)
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
