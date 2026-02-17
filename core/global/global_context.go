// Package global contains global state management logics
package global

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

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

func newGlobalContext() *Context {
	ctx := context.Background()
	ctx, err := log.WithNewEnvLogger(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %w", err))
	}
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, os.Kill)
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

func (c *Context) EventListeners() EventListenerMap {
	c.mu.RLock()
	defer c.mu.RUnlock()
	listeners := c.Value(ctxutils.EventListeners).(EventListenerMap)
	copyMap := make(EventListenerMap, len(listeners))
	for key, ls := range listeners {
		clone := make([]func(map[string]any), len(ls))
		copy(clone, ls)
		copyMap[key] = clone
	}
	return copyMap
}

func (c *Context) AddEventListener(event string, listener func(map[string]any)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	listeners := c.Value(ctxutils.EventListeners).(EventListenerMap)
	listeners[event] = append(listeners[event], listener)
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
	value := c().Value(ctxKey("typed", name))
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
	return log.Of(c()).Named(name)
}
