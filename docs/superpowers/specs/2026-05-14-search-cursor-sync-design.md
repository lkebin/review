# Design: Search Mode Cursor Sync with Terminal

**Date:** 2026-05-14

## Problem

The search bar currently renders a hardcoded `▌` character as a fake input cursor when `typing=true`. This character has a fixed shape that never matches the terminal's actual cursor configuration (beam, block, underline, blinking, etc.), creating a visual inconsistency.

## Goal

When the user enters search mode, show the real terminal cursor at the end of the search query input, so the cursor shape and behavior automatically match the terminal's configuration.

## Approach

Use ANSI cursor save/restore sequences within `RenderSearchBar` to park the terminal cursor at the correct position, combined with `tea.ShowCursor()` / `tea.HideCursor()` commands on mode transitions.

## Changes

### `internal/ui/statusbar.go` — `RenderSearchBar`

When `typing=true`:
- Remove the `▌` character.
- After rendering `/query`, insert `\033[s` (ANSI save cursor position).
- Render the gap and `[panel]` indicator as usual.
- Append `\033[u` (ANSI restore cursor position) at the very end of the returned string.

Result: bubbletea writes the full status bar line, then the restore sequence leaves the terminal cursor sitting just after the query text.

When `typing=false`: no change — no cursor character, pure display.

### `internal/ui/app.go` — mode transitions

| Event | Added command |
|---|---|
| `ActionSearchOpen` | `tea.ShowCursor` |
| Enter confirms search (`handleSearchKey`) | `tea.HideCursor` |
| Esc cancels search (`handleSearchKey`) | `tea.HideCursor` |

All other states (file list, diff view, confirmed query display) remain cursor-hidden.

## Out of Scope

- Changing cursor shape on mode entry (e.g. switching to beam on insert). Not needed.
- Blinking cursor management. The terminal's own setting governs this.
- Using `charmbracelet/bubbles/cursor` component. Unnecessary complexity.

## Risks

- A small number of terminals do not implement ANSI save/restore (`\033[s` / `\033[u`). In practice these are vanishingly rare in modern environments (xterm, iTerm2, kitty, alacritty, tmux all support it). Fallback: worst case the cursor appears at end of the status bar line rather than mid-line, which is acceptable.
