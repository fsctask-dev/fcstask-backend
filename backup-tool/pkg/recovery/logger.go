package recovery

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

type RestoreLogger struct {
	mu     sync.Mutex
	level  LogLevel
	logger *log.Logger
	file   *os.File
	stdout bool
}

func NewRestoreLogger(logFile string, debug bool) *RestoreLogger {
	level := LevelInfo
	if debug {
		level = LevelDebug
	}
	var writers []io.Writer
	if logFile != "" {
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err == nil {
			f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				writers = append(writers, f)
				rl := &RestoreLogger{level: level, file: f, stdout: true}
				rl.logger = log.New(io.MultiWriter(append(writers, os.Stdout)...), "", 0)
				return rl
			}
		}
	}
	rl := &RestoreLogger{level: level, stdout: true}
	rl.logger = log.New(os.Stdout, "", 0)
	return rl
}

func (l *RestoreLogger) formatMessage(levelStr, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s [RESTORE %s] %s", timestamp, levelStr, msg)
}

func (l *RestoreLogger) log(level LogLevel, levelStr, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(l.formatMessage(levelStr, format, args...))
}

func (l *RestoreLogger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, "DEBUG", format, args...)
}

func (l *RestoreLogger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, "INFO ", format, args...)
}

func (l *RestoreLogger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, "WARN ", format, args...)
}

func (l *RestoreLogger) Error(format string, args ...interface{}) {
	l.log(LevelError, "ERROR", format, args...)
}

func (l *RestoreLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
