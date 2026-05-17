// internal/ui/app.go
package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lkebin/review/internal/diff"
	"github.com/lkebin/review/internal/git"
	"github.com/lkebin/review/internal/highlight"
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

	// Cursor positioner for search mode (nil in tests)
	cursorPos *cursorPositioner

	// State
	files       []FileInfo
	focus       FocusType
	listWidth   int
	showHelp    bool
	helpOffset  int
	searchMode  bool
	searchQuery string
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
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		// Esc clears an active (confirmed) search query.
		if m.searchQuery != "" && msg.String() == "esc" {
			m.searchQuery = ""
			return m, nil
		}
		if m.showHelp {
			max := m.helpMaxOffset()
			switch msg.String() {
			case "j", "down":
				if m.helpOffset < max {
					m.helpOffset++
				}
			case "k", "up":
				if m.helpOffset > 0 {
					m.helpOffset--
				}
			case "ctrl+d":
				m.helpOffset += m.height / 4
				if m.helpOffset > max {
					m.helpOffset = max
				}
			case "ctrl+u":
				m.helpOffset -= m.height / 4
				if m.helpOffset < 0 {
					m.helpOffset = 0
				}
			default:
				m.showHelp = false
				m.helpOffset = 0
			}
			return m, nil
		}
		return m.handleAction(m.keys.HandleKey(msg, m.focus))
	}
	return m, nil
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.searchMode = false
		m.cursorPos.clearCol()
		model, cmd := m.doSearch(true) // jump to first match
		return model, tea.Batch(tea.HideCursor, cmd)
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		m.cursorPos.clearCol()
		return m, tea.HideCursor
	case "backspace", "ctrl+h":
		if len(m.searchQuery) > 0 {
			runes := []rune(m.searchQuery)
			m.searchQuery = string(runes[:len(runes)-1])
		}
		m.cursorPos.setCol(searchPromptCol(m.searchQuery))
	default:
		if msg.Type == tea.KeyRunes {
			m.searchQuery += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			m.searchQuery += " "
		}
		m.cursorPos.setCol(searchPromptCol(m.searchQuery))
	}
	return m, nil
}

// searchPromptCol returns the 1-indexed terminal column immediately after "/query".
func searchPromptCol(query string) int32 {
	return int32(2 + lipgloss.Width(query))
}

