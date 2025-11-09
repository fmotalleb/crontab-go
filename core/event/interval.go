package event

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/global"
)

func init() {
	eg.Register(newIntervalGenerator)
}

func newIntervalGenerator(log *zap.Logger, cfg *config.JobEvent) (abstraction.EventGenerator, bool) {
	if cfg.Interval > 0 {
		return NewInterval(cfg.Interval, log), true
	}
	return nil, false
}

type Interval struct {
	duration time.Duration
	logger   *zap.Logger
	ticker   *time.Ticker
}

func NewInterval(schedule time.Duration, logger *zap.Logger) abstraction.EventGenerator {
	return &Interval{
		duration: schedule,
		logger: logger.
			With(
				zap.String("scheduler", "interval"),
				zap.Duration("interval", schedule),
			),
	}
}

// BuildTickChannel implements abstraction.Scheduler.
func (c *Interval) BuildTickChannel(ed abstraction.EventDispatcher) {
	if c.ticker != nil {
		c.logger.Fatal("already built the ticker channel")
	}

	c.ticker = time.NewTicker(c.duration)
	ctx, cancel := context.WithCancel(global.CTX())
	defer cancel()
	for i := range c.ticker.C {
		event := NewMetaData(
			"interval",
			map[string]any{
				"interval": c.duration.String(),
				"time":     i.Format(time.RFC3339),
			},
		)
		ed.Emit(ctx, event)
	}
}
