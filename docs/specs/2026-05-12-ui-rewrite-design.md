# UI Layer Rewrite Design

## Overview

Rewrite `internal/ui` to improve key bindings (Vim-consistent), layout (clean and practical), diff rendering (fixed line number column, inline diff highlighting), and syntax highlighting performance (Chroma caching). The git, diff, and highlight layers remain unchanged.

## Scope (MVP)

**In scope:**
- Layout: left file list + right single-file diff view + bottom status bar
- Key bindings: Vim prefix key state machine (`Ctrl+W`, `gg`)
- Diff rendering: fixed-width line number column (dynamic per file), content wrapping without invading line number area
- Cursorline: full-row highlight in diff view, `j/k` moves cursor
- Inline diff: character-level highlighting for adjacent add/remove line pairs
- Syntax highlighting: Chroma whole-file tokenize + cache per file
- Visual style: plain 256-color scheme, centralized `Theme` struct

**Out of scope (future):**
- `/` search within diff
- `e` to open file in `$EDITOR`
- Staged/unstaged toggle in UI
- Dynamic context line expansion
- Custom color scheme configuration (Theme struct is ready, loading config is deferred)

## Architecture

### Preserved Layers

- `internal/git` — git command wrapper
- `internal/diff` — unified diff parser (DiffLine, LineType, etc.)
- `internal/highlight` — Chroma-based syntax highlighter

### Rewritten Layer: `internal/ui`

```
internal/ui/
├── app.go           # Top-level Model, composes sub-components, global event dispatch
├── keymap.go        # Key mapping + prefix key state machine → semantic Actions
├── filelist.go      # Left panel: flat file list with status badges
├── diffview.go      # Right panel: single-file diff with cursorline
├── lineno.go        # Line number column rendering (dynamic width calculation)
├── inlinediff.go    # Character-level inline diff (Myers diff on adjacent +/- pairs)
├── styles.go        # All lipgloss styles, Theme struct, plain 256-color scheme
├── statusbar.go     # Bottom status bar
└── viewport.go      # Custom viewport (replaces bubbles/viewport)
```

## Key Bindings

### Prefix Key State Machine

States:
- `Normal` — waiting for input
- `WaitingCtrlW` — received `Ctrl+W`, waiting for direction key
- `WaitingG` — received `g`, waiting for second `g`

Timeout: 500ms, reverts to `Normal` if no follow-up key.

`keymap.go` exposes `HandleKey(msg tea.KeyMsg) Action`. `app.go` dispatches Actions to sub-components. Key parsing is fully decoupled from business logic.

### Key Map

| Key | Focus: file list | Focus: diff view |
|---|---|---|
| `j` / `k` | Move selection, load diff for selected file | Move cursorline up/down |
| `Enter` | Switch focus to diff view | — |
| `Ctrl+W h` | — | Focus to file list |
| `Ctrl+W l` | Focus to diff view | — |
| `Ctrl+W >` / `<` | Adjust panel width | Adjust panel width |
| `gg` | Jump to first file | Jump to top of diff |
| `G` | Jump to last file | Jump to bottom of diff |
| `Ctrl+D` / `Ctrl+U` | — | Half-page scroll |
| `Ctrl+F` / `Ctrl+B` | — | Full-page scroll |
| `n` / `N` | — | Jump to next/previous hunk |
| `q` | Quit | Quit |
| `?` | Show help | Show help |

## Layout

```
┌─ file list ──────────┐┌─ diff view ──────────────────────┐
│ M src/utils.go       ││  12   15  │+import "fmt"         │
│ A src/helper.go      ││  13   16  │ func main() {        │
│ M internal/parser.go ││      17   │+    fmt.Println()    │
│ D internal/old.go    ││  14       │-    println()        │
│ R config.yaml        ││  15   18  │ }                    │
│                      ││                                  │
├──────────────────────┤├──────────────────────────────────┤
│ main │ 5 files       ││ src/utils.go  +15 -3             │
└──────────────────────┘└──────────────────────────────────┘
```

### File List

- Flat list with full file paths (no directory grouping)
- Status badge per file: `M` (yellow), `A` (green), `D` (red), `R` (purple), `C` (cyan)
- Selected line highlight (consistent with diff view cursorline style)
- Default width: max file path length + padding, capped at 50 columns
- `Ctrl+W >` / `<` to adjust width

### Diff View

- Displays diff for the currently selected file
- Line number column + separator `│` + content column
- Bottom shows current file name and +/- stats

## Diff Rendering

### Line Number Column

- Width dynamically calculated per file based on max line number
  - e.g., max line 230 → 3 digits, column format: `%3d %3d │` (total 8 chars)
  - max line 12000 → 5 digits, column format: `%5d %5d │` (total 12 chars)
- Fixed background color (color 233), not affected by add/remove line colors
- Hunk headers (`@@...@@`): line number column shows `··`
- Lines with no old/new number: show blank spaces in that position

### Content Column

- Added lines: dark green background (color 22) spanning full content width
- Removed lines: dark red background (color 52) spanning full content width
- Context lines: no background
- Hunk headers: gray foreground (color 140), no background

