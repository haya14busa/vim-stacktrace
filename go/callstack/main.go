package callstack

import (
	"log"
	"os"
	"os/user"

	vim "github.com/haya14busa/vim-go-client"
)

var homedir string

type myHandler struct{}

func (h *myHandler) Serve(cli *vim.Client, msg *vim.Message) {
	if msg.MsgID > 0 {
		r, err := (&Vim{c: cli}).callstack()
		var body interface{} = r
		if err != nil {
			body = &Err{Error: err.Error()}
		}
		cli.Send(&vim.Message{
			MsgID: msg.MsgID,
			Body:  body,
		})
	}
}

func Main() {
	usr, _ := user.Current()
	homedir = usr.HomeDir

	handler := &myHandler{}
	cli := vim.NewClient(vim.NewReadWriter(os.Stdin, os.Stdout), handler)
	log.Fatal(cli.Start())
}
