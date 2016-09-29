package stacktrace

import (
	"fmt"
	"regexp"
	"strings"
)

// Error represents Vim script error
type Error struct {
	// Throwpint similar to v:throwpint. You can build stacktrace from this using
	// Vim.Build()
	// e.g.
	//   function F[5]..<lambda>3[1]..<SNR>13_test3[2]
	//   /path/to/file.vim[14]
	Throwpoint string

	// Vim script error message
	// e.g.
	//   E121: Undefined variable: err1
	//   E15: Invalid expression: err1
	Messages []string
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
func Histerrs(msghist string) ([]*Error, error) {
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
	return errors, nil
}
