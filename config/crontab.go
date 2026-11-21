package config

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // Make CRON_TZ work in containers without system zoneinfo.

	"github.com/robfig/cron/v3"
)

// MaxCrontabRandomDelay is the Cronie limit in minutes.
const MaxCrontabRandomDelay = 24 * 60

// Cronie chooses one scale for all RANDOM_DELAY limits for the life of the daemon.
var crontabDelayScale = rand.Float64()

// CrontabTiming configures Cronie scheduling semantics for a cron event.
type CrontabTiming struct {
	TimeZone    string `mapstructure:"time-zone" json:"time-zone,omitempty"`
	RandomDelay int    `mapstructure:"random-delay" json:"random-delay,omitempty"`
}

// CrontabCommand stores the literal input for a local crontab command.
type CrontabCommand struct {
	Environment []string `mapstructure:"environment" json:"environment,omitempty"`
	Stdin       string   `mapstructure:"stdin" json:"stdin,omitempty"`
	Quiet       bool     `mapstructure:"quiet" json:"quiet,omitempty"`
}

// ValidateCron checks the schedule without consuming randomness.
func (e JobEvent) ValidateCron() error {
	_, err := e.parseCron(func(int) int { return 0 }, 0)
	return err
}

// CronSchedule builds the schedule and fixes the random field choices for this entry.
func (e JobEvent) CronSchedule() (cron.Schedule, error) {
	return e.parseCron(rand.IntN, crontabDelayScale)
}

func (e JobEvent) parseCron(pick func(int) int, delayScale float64) (cron.Schedule, error) {
	if e.Crontab == nil {
		return DefaultCronParser.Parse(e.Cron)
	}
	if e.Crontab.RandomDelay < 0 || e.Crontab.RandomDelay > MaxCrontabRandomDelay {
		return nil, fmt.Errorf("RANDOM_DELAY must be between 0 and %d minutes", MaxCrontabRandomDelay)
	}
	location, err := loadCrontabLocation(e.Crontab.TimeZone)
	if err != nil {
		return nil, err
	}
	fields := strings.Fields(e.Cron)
	if strings.HasPrefix(e.Cron, "@every ") {
		return DefaultCronParser.Parse(e.Cron)
	}
	schedule, err := parseCrontabCalendar(fields, location, pick)
	if err != nil {
		return nil, err
	}
	delay := time.Duration(float64(e.Crontab.RandomDelay)*delayScale) * time.Minute
	if delay > 0 {
		return delayedSchedule{Schedule: schedule, delay: delay}, nil
	}
	return schedule, nil
}

func loadCrontabLocation(zone string) (*time.Location, error) {
	location, err := time.LoadLocation(zone)
	if err != nil {
		return nil, fmt.Errorf("invalid CRON_TZ %q: %w", zone, err)
	}
	if zone == "" {
		return time.Local, nil
	}
	return location, nil
}

type delayedSchedule struct {
	cron.Schedule
	delay time.Duration
}

func (s delayedSchedule) Next(t time.Time) time.Time {
	next := s.Schedule.Next(t.Add(-s.delay))
	if next.IsZero() {
		return next
	}
	return next.Add(s.delay)
}

func parseCrontabCalendar(fields []string, location *time.Location, pick func(int) int) (cron.Schedule, error) {
	if len(fields) == 1 && strings.HasPrefix(fields[0], "@") {
		return DefaultCronParser.Parse("CRON_TZ=" + location.String() + " " + fields[0])
	}
	// An explicit converter matcher can select a schedule with seconds.
	if len(fields) == 5 {
		fields = append([]string{"0"}, fields...)
	}
	if len(fields) != 6 {
		return nil, errors.New("expected five schedule fields or a calendar alias")
	}

	var schedule cron.SpecSchedule
	schedule.Location = location
	bits := []*uint64{&schedule.Second, &schedule.Minute, &schedule.Hour, &schedule.Dom, &schedule.Month, &schedule.Dow}
	for i, field := range fields {
		bounds := crontabBounds[i]
		values, err := bounds.parse(field, pick)
		if err != nil {
			return nil, fmt.Errorf("invalid %s field %q: %w", bounds.label, field, err)
		}
		*bits[i] = values
	}
	if schedule.Dow&(1<<7) != 0 {
		schedule.Dow |= 1 // Both 0 and 7 are Sunday in crontab.
	}
	// robfig uses the high bit of these masks to switch from OR to AND.
	const starBit = uint64(1) << 63
	if strings.HasPrefix(fields[3], "*") {
		schedule.Dom |= starBit
	}
	if strings.HasPrefix(fields[5], "*") {
		schedule.Dow |= starBit
	}
	return &schedule, nil
}

type crontabField struct {
	label    string
	min, max int
	names    string // Space-separated, starting at min.
}

var crontabBounds = [...]crontabField{
	{label: "second", min: 0, max: 59},
	{label: "minute", min: 0, max: 59},
	{label: "hour", min: 0, max: 23},
	{label: "day of month", min: 1, max: 31},
	{label: "month", min: 1, max: 12, names: "jan feb mar apr may jun jul aug sep oct nov dec"},
	{label: "day of week", min: 0, max: 7, names: "sun mon tue wed thu fri sat"},
}

func (f crontabField) parse(field string, pick func(int) int) (uint64, error) {
	var bits uint64
	for item := range strings.SplitSeq(field, ",") {
		first, last, step, err := f.parseRange(item, pick)
		if err != nil {
			return 0, err
		}
		for value := first; value <= last; value += step {
			bits |= 1 << value
		}
	}
	return bits, nil
}

func (f crontabField) parseRange(item string, pick func(int) int) (first, last, step int, err error) {
	expr, stepText, hasStep := strings.Cut(item, "/")
	step = 1
	if hasStep {
		step, err = strconv.Atoi(stepText)
		if err != nil || step <= 0 {
			return 0, 0, 0, fmt.Errorf("invalid step %q", stepText)
		}
		// A step larger than the field's span selects only the first value.
		step = min(step, f.max-f.min+1)
	}
	if expr == "*" {
		return f.min, f.max, step, nil
	}
	if low, high, random := strings.Cut(expr, "~"); random {
		if hasStep {
			return 0, 0, 0, errors.New("a random range cannot have a step")
		}
		first, last, err = f.bounds(low, high, true)
		if err != nil {
			return 0, 0, 0, err
		}
		value := first + pick(last-first+1)
		return value, value, 1, nil
	}
	if low, high, found := strings.Cut(expr, "-"); found {
		first, last, err = f.bounds(low, high, false)
		return first, last, step, err
	}
	first, err = f.value(expr)
	last = first
	if hasStep {
		last = f.max
	}
	return first, last, step, err
}

func (f crontabField) bounds(low, high string, optional bool) (int, int, error) {
	first, last := f.min, f.max
	var err error
	if low != "" || !optional {
		first, err = f.value(low)
		if err != nil {
			return 0, 0, err
		}
	}
	if high != "" || !optional {
		last, err = f.value(high)
		if err != nil {
			return 0, 0, err
		}
	}
	if first > last {
		return 0, 0, fmt.Errorf("range start %d exceeds end %d", first, last)
	}
	return first, last, nil
}

func (f crontabField) value(text string) (int, error) {
	for i, name := range strings.Fields(f.names) {
		if strings.EqualFold(text, name) {
			return f.min + i, nil
		}
	}
	value, err := strconv.Atoi(text)
	if err != nil || value < f.min || value > f.max {
		return 0, fmt.Errorf("%q must be within %d-%d or a supported name", text, f.min, f.max)
	}
	return value, nil
}
