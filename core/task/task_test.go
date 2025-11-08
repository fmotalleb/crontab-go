package task_test

// import (
// 	"context"
// 	"testing"

// 	"github.com/alecthomas/assert/v2"
// 	"github.com/sirupsen/logrus"

// 	"github.com/FMotalleb/crontab-go/config"
// 	"github.com/FMotalleb/crontab-go/core/task"
// 	"github.com/FMotalleb/crontab-go/ctxutils"
// 	mocklogger "github.com/FMotalleb/crontab-go/logger/mock_logger"
// )

// func TestCompileTask_NonExistingTask(t *testing.T) {
// 	ctx := t.Context()
// 	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	taskConfig := config.Task{}
// 	assert.Panics(
// 		t,
// 		func() {
// 			task.Build(ctx, log, taskConfig)
// 		},
// 	)
// }

// func TestCompileTask_GetTask(t *testing.T) {
// 	ctx := t.Context()
// 	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	taskConfig := config.Task{
// 		Get: "test",
// 	}
// 	exe := task.Build(ctx, log, taskConfig)
// 	assert.NotEqual(t, nil, exe)
// }

// func TestCompileTask_CommandTask(t *testing.T) {
// 	ctx := t.Context()
// 	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	taskConfig := config.Task{
// 		Command: "test",
// 	}
// 	exe := task.Build(ctx, log, taskConfig)
// 	assert.NotEqual(t, exe, nil)
// }

// func TestCompileTask_PostTask(t *testing.T) {
// 	ctx := t.Context()
// 	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	taskConfig := config.Task{
// 		Post: "test",
// 	}
// 	exe := task.Build(ctx, log, taskConfig)
// 	assert.NotEqual(t, exe, nil)
// }

// func TestCompileTask_WithHooks(t *testing.T) {
// 	ctx := t.Context()
// 	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	taskConfig := config.Task{
// 		Command: "test",
// 		OnDone: []config.Task{
// 			{
// 				Command: "test",
// 			},
// 		},
// 		OnFail: []config.Task{
// 			{
// 				Command: "test",
// 			},
// 		},
// 	}
// 	exe := task.Build(ctx, log, taskConfig)
// 	assert.NotEqual(t, exe, nil)
// }
