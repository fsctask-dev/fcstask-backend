package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"fcstask-backend/backup-service/internal/backup"
	"fcstask-backend/backup-service/internal/config"
	"fcstask-backend/backup-service/internal/health"
	"fcstask-backend/backup-service/internal/lock"
	"fcstask-backend/backup-service/internal/logging"
	"fcstask-backend/backup-service/internal/pg"
	"fcstask-backend/backup-service/internal/restore"
	"fcstask-backend/backup-service/internal/scheduler"
	"fcstask-backend/backup-service/internal/storage"
	"fcstask-backend/backup-service/internal/wal"
)

const lockFileName = ".backup-service.lock"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return errors.New("no subcommand given")
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "backup":
		return cmdBackup(rest)
	case "serve":
		return cmdServe(rest)
	case "restore":
		return cmdRestore(rest)
	case "list":
		return cmdList(rest)
	case "verify":
		return cmdVerify(rest)
	case "wal-backup":
		return cmdWALBackup(rest)
	case "wal-archive":
		return cmdWALArchive(rest)
	case "wal-restore":
		return cmdWALRestore(rest)
	case "wal-list":
		return cmdWALList(rest)
	case "-h", "--help", "help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown subcommand %q", cmd)
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, `backup-service - PostgreSQL backup & PITR microservice

usage:
  backup-service backup  [--config FILE] [--full | --incremental]
  backup-service serve   [--config FILE]
  backup-service restore [--config FILE] (--time RFC3339 | --latest)
  backup-service list    [--config FILE]
  backup-service verify  [--config FILE]

  backup-service wal-backup  [--config FILE]
  backup-service wal-archive [--config FILE]
  backup-service wal-restore [--config FILE] (--time RFC3339 | --latest)
  backup-service wal-list    [--config FILE]

examples:
  backup-service backup --config config.yaml --full
  backup-service serve  --config config.yaml
  backup-service restore --config config.yaml --time 2026-06-07T00:00:00Z
  backup-service restore --config config.yaml --latest

  backup-service wal-backup  --config config.yaml
  backup-service wal-archive --config config.yaml
  backup-service wal-restore --config config.yaml --time 2026-06-07T00:00:00Z
`)
}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
}

func loadConfigAndLogger(configPath string) (*config.Config, *logging.Logger, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, err
	}
	logger, err := logging.New(cfg.Logging,
		cfg.Source.Password,
		cfg.RestoreTarget.Password,
		cfg.WAL.Replication.Password,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("init logger: %w", err)
	}
	return cfg, logger, nil
}

func gracefulShutdown(logger *logging.Logger, stop func()) {
	const timeout = 30 * time.Second
	done := make(chan struct{})
	go func() {
		stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		logger.Warn("Graceful shutdown timed out after %s; exiting", timeout)
	}
}

func operationContext(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, d)
}

