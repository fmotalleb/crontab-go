package webserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestWebServerRequiresCredentialsAndOnlyAcceptsPOST(t *testing.T) {
	server := NewWebServer(context.Background(), "127.0.0.1", 8080, false, &AuthConfig{
		Username: "scheduler",
		Password: "secret-password",
	})
	engine := server.newEngine()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/events/missing/emit", nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	request = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events/missing/emit", nil)
	request.SetBasicAuth("scheduler", "secret-password")
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	assert.Equal(t, http.StatusMethodNotAllowed, response.Code)

	request = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/events/missing/emit", nil)
	request.SetBasicAuth("scheduler", "secret-password")
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestWebServerUnauthenticatedAccessRequiresExplicitOptIn(t *testing.T) {
	server := NewWebServer(context.Background(), "127.0.0.1", 8080, false, &AuthConfig{
		AllowUnauthenticated: true,
	})
	engine := server.newEngine()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/events/missing/emit", nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)
}
