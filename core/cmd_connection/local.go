package connection

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.uber.org/zap"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/cmd_connection/command"
	credential "github.com/FMotalleb/crontab-go/core/os_credential"
)

func init() {
	cg.Register(NewLocalCMDConn)
}

// Local represents a local command connection.
type Local struct {
	log *zap.Logger
	cmd *exec.Cmd
}

// NewLocalCMDConn creates a new instance of Local command connection.
func NewLocalCMDConn(log *zap.Logger, cfg *config.TaskConnection) (abstraction.CmdConnection, bool) {
	if !cfg.Local {
		return nil, false
	}
	res := &Local{
		log: log.With(
			zap.String("connection", "local"),
		),
	}
	return res, true
}

// Prepare prepares the command for execution.
// It sets up the command with the provided context, task, and environment.
// It returns an error if the preparation fails.
func (l *Local) Prepare(ctx context.Context, task *config.Task) error {
	cmdCtx := command.NewCtx(ctx, task.Env, l.log)
	workingDir := task.WorkingDirectory
	if workingDir == "" {
		var e error
		workingDir, e = os.Getwd()
		if e != nil {
			return fmt.Errorf("cannot get current working directory: %w", e)
		}
	}

	shell, commandArg, environ := cmdCtx.BuildExecuteParams(task.Command)
	l.cmd = exec.CommandContext(
		ctx,
		shell,
		commandArg...,
	)
	l.log = l.log.With(
		zap.String("cmd", task.Command),
		zap.String("working_directory", workingDir),
		zap.String("shell", shell),
		zap.Strings("shell_args", commandArg),
	)
	credential.SetUser(l.log, l.cmd, task.UserName, task.GroupName)
	l.cmd.Env = environ
	l.cmd.Dir = workingDir

	// Add additional logging fields if needed
	l.log.Debug("command prepared")

	return nil
}

// Connect establishes the command connection.
// It returns an error if the connection cannot be established.
func (l *Local) Connect() error {
	return nil
}

// Disconnect closes the command connection.
// It returns an error if the disconnection process fails.
func (l *Local) Disconnect() error {
	return nil
}

// Execute executes the command and returns the output.
// It captures the command's standard output and standard error.
// It returns the output and an error, if any.
func (l *Local) Execute() ([]byte, error) {
	var res bytes.Buffer
	l.cmd.Stdout = &res
	l.cmd.Stderr = &res
	log := l.log.Named("execute")
	if err := l.cmd.Start(); err != nil {
		log.Warn("failed to start the command", zap.Error(err))
		return []byte{}, err
	} else if err := l.cmd.Wait(); err != nil {
		output := res.Bytes()
		log.Warn("command execution failed", zap.String("output", strings.TrimSpace(res.String())), zap.Error(err))
		return output, err
	}
	l.log.Debug("command output", zap.String("output", strings.TrimSpace(res.String())))
	return res.Bytes(), nil
}
