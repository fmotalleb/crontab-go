package config

import (
	"fmt"

	"github.com/fmotalleb/go-tools/log"
)

// Validate checks the validity of the Config struct.
// It ensures that the log format and log level are valid, and all jobs within the config are also valid.
// If any validation fails, it returns an error with the specific validation error.
// Otherwise, it returns nil.
func (cfg *Config) Validate() error {
	log := log.NewBuilder().FromEnv().MustBuild().Named("Config.Validator")
	// Validate log format
	if err := cfg.LogFormat.Validate(); err != nil {
		return err
	}

	// Validate log level
	if err := cfg.LogLevel.Validate(); err != nil {
		return err
	}

	if err := validateWebserverConfig(cfg); err != nil {
		return err
	}

	// Validate each job in the config
	for _, job := range cfg.Jobs {
		if err := job.Validate(log); err != nil {
			return err
		}
	}

	// All validations passed
	return nil
}

func validateWebserverConfig(cfg *Config) error {
	log := log.NewBuilder().FromEnv().MustBuild()
	if cfg.WebServerAddress == "" {
		log.Warn("no webserver address specified")
		return nil
	}
	if cfg.WebServerAddress != "" && cfg.WebServerPort == 0 {
		return fmt.Errorf("address: %s:%d is not a valid address", cfg.WebServerAddress, cfg.WebServerPort)
	}
	if len(cfg.WebServerPassword) < 8 {
		log.Warn(
			"webserver password is weak",
		)
	}

	return nil
}
