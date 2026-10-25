package task

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.uber.org/zap"
)

func TestNewHTTPClient_DefaultVerifiesTLS(t *testing.T) {
	c := newHTTPClient(false)
	tr, ok := c.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.True(t, tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify)
}

func TestNewHTTPClient_InsecureSkipsVerify(t *testing.T) {
	c := newHTTPClient(true)
	tr, ok := c.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.True(t, tr.TLSClientConfig != nil)
	assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
}

func TestRedactedURLOmitsCredentialsAndPath(t *testing.T) {
	assert.Equal(t, "https://example.com:8443", redactedURL("https://user:secret@example.com:8443/path/token?access=secret"))
	assert.Equal(t, "<invalid-url>", redactedURL("http://["))
}

func TestDoHTTP_StatusError(t *testing.T) {
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.test/", nil)
	assert.NoError(t, err)
	headers := map[string]string{}
	client := &http.Client{Transport: statusTransport{}}
	_, err = doHTTP(client, req, &headers, io.Discard, zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

type statusTransport struct{}

func (statusTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Status:     "500 Internal Server Error",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("oops\n")),
	}, nil
}

type failingBodyReader struct{}

func (failingBodyReader) Read([]byte) (int, error) { return 0, errors.New("stream interrupted") }

type failingBodyTransport struct{}

func (failingBodyTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(failingBodyReader{}),
	}, nil
}

func TestDoHTTP_FailingResponseBody(t *testing.T) {
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.test/", nil)
	assert.NoError(t, err)
	headers := map[string]string{}
	client := &http.Client{Transport: failingBodyTransport{}}
	_, err = doHTTP(client, req, &headers, io.Discard, zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream interrupted")
}
