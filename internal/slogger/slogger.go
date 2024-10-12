// Package slogger represents a logger.
package slogger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// LogFormat represents the format of the log output.
type LogFormat string

const (
	// LogFormatJSON represents the JSON log format.
	LogFormatJSON LogFormat = "json"

	// LogFormatText represents the text log format.
	LogFormatText LogFormat = "text"
)

// logger represents a logger.
type logger struct {
	level     *slog.LevelVar
	format    LogFormat
	addSource bool
}

// NewLogger creates a new logger.
func NewLogger(opts ...Option) *slog.Logger {
	logg := &logger{
		level:  &slog.LevelVar{}, // Default log level is INFO.
		format: LogFormatJSON,
	}

	// Apply options.
	for _, opt := range opts {
		opt(logg)
	}

	slogOpts := &slog.HandlerOptions{
		AddSource: logg.addSource,
		Level:     logg.level,
	}

	var logHandler slog.Handler = slog.NewJSONHandler(os.Stdout, slogOpts)
	if logg.format == LogFormatText {
		logHandler = slog.NewTextHandler(os.Stdout, slogOpts)
	}

	return slog.New(logHandler)
}

// Option represents an option for Logger.
type Option func(l *logger)

// WithLevel is an option for Logger that sets the log level.
//
// The level must be one of the following:
//
//   - slog.LevelDebug
//   - slog.LevelInfo
//   - slog.LevelWarn
//   - slog.LevelError
func WithLevel(level slog.Level) Option {
	return func(l *logger) {
		l.level.Set(level)
	}
}

// WithFormat is an option for Logger that sets the format of the log output.
// The format must be one of the following:
//
//   - LogFormatJSON: Output logs in JSON format.
//   - LogFormatText: Output logs in plain text format.
func WithFormat(format LogFormat) Option {
	return func(l *logger) {
		l.format = format
	}
}

// WithAddSource is an option for Logger that sets whether to add source file and line information to each log message.
func WithAddSource(addSource bool) Option {
	return func(l *logger) {
		l.addSource = addSource
	}
}

// ParseLogLevel parses a log level as a slog.Level from a string.
//
// The level must be one of the following:
//
//   - debug
//   - info
//   - warn
//   - error
//
// If the level is not recognized, an error is returned.
func ParseLogLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unknown log level: %s", level)
	}
}
