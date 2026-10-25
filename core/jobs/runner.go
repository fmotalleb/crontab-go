package jobs

import (
	"fmt"

	"github.com/maniartech/signals"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/fmotalleb/go-tools/debouncer"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/concurrency"
	"github.com/fmotalleb/crontab-go/core/event"
	"github.com/fmotalleb/crontab-go/core/global"
)

func InitializeJobs(jobs []*config.JobConfig, variables map[string]string) error {
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
			return fmt.Errorf("job %q: %w", job.Name, err)
		}
		logger := log.With(
			zap.String("job.name", job.Name),
			zap.Uint("job.concurrency", job.Concurrency),
		)
		if validationErr := job.Validate(logger.Named("Validator")); validationErr != nil {
			return fmt.Errorf("failed to validate job %q: %w", job.Name, validationErr)
		}
		var signal abstraction.EventDispatcher = signals.NewSync[abstraction.Event]()
		var debounceAttr attribute.KeyValue
		if job.Debounce > 0 {
			debounceAttr = attribute.String("event.debounce", job.Debounce.String())
		}
		signal = event.NewSpanDispatcher(signal, debounceAttr)
		if job.Debounce > 0 {
			signal = debouncer.NewDebouncedSignal(signal, job.Debounce)
		}
		global.CountSignals(signal,
			"events",
			"amount of events dispatched for this job",
			prometheus.Labels{
				"job": job.Name,
			},
		)
		tasks, doneHooks, failHooks, err := initTasks(*job, logger.Named("Task"))
		if err != nil {
			return fmt.Errorf("job %q: %w", job.Name, err)
		}
		logger.Debug("Tasks initialized")

		taskHandler(logger.Named("TaskRunner"), signal, tasks, doneHooks, failHooks, lock, job.Name, variables)
		if err := buildSignal(signal, *job, logger.Named("SignalGen")); err != nil {
			return fmt.Errorf("job %q: %w", job.Name, err)
		}

		logger.Debug("EventLoop initialized")
	}
	log.Info("Jobs Are Ready")
	return nil
}

func buildSignal(ed abstraction.EventDispatcher, job config.JobConfig, logger *zap.Logger) error {
	events, err := initEvents(job, logger)
	if err != nil {
		return err
	}
	logger.Debug("Events initialized")

	initEventSignal(ed, events, logger)
	return nil
}
