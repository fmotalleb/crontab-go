package jobs

import (
	"github.com/maniartech/signals"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/fmotalleb/go-tools/debouncer"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/concurrency"
	"github.com/fmotalleb/crontab-go/core/global"
)

func InitializeJobs(jobs []*config.JobConfig) {
	log := global.Logger("Cron")
	for _, job := range jobs {
		if job.Disabled {
			log.Warn("job is disabled", zap.String("job.name", job.Name))
			continue
		}
		// Setting default value of concurrency
		if job.Concurrency == 0 {
			job.Concurrency = 1
		}

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
		var signal abstraction.EventDispatcher = signals.NewSync[abstraction.Event]()
		if job.Debounce >= 0 {
			signal = debouncer.NewDebouncedSignal(signal, job.Debounce)
		}
		global.CountSignals(signal,
			"events",
			"amount of events dispatched for this job",
			prometheus.Labels{
				"job": job.Name,
			},
		)
		tasks, doneHooks, failHooks := initTasks(*job, logger.Named("Task"))
		logger.Debug("Tasks initialized")

		taskHandler(logger.Named("TaskRunner"), signal, tasks, doneHooks, failHooks, lock)
		buildSignal(signal, *job, logger.Named("SignalGen"))

		logger.Debug("EventLoop initialized")
	}
	log.Info("Jobs Are Ready")
}

func buildSignal(ed abstraction.EventDispatcher, job config.JobConfig, logger *zap.Logger) {
	events := initEvents(job, logger)
	logger.Debug("Events initialized")

	initEventSignal(ed, events, logger)
}
