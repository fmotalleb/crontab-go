package event

import (
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
	"github.com/fmotalleb/crontab-go/core/concurrency"
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
		con := utils.FirstNonZeroForced(d.Connection,
			"unix:///var/run/docker.sock",
		)

		e := NewDockerEvent(
			con,
			d.Name,
			d.Image,
			d.Actions,
			d.Labels,
			utils.FirstNonZeroForced(d.ErrorLimit, 1),
			utils.FirstNonZeroForced(d.ErrorLimitPolicy, config.ErrorPolReconnect),
			utils.FirstNonZeroForced(d.ErrorThrottle, time.Second*5),
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
		return dockerEvent.shouldReconnect()
	}
	defer cli.Close()

	ctx, cancel := context.WithCancel(global.CTX())
	defer cancel()

	msg, errs := cli.Events(ctx, events.ListOptions{})
	errCount := concurrency.NewLockedValue(uint(0))

	for {
		select {
		case err := <-errs:
			if err == nil {
				continue
			}

			dockerEvent.log.Warn("received an error from docker", zap.Error(err))
			if dockerEvent.errorThreshold == 0 {
				continue
			}

			count := errCount.Get() + 1
			errCount.Set(count)

			if count >= dockerEvent.errorThreshold {
				switch dockerEvent.errorPolicy {
				case config.ErrorPolGiveUp:
					dockerEvent.log.Error("consecutive errors from docker, giving up", zap.Uint("errors", count))
					return false
				case config.ErrorPolKill:
					dockerEvent.log.Fatal("consecutive errors from docker, killing instance", zap.Uint("errors", count))
				case config.ErrorPolReconnect:
					dockerEvent.log.Warn("consecutive errors from docker, reconnecting", zap.Uint("errors", count))
					return true
				default:
					dockerEvent.log.Fatal("unexpected ErrorLimitPolicy", zap.Any("policy", dockerEvent.errorPolicy))
				}
				errCount.Set(0)
			}

			if dockerEvent.errorThrottle > 0 {
				time.Sleep(dockerEvent.errorThrottle)
			}

		case event := <-msg:
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
			errCount.Set(0)
		}
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
		if attrib, ok := msg.Actor.Attributes[k]; !ok && !matcher.MatchString(attrib) {
			return false
		}
	}
	return true
}

func (dockerEvent *DockerEvent) shouldReconnect() bool {
	switch dockerEvent.errorPolicy {
	case config.ErrorPolReconnect:
		dockerEvent.log.Warn("retrying docker connection after failure")
		time.Sleep(dockerEvent.errorThrottle)
		return true
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
