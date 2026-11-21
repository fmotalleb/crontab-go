package task

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/common"
)

func init() {
	tg.Register(NewGet)
}

// NewGet creates a GET task with the configured retry and timeout settings.
// It returns nil, false when no GET URL is configured.
func NewGet(logger *zap.Logger, task *config.Task) (abstraction.Executable, bool) {
	if task.Get == "" {
		return nil, false
	}
	get := &Get{
		address: task.Get,
		task:    task,
		headers: &task.Headers,
		log: logger.With(
			zap.String("url", redactedURL(task.Get)),
			zap.String("method", "get"),
		),
	}
	get.ConfigRetryFrom(task)
	get.SetTimeout(task.Timeout)
	get.SetMetaName("get: " + task.Get)
	get.Action = get
	return get, true
}

type Get struct {
	common.Executable
	common.Cancelable
	common.Timeout
	task    *config.Task
	address string
	headers *map[string]string
	log     *zap.Logger
}

// Do sends one GET request with the task headers and timeout, writing the
// response body to stderr with an execution prefix. Cancel can interrupt the request.
// It returns request construction, transport, response copying, or HTTP status
// errors for responses of 400 or higher, and converts recovered panics to errors.
// Response body close errors are ignored.
func (g *Get) Do(ctx context.Context) (e error) {
	execID := newExecutionID()
	ctx = populateVars(ctx, g.task)
	_, span := taskTracer.Start(
		ctx, "http.get",
		trace.WithAttributes(
			attribute.String("url.full", redactedURL(g.address)),
			attribute.String("http.request.method", "GET"),
			attribute.String("execution.id", execID),
			attribute.Bool("tls.insecure_skip_verify", g.task.Insecure),
		),
	)
	defer span.End()
	defer func() {
		if e != nil {
			span.RecordError(e)
			span.SetStatus(codes.Error, e.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}()
	log := g.log.With(
		zap.String("id", execID),
		zap.Time("start", time.Now()),
	)
	log.Info("get started")
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				log.Error("panic recovered", zap.Error(err))
				e = err
			} else {
				log.Error("panic recovered", zap.Any("error", r))
				e = fmt.Errorf("panic: %v", r)
			}
		}
	}()

	localCtx, cancel := g.ApplyTimeout(ctx)
	defer cancel()
	unregisterCancel := g.RegisterCancel(cancel)
	defer unregisterCancel()

	g.setURLAttrs(span)
	g.setRetryAttrs(span)

	req, err := http.NewRequestWithContext(localCtx, http.MethodGet, g.address, nil)
	if err != nil {
		log.Warn("cannot create the request (pre-send)", zap.Error(err))
		return err
	}
	statusCode, execErr := doHTTP(newHTTPClient(g.task.Insecure), req, g.headers, NewPrefixWriter(os.Stderr, execID), log)
	span.SetAttributes(attribute.Int("http.response.status_code", statusCode))
	return execErr
}

func (g *Get) setURLAttrs(span trace.Span) {
	parsed, err := url.Parse(g.address)
	if err != nil {
		return
	}
	span.SetAttributes(
		attribute.String("url.scheme", parsed.Scheme),
		attribute.String("server.address", parsed.Hostname()),
	)
	if port := parsed.Port(); port != "" {
		span.SetAttributes(attribute.String("server.port", port))
	}
}

func (g *Get) setRetryAttrs(span trace.Span) {
	span.SetAttributes(
		attribute.Int64("task.timeout_nanos", int64(g.task.Timeout)),
		attribute.Int64("retry.max_retries", int64(g.task.Retries)),
	)
}
