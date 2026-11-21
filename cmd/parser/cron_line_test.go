package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/spf13/viper"

	"github.com/fmotalleb/crontab-go/config"
)

func TestParseEnvironment(t *testing.T) {
	tests := []struct {
		line  string
		name  string
		value string
	}{
		{"Mixed_Case = unquoted # literal  ", "Mixed_Case", "unquoted # literal"},
		{`NAME = '  $HOME {{.Vars.Foo}}  '`, "NAME", "  $HOME {{.Vars.Foo}}  "},
		{`"quoted name"="a=b"`, "quoted name", "a=b"},
		{"EMPTY = ", "EMPTY", ""},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			name, value, ok, err := parseEnvironment(tt.line)
			assert.NoError(t, err)
			assert.True(t, ok)
			assert.Equal(t, tt.name, name)
			assert.Equal(t, tt.value, value)
		})
	}
}

func TestParseConfigSnapshotsEnvironmentAndKeepsCommands(t *testing.T) {
	cron := NewCronString("  # comment\r\nMixed=first\r\n* * * * * 42 command # keep\r\n" +
		"Mixed=second\nCRON_TZ=UTC\nRANDOM_DELAY=30\n* * * * * echo {{.Vars.X}}\n@reboot cat%first%second")
	cfg, err := cron.ParseConfig("", false)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(cfg.Jobs))
	first, second, reboot := cfg.Jobs[0], cfg.Jobs[1], cfg.Jobs[2]
	assert.Equal(t, []string{"Mixed=first"}, first.Tasks[0].Crontab.Environment)
	assert.Equal(t, "42 command # keep", first.Tasks[0].Command)
	assert.NotEqual(t, first.Name, second.Name)
	assert.Equal(t, []string{"CRON_TZ=UTC", "Mixed=second", "RANDOM_DELAY=30"}, second.Tasks[0].Crontab.Environment)
	assert.Equal(t, "echo {{.Vars.X}}", second.Tasks[0].Command)
	assert.Equal(t, "UTC", second.Events[0].Crontab.TimeZone)
	assert.Equal(t, 30, second.Events[0].Crontab.RandomDelay)
	assert.True(t, reboot.Events[0].OnInit)
	assert.Equal(t, "first\nsecond\n", reboot.Tasks[0].Crontab.Stdin)
}

func TestParseSystemConfigWithCustomMatcher(t *testing.T) {
	cron := NewCronString("- @every 15s service-user -n  echo job\n" +
		"* * * * * * service-user echo job")
	cfg, err := cron.ParseConfig(`@every [0-9]+s|(\* ){5}\*`, true)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(cfg.Jobs))
	assert.Equal(t, "service-user", cfg.Jobs[0].Tasks[0].UserName)
	assert.True(t, cfg.Jobs[0].Tasks[0].Crontab.Quiet)
	assert.Equal(t, "echo job", cfg.Jobs[0].Tasks[0].Command)
	assert.Equal(t, "@every 15s", cfg.Jobs[0].Events[0].Cron)
	assert.Equal(t, "* * * * * *", cfg.Jobs[1].Events[0].Cron)
}

func TestParseFileReportsInvalidEntry(t *testing.T) {
	for _, entry := range []string{
		"60 * * * * echo outside range",
		"@unknown echo unknown",
		"* * * * *",
		"* * * * * -x command",
		"NAME='missing quote",
		"RANDOM_DELAY=-1",
		"CRON_TZ=bad-zone\n* * * * * echo bad zone",
	} {
		t.Run(entry, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "test.crontab")
			assert.NoError(t, os.WriteFile(path, []byte("# first line\n"+entry+"\n"), 0o600))
			cfg, err := ParseFile(path, "", false)
			assert.Error(t, err)
			assert.True(t, cfg == nil)
			assert.Contains(t, err.Error(), path+": line ")
		})
	}
}

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		line    string
		command string
		input   string
	}{
		{`printf 'x\%s\n'`, `printf 'x%s\n'`, ""},
		{`cat%one\%two%%four%`, "cat", "one%two\n\nfour\n"},
		{`cat \\%one`, `cat \\`, "one\n"},
		{`cat%`, "cat", ""},
		{`cat%a\zb`, "cat", "a\\zb\n"},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			command, input := splitCommand(tt.line)
			assert.Equal(t, tt.command, command)
			assert.Equal(t, tt.input, input)
		})
	}
}

func TestConvertedConfigRoundTrip(t *testing.T) {
	source := NewCronString("HOME=/tmp\nMixed=one\nMIXED=two\nCRON_TZ=UTC\nRANDOM_DELAY=5\n6~15 * * * * cat%line\n")
	original, err := source.ParseConfig("", false)
	assert.NoError(t, err)
	yaml, err := generateYamlFromCfg(original)
	assert.NoError(t, err)
	settings := viper.New()
	settings.SetConfigType("yaml")
	assert.NoError(t, settings.ReadConfig(strings.NewReader(yaml)))
	var converted config.Config
	assert.NoError(t, settings.Unmarshal(&converted))
	assert.Equal(t, original, &converted)
}
