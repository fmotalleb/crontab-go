// Package parser manages holds the logic behind the sub command `parse`
// this package is responsible for parsing a crontab file into valid config yaml file
package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/fmotalleb/go-tools/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	cfg       = &parserConfig{}
	ParserCmd = &cobra.Command{
		Use:       "parse <crontab file path>",
		ValidArgs: []string{"crontab file path"},
		Short:     "Parse crontab syntax and converts it into yaml syntax for crontab-go",
		Run:       run,
	}
)

func run(cmd *cobra.Command, _ []string) {
	log := log.NewBuilder().FromEnv().MustBuild()
	cfg.cronFile = cmd.Flags().Arg(0)

	log.Debug("source file: ", zap.String("file", cfg.cronFile))
	cron, err := readInCron(cfg)
	if err != nil {
		log.Panic("failed to read cron file", zap.Error(err))
	}
	finalConfig, err := cron.ParseConfig(
		cfg.cronMatcher,
		cfg.hasUser,
	)
	if err != nil {
		log.Panic("cannot parse given cron file", zap.Error(err))
	}
	result, err := generateYamlFromCfg(finalConfig)
	if err != nil {
		log.Panic("failed to generate yaml", zap.Error(err))
	}
	fmt.Println("# yaml-language-server: $schema=https://raw.githubusercontent.com/fmotalleb/crontab-go/main/schema.json")
	fmt.Println(result)
	if cfg.output != "" {
		writeOutput(log, cfg, result)
	}
	log.Info("Done writing output")
	os.Exit(0)
}

func writeOutput(log *zap.Logger, cfg *parserConfig, result string) {
	outputFile, err := os.OpenFile(cfg.output, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		log.Panic("failed to open output file", zap.Error(err))
	}
	buf := bytes.NewBufferString(result)
	_, err = io.Copy(outputFile, buf)
	if err != nil {
		log.Panic("failed to write output file", zap.Error(err))
	}
}

func readInCron(cfg *parserConfig) (*CronString, error) {
	if cfg.cronFile == "" {
		return nil, errors.New("please provide a cron file path, usage: `--help`")
	}
	file, err := os.OpenFile(cfg.cronFile, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("can't open cron file: %w", err)
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("can't stat cron file: %w", err)
	}
	content := make([]byte, stat.Size())
	_, err = file.Read(content)
	if err != nil {
		return nil, fmt.Errorf("can't open cron file: %w", err)
	}
	str := string(content)
	cron := NewCronString(str)
	return &cron, nil
}

func init() {
	ParserCmd.PersistentFlags().StringVarP(&cfg.output, "output", "o", "", "output file to write configuration to")
	ParserCmd.PersistentFlags().BoolVarP(&cfg.hasUser, "with-user", "u", false, "indicates that whether the given cron file has user field")
	ParserCmd.PersistentFlags().StringVar(&cfg.cronMatcher, "matcher", `(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|(\d+(\/|-)\d+)|\d+|\*|(\*\/\d))\s*){5,7})`, "matcher for cron")
}
