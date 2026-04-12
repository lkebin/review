# Copilot Instructions for vim-code-review

## Overview

Vim plugin for code review. Provides `:Review` command to view diffs in a split-window interface with real line numbers.

## Testing

```bash
# Run all tests
./test/run.sh

# Or directly with Vim
vim -u NONE -N -S test/run_tests.vim
```

## Architecture

### Plugin Structure (Standard Vim Layout)

- `plugin/code_review.vim` - Command definitions, loaded once on startup
- `autoload/code_review.vim` - Main implementation, lazy-loaded when commands are invoked
- `syntax/code_review.vim` - Syntax highlighting for diff view
- `syntax/code_review_files.vim` - Syntax highlighting for file list

### Key Functions

- `code_review#start(bang, ...)` - Entry point; creates review tab with file list (left) and diff viewer (right)
- `code_review#open_diff()` - Displays diff for selected file in right pane
- `code_review#close()` - Closes the review tab
- `code_review#complete(arglead, cmdline, cursorpos)` - Tab completion for git refs
- `s:calculate_line_numbers(output)` - Parses hunk headers to calculate real line numbers
- `s:prepend_line_numbers(output, line_numbers)` - Prepends line numbers to diff content

### State Management

- `t:is_code_review` - Tab-local variable marking review tabs
- `b:code_review_target` - Buffer-local variable storing the comparison target (branch/HEAD)

## Conventions

### Vimscript Style

- Use `abort` on all functions to stop on errors
- Use `silent!` for commands that may fail harmlessly
- Prefix autoload functions with the filename: `code_review#functionname()`
- Prefix script-local functions with `s:`: `s:functionname()`
- Use `setlocal` for buffer-specific options, never `set`

### Buffer Setup Pattern

Review buffers use these standard options:
```vim
setlocal buftype=nofile
setlocal bufhidden=wipe
setlocal noswapfile
setlocal nomodifiable  " Set after content is populated
```

### Virtual Text Pattern

For features requiring virtual text, use this compatibility pattern:
```vim
if has('nvim-0.10')
  " Use nvim_buf_set_extmark() with virt_text_pos='inline'
else
  " Prepend line numbers to buffer content (Vim fallback)
endif
```

### Requirements

- Vim 9.0+ or Neovim
- Git (uses direct git commands, no vim-fugitive dependency)

### Compatibility

- Neovim 0.10+: Uses inline virtual text for line numbers
- Vim / older Neovim: Prepends line numbers to buffer content
