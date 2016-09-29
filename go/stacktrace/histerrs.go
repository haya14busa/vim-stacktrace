package stacktrace

import (
	"fmt"
	"regexp"
	"strings"
)

// Error represents Vim script error
// vimdoc:type:
//	Error *stacktrace-type-error*
type Error struct {
	// Throwpint similar to v:throwpint. You can build stacktrace from this using
	// Vim.Build()
	// e.g.
	//   function F[5]..<lambda>3[1]..<SNR>13_test3[2]
	//   /path/to/file.vim[14]
	Throwpoint string `json:"throwpoint"`

	// Vim script error message
	// e.g.
	//   E121: Undefined variable: err1
	//   E15: Invalid expression: err1
	Messages []string `json:"messages"`
}

const detectedLinePrefix = "Error detected while processing "

var (
	histerrsLineRegex = regexp.MustCompile(`^line\s+(\d+):$`)
	histerrsErrRegex  = regexp.MustCompile(`^E\d+:`)
)

type histState int

const (
	histDefault histState = iota
	histDetecting
	histLine
	histErrmsg
)

// Histerrs parses given message history and returns all errors. :h :message
// Example(msghist):
//   Error detected while processing function Main[2]..<SNR>96_test[1]..<SNR>96_test2[1]..F:
//   line    3:
//   E121: Undefined variable: err1
//   E15: Invalid expression: err1
//   line    4:
//   E121: Undefined variable: err2
//   E15: Invalid expression: err2
//   Error detected while processing /path/to/file.vim:
//   line   33:
//   E605: Exception not caught: 0
//
// vimdoc:func:
//	stacktrace#histerrs([{string}])	*stacktrace#histerrs()*
//		Parses message history and returns list of error |stacktrace-type-error|.
//		|:message| content is used by default.
func Histerrs(msghist string) []*Error {
	var errors []*Error
	e := &Error{}
	state := histDefault
	basethrowpoint := ""

	reset := func() {
		e = &Error{}
		basethrowpoint = ""
		state = histDefault
	}

	push := func() {
		errors = append(errors, e)
		reset()
	}

	// (reset) for invalid move
	//
	//                      +----<<<----(push)----<<<----+
	//                      |                            |
	//                      |             +-<-(push)-<-+ |
	//                      |             |            | |
	// histDefault -> histDetecting -> histLine -> histErrmsg
	//  | |     |                                   |     | |
	//  | +->>>-+                                   +->>>-+ |
	//  |                                                   |
	//  +----------<<<-------(push)--------------<<<<-------+

	// append empty line to make sure to push the last error.
	lines := append(strings.Split(msghist, "\n"), "")
	for _, line := range lines {
		switch state {
		case histDefault:
			if strings.HasPrefix(line, detectedLinePrefix) {
				state = histDetecting
				basethrowpoint = line[len(detectedLinePrefix) : len(line)-1]
			}
		case histDetecting:
			ms := histerrsLineRegex.FindStringSubmatch(line)
			if len(ms) == 2 {
				state = histLine
				e.Throwpoint = fmt.Sprintf("%s[%s]", basethrowpoint, ms[1])
			} else {
				reset()
			}
		case histLine:
			if histerrsErrRegex.MatchString(line) {
				state = histErrmsg
				e.Messages = append(e.Messages, line)
			} else {
				reset()
			}
		case histErrmsg:
			if histerrsErrRegex.MatchString(line) {
				e.Messages = append(e.Messages, line)
				continue
			}
			ms := histerrsLineRegex.FindStringSubmatch(line)
			if len(ms) == 2 {
				savebasethrowpoint := basethrowpoint
				push()
				basethrowpoint = savebasethrowpoint
				state = histLine // after push()
				e.Throwpoint = fmt.Sprintf("%s[%s]", basethrowpoint, ms[1])
			} else if strings.HasPrefix(line, detectedLinePrefix) {
				push()
				state = histDetecting
				basethrowpoint = line[len(detectedLinePrefix) : len(line)-1]
			} else {
				push()
			}
		}
	}
	return errors
}

// Fromhist returns selected stacktrace from errors in message history.
//
// vimdoc:func:
//	stacktrace#fromhist()	*stacktrace#fromhist()*
//		Show error candidates from |message-history| and returns stacktrace of
//		selected error |stacktrace-type-stacktrace|.
func (cli *Vim) Fromhist() (*Stacktrace, error) {
	msghist, err := cli.callstrfunc("execute", ":message")
	if err != nil {
		return nil, err
	}
	selected, err := cli.selectError(msghist)
	if err != nil || selected == nil {
		return nil, err
	}
	stacktrace, err := cli.Build(selected.Throwpoint)
	if err != nil {
		return nil, err
	}
	// Add error messages
	if len(stacktrace.Stacks) > 0 {
		last := stacktrace.Stacks[len(stacktrace.Stacks)-1]
		last.Text = strings.Join(selected.Messages, ", ") + " : " + last.Text
	}
	return stacktrace, nil
}

// selectError selectes error from msg, it may return nil.
func (cli *Vim) selectError(msghist string) (*Error, error) {
	histerrs := Histerrs(msghist)

	if len(histerrs) == 0 {
		return nil, nil
	} else if len(histerrs) == 1 {
		return histerrs[0], nil
	}

	candidates := make([]string, 0, len(histerrs))
	for i, histerr := range histerrs {
		s := fmt.Sprintf("%d. %v: %v", i+1, histerr.Throwpoint, strings.Join(histerr.Messages, ", "))
		candidates = append(candidates, s)
	}

	j, err := inputlist(cli, candidates)
	if err != nil {
		return nil, err
	}
	if j == 0 { // canceled
		return nil, nil
	} else if j < 0 || len(candidates) < j {
		return nil, fmt.Errorf("selected invalid number: %v", j)
	}
	return histerrs[j-1], nil
}
