// Package endpoint implements the logic behind each endpoint of the webserver
package endpoint

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/fmotalleb/crontab-go/core/global"
)

type EventDispatchEndpoint struct{}

func NewEventDispatchEndpoint() *EventDispatchEndpoint {
	return &EventDispatchEndpoint{}
}

func (ed *EventDispatchEndpoint) Endpoint(c echo.Context) error {
	e := c.Param("event")

	metaData := make(map[string]any)
	for key, values := range c.Request().URL.Query() {
		metaData[key] = values
	}

	listeners := global.CTX().EventListeners()[e]
	if len(listeners) == 0 {
		return c.String(http.StatusNotFound, fmt.Sprintf("event: '%s' not found", e))
	}
	listenerCount := len(listeners)
	for _, listener := range listeners {
		go listener(metaData)
	}
	return c.String(http.StatusOK, fmt.Sprintf("event: '%s' emitted, %d listeners where found", e, listenerCount))
}
