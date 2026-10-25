package jobs

import (
	"context"
	"maps"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/core/concurrency"
	"github.com/fmotalleb/crontab-go/core/global"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

var tracer = otel.Tracer("crontab-go/task_handler")

type taskTracker struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	stopped bool
}

// start registers an active execution unless shutdown has begun.
// A true result must be paired with one call to done.
func (t *taskTracker) start() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopped {
		return false
	}
	t.wg.Add(1)
	return true
}

// done marks one registered execution, including its hooks, as finished.
func (t *taskTracker) done() {
	t.wg.Done()
}

// stop permanently prevents new executions from being registered.
func (t *taskTracker) stop() {
	t.mu.Lock()
	t.stopped = true
	t.mu.Unlock()
}

// stopAndWait prevents new registrations and waits for active executions.
// It returns nil when they finish or ctx.Err() if the wait is canceled.
func (t *taskTracker) stopAndWait(ctx context.Context) error {
	t.stop()
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

var (
	trackersMu sync.Mutex
	trackers   []*taskTracker
)

// WaitForShutdown prevents new task launches for currently registered handlers
// and waits for their active tasks and hooks. It returns ctx.Err() if waiting is
// canceled; active work is not canceled by this function.
func WaitForShutdown(ctx context.Context) error {
	trackersMu.Lock()
	active := append([]*taskTracker(nil), trackers...)
	trackersMu.Unlock()
	for _, tracker := range active {
		tracker.stop()
	}
	for _, tracker := range active {
		if err := tracker.stopAndWait(ctx); err != nil {
			return err
		}
	}
	return nil
}

// taskHandler registers a listener that launches tasks with event metadata and
// a separate copy of variables for each execution. Dispatch waits for concurrency
// slots and stops launching when the event context is canceled or shutdown begins.
func taskHandler(
	logger *zap.Logger,
	ed abstraction.EventDispatcher,
	tasks []abstraction.Executable,
	doneHooks []abstraction.Executable,
	failHooks []abstraction.Executable,
	lock *concurrency.ConcurrentPool,
	jobName string,
	variables map[string]string,
) {
	logger.Debug("Spawning task handler")
	tracker := &taskTracker{}
	trackersMu.Lock()
	trackers = append(trackers, tracker)
	trackersMu.Unlock()
	ed.AddListener(func(ctx context.Context, e abstraction.Event) {
		logger.Debug("Signal Received")
		for _, task := range tasks {
			if ctx.Err() != nil {
				return
			}
			ctxInternal := context.WithValue(ctx, ctxutils.EventData, e)
			ctxInternal = context.WithValue(ctxInternal, ctxutils.Vars, maps.Clone(variables))
			if err := lock.LockContext(ctxInternal); err != nil {
				logger.Warn("task dispatch canceled while waiting for a concurrency slot", zap.Error(err))
				return
			}
			if !tracker.start() {
				lock.Unlock()
				return
			}
			go executeTask(ctxInternal, task, doneHooks, failHooks, lock, jobName, tracker)
		}
	})
}

// executeTask runs a task and then the job's failure or done hooks according to
// its result, ignoring hook errors. It recovers task and hook panics and releases
// the caller's acquired pool slot and tracker registration on return.
func executeTask(
	c context.Context,
	task abstraction.Executable,
	doneHooks []abstraction.Executable,
	failHooks []abstraction.Executable,
	lock *concurrency.ConcurrentPool,
	jobName string,
	tracker *taskTracker,
) {
	defer lock.Unlock()
	defer tracker.done()
	defer func() {
		if r := recover(); r != nil {
			global.Logger("task_handler").Error("task panicked", zap.Any("recover", r))
		}
	}()

	ctx, span := tracer.Start(c, "job."+jobName+"/task.execute",
		trace.WithAttributes(
			attribute.String("job.name", jobName),
		),
	)
	defer span.End()

	ctx = context.WithValue(ctx, ctxutils.TaskKey, task)
	ctx = context.WithValue(ctx, ctxutils.JobKey, jobName)
	err := task.Execute(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		for _, hook := range failHooks {
			_ = hook.Execute(ctx)
		}
	} else {
		span.SetStatus(codes.Ok, "")
		for _, hook := range doneHooks {
			_ = hook.Execute(ctx)
		}
	}
}
