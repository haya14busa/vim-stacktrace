"=============================================================================
" FILE: autoload/callstack.vim
" AUTHOR: haya14busa
" License: MIT license
"=============================================================================
scriptencoding utf-8
let s:save_cpo = &cpo
set cpo&vim

let g:callstack#debug = v:false

function! callstack#get() abort
  return ch_evalexpr(s:job_start(), 0)
endfunction

function! s:err_cb(ch, msg) abort
  echom 'vim-callstack:' . a:msg
endfunction

function! s:separator() abort
  return fnamemodify('.', ':p')[-1 :]
endfunction

let s:base = expand('<sfile>:p:h:h')
let s:cmd = s:base . s:separator() . fnamemodify(s:base, ':t')

if g:callstack#debug
  let s:cmd = ['go', 'run', s:cmd . '.go']
elseif !filereadable(s:cmd)
  call system(printf("cd %s && go get -d && go build", s:base))
endif

let s:option = {
\   'in_mode': 'json',
\   'out_mode': 'json',
\   'err_cb': function('s:err_cb'),
\ }

function! s:job_start() abort
  if exists('s:job')
    if ch_status(s:job) ==# "closed"
      call job_stop(s:job)
      let s:job = job_start(s:cmd, s:option)
    endif
  else
    let s:job = job_start(s:cmd, s:option)
  endif
  return s:job
endfunction

let &cpo = s:save_cpo
unlet s:save_cpo
" __END__
" vim: expandtab softtabstop=2 shiftwidth=2 foldmethod=marker
