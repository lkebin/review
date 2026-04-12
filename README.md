# review

A terminal-based code review tool written in Go.

## Features

- **Split-view interface**: File list on the left, diff content on the right
- **Vim-style keybindings**: j/k navigation, q to quit
- **Syntax highlighting**: Powered by Chroma for 181+ languages
- **Line numbers**: Shows both old and new line numbers
- **Layout switching**: Toggle between horizontal and vertical split
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

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down (file list or diff scroll) |
| `k` / `↑` | Move up (file list or diff scroll) |
| `h` / `←` | Focus file list |
| `l` / `→` | Focus diff / Toggle layout |
| `Tab` | Switch focus between panels |
| `Enter` | View diff for selected file |
| `e` | Open file in $EDITOR |
| `r` | Refresh diff |
| `q` | Quit |

## Layout

### Horizontal (default)
```
┌─────────────────┬─────────────────────────────────────┐
│ M  src/main.go  │@@ -10,7 +10,8 @@ func main() {      │
│ A  src/util.go  │ 10  10   // Setup                   │
│ D  README.md    │ 11     - fmt.Println("old")         │
│                 │    11 + fmt.Println("new")          │
│                 │ 12  12   // Run                     │
├─────────────────┴─────────────────────────────────────┤
│ current > main | Files: 3 | 1/3 | Layout: H | [q]uit  │
└───────────────────────────────────────────────────────┘
```

### Vertical (press `l` to toggle)
```
┌───────────────────────────────────────────────────────┐
│ M  src/main.go                                        │
│ A  src/util.go                                        │
│ D  README.md                                          │
├───────────────────────────────────────────────────────┤
│@@ -10,7 +10,8 @@ func main() {                       │
│ 10  10   // Setup                                     │
├───────────────────────────────────────────────────────┤
│ current > main | Files: 3 | 1/3 | Layout: V | [q]uit  │
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
