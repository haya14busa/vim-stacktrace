package stacktrace

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	vim "github.com/haya14busa/vim-go-client"
)

var cli *vim.Client

var vimArgs = []string{"-Nu", "NONE", "-i", "NONE", "-n"}

type testHandler struct{}

func (h *testHandler) Serve(cli *vim.Client, msg *vim.Message) {}

func TestMain(m *testing.M) {
	c, closer, err := vim.NewChildClient(&testHandler{}, vimArgs)
	if err != nil {
		log.Fatal(err)
	}
	cli = c
	code := m.Run()
	closer.Close()
	os.Exit(code)
}

func TestVim_Build(t *testing.T) {
	v := &Vim{c: cli}
	tests := []struct {
		in   string
		want *Stacktrace
	}{
		{
			in: "function <SNR>13_test3, line 2",
			want: &Stacktrace{
				Stacks: []*Stack{
					{
						Funcname: "<SNR>13_test3",
						Flnum:    2,
						Text:     "<SNR>13_test3:2:",
					},
				},
			},
		},
		{
			in: "function F[5]..<lambda>3[1]..<SNR>13_test3, line 2",
			want: &Stacktrace{
				Stacks: []*Stack{
					{
						Funcname: "F",
						Flnum:    5,
						Text:     "F:5:",
					},
					{
						Funcname: "<lambda>3",
						Flnum:    1,
						Text:     "<lambda>3:1:",
					},
					{
						Funcname: "<SNR>13_test3",
						Flnum:    2,
						Text:     "<SNR>13_test3:2:",
					},
				},
			},
		},
		{
			in: "function 14[14]",
			want: &Stacktrace{
				Stacks: []*Stack{
					{
						Funcname: "{14}",
						Flnum:    14,
						Text:     "{14}:14:",
					},
				},
			},
		},
		{
			// support malformed style which doesn't have line number
			in: "function <SNR>13_test3",
			want: &Stacktrace{
				Stacks: []*Stack{
					{
						Funcname: "<SNR>13_test3",
						Flnum:    0,
						Text:     "<SNR>13_test3:0:",
					},
				},
			},
		},
		{ // file
			in: "/path/to/file.vim, line 14",
			want: &Stacktrace{
				Stacks: []*Stack{
					{
						Filename: "/path/to/file.vim",
						Lnum:     14,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		got, err := v.Build(tt.in)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			for i, e := range got.Stacks {
				t.Errorf("got :%#v", e)
				t.Errorf("want:%#v", tt.want.Stacks[i])
			}
		}
	}
}

func ExampleVim_Build() {
	scripts := `
function! F() abort
  let l:G = {-> s:test()}
  return l:G()
endfunction

function! s:test() abort
  return s:d.f()
endfunction

let s:d = {}
function! s:d.f() abort
  return s:test2()
endfunction

function! s:test2() abort
  try
    throw 'error!'
  catch
    return v:throwpoint
    " You can use stacktrace#build in Vim script
    " call stacktrace#build(v:throwpoint)
  endtry
  " If you just want the current callstack, use stacktrace#callstack()
  " call stacktrace#callstack()
endfunction
`
	tmp, _ := ioutil.TempFile("", "vim-stacktrace-test")
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	tmp.WriteString(scripts)
	filename := tmp.Name()

	cli, closer, err := vim.NewChildClient(&testHandler{}, vimArgs)
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()
	v := &Vim{c: cli}
	v.c.Ex(":source " + filename)
	throwpoint, _ := v.c.Expr("g:F()")
	stacktrace, _ := v.Build(throwpoint.(string))
	for _, stack := range stacktrace.Stacks {
		if stack.Filename != "" {
			stack.Filename = "/path/to/file.vim"
		}
		fmt.Println(stack)
	}
	// Output:
	// /path/to/file.vim:4: F:2:  return l:G()
	// :0: <lambda>1:1:
	// /path/to/file.vim:8: <SNR>2_test:1:  return s:d.f()
	// /path/to/file.vim:0: {1}:1:  return s:test2()
	// /path/to/file.vim:18: <SNR>2_test2:2:    throw 'error!'
}

func TestVim_Build_intergration(t *testing.T) {
	v := &Vim{c: cli}
	scripts := `
function! F() abort
  let l:G = {-> s:test()}
  return l:G()
endfunction

function! s:test() abort
  return s:d.f()
endfunction

let s:d = {}
function! s:d.f() abort
  return s:test2()
endfunction

function! s:test2() abort
  return printf('%s[%s]', expand('<sfile>'), expand('<slnum>'))
endfunction
`
	tmp, err := ioutil.TempFile("", "vim-stacktrace-test")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	tmp.WriteString(scripts)
	filename := tmp.Name()

	want := &Stacktrace{
		Stacks: []*Stack{
			{
				Funcname: "F",
				Flnum:    2,
				Line:     "  return l:G()",
				Filename: filename,
				Lnum:     4,
				Text:     "F:2:  return l:G()",
			},
			{
				Funcname: "<lambda>1",
				Flnum:    1,
				Line:     "",
				Filename: "",
				Lnum:     0,
				Text:     "<lambda>1:1:",
			},
			{
				Funcname: "<SNR>2_test",
				Flnum:    1,
				Line:     "  return s:d.f()",
				Filename: filename,
				Lnum:     8,
				Text:     "<SNR>2_test:1:  return s:d.f()",
			},
			{
				Funcname: "{1}",
				Flnum:    1,
				Line:     "  return s:test2()",
				Filename: filename,
				Text:     "{1}:1:  return s:test2()",
			},
			{
				Funcname: "<SNR>2_test2",
				Flnum:    1,
				Line:     `  return printf('%s[%s]', expand('<sfile>'), expand('<slnum>'))`,
				Filename: filename,
				Lnum:     17,
				Text:     `<SNR>2_test2:1:  return printf('%s[%s]', expand('<sfile>'), expand('<slnum>'))`,
			},
		},
	}

	v.c.Ex(":source " + tmp.Name())
	throwpoint, err := v.c.Expr("g:F()")
	if err != nil {
		t.Fatal(err)
	}

	got, err := v.Build(throwpoint.(string))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		for i, e := range got.Stacks {
			if !reflect.DeepEqual(e, want.Stacks[i]) {
				t.Errorf("got :%#v", e)
				t.Errorf("want:%#v", want.Stacks[i])
			}
		}
	}
}

func TestVim_buildFileStack(t *testing.T) {
	v := &Vim{c: cli}
	scripts := `line1
line2
line3
   line4
`
	tmp, err := ioutil.TempFile("", "vim-stacktrace-test")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	tmp.WriteString(scripts)
	filename := tmp.Name()

	tests := []struct {
		lnum int
		want *Stack
	}{
		{lnum: 1, want: &Stack{Lnum: 1, Line: "line1", Text: "line1", Filename: filename}},
		{lnum: 4, want: &Stack{Lnum: 4, Line: "   line4", Text: "   line4", Filename: filename}},
	}
	for _, tt := range tests {
		got, err := v.buildFileStack(filename, tt.lnum)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Vim.buildFileStack(%v) = %#+v, want %#+v", tt.lnum, got, tt.want)
		}
	}

}

func TestSeparateStack(t *testing.T) {
	tests := []struct {
		in       string
		wantBody string
		wantLnum int
	}{
		{
			in:       "<SNR>13_test[1]",
			wantBody: "<SNR>13_test",
			wantLnum: 1,
		},
		{
			in:       "<lambda>3[1]",
			wantBody: "<lambda>3",
			wantLnum: 1,
		},
		{
			in:       "F[100]",
			wantBody: "F",
			wantLnum: 100,
		},
		{
			in:       "/path/to/file[14]",
			wantBody: "/path/to/file",
			wantLnum: 14,
		},
		{
			in:       "[14].vim[24]",
			wantBody: "[14].vim",
			wantLnum: 24,
		},
		{
			in:       "<SNR>14_nolnum",
			wantBody: "<SNR>14_nolnum",
			wantLnum: 0,
		},
	}
	for _, tt := range tests {
		if gotBody, gotLnum := separateStack(tt.in); gotBody != tt.wantBody || gotLnum != tt.wantLnum {
			t.Errorf("separateStack(%v) = (%v, %v), want (%v, %v)", tt.in, gotBody, gotLnum, tt.wantBody, tt.wantLnum)
		}
	}
}

func TestNormalizeThrowpoint(t *testing.T) {

	tests := []struct {
		in   string
		want string
	}{
		{ // v:throwpoint
			in:   "function <SNR>13_test[1]..<SNR>13_test3, line 2",
			want: "function <SNR>13_test[1]..<SNR>13_test3[2]",
		},
		{ // :throw message
			in:   "Error detected while processing function <SNR>13_test[1]..<SNR>13_test3:\nline    2:",
			want: "function <SNR>13_test[1]..<SNR>13_test3[2]",
		},
		{ // v:throwpoint
			in:   "/path/to/file.vim, line 23",
			want: "/path/to/file.vim[23]",
		},
	}

	for _, tt := range tests {
		got := normalizeThrowpoint(tt.in)
		if got != tt.want {
			t.Errorf("normalizeThrowpoint(%v) = %v, tt.want %v", tt.in, got, tt.want)
		}
		if got2 := normalizeThrowpoint(got); got2 != tt.want {
			t.Errorf("normalizeThrowpoint(%v) = %v, tt.want %v", got, got2, tt.want)
		}
	}
}
