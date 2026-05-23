package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func InitLogger(lvl string, filePath string, console bool) error {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"

	level, err := zerolog.ParseLevel(lvl)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var output io.Writer = os.Stdout

	if filePath != "" {
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		if console {
			output = io.MultiWriter(os.Stdout, file)
		} else {
			output = file
		}
	}

	Log = zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	return nil
}

func With() zerolog.Context {
	return Log.With()
}
