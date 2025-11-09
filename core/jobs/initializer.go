// Package jobs implements the main functionality for the jobs in the application
package jobs

import (
	"context"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/event"
	"github.com/fmotalleb/crontab-go/core/task"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

func initEventSignal(ed abstraction.EventDispatcher, events []abstraction.EventGenerator, logger *zap.Logger) {
	for _, ev := range events {
		go ev.BuildTickChannel(ed)
	}
	logger.Debug("signals initialized")
}

func initTasks(job config.JobConfig, logger *zap.Logger) ([]abstraction.Executable, []abstraction.Executable, []abstraction.Executable) {
	tasks := make([]abstraction.Executable, 0, len(job.Tasks))
	doneHooks := make([]abstraction.Executable, 0, len(job.Hooks.Done))
	failHooks := make([]abstraction.Executable, 0, len(job.Hooks.Failed))

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxutils.JobKey, job.Name)
	for _, t := range job.Tasks {
		tasks = append(tasks, task.Build(ctx, logger, t))
	}
	logger.Debug("Compiled Tasks")
	for _, t := range job.Hooks.Done {
		doneHooks = append(doneHooks, task.Build(ctx, logger, t))
	}
	logger.Debug("Compiled Hooks.Done")
	for _, t := range job.Hooks.Failed {
		failHooks = append(failHooks, task.Build(ctx, logger, t))
	}
	logger.Debug("Compiled Hooks.Fail")
	return tasks, doneHooks, failHooks
}

func initEvents(job config.JobConfig, logger *zap.Logger) []abstraction.EventGenerator {
	events := make([]abstraction.EventGenerator, 0, len(job.Events))
	for _, sh := range job.Events {
		events = append(events, event.Build(logger, &sh))
	}
	return events
}
