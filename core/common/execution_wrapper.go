package common

import (
	"context"

	"github.com/fmotalleb/go-tools/log"
	"github.com/sethvargo/go-retry"
	"go.uber.org/zap"
)

type Action interface {
	Do(ctx context.Context) (e error)
}

type Executable struct {
	Retry
	Hooked
	Action
}

func (rh *Executable) forceRetry(ctx context.Context) error {
	err := rh.Do(ctx)
	if err != nil {
		return retry.RetryableError(err)
	}
	return nil
}

// Execute implements abstraction.Executable.
func (rh *Executable) Execute(ctx context.Context) error {
	err := rh.ExecuteRetry(ctx, rh.forceRetry)
	if err == nil {
		errs := rh.DoDoneHooks(ctx)
		if len(errs) != 0 {
			log.Of(ctx).Warn("some of on-done hooks failed", zap.Errors("errors", errs))
		}
	} else {
		errs := rh.DoFailHooks(ctx)
		if len(errs) != 0 {
			log.Of(ctx).Warn("some of on-fail hooks failed", zap.Errors("errors", errs))
		}
	}
	return err
}
