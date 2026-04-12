# Code Review in Vim

This repository contains a Vim plugin that facilitates code review directly within the Vim editor. It is based on tpope/vim-fugitive. It allows user to view changes for staged files or whole branch compared to another branch.

# Usage

Use `:Review` command to start code review. By default, it compares the current branch with its origin branch. You can specify a different branch by providing it as an argument, e.g., `:Review develop`. The plugin will open a new tab with split windows, left and right, the left window show all changed files on current branch compared to the specified branch, and the right window shows the diff of the selected file.

Use `:Review!` to review staged changes only.
