package restore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fcstask-backend/backup-service/internal/backup"
	"fcstask-backend/backup-service/internal/config"
	"fcstask-backend/backup-service/internal/logging"
	"fcstask-backend/backup-service/internal/pg"
	"fcstask-backend/backup-service/internal/sqlutil"
)

type Restorer struct {
	target    *pg.Client
	cfg       config.RestoreConfig
	pitr      config.PITRConfig
	outputDir string
	logger    *logging.Logger
	assembler *FileAssembler
}

func NewRestorer(target *pg.Client, cfg config.RestoreConfig, pitr config.PITRConfig, outputDir string, logger *logging.Logger) *Restorer {
	return &Restorer{
		target:    target,
		cfg:       cfg,
		pitr:      pitr,
		outputDir: outputDir,
		logger:    logger,
		assembler: NewFileAssembler(logger),
	}
}

func (r *Restorer) RestoreToTime(ctx context.Context, targetTime *time.Time) error {
	base, incrementals, err := r.selectChain(targetTime)
	if err != nil {
		return err
	}

	if targetTime != nil {
		r.logger.Info("Restoring to point in time %s", targetTime.UTC().Format(time.RFC3339))
	} else {
		r.logger.Info("Restoring to latest state")
	}
	r.logger.Info("Base full backup: %s (%s)", base.ID, base.SnapshotTime.Format(time.RFC3339))
	r.logger.Info("Incrementals to apply: %d", len(incrementals))

	if err := r.prepareTarget(ctx); err != nil {
		return fmt.Errorf("prepare target database: %w", err)
	}
	if err := r.restoreBase(ctx, base); err != nil {
		return fmt.Errorf("restore base backup %s: %w", base.ID, err)
	}

	bound := boundValue(targetTime)
	for i, inc := range incrementals {
		r.logger.Info("Applying incremental %d/%d: %s", i+1, len(incrementals), inc.ID)
		if err := r.applyIncremental(ctx, inc, bound); err != nil {
			return fmt.Errorf("apply incremental %s: %w", inc.ID, err)
		}
	}

	if targetTime != nil {
		if err := r.sweepSoftDeletes(ctx, bound); err != nil {
			return fmt.Errorf("apply soft-delete sweep: %w", err)
		}
	}

	r.logger.Info("Restore completed successfully")
	return nil
}

func (r *Restorer) verifyIntegrity(dir string) error {
	err := backup.VerifyChecksums(dir)
	switch {
	case err == nil:
		r.logger.Info("Integrity check passed for %s", filepath.Base(dir))
		return nil
	case errors.Is(err, backup.ErrNoChecksums):
		r.logger.Warn("No checksums recorded for %s; skipping integrity verification", filepath.Base(dir))
		return nil
	default:
		return fmt.Errorf("integrity verification failed for %s: %w", filepath.Base(dir), err)
	}
}

func (r *Restorer) selectChain(targetTime *time.Time) (*backup.Metadata, []*backup.Metadata, error) {
	backups, err := backup.ListBackups(r.outputDir)
	if err != nil {
		return nil, nil, err
	}
	if len(backups) == 0 {
		return nil, nil, fmt.Errorf("no backups found in %q", r.outputDir)
	}

	atOrBefore := func(t time.Time) bool { return targetTime == nil || !t.After(*targetTime) }

	var base *backup.Metadata
	for _, m := range backups {
		if m.Type == backup.FullBackup && atOrBefore(m.SnapshotTime) {
			base = m
		}
	}
	if base == nil {
		return nil, nil, fmt.Errorf("no full backup at or before the requested time; earliest backup is %s",
			backups[0].SnapshotTime.Format(time.RFC3339))
	}

	var incrementals []*backup.Metadata
	for _, m := range backups {
		if m.Type != backup.IncrementalBackup || m.BaseFull != base.ID {
			continue
		}
		if !m.SnapshotTime.After(base.SnapshotTime) {
			continue
		}
		if targetTime != nil && m.Since.After(*targetTime) {
			continue
		}
		incrementals = append(incrementals, m)
	}
	return base, incrementals, nil
}

func (r *Restorer) prepareTarget(ctx context.Context) error {
	db := r.target.Database()
	if !r.cfg.DropDatabase {
		return nil
	}
	r.logger.Info("Dropping and recreating target database %q", db)
	params := pg.Params{"db": db}
	if err := r.target.ExecOnDatabase(ctx, "postgres",
		`DROP DATABASE IF EXISTS :"db" WITH (FORCE)`, params); err != nil {
		return err
	}
	return r.target.ExecOnDatabase(ctx, "postgres", `CREATE DATABASE :"db"`, params)
}

func (r *Restorer) restoreBase(ctx context.Context, base *backup.Metadata) error {
	if err := r.target.CheckDumpCompatibility(ctx); err != nil {
		return err
	}
	dir := filepath.Join(r.outputDir, base.ID)
	if err := r.verifyIntegrity(dir); err != nil {
		return err
	}
	if err := r.assembler.AssembleFiles(dir, base.Splits); err != nil {
		return fmt.Errorf("reassemble parts: %w", err)
	}
	dataDir := filepath.Join(dir, "data")
	if _, err := os.Stat(dataDir); err != nil {
		return fmt.Errorf("base data directory missing: %w", err)
	}
	r.logger.Info("Running pg_restore (%d jobs)", r.cfg.Jobs)
	return r.target.PgRestoreDirectory(ctx, dataDir, r.cfg.Jobs, !r.cfg.DropDatabase)
}

