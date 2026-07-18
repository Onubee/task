package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func (l LogLevel) String() string {
	return []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[l]
}

type Logger struct {
	mu     sync.Mutex
	out    io.Writer
	level  LogLevel
	fields map[string]interface{}
}

func NewLogger(level string) *Logger {
	lvl := InfoLevel
	switch strings.ToLower(level) {
	case "debug":
		lvl = DebugLevel
	case "info":
		lvl = InfoLevel
	case "warn":
		lvl = WarnLevel
	case "error":
		lvl = ErrorLevel
	case "fatal":
		lvl = FatalLevel
	}
	return &Logger{
		out:    os.Stdout,
		level:  lvl,
		fields: make(map[string]interface{}),
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	file = file[strings.LastIndex(file, "/")+1:]

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("%s | %s | %s:%d | %s\n", timestamp, level.String(), file, line, msg)

	if _, err := l.out.Write([]byte(logLine)); err != nil {
		if _, stderrErr := os.Stderr.Write([]byte("Failed to write log: " + err.Error() + "\n")); stderrErr != nil {
			panic("cannot write to stderr: " + stderrErr.Error())
		}
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FatalLevel, format, args...)
	os.Exit(1)
}
