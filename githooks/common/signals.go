package common

// Cleanup context with registered handlers which can be executed in reverse order
// only once.
// Mainly for signal interrupts.
type InterruptContext struct {
	handlers []func()

	isRun bool
}

func (c *InterruptContext) AddHandler(handler func()) {
	c.handlers = append(c.handlers, handler)
}

func (c *InterruptContext) RunHandlers() {
	if c.isRun {
		return
	}

	for i := range c.handlers {
		c.handlers[len(c.handlers)-i-1]()
	}

	c.isRun = true
}
