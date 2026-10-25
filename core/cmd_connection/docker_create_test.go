package connection

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/config"
)

type dockerRoundTripper func(*http.Request) (*http.Response, error)

func (rt dockerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

func TestDockerCreateCleanupContext(t *testing.T) {
	for _, failure := range []string{"start cancellation", "wait cancellation", "wait error", "logs error", "none"} {
		t.Run(failure, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
			defer cancel()
			var cleanupCtx context.Context
			transport := dockerRoundTripper(func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodDelete {
					cleanupCtx = req.Context()
					assert.NoError(t, cleanupCtx.Err())
					deadline, ok := cleanupCtx.Deadline()
					assert.True(t, ok)
					assert.True(t, time.Until(deadline) > 0 && time.Until(deadline) <= 10*time.Second)
					assert.Equal(t, "1", req.URL.Query().Get("force"))
					return dockerResponse(http.StatusNoContent, ""), nil
				}
				return dockerExecuteResponse(req, failure, cancel)
			})
			cli, err := client.NewClientWithOpts(client.WithHost("http://docker.test"), client.WithVersion("1.45"),
				client.WithHTTPClient(&http.Client{Transport: transport}))
			assert.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, cli.Close()) })
			d := &DockerCreateConnection{
				ctx: ctx, cli: cli, log: zap.NewNop(),
				conn: &config.TaskConnection{}, containerConfig: &container.Config{},
			}
			err = d.Execute(io.Discard, io.Discard)
			if failure == "none" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.NotEqual(t, nil, cleanupCtx, "container removal must be attempted")
			assert.Equal(t, context.Canceled, cleanupCtx.Err(), "cleanup must cancel its context on completion")
		})
	}
}

func dockerExecuteResponse(req *http.Request, failure string, cancel context.CancelFunc) (*http.Response, error) {
	switch {
	case strings.HasSuffix(req.URL.Path, "/create"):
		return dockerResponse(http.StatusCreated, `{"Id":"test-container"}`), nil
	case strings.HasSuffix(req.URL.Path, "/start"):
		if failure == "start cancellation" {
			cancel()
			return nil, context.Canceled
		}
		return dockerResponse(http.StatusNoContent, ""), nil
	case strings.HasSuffix(req.URL.Path, "/wait"):
		if failure == "wait cancellation" {
			cancel()
			return nil, context.Canceled
		}
		if failure == "wait error" {
			return dockerResponse(http.StatusInternalServerError, `{"message":"wait failed"}`), nil
		}
		return dockerResponse(http.StatusOK, `{"StatusCode":0}`), nil
	default:
		if failure == "logs error" {
			return dockerResponse(http.StatusInternalServerError, `{"message":"logs failed"}`), nil
		}
		return dockerResponse(http.StatusOK, ""), nil
	}
}

func dockerResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
