package config

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestCrontabCalendar(t *testing.T) {
	tests := []struct {
		spec  string
		after string
		next  string
	}{
		{"0 12 * JAN-MAR,Nov 1-7/2", "2026-02-07T12:00:00Z", "2026-02-08T12:00:00Z"},
		{"0 0 1 * 7", "2026-02-02T00:00:00Z", "2026-02-08T00:00:00Z"},
		{"0 0 */2 * MON", "2026-02-02T00:00:00Z", "2026-02-09T00:00:00Z"},
		{"@midnight", "2026-02-01T20:00:00Z", "2026-02-02T00:00:00Z"},
	}
	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			e := JobEvent{Cron: tt.spec, Crontab: &CrontabTiming{TimeZone: "UTC"}}
			schedule, err := e.CronSchedule()
			assert.NoError(t, err)
			after, err := time.Parse(time.RFC3339, tt.after)
			assert.NoError(t, err)
			assert.Equal(t, tt.next, schedule.Next(after).Format(time.RFC3339))
		})
	}
}

func TestCrontabTimeZoneAndFixedRandomTiming(t *testing.T) {
	e := JobEvent{
		Cron:    "6~15 8 * * *",
		Crontab: &CrontabTiming{TimeZone: "America/New_York", RandomDelay: 20},
	}
	picks := 0
	schedule, err := e.parseCron(func(limit int) int {
		picks++
		assert.Equal(t, 10, limit)
		return limit - 1
	}, 0.5)
	assert.NoError(t, err)
	after := time.Date(2026, time.February, 1, 13, 22, 0, 0, time.UTC)
	first := schedule.Next(after)
	assert.Equal(t, "2026-02-01T13:25:00Z", first.Format(time.RFC3339))
	assert.Equal(t, "2026-02-02T13:25:00Z", schedule.Next(first).Format(time.RFC3339))
	assert.Equal(t, 1, picks)
}

func TestCrontabRejectsInvalidFields(t *testing.T) {
	for _, spec := range []string{"~, * * * *", "20~10 * * * *", "*/0 * * * *", "* * * * 8"} {
		t.Run(spec, func(t *testing.T) {
			e := JobEvent{Cron: spec, Crontab: &CrontabTiming{}}
			assert.Error(t, e.ValidateCron())
		})
	}
}
