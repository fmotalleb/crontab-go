// Package task provides implementation of the abstraction.Executable interface for command tasks.
package task

import (
	"context"
	"errors"
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
	cmd.ConfigRetryFrom(task)
	cmd.SetTimeout(task.Timeout)
	cmd.SetMetaName("cmd: " + task.Command)
	cmd.Action = cmd
	return cmd, true
}

type Command struct {
	common.Executable
	common.Cancelable
	common.Timeout

	task *config.Task
	log  *zap.Logger
}

// Execute implements common.RetryHooked.
func (c Command) Do(ctx context.Context) (e error) {
	ctx = populateVars(ctx, c.task)
	log := c.log.With(
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
	connections := c.task.Connections
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
			l.Error("cannot prepare command", zap.Error(err))
			helpers.WarnOnErrIgnored(
				l,
				connection.Disconnect,
				"Cannot disconnect the command's connection",
			)
			return errors.Join(errors.New("failed to prepare"), err)
		}

		if err := connection.Connect(); err != nil {
			l.Error("error when tried to connect, exiting current remote", zap.Error(err))
			return errors.Join(errors.New("failed to connect"), err)
		}
		ans, err := connection.Execute()
		if err != nil {
			l.Error("failed to run command", zap.Error(err))
			return errors.Join(errors.New("failed to execute command"), err)
		}
		l.Info("command finished", zap.ByteString("result", ans), zap.Error(err))
		if err := connection.Disconnect(); err != nil {
			l.Warn("error when tried to disconnect", zap.Error(err))
			// return errors.Join(errors.New("failed to execute command"), err)
			continue
		}
	}

	return nil
}
