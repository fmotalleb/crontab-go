package config

import (
	"errors"
	"fmt"

	"github.com/fmotalleb/go-tools/log"
)

// Validate checks the validity of the Config struct.
// It ensures that the log format and log level are valid, and all jobs within the config are also valid.
// If any validation fails, it returns an error with the specific validation error.
// Otherwise, it returns nil.
func (cfg *Config) Validate() error {
	log := log.NewBuilder().FromEnv().MustBuild().Named("Config.Validator")

	if err := validateWebserverConfig(cfg); err != nil {
		return err
	}

	names := make(map[string]struct{}, len(cfg.Jobs))
	for i, job := range cfg.Jobs {
		if job == nil {
			return fmt.Errorf("job at index %d is null", i)
		}
		if job.Name != "" {
			if _, exists := names[job.Name]; exists {
				return fmt.Errorf("duplicate job name %q", job.Name)
			}
			names[job.Name] = struct{}{}
		}
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
	if cfg.WebServerPort == 0 {
		return fmt.Errorf("address: %s:%d is not a valid address", cfg.WebServerAddress, cfg.WebServerPort)
	}
	if cfg.WebServerPort > 65535 {
		return fmt.Errorf("webserver port %d is outside the valid range 1-65535", cfg.WebServerPort)
	}
	hasUsername := cfg.WebserverUsername != ""
	hasPassword := cfg.WebServerPassword != ""
	if hasUsername != hasPassword {
		return errors.New("webserver username and password must be configured together")
	}
	if !hasUsername && !cfg.AllowUnauthenticatedWebserver {
		return errors.New("webserver authentication is required; set webserver_allow_unauthenticated=true only on a trusted network")
	}
	if hasPassword && len(cfg.WebServerPassword) < 8 {
		log.Warn(
			"webserver password is weak",
		)
	}

	return nil
}
