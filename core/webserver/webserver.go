// Package webserver implements the logic for the webserver
package webserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fmotalleb/go-tools/log"
	"github.com/gin-gonic/gin"
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
	engine := gin.New()
	auth := func(*gin.Context) {}
	if s.AuthConfig != nil && s.AuthConfig.Username != "" && s.AuthConfig.Password != "" {
		auth = gin.BasicAuth(gin.Accounts{s.AuthConfig.Username: s.AuthConfig.Password})
	} else {
		s.log.Warn("received no value on username or password, ignoring any authentication, if you intended to use no authentication ignore this message")
	}
	// log := gin.LoggerWithConfig(gin.LoggerConfig{
	// 	Formatter: gin.format,
	// })
	engine.Use(
		auth,
		// log,
		gin.Recovery(),
	)

	engine.GET(
		"/foo",
		func(c *gin.Context) {
			c.String(200, "bar")
		},
	)

	ed := &endpoint.EventDispatchEndpoint{}
	engine.Any(
		"/events/:event/emit",
		ed.Endpoint,
	)
	if s.serveMetrics {
		engine.GET("/metrics", func(ctx *gin.Context) {
			promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
		})
	} else {
		engine.GET("/metrics", func(ctx *gin.Context) {
			ctx.String(http.StatusNotFound, "Metrics are disabled, please enable metrics using `WEBSERVER_METRICS=true`")
		})
	}

	err := engine.Run(fmt.Sprintf("%s:%d", s.address, s.port))
	helpers.FatalOnErr(
		s.log,
		func() error {
			return err
		},
		"Failed to start webserver: %s",
	)
}
