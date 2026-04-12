" Vim syntax file
" Language: Code Review File List
" Maintainer: vim-code-review

if exists("b:current_syntax")
  finish
endif

" File paths (must be defined before status flags for proper matching)
syntax match codeReviewPathA /[^A ]\S*/ contained
syntax match codeReviewPathM /[^M ]\S*/ contained
syntax match codeReviewPathD /[^D ]\S*/ contained
syntax match codeReviewPathR /[^R ]\S*/ contained
syntax match codeReviewPathC /[^C ]\S*/ contained

" Status flags at start of line (higher priority with nextgroup)
syntax match codeReviewStatusA /^A / contained nextgroup=codeReviewPathA
syntax match codeReviewStatusM /^M / contained nextgroup=codeReviewPathM
syntax match codeReviewStatusD /^D / contained nextgroup=codeReviewPathD
syntax match codeReviewStatusR /^R / contained nextgroup=codeReviewPathR
syntax match codeReviewStatusC /^C / contained nextgroup=codeReviewPathC

" File lines with status
syntax match codeReviewFileAdded /^A .*$/ contains=codeReviewStatusA,codeReviewPathA
syntax match codeReviewFileModified /^M .*$/ contains=codeReviewStatusM,codeReviewPathM
syntax match codeReviewFileDeleted /^D .*$/ contains=codeReviewStatusD,codeReviewPathD
syntax match codeReviewFileRenamed /^R .*$/ contains=codeReviewStatusR,codeReviewPathR
syntax match codeReviewFileCopied /^C .*$/ contains=codeReviewStatusC,codeReviewPathC

" Status flag colors (bold)
highlight default codeReviewStatusA ctermfg=Green guifg=#22863a cterm=bold gui=bold
highlight default codeReviewStatusM ctermfg=Yellow guifg=#b08800 cterm=bold gui=bold
highlight default codeReviewStatusD ctermfg=Red guifg=#b31d28 cterm=bold gui=bold
highlight default codeReviewStatusR ctermfg=Magenta guifg=#6f42c1 cterm=bold gui=bold
highlight default codeReviewStatusC ctermfg=Cyan guifg=#0366d6 cterm=bold gui=bold

" File path colors (normal weight, different color)
highlight default codeReviewPathA ctermfg=White guifg=#e1e4e8
highlight default codeReviewPathM ctermfg=White guifg=#e1e4e8
highlight default codeReviewPathD ctermfg=Gray guifg=#959da5
highlight default codeReviewPathR ctermfg=White guifg=#e1e4e8
highlight default codeReviewPathC ctermfg=White guifg=#e1e4e8

let b:current_syntax = "code_review_files"
