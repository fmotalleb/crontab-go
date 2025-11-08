package config_test

// import (
// 	"testing"

// 	"github.com/alecthomas/assert/v2"
// 	"github.com/sirupsen/logrus"

// 	"github.com/FMotalleb/crontab-go/config"
// 	mocklogger "github.com/FMotalleb/crontab-go/logger/mock_logger"
// )

// func TestJobConfig_Validate_Disabled(t *testing.T) {
// 	logger, buff := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	jobConfig := &config.JobConfig{
// 		Disabled:    true,
// 		Name:        "Test",
// 		Concurrency: 35,
// 		Tasks:       []config.Task{{}},
// 		Events: []config.JobEvent{
// 			{Interval: -1}, // Invalid interval
// 		},
// 	}

// 	err := jobConfig.Validate(log)
// 	assert.NoError(t, err, "Expected no error when job is disabled")
// 	assert.Contains(t, buff.String(), "JobConfig Test is disabled")
// }

// func TestJobConfig_Validate_Events(t *testing.T) {
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	jobConfig := &config.JobConfig{
// 		Events: []config.JobEvent{
// 			{Interval: -1}, // Invalid interval
// 		},
// 	}

// 	err := jobConfig.Validate(log)
// 	assert.Error(t, err, "Expected error due to invalid event interval")
// }

// func TestJobConfig_Validate_Tasks(t *testing.T) {
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	jobConfig := &config.JobConfig{
// 		Tasks: []config.Task{
// 			{Command: "echo", Get: "http://example.com"}, // Invalid task with both command and get
// 		},
// 	}

// 	err := jobConfig.Validate(log)
// 	assert.Error(t, err, "Expected error due to invalid task configuration")
// }

// func TestJobConfig_Validate_HooksDone(t *testing.T) {
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	jobConfig := &config.JobConfig{
// 		Hooks: config.JobHooks{
// 			Done: []config.Task{
// 				{Command: "echo", Get: "http://example.com"}, // Invalid task with both command and get
// 			},
// 		},
// 	}

// 	err := jobConfig.Validate(log)
// 	assert.Error(t, err, "Expected error due to invalid done hook task configuration")
// }

// func TestJobConfig_Validate_HooksFailed(t *testing.T) {
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	jobConfig := &config.JobConfig{
// 		Hooks: config.JobHooks{
// 			Failed: []config.Task{
// 				{Command: "echo", Get: "http://example.com"}, // Invalid task with both command and get
// 			},
// 		},
// 	}

// 	err := jobConfig.Validate(log)
// 	assert.Error(t, err, "Expected error due to invalid failed hook task configuration")
// }
