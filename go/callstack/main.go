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

		body, ok := msg.Body.(map[string]interface{})
		if !ok {
			return
		}
		id, ok := body["id"]
		if !ok {
			return
		}

		v := &Vim{c: cli}

		var ret interface{}

		if s, ok := id.(string); ok {
			switch s {
			case "callstack#get":
				r, err := v.Callstack()
				if err != nil {
					ret = &Err{Error: err.Error()}
				}
				ret = r

			case "callstack#build":
				t, ok := body["throwpoint"]
				if !ok {
					ret = &Err{Error: "throwpoint is required"}
				} else if _, ok := t.(string); !ok {
					ret = &Err{Error: "throwpoint is not string"}
				} else {
					r, err := v.Build(t.(string))
					if err != nil {
						ret = &Err{Error: err.Error()}
					}
					ret = r
				}
			}
		}

		cli.Send(&vim.Message{
			MsgID: msg.MsgID,
			Body:  ret,
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
