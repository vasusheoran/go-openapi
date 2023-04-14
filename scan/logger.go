package scan

import (
	"fmt"
	"log"
	"os"
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
	level       LogLevel
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
}

func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:       level,
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:  log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		fatalLogger: log.New(os.Stderr, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (logger *Logger) Debug(message string, keys ...interface{}) {
	if logger.level <= LogLevelDebug {
		logger.debugLogger.Output(2, fmt.Sprintf(message, keys...))
	}
}

func (logger *Logger) Info(message string, keys ...interface{}) {
	if logger.level <= LogLevelInfo {
		logger.infoLogger.Output(2, fmt.Sprintf(message, keys...))
	}
}

func (logger *Logger) Warn(message string, keys ...interface{}) {
	if logger.level <= LogLevelWarn {
		logger.warnLogger.Output(2, fmt.Sprintf(message, keys...))
	}
}

func (logger *Logger) Error(message string, keys ...interface{}) {
	if logger.level <= LogLevelError {
		logger.errorLogger.Output(2, fmt.Sprintf(message, keys...))
	}
}

func (logger *Logger) Fatal(message string, keys ...interface{}) {
	logger.fatalLogger.Output(2, fmt.Sprintf(message, keys...))
	os.Exit(1)
}

func (logger *Logger) log(level LogLevel, message string, keys ...interface{}) {
	//msg := fmt.Sprintf(message, keys)
	lvl := logger.levelToString(level)
	msg := fmt.Sprintf("[%s] %s\n", lvl, message)
	fmt.Printf(msg, keys...)
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

func (logger *Logger) SetLogLevel(level LogLevel) {
	logger.level = level
}
