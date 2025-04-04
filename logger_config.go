package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/urfave/cli"
)

const (
	pathFlagName   = "log.path"
	levelFlagName  = "log.level"
	formatFlagName = "log.format"
)

type LogFormat string

const (
	JSONLogFormat LogFormat = "json"
	TextLogFormat LogFormat = "text"
)

type LoggerConfig struct {
	Format       LogFormat
	OutputWriter io.Writer
	HandlerOpts  logging.SLoggerOptions
}

var loggerFlags = []cli.Flag{
	cli.StringFlag{
		Name:   levelFlagName,
		Usage:  `The lowest log level that will be output. Accepted options are "debug", "info", "warn", "error"`,
		Value:  "info",
		EnvVar: envVarPrefix + "LOG_LEVEL",
	},
	cli.StringFlag{
		Name:   pathFlagName,
		Usage:  "Path to file where logs will be written",
		Value:  "",
		EnvVar: envVarPrefix + "LOG_PATH",
	},
	cli.StringFlag{
		Name:   formatFlagName,
		Usage:  "The format of the log file. Accepted options are 'json' and 'text'",
		Value:  "json",
		EnvVar: envVarPrefix + "LOG_FORMAT",
	},
}

func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Format:       JSONLogFormat,
		OutputWriter: os.Stdout,
		HandlerOpts: logging.SLoggerOptions{
			AddSource:   false,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
			TimeFormat:  time.StampMilli,
			NoColor:     false,
		},
	}
}

func ReadLoggerCLIConfig(ctx *cli.Context) (*LoggerConfig, error) {
	cfg := DefaultLoggerConfig()
	format := ctx.GlobalString(formatFlagName)
	if format == "json" {
		cfg.Format = JSONLogFormat
	} else if format == "text" {
		cfg.Format = TextLogFormat
	} else {
		return nil, fmt.Errorf("invalid log file format %s", format)
	}

	path := ctx.GlobalString(pathFlagName)
	if path != "" {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		cfg.OutputWriter = io.MultiWriter(os.Stdout, f)
	}
	logLevel := ctx.GlobalString(levelFlagName)
	var level slog.Level
	err := level.UnmarshalText([]byte(logLevel))
	if err != nil {
		panic("failed to parse log level " + logLevel)
	}
	cfg.HandlerOpts.Level = level

	return &cfg, nil
}

func NewLogger(cfg LoggerConfig) (logging.Logger, error) {
	if cfg.Format == JSONLogFormat {
		return logging.NewJsonSLogger(cfg.OutputWriter, &cfg.HandlerOpts), nil
	}
	if cfg.Format == TextLogFormat {
		return logging.NewTextSLogger(cfg.OutputWriter, &cfg.HandlerOpts), nil
	}
	return nil, fmt.Errorf("unknown log format: %s", cfg.Format)
}
