// Package concurrency provides utility functions for working with goroutines.
package concurrency

import (
	"context"
	"errors"
)

// ConcurrentPool implements a simple semaphore-like structure to limit
// the number of concurrent goroutines working together.
type ConcurrentPool struct {
	sem chan struct{}
}

// NewConcurrentPool creates a new ConcurrentPool with the specified capacity.
// It returns an error if the capacity is 0.
func NewConcurrentPool(capacity uint) (*ConcurrentPool, error) {
	if capacity == 0 {
		return nil, errors.New("capacity value of a concurrent pool cannot be 0")
	}
	return &ConcurrentPool{sem: make(chan struct{}, capacity)}, nil
}

// Lock acquires a slot from the pool, waiting if necessary until one becomes available.
func (p *ConcurrentPool) Lock() {
	p.sem <- struct{}{}
}

// LockContext acquires a slot or returns when ctx is canceled.
// It returns nil on acquisition or ctx.Err() if cancellation interrupts the wait.
func (p *ConcurrentPool) LockContext(ctx context.Context) error {
	select {
	case p.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Unlock releases a slot, making it available for other goroutines.
func (p *ConcurrentPool) Unlock() {
	select {
	case <-p.sem:
	default:
		panic(errors.New("unlock called on a totally free pool"))
	}
}

func (p *ConcurrentPool) get() uint {
	return uint(len(p.sem))
}
