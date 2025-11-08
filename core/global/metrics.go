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
	tag := name
	for _, label := range []ctxutils.ContextKey{ctxutils.JobKey} {
		if value, ok := ctx.Value(label).(string); ok {
			labels[string(label)] = value
			tag = fmt.Sprintf("%s,%s=%s", tag, label, value)
		}
	}
	if c, ok := c.countersValue[tag]; ok {
		return c
	}
	c.countersValue[tag] = concurrency.NewLockedValue[float64](0)
	c.counters[tag] = promauto.NewCounterFunc(
		prometheus.CounterOpts{
			Name:        name,
			ConstLabels: labels,
			Help:        help,
			Namespace:   "crontab_go",
		},
		func() float64 {
			item, ok := c.countersValue[tag]
			if !ok {
				return 0.0
			}
			ans := item.Get()
			// a counter should not reset after it's collected by Prometheus.
			// item.Set(0)
			return ans
		},
	)
	return c.MetricCounter(ctx, name, help, labels)
}

func (c *Context) CountSignals(ctx context.Context, name string, signal abstraction.EventChannel, help string, labels prometheus.Labels) abstraction.EventChannel {
	counter := c.MetricCounter(ctx, name, help, labels)
	out := make(abstraction.EventEmitChannel)
	go func() {
		for c := range signal {
			counter.Set(counter.Get() + 1)
			out <- c
		}
	}()
	return out
}
