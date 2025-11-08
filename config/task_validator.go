package config

import (
	"encoding/json"
	"fmt"

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

func validateRetry(t *Task, log *zap.Logger) error {
	if t.RetryDelay < 0 {
		err := fmt.Errorf(
			"retry delay for tasks cannot be negative received `%d` for %+v",
			t.RetryDelay,
			t,
		)
		log.Warn("Validation failed for Task", zap.Error(err))
		return err
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