func cmdBackup(args []string) error {
	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	configPath := fs.String("config", "config.yaml", "path to configuration file")
	full := fs.Bool("full", false, "force a full base backup")
	incremental := fs.Bool("incremental", false, "force an incremental backup (fails over to full if no base exists)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *full && *incremental {
		return errors.New("--full and --incremental are mutually exclusive")
	}

	cfg, logger, err := loadConfigAndLogger(*configPath)
	if err != nil {
		return err
	}
	defer logger.Close()

	ctx, cancel := signalContext()
	defer cancel()
	ctx, tcancel := operationContext(ctx, cfg.Backup.CommandTimeout())
	defer tcancel()

	if err := os.MkdirAll(cfg.Backup.OutputDir, 0o700); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	fl := lock.New(filepath.Join(cfg.Backup.OutputDir, lockFileName))
	if err := fl.TryAcquire(); err != nil {
		if errors.Is(err, lock.ErrLocked) {
			logger.Warn("another backup-service operation is already running; aborting")
			return err
		}
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer fl.Release()

	if cfg.Backup.MinFreeSpaceGB > 0 {
		if err := storage.NewDiskChecker().CheckFreeSpace(ctx, cfg.Backup.OutputDir, cfg.Backup.MinFreeSpaceGB); err != nil {
			return fmt.Errorf("disk space check: %w", err)
		}
	}

	src := pg.New(cfg.Source)
	if err := src.Ping(ctx); err != nil {
		return fmt.Errorf("cannot reach source database: %w", err)
	}

	b := backup.NewBackuper(src, cfg.Backup, cfg.PITR, logger)
	if *full {
		b.ForceFull(true)
	}
	if *incremental {
		b.ForceIncremental(true)
	}

	start := time.Now()
	meta, err := b.CreateBackup(ctx)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	logger.Info("backup %s (%s) completed in %s", meta.ID, meta.Type, time.Since(start).Round(time.Second))
	return nil
}

func cmdServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	configPath := fs.String("config", "config.yaml", "path to configuration file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, logger, err := loadConfigAndLogger(*configPath)
	if err != nil {
		return err
	}
	defer logger.Close()

	ctx, cancel := signalContext()
	defer cancel()

	if err := os.MkdirAll(cfg.Backup.OutputDir, 0o700); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	src := pg.New(cfg.Source)
	if err := src.Ping(ctx); err != nil {
		return fmt.Errorf("cannot reach source database: %w", err)
	}
	if err := src.CheckDumpCompatibility(ctx); err != nil {
		return err
	}

	b := backup.NewBackuper(src, cfg.Backup, cfg.PITR, logger)
	fl := lock.New(filepath.Join(cfg.Backup.OutputDir, lockFileName))

	sched := scheduler.New(
		cfg.Cron.Schedule,
		cfg.Cron.RunInitialOnStart,
		b,
		fl,
		cfg.Backup.CommandTimeout(),
		logger,
	)

	var hs *health.Server
	if cfg.Health.Addr != "" {
		hs = health.New(cfg.Health.Addr, logger)
		sched.OnResult(func(err error) {
			if err != nil {
				hs.RecordError(err.Error())
				return
			}
			hs.RecordSuccess(time.Now())
		})
		hs.Start()
	}

	if err := sched.Start(ctx); err != nil {
		return fmt.Errorf("start scheduler: %w", err)
	}
	if hs != nil {
		hs.SetReady(true)
	}
	logger.Info("scheduler running with schedule %q; press Ctrl-C to stop", cfg.Cron.Schedule)

	<-ctx.Done()
	logger.Info("shutdown signal received, stopping")
	gracefulShutdown(logger, func() {
		sched.Stop()
		if hs != nil {
			sctx, scancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer scancel()
			_ = hs.Shutdown(sctx)
		}
	})
	logger.Info("backup-service stopped")
	return nil
}

func cmdRestore(args []string) error {
	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	configPath := fs.String("config", "config.yaml", "path to configuration file")
	timeStr := fs.String("time", "", "target point in time, RFC3339 (e.g. 2026-06-07T00:00:00Z)")
	latest := fs.Bool("latest", false, "restore to the latest available state")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *timeStr == "" && !*latest {
		return errors.New("either --time or --latest is required")
	}
	if *timeStr != "" && *latest {
		return errors.New("--time and --latest are mutually exclusive")
	}

	var targetTime *time.Time
	if *timeStr != "" {
		t, err := time.Parse(time.RFC3339, *timeStr)
		if err != nil {
			return fmt.Errorf("invalid --time %q (want RFC3339): %w", *timeStr, err)
		}
		tu := t.UTC()
		targetTime = &tu
	}

	cfg, logger, err := loadConfigAndLogger(*configPath)
	if err != nil {
		return err
	}
	defer logger.Close()

	if cfg.RestoreTarget.Host == "" {
		return errors.New("restore_target is not configured in the config file")
	}
	if cfg.RestoreTarget.Host == cfg.Source.Host &&
		cfg.RestoreTarget.Port == cfg.Source.Port &&
		cfg.RestoreTarget.Database == cfg.Source.Database {
		return errors.New("refusing to restore: restore_target is identical to the source database; restore must target a separate database")
	}

	ctx, cancel := signalContext()
	defer cancel()
	ctx, tcancel := operationContext(ctx, cfg.Backup.CommandTimeout())
	defer tcancel()

	fl := lock.New(filepath.Join(cfg.Backup.OutputDir, lockFileName))
	if err := fl.TryAcquire(); err != nil {
		if errors.Is(err, lock.ErrLocked) {
			logger.Warn("another backup-service operation is already running; aborting restore")
			return err
		}
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer fl.Release()

	target := pg.New(cfg.RestoreTarget)

	if err := target.Ping(ctx); err != nil {
		logger.Warn("target ping failed (this is expected if the database does not exist yet): %v", err)
	}

	r := restore.NewRestorer(target, cfg.Restore, cfg.PITR, cfg.Backup.OutputDir, logger)

	start := time.Now()
	if err := r.RestoreToTime(ctx, targetTime); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}
	if targetTime != nil {
		logger.Info("restore to %s completed in %s", targetTime.Format(time.RFC3339), time.Since(start).Round(time.Second))
	} else {
		logger.Info("restore to latest completed in %s", time.Since(start).Round(time.Second))
	}
	return nil
}

func cmdList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	configPath := fs.String("config", "config.yaml", "path to configuration file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	backups, err := backup.ListBackups(cfg.Backup.OutputDir)
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}
	if len(backups) == 0 {
		fmt.Println("no backups found in", cfg.Backup.OutputDir)
		return nil
	}

	fmt.Printf("%-32s %-12s %-22s %-22s %s\n", "ID", "TYPE", "SNAPSHOT (UTC)", "WINDOW START (UTC)", "BASE FULL")
	for _, m := range backups {
		since := "-"
		if !m.Since.IsZero() {
			since = m.Since.UTC().Format("2006-01-02 15:04:05")
		}
		fmt.Printf("%-32s %-12s %-22s %-22s %s\n",
			m.ID,
			m.Type,
			m.SnapshotTime.UTC().Format("2006-01-02 15:04:05"),
			since,
			m.BaseFull,
		)
	}
	return nil
}

func cmdVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	configPath := fs.String("config", "config.yaml", "path to configuration file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	backups, err := backup.ListBackups(cfg.Backup.OutputDir)
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}
	if len(backups) == 0 {
		fmt.Println("no backups found in", cfg.Backup.OutputDir)
		return nil
	}

	var failed, skipped int
	for _, m := range backups {
		dir := filepath.Join(cfg.Backup.OutputDir, m.ID)
		verr := backup.VerifyChecksums(dir)
		switch {
		case verr == nil:
			fmt.Printf("OK       %s\n", m.ID)
		case errors.Is(verr, backup.ErrNoChecksums):
			skipped++
			fmt.Printf("SKIP     %s (no checksums recorded)\n", m.ID)
		default:
			failed++
			fmt.Printf("FAIL     %s: %v\n", m.ID, verr)
		}
	}
	fmt.Printf("\n%d backups, %d failed, %d skipped\n", len(backups), failed, skipped)
	if failed > 0 {
		return fmt.Errorf("%d backup(s) failed integrity verification", failed)
	}
	return nil
}

const walLockFileName = ".backup-service-wal.lock"

func requireWAL(cfg *config.Config) error {
	if !cfg.WAL.Enabled {
		return errors.New("wal is not enabled in the config (set wal.enabled: true)")
	}
	return nil
}

