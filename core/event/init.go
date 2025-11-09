package event

import (
	"context"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
)

func init() {
	eg.Register(newInitGenerator)
}

func newInitGenerator(_ *zap.Logger, cfg *config.JobEvent) (abstraction.EventGenerator, bool) {
	if cfg.OnInit {
		return &Init{}, true
	}
	return nil, false
}

type Init struct{}

// BuildTickChannel implements abstraction.Scheduler.
func (c *Init) BuildTickChannel(ed abstraction.EventDispatcher) {
	ctx := context.Background()
	ed.Emit(ctx, NewMetaData("init", map[string]any{}))
}
