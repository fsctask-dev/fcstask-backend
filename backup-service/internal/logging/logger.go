package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"fcstask-backend/backup-service/internal/config"
)

type Logger struct {
	sl    *slog.Logger
	rw    *rotatingWriter
	level slog.Level
}

func New(cfg config.LoggingConfig, secrets ...string) (*Logger, error) {
	level := parseLevel(cfg.Level)

	var writers []io.Writer
	if cfg.Stdout {
		writers = append(writers, os.Stdout)
	}

	var rw *rotatingWriter
	if cfg.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0o700); err != nil {
			return nil, fmt.Errorf("create log dir: %w", err)
		}
		var err error
		rw, err = newRotatingWriter(cfg.File, cfg.MaxSizeMB, cfg.MaxBackups, cfg.MaxAgeDays)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		writers = append(writers, rw)
	}
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	var clean []string
	for _, s := range secrets {
		if s != "" {
			clean = append(clean, s)
		}
	}
	out := io.Writer(io.MultiWriter(writers...))
	if len(clean) > 0 {
		out = &redactingWriter{w: out, secrets: clean}
	}

	handler := slog.NewTextHandler(out, &slog.HandlerOptions{Level: level})
	return &Logger{sl: slog.New(handler), rw: rw, level: level}, nil
}

type redactingWriter struct {
	w       io.Writer
	secrets []string
}

func (rw *redactingWriter) Write(p []byte) (int, error) {
	s := string(p)
	for _, sec := range rw.secrets {
		s = strings.ReplaceAll(s, sec, "***")
	}
	if _, err := rw.w.Write([]byte(s)); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (l *Logger) Debug(format string, args ...any) { l.sl.Debug(fmt.Sprintf(format, args...)) }
func (l *Logger) Info(format string, args ...any)  { l.sl.Info(fmt.Sprintf(format, args...)) }
func (l *Logger) Warn(format string, args ...any)  { l.sl.Warn(fmt.Sprintf(format, args...)) }
func (l *Logger) Error(format string, args ...any) { l.sl.Error(fmt.Sprintf(format, args...)) }

func (l *Logger) Close() error {
	if l.rw != nil {
		return l.rw.Close()
	}
	return nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type rotatingWriter struct {
	mu         sync.Mutex
	path       string
	maxBytes   int64
	maxBackups int
	maxAge     time.Duration
	f          *os.File
	size       int64
}

func newRotatingWriter(path string, maxSizeMB, maxBackups, maxAgeDays int) (*rotatingWriter, error) {
	rw := &rotatingWriter{
		path:       path,
		maxBytes:   int64(maxSizeMB) * 1024 * 1024,
		maxBackups: maxBackups,
		maxAge:     time.Duration(maxAgeDays) * 24 * time.Hour,
	}
	if err := rw.open(); err != nil {
		return nil, err
	}
	return rw, nil
}

func (rw *rotatingWriter) open() error {
	f, err := os.OpenFile(rw.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return err
	}
	rw.f = f
	rw.size = info.Size()
	return nil
}

func (rw *rotatingWriter) Write(p []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.maxBytes > 0 && rw.size+int64(len(p)) > rw.maxBytes {
		if err := rw.rotate(); err != nil {

			fmt.Fprintf(os.Stderr, "log rotation failed: %v\n", err)
		}
	}
	n, err := rw.f.Write(p)
	rw.size += int64(n)
	return n, err
}

func (rw *rotatingWriter) rotate() error {
	if err := rw.f.Close(); err != nil {
		return err
	}
	backup := fmt.Sprintf("%s.%s", rw.path, time.Now().Format("20060102_150405.000"))
	if err := os.Rename(rw.path, backup); err != nil {
		return err
	}
	if err := rw.open(); err != nil {
		return err
	}
	rw.cleanup()
	return nil
}

func (rw *rotatingWriter) cleanup() {
	matches, err := filepath.Glob(rw.path + ".*")
	if err != nil {
		return
	}
	type entry struct {
		path string
		mod  time.Time
	}
	var entries []entry
	cutoff := time.Now().Add(-rw.maxAge)
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil {
			continue
		}

		if rw.maxAge > 0 && info.ModTime().Before(cutoff) {
			_ = os.Remove(m)
			continue
		}
		entries = append(entries, entry{m, info.ModTime()})
	}

	if rw.maxBackups > 0 && len(entries) > rw.maxBackups {
		sort.Slice(entries, func(i, j int) bool { return entries[i].mod.After(entries[j].mod) })
		for _, e := range entries[rw.maxBackups:] {
			_ = os.Remove(e.path)
		}
	}
}

func (rw *rotatingWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	if rw.f != nil {
		err := rw.f.Close()
		rw.f = nil
		return err
	}
	return nil
}
