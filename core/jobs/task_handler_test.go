package jobs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/maniartech/signals"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/core/concurrency"
	"github.com/fmotalleb/crontab-go/core/event"
)

type trackingExecutable struct {
	mu       sync.Mutex
	active   int
	max      int
	finished chan struct{}
}

func (e *trackingExecutable) Execute(context.Context) error {
	e.mu.Lock()
	e.active++
	if e.active > e.max {
		e.max = e.active
	}
	e.mu.Unlock()

	time.Sleep(10 * time.Millisecond)

	e.mu.Lock()
	e.active--
	e.mu.Unlock()
	e.finished <- struct{}{}
	return nil
}

func (*trackingExecutable) GetMeta() map[string]string                             { return nil }
func (*trackingExecutable) SetMetaName(string)                                     {}
func (*trackingExecutable) SetDoneHooks(context.Context, []abstraction.Executable) {}
func (*trackingExecutable) SetFailHooks(context.Context, []abstraction.Executable) {}
func (*trackingExecutable) Cancel()                                                {}

func TestTaskHandlerAcquiresPermitsBeforeSpawning(t *testing.T) {
	const count = 8
	pool, err := concurrency.NewConcurrentPool(2)
	assert.NoError(t, err)
	signal := signals.NewSync[abstraction.Event]()
	executables := make([]abstraction.Executable, count)
	tracker := &trackingExecutable{finished: make(chan struct{}, count)}
	for i := range executables {
		executables[i] = tracker
	}
	taskHandler(zap.NewNop(), signal, executables, nil, nil, pool, "bounded", nil)

	signal.Emit(context.Background(), event.NewMetaData("test", nil))
	for range count {
		select {
		case <-tracker.finished:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for dispatched tasks")
		}
	}
	assert.NoError(t, WaitForShutdown(t.Context()))

	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	assert.Equal(t, 2, tracker.max)
}
