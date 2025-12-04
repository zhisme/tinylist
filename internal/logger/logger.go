package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Level represents log level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns string representation of level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLevel parses level from string
func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger provides structured logging
type Logger struct {
	level  Level
	format string // "json" or "text"
	output io.Writer
	logger *log.Logger
}

// New creates a new logger
func New(level, format string) *Logger {
	return &Logger{
		level:  ParseLevel(level),
		format: format,
		output: os.Stdout,
		logger: log.New(os.Stdout, "", 0),
	}
}

// SetOutput sets the logger output
func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
	l.logger.SetOutput(w)
}

// log writes a log entry
func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	if l.format == "json" {
		l.logJSON(level, msg, fields)
	} else {
		l.logText(level, msg, fields)
	}
}

// logJSON writes JSON formatted log
func (l *Logger) logJSON(level Level, msg string, fields map[string]interface{}) {
	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level.String(),
		"message":   msg,
	}

	// Merge additional fields
	for k, v := range fields {
		entry[k] = v
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	l.logger.Println(string(data))
}

// logText writes text formatted log
func (l *Logger) logText(level Level, msg string, fields map[string]interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	output := fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), msg)

	if len(fields) > 0 {
		output += " "
		for k, v := range fields {
			output += fmt.Sprintf("%s=%v ", k, v)
		}
	}

	l.logger.Println(output)
}

// Debug logs debug message
func (l *Logger) Debug(msg string) {
	l.log(LevelDebug, msg, nil)
}

// Debugf logs formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, fmt.Sprintf(format, args...), nil)
}

// DebugFields logs debug message with fields
func (l *Logger) DebugFields(msg string, fields map[string]interface{}) {
	l.log(LevelDebug, msg, fields)
}

// Info logs info message
func (l *Logger) Info(msg string) {
	l.log(LevelInfo, msg, nil)
}

// Infof logs formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, fmt.Sprintf(format, args...), nil)
}

// InfoFields logs info message with fields
func (l *Logger) InfoFields(msg string, fields map[string]interface{}) {
	l.log(LevelInfo, msg, fields)
}

// Warn logs warning message
func (l *Logger) Warn(msg string) {
	l.log(LevelWarn, msg, nil)
}

// Warnf logs formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...), nil)
}

// WarnFields logs warning message with fields
func (l *Logger) WarnFields(msg string, fields map[string]interface{}) {
	l.log(LevelWarn, msg, fields)
}

// Error logs error message
func (l *Logger) Error(msg string) {
	l.log(LevelError, msg, nil)
}

// Errorf logs formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...), nil)
}

// ErrorFields logs error message with fields
func (l *Logger) ErrorFields(msg string, fields map[string]interface{}) {
	l.log(LevelError, msg, fields)
}

// With returns a logger with pre-set fields
func (l *Logger) With(fields map[string]interface{}) *LoggerWith {
	return &LoggerWith{
		logger: l,
		fields: fields,
	}
}

// LoggerWith is a logger with pre-set fields
type LoggerWith struct {
	logger *Logger
	fields map[string]interface{}
}

// Debug logs debug message with pre-set fields
func (lw *LoggerWith) Debug(msg string) {
	lw.logger.log(LevelDebug, msg, lw.fields)
}

// Info logs info message with pre-set fields
func (lw *LoggerWith) Info(msg string) {
	lw.logger.log(LevelInfo, msg, lw.fields)
}

// Warn logs warning message with pre-set fields
func (lw *LoggerWith) Warn(msg string) {
	lw.logger.log(LevelWarn, msg, lw.fields)
}

// Error logs error message with pre-set fields
func (lw *LoggerWith) Error(msg string) {
	lw.logger.log(LevelError, msg, lw.fields)
}
