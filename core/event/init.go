package event

import (
	"context"

	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/global"
)

const (
	InitEventsMetricName = "init"
	InitEventsMetricHelp = "amount of events dispatched using init"
)

func init() {
	eg.Register(newInitGenerator)
}

func newInitGenerator(_ *zap.Logger, cfg *config.JobEvent) (abstraction.EventGenerator, bool) {
	if cfg.OnInit {
		global.RegisterCounter(
			InitEventsMetricName,
			InitEventsMetricHelp,
			prometheus.Labels{"init": "once"},
		)
		return &Init{}, true
	}
	return nil, false
}

type Init struct{}

// BuildTickChannel implements abstraction.Scheduler.
func (c *Init) BuildTickChannel(ed abstraction.EventDispatcher) {
	ctx := context.Background()
	ed.Emit(ctx, NewMetaData("init", map[string]any{}))
	global.IncMetric(
		InitEventsMetricName,
		InitEventsMetricHelp,
		prometheus.Labels{"init": "once"},
	)
}
