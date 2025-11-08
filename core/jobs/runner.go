package jobs

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/cmd"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/concurrency"
	"github.com/fmotalleb/crontab-go/core/global"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

func InitializeJobs() {
	log := global.Logger("Cron")
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
		if err := job.Validate(logger.Named("Validator")); err != nil {
			log.Panic("failed to validate job", zap.String("job", job.Name), zap.Error(err))
		}

		signal := buildSignal(*job, logger.Named("SignalGen"))
		signal = global.CTX().CountSignals(c, "events", signal, "amount of events dispatched for this job", prometheus.Labels{})
		tasks, doneHooks, failHooks := initTasks(*job, logger.Named("Task"))
		logger.Debug("Tasks initialized")

		go taskHandler(c, logger.Named("TaskRunner"), signal, tasks, doneHooks, failHooks, lock)
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
