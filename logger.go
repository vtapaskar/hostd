package main

import (
	"fmt"
	"log"
	"log/syslog"
)

// Logger wraps syslog functionality
type Logger struct {
	syslog *syslog.Writer
}

// NewLogger creates a new logger with syslog integration
func NewLogger() (*Logger, error) {
	syslogWriter, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "hostd")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to syslog: %v", err)
	}

	return &Logger{
		syslog: syslogWriter,
	}, nil
}

// Close closes the syslog connection
func (l *Logger) Close() error {
	return l.syslog.Close()
}

// Critical logs a critical error message
func (l *Logger) Critical(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.syslog.Crit(msg)
	log.Printf("[CRITICAL] %s", msg)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.syslog.Err(msg)
	log.Printf("[ERROR] %s", msg)
}

// Info logs an informational message
func (l *Logger) Info(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.syslog.Info(msg)
	log.Printf("[INFO] %s", msg)
}
