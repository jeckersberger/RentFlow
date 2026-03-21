package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger interface defines the logging methods that all services use.
// This interface must remain stable to ensure backward compatibility.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	WithPrefix(prefix string) Logger
	WithCorrelationID(id string) Logger
	WithRequestID(id string) Logger
	WithField(key string, value interface{}) Logger
}

// SimpleLogger implements the Logger interface using zerolog internally.
type SimpleLogger struct {
	zerologger    zerolog.Logger
	level         string
	prefix        string
	correlationID string
	requestID     string
	fields        map[string]interface{}
}

// New creates a new logger with the given level and prefix.
func New(level, prefix string) Logger {
	if level == "" {
		level = "info"
	}
	if prefix == "" {
		prefix = "RentFlow"
	}

	// Parse log level
	logLevel := parseLevel(level)

	// Create zerolog logger with console writer
	zlog := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		Level(logLevel).
		With().
		Str("service", prefix).
		Logger()

	return &SimpleLogger{
		zerologger: zlog,
		level:      level,
		prefix:     prefix,
		fields:     make(map[string]interface{}),
	}
}

// parseLevel converts string log level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch level {
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

// buildLogger creates a logger with correlation ID, request ID, and fields
func (l *SimpleLogger) buildLogger() zerolog.Logger {
	ctx := l.zerologger.With()

	if l.correlationID != "" {
		ctx = ctx.Str("correlationID", l.correlationID)
	}

	if l.requestID != "" {
		ctx = ctx.Str("requestID", l.requestID)
	}

	for key, value := range l.fields {
		ctx = ctx.Interface(key, value)
	}

	return ctx.Logger()
}

// Debug logs a debug message
func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	if l.level != "debug" {
		return
	}
	logger := l.buildLogger()
	logger.Debug().Msg(msg)
}

// Info logs an info message
func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	logger := l.buildLogger()
	logger.Info().Msg(msg)
}

// Warn logs a warning message
func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	logger := l.buildLogger()
	logger.Warn().Msg(msg)
}

// Error logs an error message
func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	logger := l.buildLogger()
	logger.Error().Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *SimpleLogger) Fatal(msg string, args ...interface{}) {
	logger := l.buildLogger()
	logger.Fatal().Msg(msg)
}

// WithPrefix returns a new logger with a different prefix
func (l *SimpleLogger) WithPrefix(prefix string) Logger {
	newLogger := &SimpleLogger{
		zerologger:    l.zerologger.With().Str("service", prefix).Logger(),
		level:         l.level,
		prefix:        prefix,
		correlationID: l.correlationID,
		requestID:     l.requestID,
		fields:        make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithCorrelationID returns a new logger with a correlation ID set
func (l *SimpleLogger) WithCorrelationID(id string) Logger {
	newLogger := &SimpleLogger{
		zerologger:    l.zerologger,
		level:         l.level,
		prefix:        l.prefix,
		correlationID: id,
		requestID:     l.requestID,
		fields:        make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithRequestID returns a new logger with a request ID set
func (l *SimpleLogger) WithRequestID(id string) Logger {
	newLogger := &SimpleLogger{
		zerologger:    l.zerologger,
		level:         l.level,
		prefix:        l.prefix,
		correlationID: l.correlationID,
		requestID:     id,
		fields:        make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithField returns a new logger with an additional structured field
func (l *SimpleLogger) WithField(key string, value interface{}) Logger {
	newLogger := &SimpleLogger{
		zerologger:    l.zerologger,
		level:         l.level,
		prefix:        l.prefix,
		correlationID: l.correlationID,
		requestID:     l.requestID,
		fields:        make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value

	return newLogger
}
