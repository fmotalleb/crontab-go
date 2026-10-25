// Package global contains global state management logics
package global

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/fmotalleb/crontab-go/ctxutils"
)

func ctxKey(prefix string, key string) ctxutils.ContextKey {
	return ctxutils.ContextKey(fmt.Sprintf("%s:%s", prefix, key))
}

func CTX() *Context {
	return c()
}

var c = sync.OnceValue(newGlobalContext)

type (
	EventListenerMap = map[string][]func(map[string]any)
	Context          struct {
		context.Context
		mu *sync.RWMutex
	}
)

// newGlobalContext creates a context canceled by an interrupt or SIGTERM,
// with an environment-configured logger and an empty event listener registry.
// It panics if logger initialization fails.
func newGlobalContext() *Context {
	ctx := context.Background()
	ctx, err := log.WithNewEnvLogger(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %w", err))
	}
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	ctx = context.WithValue(
		ctx,
		ctxutils.EventListeners,
		EventListenerMap{},
	)
	return &Context{
		Context: ctx,
		mu:      new(sync.RWMutex),
	}
}

// Value returns the value associated with key from the embedded context,
// synchronizing the read against writers that replace it under the mutex.
func (c *Context) Value(key any) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Context.Value(key)
}

// Done returns the embedded context's done channel, synchronized with the
// writers that replace the embedded context under the mutex.
func (c *Context) Done() <-chan struct{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Context.Done()
}

// Err returns the embedded context's error, synchronized with the writers
// that replace the embedded context under the mutex.
func (c *Context) Err() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Context.Err()
}

// Deadline returns the embedded context's deadline, synchronized with the
// writers that replace the embedded context under the mutex.
func (c *Context) Deadline() (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Context.Deadline()
}

// EventListeners returns a snapshot of the listener registry.
// Changing the returned map or handler slices does not change registered listeners.
func (c *Context) EventListeners() EventListenerMap {
	c.mu.RLock()
	defer c.mu.RUnlock()
	registered := c.Context.Value(ctxutils.EventListeners).(EventListenerMap)
	listeners := make(EventListenerMap, len(registered))
	for event, handlers := range registered {
		listeners[event] = append(make([]func(map[string]any), 0, len(handlers)), handlers...)
	}
	return listeners
}

func (c *Context) AddEventListener(event string, listener func(map[string]any)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	listeners := c.Context.Value(ctxutils.EventListeners).(EventListenerMap)
	listeners[event] = append(listeners[event], listener)
	c.Context = context.WithValue(c.Context, ctxutils.EventListeners, listeners)
}

func getTypename[T any](item T) string {
	return reflect.TypeOf(item).String()
}

func Put[T any](item T) {
	name := getTypename(item)
	ctx := c()
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.Context = context.WithValue(ctx.Context, ctxKey("typed", name), item)
}

func Get[T any]() T {
	var zero T // Default zero value for type T
	name := reflect.TypeOf(zero).String()
	ctx := c()
	ctx.mu.RLock()
	value := ctx.Context.Value(ctxKey("typed", name))
	ctx.mu.RUnlock()
	if value == nil {
		return zero
	}

	// Type assertion to ensure the value is of type T
	castedValue, ok := value.(T)
	if !ok {
		return zero
	}
	return castedValue
}

func Logger(name string) *zap.Logger {
	ctx := c()
	ctx.mu.RLock()
	logger := log.Of(ctx.Context).Named(name)
	ctx.mu.RUnlock()
	return logger
}

// AttachOTelCore tees an otelzap core onto the global logger, enabling
// log records to be exported via OTLP alongside the console sink.
func AttachOTelCore(otelCore zapcore.Core) {
	ctx := c()
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	existing := log.Of(ctx.Context)
	combined := zapcore.NewTee(existing.Core(), otelCore)
	replaced := existing.WithOptions(zap.WrapCore(func(_ zapcore.Core) zapcore.Core {
		return combined
	}))
	ctx.Context = log.WithLogger(ctx.Context, replaced)
}
