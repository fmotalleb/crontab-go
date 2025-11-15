package common

// import (
// 	"testing"
// 	"time"

// 	"github.com/alecthomas/assert/v2"
// )

// func TestSetMaxRetry(t *testing.T) {
// 	r := &Retry{}
// 	r.SetMaxRetry(5)
// 	assert.Equal(t, 5, r.maxRetries)
// }

// func TestSetRetryDelay(t *testing.T) {
// 	r := &Retry{}
// 	delay := 2 * time.Second
// 	r.SetRetryDelay(delay)
// 	assert.Equal(t, r.retryDelay, delay)
// }

// func TestWaitForRetry(t *testing.T) {
// 	r := &Retry{maxRetries: 3, retryDelay: 1 * time.Second}
// 	ctx := WithRetryCount(t.Context(), 2)

// 	start := time.Now()
// 	err := r.WaitForRetry(ctx)
// 	elapsed := time.Since(start)
// 	assert.NoError(t, err)
// 	assert.True(t, elapsed > 2*time.Second)
// }

// func TestWaitForRetryMaxExceeded(t *testing.T) {
// 	r := &Retry{maxRetries: 3, retryDelay: 1 * time.Second}
// 	ctx := WithRetryCount(t.Context(), 5)

// 	err := r.WaitForRetry(ctx)
// 	assert.Error(t, err)
// }

// func TestIncreaseRetry(t *testing.T) {
// 	ctx := WithRetryCount(t.Context(), 2)
// 	newCtx := IncreaseRetry(ctx)
// 	assert.Equal(t, 3, GetRetry(newCtx))
// }

// func TestZeroValueRetry(t *testing.T) {
// 	ctx := t.Context()
// 	assert.Equal(t, 0, GetRetry(ctx))
// }
