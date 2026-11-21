package endpoint

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/labstack/echo/v4"

	"github.com/fmotalleb/crontab-go/core/global"
)

func dispatchRequest(t *testing.T, ed *EventDispatchEndpoint, event string) int {
	t.Helper()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/?value=one&value=two", nil)
	response := httptest.NewRecorder()
	c := echo.New().NewContext(request, response)
	c.SetParamNames("event")
	c.SetParamValues(event)
	assert.NoError(t, ed.Endpoint(c))
	return response.Code
}

func TestEventDispatchSchedulesListenersIndependently(t *testing.T) {
	global.CTX() // Initialize signal handling outside the synctest bubble.
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		defer close(release)
		invoked := make(chan map[string]any, 1)
		global.CTX().AddEventListener(t.Name(), func(map[string]any) { <-release })
		global.CTX().AddEventListener(t.Name(), func(data map[string]any) { invoked <- data })
		ed := NewEventDispatchEndpoint()
		response := make(chan int, 1)
		go func() { response <- dispatchRequest(t, ed, t.Name()) }()
		select {
		case status := <-response:
			assert.Equal(t, http.StatusOK, status)
		case <-time.After(time.Second):
			t.Fatal("response blocked behind a listener")
		}
		select {
		case data := <-invoked:
			assert.Equal(t, []string{"one", "two"}, data["value"].([]string))
		case <-time.After(time.Second):
			t.Fatal("later listener blocked behind the first listener")
		}
		synctest.Wait()
		assert.Equal(t, int32(1), ed.dispatches.Load())
	})
}

func TestEventDispatchHoldsCapacityUntilListenersReturn(t *testing.T) {
	global.CTX()
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		global.CTX().AddEventListener(t.Name(), func(map[string]any) { <-release })
		ed := NewEventDispatchEndpoint()
		for range maxConcurrentEventRequests {
			assert.Equal(t, http.StatusOK, dispatchRequest(t, ed, t.Name()))
		}
		assert.Equal(t, http.StatusTooManyRequests, dispatchRequest(t, ed, t.Name()))
		close(release)
		synctest.Wait()
		assert.Equal(t, int32(0), ed.dispatches.Load())
		assert.Equal(t, http.StatusOK, dispatchRequest(t, ed, t.Name()))
		synctest.Wait()
		assert.Equal(t, int32(0), ed.dispatches.Load())
	})
}

func TestEventDispatchRecoversListenerPanic(t *testing.T) {
	global.CTX()
	synctest.Test(t, func(t *testing.T) {
		global.CTX().AddEventListener(t.Name(), func(map[string]any) { panic("listener failure") })
		ed := NewEventDispatchEndpoint()
		assert.Equal(t, http.StatusOK, dispatchRequest(t, ed, t.Name()))
		synctest.Wait()
		assert.Equal(t, int32(0), ed.dispatches.Load())
	})
}
