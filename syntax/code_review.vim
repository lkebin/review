" Vim syntax file
" Language: Code Review Diff
" Maintainer: vim-code-review

if exists("b:current_syntax")
  finish
endif

" Line numbers prefix (10 chars: "  42   45 " or "          ")
syntax match codeReviewLineNr /^\s*\d*\s\+\d*\s/ contained

" Hunk headers: line numbers followed by @@ ... @@
syntax match codeReviewHunk /^\s\{10}@@.*@@.*$/

" Added lines: line numbers followed by +
syntax match codeReviewAdded /^\s*\d*\s\+\d*\s+.*$/ contains=codeReviewLineNr

" Removed lines: line numbers followed by -
syntax match codeReviewRemoved /^\s*\d*\s\+\d*\s-.*$/ contains=codeReviewLineNr

" Context lines (space after line numbers)
syntax match codeReviewContext /^\s*\d\+\s\+\d\+\s .*$/ contains=codeReviewLineNr

" Link to standard diff highlight groups (same as vim-fugitive)
highlight! link codeReviewAdded Added
highlight! link codeReviewRemoved Removed
highlight! link codeReviewHunk Statement
highlight! link codeReviewLineNr LineNr
highlight! link codeReviewContext Normal

let b:current_syntax = "code_review"
