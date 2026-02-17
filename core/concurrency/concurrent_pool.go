// Package concurrency provides utility functions for working with goroutines.
package concurrency

import (
	"errors"
	"sync"

	"github.com/fmotalleb/go-tools/concurrency"
)

// ConcurrentPool implements a simple semaphore-like structure to limit
// the number of concurrent goroutines working together.
type ConcurrentPool struct {
	lockerLock sync.Locker
	available  uint                           // total capacity of the pool
	used       *concurrency.LockedValue[uint] // number of slots currently in use
	changeChan chan struct{}                  // channel for signaling changes in the pool's state
}

// NewConcurrentPool creates a new ConcurrentPool with the specified capacity.
// It panics if the capacity is 0.
func NewConcurrentPool(capacity uint) (*ConcurrentPool, error) {
	if capacity == 0 {
		return nil, errors.New("capacity value of a concurrent poll cannot be 0")
	}
	return &ConcurrentPool{
		lockerLock: new(sync.Mutex),
		available:  capacity,
		used:       concurrency.NewLockedValue[uint](0),
		changeChan: make(chan struct{}, 1),
	}, nil
}

// Lock acquires a lock from the pool, waiting if necessary until a slot becomes available.
// It increments the used count using the reserveSlot method.
func (p *ConcurrentPool) Lock() {
	p.lockerLock.Lock()
	defer p.lockerLock.Unlock()
	if p.available > p.get() {
		p.increase()
		return
	}
	for range p.changeChan {
		if p.available > p.get() {
			p.increase()
			break
		}
	}
}

// Unlock releases a lock, making a slot available for other goroutines.
// It decrements the used count and sends a signal on the changeChan to notify waiting goroutines.
func (p *ConcurrentPool) Unlock() {
	if p.get() == 0 {
		panic(errors.New("unlock called on a totally free pool"))
	}
	p.decrease()
	select {
	case p.changeChan <- struct{}{}:
	default:
	}
}

func (p *ConcurrentPool) get() uint {
	return p.used.Get()
}

func (p *ConcurrentPool) increase() {
	p.used.Operate(
		func(current uint) uint {
			return (current + 1)
		},
	)
}

func (p *ConcurrentPool) decrease() {
	p.used.Operate(
		func(current uint) uint {
			return (current - 1)
		},
	)
}
