package ipmitool

import (
	"errors"
)

var (
	ErrClosed = errors.New("ipmitool: session closed")
)
