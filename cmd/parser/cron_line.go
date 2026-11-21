package parser

import (
	"errors"
	"regexp"
	"strings"
)

var envAssignment = regexp.MustCompile(`^(?:"([^"=]+)"|'([^'=]+)'|([^"'= \t]+))[ \t]*=(.*)$`)

func parseEnvironment(line string) (name, value string, assignment bool, err error) {
	match := envAssignment.FindStringSubmatch(line)
	if match == nil {
		return "", "", false, nil
	}
	name = match[1] + match[2] + match[3]
	value = strings.TrimSpace(match[4])
	if value != "" && (value[0] == '\'' || value[0] == '"') {
		end := strings.IndexByte(value[1:], value[0]) + 1
		if end == 0 || end != len(value)-1 {
			return "", "", false, errors.New("expected a matching quote at the end of the environment value")
		}
		value = value[1:end]
	}
	return name, value, true, nil
}

// splitCommand removes cron escapes and separates the shell command from its stdin.
func splitCommand(line string) (string, string) {
	var command, input strings.Builder
	dst := &command
	for i := 0; i < len(line); i++ {
		switch {
		case line[i] == '\\' && i+1 < len(line):
			i++
			if line[i] != '%' {
				dst.WriteByte('\\')
			}
			dst.WriteByte(line[i])
		case line[i] != '%':
			dst.WriteByte(line[i])
		case dst == &command:
			dst = &input
		default:
			input.WriteByte('\n')
		}
	}
	if data := input.String(); data != "" && !strings.HasSuffix(data, "\n") {
		input.WriteByte('\n')
	}
	return command.String(), input.String()
}
