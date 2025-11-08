package task

import (
	"context"

	"go.uber.org/zap"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/generator"
)

var tg = generator.New[*config.Task, abstraction.Executable]()

func Build(ctx context.Context, log *zap.Logger, cfg config.Task) abstraction.Executable {
	exe, ok := tg.Get(log, &cfg)
	if !ok {
		log.Panic("did not received any executable action from given task", zap.Any("config", cfg))
	}
	onDone := []abstraction.Executable{}
	for _, d := range cfg.OnDone {
		onDone = append(onDone, Build(ctx, log, d))
	}
	exe.SetDoneHooks(ctx, onDone)
	onFail := []abstraction.Executable{}
	for _, d := range cfg.OnFail {
		onFail = append(onFail, Build(ctx, log, d))
	}
	exe.SetFailHooks(ctx, onFail)
	return exe
}
