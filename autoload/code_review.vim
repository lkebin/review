" autoload/code_review.vim

" Check if virtual text is supported
function! s:has_virtual_text() abort
  if has('nvim')
    return 1
  endif
  return has('textprop') && exists('*prop_add') && has('patch-9.0.0067')
endfunction

" Calculate line numbers for diff output, returns dict mapping buf_line -> [old_line, new_line]
function! s:calculate_line_numbers(output) abort
  let line_numbers = {}
  let old_line = 0
  let new_line = 0

  let buf_line = 0
  for line in a:output
    let buf_line += 1

    " Parse hunk header: @@ -old_start,old_count +new_start,new_count @@
    let hunk_match = matchlist(line, '^@@ -\(\d\+\),\?\(\d*\) +\(\d\+\),\?\(\d*\) @@')
    if !empty(hunk_match)
      let old_line = str2nr(hunk_match[1])
      let new_line = str2nr(hunk_match[3])
      let line_numbers[buf_line] = ['@@', '@@']
      continue
    endif

    " Determine line type and track line numbers
    let prefix = strpart(line, 0, 1)

    if prefix ==# '+'
      " Added line: only exists in new file
      let line_numbers[buf_line] = ['', new_line]
      let new_line += 1
    elseif prefix ==# '-'
      " Removed line: only exists in old file
      let line_numbers[buf_line] = [old_line, '']
      let old_line += 1
    elseif prefix ==# ' ' || (prefix !=# '\' && new_line > 0)
      " Context line: exists in both
      let line_numbers[buf_line] = [old_line, new_line]
      let old_line += 1
      let new_line += 1
    endif
  endfor

  return line_numbers
endfunction

" Add virtual text line numbers (Neovim)
function! s:add_virtual_text_nvim(line_numbers) abort
  let bufnr = bufnr('%')
  let ns_id = nvim_create_namespace('code_review_linenum')
  call nvim_buf_clear_namespace(bufnr, ns_id, 0, -1)

  " Check if inline virt_text is supported (Neovim 0.10+)
  let has_inline = has('nvim-0.10')

  for [lnum, nums] in items(a:line_numbers)
    let [old_num, new_num] = nums
    
    if old_num ==# '@@'
      let text = '         '
    else
      let old_str = old_num ==# '' ? '    ' : printf('%4d', old_num)
      let new_str = new_num ==# '' ? '    ' : printf('%4d', new_num)
      let text = old_str . ' ' . new_str . ' '
    endif

    if has_inline
      call nvim_buf_set_extmark(bufnr, ns_id, str2nr(lnum) - 1, 0, {
            \ 'virt_text': [[text, 'LineNr']],
            \ 'virt_text_pos': 'inline'
            \ })
    else
      " Older Neovim: use overlay at column 0 (will cover first chars)
      " Or use eol and accept it's at end
      call nvim_buf_set_extmark(bufnr, ns_id, str2nr(lnum) - 1, 0, {
            \ 'virt_text': [[text, 'LineNr']],
            \ 'virt_text_pos': 'overlay'
            \ })
    endif
  endfor
endfunction

" Add virtual text line numbers (Vim 9.0+)
function! s:add_virtual_text_vim(line_numbers) abort
  let bufnr = bufnr('%')
  
  silent! call prop_type_delete('code_review_linenum', {'bufnr': bufnr})
  call prop_type_add('code_review_linenum', {'bufnr': bufnr, 'highlight': 'LineNr'})

  for [lnum, nums] in items(a:line_numbers)
    let [old_num, new_num] = nums
    
    if old_num ==# '@@'
      let text = '         '
    else
      let old_str = old_num ==# '' ? '    ' : printf('%4d', old_num)
      let new_str = new_num ==# '' ? '    ' : printf('%4d', new_num)
      let text = old_str . ' ' . new_str . ' '
    endif

    " Use column 1 with text_align 'right' to place before first char
    " Or use text_padding_left on column 1
    call prop_add(str2nr(lnum), 1, {
          \ 'type': 'code_review_linenum',
          \ 'text': text,
          \ 'text_align': 'above',
          \ 'text_padding_left': 0,
          \ 'bufnr': bufnr
          \ })
  endfor
