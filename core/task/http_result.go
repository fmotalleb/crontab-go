package task

import (
	"fmt"
	"net/http"
)

func validateHTTPResult(resultErr error, res *http.Response) error {
	if resultErr != nil {
		return resultErr
	}
	if res != nil && res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected http status code: %d", res.StatusCode)
	}
	return nil
}
