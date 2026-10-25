// Package endpoint implements the logic behind each endpoint of the webserver
package endpoint

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/core/global"
)

const maxConcurrentEventRequests = 64

type EventDispatchEndpoint struct {
	dispatches atomic.Int32
}

func NewEventDispatchEndpoint() *EventDispatchEndpoint {
	return &EventDispatchEndpoint{}
}

// Endpoint invokes listeners for the route's event synchronously, passing query
// parameters as slices of strings. It responds with 404 when no listeners exist
// or 200 after all listeners return, and returns any response write error.
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
	if ed.dispatches.Add(1) > maxConcurrentEventRequests {
		ed.dispatches.Add(-1)
		return c.String(http.StatusTooManyRequests, "too many event dispatch requests")
	}
	listenerCount := len(listeners)
	var pending sync.WaitGroup
	pending.Add(listenerCount)
	for _, listener := range listeners {
		go func() {
			defer pending.Done()
			defer func() {
				if r := recover(); r != nil {
					global.Logger("event_dispatch").Error("event listener panicked", zap.Any("recover", r))
				}
			}()
			listener(metaData)
		}()
	}
	go func() {
		pending.Wait()
		ed.dispatches.Add(-1)
	}()
	return c.String(http.StatusOK, fmt.Sprintf("event: '%s' emitted, %d listeners where found", e, listenerCount))
}
