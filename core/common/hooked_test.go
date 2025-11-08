package common

import (
	"context"
	"errors"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

type mockExecutable struct {
	Cancelable
	Timeout
	Hooked

	err error
}

func newTask(typeName string) *mockExecutable {
	m := new(mockExecutable)
	m.Hooked.SetMetaName(typeName)
	return m
}

func (m *mockExecutable) Execute(_ context.Context) error {
	return m.err
}

func TestSetDoneHooks(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	h := &Hooked{}
	hooks := []abstraction.Executable{}
	h.SetDoneHooks(ctx, hooks)
	assert.Equal(t, hooks, h.doneHooks)
}

func TestSetFailHooks(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	h := &Hooked{}
	failHooks := []abstraction.Executable{newTask("test_fail_list")}
	h.SetFailHooks(ctx, failHooks)
	assert.Equal(t, failHooks, h.failHooks)
}

func TestDoDoneHooks_NoErrors(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	tsk := newTask("parent_done_ok")
	tsk.SetDoneHooks(ctx, []abstraction.Executable{newTask("test_done_ok")})

	errs := tsk.DoDoneHooks(ctx)
	assert.Zero(t, errs)
}

func TestDoFailHooks_WithErrors(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	tsk := newTask("")
	errHook := newTask("task_fail")
	errHook.err = errors.New("fail")
	tsk.SetFailHooks(ctx, []abstraction.Executable{errHook})

	errs := tsk.DoFailHooks(ctx)
	assert.Equal(t, len(errs), 1)
	assert.EqualError(t, errs[0], "fail")
}

func TestExecuteTasks_MultipleErrors(t *testing.T) {
	tsk1 := newTask("doFailList")
	tsk1.err = errors.New("error1")

	tsk2 := newTask("doFailList")
	tsk2.err = errors.New("error2")
	tasks := []abstraction.Executable{
		tsk1,
		tsk2,
	}
	errs := executeTasks(t.Context(), tasks)
	assert.Equal(t, len(errs), 2)
	assert.EqualError(t, errs[0], "error1")
	assert.EqualError(t, errs[1], "error2")
}
