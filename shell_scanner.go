package ipmitool

import (
	"bufio"
	"bytes"
	"io"
)

var (
	prompt = []byte("ipmitool> ")
)

type shellScanner struct {
	*bufio.Scanner
	skip   []byte
	prompt bool
}

func (s *shellScanner) AtPrompt() bool {
	return s.prompt
}

func (s *shellScanner) Scan(skip []byte) bool {
	s.skip = skip
	return s.Scanner.Scan()
}

// Split is defined here to prevent this method on the underlying scanner from
// being called.
func (s *shellScanner) Split() {}

func (s *shellScanner) init(r io.Reader) {
	if s.Scanner != nil {
		panic("ipmitool: scanner already initialized")
	}
	s.Scanner = bufio.NewScanner(r)
	s.Scanner.Split(s.split)
}

func (s *shellScanner) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// clog.Tracef("split data=%q atEOF=%v", data, atEOF)
	// defer func() {
	// 	if err != nil {
	// 		clog.Tracef("split data=%q atEOF=%v advance=%d token=%q err=%q", data, atEOF, advance, token, err)
	// 	} else {
	// 		clog.Tracef("split data=%q atEOF=%v advance=%d token=%q", data, atEOF, advance, token)
	// 	}
	// }()

	if len(s.skip) > 0 {
		if len(data) < len(s.skip) {
			if atEOF {
				return len(data), data, nil
			}
			return
		}
		if bytes.HasPrefix(data, s.skip) {
			advance = len(s.skip)
			data = data[advance:]
		} else {
			clog.Warningf("expected %q; got %q", s.skip, data[:len(s.skip)])
		}
		s.skip = nil
	}

	s.prompt = bytes.HasSuffix(data, prompt)

	if s.prompt {
		return advance + len(data), data[:len(data)-len(prompt)], nil
	} else if atEOF {
		return advance + len(data), data, nil
	}
	return
}
