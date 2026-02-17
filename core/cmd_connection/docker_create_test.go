package connection

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestParseContainerVolumes(t *testing.T) {
	t.Run("extracts in-container targets", func(t *testing.T) {
		got, err := parseContainerVolumes([]string{
			"/host/a:/container/a",
			"/host/b:/container/b:ro",
		})
		assert.NoError(t, err)
		assert.Equal(t, map[string]struct{}{
			"/container/a": {},
			"/container/b": {},
		}, got)
	})

	t.Run("returns error on malformed volume", func(t *testing.T) {
		_, err := parseContainerVolumes([]string{"/host-only"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid docker volume format")
	})
}

func TestRetryUntilContext(t *testing.T) {
	t.Run("returns nil when operation eventually succeeds", func(t *testing.T) {
		attempts := 0
		err := retryUntilContext(context.Background(), time.Millisecond, func() error {
			attempts++
			if attempts < 3 {
				return errors.New("try again")
			}
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("returns context cancellation when operation keeps failing", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()
		err := retryUntilContext(ctx, time.Millisecond, func() error {
			return errors.New("still failing")
		})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.DeadlineExceeded))
	})
}
