package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	credential "github.com/fmotalleb/crontab-go/core/os_credential"
)

// Validate checks the validity of a Task.
// It ensures that the task has exactly one of the Get, Post, or Command fields, and validates other fields based on the specified action.
// If any validation fails, it returns an error with the specific validation error.
// Otherwise, it returns nil.
func (t *Task) Validate(log *zap.Logger) error {
	// Log the start of validation
	log = log.With(zap.Any("task", t))
	log.Debug("begin validation")
	checkList := []func(*Task, *zap.Logger) error{
		validateActionsList,
		validateCrontabCommand,
		validateCredential,
		validateFields,
		validateGetRequest,
		validateTimeout,
		validatePostData,
		validateRetry,
	}
	for _, check := range checkList {
		if err := check(t, log); err != nil {
			return err
		}
	}

	// Validate hooks
	for _, task := range append(t.OnDone, t.OnFail...) {
		if err := task.Validate(log); err != nil {
			log.Warn("hook: failed to validate", zap.Error(err))
			return err
		}
	}

	// Log the successful validation
	log.Debug("Validation successful for Task")
	return nil
}

// validateRetry returns an error if any retry delay, timeout, maximum delay,
// or jitter is negative. Zero values are allowed.
func validateRetry(t *Task, log *zap.Logger) error {
	for name, duration := range map[string]time.Duration{
		"retry delay":     t.RetryDelay,
		"retry timeout":   t.RetryTimeout,
		"retry max delay": t.RetryMaxDelay,
		"retry jitter":    t.RetryJitter,
	} {
		if duration < 0 {
			err := fmt.Errorf(
				"%s for tasks cannot be negative; received %s",
				name,
				duration,
			)
			log.Warn("Validation failed for Task", zap.Error(err))
			return err
		}
	}
	return nil
}

func validatePostData(t *Task, log *zap.Logger) error {
	if t.Data != nil {
		_, err := json.Marshal(t.Data)
		if err != nil {
			log.Warn("Validation failed for Task", zap.Error(err))
			return err
		}
	}
	return nil
}

func validateTimeout(t *Task, log *zap.Logger) error {
	if t.Timeout < 0 {
		err := fmt.Errorf(
			"timeout for tasks cannot be negative received `%d` for %+v",
			t.Timeout,
			t,
		)
		log.Warn("Validation failed for Task", zap.Error(err))
		return err
	}
	return nil
}

func validateGetRequest(t *Task, log *zap.Logger) error {
	if t.Get != "" && t.Data != nil {
		err := fmt.Errorf("GET request cannot have data field, violating GET URI: `%s`", t.Get)
		log.Warn("Validation failed for Task", zap.Error(err))
		return err
	}
	return nil
}

func validateFields(t *Task, log *zap.Logger) error {
	if t.Command != "" && (t.Data != nil || t.Headers != nil) {
		err := fmt.Errorf("command cannot have data or headers field, violating command: `%s`", t.Command)
		log.Warn("Validation failed for Task", zap.Error(err))
		return err
	}
	return nil
}

func validateCredential(t *Task, log *zap.Logger) error {
	if t.Crontab != nil {
		return nil // Checked by validateCrontabCommand.
	}
	// Only validate host credentials for tasks that execute locally
	if len(t.Connections) > 0 {
		hasLocal := false
		for _, c := range t.Connections {
			if c.Local {
				hasLocal = true
				break
			}
		}
		if !hasLocal {
			return nil
		}
	}
	if err := credential.Validate(log, t.UserName, t.GroupName); err != nil {
		log.Warn("Be careful when using credentials, in local mode you can't use credentials unless running as root", zap.Error(err))
	}
	return nil
}

func validateActionsList(t *Task, log *zap.Logger) error {
	actions := []bool{
		t.Get != "",
		t.Command != "",
		t.Post != "",
	}
	activeActions := 0
	for _, t := range actions {
		if t {
			activeActions++
		}
	}
	if activeActions != 1 {
		err := fmt.Errorf(
			"a single task should have one of (Get, Post, Command) fields, received:(Command: `%s`, Get: `%s`, Post: `%s`)",
			t.Command,
			t.Get,
			t.Post,
		)
		log.Warn("Validation failed for Task", zap.Error(err))
		return err
	}
	return nil
}

func validateCrontabCommand(t *Task, _ *zap.Logger) error {
	if t.Crontab == nil {
		return nil
	}
	if t.Command == "" {
		return errors.New("crontab command settings require a command")
	}
	if os.Geteuid() < 0 {
		return errors.New("legacy crontab execution requires Unix")
	}
	for _, connection := range t.Connections {
		if !connection.Local || connection.ImageName != "" || connection.ContainerName != "" || connection.ContainerLabel != "" {
			return errors.New("crontab command settings require a local connection")
		}
	}
	for _, assignment := range t.Crontab.Environment {
		name, _, ok := strings.Cut(assignment, "=")
		if !ok || name == "" || strings.ContainsRune(assignment, 0) {
			return errors.New("crontab environment entries must be NAME=value")
		}
	}
	if t.Crontab.Quiet && os.Geteuid() != 0 {
		return errors.New("a leading '-' in a crontab entry requires root")
	}
	return validateCrontabAccount(t)
}

func validateCrontabAccount(t *Task) error {
	if t.UserName != "" {
		account, err := user.Lookup(t.UserName)
		if err != nil {
			return fmt.Errorf("resolve crontab user %q: %w", t.UserName, err)
		}
		if uid := os.Geteuid(); uid != 0 && account.Uid != strconv.Itoa(uid) {
			return fmt.Errorf("switching to crontab user %q requires root", t.UserName)
		}
	}
	if t.GroupName != "" {
		if os.Geteuid() != 0 {
			return errors.New("setting a crontab group requires root")
		}
		if _, err := user.LookupGroup(t.GroupName); err != nil {
			return fmt.Errorf("resolve crontab group %q: %w", t.GroupName, err)
		}
	}
	return nil
}
