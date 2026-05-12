// internal/ui/app.go
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kbliu/review/internal/diff"
	"github.com/kbliu/review/internal/git"
	"github.com/kbliu/review/internal/highlight"
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
	opts   Options
	theme  Theme
	keys   *KeyMapper
	width  int
	height int

	// Components
	fileList *FileList
	diffView *DiffView

	// State
	files       []FileInfo
	focus       FocusType
	listWidth   int
	showHelp    bool
	loading     bool
	err         error
	currentFile string
	highlighter *highlight.SimpleHighlighter
}

// NewModel creates a new Model.
func NewModel(opts Options) Model {
	theme := DefaultTheme()
	return Model{
		opts:        opts,
		theme:       theme,
		keys:        NewKeyMapper(),
		fileList:    NewFileList(nil),
		diffView:    NewDiffView(80, 24, theme),
		focus:       FocusList,
		loading:     true,
		listWidth:   30,
		highlighter: highlight.New("github"),
	}
}

// --- BubbleTea messages ---

type loadFilesMsg struct {
	files []FileInfo
	err   error
}

type loadDiffMsg struct {
	file  string
	lines []diff.Line
	err   error
}

// --- Init ---

func (m Model) Init() tea.Cmd {
	return loadFilesCmd(m.opts)
}

func loadFilesCmd(opts Options) tea.Cmd {
	return func() tea.Msg {
		files, err := loadFileList(opts)
		return loadFilesMsg{files: files, err: err}
	}
}

func loadDiffCmd(opts Options, file string) tea.Cmd {
	return func() tea.Msg {
		gopts := gitOptions(opts)
		content, err := git.GetDiff(gopts, file)
		if err != nil {
			return loadDiffMsg{file: file, err: err}
		}
		lines := diff.Parse(content)
		return loadDiffMsg{file: file, lines: lines}
	}
}

// --- Update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeComponents()
		return m, nil

	case loadFilesMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.files = msg.files
		m.fileList.SetFiles(m.files)
		m.listWidth = m.fileList.CalcWidth()
		m.resizeComponents()
		if len(m.files) > 0 {
			return m, loadDiffCmd(m.opts, m.files[0].Name)
		}
		return m, nil

	case loadDiffMsg:
		m.currentFile = msg.file
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.diffView.LoadFile(msg.lines, msg.file, m.highlighter)
		return m, nil

	case tea.KeyMsg:
		// If help is shown, any key closes it
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		return m.handleAction(m.keys.HandleKey(msg, m.focus))
	}
	return m, nil
}

func (m Model) handleAction(action Action) (tea.Model, tea.Cmd) {
	switch action {
	case ActionNone:
		return m, nil
	case ActionQuit:
		return m, tea.Quit
	case ActionHelp:
		m.showHelp = true
		return m, nil

	// Focus
	case ActionFocusLeft:
		m.focus = FocusList
		return m, nil
	case ActionFocusRight:
		m.focus = FocusDiff
		return m, nil
	case ActionEnter:
		if m.focus == FocusList {
			m.focus = FocusDiff
		}
		return m, nil

	// Panel resize
	case ActionGrowPanel:
		if m.listWidth < 60 {
			m.listWidth += 2
			m.resizeComponents()
		}
		return m, nil
	case ActionShrinkPanel:
		if m.listWidth > 10 {
			m.listWidth -= 2
			m.resizeComponents()
		}
		return m, nil

	// Cursor movement
	case ActionCursorDown:
		if m.focus == FocusList {
			prevCursor := m.fileList.Cursor()
			m.fileList.CursorDown()
			if m.fileList.Cursor() != prevCursor {
				return m, loadDiffCmd(m.opts, m.fileList.SelectedFile().Name)
			}
		} else {
			m.diffView.Viewport().CursorDown()
		}
		return m, nil

	case ActionCursorUp:
		if m.focus == FocusList {
			prevCursor := m.fileList.Cursor()
			m.fileList.CursorUp()
			if m.fileList.Cursor() != prevCursor {
				return m, loadDiffCmd(m.opts, m.fileList.SelectedFile().Name)
			}
		} else {
			m.diffView.Viewport().CursorUp()
		}
		return m, nil

	case ActionTop:
		if m.focus == FocusList {
			m.fileList.GotoTop()
			if len(m.files) > 0 {
				return m, loadDiffCmd(m.opts, m.fileList.SelectedFile().Name)
			}
		} else {
			m.diffView.Viewport().GotoTop()
		}
		return m, nil

	case ActionBottom:
		if m.focus == FocusList {
			m.fileList.GotoBottom()
			if len(m.files) > 0 {
				return m, loadDiffCmd(m.opts, m.fileList.SelectedFile().Name)
			}
		} else {
			m.diffView.Viewport().GotoBottom()
		}
		return m, nil

	case ActionHalfPageDown:
		m.diffView.Viewport().HalfPageDown()
		return m, nil
	case ActionHalfPageUp:
		m.diffView.Viewport().HalfPageUp()
		return m, nil
	case ActionPageDown:
		m.diffView.Viewport().PageDown()
		return m, nil
	case ActionPageUp:
		m.diffView.Viewport().PageUp()
		return m, nil
	case ActionNextHunk:
		m.diffView.Viewport().NextHunk()
		return m, nil
	case ActionPrevHunk:
		m.diffView.Viewport().PrevHunk()
		return m, nil
	}

	return m, nil
}

func (m *Model) resizeComponents() {
	contentHeight := m.height - 1 // 1 for status bar
	if contentHeight < 0 {
		contentHeight = 0
	}
	diffWidth := m.width - m.listWidth
	if diffWidth < 1 {
		diffWidth = 1
	}
	m.diffView.Resize(diffWidth, contentHeight)
}

// --- View ---

func (m Model) View() string {
	if m.loading && len(m.files) == 0 {
		return "Loading..."
	}
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit."
	}
	if len(m.files) == 0 {
		return "No changes found.\n\nPress q to quit."
	}
	if m.showHelp {
		return m.renderHelp()
	}

	contentHeight := m.height - 1
	if contentHeight < 0 {
		contentHeight = 0
	}

	// File list
	listView := m.fileList.Render(m.listWidth, contentHeight, m.theme)

	// Diff view
	diffContent := m.diffView.Render()

	// Compose horizontally
	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, diffContent)

	// Status bar
	selected := m.fileList.SelectedFile()
	bar := RenderStatusBar(
		m.opts.Target, len(m.files),
		m.currentFile, selected.Added, selected.Removed,
		m.width, m.theme,
	)

	return body + "\n" + bar
}

func (m Model) renderHelp() string {
	help := `Keyboard Shortcuts

  j/k          Navigate files / Move cursor in diff
  Enter        Switch focus to diff view
  Ctrl+W h/l   Switch focus between panels
  Ctrl+W >/<   Adjust panel width
  gg / G       Go to top / bottom
  Ctrl+D/U     Half page down / up
  Ctrl+F/B     Page down / up
  n / N        Next / previous hunk
  ?            Toggle this help
  q            Quit

Press any key to close...`

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		style.Render(help))
}

// --- Data bridge ---

func gitOptions(opts Options) git.Options {
	return git.Options{
		Target:       opts.Target,
		Staged:       opts.Staged,
		ContextLines: opts.ContextLines,
	}
}

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
