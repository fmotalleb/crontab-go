package event

import (
	"go.uber.org/zap"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
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
func (c *Init) BuildTickChannel() abstraction.EventChannel {
	notifyChan := make(abstraction.EventEmitChannel)

	go func() {
		notifyChan <- NewMetaData("init", map[string]any{})
		close(notifyChan)
	}()

	return notifyChan
}
