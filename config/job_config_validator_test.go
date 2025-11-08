package config_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.uber.org/zap"

	"github.com/FMotalleb/crontab-go/config"
)

func TestJobConfig_Validate_Disabled(t *testing.T) {
	jobConfig := &config.JobConfig{
		Disabled:    true,
		Name:        "Test",
		Concurrency: 35,
		Tasks:       []config.Task{{}},
		Events: []config.JobEvent{
			{Interval: -1}, // Invalid interval
		},
	}

	err := jobConfig.Validate(zap.NewNop())
	assert.NoError(t, err, "Expected no error when job is disabled")
}

func TestJobConfig_Validate_Events(t *testing.T) {
	jobConfig := &config.JobConfig{
		Events: []config.JobEvent{
			{Interval: -1}, // Invalid interval
		},
	}

	err := jobConfig.Validate(zap.NewNop())
	assert.Error(t, err, "Expected error due to invalid event interval")
}

func TestJobConfig_Validate_Tasks(t *testing.T) {
	jobConfig := &config.JobConfig{
		Tasks: []config.Task{
			{Command: "echo", Get: "http://example.com"}, // Invalid task with both command and get
		},
	}

	err := jobConfig.Validate(zap.NewNop())
	assert.Error(t, err, "Expected error due to invalid task configuration")
}

func TestJobConfig_Validate_HooksDone(t *testing.T) {
	jobConfig := &config.JobConfig{
		Hooks: config.JobHooks{
			Done: []config.Task{
				{Command: "echo", Get: "http://example.com"}, // Invalid task with both command and get
			},
		},
	}

	err := jobConfig.Validate(zap.NewNop())
	assert.Error(t, err, "Expected error due to invalid done hook task configuration")
}

func TestJobConfig_Validate_HooksFailed(t *testing.T) {
	jobConfig := &config.JobConfig{
		Hooks: config.JobHooks{
			Failed: []config.Task{
				{Command: "echo", Get: "http://example.com"}, // Invalid task with both command and get
			},
		},
	}

	err := jobConfig.Validate(zap.NewNop())
	assert.Error(t, err, "Expected error due to invalid failed hook task configuration")
}
