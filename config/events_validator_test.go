package config_test

// import (
// 	"testing"

// 	"github.com/alecthomas/assert/v2"

// 	"github.com/FMotalleb/crontab-go/config"
// 	mocklogger "github.com/FMotalleb/crontab-go/logger/mock_logger"
// )

// func TestJobEvent_Validate_WebEvent(t *testing.T) {
// 	event := config.JobEvent{
// 		Interval: 0,
// 		Cron:     "",
// 		OnInit:   false,
// 		WebEvent: "test-event",
// 	}
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)

// 	err := event.Validate(log)

// 	assert.NoError(t, err)
// }

// func TestJobEvent_Validate_PositiveInterval(t *testing.T) {
// 	event := config.JobEvent{
// 		Interval: 10,
// 		Cron:     "",
// 		OnInit:   false,
// 		WebEvent: "",
// 	}
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)

// 	err := event.Validate(log)

// 	assert.NoError(t, err)
// }

// func TestJobEvent_Validate_CorrectCron(t *testing.T) {
// 	event := config.JobEvent{
// 		Interval: 0,
// 		Cron:     "* * * * *",
// 		OnInit:   false,
// 	}
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)

// 	err := event.Validate(log)
// 	assert.NoError(t, err)
// }

// func TestJobEvent_Validate_NegativeInterval(t *testing.T) {
// 	event := config.JobEvent{
// 		Interval: -10,
// 		Cron:     "",
// 		OnInit:   false,
// 	}
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)

// 	err := event.Validate(log)

// 	expectedErr := "received a negative time in interval: `-10ns`"

// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), expectedErr)
// }

// func TestJobEvent_Validate_InvalidCronExpression(t *testing.T) {
// 	event := config.JobEvent{
// 		Interval: 0,
// 		Cron:     "invalid_cron_expression",
// 		OnInit:   false,
// 	}
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)

// 	err := event.Validate(log)

// 	assert.Error(t, err)
// }

// func TestJobEvent_Validate_MultipleActiveSchedules(t *testing.T) {
// 	event := config.JobEvent{
// 		Interval: 60,
// 		Cron:     "0 0 * * *",
// 		OnInit:   true,
// 	}
// 	logger, _ := mocklogger.HijackOutput(logrus.New())
// 	log := logrus.NewEntry(logger)

// 	err := event.Validate(log)

// 	expectedErr := "a single event must have one of "

// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), expectedErr)
// }
