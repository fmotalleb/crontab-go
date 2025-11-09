package event

import (
	"context"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/global"
)

func init() {
	eg.Register(newWebEventGenerator)
}

func newWebEventGenerator(_ *zap.Logger, cfg *config.JobEvent) (abstraction.EventGenerator, bool) {
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
			ed.Emit(ctx, event)
		},
	)
	<-ctx.Done()
}
