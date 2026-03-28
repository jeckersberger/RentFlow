package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New creates a zerolog.Logger configured for the given level.
// In development (LOG_LEVEL=debug or ENV=development), it uses a
// human-friendly console writer. Otherwise it outputs structured JSON.
func New(level string) zerolog.Logger {
	lvl := parseLevel(level)

	var w io.Writer
	if isDev() {
		w = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	} else {
		w = os.Stdout
	}

	return zerolog.New(w).
		Level(lvl).
		With().
		Timestamp().
		Caller().
		Logger()
}

// WithService returns a child logger tagged with the service name.
func WithService(logger zerolog.Logger, service string) zerolog.Logger {
	return logger.With().Str("service", service).Logger()
}

func parseLevel(level string) zerolog.Level {
	switch level {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

func isDev() bool {
	env := os.Getenv("ENV")
	return env == "" || env == "development"
}
