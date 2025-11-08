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
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/core/concurrency"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

func ctxKey(prefix string, key string) ctxutils.ContextKey {
	return ctxutils.ContextKey(fmt.Sprintf("%s:%s", prefix, key))
}

func CTX() *Context {
	return c
}

var c = newGlobalContext()

type (
	EventListenerMap = map[string][]func(map[string]any)
	Context          struct {
		context.Context
		lock          *sync.RWMutex
		countersValue map[string]*concurrency.LockedValue[float64]
		counters      map[string]prometheus.CounterFunc
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
		Context:       ctx,
		lock:          new(sync.RWMutex),
		countersValue: make(map[string]*concurrency.LockedValue[float64]),
		counters:      make(map[string]prometheus.CounterFunc),
	}
}

func (c *Context) EventListeners() EventListenerMap {
	c.lock.RLock()
	defer c.lock.RUnlock()
	listeners := c.Value(ctxutils.EventListeners)
	return listeners.(EventListenerMap)
}

func (c *Context) AddEventListener(event string, listener func(map[string]any)) {
	c.lock.Lock()
	defer c.lock.Unlock()
	listeners := c.Value(ctxutils.EventListeners).(EventListenerMap)
	listeners[event] = append(listeners[event], listener)
	c.Context = context.WithValue(c.Context, ctxutils.EventListeners, listeners)
}

func getTypename[T any](item T) string {
	return reflect.TypeOf(item).String()
}

func Put[T any](item T) {
	name := getTypename(item)
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Context = context.WithValue(c.Context, ctxKey("typed", name), item)
}

func Get[T any]() T {
	var zero T // Default zero value for type T
	name := reflect.TypeOf(zero).String()
	println(name)
	value := c.Value(ctxKey("typed", name))
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
	return log.Of(c).Named(name)
}
