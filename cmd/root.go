// Package cmd manages the command line interface/configuration file handling logic
package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/fmotalleb/go-tools/env"
	"github.com/fmotalleb/go-tools/git"
	"github.com/fmotalleb/go-tools/log"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fmotalleb/crontab-go/cmd/parser"
	"github.com/fmotalleb/crontab-go/config"
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
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			log.SetDebugDefaults()
		}
	},
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

	if ll := os.Getenv("LOG_LEVEL"); ll != "" {
		os.Setenv("ZAPLOG_LEVEL", ll)
	}
	if ltf := os.Getenv("LOG_TIMESTAMP_FORMAT"); ltf != "" {
		os.Setenv("ZAPLOG_TIME_FORMAT", ltf)
	}
	if lf := os.Getenv("LOG_FORMAT"); lf == "ansi" {
		os.Setenv("ZAPLOG_DEVELOPMENT", "true")
	}
	logStdout := env.BoolOr("LOG_STDOUT", false)
	if lf := os.Getenv("LOG_FILE"); lf != "" {
		rlf := lf
		if logStdout {
			rlf = "stdout," + lf
		}
		os.Setenv("ZAPLOG_OUTPUT_PATHS", rlf)
		os.Setenv("ZAPLOG_ERROR_PATHS", rlf)
	}

	rootCmd.AddCommand(parser.ParserCmd)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable debug logger")

	// cobra.OnInitialize()
}

func warnOnErr(err error, message string) {
	if err != nil {
		fmt.Printf("%s, %v", message, err)
	}
}

func panicOnErr(err error, message string) {
	if err != nil {
		panic(fmt.Errorf("%s, %w", message, err))
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
		CFG.Validate(),
		"Failed to initialize config file: %s",
	)
}

func setupEnv() {
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
