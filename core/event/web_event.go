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
	WebEventsMetricName = "webserver_events"
	WebEventsMetricHelp = "amount of events dispatched using webserver"
)

func init() {
	eg.Register(newWebEventGenerator)
}

func newWebEventGenerator(_ *zap.Logger, cfg *config.JobEvent) (abstraction.EventGenerator, bool) {
	if cfg.WebEvent != "" {
		global.RegisterCounter(
			WebEventsMetricName,
			WebEventsMetricHelp,
			prometheus.Labels{"event_name": cfg.WebEvent},
		)
		return NewWebEventListener(cfg.WebEvent), true
	}
	return nil, false
}

type WebEventListener struct {
	event string
}

func NewWebEventListener(event string) abstraction.EventGenerator {
	return &WebEventListener{
		event: event,
	}
}

// BuildTickChannel implements abstraction.Scheduler.
func (w *WebEventListener) BuildTickChannel(ed abstraction.EventDispatcher) {
	ctx, cancel := context.WithCancel(global.CTX())
	defer cancel()
	global.CTX().AddEventListener(
		w.event, func(params map[string]any) {
			event := NewMetaData(
				"web",
				map[string]any{
					"event":  w.event,
					"params": params,
				})
			global.IncMetric(
				WebEventsMetricName,
				WebEventsMetricHelp,
				prometheus.Labels{"file": w.event},
			)
			ed.Emit(ctx, event)
		},
	)
	<-ctx.Done()
}
