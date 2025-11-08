package connection

import (
	"bytes"
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/cmd_connection/command"
)

func init() {
	cg.Register(NewDockerAttachConnection)
}

type DockerAttachConnection struct {
	conn        *config.TaskConnection
	log         *zap.Logger
	cli         *client.Client
	execCFG     *container.ExecOptions
	containerID string
	ctx         context.Context
}

// NewDockerAttachConnection creates a new DockerAttachConnection instance.
// It initializes the connection configuration and logging fields.
// Parameters:
// - log: A zap.Logger instance for logging purposes.
// - conn: A TaskConnection instance containing the connection configuration.
// Returns:
// - A new instance of DockerAttachConnection implementing the CmdConnection interface.
func NewDockerAttachConnection(log *zap.Logger, conn *config.TaskConnection) (abstraction.CmdConnection, bool) {
	if conn.ContainerName == "" {
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
	d.containerID = d.conn.ContainerName
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
		Tty:          true,
		Privileged:   true,
		Env:          environments,
		WorkingDir:   task.WorkingDirectory,
		User:         task.UserName,
		Detach:       false,
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

// Execute runs the command in the Docker container and captures the output.
// It creates an exec instance, attaches to it, and reads the command output.
// Returns:
// - A byte slice containing the command output.
// - An error if the execution fails, otherwise nil.
func (d *DockerAttachConnection) Execute() ([]byte, error) {
	// Create the exec instance
	exec, err := d.cli.ContainerExecCreate(d.ctx, d.containerID, *d.execCFG)
	if err != nil {
		return nil, err
	}

	// Attach to the exec instance
	resp, err := d.cli.ContainerExecAttach(
		d.ctx,
		exec.ID,
		container.ExecStartOptions{
			Tty: true,
		},
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		resp.Close()
	}()

	writer := bytes.NewBuffer([]byte{})
	// Print the command output
	wrote, err := io.Copy(writer, resp.Reader)
	d.log.Debug("output of stdout is fetched", zap.Int64("bytes", wrote))
	if err != nil {
		d.log.Debug("copy of std is failed", zap.Int64("until-err", wrote), zap.Error(err))
		return writer.Bytes(), err
	}
	return writer.Bytes(), nil
}

// Disconnect closes the connection to the Docker daemon.
// Returns:
// - An error if the disconnection fails, otherwise nil.
func (d *DockerAttachConnection) Disconnect() error {
	return d.cli.Close()
}
