package task

import (
	"bytes"
	"context"
	"net/http"

	"github.com/fmotalleb/go-tools/log"
	"github.com/fmotalleb/go-tools/template"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

func logHTTPResponse(r *http.Response) (string, error) {
	result := bytes.NewBuffer([]byte{})
	err := r.Write(result)
	return result.String(), err
}

func populateVars(ctx context.Context, task *config.Task) context.Context {
	var ok bool
	var old map[string]string
	if old, ok = ctx.Value(ctxutils.Vars).(map[string]string); !ok {
		old = make(map[string]string, 0)
	}
	varTable := old
	for k, v := range task.Vars {
		var err error
		varTable[k], err = template.EvaluateTemplate(v, varTable)
		if err != nil {
			log.Of(ctx).Error(
				"failed to evaluate template on variable",
				zap.String("key", k),
				zap.String("value", v),
				zap.Error(err),
			)
		}
	}
	return context.WithValue(ctx, ctxutils.Vars, varTable)
}
