package common

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/core/global"
)

type Hooked struct {
	metaName  string
	doneHooks []abstraction.Executable
	failHooks []abstraction.Executable
}

func (h *Hooked) SetMetaName(metaName string) {
	h.metaName = metaName
	global.RegisterCounter(
		global.OKMetricName,
		global.OKMetricHelp,
		h.GetMeta(),
	)
	global.RegisterCounter(
		global.ErrMetricName,
		global.ErrMetricHelp,
		h.GetMeta(),
	)
}

func (h *Hooked) GetMeta() map[string]string {
	return prometheus.Labels{
		"task": h.metaName,
	}
}

func (h *Hooked) SetDoneHooks(_ context.Context, hooks []abstraction.Executable) {
	h.doneHooks = hooks
}

func (h *Hooked) SetFailHooks(_ context.Context, failHooks []abstraction.Executable) {
	h.failHooks = failHooks
}

func (h *Hooked) DoDoneHooks(ctx context.Context) []error {
	global.IncMetric(
		global.OKMetricName,
		global.OKMetricHelp,
		h.GetMeta(),
	)
	ctx = ResetRetries(ctx)
	return executeTasks(ctx, h.doneHooks)
}

func (h *Hooked) DoFailHooks(ctx context.Context) []error {
	global.IncMetric(
		global.ErrMetricName,
		global.ErrMetricHelp,
		h.GetMeta(),
	)
	ctx = ResetRetries(ctx)
	return executeTasks(ctx, h.failHooks)
}

func executeTasks(ctx context.Context, tasks []abstraction.Executable) []error {
	errs := []error{}
	for _, exe := range tasks {
		if err := exe.Execute(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
