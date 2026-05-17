# review

A terminal-based code review tool written in Go.

## Features

- **Split-view interface**: File list on the left, diff content on the right
- **File statistics**: Shows added/removed lines per file
- **Syntax highlighting**: Powered by Chroma for 180+ languages
- **Line numbers**: Shows both old and new line numbers side by side
- **Inline diff**: Character-level change highlighting within modified lines
- **Search**: Vim-style `/` search within the focused panel
- **Hunk navigation**: Jump between diff hunks with `n`/`N`
- **Adjustable panel width**: Resize the file list with `>` / `<`
- **Line wrapping**: Long lines wrap and preserve background colors

## Installation

```bash
go install github.com/lkebin/review/cmd/review@latest
```

Or build from source:

```bash
git clone https://github.com/lkebin/review.git
cd review
go build ./cmd/review
```

## Usage

```bash
# Review working tree vs HEAD (default)
review

# Review changes against a branch or commit
review main
review HEAD~3

# Review staged changes
review --staged

# Diff between two refs
review main..feature

# Adjust context lines
review -U 10 main
```

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `Ctrl+D` / `Ctrl+U` | Half page down / up |
| `Ctrl+F` / `Ctrl+B` | Page down / up |

### Panels

| Key | Action |
|-----|--------|
| `Tab` | Toggle focus between file list and diff |
| `Enter` | Focus diff view |
| `>` / `<` | Grow / shrink file list panel |

### Search

| Key | Action |
|-----|--------|
| `/` | Open search (searches focused panel) |
| `Enter` | Confirm search, jump to first match |
| `n` / `N` | Next / previous match (or hunk when no search) |
| `Esc` | Cancel search / clear query |

### Other

| Key | Action |
|-----|--------|
| `?` | Toggle help |
| `q` | Quit |

## Layout

```
┌──────────────────┬──────────────────────────────────────┐
│ M  build.sh      │ 19 21   . get_params.sh && rm ...    │
│ A  main.go       │ 21      -params=("-gu:gitusername"   │
│ D  old.go        │ 23      +params=("-gu:gitusername"   │
│ R  util.go       │ ·· ··  @@ -31,4 +33,10 @@           │
│                  │ 31 33   if [[ -z "$target" ]]; then  │
├──────────────────┴──────────────────────────────────────┤
│ main │ 4 files                       build.sh +11 -3    │
└─────────────────────────────────────────────────────────┘
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
