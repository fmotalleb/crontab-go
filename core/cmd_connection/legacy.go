package connection

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/fmotalleb/crontab-go/config"
	credential "github.com/fmotalleb/crontab-go/core/os_credential"
)

func prepareCrontabCommand(ctx context.Context, task *config.Task) (*exec.Cmd, error) {
	account, err := crontabAccount(task.UserName)
	if err != nil {
		return nil, fmt.Errorf("resolve crontab account: %w", err)
	}
	shell, home, env := crontabEnvironment(task.Crontab.Environment, account)
	// Cronie uses the shell as a path, relative to the job's directory if needed.
	if shell == "" {
		return nil, errors.New("crontab SHELL is empty")
	}
	shellPath := shell
	if !strings.ContainsRune(shellPath, '/') {
		shellPath = "./" + shellPath
	}
	cmd := exec.CommandContext(ctx, shellPath, "-c", task.Command)
	cmd.Args[0] = shell
	cmd.Env = env
	cmd.Dir = home
	if task.WorkingDirectory != "" {
		cmd.Dir = task.WorkingDirectory
	}
	if cmd.Dir == "" {
		return nil, errors.New("crontab HOME is empty")
	}
	cmd.Stdin = strings.NewReader(task.Crontab.Stdin)
	cmd.WaitDelay = time.Second
	if err := credential.SetCrontabUser(cmd, account, task.GroupName); err != nil {
		return nil, err
	}
	return cmd, nil
}

func crontabAccount(username string) (*user.User, error) {
	if username != "" {
		return user.Lookup(username)
	}
	return user.Current()
}

func crontabEnvironment(assignments []string, account *user.User) (shell, home string, env []string) {
	shell, home = "/bin/sh", account.HomeDir
	env = []string{"SHELL=" + shell, "HOME=" + home, "PATH=/usr/bin:/bin"}
	for _, assignment := range assignments {
		key, value, _ := strings.Cut(assignment, "=")
		switch key {
		case "SHELL":
			shell = value
		case "HOME":
			home = value
		}
		env = append(env, assignment)
	}
	// os/exec retains the last value for each name. Cronie sets the account
	// identifiers after reading the table.
	env = append(env, "LOGNAME="+account.Username, "USER="+account.Username)
	return shell, home, env
}
