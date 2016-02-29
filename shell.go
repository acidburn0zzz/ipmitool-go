package ipmitool

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/kr/pty"
)

type Shell struct {
	conf Config

	cmd *exec.Cmd
	pty *os.File

	once sync.Once
	cmds chan *command

	scanner shellScanner
	completion
}

func (c *Config) NewShell() (*Shell, error) {
	cmd := exec.Command("ipmitool", c.args("shell")...)

	pty, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	s := &Shell{
		conf: *c,
		cmd:  cmd,
		pty:  pty,
	}
	s.init()

	if b := s.readUntilPrompt(nil); len(b) > 0 {
		clog.Error(string(b))
	}
	go s.loop()

	return s, nil
}

func (s *Shell) Close() error {
	s.close(true)
	return s.Wait()
}

func (s *Shell) Done() <-chan struct{} {
	return s.done
}

func (s *Shell) Err() error {
	return s.err
}

func (s *Shell) Exec(str string, out io.Writer) (err error) {
	if i := strings.IndexByte(str, '\n'); i != -1 {
		str = strings.TrimSpace(str[:i])
	}
	if len(str) == 0 {
		return errors.New("ipmitool: empty command given")
	}
	if s.isDone() {
		return s.errOr(ErrClosed)
	}

	cmd := newCommand()
	cmd.str = str
	cmd.out = out
	cmd.Add(1)

	select {
	case s.cmds <- cmd:
		cmd.Wait()
	case <-s.done:
		cmd.Done() // decr the counter so we can use it again
		err = s.errOr(ErrClosed)
	}
	recycleCommand(cmd)
	return
}

func (s *Shell) Wait() error {
	select {
	case <-s.done:
		return s.err
	}
}

func (s *Shell) close(evict bool) {
	f := func() {
		if evict {
			s.evict()
		}
		close(s.cmds)
	}
	s.once.Do(f)
}

func (s *Shell) exec(cmd string) []byte {
	clog.Tracef("event=send cmd=%q", cmd)
	start := time.Now()

	if _, err := fmt.Fprintln(s.pty, cmd); err != nil {
		clog.Errorf("event=send_fail cmd=%q err=%q", cmd, err)
		return nil
	}

	out := s.readUntilPrompt([]byte(cmd + "\r\n"))
	clog.Tracef("event=sent cmd=%q duration=%s", cmd, time.Since(start))
	return out
}

func (s *Shell) init() {
	if s.cmds != nil {
		panic("ipmitool: shell already initialized")
	}
	s.cmds = make(chan *command)
	s.scanner.init(s.pty)
	s.completion.init()
}

func (s *Shell) loop() {
	for cmd := range s.cmds {
		cmd.exec(s)
	}

	if out := bytes.TrimSpace(s.exec("exit")); len(out) > 0 {
		clog.Error(string(out))
	}
}

func (s *Shell) readUntilPrompt(skip []byte) []byte {
	if !s.scanner.Scan(skip) {
		if err := s.scanner.Err(); err != nil {
			clog.Error(err)
		}
		s.setErr(s.cmd.Wait())
		return nil
	}
	if !s.scanner.AtPrompt() {
		s.setErr(s.cmd.Wait())
	}
	return s.scanner.Bytes()
}
