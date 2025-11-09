package event

import (
	"bufio"
	"cmp"
	"context"
	"errors"
	"io"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/global"
)

const (
	LogEventsMetricName = "log_file"
	LogEventsMetricHelp = "amount of events dispatched using log-file"
)

// TODO[epic=events] add watch method (probably after fs watcher is implemented)

func init() {
	eg.Register(newLogListenerGenerator)
}

func newLogListenerGenerator(log *zap.Logger, cfg *config.JobEvent) (abstraction.EventGenerator, bool) {
	if cfg.LogFile != "" {
		listener, err := NewLogFile(
			cfg.LogFile,
			cfg.LogLineBreaker,
			cfg.LogMatcher,
			cfg.LogCheckCycle,
			log,
		)
		if err != nil {
			log.Error("Error creating LogFileListener", zap.Error(err))
			return nil, false
		}
		return listener, true
	}
	return nil, false
}

// LogFile represents a log file that triggers an event when its content changes.
type LogFile struct {
	logger      *zap.Logger
	filePath    string
	lineBreaker string
	matcher     regexp.Regexp
	// possibly will be deprecated or changed to an struct
	checkCycle time.Duration
}

// NewLogFile creates a new LogFile with the given parameters.
func NewLogFile(filePath string, lineBreaker string, matcherStr string, checkCycle time.Duration, logger *zap.Logger) (*LogFile, error) {
	lineBreaker = cmp.Or(lineBreaker, "\n")
	matcherStr = cmp.Or(matcherStr, ".")
	checkCycle = cmp.Or(checkCycle, time.Second)

	matcher, err := regexp.Compile(
		matcherStr,
	)
	if err != nil {
		return nil, err
	}
	global.RegisterCounter(
		LogEventsMetricName,
		LogEventsMetricHelp,
		prometheus.Labels{"file": filePath},
	)
	return &LogFile{
		logger: logger.With(
			zap.String("scheduler", "log_file"),
			zap.String("file", filePath),
			zap.String("line_breaker", lineBreaker),
			zap.String("matcher", matcherStr),
			zap.Duration("check_cycle", checkCycle),
		),
		filePath:    filePath,
		lineBreaker: lineBreaker,
		matcher:     *matcher,
		checkCycle:  checkCycle,
	}, nil
}

// BuildTickChannel implements abstraction.Scheduler.
func (lf *LogFile) BuildTickChannel(ed abstraction.EventDispatcher) {
	file, err := os.Open(lf.filePath)
	if err != nil {
		lf.logger.Error("failed to open log file", zap.Error(err))
	}
	defer func() {
		if err = file.Close(); err != nil {
			lf.logger.Warn("failed to close log file", zap.Error(err))
		}
	}()
	reader := bufio.NewReader(file)
	_, err = reader.Discard(math.MaxInt64)

	ctx, cancel := context.WithCancel(global.CTX())
	defer cancel()
	if err != nil && !errors.Is(err, io.EOF) {
		lf.logger.Warn("error skipping initial data", zap.Error(err))
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		data, err := reader.ReadString(byte(0))
		if err != nil && err != io.EOF {
			lf.logger.Error("error reading log file", zap.Error(err))
			return
		}
		for _, line := range strings.Split(data, lf.lineBreaker) {
			matches := lf.matcher.FindStringSubmatch(line)
			if matches != nil {
				names := lf.matcher.SubexpNames()
				event := NewMetaData(
					"log-file",
					map[string]any{
						"file":   lf.filePath,
						"line":   line,
						"groups": reshapeRegxpMatch(names, matches),
					},
				)
				ed.Emit(ctx, event)
				global.IncMetric(
					LogEventsMetricName,
					LogEventsMetricHelp,
					prometheus.Labels{"file": lf.filePath},
				)
			}
		}
		time.Sleep(lf.checkCycle)
	}
}

func reshapeRegxpMatch(keys []string, matches []string) map[string]string {
	result := make(map[string]string)

	for i, key := range keys {
		if key != "" {
			result[key] = matches[i]
		} else if i == 0 {
			result["0"] = matches[i]
		}
	}
	return result
}