func (r *Restorer) applyIncremental(ctx context.Context, inc *backup.Metadata, bound string) error {
	dir := filepath.Join(r.outputDir, inc.ID)
	if err := r.verifyIntegrity(dir); err != nil {
		return err
	}
	if err := r.assembler.AssembleFiles(dir, inc.Splits); err != nil {
		return fmt.Errorf("reassemble parts: %w", err)
	}

	script, cleanup, err := r.buildApplyScript(ctx, dir, inc)
	if err != nil {
		return err
	}
	defer cleanup()

	if script == "" {
		r.logger.Info("Incremental %s has no rows to apply", inc.ID)
		return nil
	}
	return r.target.RunScriptFile(ctx, script, pg.Params{"bound": bound})
}

func (r *Restorer) buildApplyScript(ctx context.Context, backupDir string, inc *backup.Metadata) (string, func(), error) {
	var sb strings.Builder
	hasWork := false

	cleanRoot := filepath.Clean(backupDir)
	for i, t := range inc.Tables {
		dataPath := filepath.Clean(filepath.Join(backupDir, t.DataFile))
		if dataPath != cleanRoot && !strings.HasPrefix(dataPath, cleanRoot+string(os.PathSeparator)) {
			return "", nil, fmt.Errorf("data file %q escapes backup directory", t.DataFile)
		}
		if info, err := os.Stat(dataPath); err != nil || info.Size() == 0 {
			continue
		}

		current, err := r.target.Columns(ctx, t.Schema, t.Name)
		if err != nil {
			return "", nil, err
		}
		pk, err := r.target.PrimaryKey(ctx, t.Schema, t.Name)
		if err != nil {
			return "", nil, err
		}
		if len(current) == 0 || len(pk) == 0 {
			r.logger.Warn("Skipping %s.%s during apply (missing columns or primary key)", t.Schema, t.Name)
			continue
		}

		cols := t.Columns
		if len(cols) == 0 {
			cols = current
		}
		currentSet := map[string]bool{}
		for _, c := range current {
			currentSet[c] = true
		}
		for _, c := range cols {
			if !currentSet[c] {
				return "", nil, fmt.Errorf("column %q of %s.%s exists in backup but not in target; schema drift", c, t.Schema, t.Name)
			}
		}

		qual := sqlutil.QuoteQualified(t.Schema, t.Name)
		tmp := fmt.Sprintf("_cdc_load_%d", i)
		colList := sqlutil.JoinIdents(cols)
		upsert := buildUpsert(pk, cols)

		identity, err := r.target.HasIdentityAlways(ctx, t.Schema, t.Name)
		if err != nil {
			return "", nil, err
		}
		overriding := ""
		if identity {
			overriding = "OVERRIDING SYSTEM VALUE "
		}

		fmt.Fprintf(&sb, "BEGIN;\n")
		fmt.Fprintf(&sb, "CREATE TEMP TABLE %s (LIKE %s INCLUDING DEFAULTS) ON COMMIT DROP;\n", tmp, qual)
		fmt.Fprintf(&sb, "\\copy %s (%s) FROM '%s' WITH (FORMAT csv, HEADER true)\n", tmp, colList, sqlutil.EscapeLiteral(dataPath))
		fmt.Fprintf(&sb, "DELETE FROM %s WHERE %s > :'bound'::timestamptz;\n", tmp, t.ChangeExpr)
		fmt.Fprintf(&sb, "INSERT INTO %s (%s) %sSELECT %s FROM %s %s;\n", qual, colList, overriding, colList, tmp, upsert)
		fmt.Fprintf(&sb, "COMMIT;\n\n")
		hasWork = true
	}

	if !hasWork {
		return "", func() {}, nil
	}

	f, err := os.CreateTemp("", "cdc-apply-*.sql")
	if err != nil {
		return "", nil, err
	}
	path := f.Name()
	cleanup := func() { _ = os.Remove(path) }
	if _, err := f.WriteString(sb.String()); err != nil {
		f.Close()
		cleanup()
		return "", nil, err
	}
	if err := f.Close(); err != nil {
		cleanup()
		return "", nil, err
	}
	return path, cleanup, nil
}

func (r *Restorer) sweepSoftDeletes(ctx context.Context, bound string) error {
	if len(r.pitr.DeletedAtColumns) == 0 {
		return nil
	}
	tables, _, err := r.target.DiscoverCDCTables(ctx,
		r.pitr.Schemas, r.pitr.CreatedAtColumns, r.pitr.UpdatedAtColumns, r.pitr.DeletedAtColumns)
	if err != nil {
		return err
	}
	for _, t := range tables {
		if t.DeletedAt == "" {
			continue
		}
		qual := sqlutil.QuoteQualified(t.Schema, t.Name)
		col := sqlutil.QuoteIdent(t.DeletedAt)
		sql := fmt.Sprintf("DELETE FROM %s WHERE %s IS NOT NULL AND %s <= :'bound'::timestamptz", qual, col, col)
		if err := r.target.Exec(ctx, sql, pg.Params{"bound": bound}); err != nil {
			return fmt.Errorf("sweep %s.%s: %w", t.Schema, t.Name, err)
		}
	}
	return nil
}

func boundValue(t *time.Time) string {
	if t == nil {
		return "infinity"
	}
	return t.UTC().Format("2006-01-02 15:04:05.999999") + "+00"
}

func buildUpsert(pk, cols []string) string {
	pkSet := map[string]bool{}
	for _, c := range pk {
		pkSet[c] = true
	}
	var sets []string
	for _, c := range cols {
		if pkSet[c] {
			continue
		}
		sets = append(sets, fmt.Sprintf("%s = EXCLUDED.%s", sqlutil.QuoteIdent(c), sqlutil.QuoteIdent(c)))
	}
	if len(sets) == 0 {
		return fmt.Sprintf("ON CONFLICT (%s) DO NOTHING", sqlutil.JoinIdents(pk))
	}
	return fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s", sqlutil.JoinIdents(pk), strings.Join(sets, ", "))
}
