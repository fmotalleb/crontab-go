// Package cmd manages the command line interface/configuration file handling logic
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fmotalleb/go-tools/defaulter"
	"github.com/fmotalleb/go-tools/env"
	"github.com/fmotalleb/go-tools/git"
	"github.com/fmotalleb/go-tools/log"
	"github.com/fmotalleb/gotel"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/cmd/parser"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/global"
	"github.com/fmotalleb/crontab-go/core/jobs"
	coretask "github.com/fmotalleb/crontab-go/core/task"
	"github.com/fmotalleb/crontab-go/core/webserver"
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
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := initConfig(); err != nil {
			return err
		}
		cronInstance := cron.New(cron.WithSeconds())
		global.Put(cronInstance)
		cronInstance.Start()
		l := global.Logger("cron")
		l.Info("Booting up")

		otelResult, otelErr := gotel.Setup(global.CTX(),
			gotel.WithVersion(git.String()),
			gotel.WithConfig(CFG.Observability),
			gotel.WithLogger(l),
		)
		if otelErr != nil {
			l.Warn("observability setup error", zap.Error(otelErr))
		}
		if otelResult.LogCore != nil {
			global.AttachOTelCore(otelResult.LogCore)
		}
		defer otelResult.Shutdown(context.Background()) //nolint:errcheck // shutdown is best-effort

		if err := jobs.InitializeJobs(CFG.Jobs, CFG.Variables); err != nil {
			return err
		}
		var webServerErrors chan error
		if CFG.WebServerAddress != "" {
			webServerErrors = make(chan error, 1)
			go func() {
				webServerErrors <- webserver.NewWebServer(
					global.CTX(),
					CFG.WebServerAddress,
					CFG.WebServerPort,

					CFG.WebServerMetrics,
					&webserver.AuthConfig{
						Username:             CFG.WebserverUsername,
						Password:             CFG.WebServerPassword,
						AllowUnauthenticated: CFG.AllowUnauthenticatedWebserver,
					},
				).Serve()
			}()
		}
		select {
		case <-global.CTX().Done():
		case err := <-webServerErrors:
			if err != nil {
				return err
			}
			if global.CTX().Err() == nil {
				return errors.New("webserver stopped unexpectedly")
			}
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		select {
		case <-cronInstance.Stop().Done():
			l.Info("cron stopped")
		case <-shutdownCtx.Done():
			l.Warn("timed out waiting for cron jobs to stop")
		}
		taskShutdownCtx, taskCancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := jobs.WaitForShutdown(taskShutdownCtx); err != nil {
			l.Warn("timed out waiting for tasks to stop", zap.Error(err))
		}
		taskCancel()
		coretask.CloseHTTPIdleConnections()
		return nil
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

// initConfig loads CFG from the selected configuration file and environment,
// applies defaults, and validates it. It returns wrapped errors from reading,
// unmarshaling, applying defaults, or validation.
func initConfig() error {
	setupEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("cannot read config file: %w", err)
	}
	if err := viper.Unmarshal(CFG); err != nil {
		return fmt.Errorf("cannot unmarshal config file: %w", err)
	}
	if err := defaulter.ApplyDefaults(CFG, CFG); err != nil {
		return fmt.Errorf("cannot apply config defaults: %w", err)
	}
	if err := CFG.Validate(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}
	return nil
}

// setupEnv binds webserver environment aliases and enables automatic environment lookup.
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
			"webserver_allow_unauthenticated",
			"WEBSERVER_ALLOW_UNAUTHENTICATED",
			"ALLOW_UNAUTHENTICATED_WEBSERVER",
		),
		"Cannot bind webserver_allow_unauthenticated env variable: %s",
	)

	viper.AutomaticEnv()
}