func cmdWALBackup(args []string) error {
	fs := flag.NewFlagSet("wal-backup", flag.ContinueOnError)
	configPath := fs.String("config", "config.yaml", "path to configuration file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, logger, err := loadConfigAndLogger(*configPath)
	if err != nil {
		return err
	}
	defer logger.Close()
	if err := requireWAL(cfg); err != nil {
		return err
	}

	ctx, cancel := signalContext()
	defer cancel()
	ctx, tcancel := operationContext(ctx, cfg.Backup.CommandTimeout())
	defer tcancel()

	if err := os.MkdirAll(cfg.WAL.BaseBackupDir, 0o700); err != nil {
		return fmt.Errorf("create base backup dir: %w", err)
	}

	if cfg.Backup.MinFreeSpaceGB > 0 {
		if err := storage.NewDiskChecker().CheckFreeSpace(ctx, cfg.WAL.BaseBackupDir, cfg.Backup.MinFreeSpaceGB); err != nil {
			return fmt.Errorf("disk space check: %w", err)
		}
	}

	fl := lock.New(filepath.Join(cfg.WAL.BaseBackupDir, walLockFileName))
	if err := fl.TryAcquire(); err != nil {
		if errors.Is(err, lock.ErrLocked) {
			logger.Warn("another WAL base backup is already running; aborting")
			return err
		}
		return fmt.Errorf("acquire wal lock: %w", err)
	}
	defer fl.Release()

	b := wal.NewBaseBackuper(cfg.WAL, logger)
	start := time.Now()
	meta, err := b.TakeBaseBackup(ctx)
	if err != nil {
		return fmt.Errorf("wal base backup failed: %w", err)
	}
	logger.Info("WAL base backup %s completed in %s", meta.ID, time.Since(start).Round(time.Second))
	return nil
}

func cmdWALArchive(args []string) error {
	fs := flag.NewFlagSet("wal-archive", flag.ContinueOnError)
	configPath := fs.String("config", "config.yaml", "path to configuration file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, logger, err := loadConfigAndLogger(*configPath)
	if err != nil {
		return err
	}
	defer logger.Close()
	if err := requireWAL(cfg); err != nil {
		return err
	}

	ctx, cancel := signalContext()
	defer cancel()

	var hs *health.Server
	if cfg.Health.Addr != "" {
		hs = health.New(cfg.Health.Addr, logger)
		hs.Start()
		hs.SetReady(true)
	}

	archiver := wal.NewArchiver(cfg.WAL, logger)
	logger.Info("WAL archiver starting; press Ctrl-C to stop")
	err = archiver.Run(ctx)
	if hs != nil {
		sctx, scancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = hs.Shutdown(sctx)
		scancel()
	}
	if err != nil {
		return fmt.Errorf("wal archiver: %w", err)
	}
	return nil
}

func cmdWALRestore(args []string) error {
	fs := flag.NewFlagSet("wal-restore", flag.ContinueOnError)
	configPath := fs.String("config", "config.yaml", "path to configuration file")
	timeStr := fs.String("time", "", "target point in time, RFC3339")
	latest := fs.Bool("latest", false, "recover to the latest available WAL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *timeStr == "" && !*latest {
		return errors.New("either --time or --latest is required")
	}
	if *timeStr != "" && *latest {
		return errors.New("--time and --latest are mutually exclusive")
	}

	var targetTime *time.Time
	if *timeStr != "" {
		t, err := time.Parse(time.RFC3339, *timeStr)
		if err != nil {
			return fmt.Errorf("invalid --time %q (want RFC3339): %w", *timeStr, err)
		}
		tu := t.UTC()
		targetTime = &tu
	}

	cfg, logger, err := loadConfigAndLogger(*configPath)
	if err != nil {
		return err
	}
	defer logger.Close()
	if err := requireWAL(cfg); err != nil {
		return err
	}

	ctx, cancel := signalContext()
	defer cancel()
	ctx, tcancel := operationContext(ctx, cfg.Backup.CommandTimeout())
	defer tcancel()

	r := wal.NewRestorer(cfg.WAL, logger)
	start := time.Now()
	if err := r.RestoreToTime(ctx, targetTime); err != nil {
		return fmt.Errorf("wal restore failed: %w", err)
	}
	logger.Info("WAL restore prepared in %s", time.Since(start).Round(time.Second))
	return nil
}

func cmdWALList(args []string) error {
	fs := flag.NewFlagSet("wal-list", flag.ContinueOnError)
	configPath := fs.String("config", "config.yaml", "path to configuration file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	if !cfg.WAL.Enabled {
		return requireWAL(cfg)
	}

	bases, err := wal.ListBaseBackups(cfg.WAL.BaseBackupDir)
	if err != nil {
		return fmt.Errorf("list base backups: %w", err)
	}
	if len(bases) == 0 {
		fmt.Println("no base backups found in", cfg.WAL.BaseBackupDir)
		return nil
	}
	fmt.Printf("%-28s %-22s %-22s %s\n", "ID", "STARTED (UTC)", "FINISHED (UTC)", "START WAL")
	for _, m := range bases {
		fmt.Printf("%-28s %-22s %-22s %s\n",
			m.ID,
			m.StartedAt.UTC().Format("2006-01-02 15:04:05"),
			m.FinishedAt.UTC().Format("2006-01-02 15:04:05"),
			m.StartWALFile,
		)
	}
	return nil
}
