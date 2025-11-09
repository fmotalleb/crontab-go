// Package event contains all event emitters supported by this package.
package event

import (
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/global"
)

func init() {
	eg.Register(newCronGenerator)
}

func newCronGenerator(log *zap.Logger, cfg *config.JobEvent) (abstraction.EventGenerator, bool) {
	if cfg.Cron != "" {
		return NewCron(cfg.Cron, global.Get[*cron.Cron](), log), true
	}
	return nil, false
}

type Cron struct {
	cronSchedule string
	logger       *zap.Logger
	cron         *cron.Cron
	entry        *cron.EntryID
}

func NewCron(schedule string, c *cron.Cron, logger *zap.Logger) abstraction.EventGenerator {
	cron := &Cron{
		cronSchedule: schedule,
		cron:         c,
		logger: logger.
			With(
				zap.String("scheduler", "cron"),
				zap.String("cron", schedule),
			),
	}
	return cron
}

// BuildTickChannel implements abstraction.Scheduler.
func (c *Cron) BuildTickChannel(ed abstraction.EventDispatcher) {
	if c.entry != nil {
		c.logger.Fatal("already built the ticker channel")
	}
	notifyChan := make(chan abstraction.Event)
	schedule, err := config.DefaultCronParser.Parse(c.cronSchedule)
	if err != nil {
		c.logger.Warn("cannot initialize cron", zap.Error(err))
	} else {
		entry := c.cron.Schedule(
			schedule,
			&cronJob{
				logger:    c.logger,
				scheduler: c.cronSchedule,
				notify:    notifyChan,
			},
		)
		c.entry = &entry
	}
	ctx, cancel := context.WithCancel(global.CTX().Context)
	defer cancel()
	for e := range notifyChan {
		ed.Emit(ctx, e)
	}
}

type cronJob struct {
	logger    *zap.Logger
	scheduler string
	notify    chan<- abstraction.Event
}

func (j *cronJob) Run() {
	j.logger.Debug("cron tick received")
	j.notify <- NewMetaData(
		"cron",
		map[string]any{
			"schedule": j.scheduler,
		},
	)
}
