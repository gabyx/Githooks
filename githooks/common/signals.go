package common

// Cleanup context with registered handlers which can be executed in reverse order.
// Mainly for signal intertupts.
type InterruptContext struct {
	handlers []func()
}

func (c *InterruptContext) AddHandler(handler func()) {
	c.handlers = append(c.handlers, handler)
}

func (c *InterruptContext) RunHandlers() {
	for i := range c.handlers {
		c.handlers[len(c.handlers)-i-1]()
	}
}
