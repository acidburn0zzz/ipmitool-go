package ipmitool

type completion struct {
	done chan struct{}
	err  error
}

func (c *completion) Done() <-chan struct{} {
	return c.done
}

func (c *completion) Err() error {
	return c.err
}

func (c *completion) errOr(err error) error {
	if c.err != nil {
		return c.err
	}
	return err
}

func (c *completion) init() {
	if c.done != nil {
		panic("ipmitool: completion already initialized")
	}
	c.done = make(chan struct{})
}

func (c *completion) isDone() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

func (c *completion) setErr(err error) {
	c.err = err
	close(c.done)
}
