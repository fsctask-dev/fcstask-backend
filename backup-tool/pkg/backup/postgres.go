package backup

import (
	"bytes"
	"encoding/json"
	"fcstask-backend/backup-tool/config"
	"fcstask-backend/backup-tool/pkg/logging"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type BackupType string

const (
	FullBackup        BackupType = "full"
	IncrementalBackup BackupType = "incremental"
)

type BackupMetadata struct {
	Type         BackupType `json:"type"`
	Timestamp    time.Time  `json:"timestamp"`
	ParentBackup string     `json:"parent_backup,omitempty"`
	WalStart     string     `json:"wal_start,omitempty"`
	WalEnd       string     `json:"wal_end,omitempty"`
	Size         int64      `json:"size"`
	FileCount    int        `json:"file_count"`
}

type Backuper struct {
	config          config.PostgresConfig
	backup          config.BackupConfig
	logger          *logging.Logger
	incremental     bool
	lastBackupDir   string
	lastWalPosition string
}

func NewBackuper(pgConfig config.PostgresConfig, backupConfig config.BackupConfig, logger *logging.Logger) *Backuper {
	return &Backuper{
		config:          pgConfig,
		backup:          backupConfig,
		logger:          logger,
		incremental:     false,
		lastWalPosition: "",
	}
}

func (b *Backuper) SetIncrementalMode(enabled bool) {
	b.incremental = enabled
}

func (b *Backuper) CreateBackup() error {
	timestamp := time.Now().Format("20060102_150405")

	backupType := FullBackup
	if b.incremental && b.findLastBackup() {
		backupType = IncrementalBackup
	}

	backupDir := filepath.Join(b.backup.OutputDir, fmt.Sprintf("%s_%s", backupType, timestamp))

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	b.logger.Info("Creating %s backup in directory: %s", backupType, backupDir)

	metadata := BackupMetadata{
		Type:      backupType,
		Timestamp: time.Now(),
	}

	var err error
	if backupType == FullBackup {
		err = b.createFullBackup(backupDir, &metadata)
	} else {
		err = b.createIncrementalBackup(backupDir, &metadata)
	}

	if err != nil {
		return err
	}

	if err := b.saveMetadata(backupDir, metadata); err != nil {
		b.logger.Warn("Failed to save metadata: %v", err)
	}

	if err := b.splitLargeFiles(backupDir); err != nil {
		b.logger.Warn("Failed to split files: %v", err)
	}

	b.cleanOldBackups()
	b.lastBackupDir = backupDir

	return nil
}

func (b *Backuper) createFullBackup(backupDir string, metadata *BackupMetadata) error {
	cmdArgs := []string{
		"-h", b.config.Host,
		"-p", fmt.Sprintf("%d", b.config.Port),
		"-U", b.config.User,
		"-d", b.config.Database,
		"-F", "d",
		"-f", backupDir,
		"-j", "4",
		"-v",
	}

	env := append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", b.config.Password))

	cmd := exec.Command("pg_dump", cmdArgs...)
	cmd.Env = env

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	b.logger.Info("Starting full pg_dump...")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failed: %v, stderr: %s", err, stderr.String())
	}

	size, fileCount, err := b.getBackupInfo(backupDir)
	if err == nil {
		metadata.Size = size
		metadata.FileCount = fileCount
	}

	walPos, err := b.getCurrentWalPosition()
	if err == nil {
		metadata.WalEnd = walPos
		b.lastWalPosition = walPos
	}

	b.logger.Info("Full backup completed successfully. WAL position: %s", b.lastWalPosition)
	return nil
}

