package connection

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/cmd_connection/command"
	"github.com/fmotalleb/crontab-go/core/utils"
	"github.com/fmotalleb/crontab-go/helpers"
)

func init() {
	cg.RegisterWithPriority(NewDockerCreateConnection, 30)
}

// DockerCreateConnection is a struct that manages the creation and execution of Docker containers.
type DockerCreateConnection struct {
	conn            *config.TaskConnection
	log             *zap.Logger
	cli             *client.Client
	containerConfig *container.Config
	hostConfig      *container.HostConfig
	networkConfig   *network.NetworkingConfig
	ctx             context.Context
}

// NewDockerCreateConnection initializes a new DockerCreateConnection instance.
// Parameters:
// - log: A zap.Logger instance for logging.
// - conn: A TaskConnection instance containing the connection configuration.
// Returns:
// - A new instance of DockerCreateConnection.
func NewDockerCreateConnection(log *zap.Logger, conn *config.TaskConnection) (abstraction.CmdConnection, bool) {
	if conn.ImageName == "" {
		return nil, false
	}
	res := &DockerCreateConnection{
		conn: conn,
		log: log.With(
			zap.String("connection", "docker"),
			zap.String("docker-mode", "create"),
		),
	}
	return res, true
}

// Prepare sets up the Docker container configuration based on the provided task.
// Parameters:
// - ctx: A context.Context instance for managing the lifecycle of the container.
// - task: A Task instance containing the task configuration.
// Returns:
// - An error if the preparation fails, otherwise nil.
func (d *DockerCreateConnection) Prepare(ctx context.Context, task *config.Task) error {
	cmdCtx := command.NewCtx(ctx, task.Env, d.log)
	d.ctx = ctx
	if d.conn.DockerConnection == "" {
		d.log.Debug("No explicit docker connection specified, using default: `unix:///var/run/docker.sock`")
		d.conn.DockerConnection = "unix:///var/run/docker.sock"
	}

	shell, shellArgs, environments := cmdCtx.BuildExecuteParams(task.Command)
	cmd := append(
		[]string{shell},
		shellArgs...,
	)
	volumes := make(map[string]struct{})
	for _, volume := range d.conn.Volumes {
		parts := utils.EscapedSplit(volume, ':')
		if len(parts) < 2 {
			continue
		}
		inContainer := parts[1]
		volumes[inContainer] = struct{}{}
	}
	// Create an exec configuration
	d.containerConfig = &container.Config{
		AttachStdout: true,
		AttachStderr: true,
		Env:          environments,
		WorkingDir:   task.WorkingDirectory,
		User:         task.UserName,
		Cmd:          cmd,
		Image:        d.conn.ImageName,
		Volumes:      volumes,
		Entrypoint:   []string{},
		Shell:        []string{"/bin/sh", "-c"},
	}
	d.hostConfig = &container.HostConfig{
		Binds: d.conn.Volumes,
		// AutoRemove: true,
	}
	endpointsConfig := make(map[string]*network.EndpointSettings)
	for _, networkName := range d.conn.Networks {
		endpointsConfig[networkName] = &network.EndpointSettings{}
	}
	d.networkConfig = &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}
	return nil
}

// Connect establishes a connection to the Docker daemon.
// Returns:
// - An error if the connection fails, otherwise nil.
func (d *DockerCreateConnection) Connect() error {
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

// Execute creates a container for the prepared command, waits for it to finish,
// and copies its logs to stdout and stderr. Start failures are retried every
// second until the prepared context is canceled. It returns creation, wait,
// log retrieval, output copying, cancellation, or nonzero exit status errors.
// Container removal is attempted on return; removal and log close errors are ignored.
func (d *DockerCreateConnection) Execute(stdout, stderr io.Writer) error {
	ctx := d.ctx
	// Create the exec instance

	exec, err := d.cli.ContainerCreate(
		ctx,
		d.containerConfig,
		d.hostConfig,
		d.networkConfig,
		nil,
		d.conn.ContainerName,
	)
	if err != nil {
		return err
	}

	d.log.Debug("container created", zap.Any("response", exec), zap.Strings("warnings", exec.Warnings))

	defer helpers.WarnOnErrIgnored(
		d.log,
		func() error {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			defer cancel()
			return d.cli.ContainerRemove(cleanupCtx, exec.ID,
				container.RemoveOptions{
					Force: true,
				},
			)
		},
		"cannot remove the container",
	)

	for {
		err = d.cli.ContainerStart(
			ctx,
			exec.ID,
			container.StartOptions{},
		)
		if err == nil {
			break
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		d.log.Warn("container start failed, retrying", zap.Error(err))
		timer := time.NewTimer(time.Second)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}

	d.log.Debug("container started", zap.Any("container", exec))

	waitResult, waitErrors := d.cli.ContainerWait(ctx, exec.ID, container.WaitConditionNotRunning)
	var result container.WaitResponse
	select {
	case waitErr := <-waitErrors:
		if waitErr != nil {
			return fmt.Errorf("wait for container %s: %w", exec.ID, waitErr)
		}
		return fmt.Errorf("wait for container %s returned without a result", exec.ID)
	case result = <-waitResult:
	case <-ctx.Done():
		return ctx.Err()
	}
	if result.Error != nil {
		return fmt.Errorf("container %s failed: %s", exec.ID, result.Error.Message)
	}

	// Read the completed container's logs so fast commands cannot finish before
	// output collection begins.
	resp, err := d.cli.ContainerLogs(
		ctx,
		exec.ID,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     false,
			Details:    true,
		},
	)
	if err != nil {
		return err
	}
	defer helpers.WarnOnErrIgnored(
		d.log,
		func() error {
			return resp.Close()
		},
		"cannot close the container's logs",
	)

	// Demultiplex the container log frames into separate stdout/stderr streams
	wrote, err := stdcopy.StdCopy(stdout, stderr, resp)
	d.log.Debug("output of stdout is fetched", zap.Int64("bytes", wrote))
	if err != nil {
		d.log.Debug("copy of std is failed", zap.Int64("until-err", wrote), zap.Error(err))
		return err
	}
	return checkDockerExitStatus("container "+exec.ID, result.StatusCode)
}

// Disconnect closes the connection to the Docker daemon.
// Returns:
// - An error if the disconnection fails, otherwise nil.
func (d *DockerCreateConnection) Disconnect() error {
	return d.cli.Close()
}
