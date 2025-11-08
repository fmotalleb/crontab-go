package jobs

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/concurrency"
	"github.com/FMotalleb/crontab-go/core/global"
	"github.com/FMotalleb/crontab-go/ctxutils"
)

func InitializeJobs() {
	log := global.Logger("Initializer")
	for _, job := range cmd.CFG.Jobs {
		if job.Disabled {
			log.Warn("job is disabled", zap.String("job.name", job.Name))
			continue
		}
		// Setting default value of concurrency
		if job.Concurrency == 0 {
			job.Concurrency = 1
		}
		c := global.CTX().Context
		c = context.WithValue(c, ctxutils.JobKey, job.Name)

		lock, err := concurrency.NewConcurrentPool(job.Concurrency)
		if err != nil {
			log.Panic("failed to validate job", zap.String("job.name", job.Name), zap.Error(err))
		}
		logger := log.With(
			zap.String("job.name", job.Name),
			zap.Uint("job.concurrency", job.Concurrency),
		)
		if err := job.Validate(logger); err != nil {
			log.Panic("failed to validate job", zap.String("job", job.Name), zap.Error(err))
		}

		signal := buildSignal(*job, logger)
		signal = global.CTX().CountSignals(c, "events", signal, "amount of events dispatched for this job", prometheus.Labels{})
		tasks, doneHooks, failHooks := initTasks(*job, logger)
		logger.Debug("Tasks initialized")

		go taskHandler(c, logger, signal, tasks, doneHooks, failHooks, lock)
		logger.Debug("EventLoop initialized")
	}
	log.Info("Jobs Are Ready")
}

func buildSignal(job config.JobConfig, logger *zap.Logger) abstraction.EventChannel {
	events := initEvents(job, logger)
	logger.Debug("Events initialized")

	signal := initEventSignal(events, logger)

	return signal
}
