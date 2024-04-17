package common

import "sync"

// Cleanup context with registered handlers which can be executed in reverse order
// only once.
// Mainly for signal interrupts.
type InterruptContext struct {
	handlers []func()
	isRun    bool
	lock     sync.Mutex
}

func (c *InterruptContext) AddHandler(handler func()) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.handlers = append(c.handlers, handler)
}

func (c *InterruptContext) RunHandlers() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.isRun {
		return
	}

	for i := range c.handlers {
		c.handlers[len(c.handlers)-i-1]()
	}

	c.isRun = true
}
