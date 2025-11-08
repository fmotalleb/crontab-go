// Package jobs implements the main functionality for the jobs in the application
package jobs

import (
	"context"

	"go.uber.org/zap"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/event"
	"github.com/FMotalleb/crontab-go/core/task"
	"github.com/FMotalleb/crontab-go/core/utils"
	"github.com/FMotalleb/crontab-go/ctxutils"
)

func initEventSignal(events []abstraction.EventGenerator, logger *zap.Logger) abstraction.EventChannel {
	signals := make([]abstraction.EventChannel, 0, len(events))
	for _, ev := range events {
		signals = append(signals, ev.BuildTickChannel())
	}
	logger.Debug("Signals Built")
	signal := utils.ZipChannels(signals...)
	return signal
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
