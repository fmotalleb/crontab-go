// Package task provides implementation of the abstraction.Executable interface for command tasks.
package task

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
			zap.String("command.hash", shortHash(task.Command)),
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

// Do implements common.Action.
func (c *Command) Do(ctx context.Context) (e error) {
	execID := newExecutionID()
	ctx, span := taskTracer.Start(
		ctx, "task.command",
		trace.WithAttributes(
			attribute.String("command.hash", shortHash(c.task.Command)),
			attribute.String("execution.id", execID),
		),
	)
	defer span.End()
	c.setCommandAttrs(span)
	defer func() {
		if e != nil {
			span.RecordError(e)
			span.SetStatus(codes.Error, e.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}()
	ctx = populateVars(ctx, c.task)
	log := c.log.With(
		zap.String("id", execID),
		zap.Time("start", time.Now()),
	)
	log.Info("command started")
	defer c.recoverPanic(span, &e)
	localCtx, cancel := c.ApplyTimeout(ctx)
	defer cancel()
	unregisterCancel := c.RegisterCancel(cancel)
	defer unregisterCancel()
	connections := c.task.Connections
	if len(connections) == 0 {
		connections = []config.TaskConnection{{Local: true}}
		log.Debug("no explicit Connection provided using local task connection by default")
	}
	for _, conn := range connections {
		if err := c.executeConnection(localCtx, span, conn, execID, log); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) setCommandAttrs(span trace.Span) {
	span.SetAttributes(
		attribute.Int64("task.timeout_nanos", int64(c.task.Timeout)),
		attribute.Int64("retry.max_retries", int64(c.task.Retries)),
	)
	if c.task.UserName != "" {
		span.SetAttributes(attribute.String("command.user", c.task.UserName))
	}
	if c.task.GroupName != "" {
		span.SetAttributes(attribute.String("command.group", c.task.GroupName))
	}
	if c.task.WorkingDirectory != "" {
		span.SetAttributes(attribute.String("command.working_directory", c.task.WorkingDirectory))
	}
	if len(c.task.Env) > 0 {
		keys := make([]string, 0, len(c.task.Env))
		for k := range c.task.Env {
			keys = append(keys, k)
		}
		span.SetAttributes(attribute.StringSlice("command.env.keys", keys))
	}
}

func (c *Command) recoverPanic(span trace.Span, e *error) {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			c.log.Error("panic recovered", zap.Error(err))
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			*e = err
		} else {
			c.log.Error("panic recovered", zap.Any("error", r))
			err := fmt.Errorf("panic: %v", r)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			*e = err
		}
	}
}

func (c *Command) executeConnection(
	ctx context.Context,
	span trace.Span,
	conn config.TaskConnection,
	execID string,
	log *zap.Logger,
) error {
	connType := connectionType(conn)
	span.SetAttributes(attribute.String("command.connection.type", connType))
	l := log.With(zap.String("connection", connType))
	if conn.DockerConnection != "" {
		span.SetAttributes(attribute.String("docker.endpoint", conn.DockerConnection))
	}
	if conn.ImageName != "" {
		span.SetAttributes(attribute.String("docker.image", conn.ImageName))
	}
	if conn.ContainerName != "" {
		span.SetAttributes(attribute.String("docker.container.name", conn.ContainerName))
	}
	if len(conn.Networks) > 0 {
		span.SetAttributes(attribute.StringSlice("docker.networks", conn.Networks))
	}
	cmdConn, connErr := connection.Get(&conn, l)
	if connErr != nil {
		return connErr
	}
	if err := c.runPhase(ctx, "command.prepare", func() error {
		return cmdConn.Prepare(ctx, c.task)
	}); err != nil {
		return err
	}
	if err := c.runPhase(ctx, "command.connect", func() error {
		return cmdConn.Connect()
	}); err != nil {
		return err
	}
	defer helpers.WarnOnErrIgnored(l, cmdConn.Disconnect, "error when tried to disconnect")
	var execErr error
	if err := c.runPhase(ctx, "command.execute", func() error {
		out := NewPrefixWriter(os.Stderr, execID)
		execErr = cmdConn.Execute(out, out)
		return execErr
	}); err != nil {
		if code, ok := extractExitCode(execErr); ok {
			span.SetAttributes(attribute.Int("command.exit_code", code))
		}
		return err
	}
	l.Info("command finished")
	return nil
}

func (c *Command) runPhase(
	ctx context.Context,
	phase string,
	fn func() error,
) error {
	_, child := taskTracer.Start(ctx, phase, trace.WithSpanKind(trace.SpanKindInternal))
	defer child.End()
	err := fn()
	if err != nil {
		child.RecordError(err)
		child.SetStatus(codes.Error, err.Error())
		return errors.Join(errors.New("failed to "+phase), err)
	}
	child.SetStatus(codes.Ok, "")
	return nil
}

func connectionType(conn config.TaskConnection) string {
	switch {
	case conn.ImageName != "":
		return "docker-create"
	case conn.ContainerName != "", conn.ContainerLabel != "":
		return "docker-attach"
	case conn.Local:
		return "local"
	default:
		return "unknown"
	}
}

func extractExitCode(err error) (int, bool) {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), true
	}
	return 0, false
}
