package event

import (
	"bufio"
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
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
	matcher      *regexp.Regexp
	checkCycle   time.Duration
	metricLabels prometheus.Labels
}

func NewLogFile(filePath, matcherStr string, checkCycle time.Duration, logger *zap.Logger) (*LogFile, error) {
	if checkCycle < 0 {
		return nil, fmt.Errorf("log check cycle cannot be negative: %s", checkCycle)
	}
	matcherStr = cmp.Or(matcherStr, ".")
	checkCycle = cmp.Or(checkCycle, time.Second)

	matcher, err := regexp.Compile(matcherStr)
	if err != nil {
		return nil, fmt.Errorf("invalid log matcher: %w", err)
	}
	metricLabels := prometheus.Labels{
		"file":        filePath,
		"matcher":     matcherStr,
		"check_cycle": checkCycle.String(),
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
			zap.String("matcher", matcherStr),
			zap.Duration("check_cycle", checkCycle),
		),
		filePath:     filePath,
		matcher:      matcher,
		checkCycle:   checkCycle,
		metricLabels: metricLabels,
	}, nil
}

func (lf *LogFile) BuildTickChannel(ed abstraction.EventDispatcher) {
	ctx, cancel := context.WithCancel(global.CTX())
	defer cancel()
	var (
		file     *os.File
		reader   *bufio.Reader
		fileInfo os.FileInfo
		pending  string
		offset   int64
	)

	defer func() {
		if file != nil {
			if cerr := file.Close(); cerr != nil {
				lf.logger.Warn("failed to close log file", zap.Error(cerr))
			}
		}
	}()

	ticker := time.NewTicker(lf.checkCycle)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pathInfo, statErr := os.Stat(lf.filePath)
			if statErr != nil {
				if !errors.Is(statErr, os.ErrNotExist) {
					lf.logger.Warn("failed to stat log file", zap.String("path", lf.filePath), zap.Error(statErr))
				}
				continue
			}
			switch {
			case file == nil:
				file, reader, fileInfo, offset = lf.openFile(true)
				if file == nil {
					continue
				}
			case !os.SameFile(fileInfo, pathInfo):
				if err := file.Close(); err != nil {
					lf.logger.Warn("failed to close rotated log file", zap.Error(err))
				}
				file, reader, fileInfo, offset = lf.openFile(false)
				pending = ""
				if file == nil {
					continue
				}
			case pathInfo.Size() < offset:
				if _, err := file.Seek(0, io.SeekStart); err != nil {
					lf.logger.Warn("failed to seek truncated log file", zap.Error(err))
					continue
				}
				reader = bufio.NewReader(file)
				offset = 0
				pending = ""
			}
			lf.processNewLines(ctx, reader, ed, &pending, &offset)
		}
	}
}

func (lf *LogFile) openFile(skipExisting bool) (*os.File, *bufio.Reader, os.FileInfo, int64) {
	file, err := os.Open(lf.filePath)
	if err != nil {
		lf.logger.Warn("failed to open log file", zap.String("path", lf.filePath), zap.Error(err))
		return nil, nil, nil, 0
	}
	info, err := file.Stat()
	if err != nil {
		lf.logger.Warn("failed to stat opened log file", zap.Error(err))
		_ = file.Close()
		return nil, nil, nil, 0
	}
	var offset int64
	if skipExisting {
		offset, err = file.Seek(0, io.SeekEnd)
		if err != nil {
			lf.logger.Warn("failed to seek to end of log file", zap.Error(err))
			_ = file.Close()
			return nil, nil, nil, 0
		}
	}
	return file, bufio.NewReader(file), info, offset
}

func (lf *LogFile) processNewLines(ctx context.Context, reader *bufio.Reader, ed abstraction.EventDispatcher, pending *string, offset *int64) {
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			*pending += line
			*offset += int64(len(line))
			break // no new data yet
		}
		if err != nil {
			lf.logger.Error("failed reading log file", zap.Error(err))
			return
		}
		*offset += int64(len(line))
		line = *pending + line
		*pending = ""
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
