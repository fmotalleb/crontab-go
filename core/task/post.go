package task

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/common"
	"github.com/fmotalleb/crontab-go/helpers"
)

func init() {
	tg.Register(NewPost)
}

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
			zap.String("url", task.Post),
			zap.String("method", "post"),
		),
	}
	post.SetMaxRetry(task.Retries)
	post.SetRetryDelay(task.RetryDelay)
	post.SetTimeout(task.Timeout)
	post.SetMetaName("post: " + task.Post)
	return post, true
}

type Post struct {
	common.Hooked
	common.Cancelable
	common.Retry
	common.Timeout
	task *config.Task

	address string
	headers *map[string]string
	data    *any
	log     *zap.Logger
}

// Execute implements abstraction.Executable.
func (p *Post) Execute(ctx context.Context) (e error) {
	ctx = populateVars(ctx, p.task)
	r := common.GetRetry(ctx)
	log := p.log.With(
		zap.Any("retry", r),
		zap.Time("start", time.Now()),
	)
	defer func() {
		err := recover()
		if err != nil {
			if err, ok := err.(error); ok {
				log.Warn("recovering command execution from a fatal error", zap.Error(err))
				return
			}
			log.Warn("a non-error panic accord", zap.Any("error", err))
		}
	}()

	if err := p.WaitForRetry(ctx); err != nil {
		p.DoFailHooks(ctx)
		return err
	}

	ctx = common.IncreaseRetry(ctx)

	var localCtx context.Context
	var cancel context.CancelFunc
	localCtx, cancel = p.ApplyTimeout(ctx)
	p.SetCancel(cancel)

	client := &http.Client{}
	var dataReader *bytes.Reader
	if p.data != nil {
		data, err := json.Marshal(p.data)
		if err != nil {
			log.Warn("cannot marshal the given body (pre-send)", zap.Error(err))
			return p.Execute(ctx)
		}
		dataReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(localCtx, http.MethodPost, p.address, dataReader)
	log.Debug("sending post http request")
	if err != nil {
		log.Warn("cannot create the request (pre-send)", zap.Error(err))
		return p.Execute(ctx)
	}

	for key, val := range *p.headers {
		req.Header.Add(key, val)
	}

	res, err := client.Do(req)

	if res != nil {
		if res.Body != nil {
			defer helpers.WarnOnErrIgnored(
				log,
				res.Body.Close,
				"cannot close response body",
			)
		}
		log = log.With(zap.Int("status", res.StatusCode))
		log.Info("received response with status", zap.String("status", res.Status))
		if log.Level() >= zap.DebugLevel {
			ans, respErr := logHTTPResponse(res)
			log.Debug("fetched data", zap.String("response", ans), zap.Error(respErr))
		}
	}

	if err != nil || (res != nil && res.StatusCode >= 400) {
		log.Warn("request failed", zap.Error(err))
		return p.Execute(ctx)
	}

	p.DoDoneHooks(ctx)
	return nil
}
