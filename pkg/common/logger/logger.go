package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
}

type SimpleLogger struct {
	level  string
	prefix string
}

// New creates a new logger with the given level
func New(level, prefix string) Logger {
	if level == "" {
		level = "info"
	}
	if prefix == "" {
		prefix = "RentFlow"
	}
	return &SimpleLogger{
		level:  level,
		prefix: prefix,
	}
}

func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	if l.level == "debug" {
		log.Printf("[DEBUG] %s: "+msg+"\n", append([]interface{}{l.prefix}, args...)...)
	}
}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] %s: "+msg+"\n", append([]interface{}{l.prefix}, args...)...)
}

func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] %s: "+msg+"\n", append([]interface{}{l.prefix}, args...)...)
}

func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] %s: "+msg+"\n", append([]interface{}{l.prefix}, args...)...)
}

func (l *SimpleLogger) Fatal(msg string, args ...interface{}) {
	log.Printf("[FATAL] %s: "+msg+"\n", append([]interface{}{l.prefix}, args...)...)
	os.Exit(1)
}

func (l *SimpleLogger) WithPrefix(prefix string) Logger {
	return &SimpleLogger{
		level:  l.level,
		prefix: prefix,
	}
}
