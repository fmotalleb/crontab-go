package task

import (
	"bytes"
	"context"
	"encoding/json"
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
	tg.Register(NewPost)
}

// NewPost creates a POST task with the configured retry and timeout settings.
// It returns nil, false when no POST URL is configured.
func NewPost(logger *zap.Logger, task *config.Task) (abstraction.Executable, bool) {
	if task.Post == "" {
		return nil, false
	}
	post := &Post{
		address: task.Post,
		headers: &task.Headers,
		data:    &task.Data,
		task:    task,
		log: logger.With(
			zap.String("url", redactedURL(task.Post)),
			zap.String("method", "post"),
		),
	}
	post.ConfigRetryFrom(task)
	post.SetTimeout(task.Timeout)
	post.SetMetaName("post: " + task.Post)
	post.Action = post
	return post, true
}

type Post struct {
	common.Executable
	common.Cancelable
	common.Timeout
	task *config.Task

	address string
	headers *map[string]string
	data    *any
	log     *zap.Logger
}

// Do sends one POST request with JSON task data, headers, and the task timeout,
// writing the response body to stderr with an execution prefix. Cancel can
// interrupt the request. It returns JSON encoding, request construction, transport,
// response copying, or HTTP status errors for responses of 400 or higher, and
// converts recovered panics to errors. Response body close errors are ignored.
func (p *Post) Do(ctx context.Context) (e error) {
	execID := newExecutionID()
	ctx = populateVars(ctx, p.task)
	_, span := taskTracer.Start(
		ctx, "http.post",
		trace.WithAttributes(
			attribute.String("url.full", redactedURL(p.address)),
			attribute.String("http.request.method", "POST"),
			attribute.String("execution.id", execID),
			attribute.Bool("tls.insecure_skip_verify", p.task.Insecure),
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
	log := p.log.With(
		zap.String("id", execID),
		zap.Time("start", time.Now()),
	)
	log.Info("post started")
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

	localCtx, cancel := p.ApplyTimeout(ctx)
	defer cancel()
	unregisterCancel := p.RegisterCancel(cancel)
	defer unregisterCancel()

	p.setURLAttrs(span)
	p.setRetryAttrs(span)

	var dataReader *bytes.Reader
	if p.data != nil {
		data, err := json.Marshal(p.data)
		if err != nil {
			log.Warn("cannot marshal the given body (pre-send)", zap.Error(err))
			return err
		}
		span.SetAttributes(attribute.Int("http.request.body.size", len(data)))
		dataReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(localCtx, http.MethodPost, p.address, dataReader)
	if err != nil {
		log.Warn("cannot create the request (pre-send)", zap.Error(err))
		return err
	}
	statusCode, execErr := doHTTP(newHTTPClient(p.task.Insecure), req, p.headers, NewPrefixWriter(os.Stderr, execID), log)
	span.SetAttributes(attribute.Int("http.response.status_code", statusCode))
	return execErr
}

func (p *Post) setURLAttrs(span trace.Span) {
	parsed, err := url.Parse(p.address)
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

func (p *Post) setRetryAttrs(span trace.Span) {
	span.SetAttributes(
		attribute.Int64("task.timeout_nanos", int64(p.task.Timeout)),
		attribute.Int64("retry.max_retries", int64(p.task.Retries)),
	)
}
