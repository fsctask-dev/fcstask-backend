package wal

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fcstask-backend/backup-service/internal/config"
	"fcstask-backend/backup-service/internal/logging"
)

type Restorer struct {
	cfg    config.WALConfig
	logger *logging.Logger
}

func NewRestorer(cfg config.WALConfig, logger *logging.Logger) *Restorer {
	return &Restorer{cfg: cfg, logger: logger}
}

func (r *Restorer) RestoreToTime(ctx context.Context, targetTime *time.Time) error {
	if r.cfg.RestoreDataDir == "" {
		return fmt.Errorf("wal.restore_data_dir is not configured")
	}

	base, err := r.selectBase(targetTime)
	if err != nil {
		return err
	}
	if targetTime != nil {
		r.logger.Info("Physical recovery to %s", targetTime.UTC().Format(time.RFC3339))
	} else {
		r.logger.Info("Physical recovery to latest")
	}
	r.logger.Info("Using base backup %s (%s)", base.ID, base.StartedAt.Format(time.RFC3339))

	if entries, err := os.ReadDir(r.cfg.ArchiveDir); err != nil || len(entries) == 0 {
		r.logger.Warn("WAL archive %q is missing or empty; recovery may not reach the target time", r.cfg.ArchiveDir)
	}
	r.checkArchiveCoverage(base)
	if !toolAvailable("pg_archivecleanup") {
		r.logger.Warn("pg_archivecleanup not found on PATH; WAL archive retention will not run")
	}

	if err := r.materializeDataDir(base); err != nil {
		return err
	}
	if err := r.writeRecoveryConfig(targetTime); err != nil {
		return err
	}

	if r.cfg.PgCtlPath == "" {
		r.logger.Info("Recovered data directory prepared at %s", r.cfg.RestoreDataDir)
		r.logger.Info("Start it with: %s -D %s -o \"-p %d\" start",
			"pg_ctl", r.cfg.RestoreDataDir, r.restorePort())
		return nil
	}
	return r.startInstance(ctx)
}

func (r *Restorer) checkArchiveCoverage(base *BaseMetadata) {
	if base.StartWALFile == "" {
		r.logger.Warn("Base backup %s has no recorded start WAL file; cannot verify archive coverage, recovery may fail to reach the target", base.ID)
		return
	}
	entries, err := os.ReadDir(r.cfg.ArchiveDir)
	if err != nil {
		r.logger.Warn("Cannot read WAL archive %q to verify coverage of start segment %s: %v", r.cfg.ArchiveDir, base.StartWALFile, err)
		return
	}
	for _, e := range entries {
		if strings.TrimSuffix(e.Name(), ".partial") == base.StartWALFile {
			r.logger.Info("Verified start WAL segment %s is present in the archive", base.StartWALFile)
			return
		}
	}
	r.logger.Warn("GAP DETECTED: start WAL segment %s for base backup %s is NOT present in archive %q. "+
		"Recovery cannot replay from the base and will almost certainly fail. "+
		"Ensure the WAL archiver and replication slot were running before the base backup was taken.",
		base.StartWALFile, base.ID, r.cfg.ArchiveDir)
}

func (r *Restorer) selectBase(targetTime *time.Time) (*BaseMetadata, error) {
	bases, err := ListBaseBackups(r.cfg.BaseBackupDir)
	if err != nil {
		return nil, err
	}
	if len(bases) == 0 {
		return nil, fmt.Errorf("no base backups found in %q", r.cfg.BaseBackupDir)
	}
	var base *BaseMetadata
	for _, m := range bases {
		if targetTime == nil || !m.StartedAt.After(*targetTime) {
			base = m
		}
	}
	if base == nil {
		return nil, fmt.Errorf("no base backup at or before the requested time; earliest base is %s",
			bases[0].StartedAt.Format(time.RFC3339))
	}
	return base, nil
}

func (r *Restorer) materializeDataDir(base *BaseMetadata) error {
	dst := r.cfg.RestoreDataDir
	if entries, err := os.ReadDir(dst); err == nil && len(entries) > 0 {
		return fmt.Errorf("restore_data_dir %q is not empty; refusing to overwrite", dst)
	}
	src := filepath.Join(r.cfg.BaseBackupDir, base.ID)

	r.logger.Info("Copying base backup into %s", dst)
	if err := copyTree(src, dst); err != nil {
		return fmt.Errorf("copy base backup: %w", err)
	}

	if err := os.Chmod(dst, 0o700); err != nil {
		return fmt.Errorf("chmod data dir: %w", err)
	}

	_ = os.Remove(filepath.Join(dst, baseMetadataFile))

	if entries, err := os.ReadDir(filepath.Join(dst, "pg_tblspc")); err == nil {
		for _, e := range entries {
			if e.Name() == "." || e.Name() == ".." {
				continue
			}
			r.logger.Warn("Base backup references tablespaces (pg_tblspc/%s); their symlinks are copied verbatim and likely need manual remapping before the recovered instance will start", e.Name())
		}
	}
	if _, err := os.Stat(filepath.Join(dst, "tablespace_map")); err == nil {
		r.logger.Warn("tablespace_map present: recovered instance requires tablespace directories to exist; review before starting")
	}
	return nil
}

func (r *Restorer) writeRecoveryConfig(targetTime *time.Time) error {
	archive := r.cfg.ArchiveDir
	var b strings.Builder
	b.WriteString("\n# --- added by backup-service WAL restore ---\n")

	fmt.Fprintf(&b, "restore_command = 'cp \"%s/%%f\" \"%%p\"'\n", escapeConf(archive))
	if targetTime != nil {
		fmt.Fprintf(&b, "recovery_target_time = '%s'\n", targetTime.UTC().Format("2006-01-02 15:04:05+00"))

		b.WriteString("recovery_target_inclusive = true\n")
	}

	b.WriteString("recovery_target_action = 'promote'\n")
	b.WriteString("recovery_target_timeline = 'latest'\n")

	confPath := filepath.Join(r.cfg.RestoreDataDir, "postgresql.auto.conf")
	f, err := os.OpenFile(confPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open postgresql.auto.conf: %w", err)
	}
	if _, err := f.WriteString(b.String()); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	signal := filepath.Join(r.cfg.RestoreDataDir, recoverySignal)
	if err := os.WriteFile(signal, nil, 0o600); err != nil {
		return fmt.Errorf("create recovery.signal: %w", err)
	}
	r.logger.Info("Recovery configuration written (restore_command + recovery target)")
	return nil
}

func (r *Restorer) startInstance(ctx context.Context) error {
	port := r.restorePort()
	r.logger.Info("Starting recovered instance on port %d", port)
	_, err := runCommand(ctx, r.cfg.Replication, r.cfg.PgCtlPath,
		"-D", r.cfg.RestoreDataDir,
		"-o", fmt.Sprintf("-p %d", port),
		"-w",
		"start",
	)
	if err != nil {
		return fmt.Errorf("start recovered instance: %w", err)
	}
	r.logger.Info("Recovered instance started; it will replay WAL and promote at the target")
	return nil
}

func (r *Restorer) restorePort() int {
	if r.cfg.RestorePort > 0 {
		return r.cfg.RestorePort
	}
	return 5433
}

func escapeConf(s string) string { return strings.ReplaceAll(s, "'", "''") }

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if info.Mode()&os.ModeSymlink != 0 {

			link, lerr := os.Readlink(path)
			if lerr != nil {
				return lerr
			}
			return os.Symlink(link, target)
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
