package scan

import (
	"fmt"
	"os"
	"strings"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

type Logger struct {
	level LogLevel
}

func NewLogger(levelStr string) *Logger {
	return &Logger{level: GetLogLevel(levelStr)}
}

func (logger *Logger) Debug(message string, keys ...string) {
	if logger.level <= LogLevelDebug {
		logger.log(LogLevelDebug, message, keys...)
	}
}

func (logger *Logger) Info(message string, keys ...string) {
	if logger.level <= LogLevelInfo {
		logger.log(LogLevelInfo, message, keys...)
	}
}

func (logger *Logger) Warn(message string, keys ...string) {
	if logger.level <= LogLevelWarn {
		logger.log(LogLevelWarn, message, keys...)
	}
}

func (logger *Logger) Error(message string, keys ...string) {
	if logger.level <= LogLevelError {
		logger.log(LogLevelError, message, keys...)
	}
}

func (logger *Logger) Fatal(message string, keys ...string) {
	logger.log(LogLevelFatal, message, keys...)
	os.Exit(1)
}

func (logger *Logger) log(level LogLevel, message string, keys ...string) {
	lvl := logger.levelToString(level)
	msg := fmt.Sprintf("[%s] %s\n", lvl, message)
	fmt.Printf(msg, keys)
}

func (logger *Logger) levelToString(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	}
	return ""
}

func GetLogLevel(levelStr string) LogLevel {
	level := LogLevelInfo
	switch strings.ToLower(levelStr) {
	case "debug":
		level = LogLevelDebug
	case "info":
		level = LogLevelInfo
	case "warn":
		level = LogLevelWarn
	case "error":
		level = LogLevelError
	case "fatal":
		level = LogLevelFatal
	}

	return level
}
