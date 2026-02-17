package global

import (
	"sync/atomic"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestEventListenersReturnsCopy(t *testing.T) {
	eventName := "test-event-listeners-copy"
	var called atomic.Int64
	listener := func(map[string]any) {
		called.Add(1)
	}

	CTX().AddEventListener(eventName, listener)
	firstRead := CTX().EventListeners()
	firstRead[eventName] = []func(map[string]any){}

	secondRead := CTX().EventListeners()
	assert.True(t, len(secondRead[eventName]) > 0)

	// Ensure original listener remains callable.
	secondRead[eventName][0](map[string]any{})
	assert.Equal(t, int64(1), called.Load())
}
