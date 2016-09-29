package stacktrace

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	vim "github.com/haya14busa/vim-go-client"
)

var cli *vim.Client

var defaultServeFunc = func(cli *vim.Client, msg *vim.Message) {}

var vimArgs = []string{"-Nu", "NONE", "-i", "NONE", "-n"}

var waitLog = func() { time.Sleep(1 * time.Millisecond) }

type testHandler struct {
	f func(cli *vim.Client, msg *vim.Message)
}

func (h *testHandler) Serve(cli *vim.Client, msg *vim.Message) {
	fn := h.f
	if fn == nil {
		fn = defaultServeFunc
	}
	fn(cli, msg)
}

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
				Entries: []*Entry{
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
				Entries: []*Entry{
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
				Entries: []*Entry{
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
				Entries: []*Entry{
					{
						Funcname: "<SNR>13_test3",
						Flnum:    0,
						Text:     "<SNR>13_test3:0:",
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
			for i, e := range got.Entries {
				t.Errorf("got :%#v", e)
				t.Errorf("want:%#v", tt.want.Entries[i])
			}
		}
	}
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
		Entries: []*Entry{
			&Entry{
				Funcname: "F",
				Flnum:    2,
				Line:     "  return l:G()",
				Filename: filename,
				Lnum:     4,
				Text:     "F:2:  return l:G()",
			},
			&Entry{
				Funcname: "<lambda>1",
				Flnum:    1,
				Line:     "",
				Filename: "",
				Lnum:     0,
				Text:     "<lambda>1:1:",
			},
			&Entry{
				Funcname: "<SNR>2_test",
				Flnum:    1,
				Line:     "  return s:d.f()",
				Filename: filename,
				Lnum:     8,
				Text:     "<SNR>2_test:1:  return s:d.f()",
			},
			&Entry{
				Funcname: "{1}",
				Flnum:    1,
				Line:     "  return s:test2()",
				Filename: filename,
				Text:     "{1}:1:  return s:test2()",
			},
			&Entry{
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
		for i, e := range got.Entries {
			if !reflect.DeepEqual(e, want.Entries[i]) {
				t.Errorf("got :%#v", e)
				t.Errorf("want:%#v", want.Entries[i])
			}
		}
	}
}

func TestNormalizeThrowpoint(t *testing.T) {
	want := "function <SNR>13_test[1]..<SNR>13_test3[2]"

	{ // v:throwpoint
		in := "function <SNR>13_test[1]..<SNR>13_test3, line 2"
		got := normalizeThrowpoint(in)
		if got != want {
			t.Errorf("normalizeThrowpoint(%v) = %v, want %v", in, got, want)
		}
		if got2 := normalizeThrowpoint(got); got2 != want {
			t.Errorf("normalizeThrowpoint(%v) = %v, want %v", got, got2, want)
		}
	}

	{ // :throw message
		in := "Error detected while processing function <SNR>13_test[1]..<SNR>13_test3:\nline    2:"
		got := normalizeThrowpoint(in)
		if got != want {
			t.Errorf("normalizeThrowpoint(%v) = %v, want %v", in, got, want)
		}
		if got2 := normalizeThrowpoint(got); got2 != want {
			t.Errorf("normalizeThrowpoint(%v) = %v, want %v", got, got2, want)
		}
	}

}
