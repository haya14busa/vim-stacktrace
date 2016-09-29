"=============================================================================
" FILE: autoload/stacktrace.vim
" AUTHOR: haya14busa
" License: MIT license
"=============================================================================
scriptencoding utf-8
let s:save_cpo = &cpo
set cpo&vim

let g:stacktrace#debug = v:false

function! stacktrace#callstack() abort
  return ch_evalexpr(s:job_start(), {'id': 'stacktrace#callstack'})
endfunction

function! stacktrace#build(throwpoint) abort
  return ch_evalexpr(s:job_start(), {'id': 'stacktrace#build', 'throwpoint': a:throwpoint})
endfunction

function! stacktrace#histerrs(...) abort
  let msghist = get(a:, 1, '')
  if msghist ==# ''
    let msghist = execute(':message')
  endif
  return ch_evalexpr(s:job_start(), {'id': 'stacktrace#histerrs', 'msghist': msghist})
endfunction

function! s:err_cb(ch, msg) abort
  echom 'vim-stacktrace:' . a:msg
endfunction

function! s:separator() abort
  return fnamemodify('.', ':p')[-1 :]
endfunction

let s:is_windows = has('win16') || has('win32') || has('win64') || has('win95')

let s:base = expand('<sfile>:p:h:h')
let s:basecmd = s:base . s:separator() . fnamemodify(s:base, ':t')
let s:cmd = s:basecmd . (s:is_windows ? '.exe' : '')

if g:stacktrace#debug
  let s:cmd = ['go', 'run', s:basecmd . '.go']
elseif !filereadable(s:cmd)
  call system(printf('cd %s && go get -d && go build', s:base))
endif

let s:option = {
\   'in_mode': 'json',
\   'out_mode': 'json',
\   'err_cb': function('s:err_cb'),
\ }

function! s:job_start() abort
  if exists('s:job')
    if ch_status(s:job) ==# 'closed'
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
