package cron

import (
	"fcstask-backend/backup-tool/pkg/backup"
	"fcstask-backend/backup-tool/pkg/logging"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron     *cron.Cron
	schedule string
	backuper *backup.Backuper
	logger   *logging.Logger
	running  bool
}

func NewScheduler(cronExpr string, backuper *backup.Backuper, logger *logging.Logger) *Scheduler {
	return &Scheduler{
		schedule: cronExpr,
		backuper: backuper,
		logger:   logger,
		running:  false,
	}
}

func (s *Scheduler) Start() {
	if s.running {
		s.logger.Warn("Scheduler is already running")
		return
	}

	s.cron = cron.New(cron.WithSeconds())

	_, err := s.cron.AddFunc(s.schedule, func() {
		s.logger.Info("Starting scheduled backup job")
		startTime := time.Now()

		if err := s.backuper.CreateBackup(); err != nil {
			s.logger.Error("Scheduled backup failed: %v", err)
			return
		}

		duration := time.Since(startTime)
		s.logger.Info("Scheduled backup completed in %v", duration)
	})

	if err != nil {
		s.logger.Error("Failed to parse cron expression '%s': %v", s.schedule, err)
		return
	}

	go func() {
		s.logger.Info("Running initial backup")
		if err := s.backuper.CreateBackup(); err != nil {
			s.logger.Error("Initial backup failed: %v", err)
		}
	}()

	s.cron.Start()
	s.running = true

	s.logger.Info("Scheduler started with schedule: %s", s.schedule)
	s.logNextRun()
}

func (s *Scheduler) logNextRun() {
	if s.cron != nil {
		entries := s.cron.Entries()
		if len(entries) > 0 {
			nextRun := entries[0].Next
			s.logger.Info("Next backup scheduled at: %v", nextRun.Format("2006-01-02 15:04:05"))
		}
	}
}

func (s *Scheduler) Stop() {
	if s.running && s.cron != nil {
		s.cron.Stop()
		s.running = false
		s.logger.Info("Scheduler stopped")
	}
}
