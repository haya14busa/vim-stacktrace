package stacktrace

import (
	"fmt"
	"log"
	"os"
	"os/user"

	vim "github.com/haya14busa/vim-go-client"
)

var homedir string

func init() {
	usr, _ := user.Current()
	homedir = usr.HomeDir
}

type myHandler struct{}

func (h *myHandler) Serve(cli *vim.Client, msg *vim.Message) {
	if msg.MsgID > 0 {
		ret, err := (&Vim{c: cli}).handle(msg.Body)
		if err != nil {
			ret = &Err{Error: err.Error()}
		}
		cli.Send(&vim.Message{
			MsgID: msg.MsgID,
			Body:  ret,
		})
	}
}

func (cli *Vim) handle(msgBody vim.Body) (interface{}, error) {
	body, ok := msgBody.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("message body is invalid: %+v", msgBody)
	}
	id, ok := body["id"]
	if !ok {
		return nil, fmt.Errorf("id field is not in message body: %v", body)
	}
	if s, ok := id.(string); ok {
		switch s {
		case "stacktrace#callstack":
			return cli.Callstack()
		case "stacktrace#build":
			t, ok := body["throwpoint"]
			if !ok {
				return nil, fmt.Errorf("throwpoint is required in message body: %v", body)
			} else if _, ok := t.(string); !ok {
				return nil, fmt.Errorf("throwpoint is not string: %+v", t)
			}
			return cli.Build(t.(string))
		default:
			return nil, fmt.Errorf("got an unexpected id: %v", s)
		}
	}
	return nil, fmt.Errorf("id is not string: %+v", id)
}

// Main func.
func Main() {
	handler := &myHandler{}
	cli := vim.NewClient(vim.NewReadWriter(os.Stdin, os.Stdout), handler)
	log.Fatal(cli.Start())
}