endfunction

" Prepend line numbers to buffer content (fallback for older Vim)
function! s:prepend_line_numbers(output, line_numbers) abort
  let result = []
  let buf_line = 0
  for line in a:output
    let buf_line += 1
    if has_key(a:line_numbers, buf_line)
      let [old_num, new_num] = a:line_numbers[buf_line]
      
      if old_num ==# '@@'
        call add(result, '          ' . line)
      else
        let old_str = old_num ==# '' ? '    ' : printf('%4d', old_num)
        let new_str = new_num ==# '' ? '    ' : printf('%4d', new_num)
        call add(result, old_str . ' ' . new_str . ' ' . line)
      endif
    else
      call add(result, '          ' . line)
    endif
  endfor
  return result
endfunction

" Add line numbers to the diff buffer (call AFTER setline)
function! s:add_line_numbers_virtual(line_numbers) abort
  setlocal nonumber norelativenumber
  
  if has('nvim')
    call s:add_virtual_text_nvim(a:line_numbers)
  elseif s:has_virtual_text()
    call s:add_virtual_text_vim(a:line_numbers)
  endif
endfunction

" Process diff output - returns [display_lines, line_numbers]
function! s:process_diff_output(output) abort
  let line_numbers = s:calculate_line_numbers(a:output)
  
  " Only Neovim 0.10+ supports inline virtual text at line start
  " For Vim and older Neovim, prepend to content
  if has('nvim-0.10')
    return [a:output, line_numbers]
  else
    return [s:prepend_line_numbers(a:output, line_numbers), {}]
  endif
endfunction

" Close review tab safely
function! s:close_review_tab() abort
  " Prevent recursive calls
  if exists('s:closing_tab') && s:closing_tab
    return
  endif
  let s:closing_tab = 1
  
  " Defer the close to avoid conflicts with other plugins
  if exists('*timer_start')
    call timer_start(0, {-> s:do_close_review_tab()})
  else
    call s:do_close_review_tab()
  endif
endfunction

function! s:do_close_review_tab() abort
  " Check if still in a code review tab and close it
  if exists('t:is_code_review')
    silent! tabclose
  endif
  let s:closing_tab = 0
endfunction

function! code_review#start(bang, ...) abort
  if !exists('g:loaded_fugitive')
    echoerr "vim-code-review requires tpope/vim-fugitive"
    return
  endif

  let target = a:0 > 0 ? a:1 : ''

  if empty(target)
    if a:bang
      let target = 'HEAD'
    else
      echo "Usage: :Review <branch> (or :Review! for local changes)"
      return
    endif
  endif

  let diff_cmd = 'git diff --name-only ' . shellescape(target)

  let files = systemlist(diff_cmd)

  if v:shell_error
    echoerr "Error running git diff: " . join(files, "\n")
    return
  endif

  if empty(files)
    echo "No changes found."
    return
  endif

  " Open new tab
  tabnew
  let t:is_code_review = 1

  " Setup the file list window (left side)
  " We use a vertical split.
  let list_buf_name = 'Code Review Files'
  silent! execute 'file ' . list_buf_name
  
  " Configure the buffer
  setlocal buftype=nofile
  setlocal bufhidden=wipe
  setlocal noswapfile
  setlocal nobuflisted
  setlocal nonumber
  setlocal norelativenumber
  setlocal cursorline
  setlocal statusline=[Code\ Review]\ Files
  
  " Save state variables in the buffer
  let b:code_review_target = target
  let b:code_review_tab = tabpagenr()

  " Auto-close tab when this buffer is closed
  autocmd BufWipeout <buffer> call s:close_review_tab()

  " Populate file list
  call setline(1, files)
  setlocal nomodifiable

  " Maps
  nnoremap <buffer> <silent> <CR> :call code_review#open_diff()<CR>
  nnoremap <buffer> <silent> q :tabclose<CR>

  " Resize list window to fit content width, max 40?
  let width = 30
  for f in files
    if len(f) + 2 > width
      let width = len(f) + 2
    endif
  endfor
  if width > 50 | let width = 50 | endif
  execute 'vertical resize ' . width

  " Move to the right window (create it)
  " Currently we are in the only window. Split right.
  " We want the list to be on the left.
  " So split, move right.
  " Use 'vertical new' without 'topleft' or 'botright' usually splits the current window.
  " Since we want the NEW window to be on the RIGHT, we can use 'vertical rightbelow new'
  " or simply 'set splitright' temporarily.
  
  " Better approach:
  " We are in the file list buffer.
  " 'vertical new' splits the window.
  " By default, 'splitright' option determines side.
  " Let's force it.
  vertical rightbelow new

  " Now we have [List] | [Empty]
  " We are in the new window (right).
  wincmd l
  " Setup an initial empty buffer or help text
  setlocal buftype=nofile
  setlocal bufhidden=wipe
  setlocal noswapfile
  call setline(1, ["Select a file from the left to view diff."])
  setlocal nomodifiable
  
  " Go back to list
  wincmd h
