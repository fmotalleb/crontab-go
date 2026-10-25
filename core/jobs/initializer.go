// Package jobs implements the main functionality for the jobs in the application
package jobs

import (
	"context"
	"fmt"

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

func initTasks(job config.JobConfig, logger *zap.Logger) ([]abstraction.Executable, []abstraction.Executable, []abstraction.Executable, error) {
	tasks := make([]abstraction.Executable, 0, len(job.Tasks))
	doneHooks := make([]abstraction.Executable, 0, len(job.Hooks.Done))
	failHooks := make([]abstraction.Executable, 0, len(job.Hooks.Failed))

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxutils.JobKey, job.Name)
	for _, t := range job.Tasks {
		exe, err := task.Build(ctx, logger, t)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to build task: %w", err)
		}
		tasks = append(tasks, exe)
	}
	logger.Debug("Compiled Tasks")
	for _, t := range job.Hooks.Done {
		exe, err := task.Build(ctx, logger, t)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to build done hook: %w", err)
		}
		doneHooks = append(doneHooks, exe)
	}
	logger.Debug("Compiled Hooks.Done")
	for _, t := range job.Hooks.Failed {
		exe, err := task.Build(ctx, logger, t)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to build fail hook: %w", err)
		}
		failHooks = append(failHooks, exe)
	}
	logger.Debug("Compiled Hooks.Fail")
	return tasks, doneHooks, failHooks, nil
}

func initEvents(job config.JobConfig, logger *zap.Logger) ([]abstraction.EventGenerator, error) {
	events := make([]abstraction.EventGenerator, 0, len(job.Events))
	for _, sh := range job.Events {
		ev, err := event.Build(logger, &sh)
		if err != nil {
			return nil, fmt.Errorf("failed to build event: %w", err)
		}
		events = append(events, ev)
	}
	return events, nil
}
