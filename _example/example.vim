" vim -c 'set runtimepath+=.' -S _example/example.vim

function! s:test() abort
  return s:test2()
endfunction

function! s:test2() abort
  return F()
endfunc

function! F() abort
  let l:G = {-> s:test3()}
  " ...
  return l:G()
endfunction

function! s:test3() abort
  return stacktrace#callstack()
endfunction

function! Main() abort
  call ch_logfile('/tmp/vimchlog.txt', 'w')
  call setqflist(s:test().stacks)
  copen
  cfirst
endfunction

if expand('%:p') ==# expand('<sfile>:p') || expand('%:p') ==# ''
  call Main()
endif
