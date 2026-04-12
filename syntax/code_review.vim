" Vim syntax file
" Language: Code Review Diff
" Maintainer: vim-code-review

if exists("b:current_syntax")
  finish
endif

" Hunk headers: @@ -old,count +new,count @@
syntax match codeReviewHunk /^.*@@\s*-\d\+.*+\d\+.*@@.*$/

" Added lines: contain + after line numbers
syntax match codeReviewAdded /^.\{-}+.*$/ contains=codeReviewLineNr

" Removed lines: contain - after line numbers  
syntax match codeReviewRemoved /^.\{-}-.*$/ contains=codeReviewLineNr

" Line numbers at the start
syntax match codeReviewLineNr /^\s*\d*\s\+\d*\s\+/ contained

" Link to standard diff highlight groups from colorscheme
highlight! link codeReviewAdded DiffAdd
highlight! link codeReviewRemoved DiffDelete
highlight! link codeReviewHunk DiffChange
highlight! link codeReviewLineNr LineNr

let b:current_syntax = "code_review"
