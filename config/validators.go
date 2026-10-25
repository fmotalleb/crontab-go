package config

import (
	"errors"
	"fmt"

	"github.com/fmotalleb/go-tools/log"
)

// Validate checks webserver settings and jobs, rejecting nil jobs and duplicate
// nonempty job names, including names of disabled jobs. It returns the first
// configuration or job validation error, or nil when all checks pass.
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

// validateWebserverConfig skips checks when no address is set. Otherwise it
// requires a port in 1-65535 and either both credentials or empty credentials
// with explicit permission for unauthenticated access. It returns an error for invalid settings;
// passwords shorter than eight bytes are allowed.
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
