package logging

import (
	"fcstask-backend/backup-tool/config"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	mu       sync.Mutex
	config   LoggingConfig
	file     *os.File
	logLevel int
	writers  []io.Writer
	loggers  map[int]*log.Logger
}

type LoggingConfig struct {
	Level      string
	File       string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
}

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)

func NewLogger(cfg config.LoggingConfig) *Logger {
	l := &Logger{
		config: LoggingConfig{
			Level:      cfg.Level,
			File:       cfg.File,
			MaxSizeMB:  cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAgeDays: cfg.MaxAgeDays,
		},
		writers: []io.Writer{os.Stdout},
		loggers: make(map[int]*log.Logger),
	}

	switch strings.ToLower(l.config.Level) {
	case "debug":
		l.logLevel = DEBUG
	case "info":
		l.logLevel = INFO
	case "warn":
		l.logLevel = WARN
	case "error":
		l.logLevel = ERROR
	default:
		l.logLevel = INFO
	}

	if l.config.File != "" {
		if err := l.openLogFile(); err != nil {
			fmt.Printf("Failed to open log file: %v\n", err)
		}
	}

	l.initLoggers()
	return l
}

func (l *Logger) openLogFile() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
	}

	if err := os.MkdirAll(filepath.Dir(l.config.File), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(l.config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.file = file
	l.writers = append([]io.Writer{file}, os.Stdout)
	l.initLoggers()

	return nil
}

func (l *Logger) initLoggers() {
	for level := DEBUG; level <= ERROR; level++ {
		l.loggers[level] = log.New(io.MultiWriter(l.writers...), "", 0)
	}
}

func (l *Logger) checkRotation() {
	if l.config.File == "" {
		return
	}

	info, err := os.Stat(l.config.File)
	if err != nil {
		return
	}

	maxBytes := int64(l.config.MaxSizeMB) * 1024 * 1024
	if info.Size() < maxBytes {
		return
	}

	l.rotateLogs()
}

func (l *Logger) rotateLogs() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.config.File == "" {
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	backupFile := fmt.Sprintf("%s.%s", l.config.File, timestamp)

	if err := os.Rename(l.config.File, backupFile); err != nil {
		fmt.Printf("Failed to rotate log: %v\n", err)
		return
	}

	l.openLogFile()
	l.cleanOldLogs()
}

func (l *Logger) cleanOldLogs() {
	if l.config.MaxBackups <= 0 && l.config.MaxAgeDays <= 0 {
		return
	}

	dir := filepath.Dir(l.config.File)
	pattern := filepath.Base(l.config.File) + ".*"

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -l.config.MaxAgeDays)
	var toDelete []string

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		if (l.config.MaxAgeDays > 0 && info.ModTime().Before(cutoff)) ||
			(l.config.MaxBackups > 0 && len(matches) > l.config.MaxBackups) {
			toDelete = append(toDelete, match)
		}
	}

	for _, file := range toDelete {
		os.Remove(file)
	}
}

func (l *Logger) log(level int, format string, args ...interface{}) {
	if level < l.logLevel {
		return
	}

	l.checkRotation()

	var prefix string
	switch level {
	case DEBUG:
		prefix = "[DEBUG]"
	case INFO:
		prefix = "[INFO]"
	case WARN:
		prefix = "[WARN]"
	case ERROR:
		prefix = "[ERROR]"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logMessage := fmt.Sprintf("%s %s %s", timestamp, prefix, message)

	l.mu.Lock()
	defer l.mu.Unlock()

	if logger, ok := l.loggers[level]; ok {
		logger.Println(logMessage)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}
