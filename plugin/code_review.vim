" plugin/code_review.vim

if exists('g:loaded_code_review')
  finish
endif
let g:loaded_code_review = 1

command! -bang -nargs=? -complete=customlist,code_review#complete Review call code_review#start(<bang>0, <f-args>)
command! ReviewClose call code_review#close()
