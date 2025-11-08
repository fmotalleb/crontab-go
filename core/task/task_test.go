package task_test

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/task"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

func TestCompileTask_NonExistingTask(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{}
	assert.Panics(
		t,
		func() {
			task.Build(ctx, zap.NewNop(), taskConfig)
		},
	)
}

func TestCompileTask_GetTask(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{
		Get: "test",
	}
	exe := task.Build(ctx, zap.NewNop(), taskConfig)
	assert.NotEqual(t, nil, exe)
}

func TestCompileTask_CommandTask(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{
		Command: "test",
	}
	exe := task.Build(ctx, zap.NewNop(), taskConfig)
	assert.NotEqual(t, exe, nil)
}

func TestCompileTask_PostTask(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{
		Post: "test",
	}
	exe := task.Build(ctx, zap.NewNop(), taskConfig)
	assert.NotEqual(t, exe, nil)
}

func TestCompileTask_WithHooks(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{
		Command: "test",
		OnDone: []config.Task{
			{
				Command: "test",
			},
		},
		OnFail: []config.Task{
			{
				Command: "test",
			},
		},
	}
	exe := task.Build(ctx, zap.NewNop(), taskConfig)
	assert.NotEqual(t, exe, nil)
}
