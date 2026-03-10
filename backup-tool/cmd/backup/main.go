package main

import (
	"fcstask-backend/backup-tool/config"
	"fcstask-backend/backup-tool/pkg/backup"
	"fcstask-backend/backup-tool/pkg/cron"
	"fcstask-backend/backup-tool/pkg/logging"
	"fcstask-backend/backup-tool/pkg/storage"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		configPath  string
		runNow      bool
		incremental bool
	)

	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
	flag.BoolVar(&runNow, "run-now", false, "Run backup immediately and exit")
	flag.BoolVar(&incremental, "incremental", false, "Use incremental backup mode")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		panic(err)
	}

	logger := logging.NewLogger(cfg.Logging)
	defer logger.Close()

	logger.Info("Starting backup tool")

	diskChecker := storage.NewDiskChecker()
	if err := diskChecker.CheckFreeSpace(cfg.Backup.OutputDir, cfg.Backup.MinFreeSpaceGB); err != nil {
		logger.Error("Disk space check failed: %v", err)
		os.Exit(1)
	}

	backuper := backup.NewBackuper(cfg.Postgres, cfg.Backup, logger)

	if incremental {
		backuper.SetIncrementalMode(true)
		logger.Info("Incremental backup mode enabled")
	}

	if runNow {
		if err := backuper.CreateBackup(); err != nil {
			logger.Error("Backup failed: %v", err)
			os.Exit(1)
		}
		logger.Info("Backup completed successfully")
		return
	}

	scheduler := cron.NewScheduler(cfg.Cron.Schedule, backuper, logger)
	scheduler.Start()

	logger.Info("Scheduler started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutdown signal received")

	scheduler.Stop()
	logger.Info("Application stopped")
}
