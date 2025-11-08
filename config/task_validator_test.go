package config_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.uber.org/zap"

	"github.com/FMotalleb/crontab-go/config"
)

func TestTaskValidate_NegativeTimeout(t *testing.T) {
	task := &config.Task{
		Command: "command",
		Timeout: -1,
	}

	err := task.Validate(zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout for tasks cannot be negative")
}

func TestTaskValidate_NegativeRetryDelay(t *testing.T) {
	task := &config.Task{
		Command:    "command",
		RetryDelay: -1,
	}

	err := task.Validate(zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "retry delay for tasks cannot be negative")
}

func TestTaskValidate_NegativeTimeoutAndRetryDelay(t *testing.T) {
	task := &config.Task{
		Command: "command",
		Timeout: -1,
		// RetryDelay: -1,
	}

	err := task.Validate(zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout for tasks cannot be negative")
	// assert.Contains(t, err.Error(), "retry delay for jobs cannot be negative")
}

func TestTaskValidate_ValidTimeoutAndRetryDelay(t *testing.T) {
	task := &config.Task{
		Command:    "command",
		Timeout:    10,
		RetryDelay: 5,
	}

	err := task.Validate(zap.NewNop())
	assert.NoError(t, err)
}

func TestTaskValidate_ValidTask(t *testing.T) {
	task := &config.Task{
		Command:    "echo 'Hello, World!'",
		Timeout:    10,
		RetryDelay: 5,
	}

	err := task.Validate(zap.NewNop())
	assert.NoError(t, err)
}

func TestTaskValidate_InvalidPostData(t *testing.T) {
	task := &config.Task{
		Post:       "http://localhost",
		Timeout:    10,
		Data:       map[any]any{},
		RetryDelay: 5,
	}

	err := task.Validate(zap.NewNop())
	assert.Error(t, err)
}

func TestTaskValidate_PostData(t *testing.T) {
	task := &config.Task{
		Post:       "http://localhost",
		Timeout:    10,
		Data:       map[string]any{},
		RetryDelay: 5,
	}

	err := task.Validate(zap.NewNop())
	assert.NoError(t, err)
}

func TestTaskValidate_CredentialLog(t *testing.T) {
	task := &config.Task{
		Command:    "test",
		Timeout:    10,
		RetryDelay: 5,
		UserName:   "testuser",
	}

	err := task.Validate(zap.NewNop())
	assert.NoError(t, err)
	// if runtime.GOOS == "windows" {
	// 	assert.Contains(t, buffer.String(), "windows os does not have capability")
	// } else {
	// 	assert.Contains(t, buffer.String(), "Be careful when using credentials, in local mode you can't use credentials unless running as root")
	// }
}

func TestTaskValidate_InvalidTaskWithData(t *testing.T) {
	task := &config.Task{
		Command:    "test",
		Data:       map[string]any{},
		Timeout:    10,
		RetryDelay: 5,
	}

	err := task.Validate(zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command cannot have data or headers field, violating command")
}

func TestTaskValidate_InvalidTaskWithHeader(t *testing.T) {
	task := &config.Task{
		Command:    "test",
		Headers:    map[string]string{},
		Timeout:    10,
		RetryDelay: 5,
	}

	err := task.Validate(zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command cannot have data or headers field, violating command")
}

func TestTaskValidate_InvalidGetWithData(t *testing.T) {
	task := &config.Task{
		Get:        "http://test",
		Data:       map[string]string{},
		Timeout:    10,
		RetryDelay: 5,
	}

	err := task.Validate(zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GET request cannot have data field, violating GET URI")
}

func TestTaskValidate_ValidCommandWithErrorHook(t *testing.T) {
	task := &config.Task{
		Command:    "test",
		Timeout:    10,
		RetryDelay: 5,
		OnDone: []config.Task{
			{},
		},
	}

	err := task.Validate(zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "a single task should have one of")
}
