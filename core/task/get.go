package task

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/common"
	"github.com/FMotalleb/crontab-go/helpers"
)

func init() {
	tg.Register(NewGet)
}

func NewGet(logger *zap.Logger, task *config.Task) (abstraction.Executable, bool) {
	if task.Get == "" {
		return nil, false
	}
	get := &Get{
		address: task.Get,
		headers: &task.Headers,
		log: logger.With(
			zap.String("url", task.Get),
			zap.String("method", "get"),
		),
	}
	get.SetMaxRetry(task.Retries)
	get.SetRetryDelay(task.RetryDelay)
	get.SetTimeout(task.Timeout)
	get.SetMetaName("get: " + task.Get)
	return get, true
}

type Get struct {
	common.Hooked
	common.Cancelable
	common.Retry
	common.Timeout

	address string
	headers *map[string]string
	log     *zap.Logger
}

// Execute implements abstraction.Executable.
func (g *Get) Execute(ctx context.Context) (e error) {
	r := common.GetRetry(ctx)
	log := g.log.With(
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

	if err := g.WaitForRetry(ctx); err != nil {
		g.DoFailHooks(ctx)
		return err
	}
	ctx = common.IncreaseRetry(ctx)

	localCtx, cancel := g.ApplyTimeout(ctx)
	g.SetCancel(cancel)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(localCtx, http.MethodGet, g.address, nil)
	log.Debug("sending get http request")
	if err != nil {
		log.Warn("cannot create the request (pre-send)", zap.Error(err))
		return g.Execute(ctx)
	}
	for key, val := range *g.headers {
		req.Header.Add(key, val)
	}
	res, err := client.Do(req)
	if res != nil {
		if res.Body != nil {
			defer helpers.WarnOnErrIgnored(
				log,
				res.Body.Close,
				"cannot close response body: %s",
			)
		}
		log = log.With(zap.Int("status", res.StatusCode))
		log.Info("received response with status", zap.String("status", res.Status))

		if log.Level() >= zap.DebugLevel {
			ans, err := logHTTPResponse(res)
			log.Debug("fetched data", zap.String("response", ans), zap.Error(err))
		}
	}
	if err != nil || res.StatusCode >= 400 {
		log.Warn("request failed", zap.Error(err))
		return g.Execute(ctx)
	}
	g.DoDoneHooks(ctx)
	return nil
}
