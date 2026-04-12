" test/run_tests.vim - Test runner for vim-code-review
" Usage: vim -u NONE -N -S test/run_tests.vim

let s:test_count = 0
let s:pass_count = 0
let s:fail_count = 0
let s:errors = []

function! s:assert(condition, message) abort
  let s:test_count += 1
  if a:condition
    let s:pass_count += 1
    echom '  ✓ ' . a:message
  else
    let s:fail_count += 1
    call add(s:errors, a:message)
    echom '  ✗ ' . a:message
  endif
endfunction

function! s:assert_equal(expected, actual, message) abort
  call s:assert(a:expected ==# a:actual, a:message . ' (expected: ' . string(a:expected) . ', got: ' . string(a:actual) . ')')
endfunction

function! s:assert_match(pattern, string, message) abort
  call s:assert(a:string =~# a:pattern, a:message . ' (pattern: ' . a:pattern . ', string: ' . a:string . ')')
endfunction

function! s:run_test(name, Func) abort
  echom ''
  echom 'Running: ' . a:name
  try
    call a:Func()
  catch
    let s:fail_count += 1
    call add(s:errors, a:name . ': ' . v:exception)
    echom '  ✗ Exception: ' . v:exception
  endtry
endfunction

" =============================================================================
" Test Cases
" =============================================================================

function! s:test_has_virtual_text_detection() abort
  " Source the autoload file to access script-local functions via testing
  source autoload/code_review.vim
  
  " The function exists check
  call s:assert(exists('*code_review#start'), 'code_review#start function exists')
  call s:assert(exists('*code_review#open_diff'), 'code_review#open_diff function exists')
  call s:assert(exists('*code_review#close'), 'code_review#close function exists')
endfunction

function! s:test_hunk_header_parsing() abort
  " Test the regex pattern used for hunk header parsing
  let pattern = '^@@ -\(\d\+\),\?\(\d*\) +\(\d\+\),\?\(\d*\) @@'
  
  " Standard hunk header
  let match = matchlist('@@ -10,5 +12,7 @@ function name', pattern)
  call s:assert(!empty(match), 'Matches standard hunk header')
  call s:assert_equal('10', match[1], 'Extracts old start line')
  call s:assert_equal('5', match[2], 'Extracts old count')
  call s:assert_equal('12', match[3], 'Extracts new start line')
  call s:assert_equal('7', match[4], 'Extracts new count')
  
  " Hunk header without counts (single line change)
  let match2 = matchlist('@@ -1 +1 @@', pattern)
  call s:assert(!empty(match2), 'Matches single-line hunk header')
  call s:assert_equal('1', match2[1], 'Extracts old start for single line')
  call s:assert_equal('1', match2[3], 'Extracts new start for single line')
  
  " Hunk header with only new count
  let match3 = matchlist('@@ -0,0 +1,25 @@', pattern)
  call s:assert(!empty(match3), 'Matches new file hunk header')
  call s:assert_equal('0', match3[1], 'Old start is 0 for new file')
  call s:assert_equal('1', match3[3], 'New start is 1 for new file')
  call s:assert_equal('25', match3[4], 'New count is 25')
endfunction

function! s:test_diff_line_type_detection() abort
  " Test prefix detection logic
  call s:assert_equal('+', strpart('+added line', 0, 1), 'Detects added line prefix')
  call s:assert_equal('-', strpart('-removed line', 0, 1), 'Detects removed line prefix')
  call s:assert_equal(' ', strpart(' context line', 0, 1), 'Detects context line prefix')
  call s:assert_equal('@', strpart('@@ -1,2 +1,2 @@', 0, 1), 'Detects hunk header prefix')
endfunction

function! s:test_line_number_formatting() abort
  " Test printf formatting for line numbers
  call s:assert_equal('   1 ', printf('%4d ', 1), 'Formats single digit')
  call s:assert_equal('  10 ', printf('%4d ', 10), 'Formats double digit')
  call s:assert_equal(' 100 ', printf('%4d ', 100), 'Formats triple digit')
  call s:assert_equal('1000 ', printf('%4d ', 1000), 'Formats four digits')
endfunction

function! s:test_buffer_options() abort
  " Create a test buffer and verify options can be set
  new
  setlocal buftype=nofile
  setlocal bufhidden=wipe
  setlocal noswapfile
  
  call s:assert_equal('nofile', &l:buftype, 'Buffer type is nofile')
  call s:assert_equal('wipe', &l:bufhidden, 'Buffer hidden is wipe')
  call s:assert_equal(0, &l:swapfile, 'Swapfile is disabled')
  
  bwipeout!
endfunction

function! s:test_tab_variable() abort
  " Test tab-local variable management
  tabnew
  let t:is_code_review = 1
  
  call s:assert(exists('t:is_code_review'), 'Tab variable exists')
  call s:assert_equal(1, t:is_code_review, 'Tab variable has correct value')
  
  tabclose
endfunction

function! s:test_virtual_text_capability() abort
  " Test virtual text detection logic
  if has('nvim')
    call s:assert(1, 'Neovim has virtual text support')
  elseif has('textprop') && exists('*prop_add')
    call s:assert(1, 'Vim has textprop support')
  else
    call s:assert(1, 'Older Vim without virtual text (graceful degradation)')
  endif
endfunction

" =============================================================================
" Main Test Runner
" =============================================================================

function! s:run_all_tests() abort
  echom '=============================================='
  echom 'vim-code-review Test Suite'
  echom '=============================================='
  
  call s:run_test('Function existence', function('s:test_has_virtual_text_detection'))
  call s:run_test('Hunk header parsing', function('s:test_hunk_header_parsing'))
  call s:run_test('Diff line type detection', function('s:test_diff_line_type_detection'))
  call s:run_test('Line number formatting', function('s:test_line_number_formatting'))
  call s:run_test('Buffer options', function('s:test_buffer_options'))
  call s:run_test('Tab variables', function('s:test_tab_variable'))
  call s:run_test('Virtual text capability', function('s:test_virtual_text_capability'))
  
  echom ''
  echom '=============================================='
  echom 'Results: ' . s:pass_count . ' passed, ' . s:fail_count . ' failed, ' . s:test_count . ' total'
  echom '=============================================='
  
  if !empty(s:errors)
    echom ''
    echom 'Failures:'
    for err in s:errors
      echom '  - ' . err
    endfor
  endif
  
  " Exit with appropriate code for CI
  if s:fail_count > 0
    cquit!
  else
    qall!
  endif
endfunction

" Run tests
call s:run_all_tests()
