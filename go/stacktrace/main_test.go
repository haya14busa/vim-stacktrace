package stacktrace

import (
	"log"
	"os"
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

func TestVimHandle(t *testing.T) {
	v := &Vim{c: cli}
	tests := []struct {
		in vim.Body
	}{
		{map[string]interface{}{"id": "stacktrace#build", "throwpoint": "function F[1]"}},
		{map[string]interface{}{"id": "stacktrace#histerrs", "msghist": ""}},
		{map[string]interface{}{"id": "stacktrace#fromhist"}},
	}
	for _, tt := range tests {
		if _, err := v.handle(tt.in); err != nil {
			t.Errorf("Vim.handle(%v) got an unexpected error %v", tt.in, err)
		}
	}
}

func TestVimHandle_error(t *testing.T) {
	v := &Vim{c: cli}
	tests := []struct {
		in vim.Body
	}{
		{0},
		{map[string]interface{}{}},
		{map[string]interface{}{"id": 1}},
		{map[string]interface{}{"id": "unexpected id"}},
		{map[string]interface{}{"id": "stacktrace#callstack"}}, // <sfile> is empty
		{map[string]interface{}{"id": "stacktrace#build"}},
		{map[string]interface{}{"id": "stacktrace#build", "throwpoint": 1}},
		{map[string]interface{}{"id": "stacktrace#histerrs"}},
		{map[string]interface{}{"id": "stacktrace#histerrs", "msghist": 1}},
	}
	for _, tt := range tests {
		got, err := v.handle(tt.in)
		if err == nil {
			t.Errorf("Vim.handle(%v) = %v, but want error", tt.in, got)
		}
		t.Log(err)
	}
}

func TestMyHandler_Serve(t *testing.T) {
	tests := []struct {
		in *vim.Message
	}{
		{&vim.Message{}},                  // no id
		{&vim.Message{MsgID: 1, Body: 1}}, // invalid body
		{&vim.Message{ // ok
			MsgID: 1,
			Body:  map[string]interface{}{"id": "stacktrace#build", "throwpoint": "function F[1]"},
		}},
	}
	handler := &myHandler{}
	for _, tt := range tests {
		// Just checking it works without panic
		handler.Serve(cli, tt.in)
	}
}
