package concurrency

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestConcurrentPool_PanicCase(t *testing.T) {
	_, err := NewConcurrentPool(0)
	assert.Error(t, err)
}

func TestConcurrentPool_LockUnlock(t *testing.T) {
	t.Run("Lock and Unlock with capacity 1", func(t *testing.T) {
		pool, err := NewConcurrentPool(1)
		assert.NoError(t, err)
		pool.Lock()
		assert.Equal(t, 1, pool.get())
		pool.Unlock()
		assert.Equal(t, 0, pool.get())
	})

	t.Run("Lock and Unlock with capacity 2", func(t *testing.T) {
		pool, err := NewConcurrentPool(2)
		assert.NoError(t, err)
		pool.Lock()
		assert.Equal(t, 1, pool.get())
		pool.Lock()
		assert.Equal(t, 2, pool.get())
		pool.Unlock()
		assert.Equal(t, 1, pool.get())
		pool.Unlock()
		assert.Equal(t, 0, pool.get())
	})

	t.Run("Unlock on a totally free pool", func(t *testing.T) {
		pool, err := NewConcurrentPool(1)
		assert.NoError(t, err)
		assert.Panics(t, pool.Unlock)
	})
}

func TestConcurrentPool_LockUnlockGoroutine(t *testing.T) {
	t.Run("Lock and Unlock with capacity 1 (inside 2 goroutine)", func(t *testing.T) {
		pool, err := NewConcurrentPool(1)

		assert.NoError(t, err)
		chn := make(chan int64)
		for i := 0; i < 2; i++ {
			go func() {
				pool.Lock()
				defer pool.Unlock()
				time.Sleep(time.Millisecond * 50)
				chn <- time.Now().UnixMilli()
			}()
		}
		end1 := <-chn
		end2 := <-chn
		diff := end2 - end1
		assert.NotZero(t, diff)
		assert.True(t, diff >= 50)
	})
}

func TestConcurrentPool_LockContextReturnsOnCancellation(t *testing.T) {
	pool, err := NewConcurrentPool(1)
	assert.NoError(t, err)
	pool.Lock()
	defer pool.Unlock()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	assert.Error(t, pool.LockContext(ctx))
	assert.Equal(t, 1, pool.get())
}
