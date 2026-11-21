// Package webserver implements the logic for the webserver
package webserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/core/webserver/endpoint"
)

type AuthConfig struct {
	Username             string
	Password             string
	AllowUnauthenticated bool
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

// newEngine configures event and metrics routes with a 1 MB request body limit.
// All routes require Basic authentication unless unauthenticated access is
// explicitly allowed with both credentials empty; other configurations yield 401.
// Event dispatch accepts POST and returns 429 when 64 dispatch requests are active.
func (s *WebServer) newEngine() *echo.Echo {
	engine := echo.New()

	switch {
	case s.AuthConfig != nil && s.Username != "" && s.Password != "":
		engine.Use(middleware.BasicAuth(func(username, password string, _ echo.Context) (bool, error) {
			if username == s.Username && password == s.Password {
				return true, nil
			}
			return false, nil
		}))
	case s.AuthConfig != nil && s.AllowUnauthenticated && s.Username == "" && s.Password == "":
		s.log.Warn("webserver event dispatch is unauthenticated; use only on a trusted network")
	default:
		engine.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return c.String(http.StatusUnauthorized, "webserver authentication is not configured")
			}
		})
	}

	engine.Use(
		middleware.BodyLimit("1M"),
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogURI:      true,
			LogStatus:   true,
			LogError:    true,
			HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
			LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
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

	ed := endpoint.NewEventDispatchEndpoint()
	engine.POST("/events/:event/emit", ed.Endpoint)
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
	engine.HideBanner = true
	engine.HidePort = true
	engine.Debug = s.log.Level() == zap.DebugLevel
	return engine
}

// Serve listens for HTTP requests until the server stops or s.ctx is canceled.
// On cancellation it attempts graceful shutdown with a ten-second timeout.
// It wraps listen and serve errors, treating cancellation during listen and
// http.ErrServerClosed as success. Shutdown errors are not returned.
func (s *WebServer) Serve() error {
	engine := s.newEngine()
	addr := fmt.Sprintf("%s:%d", s.address, s.port)
	s.log.Info("starting server", zap.String("address", addr))
	listener, err := new(net.ListenConfig).Listen(s.ctx, "tcp", addr)
	if err != nil {
		if s.ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("failed to listen for web requests: %w", err)
	}
	startErr := make(chan error, 1)
	go func() {
		startErr <- engine.Server.Serve(listener)
	}()
	select {
	case err := <-startErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("webserver stopped: %w", err)
		}
		return nil
	case <-s.ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := engine.Shutdown(shutdownCtx); err != nil {
			s.log.Warn("failed to shut down webserver cleanly", zap.Error(err))
		}
		if err := <-startErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("webserver stopped with an error: %w", err)
		}
		return nil
	}
}
