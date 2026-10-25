package config_test

import (
	"encoding/json"
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
	Name:     "valid",
	Disabled: false,
	Tasks: []config.Task{
		{
			Post: "https://localhost",
		},
	},
	Events: []config.JobEvent{{OnInit: true}},
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

func TestConfig_Validate_RejectsNilJob(t *testing.T) {
	cfg := &config.Config{Jobs: []*config.JobConfig{nil}}
	assert.Error(t, cfg.Validate())
}

func TestConfig_Validate_WebserverRequiresExplicitAuthenticationPolicy(t *testing.T) {
	cfg := &config.Config{WebServerAddress: "127.0.0.1", WebServerPort: 8080}
	assert.Error(t, cfg.Validate())

	cfg.AllowUnauthenticatedWebserver = true
	assert.NoError(t, cfg.Validate())
}

func TestConfig_Validate_WebserverRejectsPartialCredentials(t *testing.T) {
	cfg := &config.Config{
		WebServerAddress:  "127.0.0.1",
		WebServerPort:     8080,
		WebserverUsername: "scheduler",
	}
	assert.Error(t, cfg.Validate())
}

func TestConfig_JSONUsesSchemaKeys(t *testing.T) {
	cfg := config.Config{
		Jobs: []*config.JobConfig{{
			Name:   "docker",
			Events: []config.JobEvent{{Docker: &config.DockerEvent{ErrorLimit: 3}}},
			Tasks:  []config.Task{{Command: "true", WorkingDirectory: "/work"}},
		}},
		Variables: map[string]string{"region": "test"},
	}
	encoded, err := json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Contains(t, string(encoded), `"working-dir":"/work"`)
	assert.Contains(t, string(encoded), `"error-limit-count":3`)
	assert.Contains(t, string(encoded), `"variables":{"region":"test"}`)
}
