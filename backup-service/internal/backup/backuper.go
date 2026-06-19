package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fcstask-backend/backup-service/internal/config"
	"fcstask-backend/backup-service/internal/logging"
	"fcstask-backend/backup-service/internal/pg"
	"fcstask-backend/backup-service/internal/sqlutil"
)

type BackupDecision string

const (
	DecisionFull        BackupDecision = "full"
	DecisionIncremental BackupDecision = "incremental"
)

type Backuper struct {
	src              *pg.Client
	cfg              config.BackupConfig
	pitr             config.PITRConfig
	logger           *logging.Logger
	forceFull        bool
	forceIncremental bool
}

func NewBackuper(src *pg.Client, cfg config.BackupConfig, pitr config.PITRConfig, logger *logging.Logger) *Backuper {
	return &Backuper{src: src, cfg: cfg, pitr: pitr, logger: logger}
}

func (b *Backuper) ForceFull(v bool) { b.forceFull = v }

func (b *Backuper) ForceIncremental(v bool) { b.forceIncremental = v }

func (b *Backuper) CreateBackup(ctx context.Context) (*Metadata, error) {
	if err := os.MkdirAll(b.cfg.OutputDir, 0o700); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	if err := b.src.CheckDumpCompatibility(ctx); err != nil {
		return nil, err
	}

	existing, err := ListBackups(b.cfg.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("list existing backups: %w", err)
	}

	snapshot, err := b.dbNow(ctx)
	if err != nil {
		return nil, fmt.Errorf("read server time: %w", err)
	}

	decision, parent := b.decideBackup(existing)

	typ := FullBackup
	if decision == DecisionIncremental {
		typ = IncrementalBackup
	}

	id := fmt.Sprintf("%s_%s", typ, snapshot.Format("20060102_150405"))
	dir := filepath.Join(b.cfg.OutputDir, id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create backup dir %q: %w", dir, err)
	}

	meta := &Metadata{
		Type:         typ,
		ID:           id,
		Database:     b.src.Database(),
		StartedAt:    time.Now().UTC(),
		SnapshotTime: snapshot,
	}

	b.logger.Info("Creating %s backup %s", typ, id)

	switch typ {
	case FullBackup:
		meta.BaseFull = id
		err = b.createFull(ctx, dir)
	default:
		meta.Parent = parent.ID
		meta.BaseFull = parent.BaseFull
		meta.Since = parent.SnapshotTime
		err = b.createIncremental(ctx, dir, meta, parent.SnapshotTime, snapshot)
	}
	if err != nil {
		return nil, fmt.Errorf("%s backup failed: %w", typ, err)
	}

	if err := b.splitLargeFiles(dir, meta); err != nil {
		return nil, fmt.Errorf("split files: %w", err)
	}

	if err := WriteChecksums(dir); err != nil {
		return nil, fmt.Errorf("write checksums: %w", err)
	}

	if size, count, derr := dirSize(dir); derr == nil {
		meta.SizeBytes = size
		meta.FileCount = count
	}
	meta.FinishedAt = time.Now().UTC()

	if err := writeMetadata(dir, meta); err != nil {
		return nil, fmt.Errorf("write metadata: %w", err)
	}

	if err := b.cleanOldBackups(); err != nil {
		b.logger.Warn("Retention cleanup failed: %v", err)
	}

	b.logger.Info("Backup %s completed (%d files, %.2f MiB)", id, meta.FileCount, float64(meta.SizeBytes)/(1<<20))
	return meta, nil
}

func (b *Backuper) decideBackup(backups []*Metadata) (BackupDecision, *Metadata) {
	if b.forceFull {
		return DecisionFull, nil
	}

	var last *Metadata
	if len(backups) > 0 {
		last = backups[len(backups)-1]
	}

	if b.forceIncremental {
		if last == nil {
			return DecisionFull, nil
		}
		return DecisionIncremental, last
	}

	if !b.pitr.Enabled || b.cfg.FullBackupEvery <= 1 {
		return DecisionFull, nil
	}
	if last == nil {
		return DecisionFull, nil
	}

	chainLen := 0
	for _, m := range backups {
		if m.BaseFull == last.BaseFull {
			chainLen++
		}
	}
	if chainLen >= b.cfg.FullBackupEvery {
		return DecisionFull, nil
	}
	return DecisionIncremental, last
}

func (b *Backuper) createFull(ctx context.Context, dir string) error {
	dataDir := filepath.Join(dir, dataSubdir)
	b.logger.Info("Running pg_dump (directory format, %d jobs)", b.cfg.PgDumpJobs)
	return b.src.PgDumpDirectory(ctx, dataDir, b.cfg.PgDumpJobs)
}

