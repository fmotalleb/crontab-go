// Package config contains the configuration information of the application.
package config

import (
	"time"
)

// Config represents the configuration for the crontab application.
type Config struct {
	// Command executor configs
	Shell     string   `mapstructure:"shell" json:"shell,omitempty"`
	ShellArgs []string `mapstructure:"shell_args" json:"shell_args,omitempty"`

	// Web-server config
	WebServerAddress  string `mapstructure:"webserver_address" json:"webserver_listen_address,omitempty"`
	WebServerPort     uint   `mapstructure:"webserver_port" json:"webserver_port,omitempty"`
	WebserverUsername string `mapstructure:"webserver_username" json:"webserver_username,omitempty"`
	WebServerPassword string `mapstructure:"webserver_password" json:"webserver_password,omitempty"`
	WebServerMetrics  bool   `mapstructure:"webserver_metrics" json:"webserver_metrics,omitempty"`

	Jobs []*JobConfig `mapstructure:"jobs" json:"jobs"`
}

// JobConfig represents the configuration for a specific job.
type JobConfig struct {
	Name        string        `mapstructure:"name" json:"name,omitempty"`
	Description string        `mapstructure:"description" json:"description,omitempty"`
	Disabled    bool          `mapstructure:"disabled" json:"disabled,omitempty"`
	Concurrency uint          `mapstructure:"concurrency" json:"concurrency,omitempty"`
	Tasks       []Task        `mapstructure:"tasks" json:"tasks,omitempty"`
	Events      []JobEvent    `mapstructure:"events" json:"events"`
	Hooks       JobHooks      `mapstructure:"hooks" json:"hooks,omitempty"`
	Debounce    time.Duration `mapstructure:"debounce" json:"debounce,omitempty"`
}

// JobEvent represents the scheduling configuration for a job.
type JobEvent struct {
	Cron     string        `mapstructure:"cron" json:"cron,omitempty"`
	Interval time.Duration `mapstructure:"interval" json:"interval,omitempty"`
	OnInit   bool          `mapstructure:"on-init" json:"on-init,omitempty"`
	WebEvent string        `mapstructure:"web-event" json:"web-event,omitempty"`
	Docker   *DockerEvent  `mapstructure:"docker" json:"docker,omitempty"`

	LogFile        string        `mapstructure:"log-file" json:"log-file,omitempty"`
	LogCheckCycle  time.Duration `mapstructure:"log-check-cycle" json:"log-check-cycle,omitempty"`
	LogLineBreaker string        `mapstructure:"log-line-breaker" json:"log-line-breaker,omitempty"`
	LogMatcher     string        `mapstructure:"log-matcher" json:"log-matcher,omitempty"`
}

// DockerEvent represents a Docker event configuration.
type DockerEvent struct {
	Connection       string            `mapstructure:"connection" json:"connection,omitempty"`
	Name             string            `mapstructure:"name" json:"name,omitempty"`
	Image            string            `mapstructure:"image" json:"image,omitempty"`
	Actions          []string          `mapstructure:"actions" json:"actions,omitempty"`
	Labels           map[string]string `mapstructure:"labels" json:"labels,omitempty"`
	ErrorLimit       uint              `mapstructure:"error-limit-count" json:"error-limit,omitempty"`
	ErrorLimitPolicy ErrorLimitPolicy  `mapstructure:"error-limit-policy" json:"error-limit-policy,omitempty"`
	ErrorThrottle    time.Duration     `mapstructure:"error-throttle" json:"error-throttle,omitempty"`
}

// JobHooks represents the hooks configuration for a job.
type JobHooks struct {
	Done   []Task `mapstructure:"done" json:"done,omitempty"`
	Failed []Task `mapstructure:"failed" json:"failed,omitempty"`
}

// Task represents the configuration for a task within a job.
type Task struct {
	// Http Requests
	Post    string            `mapstructure:"post" json:"post,omitempty"`
	Get     string            `mapstructure:"get" json:"get,omitempty"`
	Headers map[string]string `mapstructure:"headers" json:"headers,omitempty"`
	Data    any               `mapstructure:"data" json:"data,omitempty"`

	// Command params
	Command          string            `mapstructure:"command" json:"command,omitempty"`
	WorkingDirectory string            `mapstructure:"working-dir" json:"working-directory,omitempty"`
	UserName         string            `mapstructure:"user" json:"user,omitempty"`
	GroupName        string            `mapstructure:"group" json:"group,omitempty"`
	Env              map[string]string `mapstructure:"env" json:"env,omitempty"`
	Connections      []TaskConnection  `mapstructure:"connections" json:"connections,omitempty"`

	// Retry & Timeout config
	Retries       uint64        `mapstructure:"retries" json:"retries,omitempty"`
	RetryDelay    time.Duration `mapstructure:"retry-delay" default:"15s" json:"retry-delay,omitempty"`
	RetryTimeout  time.Duration `mapstructure:"retry-timeout" json:"retry-timeout,omitempty"`
	RetryMaxDelay time.Duration `mapstructure:"retry-max-delay" json:"retry-max-delay,omitempty"`
	RetryJitter   time.Duration `mapstructure:"retry-jitter" json:"retry-jitter,omitempty"`
	RetryModifier string        `mapstructure:"retry-mode" json:"retry-mode,omitempty"`

	Timeout time.Duration `mapstructure:"timeout" json:"timeout,omitempty"`

	// Hooks
	OnDone []Task `mapstructure:"on-done" json:"on-done,omitempty"`
	OnFail []Task `mapstructure:"on-fail" json:"on-fail,omitempty"`

	// Misc
	Vars map[string]string `mapstructure:"vars" json:"vars,omitempty"`
}

// TaskConnection represents the connection configuration for a task.
type TaskConnection struct {
	Local            bool     `mapstructure:"local" json:"local,omitempty"`
	DockerConnection string   `mapstructure:"docker" json:"docker,omitempty"`
	ContainerName    string   `mapstructure:"container" json:"container,omitempty"`
	ImageName        string   `mapstructure:"image" json:"image,omitempty"`
	Volumes          []string `mapstructure:"volumes" json:"volumes,omitempty"`
	Networks         []string `mapstructure:"networks" json:"networks,omitempty"`
}

type ErrorLimitPolicy string

const (
	ErrorPolKill      ErrorLimitPolicy = "kill"
	ErrorPolGiveUp    ErrorLimitPolicy = "give-up"
	ErrorPolReconnect ErrorLimitPolicy = "reconnect"
)
