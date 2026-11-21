package global

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestEventListenersReturnsSnapshot(t *testing.T) {
	const eventName = "test-listener-snapshot"
	CTX().AddEventListener(eventName, func(map[string]any) {})

	snapshot := CTX().EventListeners()
	delete(snapshot, eventName)
	if len(CTX().EventListeners()[eventName]) != 1 {
		t.Fatal("mutating the returned listener map changed registered listeners")
	}
}

func TestAttachOTelCorePreservesLogConfig(t *testing.T) {
	output := filepath.Join(t.TempDir(), "app.log")
	t.Setenv("ZAPLOG_LEVEL", "warn")
	t.Setenv("ZAPLOG_OUTPUT_PATHS", output)
	t.Setenv("ZAPLOG_DEVELOPMENT", "false")
	t.Setenv("ZAPLOG_TIME_FORMAT", "epoch")
	t.Setenv("ZAPLOG_LEVEL_FORMAT", "capital")
	t.Setenv("ZAPLOG_DISABLE_CALLER", "true")

	ctx, err := log.WithNewEnvLogger(t.Context())
	assert.NoError(t, err)
	previous := c
	state := &Context{Context: ctx, mu: new(sync.RWMutex)}
	c = func() *Context { return state }
	t.Cleanup(func() { c = previous })

	telemetry, exported := observer.New(zap.DebugLevel)
	AttachOTelCore(telemetry)
	logger := Logger("worker").With(zap.String("job", "backup"))
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")

	assert.Equal(t, zap.WarnLevel, logger.Level())
	assert.Equal(t, 1, exported.Len())
	entry := exported.All()[0]
	assert.Equal(t, "warning message", entry.Message)
	assert.Equal(t, "worker", entry.LoggerName)
	assert.Equal(t, "backup", entry.ContextMap()["job"])

	data, err := os.ReadFile(output)
	assert.NoError(t, err)
	var record map[string]any
	assert.NoError(t, json.Unmarshal(data, &record))
	assert.Equal(t, "warning message", record["message"])
	assert.Equal(t, "worker", record["logger"])
	assert.Equal(t, "WARN", record["level"])
	assert.Equal(t, "backup", record["job"])
	_, epochTimestamp := record["timestamp"].(float64)
	assert.True(t, epochTimestamp)
	_, caller := record["caller"]
	assert.False(t, caller)
}