endfunction

function! code_review#open_diff() abort
  let fname = getline('.')
  if empty(fname) | return | endif

  let target = b:code_review_target
  let list_win_id = win_getid()

  " Navigate to the right side
  " Strategy: Close all windows except the current list window, then create a new split
  
  " This is destructive to other splits, but consistent with 'Review Mode'
  let win_count = winnr('$')
  
  " We can't just 'only' because we want to keep the current window (list) 
  " and we want it to be on the left.
  
  " Simple approach:
  " 1. Go to right window (if exists)
  wincmd l
  if win_getid() == list_win_id
     " We were alone? Create split
     vertical rightbelow new
  endif

  " Now we are in a window that is NOT the list.
  " This will be our target window.
  let target_win_id = win_getid()

  " Close all other windows that are not the list and not the current target.
  let wins_to_close = []
  for i in range(1, winnr('$'))
    let wid = win_getid(i)
    if wid != list_win_id && wid != target_win_id
      call add(wins_to_close, wid)
    endif
  endfor

  for wid in wins_to_close
    call win_execute(wid, 'close')
  endfor

  " Now we have [List] | [Current]
  " Setup buffer for diff view
  " We reuse the current window (which is the right split)
  enew
  setlocal buftype=nofile
  setlocal bufhidden=wipe
  setlocal noswapfile
  setlocal nonumber norelativenumber
  setlocal filetype=code_review
  
  " Auto-close tab when this buffer is closed
  autocmd BufWipeout <buffer> call s:close_review_tab()
  
  " Construct git command
  let cmd = 'git diff --no-color ' . shellescape(target) . ' -- ' . shellescape(fname)

  let output = systemlist(cmd)
  
  if v:shell_error
    call setline(1, ["Error generating diff:"] + output)
  else
    if empty(output)
      call setline(1, "No differences found (binary file or empty?).")
    else
      " Filter out git diff header (lines before the first hunk starting with @@)
      let hunk_idx = -1
      for i in range(len(output))
        if output[i] =~# '^@@'
          let hunk_idx = i
          break
        endif
      endfor

      if hunk_idx >= 0
        let diff_lines = output[hunk_idx : -1]
        let [display_lines, line_numbers] = s:process_diff_output(diff_lines)
        call setline(1, display_lines)
        if !empty(line_numbers)
          call s:add_line_numbers_virtual(line_numbers)
        endif
      else
        " No hunks found (e.g. binary file, mode change only), show as is
        call setline(1, output)
      endif
    endif
  endif
  
  setlocal nomodifiable
  
  " Return focus to list window so user can quickly pick next file
  call win_gotoid(list_win_id)
endfunction

function! code_review#close() abort
  if exists('t:is_code_review')
    tabclose
  else
    " Check if any other tab is a code review tab
    for i in range(1, tabpagenr('$'))
      let vars = gettabvar(i, '')
      if has_key(vars, 'is_code_review')
        execute 'tabclose ' . i
        return
      endif
    endfor
    echo "No active code review session found."
  endif
endfunction
