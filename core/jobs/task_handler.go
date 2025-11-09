package jobs

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

func taskHandler(
	logger *zap.Logger,
	ed abstraction.EventDispatcher,
	tasks []abstraction.Executable,
	doneHooks []abstraction.Executable,
	failHooks []abstraction.Executable,
	lock sync.Locker,
) {
	logger.Debug("Spawning task handler")
	ed.AddListener(func(ctx context.Context, e abstraction.Event) {
		logger.Debug("Signal Received")
		for _, task := range tasks {
			ctxInternal := context.WithValue(ctx, ctxutils.EventData, e)
			go executeTask(ctxInternal, task, doneHooks, failHooks, lock)
		}
	})
}

func executeTask(
	c context.Context,
	task abstraction.Executable,
	doneHooks []abstraction.Executable,
	failHooks []abstraction.Executable,
	lock sync.Locker,
) {
	lock.Lock()
	defer lock.Unlock()
	ctx := context.WithValue(c, ctxutils.TaskKey, task)
	err := task.Execute(ctx)
	switch err {
	case nil:
		for _, task := range doneHooks {
			_ = task.Execute(ctx)
		}
	default:
		for _, task := range failHooks {
			_ = task.Execute(ctx)
		}
	}
}
