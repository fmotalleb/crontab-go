package config_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/fmotalleb/crontab-go/config"
)

var failJob config.JobConfig = config.JobConfig{
	Disabled: false,
	Tasks: []config.Task{
		{},
	},
}

var okJob config.JobConfig = config.JobConfig{
	Disabled: false,
	Tasks: []config.Task{
		{
			Post: "https://localhost",
		},
	},
}

func TestConfig_Validate_JobFails(t *testing.T) {
	cfg := &config.Config{
		Jobs: []*config.JobConfig{&failJob},
	}
	err := cfg.Validate()
	assert.Error(t, err)
}

func TestConfig_Validate_AllValidationsPass(t *testing.T) {
	cfg := &config.Config{
		Jobs: []*config.JobConfig{
			&okJob,
		},
	}
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_NoJobs(t *testing.T) {
	cfg := &config.Config{
		Jobs: []*config.JobConfig{},
	}
	err := cfg.Validate()
	assert.NoError(t, err)
}
