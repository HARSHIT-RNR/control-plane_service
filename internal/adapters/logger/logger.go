package logger

import (
	"log"
	"os"
)

// Logger provides structured logging capabilities
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	warnLogger  *log.Logger
}

// New creates a new logger instance
func New() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lshortfile),
		warnLogger:  log.New(os.Stdout, "[WARN] ", log.LstdFlags|log.Lshortfile),
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.warnLogger.Printf(format, v...)
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.errorLogger.Fatalf(format, v...)
}
