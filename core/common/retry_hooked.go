package common

import (
	"context"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
)

type RetryHooked struct {
	Retry
	Hooked
}

// Execute implements abstraction.Executable.
func (rh *RetryHooked) Execute(ctx context.Context) error {
	err := rh.ExecuteRetry(ctx)
	if err != nil {
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