func (b *Backuper) createIncremental(ctx context.Context, dir string, meta *Metadata, since, until time.Time) error {
	tables, skipped, err := b.src.DiscoverCDCTables(ctx,
		b.pitr.Schemas, b.pitr.CreatedAtColumns, b.pitr.UpdatedAtColumns, b.pitr.DeletedAtColumns)
	if err != nil {
		return err
	}
	for _, s := range skipped {
		b.logger.Warn("Skipping table for CDC: %s", s)
	}
	if len(tables) == 0 {
		return fmt.Errorf("no CDC-eligible tables found; cannot create incremental backup")
	}

	cdcDir := filepath.Join(dir, cdcSubdir)
	if err := os.MkdirAll(cdcDir, 0o700); err != nil {
		return err
	}

	sinceLit := sqlTimestamp(since)
	untilLit := sqlTimestamp(until)

	refs := make([]pg.TableRef, len(tables))
	cols := make([][]string, len(tables))
	files := make([]string, len(tables))
	copies := make([]pg.Copy, len(tables))
	for i, t := range tables {
		ref := pg.TableRef{
			Schema: t.Schema, Name: t.Name,
			CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt, DeletedAt: t.DeletedAt,
		}
		refs[i] = ref

		tcols, err := b.src.Columns(ctx, t.Schema, t.Name)
		if err != nil {
			return err
		}
		if len(tcols) == 0 {
			return fmt.Errorf("table %s.%s has no columns", t.Schema, t.Name)
		}
		cols[i] = tcols

		changeExpr := ref.ChangeExpr()
		fileName := fmt.Sprintf("%03d_%s.%s.csv", i, safeComponent(t.Schema), safeComponent(t.Name))
		files[i] = fileName
		copies[i] = pg.Copy{
			SelectSQL: fmt.Sprintf("SELECT %s FROM %s WHERE %s > %s AND %s <= %s",
				sqlutil.JoinIdents(tcols), ref.Qualified(), changeExpr, sinceLit, changeExpr, untilLit),
			DestPath: filepath.Join(cdcDir, fileName),
		}
	}

	counts, err := b.src.RunCopyTransaction(ctx, copies)
	if err != nil {
		return fmt.Errorf("capture incremental snapshot: %w", err)
	}

	for i, t := range tables {
		ref := refs[i]
		b.logger.Info("CDC %s.%s: %d changed rows", t.Schema, t.Name, counts[i])
		meta.Tables = append(meta.Tables, TableInfo{
			Schema:     t.Schema,
			Name:       t.Name,
			CreatedAt:  t.CreatedAt,
			UpdatedAt:  t.UpdatedAt,
			DeletedAt:  t.DeletedAt,
			ChangeExpr: ref.ChangeExpr(),
			Columns:    cols[i],
			DataFile:   filepath.Join(cdcSubdir, files[i]),
			RowCount:   counts[i],
		})
	}
	return nil
}

func (b *Backuper) dbNow(ctx context.Context) (time.Time, error) {
	s, err := b.src.QueryScalar(ctx, "SELECT extract(epoch from now())", nil)
	if err != nil {
		return time.Time{}, err
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse server epoch %q: %w", s, err)
	}
	sec := int64(f)
	nsec := int64((f - float64(sec)) * 1e9)
	return time.Unix(sec, nsec).UTC(), nil
}

func (b *Backuper) splitLargeFiles(dir string, meta *Metadata) error {
	if b.cfg.SplitSizeMB <= 0 {
		return nil
	}
	maxBytes := int64(b.cfg.SplitSizeMB) * (1 << 20)
	splitter := NewFileSplitter(b.cfg.SplitSizeMB, b.logger)

	var toSplit []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || IsPart(path) || strings.HasSuffix(path, metadataFile) {
			return nil
		}
		if info.Size() > maxBytes {
			toSplit = append(toSplit, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, path := range toSplit {
		sum, size, err := fileSHA256(path)
		if err != nil {
			return fmt.Errorf("checksum %q: %w", path, err)
		}
		b.logger.Info("Splitting %s (%.2f MiB)", path, float64(size)/(1<<20))
		parts, err := splitter.SplitFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		meta.Splits = append(meta.Splits, SplitInfo{
			File:   filepath.ToSlash(rel),
			Parts:  parts,
			Size:   size,
			SHA256: sum,
		})
	}
	return nil
}

func (b *Backuper) cleanOldBackups() error {
	if b.cfg.RetentionDays <= 0 {
		return nil
	}
	backups, err := ListBackups(b.cfg.OutputDir)
	if err != nil {
		return err
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -b.cfg.RetentionDays)

	newest := map[string]time.Time{}
	members := map[string][]string{}
	for _, m := range backups {
		members[m.BaseFull] = append(members[m.BaseFull], m.ID)
		if m.SnapshotTime.After(newest[m.BaseFull]) {
			newest[m.BaseFull] = m.SnapshotTime
		}
	}
	for base, latest := range newest {
		if latest.Before(cutoff) {
			for _, id := range members[base] {
				path := filepath.Join(b.cfg.OutputDir, id)
				b.logger.Info("Removing expired backup: %s", id)
				if err := os.RemoveAll(path); err != nil {
					b.logger.Warn("Failed to remove %s: %v", path, err)
				}
			}
		}
	}
	return nil
}

func sqlTimestamp(t time.Time) string {
	return "'" + t.UTC().Format("2006-01-02 15:04:05.999999") + "+00'::timestamptz"
}
