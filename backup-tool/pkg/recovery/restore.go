package recovery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fcstask-backend/backup-tool/pkg/backup"
)

type Restorer struct {
	config    RestoreConfig
	logger    *RestoreLogger
	assembler *FileAssembler
}

func NewRestorer(cfg RestoreConfig, logger *RestoreLogger) *Restorer {
	return &Restorer{
		config:    cfg,
		logger:    logger,
		assembler: NewFileAssembler(logger),
	}
}

func (r *Restorer) Restore(backupIdentifier string) error {
	r.logger.Info("Starting restore process for identifier: %s", backupIdentifier)

	targetDir, err := r.locateBackupDir(backupIdentifier)
	if err != nil {
		return fmt.Errorf("failed to locate backup: %w", err)
	}
	r.logger.Info("Found backup directory: %s", targetDir)

	meta, err := r.loadMetadata(targetDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}
	r.logger.Info("Backup type: %s, timestamp: %v", meta.Type, meta.Timestamp)

	chain, err := r.buildBackupChain(targetDir, meta)
	if err != nil {
		return fmt.Errorf("failed to build backup chain: %w", err)
	}

	for i, item := range chain {
		r.logger.Info("--- Step %d: processing %s backup at %s", i+1, item.Type, item.Dir)

		if err := r.assembler.AssembleFiles(item.Dir); err != nil {
			return fmt.Errorf("failed to assemble files in %s: %w", item.Dir, err)
		}

		if item.Type == backup.FullBackup {
			if err := r.restoreFullBackup(item.Dir); err != nil {
				return fmt.Errorf("full backup restore failed: %w", err)
			}
		} else {
			if err := r.applyIncrementalBackup(item.Dir); err != nil {
				return fmt.Errorf("incremental backup apply failed: %w", err)
			}
		}
	}

	r.logger.Info("Restore completed successfully")
	return nil
}

func (r *Restorer) locateBackupDir(identifier string) (string, error) {
	if identifier == "latest" {
		return r.findLatestBackup()
	}
	fullPath := identifier
	if !filepath.IsAbs(identifier) {
		fullPath = filepath.Join(r.config.BackupRoot, identifier)
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", fullPath)
	}
	return fullPath, nil
}

func (r *Restorer) findLatestBackup() (string, error) {
	entries, err := os.ReadDir(r.config.BackupRoot)
	if err != nil {
		return "", err
	}
	var latestDir string
	var latestTime time.Time
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "full_") && !strings.HasPrefix(name, "incremental_") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestDir = filepath.Join(r.config.BackupRoot, name)
		}
	}
	if latestDir == "" {
		return "", fmt.Errorf("no backup found in %s", r.config.BackupRoot)
	}
	return latestDir, nil
}

func (r *Restorer) loadMetadata(backupDir string) (*backup.BackupMetadata, error) {
	metaPath := filepath.Join(backupDir, "backup_metadata.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var meta backup.BackupMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

type chainItem struct {
	Dir  string
	Type backup.BackupType
	Meta *backup.BackupMetadata
}

func (r *Restorer) buildBackupChain(startDir string, startMeta *backup.BackupMetadata) ([]chainItem, error) {
	var chain []chainItem
	currentDir := startDir
	currentMeta := startMeta

	for {
		chain = append([]chainItem{{Dir: currentDir, Type: currentMeta.Type, Meta: currentMeta}}, chain...)

		if currentMeta.Type == backup.FullBackup {
			break
		}
		if currentMeta.ParentBackup == "" {
			return nil, fmt.Errorf("incremental backup has no parent: %s", currentDir)
		}
		parentDir := filepath.Join(r.config.BackupRoot, currentMeta.ParentBackup)
		parentMeta, err := r.loadMetadata(parentDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load parent metadata %s: %w", parentDir, err)
		}
		currentDir = parentDir
		currentMeta = parentMeta
	}
	return chain, nil
}

func (r *Restorer) restoreFullBackup(backupDir string) error {
	r.logger.Info("Restoring full backup from %s", backupDir)

	if r.config.DropDatabase {
		if err := r.dropAndCreateDatabase(); err != nil {
			return fmt.Errorf("failed to drop/create database: %w", err)
		}
	}

	args := []string{
		"-h", r.config.Target.Host,
		"-p", fmt.Sprintf("%d", r.config.Target.Port),
		"-U", r.config.Target.User,
		"-d", r.config.Target.Database,
		"-F", "d",
		"-j", fmt.Sprintf("%d", r.config.Jobs),
		"-v",
		backupDir,
	}

	env := append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", r.config.Target.Password))
	cmd := exec.Command("pg_restore", args...)
	cmd.Env = env

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	r.logger.Debug("Running pg_restore %s", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_restore failed: %v, stderr: %s", err, stderr.String())
	}
	r.logger.Info("Full backup restored successfully")
	return nil
}

func (r *Restorer) applyIncrementalBackup(backupDir string) error {
	r.logger.Info("Applying incremental backup from %s", backupDir)

	walDir := filepath.Join(backupDir, "wal")
	info, err := os.Stat(walDir)
	if err != nil {
		r.logger.Warn("WAL directory not found in %s, skipping", backupDir)
		return nil
	}
	if !info.IsDir() {
		r.logger.Warn("WAL path is not a directory, skipping")
		return nil
	}

	if r.config.WalDestinationDir == "" {
		r.logger.Info("WAL files are available at %s. To perform PITR, copy them manually.", walDir)
		return nil
	}

	if err := r.copyWalFiles(walDir, r.config.WalDestinationDir); err != nil {
		return fmt.Errorf("failed to copy WAL files: %w", err)
	}
	r.logger.Info("WAL files copied to %s", r.config.WalDestinationDir)
	return nil
}

func (r *Restorer) copyWalFiles(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if err := r.copyFile(srcPath, dstPath); err != nil {
			r.logger.Warn("Failed to copy %s: %v", e.Name(), err)
		}
	}
	return nil
}

func (r *Restorer) copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func (r *Restorer) dropAndCreateDatabase() error {
	r.logger.Info("Dropping and recreating database %s", r.config.Target.Database)

	dropCmd := exec.Command("psql",
		"-h", r.config.Target.Host,
		"-p", fmt.Sprintf("%d", r.config.Target.Port),
		"-U", r.config.Target.User,
		"-d", "postgres",
		"-c", fmt.Sprintf("DROP DATABASE IF EXISTS %s;", r.config.Target.Database),
	)
	dropCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", r.config.Target.Password))
	if out, err := dropCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("drop database failed: %v, output: %s", err, out)
	}

	createCmd := exec.Command("psql",
		"-h", r.config.Target.Host,
		"-p", fmt.Sprintf("%d", r.config.Target.Port),
		"-U", r.config.Target.User,
		"-d", "postgres",
		"-c", fmt.Sprintf("CREATE DATABASE %s;", r.config.Target.Database),
	)
	createCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", r.config.Target.Password))
	if out, err := createCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("create database failed: %v, output: %s", err, out)
	}
	return nil
}
