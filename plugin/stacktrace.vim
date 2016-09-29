"=============================================================================
" FILE: plugin/stacktrace.vim
" AUTHOR: haya14busa
" License: MIT license
"=============================================================================
scriptencoding utf-8
if expand('%:p') ==# expand('<sfile>:p')
  unlet! g:loaded_stacktrace
endif
if exists('g:loaded_stacktrace')
  finish
endif
let g:loaded_stacktrace = 1
let s:save_cpo = &cpo
set cpo&vim

command! CStacktraceFromhist call s:fromhist('c')
command! LStacktraceFromhist call s:fromhist('l')

function! s:fromhist(type) abort
  let stacktrace = stacktrace#fromhist()
  if stacktrace isnot# v:null
    let locs = stacktrace.stacks
    if a:type is# 'c'
      call setqflist(locs)
    elseif a:type is# 'l'
      call setloclist(0, locs)
    endif
  endif
endfunction

let &cpo = s:save_cpo
unlet s:save_cpo
" __END__
" vim: expandtab softtabstop=2 shiftwidth=2 foldmethod=marker
