package global

import (
	"context"
	"maps"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/core/concurrency"
)

const namespace = "crontab_go"

type Metrics = map[string]prometheus.CounterVec

var collectors = concurrency.NewLockedValue(make(Metrics, 0))

func IncMetric(name string, help string, labels prometheus.Labels) {
	collectors.Operate(
		func(old Metrics) Metrics {
			m := maps.Clone(old)
			if vec, ok := m[name]; ok {
				if olderVec, err := vec.GetMetricWith(labels); err != nil {
					olderVec.Add(1)
				} else {
					c := vec.With(labels)
					c.Add(1)
				}
				return m
			}

			keys := make([]string, 0, len(labels))
			for key := range labels {
				keys = append(keys, key)
			}
			vec := *promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      name,
					Help:      help,
				},
				keys,
			)
			counter := vec.With(labels)
			counter.Add(1)
			m[name] = vec
			return m
		})
}

func CountSignals(signal abstraction.EventDispatcher, name string, help string, labels prometheus.Labels) {
	signal.AddListener(func(_ context.Context, _ abstraction.Event) {
		IncMetric(name, help, labels)
	})
}
