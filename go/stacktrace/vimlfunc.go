package stacktrace

import "fmt"

// func (cli *Vim) debug(msg interface{}) {
// 	cli.c.Ex("echom " + strconv.Quote(fmt.Sprintf("%+#v", msg)))
// }

var inputlist = func(cli *Vim, candidates []string) (int, error) {
	return cli.callintfunc("inputlist", candidates)
}

func (cli *Vim) sfile() (string, error) {
	return cli.callstrfunc("expand", "<sfile>")
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

func (cli *Vim) callintfunc(f string, args ...interface{}) (int, error) {
	ret, err := cli.c.Call(f, args...)
	if err != nil {
		return 0, err
	}
	s, ok := ret.(float64)
	if ok {
		return int(s), nil
	}
	return 0, fmt.Errorf("%v(%v) is not float64: %v", f, args, ret)
}
