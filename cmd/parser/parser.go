// Package parser reads legacy crontab files and implements the parse subcommand.
package parser

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var ParserCmd = &cobra.Command{
	Use:   "parse <crontab file path>",
	Args:  cobra.ExactArgs(1),
	Short: "Convert a crontab file to crontab-go YAML",
	RunE:  run,
}

var (
	output  string
	matcher string
	hasUser bool
)

func run(cmd *cobra.Command, args []string) error {
	cfg, err := ParseFile(args[0], matcher, hasUser)
	if err != nil {
		return err
	}
	result, err := generateYamlFromCfg(cfg)
	if err != nil {
		return err
	}
	if output != "" {
		if writeErr := os.WriteFile(output, []byte(result), 0o644); writeErr != nil {
			return fmt.Errorf("write converted config: %w", writeErr)
		}
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(),
		"# yaml-language-server: $schema=https://raw.githubusercontent.com/fmotalleb/crontab-go/main/schema.json\n"+result)
	return err
}

func init() {
	ParserCmd.Flags().StringVarP(&output, "output", "o", "", "output file to write configuration to")
	ParserCmd.Flags().BoolVarP(&hasUser, "with-user", "u", false, "read a username between each schedule and command")
	ParserCmd.Flags().StringVar(&matcher, "matcher", "", "override the regular expression used to extract the schedule")
}
