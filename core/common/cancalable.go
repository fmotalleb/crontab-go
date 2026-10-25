package common

import "sync"

type Cancelable struct {
	mu     sync.Mutex
	cancel func()
	active map[uint64]func()
	nextID uint64
}

func (c *Cancelable) SetCancel(cancel func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancel = cancel
}

func (c *Cancelable) Cancel() {
	c.mu.Lock()
	cancel := c.cancel
	active := c.active
	c.active = nil
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	for _, cancel := range active {
		cancel()
	}
}

// RegisterCancel adds one active execution cancellation and returns its unregister function.
func (c *Cancelable) RegisterCancel(cancel func()) func() {
	c.mu.Lock()
	if c.active == nil {
		c.active = make(map[uint64]func())
	}
	c.nextID++
	id := c.nextID
	c.active[id] = cancel
	c.mu.Unlock()
	return func() {
		c.mu.Lock()
		delete(c.active, id)
		c.mu.Unlock()
	}
}
