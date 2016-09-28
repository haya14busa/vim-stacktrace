package callstack

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	vimlparser "github.com/haya14busa/go-vimlparser"
	"github.com/haya14busa/go-vimlparser/ast"
	vim "github.com/haya14busa/vim-go-client"
)

var (
	fileFuncLines   = make(map[string]map[string]int)
	fileFuncLinesMu sync.RWMutex
)

type Err struct {
	Error string `json:"error"`
}

type Callstack struct {
	Entries []*Entry `json:"entries"`
}

type Entry struct {
	// Function name including <SNR> for script local function
	Funcname string `json:"funcname"`

	// The line number relative to the start of the function
	Flnum int `json:"flnum"`

	// Line text. It's empty if the func is lambda or partial
	Line string `json:"line,omitempty"`

	// Filename is empty if func is defined in Ex-command line
	Filename string `json:"filename,omitempty"`
	// The line number relative to the start of the file
	Lnum int `json:"lnum,omitempty"`

	// Text for quickfix or location list
	Text string `json:"text,omitempty"`
}

type Vim struct {
	c *vim.Client
}

func (cli *Vim) debug(msg interface{}) {
	cli.c.Ex("echom " + strconv.Quote(fmt.Sprintf("%+#v", msg)))
}

func (cli *Vim) sfile() (string, error) {
	return cli.callstrfunc("expand", "<sfile>")
}

func (cli *Vim) slnum() (string, error) {
	return cli.callstrfunc("expand", "<slnum>")
}

func (cli *Vim) function(funcname string) (string, error) {
	return cli.callstrfunc("execute", fmt.Sprintf(":verbose function %v", funcname))
}

func (cli *Vim) callstrfunc(f string, args ...interface{}) (string, error) {
	ret, err := cli.c.Call(f, args...)
	if err != nil {
		return "", err
	}
	s, ok := ret.(string)
	if ok {
		return s, nil
	}
	return "", fmt.Errorf("%v(%v) is not string: %v", f, args, ret)
}

func (cli *Vim) Callstack() (*Callstack, error) {
	sfile, err := cli.sfile()
	if err != nil {
		return nil, err
	}
	ss := strings.Split(sfile, "..")
	throwpoint := strings.Join(ss[:len(ss)-1], "..")
	return cli.build(throwpoint)
}

// throwpoint should be normalized
func (cli *Vim) build(throwpoint string) (*Callstack, error) {
	if !strings.HasPrefix(throwpoint, "function ") {
		return nil, fmt.Errorf("doesn't called in function")
	}

	fileFuncLinesMu.Lock()
	fileFuncLines = make(map[string]map[string]int)
	fileFuncLinesMu.Unlock()

	var es []*Entry
	ss := strings.Split(throwpoint[len("function "):], "..")
	for _, e := range ss {
		i := strings.Index(e, "[")
		funcname := e[:i]
		flnum, err := strconv.Atoi(e[i+1 : len(e)-1])
		if err != nil {
			return nil, err
		}
		e, err := cli.buildEntry(funcname, flnum)
		if err != nil {
			return nil, err
		}
		es = append(es, e)
	}
	return &Callstack{Entries: es}, nil
}

func (cli *Vim) Build(throwpoint string) (*Callstack, error) {
	return cli.build(normalizeThrowpoint(throwpoint))
}

// function <SNR>13_test[1]..<SNR>13_test2[1]..F[3]..<lambda>1[1]..<SNR>13_test3, line 2
// Error detected while processing function <SNR>13_test[1]..<SNR>13_test3:
// line    2:
func normalizeThrowpoint(throwpoint string) string {
	i := strings.Index(throwpoint, ", line ")
	if i != -1 {
		lnum := throwpoint[i+len(", line "):]
		return fmt.Sprintf("%s[%s]", throwpoint[:i], lnum)
	}

	if strings.HasPrefix(throwpoint, "Error detected while processing ") {
		throwpoint = throwpoint[len("Error detected while processing "):]
	}

	j := strings.Index(throwpoint, ":\nline")
	if j != -1 {
		lnum := strings.TrimLeft(throwpoint[j+len(":\nline"):len(throwpoint)-1], " ")
		return fmt.Sprintf("%s[%s]", throwpoint[:j], lnum)
	}

	return throwpoint
}

// e: {funcname}[{line}]
func (cli *Vim) buildEntry(funcname string, flnum int) (*Entry, error) {
	e := &Entry{
		Funcname: funcname,
		Flnum:    flnum,
		Text:     fmt.Sprintf("%s:%d:", funcname, flnum),
	}

	f, err := cli.function(funcname)
	if err != nil {
		return e, nil
	}
	lines := strings.Split(strings.Trim(f, "\n"), "\n")

	// Get filename from Last set from ..., empty if func doen't not have Last
	// set from
	file := ""
	if strings.HasPrefix(lines[1], "\tLast set from ") {
		file = lines[1][len("\tLast set from "):]
		if strings.HasPrefix(file, "~/") {
			file = strings.Replace(file, "~", homedir, 1)
		}
		l := make([]string, 0, len(lines)-1)
		l = append(l, lines[0])
		l = append(l, lines[1:]...)
		lines = l
	}
	e.Filename = file

	// Get line text
	targeti := 0
	for i, l := range lines {
		if strings.HasPrefix(l, strconv.Itoa(flnum)) {
			targeti = i
		}
	}
	lastl := lines[len(lines)-2]
	numfield := lastl[:strings.Index(lastl, " ")]
	e.Line = lines[targeti][len(numfield)+2:]
	e.Text += e.Line

	if e.Filename != "" {
		e.Lnum = cli.funcLnum(funcname, file) + flnum
	}

	return e, nil
}

func (cli *Vim) funcLnum(funcname, file string) int {
	if strings.HasPrefix(funcname, "<SNR>") {
		funcname = "s:" + funcname[strings.Index(funcname, "_")+1:]
	}

	fileFuncLinesMu.Lock()
	defer fileFuncLinesMu.Unlock()
	if funclines, ok := fileFuncLines[file]; ok {
		return funclines[funcname]
	}
	f, err := os.Open(file)
	if err != nil {
		return 0
	}
	defer f.Close()

	node, err := vimlparser.ParseFile(f, file, &vimlparser.ParseOption{})
	if err != nil {
		return 0
	}
	fs := funcLines(node)
	fileFuncLines[file] = fs
	return fs[funcname]
}

func funcLines(node ast.Node) map[string]int {
	funcs := make(map[string]int)
	ast.Inspect(node, func(n ast.Node) bool {
		switch f := n.(type) {
		case *ast.Function:
			switch fname := f.Name.(type) {
			case *ast.Ident:
				funcs[fname.Name] = f.Pos().Line
			}
		}
		return true
	})
	return funcs
}
