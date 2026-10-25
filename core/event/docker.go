package event

import (
	"cmp"
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/global"
	"github.com/fmotalleb/crontab-go/core/utils"
)

const (
	DockerEventsMetricName = "docker"
	DockerEventsMetricHelp = "amount of events dispatched using docker"
)

func init() {
	eg.Register(newDockerGenerator)
}

func newDockerGenerator(log *zap.Logger, cfg *config.JobEvent) (abstraction.EventGenerator, bool) {
	if cfg.Docker != nil {
		d := cfg.Docker
		con := cmp.Or(d.Connection,
			"unix:///var/run/docker.sock",
		)

		e := NewDockerEvent(
			con,
			d.Name,
			d.Image,
			d.Actions,
			d.Labels,
			cmp.Or(d.ErrorLimit, 1),
			cmp.Or(d.ErrorLimitPolicy, config.ErrorPolReconnect),
			cmp.Or(d.ErrorThrottle, time.Second*5),
			log,
		)
		return e, true
	}
	return nil, false
}

type DockerEvent struct {
	connection       string
	containerMatcher regexp.Regexp
	imageMatcher     regexp.Regexp
	actions          *utils.List[events.Action]
	labels           map[string]regexp.Regexp
	errorThreshold   uint
	errorPolicy      config.ErrorLimitPolicy
	errorThrottle    time.Duration
	log              *zap.Logger
	metricLabels     prometheus.Labels
}

func NewDockerEvent(
	connection string,
	containerMatcher string,
	imageMatcher string,
	actions []string,
	labels map[string]string,
	errorLimit uint,
	errorPolicy config.ErrorLimitPolicy,
	errorThrottle time.Duration,
	logger *zap.Logger,
) abstraction.EventGenerator {
	metricLabels := prometheus.Labels{
		"connection":       connection,
		"containerMatcher": containerMatcher,
		"imageMatcher":     imageMatcher,
		"actions":          strings.Join(actions, "||"),
	}
	global.RegisterCounter(
		DockerEventsMetricName,
		DockerEventsMetricHelp,
		metricLabels,
	)
	return &DockerEvent{
		connection:       connection,
		containerMatcher: *regexp.MustCompile(containerMatcher),
		imageMatcher:     *regexp.MustCompile(imageMatcher),
		actions:          toAction(actions),
		labels:           reshapeLabelMatcher(labels),
		errorThreshold:   errorLimit,
		errorPolicy:      errorPolicy,
		errorThrottle:    errorThrottle,
		log:              logger,
		metricLabels:     metricLabels,
	}
}

func (dockerEvent *DockerEvent) BuildTickChannel(ed abstraction.EventDispatcher) {
	for {
		if !dockerEvent.connectAndListen(ed) {
			return // stop if policy says to give up
		}
	}
}

func (dockerEvent *DockerEvent) connectAndListen(ed abstraction.EventDispatcher) bool {
	cli, err := client.NewClientWithOpts(
		client.WithHost(dockerEvent.connection),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		dockerEvent.log.Warn("failed to connect to docker", zap.Error(err))
		return dockerEvent.shouldReconnect(global.CTX())
	}
	defer cli.Close()

	ctx, cancel := context.WithCancel(global.CTX())
	defer cancel()

	msg, errs := cli.Events(ctx, events.ListOptions{})
	var errorCount uint

	for {
		select {
		case err, ok := <-errs:
			if !ok {
				dockerEvent.log.Warn("docker error stream closed; reconnecting")
				return dockerEvent.shouldReconnect(ctx)
			}
			action := dockerEvent.handleStreamError(ctx, &errorCount, err)
			switch action {
			case dockerReconnect:
				return dockerEvent.shouldReconnect(ctx)
			case dockerStop:
				return false
			}

		case event, ok := <-msg:
			if !ok {
				dockerEvent.log.Warn("docker event stream closed; reconnecting")
				return dockerEvent.shouldReconnect(ctx)
			}
			dockerEvent.log.Debug("received an event from docker", zap.Any("event", event))
			if dockerEvent.matches(&event) {
				meta := NewMetaData("docker", map[string]any{
					"scope":      event.Scope,
					"action":     event.Action,
					"actor":      event.Actor.ID,
					"attributes": event.Actor.Attributes,
				})
				ed.Emit(ctx, meta)
				global.IncMetric(
					DockerEventsMetricName,
					DockerEventsMetricHelp,
					dockerEvent.metricLabels,
				)
			}
			errorCount = 0
		case <-ctx.Done():
			return false
		}
	}
}

