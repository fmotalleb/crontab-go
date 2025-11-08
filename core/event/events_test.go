package event_test

// import (
// 	"errors"
// 	"testing"

// 	"github.com/alecthomas/assert/v2"
// 	"github.com/robfig/cron/v3"
// 	"github.com/sirupsen/logrus"

// 	"github.com/FMotalleb/crontab-go/config"
// 	"github.com/FMotalleb/crontab-go/core/event"
// 	"github.com/FMotalleb/crontab-go/core/global"
// 	mocklogger "github.com/FMotalleb/crontab-go/logger/mock_logger"
// )

// func prepareState() {
// 	cr := cron.New()
// 	global.Put(cr)
// }

// // TestCompileEvent_IntervalZero tests that CompileEvents returns nil when Interval is zero.
// func TestCompileEvent_IntervalZero(t *testing.T) {
// 	sh := &config.JobEvent{Interval: 0}
// 	prepareState()
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	event := event.Build(log, sh)
// 	assert.Equal(t, event, nil)
// }

// func TestCompileEvent_IntervalNonZero(t *testing.T) {
// 	sh := &config.JobEvent{Interval: 15}
// 	prepareState()
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)
// 	sch := event.Build(log, sh)
// 	_, ok := sch.(*event.Interval)
// 	assert.Equal(t, ok, true)
// }

// // TestCompileEvent_IntervalZeroWithCronSet tests CompileEvents with Interval zero but Cron expression set.
// func TestCompileEvent_IntervalZeroWithCronSet(t *testing.T) {
// 	sh := &config.JobEvent{Cron: "0 * * * *", Interval: 0}
// 	prepareState()
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)

// 	e := event.Build(log, sh)
// 	if _, ok := e.(*event.Cron); !ok {
// 		t.Errorf("Expected Cron events, got %T", e)
// 	}
// }

// // TestCompileEvent_IntervalZeroWithOnInitSet tests CompileEvents with Interval zero and OnInit set.
// func TestCompileEvent_IntervalZeroWithOnInitSet(t *testing.T) {
// 	sh := &config.JobEvent{OnInit: true, Interval: 0}
// 	prepareState()
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)

// 	e := event.Build(log, sh)
// 	if _, ok := e.(*event.Init); !ok {
// 		t.Errorf("Expected Init events, got %T", e)
// 	}
// }

// // TestCompileEvent_IntervalZeroWithAllFieldsEmpty tests CompileEvents with Interval zero and all other fields empty.
// func TestCompileEvent_IntervalZeroWithAllFieldsEmpty(t *testing.T) {
// 	sh := &config.JobEvent{Interval: 0}
// 	prepareState()
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)

// 	e := event.Build(log, sh)
// 	if e != nil {
// 		t.Errorf("Expected nil, got %v", e)
// 	}
// }

// // TestCompileEvent_IntervalZeroWithCronAndOnInitSet tests CompileEvent with Interval zero, Cron expression, and OnInit set.
// func TestCompileEvent_IntervalZeroWithCronAndOnInitSet(t *testing.T) {
// 	sh := &config.JobEvent{Cron: "0 * * * *", OnInit: true, Interval: 0}
// 	prepareState()
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)

// 	e := event.Build(log, sh)
// 	if _, ok := e.(*event.Cron); !ok {
// 		t.Errorf("Expected Cron event, got %T", e)
// 	}
// }

// func TestNewMetaData(t *testing.T) {
// 	extra := map[string]any{"key1": "value1"}
// 	m := event.NewMetaData("emitter1", extra)
// 	assert.Equal(t, "emitter1", m.Emitter)
// 	assert.Equal(t, extra, m.Extra)
// }

// func TestNewErrMetaData(t *testing.T) {
// 	err := errors.New("something went wrong")
// 	m := event.NewErrMetaData("emitter2", err)
// 	assert.Equal(t, "emitter2", m.Emitter)
// 	expectedExtra := map[string]any{"error": "something went wrong"}
// 	assert.Equal(t, expectedExtra, m.Extra)
// }

// func TestGetData(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		emitter  string
// 		extra    map[string]any
// 		expected map[string]any
// 	}{
// 		{
// 			name:    "nil extra map",
// 			emitter: "e1",
// 			extra:   nil,
// 			expected: map[string]any{
// 				"emitter": "e1",
// 			},
// 		},
// 		{
// 			name:    "empty extra map",
// 			emitter: "e2",
// 			extra:   map[string]any{},
// 			expected: map[string]any{
// 				"emitter": "e2",
// 			},
// 		},
// 		{
// 			name:    "non-empty extra map",
// 			emitter: "e3",
// 			extra: map[string]any{
// 				"foo": "bar",
// 			},
// 			expected: map[string]any{
// 				"foo":     "bar",
// 				"emitter": "e3",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			m := &event.MetaData{
// 				Emitter: tt.emitter,
// 				Extra:   tt.extra,
// 			}
// 			data := m.GetData()
// 			assert.Equal(t, tt.expected, data)
// 		})
// 	}
// }
