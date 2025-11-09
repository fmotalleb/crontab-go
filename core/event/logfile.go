package event

import (
	"bufio"
	"cmp"
	"context"
	"errors"
	"fmt"
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
	if cfg.LogFile == "" {
		return nil, false
	}

	listener, err := NewLogFile(
		cfg.LogFile,
		cfg.LogLineBreaker,
		cfg.LogMatcher,
		cfg.LogCheckCycle,
		log,
	)
	if err != nil {
		log.Error("failed to create LogFile listener", zap.Error(err))
		return nil, false
	}
	return listener, true
}

// LogFile represents a log file that triggers an event when its content changes.
type LogFile struct {
	logger       *zap.Logger
	filePath     string
	lineBreaker  string
	matcher      *regexp.Regexp
	checkCycle   time.Duration
	metricLabels prometheus.Labels
}

func NewLogFile(filePath, lineBreaker, matcherStr string, checkCycle time.Duration, logger *zap.Logger) (*LogFile, error) {
	lineBreaker = cmp.Or(lineBreaker, "\n")
	matcherStr = cmp.Or(matcherStr, ".")
	checkCycle = cmp.Or(checkCycle, time.Second)

	matcher, err := regexp.Compile(matcherStr)
	if err != nil {
		return nil, fmt.Errorf("invalid log matcher: %w", err)
	}
	metricLabels := prometheus.Labels{
		"file":         filePath,
		"line_breaker": lineBreaker,
		"matcher":      matcherStr,
		"check_cycle":  checkCycle.String(),
	}
	global.RegisterCounter(
		LogEventsMetricName,
		LogEventsMetricHelp,
		metricLabels,
	)
	return &LogFile{
		logger: logger.With(
			zap.String("scheduler", "log_file"),
			zap.String("file", filePath),
			zap.String("line_breaker", lineBreaker),
			zap.String("matcher", matcherStr),
			zap.Duration("check_cycle", checkCycle),
		),
		filePath:     filePath,
		lineBreaker:  lineBreaker,
		matcher:      matcher,
		checkCycle:   checkCycle,
		metricLabels: metricLabels,
	}, nil
}

func (lf *LogFile) BuildTickChannel(ed abstraction.EventDispatcher) {
	ctx, cancel := context.WithCancel(global.CTX())
	defer cancel()

	file, err := os.Open(lf.filePath)
	if err != nil {
		lf.logger.Error("failed to open log file", zap.String("path", lf.filePath), zap.Error(err))
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			lf.logger.Warn("failed to close log file", zap.Error(cerr))
		}
	}()

	reader := bufio.NewReader(file)
	// Skip existing content
	if _, err := reader.Discard(math.MaxInt64); err != nil && !errors.Is(err, io.EOF) {
		lf.logger.Warn("failed to skip initial data", zap.Error(err))
	}

	ticker := time.NewTicker(lf.checkCycle)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lf.processNewLines(ctx, reader, ed)
		}
	}
}

func (lf *LogFile) processNewLines(ctx context.Context, reader *bufio.Reader, ed abstraction.EventDispatcher) {
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break // no new data yet
		}
		if err != nil {
			lf.logger.Error("failed reading log file", zap.Error(err))
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		if matches := lf.matcher.FindStringSubmatch(line); matches != nil {
			event := NewMetaData("log-file", map[string]any{
				"file":   lf.filePath,
				"line":   line,
				"groups": reshapeRegexpMatch(lf.matcher.SubexpNames(), matches),
			})
			ed.Emit(ctx, event)
			global.IncMetric(
				LogEventsMetricName,
				LogEventsMetricHelp,
				lf.metricLabels,
			)
		}
	}
}

func reshapeRegexpMatch(keys, matches []string) map[string]string {
	result := make(map[string]string, len(matches))
	for i, key := range keys {
		if key == "" {
			if i == 0 {
				result["0"] = matches[i]
			}
			continue
		}
		result[key] = matches[i]
	}
	return result
}
