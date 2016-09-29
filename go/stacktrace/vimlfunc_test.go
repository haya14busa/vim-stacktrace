package stacktrace

import "testing"

func TestVimSfile(t *testing.T) {
	v := &Vim{c: cli}
	got, err := v.sfile()
	if err == nil {
		t.Errorf("Vim.sfile() returns %v, want error", got)
	}
}

func TestVimfunction(t *testing.T) {
	v := &Vim{c: cli}
	{
		got, err := v.function("hoge")
		if err == nil {
			t.Errorf("Vim.function('hoge') returns %v, want error", got)
		}
	}
	{
		f := `
function! Hoge()
endfunction
`
		v.c.Call("execute", f)

		got, err := v.function("Hoge")
		if err != nil {
			t.Errorf("Vim.function('Hoge') got an error %v", err)
		}
		if got == "" {
			t.Error("Vim.function('Hoge') == ''")
		}
	}
}

func TestVimCallstrfunc(t *testing.T) {
	v := &Vim{c: cli}
	{
		got, err := v.callstrfunc("substitute", "vim-callstack", "callstack", "stacktrace", "g")
		if err != nil {
			t.Error(err)
		}
		if want := "vim-stacktrace"; got != want {
			t.Errorf("substitute(...) = %v, want %v", got, want)
		}
	}

	{
		_, err := v.callstrfunc("pow", 3, 3)
		if err == nil {
			t.Error("Vim.callstrfunc('pow', 3, 3) want error, but got nil")
		}
	}
}

func TestVimCallintfunc(t *testing.T) {
	v := &Vim{c: cli}
	{
		_, err := v.callintfunc("substitute", "vim-callstack", "callstack", "stacktrace", "g")
		if err == nil {
			t.Error("Vim.callintfunc('substitute', ...) want error, but got nil")
		}
	}

	{
		got, err := v.callintfunc("pow", 3, 3)
		if err != nil {
			t.Error(err)
		}
		if want := 27; got != want {
			t.Errorf("Vim.callintfunc('pow', 3, 3) = %v, want %v", got, want)
		}
	}
}
