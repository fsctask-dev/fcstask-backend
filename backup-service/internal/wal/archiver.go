package wal

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"fcstask-backend/backup-service/internal/config"
	"fcstask-backend/backup-service/internal/logging"
)

const (
	archiverMinBackoff = 2 * time.Second
	archiverMaxBackoff = 60 * time.Second
)

type Archiver struct {
	cfg    config.WALConfig
	logger *logging.Logger
}

func NewArchiver(cfg config.WALConfig, logger *logging.Logger) *Archiver {
	return &Archiver{cfg: cfg, logger: logger}
}

func (a *Archiver) EnsureSlot(ctx context.Context) error {
	if !toolAvailable("pg_receivewal") {
		return fmt.Errorf("pg_receivewal not found on PATH")
	}
	if a.cfg.SlotName == "" {
		return nil
	}
	args := connArgs(a.cfg.Replication)
	args = append(args,
		"--create-slot", "--if-not-exists",
		"--slot", a.cfg.SlotName,
		"--no-password",
	)
	out, err := runCommand(ctx, a.cfg.Replication, "pg_receivewal", args...)
	if err != nil {
		if strings.Contains(strings.ToLower(out+err.Error()), "already exists") {
			return nil
		}
		return err
	}
	a.logger.Info("Replication slot %q is ready", a.cfg.SlotName)
	return nil
}

func (a *Archiver) Run(ctx context.Context) error {
	if !toolAvailable("pg_receivewal") {
		return fmt.Errorf("pg_receivewal not found on PATH")
	}
	if err := os.MkdirAll(a.cfg.ArchiveDir, 0o700); err != nil {
		return fmt.Errorf("create archive dir: %w", err)
	}
	if err := a.EnsureSlot(ctx); err != nil {
		a.logger.Warn("Could not ensure replication slot: %v", err)
	}

	backoff := archiverMinBackoff
	for {
		if ctx.Err() != nil {
			a.logger.Info("WAL archiver stopped")
			return nil
		}

		start := time.Now()
		err := a.stream(ctx)
		if ctx.Err() != nil {
			a.logger.Info("WAL archiver stopped")
			return nil
		}
		if err == nil {
			backoff = archiverMinBackoff
			continue
		}

		if time.Since(start) > archiverMaxBackoff {
			backoff = archiverMinBackoff
		}
		a.logger.Error("WAL stream interrupted: %v; retrying in %s", err, backoff)

		select {
		case <-ctx.Done():
			a.logger.Info("WAL archiver stopped")
			return nil
		case <-time.After(backoff):
		}
		if backoff < archiverMaxBackoff {
			backoff *= 2
			if backoff > archiverMaxBackoff {
				backoff = archiverMaxBackoff
			}
		}
	}
}

func (a *Archiver) stream(ctx context.Context) error {
	args := connArgs(a.cfg.Replication)
	args = append(args,
		"-D", a.cfg.ArchiveDir,
		"--no-password",
		"-v",
	)
	if a.cfg.SlotName != "" {
		args = append(args, "--slot", a.cfg.SlotName)
	}
	a.logger.Info("Streaming WAL into %s", a.cfg.ArchiveDir)
	_, err := runCommand(ctx, a.cfg.Replication, "pg_receivewal", args...)
	return err
}
