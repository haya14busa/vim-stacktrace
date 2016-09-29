package main

import (
	"testing"

	vim "github.com/haya14busa/vim-go-client"
)

type testHandler struct{}

func (h *testHandler) Serve(cli *vim.Client, msg *vim.Message) {}

func TestExample(t *testing.T) {
	vimArgs := []string{"-Nu", "NONE", "-i", "NONE", "-n", "-c", "set runtimepath+=."}
	cli, closer, err := vim.NewChildClient(&testHandler{}, vimArgs)
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()
	// use execute() instead of cli.Ex to wait execution
	cli.Call("execute", ":source ./_example/example.vim")
	cli.Expr("Main()")
	got, err := cli.Expr("getqflist()")
	if err != nil {
		t.Fatal(err)
	}
	qflist, ok := got.([]interface{})
	if !ok {
		t.Fatal("getqflist should return list")
	}
	if len(qflist) == 0 {
		t.Fatal("qflist is not set")
	}
	first := qflist[0].(map[string]interface{})
	wantLnum, err := cli.Call("line", ".")
	if err != nil {
		t.Fatal(err)
	}
	gotLnum := first["lnum"].(float64)
	if gotLnum != wantLnum {
		t.Errorf("got lnum:%v, want lnum:%v", gotLnum, wantLnum)
	}
}
