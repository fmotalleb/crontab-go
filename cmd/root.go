// Package cmd manages the command line interface/configuration file handling logic
package cmd

import (
	"os"
	"runtime"

	"github.com/fmotalleb/go-tools/git"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/FMotalleb/crontab-go/cmd/parser"
	"github.com/FMotalleb/crontab-go/config"
)

var (
	cfgFile string
	CFG     *config.Config = &config.Config{}
)

var rootCmd = &cobra.Command{
	Use:   "crontab-go",
	Short: "Crontab replacement for containers",
	Long: `Cronjob-go is a powerful, lightweight, and highly configurable Golang application
designed to replace the traditional crontab in Docker environments.
With its seamless integration and easy-to-use YAML configuration,
Cronjob-go simplifies the process of scheduling and managing recurring tasks
within your containerized applications.`,
	Version: git.String(),
	Run: func(_ *cobra.Command, _ []string) {
		initConfig()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	_ = godotenv.Load()

	rootCmd.AddCommand(parser.ParserCmd)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is config.yaml)")

	// cobra.OnInitialize()
}

func warnOnErr(err error, message string) {
	if err != nil {
		logrus.Warnf(message, err)
	}
}

func panicOnErr(err error, message string) {
	if err != nil {
		logrus.Panicf(message, err)
	}
}

func initConfig() {
	if runtime.GOOS == "windows" {
		viper.SetDefault("shell", "C:\\WINDOWS\\system32\\cmd.exe")
		viper.SetDefault("shell_args", "/c")
	} else {
		viper.SetDefault("shell", "/bin/sh")
		viper.SetDefault("shell_args", "-c")
	}

	setupEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	panicOnErr(
		viper.ReadInConfig(),
		"Cannot read the config file: %s",
	)
	panicOnErr(
		viper.Unmarshal(CFG),
		"Cannot unmarshal the config file: %s",
	)
	panicOnErr(
		CFG.Validate(logrus.WithField("section", "config.validation")),
		"Failed to initialize config file: %s",
	)
}

func setupEnv() {
	viper.SetDefault("log_timestamp_format", "2006-01-02T15:04:05Z07:00")
	warnOnErr(
		viper.BindEnv(
			"log_timestamp_format",
			"timestamp_format",
		),
		"Cannot bind log_timestamp_format env variable: %s",
	)
	viper.SetDefault("log_format", "ansi")
	warnOnErr(
		viper.BindEnv(
			"log_format",
			"output_format",
		),
		"Cannot bind log_format env variable: %s",
	)
	warnOnErr(
		viper.BindEnv(
			"log_file",
			"output_file",
		),
		"Cannot bind log_file env variable: %s",
	)
	viper.SetDefault("log_stdout", true)
	warnOnErr(
		viper.BindEnv(
			"log_stdout",
			"print",
		),
		"Cannot bind log_stdout env variable: %s",
	)
	warnOnErr(
		viper.BindEnv(
			"webserver_port",
			"listen_port",
		),
		"Cannot bind webserver_port env variable: %s",
	)
	warnOnErr(
		viper.BindEnv(
			"webserver_address",
			"webserver_listen_address",
			"listen_address",
		),
		"Cannot bind webserver_address env variable: %s",
	)
	warnOnErr(
		viper.BindEnv(
			"webserver_password",
			"password",
		),
		"Cannot bind webserver_password env variable: %s",
	)

	warnOnErr(
		viper.BindEnv(
			"webserver_metrics",
			"prometheus_metrics",
		),
		"Cannot bind webserver_metrics env variable: %s",
	)

	warnOnErr(
		viper.BindEnv(
			"webserver_username",
			"username",
		),
		"Cannot bind webserver_username env variable: %s",
	)

	viper.SetDefault("log_level", "info")
	warnOnErr(
		viper.BindEnv(
			"log_level",
			"level",
		),
		"Cannot bind log_level env variable: %s",
	)

	warnOnErr(
		viper.BindEnv(
			"shell",
		),
		"Cannot bind shell env variable: %s",
	)
	warnOnErr(
		viper.BindEnv(
			"shell_args",
		),
		"Cannot bind shell_args env variable: %s",
	)

	viper.AutomaticEnv()
}
