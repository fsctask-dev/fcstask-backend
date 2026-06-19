package wal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"fcstask-backend/backup-service/internal/config"
	"fcstask-backend/backup-service/internal/logging"
)

type BaseMetadata struct {
	ID           string    `json:"id"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	StartWALFile string    `json:"start_wal_file"`
	Host         string    `json:"host"`
	Port         int       `json:"port"`
}

type BaseBackuper struct {
	cfg    config.WALConfig
	logger *logging.Logger
}

func NewBaseBackuper(cfg config.WALConfig, logger *logging.Logger) *BaseBackuper {
	return &BaseBackuper{cfg: cfg, logger: logger}
}

func (b *BaseBackuper) TakeBaseBackup(ctx context.Context) (*BaseMetadata, error) {
	if !toolAvailable("pg_basebackup") {
		return nil, fmt.Errorf("pg_basebackup not found on PATH")
	}
	if err := os.MkdirAll(b.cfg.BaseBackupDir, 0o700); err != nil {
		return nil, fmt.Errorf("create base backup root: %w", err)
	}

	started := time.Now().UTC()
	id := basePrefix + started.Format("20060102_150405")
	dest := filepath.Join(b.cfg.BaseBackupDir, id)
	if _, err := os.Stat(dest); err == nil {
		return nil, fmt.Errorf("base backup directory already exists: %s", dest)
	}

	args := connArgs(b.cfg.Replication)
	args = append(args,
		"-D", dest,
		"-F", "p",
		"-X", "stream",
		"--no-password",
		"-l", id,
	)
	if b.cfg.FastCheckpoint {
		args = append(args, "-c", "fast")
	}

	b.logger.Info("Running pg_basebackup into %s", dest)
	if _, err := runCommand(ctx, b.cfg.Replication, "pg_basebackup", args...); err != nil {

		_ = os.RemoveAll(dest)
		return nil, err
	}

	meta := &BaseMetadata{
		ID:         id,
		StartedAt:  started,
		FinishedAt: time.Now().UTC(),
		Host:       b.cfg.Replication.Host,
		Port:       b.cfg.Replication.Port,
	}
	if wal, err := parseStartWALFile(filepath.Join(dest, "backup_label")); err == nil {
		meta.StartWALFile = wal
	} else {
		b.logger.Warn("Could not parse backup_label start WAL file: %v", err)
	}

	if err := writeBaseMetadata(dest, meta); err != nil {
		return nil, fmt.Errorf("write base metadata: %w", err)
	}
	b.logger.Info("Base backup %s completed (start WAL %s)", id, meta.StartWALFile)

	if err := b.cleanup(ctx); err != nil {
		b.logger.Warn("WAL retention cleanup failed: %v", err)
	}
	return meta, nil
}

func ListBaseBackups(root string) ([]*BaseMetadata, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []*BaseMetadata
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m, err := readBaseMetadata(filepath.Join(root, e.Name()))
		if err != nil {
			continue
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.Before(out[j].StartedAt) })
	return out, nil
}

func (b *BaseBackuper) cleanup(ctx context.Context) error {
	if b.cfg.RetentionDays <= 0 {
		return nil
	}
	bases, err := ListBaseBackups(b.cfg.BaseBackupDir)
	if err != nil {
		return err
	}
	if len(bases) == 0 {
		return nil
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -b.cfg.RetentionDays)

	newest := bases[len(bases)-1]
	for _, m := range bases {
		if m.ID == newest.ID {
			continue
		}
		if m.FinishedAt.Before(cutoff) {
			path := filepath.Join(b.cfg.BaseBackupDir, m.ID)
			b.logger.Info("Removing expired base backup: %s", m.ID)
			if err := os.RemoveAll(path); err != nil {
				b.logger.Warn("Failed to remove %s: %v", path, err)
			}
		}
	}

	remaining, err := ListBaseBackups(b.cfg.BaseBackupDir)
	if err != nil || len(remaining) == 0 {
		return err
	}
	oldest := remaining[0]
	if oldest.StartWALFile == "" {
		return nil
	}
	if !toolAvailable("pg_archivecleanup") {
		b.logger.Warn("pg_archivecleanup not found; skipping WAL archive trim")
		return nil
	}
	b.logger.Info("Trimming WAL archive before %s", oldest.StartWALFile)
	_, err = runCommand(ctx, b.cfg.Replication, "pg_archivecleanup", b.cfg.ArchiveDir, oldest.StartWALFile)
	return err
}

func writeBaseMetadata(dir string, m *BaseMetadata) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, baseMetadataFile+".tmp")
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, baseMetadataFile))
}

func readBaseMetadata(dir string) (*BaseMetadata, error) {
	data, err := os.ReadFile(filepath.Join(dir, baseMetadataFile))
	if err != nil {
		return nil, err
	}
	var m BaseMetadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
