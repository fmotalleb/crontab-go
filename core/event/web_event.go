package event

import (
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/global"
)

func init() {
	eg.Register(newWebEventGenerator)
}

func newWebEventGenerator(_ *logrus.Entry, cfg *config.JobEvent) (abstraction.EventGenerator, bool) {
	if cfg.WebEvent != "" {
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
func (w *WebEventListener) BuildTickChannel() abstraction.EventChannel {
	notifyChan := make(abstraction.EventEmitChannel)
	global.CTX().AddEventListener(
		w.event, func(params map[string]any) {
			notifyChan <- NewMetaData(
				"web",
				map[string]any{
					"event":  w.event,
					"params": params,
				})
		},
	)
	return notifyChan
}
