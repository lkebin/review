# Code Review in Vim

A Vim plugin for code review directly within the Vim editor. View changes for local modifications or compare branches in a split-window interface.

## Features

- Split-window interface: file list on left, diff on right
- Real line numbers (old and new) displayed in diff view
- File status indicators (A/M/D/R/C) with syntax highlighting
- Tab completion for git refs (branches, tags)
- Works with Vim 9.0+ and Neovim

## Requirements

- Vim 9.0+ or Neovim
- Git

## Usage

- `:Review <branch>` - Compare current branch with specified branch
- `:Review!` - Review local changes (staged and unstaged vs HEAD)
- `:ReviewClose` - Close the review tab

In the file list window:
- `<CR>` - Open diff for selected file
- `q` - Close review tab
