package task

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/common"
	"github.com/fmotalleb/crontab-go/helpers"
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
		task:    task,
		headers: &task.Headers,
		log: logger.With(
			zap.String("url", task.Get),
			zap.String("method", "get"),
		),
	}
	get.ConfigRetryFrom(task)
	get.SetTimeout(task.Timeout)
	get.SetMetaName("get: " + task.Get)
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

// Execute implements abstraction.Executable.
func (g *Get) Do(ctx context.Context) (e error) {
	ctx = populateVars(ctx, g.task)
	log := g.log.With(
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

	localCtx, cancel := g.ApplyTimeout(ctx)
	g.SetCancel(cancel)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(localCtx, http.MethodGet, g.address, nil)
	log.Debug("sending get http request")
	if err != nil {
		log.Warn("cannot create the request (pre-send)", zap.Error(err))
		return err
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
	if err != nil {
		log.Warn("request failed", zap.Error(err))
		return err
	}
	if res != nil && res.StatusCode >= 400 {
		err = fmt.Errorf("unexpected http status code: %d", res.StatusCode)
		log.Warn("request failed", zap.Error(err))
		return err
	}
	return nil
}
