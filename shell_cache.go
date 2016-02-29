package ipmitool

import (
	"container/list"
	"sync"
)

const (
	MaxShells = 128
)

var shells struct {
	sync.Mutex
	l list.List
	m map[Config]*list.Element
}

func maybeEvict() {
	for len(shells.m) > MaxShells {
		sh, ok := shells.l.Remove(shells.l.Back()).(*Shell)
		if !ok {
			return
		}
		go sh.close(false)
	}
}

func (c *Config) GetShell() (*Shell, error) {
	shells.Lock()
	defer shells.Unlock()

	if shells.m == nil {
		shells.m = make(map[Config]*list.Element)
		shells.l.Init()
	} else if elem := shells.m[*c]; elem != nil {
		shells.l.MoveToFront(elem)
		return elem.Value.(*Shell), nil
	}

	s, err := c.NewShell()
	if err != nil {
		return nil, err
	}
	shells.m[*c] = shells.l.PushFront(s)
	maybeEvict()
	return s, nil
}

func (s *Shell) evict() {
	shells.Lock()
	defer shells.Unlock()

	if shells.m == nil {
		return
	}
	if elem := shells.m[s.conf]; elem != nil && elem.Value == s {
		shells.l.Remove(elem)
		delete(shells.m, s.conf)
	}
}
