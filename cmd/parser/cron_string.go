package parser

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/fmotalleb/crontab-go/config"
)

type CronString struct {
	string
}

func NewCronString(cron string) CronString {
	return CronString{cron}
}

// ParseFile loads a legacy crontab and adds its source path to parse errors.
func ParseFile(path string, matcher string, hasUser bool) (*config.Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read crontab: %w", err)
	}
	parsed := NewCronString(string(content))
	cfg, err := parsed.ParseConfig(matcher, hasUser)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}

type crontabLineParser struct {
	matcher *regexp.Regexp
	hasUser bool
	env     map[string]string
	delay   int
}

// ParseConfig reads the entries in order, using the environment at each line.
// A nonempty pattern overrides how the schedule is extracted from the line.
func (s CronString) ParseConfig(pattern string, hasUser bool) (*config.Config, error) {
	parser := crontabLineParser{hasUser: hasUser, env: make(map[string]string)}
	if pattern != "" {
		matcher, err := regexp.Compile(`^(` + pattern + `)[ \t]+(.*)$`)
		if err != nil {
			return nil, fmt.Errorf("invalid schedule matcher: %w", err)
		}
		parser.matcher = matcher
	}
	cfg := &config.Config{Jobs: []*config.JobConfig{}}
	for index, raw := range strings.Split(s.string, "\n") {
		job, err := parser.parseLine(strings.TrimSuffix(raw, "\r"), index+1)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", index+1, err)
		}
		if job != nil {
			cfg.Jobs = append(cfg.Jobs, job)
		}
	}
	return cfg, nil
}

func (p *crontabLineParser) parseLine(line string, lineNumber int) (*config.JobConfig, error) {
	line = strings.TrimLeft(line, " \t")
	if line == "" || strings.HasPrefix(line, "#") {
		return nil, nil
	}
	if strings.ContainsRune(line, 0) {
		return nil, errors.New("NUL byte in crontab")
	}
	name, value, assignment, err := parseEnvironment(line)
	if err != nil {
		return nil, err
	}
	if assignment {
		return nil, p.setEnvironment(name, value)
	}

	quiet := strings.HasPrefix(line, "-")
	line = strings.TrimLeft(strings.TrimPrefix(line, "-"), " \t")
	schedule, remainder, err := p.splitSchedule(line)
	if err != nil {
		return nil, err
	}
	var user string
	if p.hasUser {
		user, remainder = nextField(remainder)
		if user == "" {
			return nil, errors.New("missing system crontab username")
		}
	}
	commandText, err := stripMailOption(remainder)
	if err != nil {
		return nil, err
	}
	command, input := splitCommand(commandText)
	if strings.TrimSpace(command) == "" {
		return nil, errors.New("missing command")
	}
	event, err := p.event(schedule)
	if err != nil {
		return nil, err
	}
	environment := make([]string, 0, len(p.env))
	for _, key := range slices.Sorted(maps.Keys(p.env)) {
		environment = append(environment, key+"="+p.env[key])
	}
	return &config.JobConfig{
		Name:   fmt.Sprintf("FromCron: %d (%s)", lineNumber, schedule),
		Events: []config.JobEvent{event},
		Tasks: []config.Task{{
			Command:  command,
			UserName: user,
			Crontab:  &config.CrontabCommand{Environment: environment, Stdin: input, Quiet: quiet},
		}},
	}, nil
}

func (p *crontabLineParser) setEnvironment(name, value string) error {
	if name == "RANDOM_DELAY" {
		delay, err := strconv.Atoi(value)
		if err != nil || delay < 0 || delay > config.MaxCrontabRandomDelay {
			return fmt.Errorf("RANDOM_DELAY must be an integer from 0 to %d minutes", config.MaxCrontabRandomDelay)
		}
		p.delay = delay
	}
	p.env[name] = value
	return nil
}

func (p crontabLineParser) event(schedule string) (config.JobEvent, error) {
	if schedule == "@reboot" {
		return config.JobEvent{OnInit: true}, nil
	}
	event := config.JobEvent{
		Cron: schedule,
		Crontab: &config.CrontabTiming{
			TimeZone:    p.env["CRON_TZ"],
			RandomDelay: p.delay,
		},
	}
	if err := event.ValidateCron(); err != nil {
		return config.JobEvent{}, err
	}
	return event, nil
}

func (p crontabLineParser) splitSchedule(line string) (string, string, error) {
	if p.matcher != nil {
		match := p.matcher.FindStringSubmatch(line)
		if match == nil {
			return "", "", errors.New("entry does not match the schedule matcher")
		}
		return strings.Join(strings.Fields(match[1]), " "), match[len(match)-1], nil
	}
	first, rest := nextField(line)
	if first == "@every" {
		duration, command := nextField(rest)
		return first + " " + duration, command, nil
	}
	if strings.HasPrefix(first, "@") {
		return first, rest, nil
	}
	fields := []string{first}
	for len(fields) < 5 {
		var field string
		field, rest = nextField(rest)
		if field == "" {
			return "", "", errors.New("missing schedule fields")
		}
		fields = append(fields, field)
	}
	return strings.Join(fields, " "), rest, nil
}

func nextField(text string) (string, string) {
	text = strings.TrimLeft(text, " \t")
	if i := strings.IndexAny(text, " \t"); i >= 0 {
		return text[:i], strings.TrimLeft(text[i:], " \t")
	}
	return text, ""
}

func stripMailOption(command string) (string, error) {
	if !strings.HasPrefix(command, "-") {
		return command, nil
	}
	option, rest := nextField(command)
	// Cronie's -n option affects mail delivery only.
	if option != "-n" || strings.HasPrefix(rest, "-") {
		return "", errors.New("invalid crontab command option")
	}
	return rest, nil
}