### Content Wrapping

- When content exceeds available width, wrap within content column
- Continuation lines: line number column shows blank space, content indented with `│` marker
- Line number area is never invaded by wrapped content

### Cursorline

- Full-row background highlight from line number column to end of content
- When cursor is on added line: brighter green background (color 28)
- When cursor is on removed line: brighter red background (color 88)
- When cursor is on context line: dark gray background (color 236)

## Inline Diff (Character-Level)

### Pairing Rule

Consecutive removed lines followed by consecutive added lines are paired 1:1 in order. Unpaired lines get no inline diff treatment.

### Algorithm

Myers diff on the character sequences of each paired line to identify changed spans.

### Rendering

- Within a removed line, deleted characters: deeper red background (color 88)
- Within an added line, inserted characters: deeper green background (color 28)

Example:
```
│  14       │-    println("hello")         │  ← "hello" in deeper red
│      17   │+    fmt.Println("hello")     │  ← "fmt.P" in deeper green
```

## Syntax Highlighting

### Optimization

Current problem: `highlight.HighlightDiffLine()` is called per-line, creating a new Chroma lexer/iterator each time, causing stuttering.

New approach — add a new method to `internal/highlight`:
1. `TokenizeFile(filename string, lines []string) [][]Token` — accepts the filename (for lexer detection) and all code lines (prefixes stripped)
2. Internally: concatenate lines, call Chroma tokenize once, split result by line
3. Returns per-line token slices, cached by the caller (diffview component)
4. The existing `HighlightDiffLine()` method remains for backward compatibility but is not used by the new UI

### Integration with Diff Colors

Syntax highlighting colors apply to foreground text. Diff line background (add/remove/context) and inline diff background are independent layers. The visual stack:
1. Line background (add green / remove red / context none)
2. Inline diff background (deeper green / deeper red for changed characters)
3. Syntax highlighting foreground (Chroma token colors)
4. Cursorline (shifts background one shade brighter)

## Custom Viewport

### Why Not `bubbles/viewport`

`bubbles/viewport` treats content as a single string block. We need:
1. Fixed line number column that isn't affected by content wrapping
2. Cursorline tracking on logical lines (viewport only knows text offset)
3. Continuation lines from wrapping must not occupy line number column

### Design

```go
type Viewport struct {
    width, height int    // visible area dimensions
    offset        int    // first visible logical line index
    cursor        int    // cursorline logical line index
    lines         []Line // all logical lines (diff lines)
}

type Line struct {
    LeftNo   string   // old file line number (empty = not present)
    RightNo  string   // new file line number
    Content  string   // highlighted content (with ANSI codes)
    LineType LineType // Added/Removed/Context/HunkHeader
}
```

### Scrolling

- `j/k`: move cursor, adjust offset when cursor leaves visible area (Vim `scrolloff` behavior, keep a few lines of context around cursor)
- `Ctrl+D/U`: move cursor and offset by half page
- `Ctrl+F/B`: full page
- `gg/G`: jump to first/last line

## Visual Style

### Color Scheme (256-color)

| Element | Foreground | Background | Notes |
|---|---|---|---|
| Normal text | 252 (light gray) | terminal default | No forced background |
| Added line | 252 | 22 (dark green) | Soft, not glaring |
| Removed line | 252 | 52 (dark red) | Same |
| Inline diff (add) | — | 28 (bright green) | Over add line background |
| Inline diff (del) | — | 88 (bright red) | Over del line background |
| Cursorline | — | 236 (dark gray) | Mixed color on add/del lines |
| Cursorline + added | — | 28 | One shade brighter than normal add |
| Cursorline + removed | — | 88 | One shade brighter than normal del |
| Line number column | 240 (gray) | 233 (very dark gray) | Fixed background, visual separation |
| Line number separator `│` | 238 (dim gray) | 233 | Subtle |
| Hunk header `@@` | 140 (purple-gray) | terminal default | Unobtrusive |
| File status M | 178 (yellow) | — | — |
| File status A | 40 (green) | — | — |
| File status D | 167 (red) | — | — |
| File status R | 133 (purple) | — | — |
| File status C | 73 (cyan) | — | — |
| File list selected | 16 (black) | 75 (blue) | Clear selection |
| Status bar | 252 | 236 | Simple bottom bar |

### Theme Struct

All colors defined in a `Theme` struct in `styles.go`. MVP hardcodes one theme. Future: load from config file to override.

### De-decoration

- No all-caps text
- No brand labels
- No rounded fancy borders
- Headers show only necessary information
- Borders use simple line-drawing characters

## Data Flow

```
app.Init()
  → git.GetFiles() → []FileInfo
  → display file list

User selects file (j/k in file list):
  → git.GetDiff(file) → []DiffLine
  → highlight.TokenizeFile(diffLines) → cached tokens (one Chroma call)
  → inlinediff.Compute(diffLines) → mark character-level changes
  → viewport renders visible lines with line numbers + highlighting + inline diff
```
