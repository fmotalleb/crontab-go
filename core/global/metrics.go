package global

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/core/concurrency"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

func (c *Context) MetricCounter(
	ctx context.Context,
	name string,
	help string,
	labels prometheus.Labels,
) *concurrency.LockedValue[float64] {
	// ensure labels map exists so we can safely add const labels
	if labels == nil {
		labels = prometheus.Labels{}
	}
	// attach job info from context if present and build a unique tag
	tag := name
	if value, ok := ctx.Value(ctxutils.JobKey).(string); ok {
		labels[string(ctxutils.JobKey)] = value
		tag = fmt.Sprintf("%s,%s=%s", tag, ctxutils.JobKey, value)
	}

	// fast path: if already created, return it
	c.mu.RLock()
	if existing, ok := c.countersValue[tag]; ok {
		c.mu.RUnlock()
		return existing
	}
	c.mu.RUnlock()

	// create new locked value and register Prometheus counter func
	lv := concurrency.NewLockedValue[float64](0)

	c.mu.Lock()
	// double-check inside write lock
	if existing, ok := c.countersValue[tag]; ok {
		c.mu.Unlock()
		return existing
	}
	c.countersValue[tag] = lv
	// counter func should safely read the locked value under the context mutex
	c.counters[tag] = promauto.NewCounterFunc(
		prometheus.CounterOpts{
			Name:        name,
			ConstLabels: labels,
			Help:        help,
			Namespace:   "crontab_go",
		},
		func() float64 {
			c.mu.RLock()
			item, ok := c.countersValue[tag]
			c.mu.RUnlock()
			if !ok {
				return 0.0
			}
			return item.Get()
		},
	)
	c.mu.Unlock()

	return lv
}

func (c *Context) CountSignals(ctx context.Context, name string, signal abstraction.EventDispatcher, help string, labels prometheus.Labels) {
	counter := c.MetricCounter(ctx, name, help, labels)
	signal.AddListener(func(_ context.Context, _ abstraction.Event) {
		counter.Set(counter.Get() + 1)
	})
}
