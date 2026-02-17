package task

import (
	"errors"
	"net/http"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestValidateHTTPResult(t *testing.T) {
	t.Run("returns original transport error", func(t *testing.T) {
		expErr := errors.New("network failed")
		err := validateHTTPResult(expErr, nil)
		assert.Equal(t, expErr, err)
	})

	t.Run("fails for bad status code", func(t *testing.T) {
		err := validateHTTPResult(nil, &http.Response{StatusCode: http.StatusInternalServerError})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected http status code: 500")
	})

	t.Run("passes for successful status", func(t *testing.T) {
		err := validateHTTPResult(nil, &http.Response{StatusCode: http.StatusOK})
		assert.NoError(t, err)
	})

	t.Run("passes when response is nil", func(t *testing.T) {
		err := validateHTTPResult(nil, nil)
		assert.NoError(t, err)
	})
}