func (b *Backuper) createIncrementalBackup(backupDir string, metadata *BackupMetadata) error {
	walPos, err := b.getCurrentWalPosition()
	if err != nil {
		b.logger.Warn("Failed to get WAL position, falling back to full backup: %v", err)
		return b.createFullBackup(backupDir, metadata)
	}

	if b.lastWalPosition == "" {
		b.logger.Warn("No previous WAL position found, falling back to full backup")
		return b.createFullBackup(backupDir, metadata)
	}

	metadata.WalStart = b.lastWalPosition
	metadata.WalEnd = walPos
	metadata.ParentBackup = filepath.Base(b.lastBackupDir)

	if err := b.copyWalFiles(backupDir, b.lastWalPosition, walPos); err != nil {
		b.logger.Warn("Failed to copy WAL files, falling back to full backup: %v", err)
		return b.createFullBackup(backupDir, metadata)
	}

	if err := b.saveSchemaChanges(backupDir); err != nil {
		b.logger.Warn("Failed to save schema changes: %v", err)
	}

	metadata.Type = IncrementalBackup

	size, fileCount, err := b.getBackupInfo(backupDir)
	if err == nil {
		metadata.Size = size
		metadata.FileCount = fileCount
	}

	b.lastWalPosition = walPos

	b.logger.Info("Incremental backup completed successfully. WAL range: %s -> %s",
		metadata.WalStart, metadata.WalEnd)
	return nil
}

func (b *Backuper) getCurrentWalPosition() (string, error) {
	cmd := exec.Command("psql",
		"-h", b.config.Host,
		"-p", fmt.Sprintf("%d", b.config.Port),
		"-U", b.config.User,
		"-d", b.config.Database,
		"-t",
		"-c", "SELECT pg_current_wal_lsn();")

	env := append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", b.config.Password))
	cmd.Env = env

	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("psql",
			"-h", b.config.Host,
			"-p", fmt.Sprintf("%d", b.config.Port),
			"-U", b.config.User,
			"-d", b.config.Database,
			"-t",
			"-c", "SELECT pg_current_xlog_location();")
		cmd.Env = env
		output, err = cmd.Output()
		if err != nil {
			return "", err
		}
	}

	return strings.TrimSpace(string(output)), nil
}

func (b *Backuper) copyWalFiles(destDir string, startLsn, endLsn string) error {
	destWalDir := filepath.Join(destDir, "wal")
	if err := os.MkdirAll(destWalDir, 0755); err != nil {
		return err
	}

	walArchiveDir := filepath.Join(b.backup.OutputDir, "wal_archive")
	if err := os.MkdirAll(walArchiveDir, 0755); err != nil {
		b.logger.Warn("Failed to create WAL archive directory: %v", err)
	}

	files, err := b.getWalFilesFromRange(startLsn, endLsn)
	if err != nil {
		b.logger.Warn("Failed to get WAL files via pg_waldump: %v", err)
		return b.copyAllWalFiles(walArchiveDir, destWalDir)
	}

	copiedCount := 0
	for _, walFile := range files {
		srcPath := filepath.Join(walArchiveDir, walFile)
		if _, err := os.Stat(srcPath); err != nil {
			b.logger.Debug("WAL file not found in archive: %s", walFile)
			continue
		}

		dstPath := filepath.Join(destWalDir, walFile)
		if err := copyFile(srcPath, dstPath); err != nil {
			b.logger.Warn("Failed to copy WAL file %s: %v", walFile, err)
			continue
		}
		copiedCount++
	}

	b.logger.Info("Copied %d WAL files for range %s - %s", copiedCount, startLsn, endLsn)
	return nil
}

func (b *Backuper) getWalFilesFromRange(startLsn, endLsn string) ([]string, error) {
	walArchiveDir := filepath.Join(b.backup.OutputDir, "wal_archive")

	cmd := exec.Command("pg_waldump",
		"--path", walArchiveDir,
		"--start", startLsn,
		"--end", endLsn,
		"--quiet",
		"-n", "1000")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			fileName := fields[0]
			if strings.Contains(fileName, ".partial") {
				fileName = strings.Split(fileName, ".partial")[0]
			}
			if !strings.Contains(fileName, ".history") {
				files = append(files, fileName)
			}
		}
	}

	return files, nil
}

func (b *Backuper) copyAllWalFiles(srcDir, destDir string) error {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		b.logger.Debug("WAL archive directory does not exist: %s", srcDir)
		return nil
	}

	files, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	copiedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if len(name) == 24 && (strings.HasPrefix(name, "0000") || strings.HasPrefix(name, "0001")) {
			srcPath := filepath.Join(srcDir, name)
			dstPath := filepath.Join(destDir, name)
			if err := copyFile(srcPath, dstPath); err == nil {
				copiedCount++
			}
		}
	}

	b.logger.Info("Copied %d WAL files from archive", copiedCount)
	return nil
}

