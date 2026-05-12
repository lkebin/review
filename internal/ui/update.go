package ui

import (
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbletea"
)

// loadFilesMsg is sent when files are loaded
type loadFilesMsg struct {
	files []FileInfo
	err   error
}

// loadDiffMsg is sent when diff is loaded
type loadDiffMsg struct {
	file  string
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
		return loadDiffMsg{lines: lines, file: file, err: err}
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
	case tea.MouseMsg:
		// In horizontal layout the list panel occupies columns [0, listWidth).
		// In vertical layout both panels span the full width, so X-based routing
		// is not meaningful — treat all wheel events as diff-panel scrolling.
		inListPanel := m.layout == LayoutHorizontal && msg.X < m.listWidth
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			if inListPanel {
				if m.cursor > 0 {
					m.cursor--
				}
			} else {
				if m.diffCursor > 0 {
					m.diffCursor--
					if m.diffCursor < m.diffViewport.YOffset {
						m.diffViewport.YOffset = m.diffCursor
					}
				}
			}
		case tea.MouseButtonWheelDown:
			if inListPanel {
				if m.cursor < len(m.files)-1 {
					m.cursor++
				}
			} else {
				if len(m.diffLines) > 0 && m.diffCursor < len(m.diffLines)-1 {
					m.diffCursor++
					if m.diffCursor >= m.diffViewport.YOffset+m.diffViewport.Height {
						m.diffViewport.YOffset = m.diffCursor - m.diffViewport.Height + 1
					}
				}
			}
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.diffViewport.Width = m.getDiffWidth()
		m.diffViewport.Height = m.getContentHeight() - m.getWorkspaceHeaderHeight()
		if m.diffViewport.Height < 0 {
			m.diffViewport.Height = 0
		}
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
		m.currentFile = msg.file
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

	case editorFinishedMsg:
		// Editor closed, refresh the diff
		if m.cursor < len(m.files) {
			return m, loadDiff(m.options, m.files[m.cursor].Name)
		}
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

	// Shift+L - toggle layout
	if msg.String() == "L" {
		if m.layout == LayoutHorizontal {
			m.layout = LayoutVertical
		} else {
			m.layout = LayoutHorizontal
		}
		// Resize viewport
		m.diffViewport.Width = m.getDiffWidth()
		m.diffViewport.Height = m.getContentHeight() - m.getWorkspaceHeaderHeight()
		if m.diffViewport.Height < 0 {
			m.diffViewport.Height = 0
		}
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
		if len(m.diffLines) > 0 {
			m.diffCursor = min(m.diffCursor+1, len(m.diffLines)-1)
		}
		if m.diffCursor >= m.diffViewport.YOffset+m.diffViewport.Height {
			m.diffViewport.YOffset = m.diffCursor - m.diffViewport.Height + 1
		}
		if m.diffViewport.YOffset < 0 {
			m.diffViewport.YOffset = 0
		}
		return m, nil

	case "k", "up":
		m.diffCursor = max(m.diffCursor-1, 0)
		if m.diffCursor < m.diffViewport.YOffset {
			m.diffViewport.YOffset = m.diffCursor
		}
		if m.diffViewport.YOffset < 0 {
			m.diffViewport.YOffset = 0
		}
		return m, nil

	case "ctrl+d":
		m.diffViewport.ScrollDown(m.diffViewport.Height / 2)
		if len(m.diffLines) > 0 {
			m.diffCursor = max(m.diffViewport.YOffset, min(m.diffViewport.YOffset+m.diffViewport.Height-1, m.diffCursor))
			m.diffCursor = max(0, min(len(m.diffLines)-1, m.diffCursor))
		}
		return m, nil

	case "ctrl+u":
		m.diffViewport.ScrollUp(m.diffViewport.Height / 2)
		if len(m.diffLines) > 0 {
			m.diffCursor = max(m.diffViewport.YOffset, min(m.diffViewport.YOffset+m.diffViewport.Height-1, m.diffCursor))
			m.diffCursor = max(0, min(len(m.diffLines)-1, m.diffCursor))
		}
		return m, nil

	case "ctrl+f":
		// Page down (forward)
		m.diffViewport.ScrollDown(m.diffViewport.Height)
		if len(m.diffLines) > 0 {
			m.diffCursor = max(m.diffViewport.YOffset, min(m.diffViewport.YOffset+m.diffViewport.Height-1, m.diffCursor))
			m.diffCursor = max(0, min(len(m.diffLines)-1, m.diffCursor))
		}
		return m, nil

	case "ctrl+b":
		// Page up (backward)
		m.diffViewport.ScrollUp(m.diffViewport.Height)
		if len(m.diffLines) > 0 {
			m.diffCursor = max(m.diffViewport.YOffset, min(m.diffViewport.YOffset+m.diffViewport.Height-1, m.diffCursor))
			m.diffCursor = max(0, min(len(m.diffLines)-1, m.diffCursor))
		}
		return m, nil

	case "g":
		m.diffViewport.GotoTop()
		m.diffCursor = 0
		return m, nil

	case "G":
		m.diffViewport.GotoBottom()
		if len(m.diffLines) > 0 {
			m.diffCursor = len(m.diffLines) - 1
		}
		return m, nil

	case "e":
		// Open file in external editor
		if m.cursor < len(m.files) {
			return m, openInEditor(m.files[m.cursor].Name)
		}
		return m, nil

	case "r":
		// Refresh - reload files and diff
		return m, loadFiles(m.options)
	}

	return m, nil
}

// Helper methods for dimensions
func (m Model) getListWidth() int {
	if m.layout == LayoutHorizontal {
		return m.listWidth
	}
	return m.width
}

func (m Model) getDiffWidth() int {
	if m.layout == LayoutHorizontal {
		return m.width - m.listWidth
	}
	return m.width
}

func (m Model) getContentHeight() int {
	// Reserve 2 lines for top header and bottom bar
	return m.height - 2
}

func (m Model) getWorkspaceHeaderHeight() int {
	return 3
}

// openInEditor opens the given file in $EDITOR
func openInEditor(filename string) tea.Cmd {
	return func() tea.Msg {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		cmd := exec.Command(editor, filename)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Suspend bubbletea, run editor, resume
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			return editorFinishedMsg{err: err}
		})
	}
}

type editorFinishedMsg struct {
	err error
}