// doSearch performs a next/prev search jump on the currently focused panel.
func (m Model) doSearch(forward bool) (tea.Model, tea.Cmd) {
	if m.searchQuery == "" {
		return m, nil
	}
	if m.focus == FocusList {
		var moved bool
		if forward {
			moved = m.fileList.SearchNext(m.searchQuery)
		} else {
			moved = m.fileList.SearchPrev(m.searchQuery)
		}
		if moved {
			return m, loadDiffCmd(m.opts, m.fileList.SelectedFile().Name)
		}
	} else {
		if forward {
			m.diffView.Viewport().SearchNext(m.searchQuery)
		} else {
			m.diffView.Viewport().SearchPrev(m.searchQuery)
		}
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
		m.helpOffset = 0
		return m, nil
	case ActionSearchOpen:
		m.searchMode = true
		m.searchQuery = ""
		m.cursorPos.setCol(searchPromptCol(""))
		return m, tea.ShowCursor

	// Focus
	case ActionFocusLeft:
		m.focus = FocusList
		return m, nil
	case ActionFocusRight:
		m.focus = FocusDiff
		return m, nil
	case ActionFocusToggle:
		if m.focus == FocusList {
			m.focus = FocusDiff
		} else {
			m.focus = FocusList
		}
		return m, nil
	case ActionEnter:
		if m.focus == FocusList {
			m.focus = FocusDiff
		}
		return m, nil

	// Panel resize
	case ActionGrowPanel:
		maxListWidth := max(m.width/2, 10)
		if m.listWidth < maxListWidth {
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
		if m.focus == FocusList {
			contentHeight := m.height - 2
			prevCursor := m.fileList.Cursor()
			m.fileList.PageDown(contentHeight)
			if m.fileList.Cursor() != prevCursor {
				return m, loadDiffCmd(m.opts, m.fileList.SelectedFile().Name)
			}
		} else {
			m.diffView.Viewport().PageDown()
		}
		return m, nil
	case ActionPageUp:
		if m.focus == FocusList {
			contentHeight := m.height - 2
			prevCursor := m.fileList.Cursor()
			m.fileList.PageUp(contentHeight)
			if m.fileList.Cursor() != prevCursor {
				return m, loadDiffCmd(m.opts, m.fileList.SelectedFile().Name)
			}
		} else {
			m.diffView.Viewport().PageUp()
		}
		return m, nil
	case ActionNextHunk:
		if m.searchQuery != "" {
			return m.doSearch(true)
		}
		if m.focus == FocusDiff {
			m.diffView.Viewport().NextHunk()
		}
		return m, nil
	case ActionPrevHunk:
		if m.searchQuery != "" {
			return m.doSearch(false)
		}
		if m.focus == FocusDiff {
			m.diffView.Viewport().PrevHunk()
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) resizeComponents() {
	contentHeight := max(m.height-2, 0) // status bar + command line
	diffWidth := max(m.width-m.listWidth, 1)
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

	contentHeight := max(m.height-2, 0)

	// File list
	listView := m.fileList.Render(m.listWidth, contentHeight, m.theme)

	// Diff view
	diffContent := m.diffView.Render()

	// Compose horizontally
	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, diffContent)

	// Status bar: always shows branch / file info
	selected := m.fileList.SelectedFile()
	statusBar := RenderStatusBar(
		m.opts.Target, len(m.files),
		m.currentFile, selected.Added, selected.Removed,
		m.width, m.theme,
	)

	// Command line: search input or blank
	cmdLine := RenderCmdLine(m.searchQuery, m.width, m.searchMode)

	return body + "\n" + statusBar + "\n" + cmdLine
}

func (m Model) helpLines() []string {
	targetLine := "  target: " + m.opts.Target
	if m.opts.Staged {
		targetLine = "  target: --staged (index vs HEAD)"
	}
	return strings.Split(`Key Bindings

  Navigation
  j / k        Move cursor down / up
  gg / G       Jump to top / bottom
  Ctrl+D / U   Half page down / up  (diff only)
  Ctrl+F / B   Page down / up
  n / N        Next / previous hunk  (diff only)

  Panels
  Tab          Toggle focus between panels
  Enter        Focus diff view
  > / <        Grow / shrink file list panel

  Search
  /            Open search (searches current panel)
  n / N        Next / previous match (falls back to hunk nav in diff)
  Enter        Confirm search and jump to first match
  Esc          Cancel search and clear query

  Other
  ?            Toggle this help
  q            Quit

Supported targets  (review --help for examples)

  HEAD          working tree vs HEAD (default)
  HEAD~N        N commits ago
  <branch>      working tree vs branch
  <commit>      working tree vs commit
  a..b          diff between two refs
  --staged      staged changes

Current session
`+targetLine, "\n")
}

func (m Model) helpMaxOffset() int {
	// Border (2) + padding top/bottom (2) + top indicator (1) + bottom indicator (1)
	overhead := 6
	innerHeight := max(m.height-overhead, 3)
	return max(len(m.helpLines())-innerHeight, 0)
}

func (m Model) renderHelp() string {
	lines := m.helpLines()
	maxOffset := m.helpMaxOffset()

	offset := min(m.helpOffset, maxOffset)
	overhead := 6
	innerHeight := max(m.height-overhead, 3)
	end := min(offset+innerHeight, len(lines))
	visible := strings.Join(lines[offset:end], "\n")

	var topIndicator, bottomIndicator string
	if offset > 0 {
		topIndicator = "  ▲ j/k to scroll\n"
	} else {
		topIndicator = "\n"
	}
	if offset < maxOffset {
		bottomIndicator = "\n  ▼ more"
	} else {
		bottomIndicator = "\n\n  Press any key to close"
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		style.Render(topIndicator+visible+bottomIndicator))
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
	stats, _ := git.GetFileStats(gopts) // best-effort; missing stats show 0
	result := make([]FileInfo, len(files))
	for i, f := range files {
		added, removed := 0, 0
		if s, ok := stats[f.Name]; ok {
			added = s[0]
			removed = s[1]
		}
		result[i] = FileInfo{
			Status:  f.Status,
			Name:    f.Name,
			Added:   added,
			Removed: removed,
		}
	}
	return result, nil
}
