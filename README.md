## vim-stacktrace - Stacktrace of Vim script

[![Travis Build Status](https://travis-ci.org/haya14busa/vim-stacktrace.svg?branch=master)](https://travis-ci.org/haya14busa/vim-stacktrace)
[![Coverage Status](https://coveralls.io/repos/github/haya14busa/vim-stacktrace/badge.svg?branch=master)](https://coveralls.io/github/haya14busa/vim-stacktrace?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/haya14busa/vim-stacktrace)](https://goreportcard.com/report/github.com/haya14busa/vim-stacktrace)
[![LICENSE](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/haya14busa/vim-stacktrace/go/stacktrace?status.svg)](https://godoc.org/github.com/haya14busa/vim-stacktrace/go/stacktrace)

vim-stacktrace provides a way to get a callstack or build stacktrace by error information (e.g. `v:throwpoint`, error message).
You can create quickfix list or location list from the result.

vim-stacktrace helps you to debug Vim script :bug: and to report a helpful error report to issue tracker of Vim plugins :two_hearts:

#### :mag: :bug: Demo :CStacktraceFromhist

`:CStacktraceFromhist` sets selected error stacktrace to quickfix list from message history.

![stacktracefromhist.gif (1287×800)](https://raw.githubusercontent.com/haya14busa/i/e7ef65e590e850ea37425c6ebf4479c1422ef8c8/vim-stacktrace/stacktracefromhist.gif)

#### :paw_prints: Demo stacktrace#callstack()

`stacktrace#callstack()` returns current callstack.

![anim.gif (1195×823)](https://raw.githubusercontent.com/haya14busa/i/b1065499c18fb0001198bdb911151cb47fa1759a/vim-stacktrace/anim.gif)


### Requirements
- Vim 8.0 or above
- "go" command in $PATH

### Installation

[dein.vim](https://github.com/Shougo/dein.vim) / [vim-plug](https://github.com/junegunn/vim-plug)

```vim
call dein#add('haya14busa/vim-stacktrace', {'build': 'make'})
```

```
Plug 'haya14busa/vim-stacktrace', { 'do': 'make' }
```

### Proof of Concept: Writing Vim plugin in Go lang for Vim 8.0
vim-stacktrace demonstrates a feasibility to write Vim plugin in Go lang for Vim 8.0.

Libraries which helps me to write vim-stacktrace in Go lang.

- [haya14busa/vim-go-client](https://github.com/haya14busa/vim-go-client) for communicating with Vim
- [haya14busa/go-vimlparser](https://github.com/haya14busa/go-vimlparser) for creating rich stacktrace by parsing Vim script without any noticeable delay

### :bird: Author
haya14busa (https://github.com/haya14busa)
