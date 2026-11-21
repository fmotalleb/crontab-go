package event

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/maniartech/signals"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
)

func TestLogFileStartupAndRetry(t *testing.T) {
	for _, existing := range []bool{true, false} {
		name := "created after startup"
		if existing {
			name = "existing at startup"
		}
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "events.log")
			if existing {
				assert.NoError(t, os.WriteFile(path, []byte("old\n"), 0o600))
			}
			lf, err := NewLogFile(path, "", time.Second, zap.NewNop())
			assert.NoError(t, err)
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				var lines []string
				dispatcher := signals.NewSync[abstraction.Event]()
				dispatcher.AddListener(func(_ context.Context, e abstraction.Event) {
					lines = append(lines, e.GetData()["line"].(string))
				})
				go lf.buildTickChannel(ctx, dispatcher)
				synctest.Wait()
				if !existing {
					// Let a missing-file retry pass before creating the file.
					time.Sleep(time.Second)
					synctest.Wait()
				}
				file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
				assert.NoError(t, err)
				_, err = file.WriteString("first\nsecond\n")
				assert.NoError(t, err)
				assert.NoError(t, file.Close())
				time.Sleep(time.Second)
				synctest.Wait()
				assert.Equal(t, []string{"first", "second"}, lines)
				cancel()
				synctest.Wait()
			})
		})
	}
}
