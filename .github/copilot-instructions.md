# Copilot Instructions for vim-code-review

## Overview

Vim plugin for code review built on top of vim-fugitive. Provides `:Review` command to view diffs in a split-window interface.

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

### Key Functions

- `code_review#start(bang, ...)` - Entry point; creates review tab with file list (left) and diff viewer (right)
- `code_review#open_diff()` - Displays diff for selected file in right pane
- `code_review#close()` - Closes the review tab
- `s:add_line_numbers(output)` - Adds virtual text line numbers to diff view
- `s:has_virtual_text()` - Checks for Vim 9.0+ or Neovim virtual text support

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
if has('nvim')
  " Use nvim_buf_set_extmark() with namespace
else
  " Use prop_type_add() / prop_add() for Vim 9.0+
endif
```

### Dependencies

Requires [vim-fugitive](https://github.com/tpope/vim-fugitive) - the plugin checks for `g:loaded_fugitive` before running.

### Compatibility

Line number virtual text requires Vim 9.0+ or Neovim. The feature degrades gracefully on older versions.
