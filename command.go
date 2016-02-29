package ipmitool

import (
	"io"
	"sync"
)

type command struct {
	str string
	out io.Writer
	sync.WaitGroup
}

func (c *command) exec(s *Shell) {
	defer c.Done()

	b := s.exec(c.str)

	if c.out != nil {
		c.out.Write(b)
	}
}

var commandPool = sync.Pool{
	New: func() interface{} { return new(command) },
}

func newCommand() *command {
	return commandPool.Get().(*command)
}

func recycleCommand(cmd *command) {
	cmd.str = ""
	cmd.out = nil
	commandPool.Put(cmd)
}
