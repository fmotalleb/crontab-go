package config_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/config"
)

func TestJobEvent_Validate_WebEvent(t *testing.T) {
	event := config.JobEvent{
		Interval: 0,
		Cron:     "",
		OnInit:   false,
		WebEvent: "test-event",
	}
	err := event.Validate(zap.NewNop())

	assert.NoError(t, err)
}

func TestJobEvent_Validate_PositiveInterval(t *testing.T) {
	event := config.JobEvent{
		Interval: 10,
		Cron:     "",
		OnInit:   false,
		WebEvent: "",
	}
	err := event.Validate(zap.NewNop())

	assert.NoError(t, err)
}

func TestJobEvent_Validate_CorrectCron(t *testing.T) {
	event := config.JobEvent{
		Interval: 0,
		Cron:     "* * * * *",
		OnInit:   false,
	}

	err := event.Validate(zap.NewNop())
	assert.NoError(t, err)
}

func TestJobEvent_Validate_NegativeInterval(t *testing.T) {
	event := config.JobEvent{
		Interval: -10,
		Cron:     "",
		OnInit:   false,
	}

	err := event.Validate(zap.NewNop())

	expectedErr := "received a negative time in interval: `-10ns`"

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr)
}

func TestJobEvent_Validate_InvalidCronExpression(t *testing.T) {
	event := config.JobEvent{
		Interval: 0,
		Cron:     "invalid_cron_expression",
		OnInit:   false,
	}

	err := event.Validate(zap.NewNop())

	assert.Error(t, err)
}

func TestJobEvent_Validate_MultipleActiveSchedules(t *testing.T) {
	event := config.JobEvent{
		Interval: 60,
		Cron:     "0 0 * * *",
		OnInit:   true,
	}

	err := event.Validate(zap.NewNop())

	expectedErr := "a single event must have one of "

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr)
}

func TestJobEvent_Validate_DockerInvalidImagePattern(t *testing.T) {
	event := config.JobEvent{
		Docker: &config.DockerEvent{
			Name:   ".*",
			Image:  "[invalid",
			Labels: map[string]string{},
		},
	}

	err := event.Validate(zap.NewNop())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing closing ]")
}

func TestJobEvent_Validate_DockerInvalidLabelPattern(t *testing.T) {
	event := config.JobEvent{
		Docker: &config.DockerEvent{
			Name:  ".*",
			Image: ".*",
			Labels: map[string]string{
				"service": "(broken",
			},
		},
	}

	err := event.Validate(zap.NewNop())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing closing )")
}
