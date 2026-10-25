package connection

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/cmd_connection/command"
)

func init() {
	cg.RegisterWithPriority(NewDockerAttachConnection, 20)
}

type DockerAttachConnection struct {
	conn    *config.TaskConnection
	log     *zap.Logger
	cli     *client.Client
	execCFG *container.ExecOptions
	ctx     context.Context
}

// NewDockerAttachConnection creates a new DockerAttachConnection instance.
// It initializes the connection configuration and logging fields.
// Parameters:
// - log: A zap.Logger instance for logging purposes.
// - conn: A TaskConnection instance containing the connection configuration.
// Returns:
// - A new instance of DockerAttachConnection implementing the CmdConnection interface.
func NewDockerAttachConnection(log *zap.Logger, conn *config.TaskConnection) (abstraction.CmdConnection, bool) {
	if conn.ContainerName == "" && conn.ContainerLabel == "" {
		return nil, false
	}
	res := &DockerAttachConnection{
		conn: conn,
		log: log.With(
			zap.String("connection", "docker"),
			zap.String("docker-mode", "attach"),
		),
	}
	return res, true
}

// Prepare sets up the DockerAttachConnection for executing a task.
// It reshapes the environment variables, sets the context, and creates an exec configuration.
// Parameters:
// - ctx: A context.Context instance for managing the request lifetime.
// - task: A Task instance containing the task configuration.
// Returns:
// - An error if the preparation fails, otherwise nil.
func (d *DockerAttachConnection) Prepare(ctx context.Context, task *config.Task) error {
	cmdCtx := command.NewCtx(ctx, task.Env, d.log)
	d.ctx = ctx
	// Specify the container ID or name
	if d.conn.DockerConnection == "" {
		d.log.Debug("No explicit docker connection specified, using default: `unix:///var/run/docker.sock`")
		d.conn.DockerConnection = "unix:///var/run/docker.sock"
	}
	shell, shellArgs, environments := cmdCtx.BuildExecuteParams(task.Command)
	cmd := append(
		[]string{shell},
		shellArgs...,
	)
	// Create an exec configuration
	d.execCFG = &container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Privileged:   true,
		Env:          environments,
		WorkingDir:   task.WorkingDirectory,
		User:         task.UserName,
		Cmd:          cmd,
	}
	return nil
}

// Connect establishes a connection to the Docker daemon.
// It initializes the Docker client with the specified connection settings.
// Returns:
// - An error if the connection fails, otherwise nil.
func (d *DockerAttachConnection) Connect() error {
	cli, err := client.NewClientWithOpts(
		client.WithHost(d.conn.DockerConnection),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	d.cli = cli
	return nil
}

// Execute runs the command in the Docker container and streams its output to the provided writer.
// It creates an exec instance and attaches to it.
func (d *DockerAttachConnection) Execute(stdout, stderr io.Writer) error {
	cid := d.conn.ContainerName
	if cid == "" {
		label := d.conn.ContainerLabel
		if label == "" {
			return errors.New("neither container name nor label provided")
		}

		args := filters.NewArgs()
		args.Add("label", label)

		containers, err := d.cli.ContainerList(d.ctx, container.ListOptions{
			Filters: args,
		})
		if err != nil {
			return err
		}

		if len(containers) == 0 {
			return fmt.Errorf("no container found with label %q", label)
		}

		if len(containers) != 1 {
			return fmt.Errorf("more than one container found with label %q", label)
		}
		cid = containers[0].ID
	}

	// Create the exec instance
	exec, err := d.cli.ContainerExecCreate(d.ctx, cid, *d.execCFG)
	if err != nil {
		return err
	}

	// Attach to the exec instance
	resp, err := d.cli.ContainerExecAttach(
		d.ctx,
		exec.ID,
		container.ExecStartOptions{},
	)
	if err != nil {
		return err
	}
	defer func() {
		resp.Close()
	}()

	// Demultiplex the exec frames into separate stdout/stderr streams
	wrote, err := stdcopy.StdCopy(stdout, stderr, resp.Reader)
	d.log.Debug("output of stdout is fetched", zap.Int64("bytes", wrote))
	if err != nil {
		d.log.Debug("copy of std is failed", zap.Int64("until-err", wrote), zap.Error(err))
		return err
	}
	inspect, err := waitForExec(d.ctx, d.cli, exec.ID)
	if err != nil {
		return err
	}
	return checkDockerExitStatus("docker exec "+exec.ID, int64(inspect.ExitCode))
}

func waitForExec(ctx context.Context, cli *client.Client, execID string) (container.ExecInspect, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		inspect, err := cli.ContainerExecInspect(ctx, execID)
		if err != nil {
			return container.ExecInspect{}, fmt.Errorf("inspect docker exec %s: %w", execID, err)
		}
		if !inspect.Running {
			return inspect, nil
		}
		select {
		case <-ctx.Done():
			return container.ExecInspect{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

// Disconnect closes the connection to the Docker daemon.
// Returns:
// - An error if the disconnection fails, otherwise nil.
func (d *DockerAttachConnection) Disconnect() error {
	return d.cli.Close()
}
