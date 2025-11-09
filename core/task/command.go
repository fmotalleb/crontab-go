// Package task provides implementation of the abstraction.Executable interface for command tasks.
package task

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	connection "github.com/fmotalleb/crontab-go/core/cmd_connection"
	"github.com/fmotalleb/crontab-go/core/common"
	"github.com/fmotalleb/crontab-go/helpers"
)

func init() {
	tg.Register(NewCommand)
}

func NewCommand(
	logger *zap.Logger,
	task *config.Task,
) (abstraction.Executable, bool) {
	if task.Command == "" {
		return nil, false
	}
	cmd := &Command{
		log: logger.With(
			zap.String("command", task.Command),
		),

		task: task,
	}
	cmd.SetMaxRetry(task.Retries)
	cmd.SetRetryDelay(task.RetryDelay)
	cmd.SetTimeout(task.Timeout)
	cmd.SetMetaName("cmd: " + task.Command)
	return cmd, true
}

type Command struct {
	common.Hooked
	common.Cancelable
	common.Retry
	common.Timeout

	task *config.Task
	log  *zap.Logger
}

// Execute implements abstraction.Executable.
func (c *Command) Execute(ctx context.Context) (e error) {
	ctx = populateVars(ctx, c.task)
	r := common.GetRetry(ctx)
	log := c.log.With(
		zap.Any("retry", r),
		zap.Time("start", time.Now()),
	)
	defer func() {
		err := recover()
		if err != nil {
			if err, ok := err.(error); ok {
				log.Warn("recovering command execution from a fatal error", zap.Error(err))
				return
			}
			log.Warn("a non-error panic accord", zap.Any("error", err))
		}
	}()

	if err := c.WaitForRetry(ctx); err != nil {
		c.DoFailHooks(ctx)
		return err
	}

	ctx = common.IncreaseRetry(ctx)
	connections := c.task.Connections
	if fc := getFailedConnections(ctx); len(fc) != 0 {
		connections = fc
	}
	if len(connections) == 0 {
		connections = []config.TaskConnection{
			{
				Local: true,
			},
		}
		log.Debug("no explicit Connection provided using local task connection by default")
	}
	for _, conn := range connections {
		l := log.With(
			zap.Any("is-local", conn.Local),
		)
		connection := connection.Get(&conn, l)
		cmdCtx, cancel := c.ApplyTimeout(ctx)
		c.SetCancel(cancel)

		if err := connection.Prepare(cmdCtx, c.task); err != nil {
			l.Warn("cannot prepare command", zap.Error(err))
			ctx = addFailedConnections(ctx, conn)
			helpers.WarnOnErrIgnored(
				l,
				connection.Disconnect,
				"Cannot disconnect the command's connection",
			)
			continue
		}

		if err := connection.Connect(); err != nil {
			l.Warn("error when tried to connect, exiting current remote", zap.Error(err))
			ctx = addFailedConnections(ctx, conn)
			continue
		}
		ans, err := connection.Execute()
		if err != nil {
			ctx = addFailedConnections(ctx, conn)
		}
		l.Info("command finished", zap.ByteString("result", ans), zap.Error(err))
		if err := connection.Disconnect(); err != nil {
			l.Warn("error when tried to disconnect", zap.Error(err))
			ctx = addFailedConnections(ctx, conn)
			continue
		}
	}
	if fc := getFailedConnections(ctx); len(fc) != 0 {
		return c.Execute(ctx)
	}

	if errs := c.DoDoneHooks(ctx); len(errs) != 0 {
		log.Warn("command finished successfully but its hooks failed")
	}
	return nil
}
