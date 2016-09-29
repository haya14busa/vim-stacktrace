" vim -c 'set runtimepath+=.' -S plugin/stacktrace.vim -S _example/error.vim -cmd ':source _example/error.vim'
" :call Main()
" :CStacktraceFromhist

function! s:test() abort
  return s:test2()
endfunction

function! s:test2() abort
  return F()
endfunc

function! F()
  throw err1
  throw err2
endfunction

function! Main()
  call ch_logfile('/tmp/vimchlog.txt', 'w')
  call s:test()
  throw err3
endfunction