type dockerStreamAction uint8

const (
	dockerKeepListening dockerStreamAction = iota
	dockerReconnect
	dockerStop
)

func (d *DockerEvent) handleStreamError(ctx context.Context, count *uint, err error) dockerStreamAction {
	if err == nil {
		return dockerKeepListening
	}
	d.log.Warn("received an error from docker", zap.Error(err))
	if d.errorThreshold == 0 {
		return dockerKeepListening
	}
	*count++
	if *count >= d.errorThreshold {
		switch d.errorPolicy {
		case config.ErrorPolGiveUp:
			d.log.Error("consecutive errors from docker, giving up", zap.Uint("errors", *count))
			return dockerStop
		case config.ErrorPolKill:
			d.log.Fatal("consecutive errors from docker, killing instance", zap.Uint("errors", *count))
		case config.ErrorPolReconnect:
			d.log.Warn("consecutive errors from docker, reconnecting", zap.Uint("errors", *count))
			return dockerReconnect
		default:
			d.log.Fatal("unexpected ErrorLimitPolicy", zap.Any("policy", d.errorPolicy))
		}
	}
	if d.errorThrottle <= 0 {
		return dockerKeepListening
	}
	timer := time.NewTimer(d.errorThrottle)
	defer timer.Stop()
	select {
	case <-timer.C:
		return dockerKeepListening
	case <-ctx.Done():
		return dockerStop
	}
}

func (dockerEvent *DockerEvent) matches(msg *events.Message) bool {
	if dockerEvent.actions.IsNotEmpty() && !dockerEvent.actions.Contains(msg.Action) {
		return false
	}
	if !dockerEvent.containerMatcher.MatchString(msg.Actor.Attributes["name"]) {
		return false
	}

	if !dockerEvent.imageMatcher.MatchString(msg.Actor.Attributes["image"]) {
		return false
	}

	for k, matcher := range dockerEvent.labels {
		if attrib, ok := msg.Actor.Attributes[k]; !ok || !matcher.MatchString(attrib) {
			return false
		}
	}
	return true
}

func (dockerEvent *DockerEvent) shouldReconnect(ctx context.Context) bool {
	switch dockerEvent.errorPolicy {
	case config.ErrorPolReconnect:
		dockerEvent.log.Warn("retrying docker connection after failure")
		if dockerEvent.errorThrottle <= 0 {
			return ctx.Err() == nil
		}
		timer := time.NewTimer(dockerEvent.errorThrottle)
		defer timer.Stop()
		select {
		case <-timer.C:
			return ctx.Err() == nil
		case <-ctx.Done():
			return false
		}
	case config.ErrorPolGiveUp:
		dockerEvent.log.Error("giving up on docker connection")
		return false
	case config.ErrorPolKill:
		dockerEvent.log.Fatal("docker connection failed, killing instance")
	default:
		dockerEvent.log.Fatal("unexpected ErrorLimitPolicy", zap.Any("policy", dockerEvent.errorPolicy))
	}
	return false
}

func reshapeLabelMatcher(labels map[string]string) map[string]regexp.Regexp {
	res := make(map[string]regexp.Regexp)
	for k, v := range labels {
		res[k] = *regexp.MustCompile(v)
	}
	return res
}

func toAction(acts []string) *utils.List[events.Action] {
	actions := utils.NewList[events.Action]()
	for _, act := range acts {
		actions.Add(events.Action(act))
	}
	return actions
}
