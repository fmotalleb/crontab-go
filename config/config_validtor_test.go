package config_test

// import (
// 	"testing"

// 	"github.com/alecthomas/assert/v2"
// 	"github.com/sirupsen/logrus"

// 	"github.com/FMotalleb/crontab-go/config"
// 	"github.com/FMotalleb/crontab-go/enums"
// 	mocklogger "github.com/FMotalleb/crontab-go/logger/mock_logger"
// )

// var failJob config.JobConfig = config.JobConfig{
// 	Disabled: false,
// 	Tasks: []config.Task{
// 		{},
// 	},
// }

// var okJob config.JobConfig = config.JobConfig{
// 	Disabled: false,
// 	Tasks: []config.Task{
// 		{
// 			Post: "https://localhost",
// 		},
// 	},
// }

// func TestConfig_Validate_LogFormatFails(t *testing.T) {
// 	cfg := &config.Config{
// 		LogFormat: enums.LoggerFormatType("unknown"),
// 		LogLevel:  enums.DebugLevel,
// 		Jobs:      []*config.JobConfig{},
// 	}
// 	log, _ := mocklogger.HijackOutput(logrus.New())
// 	err := cfg.Validate(log.WithField("test", "test"))
// 	assert.Error(t, err)
// 	assert.Equal(t, "given logger format: `unknown`", err.Error())
// }

// func TestConfig_Validate_LogLevelFails(t *testing.T) {
// 	cfg := &config.Config{
// 		LogFormat: enums.AnsiLogger,
// 		LogLevel:  enums.LogLevel("unknown"),
// 		Jobs:      []*config.JobConfig{},
// 	}
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	err := cfg.Validate(log)
// 	assert.Error(t, err)
// 	assert.Equal(t, "LogLevel must be one of (trace,debug,info,warn,fatal,panic) but received `unknown`", err.Error())
// }

// func TestConfig_Validate_JobFails(t *testing.T) {
// 	cfg := &config.Config{
// 		LogFormat: enums.JSONLogger,
// 		LogLevel:  enums.FatalLevel,
// 		Jobs:      []*config.JobConfig{&failJob},
// 	}
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	err := cfg.Validate(log)
// 	assert.Error(t, err)
// }

// func TestConfig_Validate_AllValidationsPass(t *testing.T) {
// 	cfg := &config.Config{
// 		LogFormat: enums.JSONLogger,
// 		LogLevel:  enums.DebugLevel,
// 		Jobs: []*config.JobConfig{
// 			&okJob,
// 		},
// 	}
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	err := cfg.Validate(log)
// 	assert.NoError(t, err)
// }

// func TestConfig_Validate_NoJobs(t *testing.T) {
// 	cfg := &config.Config{
// 		LogFormat: enums.JSONLogger,
// 		LogLevel:  enums.DebugLevel,
// 		Jobs:      []*config.JobConfig{},
// 	}
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	err := cfg.Validate(log)
// 	assert.NoError(t, err)
// }
