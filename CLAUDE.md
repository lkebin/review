# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o review ./cmd/review/

# Run a review
./review

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a single package
go test ./internal/diff/...
go test ./internal/ui/...

# Run a single test
go test -run TestName ./internal/ui/...

# Install
go install github.com/lkebin/review/cmd/review@latest

# Update dependencies
go mod tidy
```

## Project Overview

Terminal-based code review tool (TUI) written in Go. It displays git diffs in a split-view interface with syntax highlighting, inline character-level diff, hunk navigation, and vim-style search.

## Architecture

### Entrypoint
- **cmd/review/main.go** — CLI entrypoint using `urfave/cli/v2`. Parses flags (`--staged`, `--unified`/`-U`) and target ref, calls `ui.Run(opts)`.

### Internal Packages

#### `internal/git` — Git integration
- `git.GetFiles(opts)` — runs `git diff --name-status` to list changed files
- `git.GetFileStats(opts)` — runs `git diff --numstat` for added/removed counts
- `git.GetDiff(opts, file)` — runs `git diff --no-color` for a single file's diff
- Wraps `os/exec` to call the system `git` binary directly. No git library dependency.

#### `internal/diff` — Diff parsing
- `diff.Parse(output)` — parses raw git diff output into `[]Line` with typed entries (LineContext, LineAdded, LineRemoved, LineHunkHeader). Tracks old/new line numbers from hunk headers.

#### `internal/highlight` — Syntax highlighting
- `highlight.New(styleName)` — creates a highlighter using Chroma. Style defaults to "github" (256-color).
- `highlight.TokenizeFile(filename, lines)` — batch-tokenizes code lines by detecting the lexer from the filename. Falls back to plain text if no lexer matches.

#### `internal/ui` — BubbleTea TUI (the bulk of the project)
- **app.go** — Top-level `Model` struct implementing `tea.Model`. Manages focus, search, help screen, and coordinates components.
- **ui.go** — `Options` struct and `Run()` function. Creates the `tea.Program` with altscreen and mouse support.
- **filelist.go** — Left panel: scrollable list of changed files with status badges (M/A/D/R/C).
- **diffview.go** — Right panel: wraps a `Viewport` and coordinates loading/rendering diffs.
- **viewport.go** — Scrollable content area with cursor, wrapping, search, and hunk navigation. Wraps long lines correctly (the ANSI-safe break logic is here).
- **inlinediff.go** — Character-level diff using common-prefix/suffix matching. `PairDiffLines` pairs removed→added lines 1:1 within a hunk.
- **lineno.go** — Formats old/new line number pairs for display.
- **statusbar.go** — Bottom status bar showing branch, file count, current file + stats.
- **cursor_positioner.go** — Wraps stdout to position the terminal cursor for search input in altscreen mode.
- **keymap.go** — Prefix-key state machine (`g`→`gg` for top, `G` for bottom). Translates key events to semantic `Action` values.
- **styles.go** — Theme struct with 256-color ANSI values and lipgloss style builders.

### Key Design Points

- **No external git library** — the tool invokes `git` via `os/exec` for all operations. The `git` package is the only point of contact with git.
- **Batch highlighting** — `TokenizeFile` processes all lines of a file in a single Chroma call for efficiency.
- **Inline diff** — within each hunk, consecutive `-` lines are paired with consecutive `+` lines. Changed character spans are computed via common-prefix/suffix and rendered with an emphasis background color.
- **Line wrapping** — uses lipgloss width and `findANSISafeBreak` to wrap ANSI-rendered content at character boundaries without breaking escape sequences.
- **TestMain** — `internal/ui/testmain_test.go` sets lipgloss to TrueColor profile for consistent test rendering.
- **No alt screen mock** — TUI tests that don't need terminal rendering operate on the model directly without starting a `tea.Program`.

### Superpowers
- **Git** — DO NOT help me create commit
