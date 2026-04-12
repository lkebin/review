# review

A terminal-based code review tool written in Go.

## Features

- **Split-view interface**: File list on the left, diff content on the right
- **File statistics**: Shows added/removed lines per file (+add/-del)
- **Vim-style keybindings**: j/k navigation, efficient shortcuts
- **Syntax highlighting**: Powered by Chroma for 180+ languages
- **Line numbers**: Shows both old and new line numbers
- **Word wrap**: Long lines wrap with continuation indicator
- **Layout switching**: Toggle between horizontal and vertical split
- **External editor**: Open files in $EDITOR from the tool
- **Zero configuration**: Works out of the box

## Installation

```bash
go install github.com/kbliu/review/cmd/review@latest
```

Or build from source:

```bash
git clone https://github.com/kbliu/review.git
cd review
go build ./cmd/review
```

## Usage

```bash
# Review local changes (vs HEAD)
review

# Review changes against a specific branch
review main

# Review staged changes
review --staged

# Adjust context lines
review -U 10 main
```

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down / Scroll down |
| `k` / `↑` | Move up / Scroll up |
| `Enter` | View diff for selected file |

### Window Management

| Key | Action |
|-----|--------|
| `Ctrl+W` | Switch focus between file list and diff |
| `h` / `H` | Shrink / Grow file list width |
| `L` (Shift+l) | Toggle horizontal/vertical layout |

### Diff Navigation

| Key | Action |
|-----|--------|
| `g` | Go to top of diff |
| `G` | Go to bottom of diff |
| `Ctrl+D` | Half page down |
| `Ctrl+U` | Half page up |
| `Ctrl+F` | Page forward (vim style) |
| `Ctrl+B` | Page backward (vim style) |

### Actions

| Key | Action |
|-----|--------|
| `e` | Open current file in $EDITOR |
| `r` | Refresh file list |
| `?` | Show help |
| `q` | Quit |

## Layout

### Horizontal (default)
```
┌─────────────────┬─────────────────────────────────────┐
│ A  main.go      │  1   1   package main               │
│ M  util.go      │  2   2                                │
│ D  old.go       │  3      -import "fmt"               │
│ (+5/-3)         │  4      +import "strings"           │
├─────────────────┴─────────────────────────────────────┤
│ current > main | Files: 3 | 1/3 | Focus: List | [?]   │
└───────────────────────────────────────────────────────┘
```

### Vertical (press `L` to toggle)
```
┌───────────────────────────────────────────────────────┐
│ A  main.go (+5/-3)                                    │
│ M  util.go (+10/-2)                                   │
│ D  old.go (+0/-50)                                    │
├───────────────────────────────────────────────────────┤
│  1   1   package main                                 │
│  3      -import "fmt"                                 │
│  4      +import "strings"                             │
├───────────────────────────────────────────────────────┤
│ current > main | Files: 3 | 1/3 | Focus: Diff | [?]   │
└───────────────────────────────────────────────────────┘
```

## Development

```bash
# Run tests
go test ./...

# Build
go build ./cmd/review

# Run
./review
```

## License

MIT
