// Package endpoint implements the logic behind each endpoint of the webserver
package endpoint

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/fmotalleb/crontab-go/core/global"
)

type EventDispatchEndpoint struct{}

func NewEventDispatchEndpoint() *EventDispatchEndpoint {
	return &EventDispatchEndpoint{}
}

func (ed *EventDispatchEndpoint) Endpoint(c *gin.Context) {
	event := c.Param("event")

	metaData := make(map[string]any)
	for key, values := range c.Request.URL.Query() {
		metaData[key] = values
	}

	listeners := global.CTX().EventListeners()[event]
	if len(listeners) == 0 {
		c.String(http.StatusNotFound, fmt.Sprintf("event: '%s' not found", event))
		return
	}
	global.CTX().MetricCounter(
		c,
		"webserver_events",
		"amount of events dispatched using webserver",
		prometheus.Labels{"event_name": event},
	).Operate(
		func(f float64) float64 {
			return f + 1
		},
	)
	listenerCount := len(listeners)
	global.CTX().MetricCounter(c, "webserver_event_listeners_invoked", "amount of listeners invoked by `event_name` event", prometheus.Labels{"event_name": event}).Operate(
		func(f float64) float64 {
			return f + float64(listenerCount)
		},
	)
	for _, listener := range listeners {
		go listener(metaData)
	}
	c.String(http.StatusOK, fmt.Sprintf("event: '%s' emitted, %d listeners where found", event, listenerCount))
}
