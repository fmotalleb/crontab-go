//go:build windows

// Package credential is ignored on windows builds
package credential

import (
	"errors"
	"os/exec"
	"os/user"

	"go.uber.org/zap"
)

// Validate NOOP if user and group are empty log a warning if not and always returns nil.
func Validate(log *zap.Logger, usr string, grp string) error {
	if usr == "" && grp == "" {
		return nil // skip warn message if no user and group provided
	}
	log.Warn("windows os does not have capability to set user thus validation will pass but will not work")
	return nil
}

// SetUser NOOP.
func SetUser(_ *zap.Logger, _ *exec.Cmd, _ string, _ string) {
}

// SetCrontabUser reports that legacy crontab execution requires Unix credentials.
func SetCrontabUser(_ *exec.Cmd, _ *user.User, _ string) error {
	return errors.New("legacy crontab execution requires Unix")
}
