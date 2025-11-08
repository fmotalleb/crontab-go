package task

import (
	"bytes"
	"context"
	"net/http"

	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

func logHTTPResponse(r *http.Response) (string, error) {
	result := bytes.NewBuffer([]byte{})
	err := r.Write(result)
	return result.String(), err
}

func getFailedConnections(ctx context.Context) []config.TaskConnection {
	items := ctx.Value(ctxutils.FailedRemotes)
	if items != nil {
		return items.([]config.TaskConnection)
	}
	return []config.TaskConnection{}
}

func addFailedConnections(ctx context.Context, con config.TaskConnection) context.Context {
	current := getFailedConnections(ctx)
	return context.WithValue(ctx, ctxutils.FailedRemotes, append(current, con))
}
