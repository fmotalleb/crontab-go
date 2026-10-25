package task

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/helpers"
)

var (
	httpClientsOnce sync.Once
	verifiedClient  *http.Client
	insecureClient  *http.Client
)

// newHTTPClient returns a shared client with pooled connections.
// insecure selects the client that skips TLS certificate verification.
func newHTTPClient(insecure bool) *http.Client {
	httpClientsOnce.Do(func() {
		verifiedTransport := http.DefaultTransport.(*http.Transport).Clone()
		insecureTransport := verifiedTransport.Clone()
		insecureTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // explicit user opt-in
		verifiedClient = &http.Client{Transport: verifiedTransport}
		insecureClient = &http.Client{Transport: insecureTransport}
	})
	if insecure {
		return insecureClient
	}
	return verifiedClient
}

// CloseHTTPIdleConnections releases pooled connections during application shutdown.
func CloseHTTPIdleConnections() {
	if verifiedClient != nil {
		verifiedClient.CloseIdleConnections()
	}
	if insecureClient != nil {
		insecureClient.CloseIdleConnections()
	}
}

// redactedURL returns only the URL scheme and host, including any port.
// It returns "<invalid-url>" if parsing fails or the scheme or host is missing.
func redactedURL(address string) string {
	parsed, err := url.Parse(address)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "<invalid-url>"
	}
	return parsed.Scheme + "://" + parsed.Host
}

// doHTTP sends the request, streams the response body to out, and returns the HTTP status code and any error.
func doHTTP(client *http.Client, req *http.Request, headers *map[string]string, out io.Writer, log *zap.Logger) (int, error) {
	for key, val := range *headers {
		req.Header.Add(key, val)
	}
	log.Debug("sending http request")
	res, err := client.Do(req)
	if res != nil {
		if res.Body != nil {
			defer helpers.WarnOnErrIgnored(
				log,
				res.Body.Close,
				"cannot close response body",
			)
		}
		log.Info("received response with status", zap.Int("status", res.StatusCode))
		if out != nil {
			if _, cpErr := io.Copy(out, res.Body); cpErr != nil {
				log.Warn("failed to stream response body", zap.Error(cpErr))
				return res.StatusCode, fmt.Errorf("failed to stream response body: %w", cpErr)
			}
		}
	}
	if err != nil {
		log.Warn("request failed", zap.Error(err), zap.Int("status", statusCodeOf(res)))
		return statusCodeOf(res), err
	}
	if res != nil && res.StatusCode >= http.StatusBadRequest {
		statusErr := fmt.Errorf("request failed with status %s", res.Status)
		log.Warn("request failed", zap.Error(statusErr), zap.Int("status", res.StatusCode))
		return res.StatusCode, statusErr
	}
	return statusCodeOf(res), nil
}

func statusCodeOf(res *http.Response) int {
	if res == nil {
		return 0
	}
	return res.StatusCode
}
