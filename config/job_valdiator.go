package config

import (
	"fmt"
	"regexp"

	"github.com/docker/docker/api/types/events"
	"go.uber.org/zap"

	"github.com/FMotalleb/crontab-go/core/utils"
)

var acceptedActions = utils.NewList(
	events.ActionCreate,
	events.ActionStart,
	events.ActionRestart,
	events.ActionStop,
	events.ActionCheckpoint,
	events.ActionPause,
	events.ActionUnPause,
	events.ActionAttach,
	events.ActionDetach,
	events.ActionResize,
	events.ActionUpdate,
	events.ActionRename,
	events.ActionKill,
	events.ActionDie,
	events.ActionOOM,
	events.ActionDestroy,
	events.ActionRemove,
	events.ActionCommit,
	events.ActionTop,
	events.ActionCopy,
	events.ActionArchivePath,
	events.ActionExtractToDir,
	events.ActionExport,
	events.ActionImport,
	events.ActionSave,
	events.ActionLoad,
	events.ActionTag,
	events.ActionUnTag,
	events.ActionPush,
	events.ActionPull,
	events.ActionPrune,
	events.ActionDelete,
	events.ActionEnable,
	events.ActionDisable,
	events.ActionConnect,
	events.ActionDisconnect,
	events.ActionReload,
	events.ActionMount,
	events.ActionUnmount,
	events.ActionExecCreate,
	events.ActionExecStart,
	events.ActionExecDie,
	events.ActionExecDetach,
	events.ActionHealthStatus,
	events.ActionHealthStatusRunning,
	events.ActionHealthStatusHealthy,
	events.ActionHealthStatusUnhealthy,
)

func (c *JobConfig) Validate(log *zap.Logger) error {
	// Log the start of validation
	log.Debug("begin Validation")

	// Check if the job is disabled
	if c.Disabled {
		// Log the disabled job
		log.Warn("job is disabled")
		return nil
	}
	checkList := []func(*JobConfig, *zap.Logger) error{
		validateEvents,
		validateTasks,
		validateJobHooks,
	}
	for _, check := range checkList {
		if err := check(c, log); err != nil {
			return err
		}
	}

	// Log the successful validation
	log.Info("validation successful")
	return nil
}

func validateTasks(c *JobConfig, log *zap.Logger) error {
	for _, t := range c.Tasks {
		if err := t.Validate(log); err != nil {
			log.Error("Validation error in task for JobConfig", zap.Error(err))
			return err
		}
	}
	return nil
}

func validateEvents(c *JobConfig, log *zap.Logger) error {
	for _, s := range c.Events {
		if err := s.Validate(log); err != nil {
			log.Error("Validation error in event for JobConfig", zap.Error(err))
			return err
		}
	}
	return nil
}

func validateJobHooks(c *JobConfig, log *zap.Logger) error {
	for _, t := range c.Hooks.Failed {
		if err := t.Validate(log); err != nil {
			log.Error("Validation error in failed hook for JobConfig", zap.Error(err))
			return err
		}
	}
	for _, t := range c.Hooks.Done {
		if err := t.Validate(log); err != nil {
			log.Error("Validation error in done hook for JobConfig", zap.Error(err))
			return err
		}
	}
	return nil
}

// Validate checks the validity of a JobEvent configuration.
// It ensures that the event has a valid interval or cron expression, and only one of on_init, interval, or cron is set.
// It returns an error if the validation fails, otherwise, it returns nil.
func (s *JobEvent) Validate(log *zap.Logger) error {
	// Check if the interval is a negative value
	if s.Interval < 0 {
		err := fmt.Errorf("received a negative time in interval: `%v`", s.Interval)
		log.Warn("Validation failed for JobEvent", zap.Error(err))
		return err
	} else if _, err := DefaultCronParser.Parse(s.Cron); s.Cron != "" && err != nil {
		log.Warn("Validation failed for JobEvent", zap.Error(err))
		return err
	} else if s.Docker != nil {
		returnValue := dockerValidation(s, log)
		if returnValue != nil {
			return returnValue
		}
	}

	// Check the active events to ensure only one of on_init, interval, docker, or cron is set
	events := utils.NewList(s.Interval != 0,
		s.Cron != "",
		s.WebEvent != "",
		s.Docker != nil,
		s.LogFile != "",
		s.OnInit,
	)
	activeEvents := utils.Fold(events, 0, func(c int, item bool) int {
		if item {
			return c + 1
		}
		return c
	})

	if activeEvents != 1 {
		err := fmt.Errorf(
			"a single event must have one of (on-init: true,interval,cron,web-event,docker,log-file) field, received:(on_init: %t,cron: `%s`, interval: `%s`, web_event: `%s`, docker: %v,log-file: %v)",
			s.OnInit,
			s.Cron,
			s.Interval,
			s.WebEvent,
			s.Docker,
			s.LogFile,
		)
		log.Warn("Validation failed for JobEvent", zap.Error(err))
		return err
	}

	// Log the successful validation
	log.Debug("Validation successful for JobEvent", zap.Any("event", s))
	return nil
}

func dockerValidation(s *JobEvent, log *zap.Logger) error {
	// Check if regex matchers are valid
	checkList := utils.NewList[string]()
	checkList.Add(
		s.Docker.Name,
		s.Docker.Image,
	)
	for _, v := range s.Docker.Labels {
		checkList.Add(v)
	}
	err := utils.Fold(checkList, nil, func(initial error, _ string) error {
		if initial != nil {
			return initial
		}
		_, err := regexp.Compile(s.Docker.Name)
		return err
	})
	if err != nil {
		log.Warn("Validation failed for one of docker regex pattern (container name, image name, labels value)", zap.Error(err))
		return err
	}
	for _, i := range s.Docker.Actions {
		if !acceptedActions.Contains(events.Action(i)) {
			err := fmt.Errorf("given action: %#v is not allowed", i)
			log.Warn("Validation failed for one of docker actions", zap.Error(err))
			return err
		}
	}
	// Validating error handler parameters
	if s.Docker.ErrorLimit > 0 {
		log.Debug("error limit will be set to 1")
	}
	if s.Docker.ErrorLimitPolicy == "" {
		log.Info("no error policy was specified, using default policy (reconnect)")
	}
	if !utils.NewList("", ErrorPolGiveUp, ErrorPolKill, ErrorPolReconnect).Contains(s.Docker.ErrorLimitPolicy) {
		err := fmt.Errorf("given error limit policy: %#v is not allowed, possible error policies are (give-up,kill,reconnect)", s.Docker.ErrorLimitPolicy)
		log.Warn("Validation failed for docker error limit policy", zap.Error(err))
		return err
	}
	if s.Docker.ErrorThrottle < 0 {
		err := fmt.Errorf("received a negative throttle value: `%v`", s.Docker.ErrorThrottle)
		log.Warn("Validation failed for docker, throttling value error", zap.Error(err))
		return err
	}
	return nil
}
