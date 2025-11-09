// Package webserver implements the logic for the webserver
package webserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fmotalleb/go-tools/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/core/webserver/endpoint"
	"github.com/fmotalleb/crontab-go/helpers"
)

type AuthConfig struct {
	Username string
	Password string
}

type WebServer struct {
	*AuthConfig
	ctx          context.Context
	address      string
	port         uint
	log          *zap.Logger
	serveMetrics bool
}

func NewWebServer(ctx context.Context,
	address string,
	port uint,
	serveMetrics bool,
	authentication *AuthConfig,
) *WebServer {
	return &WebServer{
		ctx:          ctx,
		address:      address,
		port:         port,
		AuthConfig:   authentication,
		log:          log.Of(ctx).Named("WebServer"),
		serveMetrics: serveMetrics,
	}
}

func (s *WebServer) Serve() {
	engine := echo.New()
	auth := func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
	if s.AuthConfig != nil && s.AuthConfig.Username != "" && s.AuthConfig.Password != "" {
		auth = middleware.BasicAuth(func(username, password string, _ echo.Context) (bool, error) {
			if username == s.AuthConfig.Username && password == s.AuthConfig.Password {
				return true, nil
			}
			return false, nil
		})
	} else {
		s.log.Warn("received no value on username or password, ignoring any authentication, if you intended to use no authentication ignore this message")
	}

	engine.Use(
		auth,
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogURI:      true,
			LogStatus:   true,
			LogError:    true,
			HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				if v.Error == nil {
					s.log.Debug(
						"request served",
						zap.String("URI", v.URI),
						zap.Int("status", v.Status),
					)
				} else {
					s.log.Warn(
						"request failed",
						zap.String("URI", v.URI),
						zap.Int("status", v.Status),
						zap.Error(v.Error),
					)
				}
				return nil
			},
		}),
		middleware.Recover(),
	)

	engine.GET(
		"/foo",
		func(c echo.Context) error {
			return c.String(200, "bar")
		},
	)

	ed := &endpoint.EventDispatchEndpoint{}
	engine.Any(
		"/events/:event/emit",
		ed.Endpoint,
	)
	if s.serveMetrics {
		engine.GET("/metrics", func(c echo.Context) error {
			promhttp.Handler().ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})
	} else {
		engine.GET("/metrics", func(c echo.Context) error {
			return c.String(http.StatusNotFound, "Metrics are disabled, please enable metrics using `WEBSERVER_METRICS=true`")
		})
	}

	err := engine.Start(fmt.Sprintf("%s:%d", s.address, s.port))
	helpers.FatalOnErr(
		s.log,
		func() error {
			return err
		},
		"Failed to start webserver: %s",
	)
}