func (b *Backuper) saveSchemaChanges(backupDir string) error {
	schemaFile := filepath.Join(backupDir, "schema.sql")

	cmd := exec.Command("pg_dump",
		"-h", b.config.Host,
		"-p", fmt.Sprintf("%d", b.config.Port),
		"-U", b.config.User,
		"-d", b.config.Database,
		"-s",
		"-f", schemaFile)

	env := append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", b.config.Password))
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to dump schema: %v", err)
	}

	extFile := filepath.Join(backupDir, "extensions.sql")
	extCmd := exec.Command("psql",
		"-h", b.config.Host,
		"-p", fmt.Sprintf("%d", b.config.Port),
		"-U", b.config.User,
		"-d", b.config.Database,
		"-t",
		"-c", "SELECT 'CREATE EXTENSION IF NOT EXISTS \"' || extname || '\";' FROM pg_extension;",
		"-o", extFile)

	extCmd.Env = env
	extCmd.Run()

	return nil
}

func (b *Backuper) findLastBackup() bool {
	entries, err := os.ReadDir(b.backup.OutputDir)
	if err != nil {
		b.logger.Debug("Cannot read backup directory: %v", err)
		return false
	}

	var lastBackup string
	var lastTime time.Time

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, "full_") || strings.HasPrefix(name, "incremental_") {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.ModTime().After(lastTime) {
				lastTime = info.ModTime()
				lastBackup = name
			}
		}
	}

	if lastBackup != "" {
		b.lastBackupDir = filepath.Join(b.backup.OutputDir, lastBackup)
		b.logger.Info("Found last backup: %s", lastBackup)

		metadata, err := b.loadMetadata(b.lastBackupDir)
		if err == nil && metadata.WalEnd != "" {
			b.lastWalPosition = metadata.WalEnd
			b.logger.Info("Last WAL position from metadata: %s", b.lastWalPosition)
			return true
		}

		b.logger.Debug("No WAL position in metadata, will use full backup")
		return false
	}

	return false
}

func (b *Backuper) saveMetadata(backupDir string, metadata BackupMetadata) error {
	metadataFile := filepath.Join(backupDir, "backup_metadata.json")

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataFile, data, 0644)
}

func (b *Backuper) loadMetadata(backupDir string) (*BackupMetadata, error) {
	metadataFile := filepath.Join(backupDir, "backup_metadata.json")

	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, err
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (b *Backuper) getBackupInfo(backupDir string) (int64, int, error) {
	var totalSize int64
	var fileCount int

	err := filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++
		}
		return nil
	})

	return totalSize, fileCount, err
}

func (b *Backuper) cleanOldBackups() {
	if b.backup.RetentionDays <= 0 {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -b.backup.RetentionDays)

	entries, err := os.ReadDir(b.backup.OutputDir)
	if err != nil {
		b.logger.Error("Failed to read backup directory: %v", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, "full_") && !strings.HasPrefix(name, "incremental_") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			path := filepath.Join(b.backup.OutputDir, entry.Name())
			b.logger.Info("Removing old backup: %s", path)
			if err := os.RemoveAll(path); err != nil {
				b.logger.Warn("Failed to remove old backup: %v", err)
			}
		}
	}
}

func (b *Backuper) splitLargeFiles(backupDir string) error {
	if b.backup.SplitSizeMB <= 0 {
		return nil
	}

	splitter := NewFileSplitter(b.backup.SplitSizeMB, b.logger)

	return filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.Contains(path, ".part") || strings.HasSuffix(path, "backup_metadata.json") {
			return nil
		}

		maxBytes := int64(b.backup.SplitSizeMB) * (1 << 20)
		if info.Size() > maxBytes {
			b.logger.Info("Splitting large file: %s (%.2f MB)",
				path, float64(info.Size())/(1<<20))
			return splitter.SplitFile(path)
		}

		return nil
	})
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}
